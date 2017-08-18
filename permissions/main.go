package permissions

import (
	"errors"
	"log"
	"time"

	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/web_ext"
)

func InitUser(lifecycle *web_ext.Lifecycle) func(userName string) web_ext.User {
	db := &DB{Engine: lifecycle.ModelEngine}
	permissionGroupCache := &GroupCache{}
	var lastErr concurrency.ErrorValue

	refresh := func() {
		lastErr.Set(permissionGroupCache.refresh(db))
	}
	permissionGroupCache.Init(5*time.Minute, refresh)
	refresh()

	var administrator, visitor int64

	return func(userName string) web_ext.User {
		if e := lastErr.Get(); e != nil {
			panic(e)
		}
		if administrator == 0 {
			var adminRole Role
			if e := db.Roles().Where(orm.Cond{"name": "administrator"}).One(&adminRole); e == nil {
				administrator = adminRole.ID
			} else {
				log.Println("[warn] role administrator isnot found -", e)
			}
		}
		if visitor == 0 {
			var visitorRole Role
			if e := db.Roles().Where(orm.Cond{"name": "visitor"}).One(&visitorRole); e == nil {
				visitor = visitorRole.ID
			} else {
				log.Println("[warn] role visitor isnot found -", e)
			}
		}

		var u = &user{db: db,
			lifecycle:            lifecycle,
			permissionGroupCache: permissionGroupCache,
			administrator:        administrator,
			visitor:              visitor}
		err := db.Users().Where(orm.Cond{"name": userName}).One(&u.u)
		if err != nil {
			if userName != "admin" {
				panic(errors.New("query user with name is " + userName + "fail: " + err.Error()))
			}
			u.u.Name = userName
			return u
		}

		rolesqlStr := "select * from " + db.Roles().Name() + " as roles " +
			" where exists (select * from " + db.UsersAndRoles().Name() + " as users_roles join " +
			db.Users().Name() + " as users on users_roles.user_id = users.id where users_roles.role_id = roles.id and users.name = ?)"
		err = db.Roles().Query(rolesqlStr, userName).All(&u.roles)
		if err != nil {
			log.Println("[permission] ", rolesqlStr, userName)
			panic(errors.New("query permissions and roles with user is " + userName + " fail: " + err.Error()))
		}

		if u.administrator != 0 {
			for _, role := range u.roles {
				if role.ID == u.administrator {
					return u
				}
			}
		}

		sqlStr := "select * from " + db.PermissionGroupsAndRoles().Name() + " as pg_role " +
			" where exists (select * from " + db.UsersAndRoles().Name() + " as users_roles join " +
			db.Users().Name() + " as users on users_roles.user_id = users.id where users_roles.role_id = pg_role.role_id and users.name = ?)"

		err = db.PermissionGroupsAndRoles().Query(sqlStr, userName).All(&u.permissionsAndRoles)
		if err != nil {
			log.Println("[permission] ", sqlStr, userName)
			panic(errors.New("query permissions and roles with user is " + userName + " fail: " + err.Error()))
		}

		// sqlStr := "select * from " + db.PermissionGroupAndRoles().Name() + "as pg " +
		// 	" where exists (select * from " + db.UserAndRole().Name() + " as uar join " +
		// 	db.User().Name() + " as user on uar.user_id = user.id where role.id = uar.role_id and user.name = ?)"

		// var roles []PermissionGroupAndRole
		// err = db.Role().Query(sqlStr, user).All(&roles)
		// if err != nil {
		// 	panic(errors.New("query roles with user is " + user + "fail: " + err.Error()))
		// }

		return u
	}
}

type user struct {
	db                   *DB
	lifecycle            *web_ext.Lifecycle
	u                    User
	roles                []Role
	permissionsAndRoles  []PermissionGroupAndRole
	permissionGroupCache *GroupCache

	administrator, visitor int64
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
	if u.Name() == "admin" {
		return true
	}
	if u.administrator != 0 {
		for _, role := range u.roles {
			if role.ID == u.administrator {
				return true
			}
		}
	}

	if u.visitor != 0 && web_ext.QUERY == op {
		for _, role := range u.roles {
			if role.ID == u.visitor {
				return true
			}
		}
	}

	for _, pr := range u.permissionsAndRoles {
		enableOperation := false
		switch op {
		case web_ext.CREATE:
			enableOperation = pr.CreateOperation
		case web_ext.DELETE:
			enableOperation = pr.DeleteOperation
		case web_ext.UPDATE:
			enableOperation = pr.UpdateOperation
		case web_ext.QUERY:
			enableOperation = pr.QueryOperation
			if pr.CreateOperation ||
				pr.DeleteOperation ||
				pr.UpdateOperation {
				enableOperation = true
			}
		default:
			panic(errors.New("Operation '" + op + "' is unknown"))
		}
		if !enableOperation {
			continue
		}

		if u.hasPermission(pr.GroupID, permissionID) {
			return true
		}
	}
	return false
}

func (u *user) hasPermission(groupID int64, permissionID string) bool {
	permissions := u.permissionGroupCache.Get(groupID)
	if permissions == nil {
		log.Println("[permissions] permission group with id is", groupID, "isn't found.")
		return false
	}
	for _, id := range permissions.PermissionIDs {
		if permissionID == id {
			return true
		}
	}

	for _, tag := range permissions.PermissionTags {
		permissionList, err := GetPermissionsByTag(tag)
		if err != nil {
			panic(err)
		}

		for _, permission := range permissionList {
			if permissionID == permission.ID {
				return true
			}
		}
	}

	if permissions.ParentID != 0 {
		return u.hasPermission(permissions.ParentID, permissionID)
	}
	return false
}
