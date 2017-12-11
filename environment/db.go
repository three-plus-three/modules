package environment

import (
	"errors"
	"fmt"
)

// DbConfig 数据库配置
type DbConfig struct {
	DbType   string
	Address  string
	Port     string
	Schema   string
	Username string
	Password string
}

func (db *DbConfig) Host() string {
	if "" != db.Port && "0" != db.Port {
		return db.Address + ":" + db.Port
	}
	switch db.DbType {
	case "postgresql":
		return db.Address + ":35432"
	default:
		panic(errors.New("unknown db type - " + db.DbType))
	}
}

func (db *DbConfig) dbUrl() (string, string, error) {
	switch db.DbType {
	case "postgresql":
		if db.Port == "" {
			db.Port = "5432"
		}
		return "postgres", fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable",
			db.Address, db.Port, db.Schema, db.Username, db.Password), nil
	case "mysql":
		if db.Port == "" {
			db.Port = "3306"
		}
		return "mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?autocommit=true&parseTime=true",
			db.Username, db.Password, db.Address, db.Port, db.Schema), nil
	case "odbc_with_mssql":
		return "odbc_with_mssql", fmt.Sprintf("dsn=%s;uid=%s;pwd=%s",
			db.Schema, db.Username, db.Password), nil
	default:
		return "", "", errors.New("unknown db type - " + db.DbType)
	}
}

func (db *DbConfig) Url() (string, string) {
	dbDrv, dbUrl, err := db.dbUrl()
	if err != nil {
		panic(errors.New("unknown db type - " + db.DbType))
	}
	return dbDrv, dbUrl
}
