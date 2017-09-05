package permissions

import (
	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/errors"
)

func SaveDefaultPermissionGroups(db *DB) error {
	allDefaultGroups, err := GetDefaultPermissionGroups()
	if err != nil {
		return errors.Wrap(err, "载入缺省权限组")
	}
	var allPermissionGroups []PermissionGroup
	err = db.PermissionGroups().
		Where(orm.Cond{"is_default": "true"}).
		All(&allPermissionGroups)
	if err != nil {
		return errors.Wrap(err, "GetAllPermissionGroups")
	}
	err = syncGroups(db, allDefaultGroups, allPermissionGroups)
	if err != nil {
		return errors.Wrap(err, "载入缺省权限组")
	}
	return nil
}

func syncGroups(db *DB, allDefaultGroups []Group,
	allPermissionGroups []PermissionGroup) error {
	for _, defaultGroup := range allDefaultGroups {
		foundIndex := -1

		for idx := range allPermissionGroups {
			if defaultGroup.Name == allPermissionGroups[idx].Name {
				foundIndex = idx
				break
			}
		}

		if foundIndex >= 0 {
			err := updatePermissionGroups(db, defaultGroup,
				allPermissionGroups[foundIndex])
			if err != nil {
				return err
			}
		} else {
			err := insertPermissionGroups(db, defaultGroup, 0)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//添加权限组
func insertPermissionGroups(db *DB, group Group, parentID int64) error {
	var permissionGroup PermissionGroup
	permissionGroup.Name = group.Name
	permissionGroup.Description = group.Description
	permissionGroup.IsDefault = true
	if parentID != 0 {
		permissionGroup.ParentID = parentID
	}
	id, err := db.PermissionGroups().Nullable("parent_id").Insert(&permissionGroup)
	if err != nil {
		return errors.New("InsertPermissionGroups " + permissionGroup.Name +
			" fail:" + err.Error())
	}
	err = insertPerssionsAndGroup(db, group.PermissionIDs, group.PermissionTags, id.(int64))
	if err != nil {
		return err
	}
	if len(group.Children) != 0 {
		for _, child := range group.Children {
			err := insertPermissionGroups(db, child, id.(int64))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func insertPerssionsAndGroup(db *DB, permissionIDs []string,
	permissionTags []string, groupID int64) error {
	if len(permissionIDs) != 0 {
		for _, v := range permissionIDs {
			var perssionAndGroup PermissionAndGroup
			perssionAndGroup.GroupID = groupID
			perssionAndGroup.PermissionObject = v
			perssionAndGroup.Type = PERMISSION_ID
			_, err := db.PermissionsAndGroups().Insert(perssionAndGroup)
			if err != nil {
				return errors.New("InsertPermissionsAndGroups " + v + err.Error())
			}
		}
	}

	if len(permissionTags) != 0 {
		for _, v := range permissionTags {
			var perssionAndGroup PermissionAndGroup
			perssionAndGroup.GroupID = groupID
			perssionAndGroup.PermissionObject = v
			perssionAndGroup.Type = PERMISSION_TAG
			_, err := db.PermissionsAndGroups().Insert(perssionAndGroup)
			if err != nil {
				return errors.New("InsertPermissionsAndGroups " + v + err.Error())
			}
		}
	}
	return nil
}

func updatePermissionGroups(db *DB, group Group, peg PermissionGroup) error {
	err := db.PermissionGroups().Id(peg.ID).Delete()
	if err != nil {
		return errors.New("Delete PermissionGroups fail" + err.Error())
	}
	err = insertPermissionGroups(db, group, 0)
	if err != nil {
		return err
	}
	return nil
}
