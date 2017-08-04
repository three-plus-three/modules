package permissions

import (
	"errors"

	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/web_ext"
)

func init() {
	web_ext.ReadUser = func(lifecycle *web_ext.Lifecycle, userName string) web_ext.User {
		db := &DB{Engine: lifecycle.ModelEngine}

		var u = &user{db: db, lifecycle: lifecycle}
		err := db.User().Where(orm.Cond{"name": userName}).One(&u.u)
		if err != nil {
			panic(errors.New("query user with name is " + userName + "fail: " + err.Error()))
		}

		// sqlStr := "select * from " + db.Role().Name() + "as role " +
		//   " where exists (select * from " + db.UserAndRole().Name() + " as uar join " +
		//   db.User().Name() + " as user on uar.user_id = user.id where role.id = uar.role_id and user.name = ?)"

		// var roles []Role
		// err := db.Role().Query(sqlStr, user).All(&roles)
		// if err != nil {
		//   panic(errors.New("query roles with user is " + user + "fail: " + err.Error()))
		// }

		// sqlStr = "select * from " + db.p().Name() + "as role " +
		// 	" where exists (select * from " + db.UserAndRole().Name() + " as uar join " +
		// 	db.User().Name() + " as user on uar.user_id = user.id where role.id = uar.role_id and user.name = ?)"

		// var roles []Role
		// err := db.Role().Query(sqlStr, user).All(&roles)
		// if err != nil {
		// 	panic(errors.New("query roles with user is " + user + "fail: " + err.Error()))
		// }

		return u
	}
}

type user struct {
	db        *DB
	lifecycle *web_ext.Lifecycle
	u         User
}

func (u *user) ID() int64 {
	return u.u.ID
}

func (u *user) Name() string {
	return u.u.Name
}

func (u *user) Data(key string) interface{} {
	switch key {
	case "id":
		return u.u.ID
	case "name":
		return u.u.Name
	case "description":
		return u.u.Description
	case "attributes":
		return u.u.Attributes
	case "source":
		return u.u.Source
	case "created_at":
		return u.u.CreatedAt
	case "updated_at":
		return u.u.UpdatedAt
	default:
		if u.u.Attributes != nil {
			return u.u.Attributes[key]
		}
	}
	return nil
}

func (u *user) HasPermission(permissionID, op string) bool {

	return true
}
