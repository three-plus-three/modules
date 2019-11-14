package users

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
func UserIncludeDisabled() UserOption {
	return userIncludeDisabled{}
}

// UserManager 用户管理
type UserManager interface {
	Users(opts ...UserOption) ([]User, error)
	Groups(opts ...UserOption) ([]UserGroup, error)

	ByName(username string, opts ...UserOption) (User, error)
	ByID(userID int64, opts ...UserOption) (User, error)

	GroupByName(username string, opts ...UserOption) (UserGroup, error)
	GroupByID(groupID int64, opts ...UserOption) (UserGroup, error)
}

// UserGroup 用户组信息
type UserGroup interface {
	ID() int64

	// 用户登录名
	Name() string

	// 用户成员
	Users(opts ...UserOption) ([]User, error)
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
