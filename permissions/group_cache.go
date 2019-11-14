package permissions

import (
	"errors"
	"sync/atomic"

	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/users"
)

type Permissions = users.Permissions

var _ users.PermGroupCache = &GroupCache{}

type GroupCache struct {
	concurrency.Tickable
	permissions atomic.Value
}

func (cache *GroupCache) GetPermissionsByTag(tag string) ([]Permission, error) {
	return GetPermissionsByTag(tag)
}

//从缓存中获取权限对象
func (cache *GroupCache) Get(id int64) *Permissions {
	values := cache.getPermissions()
	if len(values) == 0 {
		return nil
	}
	return values[id]
}

//从缓存中获取子组
func (cache *GroupCache) GetChildren(id int64) []*Permissions {
	values := cache.getPermissions()
	if len(values) == 0 {
		return nil
	}
	var children []*Permissions
	for _, group := range values {
		if group.ParentID == id {
			children = append(children, group)
		}
	}
	return children
}

func (cache *GroupCache) setPermissions(values map[int64]*Permissions) {
	cache.permissions.Store(values)
}

func (cache *GroupCache) getPermissions() map[int64]*Permissions {
	o := cache.permissions.Load()
	if o == nil {
		return nil
	}
	values, ok := o.(map[int64]*Permissions)
	if !ok {
		return nil
	}
	return values
}

func (cache *GroupCache) refresh(db *DB) error {
	var valueArray []*Permissions
	err := db.PermissionGroups().Where().All(&valueArray)
	if err != nil {
		return errors.New("query permission groups fail: " + err.Error())
	}

	var values = map[int64]*Permissions{}
	for _, p := range valueArray {
		values[p.ID] = p
	}

	var pagArray []PermissionAndGroup
	err = db.PermissionsAndGroups().Where().All(&pagArray)
	if err != nil {
		return errors.New("query permission groups fail: " + err.Error())
	}
	for _, p := range pagArray {
		if old, ok := values[p.GroupID]; ok {
			if p.Type == PERMISSION_TAG {
				old.PermissionTags = append(old.PermissionTags, p.PermissionObject)
			} else {
				old.PermissionIDs = append(old.PermissionIDs, p.PermissionObject)
			}
		}
	}

	cache.setPermissions(values)
	return nil
}
