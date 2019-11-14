package web_ext

import (
	"github.com/revel/revel"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/menus"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/sso/client/revel_sso"
	"xorm.io/xorm"
)

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

	UserManager toolbox.UserManager
	GetUser     func(userName string, opts ...toolbox.UserOption) toolbox.User
	CurrentUser func(c *revel.Controller) toolbox.User
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
	modelEngine.ShowSQL(true)

	dataDrv, dataURL := env.Db.Data.Url()
	dataEngine, err := xorm.NewEngine(dataDrv, dataURL)
	if err != nil {
		return nil, err
	}
	dataEngine.ShowSQL(true)

	return &Lifecycle{
		Env:           env,
		ModelEngine:   modelEngine,
		DataEngine:    dataEngine,
		ApplicationID: serviceID,
	}, nil
}

type userManager struct {
	lifecycle *Lifecycle
}

func (um *userManager) Groups(opts ...toolbox.UserOption) ([]toolbox.UserGroup, error) {
	return []toolbox.UserGroup{}, nil
}

func (um *userManager) Users(opts ...toolbox.UserOption) ([]toolbox.User, error) {
	return []toolbox.User{}, nil
}

func (um *userManager) ByName(username string, opts ...toolbox.UserOption) (toolbox.User, error) {
	return &user{lifecycle: um.lifecycle, name: username}, nil
}

func (um *userManager) ByID(userID int64, opts ...toolbox.UserOption) (toolbox.User, error) {
	return nil, errors.NotFound(userID, "user")
}

func (um *userManager) GroupByName(groupname string, opts ...toolbox.UserOption) (toolbox.UserGroup, error) {
	return &usergroup{lifecycle: um.lifecycle, name: groupname}, nil
}

func (um *userManager) GroupByID(groupID int64, opts ...toolbox.UserOption) (toolbox.UserGroup, error) {
	return nil, errors.NotFound(groupID, "usergroup")
}

type usergroup struct {
	lifecycle *Lifecycle
	name      string
}

func (ug *usergroup) ID() int64 {
	return 1
}

func (ug *usergroup) Name() string {
	if ug.name == "" {
		return toolbox.UserAdmin
	}
	return ug.name
}

func (ug *usergroup) Users(opts ...toolbox.UserOption) ([]toolbox.User, error) {
	return nil, nil
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
		return toolbox.UserAdmin
	}

	return u.name
}

func (u *user) Nickname() string {
	if u.name == "" {
		return toolbox.UserAdmin
	}

	return u.name
}

func (u *user) HasAdminRole() bool {
	return true
}

func (u *user) HasGuestRole() bool {
	return true
}

func (u *user) HasRole(role string) bool {
	if role == toolbox.RoleAdministrator {
		return true
	}
	return false
}

func (u *user) IsMemberOf(group int64) bool {
	return false
}

func (u *user) WriteProfile(key, value string) error {
	return nil
}

func (u *user) ReadProfile(key string) (string, error) {
	return "", nil
}

func (u *user) Data(key string) interface{} {
	return nil
}

func (u *user) Roles() []string {
	return []string{toolbox.RoleAdministrator}
}

func (u *user) HasPermission(permissionName, op string) bool {
	return true
}

// InitUser 初始化用户的回调函数
var InitUser = func(lifecycle *Lifecycle) toolbox.UserManager {
	return &userManager{lifecycle: lifecycle}
}
