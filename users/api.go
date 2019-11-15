package users

import (
	"context"

	"github.com/three-plus-three/modules/users/usermodels"
)

// Action
const (
	CREATE = usermodels.CREATE
	DELETE = usermodels.DELETE
	UPDATE = usermodels.UPDATE
	QUERY  = usermodels.QUERY

	DeletePermission = usermodels.DeletePermission
	UpdatePermission = usermodels.UpdatePermission
	CreatePermission = usermodels.CreatePermission
	QueryPermission  = usermodels.QueryPermission

	// UserAdmin admin 用户名
	UserAdmin = usermodels.UserAdmin

	// UserGuest guest 用户名
	UserGuest = usermodels.UserGuest

	// UserTPTNetwork tpt_nm 用户名
	UserTPTNetwork = usermodels.UserTPTNetwork

	// RoleSuper super 角色名
	RoleSuper = usermodels.RoleSuper

	// RoleAdministrator administrator 角色名
	RoleAdministrator = usermodels.RoleAdministrator

	// RoleVisitor visitor 角色名
	RoleVisitor = usermodels.RoleVisitor

	// RoleGuest guest 角色名
	RoleGuest = usermodels.RoleGuest
)

// Option 用户选项
type Option interface {
	apply()
}

// UserIncludeDisabled 禁用的用户也返回
func UserIncludeDisabled() Option {
	return userIncludeDisabled{}
}

// UserManager 用户管理
type UserManager interface {
	Users(ctx context.Context, opts ...Option) ([]User, error)
	Usergroups(ctx context.Context, opts ...Option) ([]Usergroup, error)

	UserByName(ctx context.Context, username string, opts ...Option) (User, error)
	UserByID(ctx context.Context, userID int64, opts ...Option) (User, error)

	UsergroupByName(ctx context.Context, username string, opts ...Option) (Usergroup, error)
	UsergroupByID(ctx context.Context, groupID int64, opts ...Option) (Usergroup, error)
}

// Usergroup 用户组信息
type Usergroup interface {
	ID() int64

	// 用户登录名
	Name() string

	// 用户成员
	Users(ctx context.Context, opts ...Option) ([]User, error)
}

// User 用户信息
type User interface {
	ID() int64

	// 用户登录名
	Name() string

	// 是不是有一个管理员角色
	HasAdminRole() bool

	// 是不是有一个 Guest 角色
	// HasGuestRole() bool

	// 呢称
	Nickname() string

	// Profile 是用于保存用户在界面上的一些个性化数据
	// WriteProfile 保存 profiles
	WriteProfile(key, value string) error

	// Profile 是用于保存用户在界面上的一些个性化数据
	// ReadProfile 读 profiles
	ReadProfile(key string) (string, error)

	// 用户扩展属性
	Data(key string) interface{}

	// 用户角色列表
	Roles() []string

	// 用户是否有指定的权限
	HasPermission(permissionName, op string) bool

	// 是不是有一个指定的角色
	HasRole(string) bool

	// 本用户是不是指定的用户组的成员
	IsMemberOf(int64) bool
}

type ReadCurrentUserFunc func(context.Context) (User, error)
