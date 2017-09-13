package web_ext

import (
	"github.com/go-xorm/xorm"
	"github.com/revel/revel"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/sso/client/revel_sso"
)

const CREATE = "create"
const DELETE = "delete"
const UPDATE = "update"
const QUERY = "query"

// UserAdmin admin 用户名
const UserAdmin = "admin"

// UserGuest guest 用户名
const UserGuest = "guest"

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

	Name() string

	Data(key string) interface{}

	Roles() []string

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

	ApplicationContext string
	ApplicationRoot    string

	GetUser     func(userName string) User
	CurrentUser func(c *revel.Controller) User
	CheckUser   revel_sso.CheckFunc
	MenuList    []toolbox.Menu
}

// NewLifecycle 创建一个生命周期
func NewLifecycle(env *environment.Environment) (*Lifecycle, error) {
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
		Env:         env,
		ModelEngine: modelEngine,
		DataEngine:  dataEngine,
	}, nil
}
