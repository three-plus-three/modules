package menus

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/hub"
	hub_engine "github.com/three-plus-three/modules/hub/engine"
	"github.com/three-plus-three/modules/toolbox"
)

//go:generate genny -pkg=menus -in=../weaver/client.go -out=client-gen.go gen "ValueType=[]toolbox.Menu"
//go:generate genny -pkg=menus -in=../weaver/server.go -out=server-gen.go gen "WeaveType=[]toolbox.Menu"
//go:generate genny -pkg=menus -in=../concurrency/generic/cached.go -out=cached-gen.go gen "ValueType=[]toolbox.Menu"

// ErrAlreadyClosed  server is closed
var ErrAlreadyClosed = errors.New("server is closed")

func NewWeaver(logger *log.Logger, env *environment.Environment, core *hub_engine.Core, layout Layout, layouts map[string]Layout) (Weaver, error) {
	weaver := &menuWeaver{Logger: logger, env: env, core: core, layout: layout, layouts: layouts}
	if err := weaver.Init(); err != nil {
		return nil, err
	}

	return weaver, nil
}

// // Menu 数据库中的一个菜单项
// type Menu struct {
// 	AutoID      int64  `json:"auto_id" xorm:"id pk autoincr"`
// 	Application string `json:"application" xorm:"application"`

// 	ParentID int64 `json:"parent_id,omitempty" xorm:"parent_id"`
// 	Seqence  int64 `json:"seqence,omitempty" xorm:"seqence"`

// 	toolbox.Menu `xorm:"extends"`

// 	Container []*Menu
// }

// Layout 菜单布避生成器
type Layout interface {
	Stats() interface{}
	Generate(map[string][]toolbox.Menu) ([]toolbox.Menu, error)
}

type menuWeaver struct {
	Logger *log.Logger
	env    *environment.Environment
	core   *hub_engine.Core
	//db      *DB
	layout  Layout
	layouts map[string]Layout

	mu               sync.RWMutex
	byApplications   map[string][]toolbox.Menu
	menuList         []toolbox.Menu
	menuListByLayout map[string][]toolbox.Menu
}

func (weaver *menuWeaver) Stats() interface{} {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	apps := map[string]interface{}{}
	for name, app := range weaver.byApplications {
		// apps[name] = toMenuTree(app)
		apps[name] = app
	}

	layouts := map[string]interface{}{}
	for k, v := range weaver.layouts {
		layouts[k] = v.Stats()
	}

	return map[string]interface{}{
		"applications": apps,
		"layout":       layouts,
		"menuList":     weaver.menuList,
	}
}

func (weaver *menuWeaver) generate() ([]toolbox.Menu, error) {
	// byApps := map[string][]toolbox.Menu{}
	// for name, app := range weaver.byApplications {
	// 	byApps[name] = toMenuTree(app)
	// }
	return weaver.layout.Generate(weaver.byApplications)
}

func (weaver *menuWeaver) Init() error {

	// var allList []Menu
	// err := weaver.db.Menus().Where().All(&allList)
	// if err != nil {
	// 	return errors.New("LoadFromDB: " + err.Error())
	// }

	// byApplications := map[string]map[string]*Menu{}
	// for idx, menu := range allList {
	// 	newInGroup := byApplications[menu.Application]
	// 	if newInGroup == nil {
	// 		newInGroup = map[string]*Menu{}
	// 	}
	// 	newInGroup[menu.UID] = &allList[idx]
	// 	byApplications[menu.Application] = newInGroup
	// }

	byApplications := map[string][]toolbox.Menu{}
	filename := weaver.env.Fs.FromTMP("app_menus.json")
	in, err := os.Open(filename)
	if err != nil {
		weaver.Logger.Println("LoadFromDB:", err)
	} else {
		defer in.Close()

		err = json.NewDecoder(in).Decode(&byApplications)
		if err != nil {
			weaver.Logger.Println("LoadFromDB:", err)
		}
	}

	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.byApplications = byApplications
	weaver.menuListByLayout = nil
	weaver.menuList, err = weaver.generate()
	if err != nil {
		weaver.Logger.Println("LoadFromDB:", err)
	}
	return nil
}

func (weaver *menuWeaver) Update(app string, menuList []toolbox.Menu) error {
	weaver.mu.RLock()
	oldList := weaver.byApplications[app]
	weaver.mu.RUnlock()

	if len(menuList) == 0 && len(oldList) == 0 {
		return nil
	}

	var err error
	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	if weaver.byApplications == nil {
		weaver.byApplications = map[string][]toolbox.Menu{}
	}
	weaver.byApplications[app] = menuList
	weaver.menuListByLayout = nil
	weaver.menuList, err = weaver.generate()
	if err != nil {
		return errors.New("Generate: " + err.Error())
	}

	weaver.core.CreateTopicIfNotExists("menus.changed").
		Send(hub.Message([]byte(strconv.Itoa(len(menuList)))))

	filename := weaver.env.Fs.FromTMP("app_menus.json")
	if err = os.MkdirAll(filepath.Dir(filename), 0777); err != nil {
		weaver.Logger.Println("update menu list in app " + app + " to file fail, " + err.Error())
		return nil
	}

	out, err := os.Create(filename)
	if err != nil {
		weaver.Logger.Println("update menu list in app " + app + " to file fail, " + err.Error())
		return nil
	}
	defer out.Close()

	err = json.NewEncoder(out).Encode(weaver.byApplications)
	if err != nil {
		weaver.Logger.Println("update menu list in app " + app + " to file fail, " + err.Error())
	}
	return nil
}

func (weaver *menuWeaver) Generate(ctx string) ([]toolbox.Menu, error) {
	return weaver.read(ctx)
}

func (weaver *menuWeaver) read(layoutName string, args ...interface{}) ([]toolbox.Menu, error) {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	if len(args) == 0 {
		return weaver.menuList, nil
	}
	if len(args) != 1 {
		return nil, errors.New("arguments is too many")
	}

	if weaver.menuListByLayout != nil {
		byLayout, ok := weaver.menuListByLayout[layoutName]
		if ok {
			return byLayout, nil
		}
	}

	layout, ok := weaver.layouts[layoutName]
	if ok && layout != nil {
		return func() ([]toolbox.Menu, error) {
			weaver.mu.RUnlock()
			weaver.mu.Lock()

			defer func() {
				weaver.mu.Unlock()
				weaver.mu.RLock()
			}()

			// byApps := map[string][]toolbox.Menu{}
			// for name, app := range weaver.byApplications {
			// 	byApps[name] = toMenuTree(app)
			// }
			menuList, err := layout.Generate(weaver.byApplications)
			if err != nil {
				weaver.menuListByLayout[layoutName] = menuList
			}
			return menuList, err
		}()
	}

	return weaver.menuList, nil
}

func isSubset(allItems, subset []toolbox.Menu) bool {
	return IsSubset(allItems, subset)
}

func IsSubset(allItems, subset []toolbox.Menu) bool {
	for _, item := range subset {
		raw := SearchMenuInTree(allItems, item.UID)
		if raw == nil || !isSameMenu(item, *raw) {
			return false
		}
	}
	return true
}

// func toMenuTree(menuList map[string]*Menu) []toolbox.Menu {
// 	byID := map[int64]*Menu{}
// 	for _, menu := range menuList {
// 		if menu == nil {
// 			continue
// 		}

// 		byID[menu.AutoID] = menu
// 		if menu.Container != nil {
// 			menu.Container = menu.Container[:0]
// 		}
// 	}

// 	topMenuList := make([]*Menu, 0, 16)
// 	for _, menu := range menuList {
// 		if menu.ParentID == 0 {
// 			topMenuList = append(topMenuList, menu)
// 			continue
// 		}

// 		parent := byID[menu.ParentID]
// 		if parent == nil {
// 			panic(fmt.Errorf("菜单(%d:%s) 找不到父节点 %d", menu.AutoID, menu.UID, menu.ParentID))
// 		}
// 		parent.Container = append(parent.Container, menu)
// 	}

// 	sortMenuList(topMenuList)
// 	results := copyToMenuList(topMenuList)
// 	results = clearDividerFromList(results)
// 	return results
// }

// func sortMenuList(list []*Menu) {
// 	if len(list) == 0 {
// 		return
// 	}

// 	sort.Slice(list, func(a, b int) bool {
// 		return list[a].Seqence < list[b].Seqence
// 	})

// 	for _, menu := range list {
// 		sortMenuList(menu.Container)
// 	}
// }

// func copyToMenuList(list []*Menu) []toolbox.Menu {
// 	if len(list) == 0 {
// 		return nil
// 	}

// 	results := make([]toolbox.Menu, 0, len(list))
// 	for _, menu := range list {
// 		menu.Children = copyToMenuList(menu.Container)
// 		results = append(results, menu.Menu)
// 	}
// 	return results
// }
