package web_ext

import (
	"github.com/go-xorm/xorm"
	"github.com/revel/revel"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/sso/client/revel_sso"
)

type User interface {
	ID() int

	Name() string

	Data(key string) interface{}
}

type user struct {
	lifecycle *Lifecycle
	name      string
}

func (u *user) ID() int {
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

// Lifecycle 表示一个运行周期，它包含了所有业务相关的对象
type Lifecycle struct {
	environment.Base
	Env         *environment.Environment
	ModelEngine *xorm.Engine
	DataEngine  *xorm.Engine
	Variables   map[string]interface{}
	URLPrefix   string
	URLRoot     string

	ApplicationContext string
	ApplicationRoot    string

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
