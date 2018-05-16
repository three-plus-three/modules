package web_ext

import (
	"github.com/go-xorm/xorm"
	"github.com/revel/revel"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/menus"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/sso/client/revel_sso"
)

// Action
const (
	CREATE = "create"
	DELETE = "delete"
	UPDATE = "update"
	QUERY  = "query"
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

// InitUser 初始化用户的回调函数
var InitUser = func(lifecycle *Lifecycle) func(userName string) User {
	return func(userName string) User {
		return &user{lifecycle: lifecycle, name: userName}
	}
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

type user struct {
	lifecycle *Lifecycle
	name      string
}

func (u *user) ID() int64 {
	return 1
}

func (u *user) Name() string {
	if u.name == "" {
		return "admin"
	}

	return u.name
}

func (u *user) Nickname() string {
	if u.name == "" {
		return "admin"
	}

	return u.name
}

func (u *user) WriteProfile(key, value string) error {
	return nil
}

func (u *user) ReadProfile(key string) (interface{}, error) {
	return nil, nil
}

func (u *user) Data(key string) interface{} {
	return nil
}

func (u *user) Roles() []string {
	return []string{"administrator"}
}

func (u *user) HasPermission(permissionName, op string) bool {
	return true
}

// Lifecycle 表示一个运行周期，它包含了所有业务相关的对象
type Lifecycle struct {
	concurrency.Base
	Env         *environment.Environment
	ModelEngine *xorm.Engine
	DataEngine  *xorm.Engine
	Variables   map[string]interface{}
	URLPrefix   string
	URLRoot     string

	ApplicationID      environment.ENV_PROXY_TYPE
	ApplicationContext string
	ApplicationRoot    string

	GetUser     func(userName string) User
	CurrentUser func(c *revel.Controller) User
	CheckUser   revel_sso.CheckFunc
	menuClient  menus.Client
	menuHook    func() ([]toolbox.Menu, error)
}

// Menus 返回所有菜单
func (lifecycle *Lifecycle) Menus() []toolbox.Menu {
	var menuList []toolbox.Menu
	var err error
	if lifecycle.menuHook != nil {
		menuList, err = lifecycle.menuHook()
	} else {
		menuList, err = lifecycle.menuClient.Read()
	}
	if err != nil {
		revel.ERROR.Println("\n错误:" + err.Error())
		panic(errors.Wrap(err, "获取菜单失败"))
	}
	return menuList
}

// NewLifecycle 创建一个生命周期
func NewLifecycle(env *environment.Environment, serviceID environment.ENV_PROXY_TYPE) (*Lifecycle, error) {
	dbDrv, dbURL := env.Db.Models.Url()
	modelEngine, err := xorm.NewEngine(dbDrv, dbURL)
	if err != nil {
		return nil, err
	}

	dataDrv, dataURL := env.Db.Data.Url()
	dataEngine, err := xorm.NewEngine(dataDrv, dataURL)
	if err != nil {
		return nil, err
	}

	return &Lifecycle{
		Env:           env,
		ModelEngine:   modelEngine,
		DataEngine:    dataEngine,
		ApplicationID: serviceID,
	}, nil
}
