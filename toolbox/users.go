package toolbox

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/three-plus-three/modules/as"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/users"
)

// Action
const (
	CREATE = users.CREATE
	DELETE = users.DELETE
	UPDATE = users.UPDATE
	QUERY  = users.QUERY

	DeletePermission = users.DeletePermission
	UpdatePermission = users.UpdatePermission
	CreatePermission = users.CreatePermission
	QueryPermission  = users.QueryPermission

	// UserAdmin admin 用户名
	UserAdmin = users.UserAdmin

	// UserGuest guest 用户名
	UserGuest = users.UserGuest

	// UserTPTNetwork tpt_nm 用户名
	UserTPTNetwork = users.UserTPTNetwork

	// RoleSuper super 角色名
	RoleSuper = users.RoleSuper

	// RoleAdministrator administrator 角色名
	RoleAdministrator = users.RoleAdministrator

	// RoleVisitor visitor 角色名
	RoleVisitor = users.RoleVisitor

	// RoleGuest guest 角色名
	RoleGuest = users.RoleGuest
)

// UserOption 用户选项
type UserOption = users.UserOption
type UserManager = users.UserManager
type User = users.User
type UserGroup = users.UserGroup

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

	readUserattr := func(u User, attr string, defaultValue ...string) string {
		value := u.Data(attr)
		if value == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return ""
		}

		if len(defaultValue) > 0 {
			return as.StringWithDefault(value, defaultValue[0])
		}
		return as.StringWithDefault(value, "")
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
		u, err := um.ByName(user)
		if err != nil {
			if errors.IsNotFound(err) {
				return false
			}
			panic(errors.Wrap(err, "load user with name is '"+user+"' fail"))
		}
		if u == nil {
			return false
		}
		return u.HasPermission(permissionName, op)
	}

	funcs["username"] = func(userID interface{}, defaultValue ...string) string {
		uid, err := as.Int64(userID)
		if err != nil {
			u, ok := userID.(User)
			if ok {
				return u.Nickname()
			}

			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' is invalid user identifier"))
		}

		if uid == 0 {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}

			return ""
		}

		u, err := um.ByID(uid, users.UserIncludeDisabled())
		if err != nil && !errors.IsNotFound(err) {
			panic(errors.Wrap(err, "load user with id is '"+fmt.Sprint(userID)+"' fail"))
		}

		if u == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' isnot found"))
		}

		return u.Nickname()
	}

	funcs["userattr"] = func(userID interface{}, attr string, defaultValue ...string) string {
		uid, err := as.Int64(userID)
		if err != nil {
			u, ok := userID.(User)
			if ok {
				return readUserattr(u, attr, defaultValue...)
			}

			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' is invalid user identifier"))
		}

		if uid == 0 {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			return ""
		}

		u, err := um.ByID(uid, users.UserIncludeDisabled())
		if err != nil && !errors.IsNotFound(err) {
			panic(errors.Wrap(err, "load user with id is '"+fmt.Sprint(userID)+"' fail"))
		}

		if u == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("user id '" + fmt.Sprint(userID) + "' isnot found"))
		}

		return readUserattr(u, attr, defaultValue...)
	}

	funcs["usernames"] = func(args ...interface{}) map[int64]string {
		if len(args) > 2 {
			panic(errors.New("bad usernames arguments - " + fmt.Sprint(args)))
		}

		var usergroup UserGroup
		var group int64
		var opts = []UserOption{}
		for idx, arg := range args {
			switch v := arg.(type) {
			case bool:
				opts = []UserOption{users.UserIncludeDisabled()}
			case int:
				group = int64(v)
			case int32:
				group = int64(v)
			case int64:
				group = v
			case uint:
				group = int64(v)
			case uint32:
				group = int64(v)
			case uint64:
				group = int64(v)
			case string:
				if s := strings.ToLower(v); s == "true" {
					opts = []UserOption{users.UserIncludeDisabled()}
					break
				} else if s == "false" {
					break
				}

				i64, err := strconv.ParseInt(v, 10, 64)
				if err != nil {
					panic(fmt.Errorf("bad usernames argument(%d) - %s", idx, arg))
				}
				group = i64
			case UserGroup:
				usergroup = v
			default:
				panic(fmt.Errorf("bad usernames argument(%d) - %s", idx, arg))
			}
		}

		var userlist []User
		if usergroup != nil {
			uList, err := usergroup.Users(opts...)
			if err != nil {
				panic(errors.Wrap(err, "load users of group("+usergroup.Name()+") fail"))
			}
			userlist = uList
		} else if group != 0 {
			var err error
			usergroup, err = um.GroupByID(group, opts...)
			if err != nil {
				panic(errors.Wrap(err, "load users of group("+strconv.FormatInt(group, 10)+") fail"))
			}
			userlist, err = usergroup.Users(opts...)
			if err != nil {
				panic(errors.Wrap(err, "load users of group("+usergroup.Name()+") fail"))
			}
		} else {
			uList, err := um.Users(opts...)
			if err != nil {
				panic(errors.Wrap(err, "load all users fail"))
			}
			userlist = uList
		}

		results := map[int64]string{}
		for _, u := range userlist {
			results[u.ID()] = u.Nickname()
		}
		return results
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

		u, err := um.GroupByID(uid)
		if err != nil {
			if errors.IsNotFound(err) {
				panic(errors.New("usergroup id '" + fmt.Sprint(groupID) + "' isnot found"))
			} else {
				panic(errors.Wrap(err, "load usergroup with id is '"+fmt.Sprint(groupID)+"' fail"))
			}
		}

		if u == nil {
			if len(defaultValue) > 0 {
				return defaultValue[0]
			}
			panic(errors.New("usergroup id '" + fmt.Sprint(groupID) + "' isnot found"))
		}

		return u.Name()
	}

	funcs["usergroupnames"] = func(includeDisabled ...bool) map[int64]string {
		var opts = []UserOption{}
		if len(includeDisabled) > 0 && includeDisabled[0] {
			opts = []UserOption{users.UserIncludeDisabled()}
		}
		ugList, err := um.Groups(opts...)
		if err != nil {
			panic(errors.Wrap(err, "load all users fail"))
		}

		results := map[int64]string{}
		for _, ug := range ugList {
			results[ug.ID()] = ug.Name()
		}
		return results
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

// UserProvider 一个读用户的扩展接口
type UserProvider struct {
	UM UserManager
}

func (up *UserProvider) Read(ctx, a interface{}) (interface{}, error) {
	users, err := up.UM.Users()
	if err != nil {
		return nil, err
	}

	var names = make([][2]string, 0, len(users))
	for _, u := range users {
		names = append(names, [2]string{strconv.FormatInt(u.ID(), 10), u.Nickname()})
	}
	return names, nil
}

// UsergroupProvider 一个读用户的扩展接口
type UsergroupProvider struct {
	UM UserManager
}

func (up *UsergroupProvider) Read(ctx, a interface{}) (interface{}, error) {
	usergroups, err := up.UM.Groups()
	if err != nil {
		return nil, err
	}

	var names = make([][2]string, 0, len(usergroups))
	for _, u := range usergroups {
		names = append(names, [2]string{strconv.FormatInt(u.ID(), 10), u.Name()})
	}
	return names, nil
}
