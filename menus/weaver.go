package menus

import (
	"errors"
	"strconv"
	"sync"

	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/hub"
	hub_engine "github.com/three-plus-three/modules/hub/engine"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/util"
)

//go:generate genny -pkg=menus -in=../weaver/client.go -out=client-gen.go gen "ValueType=[]toolbox.Menu"
//go:generate genny -pkg=menus -in=../weaver/server.go -out=server-gen.go gen "WeaveType=[]toolbox.Menu"

// ErrAlreadyClosed  server is closed
var ErrAlreadyClosed = errors.New("server is closed")

func NewWeaver(core *hub_engine.Core, db *DB) (Weaver, error) {
	weaver := &menuWeaver{core: core, db: db}
	if err := weaver.LoadFromDB(); err != nil {
		return nil, err
	}

	return weaver, nil
}

// Menu 数据库中的一个菜单项
type Menu struct {
	ID          int64  `json:"id" xorm:"id pk autoincr"`
	ParentID    int64  `json:"parent_id,omitempty" xorm:"parent_id"`
	Application string `json:"application" xorm:"application"`
	Seqence     int64  `json:"seqence,omitempty" xorm:"seqence"`

	toolbox.Menu `xorm:"extends"`
}

type menuWeaver struct {
	core *hub_engine.Core
	db   *DB

	mu       sync.RWMutex
	menuList []toolbox.Menu
	byGroups map[string]map[string]*Menu
}

func (weaver *menuWeaver) LoadFromDB() error {
	var allList []Menu
	err := weaver.db.Menus().Where().All(allList)
	if err != nil {
		return err
	}

	byID := map[int64]*Menu{}
	byGroups := map[string]map[string]*Menu{}
	for idx, menu := range allList {
		byID[menu.ID] = &allList[idx]

		newInGroup := byGroups[menu.Application]
		if newInGroup == nil {
			newInGroup = map[string]*Menu{}
		}
		newInGroup[menu.Name] = &allList[idx]
		byGroups[menu.Application] = newInGroup
	}

	menuList := generateMenuTree(0, allList)

	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.byGroups = byGroups
	weaver.menuList = menuList
	return nil
}

func upsertMenuList(db *DB, app string, parentID int64, menuList []toolbox.Menu, oldInGroup map[string]*Menu) (map[string]*Menu, error) {
	newInGroup := map[string]*Menu{}
	for _, menuItem := range menuList {
		var old *Menu
		var ok bool

		if oldInGroup != nil {
			old, ok = oldInGroup[menuItem.Name]
			if ok {
				delete(oldInGroup, menuItem.Name)
			}
		}

		if !ok {
			old = &Menu{}

			err := db.Menus().Where(orm.Cond{"application": app, "name": menuItem.Name}).One(old)
			if err != nil {
				if orm.ErrNotFound != err {
					return nil, err
				}
				old.ParentID = parentID
				old.ID = 0
			}
		}

		old.Application = app
		toolbox.MergeMenuWithNoChildren(&old.Menu, &menuItem)

		var err error
		if old.ID == 0 {
			var id interface{}
			old.ParentID = parentID
			old.Seqence = 0
			id, err = db.Menus().Nullable("parent_id").Insert(old)
			old.ID = id.(int64)
		} else {
			err = db.Menus().ID(old.ID).Update(old)
		}

		if err != nil {
			return nil, err
		}

		newInGroup[menuItem.Name] = old
	}

	if oldInGroup != nil {
		for name := range oldInGroup {
			_, err := db.Menus().Where(orm.Cond{"application": app, "name": name}).Delete()
			if err != nil {
				return nil, err
			}
		}
	}
	return newInGroup, nil
}

func (weaver *menuWeaver) Update(app string, menuList []toolbox.Menu) error {
	weaver.mu.RLock()
	oldList := weaver.byGroups[app]
	weaver.mu.RUnlock()

	if len(menuList) == 0 && len(oldList) == 0 {
		return nil
	}

	newInGroup, err := func(inGroup map[string]*Menu) (map[string]*Menu, error) {
		tx, err := weaver.db.Begin()
		defer util.CloseWith(tx)

		newList, err := upsertMenuList(tx, app, 0, menuList, inGroup)
		if err != nil {
			return nil, err
		}

		return newList, tx.Commit()
	}(oldList)

	if err != nil {
		return errors.New("update menu list in app \"" + app + "\" to db fail, " + err.Error())
	}

	weaver.mu.Lock()
	defer weaver.mu.Unlock()

	oldInGroup := weaver.byGroups[app]
	weaver.byGroups[app] = newInGroup
	weaver.menuList = toolbox.MergeMenus(weaver.menuList, menuList)

	if oldInGroup != nil {
		for name := range oldInGroup {
			if _, ok := newInGroup[name]; !ok {
				toolbox.Remove(weaver.menuList, name)
			}
		}
	}

	weaver.core.CreateTopicIfNotExists("menus.changed").
		Send(hub.Message([]byte(strconv.Itoa(len(menuList)))))
	return nil
}

func (weaver *menuWeaver) Generate() ([]toolbox.Menu, error) {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	return weaver.menuList, nil
}

func isSubset(allItems, subset []toolbox.Menu) bool {
	for _, item := range subset {
		raw := toolbox.SearchMenuInTree(allItems, item.Name)
		if raw == nil || !toolbox.IsSameMenu(item, *raw) {
			return false
		}
	}
	return true
}

func generateMenuTree(parentID int64, byID []Menu) []toolbox.Menu {
	var results []toolbox.Menu
	for _, menu := range byID {
		if menu.ParentID == parentID {
			menu.Menu.Children = generateMenuTree(menu.ID, byID)
			results = append(results, menu.Menu)
		}
	}
	return results
}
