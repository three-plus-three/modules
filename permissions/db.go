package permissions

import (
	"errors"
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/runner-mei/orm"
)

const EnabledUsers = "( disabled IS NULL or disabled = false)"

type DB struct {
	orm.DB
}

func (db *DB) WithSession(sess *xorm.Session) *DB {
	return &DB{DB: orm.DB{Engine: db.Engine, Session: sess}}
}

func (db *DB) Begin() (*DB, error) {
	if db.Session != nil {
		return nil, errors.New("run in the transaction")
	}
	session := db.Engine.NewSession()
	return db.WithSession(session), nil
}

func (db *DB) PermissionGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroup{}
	}, KeyForPermissionGroups)(db.Engine).WithSession(db.Session)
}
func (db *DB) Users() *orm.Collection {
	return orm.New(func() interface{} {
		return &User{}
	}, KeyForUsers)(db.Engine).WithSession(db.Session)
}
func (db *DB) OnlineUsers() *orm.Collection {
	return orm.New(func() interface{} {
		return &OnlineUser{}
	}, KeyForOnlineUsers)(db.Engine).WithSession(db.Session)
}
func (db *DB) UsersAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndRole{}
	}, KeyForUsersAndRoles)(db.Engine).WithSession(db.Session)
}
func (db *DB) PermissionsAndGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionAndGroup{}
	}, KeyForPermissionsAndGroups)(db.Engine).WithSession(db.Session)
}
func (db *DB) Roles() *orm.Collection {
	return orm.New(func() interface{} {
		return &Role{}
	}, KeyForRoles)(db.Engine).WithSession(db.Session)
}
func (db *DB) PermissionGroupsAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroupAndRole{}
	}, KeyForPermissionGroupsAndRoles)(db.Engine).WithSession(db.Session)
}
func (db *DB) UserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserGroup{}
	}, KeyForUserGroups)(db.Engine).WithSession(db.Session)
}
func (db *DB) UsersAndUserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndUserGroup{}
	}, KeyForUsersAndUserGroups)(db.Engine).WithSession(db.Session)
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
			if !strings.Contains(err.Error(), "does not exist") &&
				!strings.Contains(err.Error(), "不存在") {
				return err
			}
		}
	}

	if err := engine.DropTables(beans...); err != nil {
		return err
	}

	return nil
}
