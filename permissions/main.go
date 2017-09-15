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

	var adminRole Role
	if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleAdministrator}).One(&adminRole); e != nil {
		adminRole.Name = web_ext.RoleAdministrator
		log.Println("[warn] role administrator isnot found -", e)
	}

	var visitorRole Role
	if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleVisitor}).One(&visitorRole); e != nil {
		visitorRole.Name = web_ext.RoleVisitor
		log.Println("[warn] role visitor isnot found -", e)
	}

	var guestRole Role
	if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleGuest}).One(&guestRole); e != nil {
		guestRole.Name = web_ext.RoleGuest
		log.Println("[warn] role visitor isnot found -", e)
	}

	return func(userName string) web_ext.User {
		if e := lastErr.Get(); e != nil {
			panic(e)
		}
		if adminRole.ID == 0 {
			if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleAdministrator}).One(&adminRole); e != nil {
				log.Println("[warn] role administrator isnot found -", e)
			}
		}
		if visitorRole.ID == 0 {
			if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleVisitor}).One(&visitorRole); e != nil {
				log.Println("[warn] role visitor isnot found -", e)
			}
		}
		if guestRole.ID == 0 {
			if e := db.Roles().Where(orm.Cond{"name": web_ext.RoleGuest}).One(&guestRole); e != nil {
				log.Println("[warn] role guest isnot found -", e)
			}
		}

		var u = &user{db: db,
			lifecycle:            lifecycle,
			permissionGroupCache: permissionGroupCache,
			administrator:        adminRole.ID,
			visitor:              visitorRole.ID}
		err := db.Users().Where(orm.Cond{"name": userName}).One(&u.u)
		if err != nil {
			switch userName {
			case web_ext.UserAdmin:
				u.u.Name = userName
				u.roleNames = []string{web_ext.RoleAdministrator}
				u.roles = []Role{adminRole}
				return u
			case web_ext.UserGuest:
				u.u.Name = userName
				u.roleNames = []string{web_ext.RoleGuest}
				u.roles = []Role{guestRole}
				return u
			default:
				panic(errors.New("query user with name is " + userName + "fail: " + err.Error()))
			}
		}

		cond := orm.Cond{"exists (select * from " + db.UsersAndRoles().Name() + " as users_roles join " +
			db.Users().Name() + " as users on users_roles.user_id = users.id " +
			" where users_roles.role_id = " + db.Roles().Name() + ".id and users.name = ?)": userName}
		err = db.Roles().Where(cond).
			All(&u.roles)
		if err != nil {
			log.Println("[permission] ", cond)
			panic(errors.New("query permissions and roles with user is " + userName + " fail: " + err.Error()))
		}

		if u.administrator != 0 {
			for _, role := range u.roles {
				if role.ID == u.administrator {
					u.Roles() // 缓存 roleNames
					return u
				}
			}
		}

		if u.u.Name == web_ext.UserAdmin {
			u.u.Name = userName
			u.roles = append(u.roles, adminRole)
			u.Roles() // 缓存 roleNames
			return u
		}

		pgRoleCond := orm.Cond{"exists (select * from " + db.UsersAndRoles().Name() + " as users_roles join " +
			db.Users().Name() + " as users on users_roles.user_id = users.id " +
			"where users_roles.role_id = " + db.PermissionGroupsAndRoles().Name() + ".role_id and users.name = ?)": userName}
		err = db.PermissionGroupsAndRoles().Where(pgRoleCond).All(&u.permissionsAndRoles)
		if err != nil {
			log.Println("[permission] ", pgRoleCond)
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

		u.Roles() // 缓存 roleNames
		return u
	}
}

type user struct {
	db                   *DB
	lifecycle            *web_ext.Lifecycle
	u                    User
	roles                []Role
	roleNames            []string
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

func (u *user) Nickname() string {
	return u.u.Name
}

func (u *user) Roles() []string {
	if len(u.roleNames) != 0 {
		return u.roleNames
	}
	if len(u.roles) == 0 {
		return nil
	}

	roleNames := make([]string, 0, len(u.roles))
	for idx := range u.roles {
		roleNames = append(roleNames, u.roles[idx].Name)
	}

	u.roleNames = roleNames
	return u.roleNames
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
	if u.Name() == web_ext.UserAdmin {
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
	group := u.permissionGroupCache.Get(groupID)
	if group == nil {
		log.Println("[permissions] permission group with id is", groupID, "isn't found.")
		return false
	}
	return u.hasPermissionInGroup(group, permissionID)
}

func (u *user) hasPermissionInGroup(group *Permissions, permissionID string) bool {
	// 在本组中查找是不是有这个权限
	for _, id := range group.PermissionIDs {
		if permissionID == id {
			return true
		}
	}

	// 在本组中查找是不是有标签含有这个权限
	for _, tag := range group.PermissionTags {
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

	// 在子组中查找这个权限
	children := u.permissionGroupCache.GetChildren(group.ID)
	if len(children) != 0 {
		for _, child := range children {
			if u.hasPermissionInGroup(child, permissionID) {
				return true
			}
		}
	}
	return false
}
