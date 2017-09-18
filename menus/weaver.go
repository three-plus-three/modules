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

func NewWeaver(core *hub_engine.Core, db *DB) (Weaver, error) {
	weaver := &menuWeaver{core: core, db: db}
	if err := weaver.LoadFromDB(); err != nil {
		return nil, err
	}

	return weaver, nil
}

// Menu 数据库中的一个菜单项
type Menu struct {
	ID       int64  `json:"id" xorm:"id pk autoincr"`
	ParentID int64  `json:"parent_id,omitempty" xorm:"parent_id"`
	Group    string `json:"group" xorm:"group"`
	Seqence  int64  `json:"seqence,omitempty" xorm:"seqence"`

	toolbox.Menu `xorm:"extends"`
}

type menuWeaver struct {
	core *hub_engine.Core
	db   *DB

	mu       sync.RWMutex
	menuList []toolbox.Menu
	byGroups map[string]map[string]*Menu
}

func upsertMenuList(db *DB, group string, parentID int64, menuList []toolbox.Menu, oldInGroup map[string]*Menu) (map[string]*Menu, error) {
	newInGroup := map[string]*Menu{}
	for _, menuItem := range menuList {

		old, ok := oldInGroup[menuItem.Name]
		if ok {
			delete(oldInGroup, menuItem.Name)
		} else {
			old = &Menu{}

			err := db.Menus().Where(orm.Cond{"group": group, "name": menuItem.Name}).One(old)
			if err != nil {
				if orm.ErrNotFound != err {
					return nil, err
				}
				old.ParentID = parentID
				old.ID = 0
			}
		}

		old.Group = group
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

	for name := range oldInGroup {
		_, err := db.Menus().Where(orm.Cond{"group": group, "name": name}).Delete()
		if err != nil {
			return nil, err
		}
	}
	return newInGroup, nil
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

		newInGroup := byGroups[menu.Group]
		if newInGroup == nil {
			newInGroup = map[string]*Menu{}
		}
		newInGroup[menu.Name] = &allList[idx]
		byGroups[menu.Group] = newInGroup
	}

	menuList := generateMenuTree(0, allList)

	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.byGroups = byGroups
	weaver.menuList = menuList
	return nil
}

func (weaver *menuWeaver) Update(group string, menuList []toolbox.Menu) error {
	newInGroup, err := func() (map[string]*Menu, error) {
		tx, err := weaver.db.Begin()
		defer util.CloseWith(tx)
		inGroup := map[string]*Menu{}

		weaver.mu.RLock()
		for k, v := range weaver.byGroups[group] {
			inGroup[k] = v
		}
		weaver.mu.RUnlock()

		newList, err := upsertMenuList(tx, group, 0, menuList, inGroup)
		if err != nil {
			return nil, err
		}

		return newList, tx.Commit()
	}()

	if err != nil {
		return errors.New("update menu list in group \"" + group + "\" to db fail, " + err.Error())
	}

	weaver.mu.Lock()
	defer weaver.mu.Unlock()

	oldInGroup := weaver.byGroups[group]
	weaver.byGroups[group] = newInGroup
	weaver.menuList = toolbox.MergeMenus(weaver.menuList, menuList)

	for name := range oldInGroup {
		if _, ok := newInGroup[name]; !ok {
			toolbox.Remove(weaver.menuList, name)
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

func isSame(newItems, oldItems []toolbox.Menu) bool {
	return toolbox.IsSameMenuArray(newItems, oldItems)
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
