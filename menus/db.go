package menus

import (
	"database/sql"
	"errors"

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
	if db.session == nil {
		return sql.ErrTxDone
	}
	db.session.Close()
	db.session = nil
	return nil
}

func (db *DB) Query(sqlStr string, args ...interface{}) orm.Queryer {
	return orm.NewWithNoInstance()(db.Engine).
		WithSession(db.session).
		Query(sqlStr, args...)
}

func (db *DB) Menus() *orm.Collection {
	return orm.New(func() interface{} {
		return &Menu{}
	})(db.Engine).WithSession(db.session)
}
