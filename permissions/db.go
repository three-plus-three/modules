package permissions

import (
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/runner-mei/orm"
)

type DB struct {
	Engine *xorm.Engine
}

func (db *DB) PermissionGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroup{}
	})(db.Engine)
}
func (db *DB) Users() *orm.Collection {
	return orm.New(func() interface{} {
		return &User{}
	})(db.Engine)
}
func (db *DB) OnlineUsers() *orm.Collection {
	return orm.New(func() interface{} {
		return &OnlineUser{}
	})(db.Engine)
}
func (db *DB) UsersAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndRole{}
	})(db.Engine)
}
func (db *DB) PermissionsAndGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionAndGroup{}
	})(db.Engine)
}
func (db *DB) Roles() *orm.Collection {
	return orm.New(func() interface{} {
		return &Role{}
	})(db.Engine)
}
func (db *DB) PermissionGroupsAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroupAndRole{}
	})(db.Engine)
}
func (db *DB) UserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserGroup{}
	})(db.Engine)
}
func (db *DB) UsersAndUserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndUserGroup{}
	})(db.Engine)
}

func InitTables(engine *xorm.Engine) error {
	beans := []interface{}{
		&PermissionGroup{},
		&User{},
		&OnlineUser{},
		&Role{},
		&UserAndRole{},
		&PermissionAndGroup{},
		&PermissionGroupAndRole{},
		&UserGroup{},
		&UserAndUserGroup{},
	}

	if err := engine.CreateTables(beans...); err != nil {
		return err
	}

	for _, bean := range beans {
		if err := engine.CreateIndexes(bean); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}

		if err := engine.CreateUniques(bean); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}
	return nil
}

func DropTables(engine *xorm.Engine) error {
	beans := []interface{}{
		&UserAndRole{},
		&UserAndUserGroup{},
		&PermissionAndGroup{},
		&PermissionGroupAndRole{},
		&PermissionGroup{},
		&UserGroup{},
		&OnlineUser{},
		&User{},
		&Role{},
	}

	for _, bean := range beans {
		if err := engine.DropIndexes(bean); err != nil {
			if !strings.Contains(err.Error(), "already exists") {
				return err
			}
		}
	}

	if err := engine.DropTables(beans...); err != nil {
		return err
	}

	return nil
}
