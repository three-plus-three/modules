package permissions

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-xorm/xorm"
	"github.com/runner-mei/orm"
)

type DB struct {
	Engine  *xorm.Engine
	session *xorm.Session
}

func (db *DB) WithSession(sess *xorm.Session) *DB {
	return &DB{Engine: db.Engine, session: sess}
}

func (db *DB) Begin() (*DB, error) {
	if db.session != nil {
		return nil, errors.New("run in the transaction")
	}
	session := db.Engine.NewSession()
	return &DB{Engine: db.Engine, session: session}, nil
}

func (db *DB) Commit() error {
	if db.session == nil {
		return sql.ErrTxDone
	}
	err := db.session.Commit()
	db.session = nil
	return err
}

func (db *DB) Rollback() error {
	if db.session == nil {
		return sql.ErrTxDone
	}
	err := db.session.Rollback()
	db.session = nil
	return err
}

func (db *DB) Close() error {
	return db.Rollback()
}

func (db *DB) PermissionGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroup{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) Users() *orm.Collection {
	return orm.New(func() interface{} {
		return &User{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) OnlineUsers() *orm.Collection {
	return orm.New(func() interface{} {
		return &OnlineUser{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) UsersAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndRole{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) PermissionsAndGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionAndGroup{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) Roles() *orm.Collection {
	return orm.New(func() interface{} {
		return &Role{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) PermissionGroupsAndRoles() *orm.Collection {
	return orm.New(func() interface{} {
		return &PermissionGroupAndRole{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) UserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserGroup{}
	})(db.Engine).WithSession(db.session)
}
func (db *DB) UsersAndUserGroups() *orm.Collection {
	return orm.New(func() interface{} {
		return &UserAndUserGroup{}
	})(db.Engine).WithSession(db.session)
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
			if !strings.Contains(err.Error(), "does not exist") {
				return err
			}
		}
	}

	if err := engine.DropTables(beans...); err != nil {
		return err
	}

	return nil
}
