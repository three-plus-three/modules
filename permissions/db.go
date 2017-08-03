package permissions

import (
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/runner-mei/orm"
)

type DB struct {
	Engine *xorm.Engine
}

func (db *DB) PermissionGroup() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroup{}
	})(db.Engine)
}
func (db *DB) User() *orm.Collection {
	return orm.New(func() interface{} {
		return &User{}
	})(db.Engine)
}
func (db *DB) UserAndRole() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndRole{}
	})(db.Engine)
}
func (db *DB) PermissionAndGroup() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionAndGroup{}
	})(db.Engine)
}
func (db *DB) Role() *orm.Collection {
	return orm.New(func() interface{} {
		return &Role{}
	})(db.Engine)
}
func (db *DB) PermissionGroupAndRole() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroupAndRole{}
	})(db.Engine)
}
func (db *DB) UserGroup() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserGroup{}
	})(db.Engine)
}
func (db *DB) UserAndUserGroup() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndUserGroup{}
	})(db.Engine)
}

func InitTables(engine *xorm.Engine) error {
	beans := []interface{}{
		&PermissionGroup{},
		&User{},
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
