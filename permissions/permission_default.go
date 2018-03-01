package permissions

import (
	"strings"

	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/errors"
)

func SaveDefaultPermissionGroups(db *DB, allDefaultGroups []Group) error {
	var allPermissionGroups []PermissionGroup
	err := db.PermissionGroups().
		Where(orm.Cond{"is_default": "true"}).
		All(&allPermissionGroups)
	if err != nil {
		return errors.Wrap(err, "GetAllPermissionGroups")
	}

	if len(allPermissionGroups) > 0 {
		deletedCount := 0
		for _, group := range allDefaultGroups {
			foundIndex := -1
			for idx := range allPermissionGroups {
				if group.Name == allPermissionGroups[idx].Name &&
					allPermissionGroups[idx].ParentID == 0 {
					if foundIndex >= 0 {
						err = deletePermissionGroups(db, allPermissionGroups[idx].ID)
						if err != nil {
							return err
						}
						deletedCount++
					} else {
						foundIndex = idx
					}
				}
			}
		}

		if deletedCount > 0 {
			allPermissionGroups = nil
			err = db.PermissionGroups().
				Where(orm.Cond{"is_default": "true"}).
				All(&allPermissionGroups)
			if err != nil {
				return errors.Wrap(err, "GetAllPermissionGroups")
			}
		}
	}

	for _, group := range allDefaultGroups {
		err = syncGroups(db, []Group{group}, allPermissionGroups, 0)
		if err != nil {
			return errors.Wrap(err, "载入缺省权限组")
		}
	}
	return nil
}

func syncGroups(db *DB, allDefaultGroups []Group,
	allPermissionGroups []PermissionGroup, parentID int64) error {

	for _, defaultGroup := range allDefaultGroups {
		foundIndex := -1
		for idx := range allPermissionGroups {
			if defaultGroup.Name == allPermissionGroups[idx].Name && allPermissionGroups[idx].ParentID == parentID {
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
			if len(defaultGroup.Children) > 0 {
				syncGroups(db, defaultGroup.Children, allPermissionGroups, allPermissionGroups[foundIndex].ID)
			}
		} else {
			err := insertPermissionGroups(db, defaultGroup, parentID)
			if err != nil {
				return err
			}
		}
	}

	for idx := range allPermissionGroups {
		foundIndex := -1
		for _, defaultGroup := range allDefaultGroups {
			if defaultGroup.Name == allPermissionGroups[idx].Name && allPermissionGroups[idx].ParentID == parentID {
				foundIndex = idx
				break
			}
		}
		if foundIndex < 0 && allPermissionGroups[idx].ParentID == parentID && parentID != 0 {
			err := deletePermissionGroups(db, allPermissionGroups[idx].ID)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func deletePermissionGroups(db *DB, groupID int64) error {
	var roleAndpermissionGroups []PermissionGroupAndRole
	err := db.PermissionGroupsAndRoles().Where(orm.Cond{"group_id": groupID}).All(&roleAndpermissionGroups)
	if err != nil {
		return errors.Wrap(err, "获取角色与权限组关系")
	}
	if len(roleAndpermissionGroups) > 0 {
		var permissionGroup PermissionGroup
		err = db.PermissionGroups().Id(groupID).Get(&permissionGroup)
		if err != nil {
			return errors.Wrap(err, "获取权限组")
		}
		if strings.Index(permissionGroup.Name, "(已删除)") < 0 {
			permissionGroup.Name = permissionGroup.Name + "(已删除)"
		}
		err := db.PermissionGroups().Id(permissionGroup.ID).Nullable("parent_id").Update(&permissionGroup)
		if err != nil {
			return errors.Wrap(err, "更新权限组失败")
		}
	} else {
		err = db.PermissionGroups().Id(groupID).Delete()
		if err != nil {
			return errors.Wrap(err, "删除权限组")
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

func updatePermissionGroups(db *DB, group Group, permissionGroup PermissionGroup) error {
	permissionGroup.Name = group.Name
	permissionGroup.Description = group.Description
	err := db.PermissionGroups().Id(permissionGroup.ID).Nullable("parent_id").Update(&permissionGroup)
	if err != nil {
		return errors.Wrap(err, "更新权限组失败")
	}

	err = updatePerssionsAndGroup(db, group.PermissionTags, permissionGroup.ID, PERMISSION_TAG)
	if err != nil {
		return err
	}

	err = updatePerssionsAndGroup(db, group.PermissionIDs, permissionGroup.ID, PERMISSION_ID)
	if err != nil {
		return err
	}

	return nil
}

//更新权限组与权限关系
func updatePerssionsAndGroup(db *DB, ids []string,
	groupID int64, permissionAndGroupType int64) error {

	var permissionsAndGroupsInDB []PermissionAndGroup
	err := db.PermissionsAndGroups().Where(orm.Cond{"group_id": groupID, "type": permissionAndGroupType}).All(&permissionsAndGroupsInDB)
	if err != nil {
		return errors.Wrap(err, "获取权限组与权限关系")
	}

	var created, deleted []string

	for _, permissionsAndGroup := range permissionsAndGroupsInDB {
		found := false
		for _, permissionID := range ids {
			if permissionID == permissionsAndGroup.PermissionObject {
				found = true
				break
			}
		}
		if !found {
			deleted = append(deleted, permissionsAndGroup.PermissionObject)
		}
	}

	for _, permissionID := range ids {
		found := false
		for _, permissionsAndGroup := range permissionsAndGroupsInDB {
			if permissionID == permissionsAndGroup.PermissionObject {
				found = true
				break
			}
		}
		if !found {
			created = append(created, permissionID)
		}
	}

	_, err = db.PermissionsAndGroups().Where(orm.Cond{"group_id": groupID}).And(orm.Cond{"permission_object IN": deleted}).Delete()
	if err != nil {
		return errors.Wrap(err, "删除全权限组与权限关系")
	}

	for _, v := range created {
		var permissionAndGroup PermissionAndGroup
		permissionAndGroup.GroupID = groupID
		permissionAndGroup.PermissionObject = v
		permissionAndGroup.Type = permissionAndGroupType
		_, err = db.PermissionsAndGroups().Insert(&permissionAndGroup)
		if err != nil {
			return errors.Wrap(err, "InsertPermissionsAndGroups")
		}
	}
	return nil
}
