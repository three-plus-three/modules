package permissions

import (
	"errors"
	"sync/atomic"
)

type Permissions struct {
	PermissionGroup
	Permissions []string
}

type GroupCache struct {
	permissions atomic.Value
}

//从缓存中获取权限对象
func (cache *GroupCache) Get(id int64) *Permissions {
	values := cache.getPermissions()
	if len(values) == 0 {
		return nil
	}
	return values[id]
}

func (cache *GroupCache) savePermissions(values map[int64]*Permissions) {
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

	cache.savePermissions(values)
}
