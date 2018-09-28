package toolbox

import (
	"fmt"

	"github.com/three-plus-three/modules/as"
	"github.com/three-plus-three/modules/errors"
)

// Action
const (
	CREATE = "create"
	DELETE = "delete"
	UPDATE = "update"
	QUERY  = "query"

	DeletePermission = DELETE
	UpdatePermission = UPDATE
	CreatePermission = CREATE
	QueryPermission  = QUERY
)

// UserAdmin admin 用户名
const UserAdmin = "admin"

// UserGuest guest 用户名
const UserGuest = "guest"

// UserTPTNetwork tpt_nm 用户名
const UserTPTNetwork = "tpt_nm"

// RoleSuper super 角色名
const RoleSuper = "super"

// RoleAdministrator administrator 角色名
const RoleAdministrator = "administrator"

// RoleVisitor visitor 角色名
const RoleVisitor = "visitor"

// RoleGuest guest 角色名
const RoleGuest = "guest"

// UserOption 用户选项
type UserOption interface {
	apply()
}

// UserIncludeDisabled 禁用的用户也返回
type UserIncludeDisabled struct{}

func (u UserIncludeDisabled) apply() {}

// UserManager 用户管理
type UserManager interface {
	ByName(username string, opts ...UserOption) User
	ByID(userID int64, opts ...UserOption) User
	GroupByID(groupID int64, opts ...UserOption) UserGroup
}

// UserGroup 用户组信息
type UserGroup interface {
	ID() int64

	// 用户登录名
	Name() string
}

// User 用户信息
type User interface {
	ID() int64

	// 用户登录名
	Name() string

	// 呢称
	Nickname() string

	// Profile 是用于保存用户在界面上的一些个性化数据
	// WriteProfile 保存 profiles
	WriteProfile(key, value string) error

	// Profile 是用于保存用户在界面上的一些个性化数据
	// ReadProfile 读 profiles
	ReadProfile(key string) (interface{}, error)

	// 用户扩展属性
	Data(key string) interface{}

	// 用户角色列表
	Roles() []string

	// 用户是否有指定的权限
	HasPermission(permissionName, op string) bool
}

type CurrentUserFunc func(ctx map[string]interface{}) (User, error)

func CurrentUserHasPermission(currentUser CurrentUserFunc, ctx map[string]interface{}, permissionName string, opList []string) bool {
	u, err := currentUser(ctx)
	if err != nil {
		panic(err)
	}

	if u == nil {
		return false
	}

	for _, op := range opList {
		if u.HasPermission(permissionName, op) {
			return true
		}
	}
	return false
}

func InitUserFuncs(um UserManager, currentUser CurrentUserFunc, funcs map[string]interface{}) {
	if um == nil {
		panic("argument userManager is nil")
	}

	if currentUser == nil {
		currentUser = CurrentUserFunc(func(ctx map[string]interface{}) (User, error) {
			o := ctx["currentUser"]
			if o == nil {
				return nil, nil
			}

			u, ok := o.(User)
			if !ok {
				return nil, nil
			}
			return u, nil
		})
	}

	funcs["current_user_has_permission"] = func(ctx map[string]interface{}, permissionName string, op ...string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, op)
	}
	funcs["current_user_has_new_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, []string{CREATE})
	}
	funcs["current_user_has_del_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, []string{DELETE})
	}
	funcs["current_user_has_edit_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, []string{UPDATE})
	}
	funcs["current_user_has_write_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, []string{CREATE, DELETE, UPDATE})
	}
	funcs["current_user_has_query_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(currentUser, ctx, permissionName, []string{QUERY})
	}
	funcs["current_user_has_menu"] = func(ctx map[string]interface{}, menu interface{}) bool {
		var menuItem *Menu
		switch m := menu.(type) {
		case *Menu:
			menuItem = m
		case Menu:
			menuItem = &m
		default:
			panic(fmt.Errorf("unknown menuItem -- %T - %v", menu, menu))
		}

		if menuItem.Title == MenuDivider {
			return true
		}

		if menuItem.Permission == "" && menuItem.UID == "" {
			return true
		}

		u, err := currentUser(ctx)
		if err != nil {
			panic(err)
		}

		if u == nil {
			return false
		}
		return hasMenu(ctx, u, menuItem)
	}

	funcs["user_has_permission"] = func(ctx map[string]interface{}, user, permissionName, op string) bool {
		u := um.ByName(user)
		if u == nil {
			return false
		}
		return u.HasPermission(permissionName, op)
	}

	funcs["username"] = func(userID interface{}, defaultValue ...string) string {
		uid, err := as.Int64(userID)
		if err != nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' is invalid user identifier"))
		}

		if userID == 0 {
			return ""
		}

		u := um.ByID(uid, UserIncludeDisabled{})
		if u == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' isnot found"))
		}

		return u.Nickname()
	}

	funcs["usergroupname"] = func(groupID interface{}, defaultValue ...string) string {
		uid, err := as.Int64(groupID)
		if err != nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(groupID) + "' is invalid user identifier"))
		}

		if groupID == 0 {
			return ""
		}

		u := um.GroupByID(uid)
		if u == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(groupID) + "' isnot found"))
		}

		return u.Name()
	}
}

func hasMenu(ctx map[string]interface{}, u User, item *Menu) bool {
	permissionID := item.Permission
	if permissionID == "" {
		permissionID = item.UID
	}

	if u.HasPermission(permissionID, QUERY) {
		return true
	}

	for _, child := range item.Children {
		if hasMenu(ctx, u, &child) {
			return true
		}
	}
	return false
}
