package menus

import (
	"bytes"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/errors"
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
	Application string `json:"application" xorm:"application"`

	ParentID int64 `json:"parent_id,omitempty" xorm:"parent_id"`
	Seqence  int64 `json:"seqence,omitempty" xorm:"seqence"`

	toolbox.Menu `xorm:"extends"`

	Container []*Menu
}

type menuWeaver struct {
	core *hub_engine.Core
	db   *DB

	mu             sync.RWMutex
	menuList       []toolbox.Menu
	byApplications map[string]map[string]*Menu
}

func (weaver *menuWeaver) LoadFromDB() error {
	var allList []Menu
	err := weaver.db.Menus().Where().All(allList)
	if err != nil {
		return errors.New("LoadFromDB: " + err.Error())
	}

	byApplications := map[string]map[string]*Menu{}
	for idx, menu := range allList {
		newInGroup := byApplications[menu.Application]
		if newInGroup == nil {
			newInGroup = map[string]*Menu{}
		}
		newInGroup[menu.Name] = &allList[idx]
		byApplications[menu.Application] = newInGroup
	}

	menuList := generateMenuTree(byApplications)

	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.byApplications = byApplications
	weaver.menuList = menuList
	return nil
}

func upsertMenuListRecursive(db *DB, parentID int64, app string, menuList []toolbox.Menu,
	oldInGroup, newInGroup map[string]*Menu, idList *[]int64) error {
	for idx, menuItem := range menuList {
		var old *Menu
		var ok bool

		if oldInGroup != nil {
			old, ok = oldInGroup[menuItem.Name]
		}
		if !ok || old == nil {
			old = &Menu{}

			err := db.Menus().Where(orm.Cond{"application": app, "name": menuItem.Name}).One(old)
			if err != nil {
				if orm.ErrNotFound != err {
					return err
				}

				old.ID = 0
				old.ParentID = 0
			}
		}

		old.Application = app
		old.Seqence = int64(idx) + 1
		toolbox.MergeMenuWithNoChildren(&old.Menu, &menuItem)

		var err error
		if old.ID == 0 {
			var id interface{}
			old.ParentID = parentID
			//old.Seqence = 0
			id, err = db.Menus().
				Nullable("parent_id").
				Insert(old)
			old.ID = id.(int64)
		} else {
			err = db.Menus().ID(old.ID).Update(old)
		}

		if err != nil {
			return err
		}

		*idList = append(*idList, old.ID)
		newInGroup[menuItem.Name] = old

		err = upsertMenuListRecursive(db, old.ID, app, menuItem.Children, oldInGroup, newInGroup, idList)
		if err != nil {
			return err
		}
	}

	return nil
}

func upsertMenuList(db *DB, parentID int64, app string, menuList []toolbox.Menu, oldInGroup map[string]*Menu) (map[string]*Menu, error) {
	newInGroup := map[string]*Menu{}
	idList := make([]int64, 0, len(menuList))
	err := upsertMenuListRecursive(db, parentID, app, menuList, oldInGroup, newInGroup, &idList)
	if err != nil {
		return nil, err
	}
	_, err = db.Menus().Where(orm.Cond{"application": app, "id NOT IN": idList}).Delete()
	if err != nil {
		return nil, err
	}
	return newInGroup, nil
}

func (weaver *menuWeaver) Update(app string, menuList []toolbox.Menu) error {
	weaver.mu.RLock()
	oldList := weaver.byApplications[app]
	weaver.mu.RUnlock()

	if len(menuList) == 0 && len(oldList) == 0 {
		return nil
	}

	newInGroup, err := func(oldListOfApp map[string]*Menu) (map[string]*Menu, error) {
		tx, err := weaver.db.Begin()
		defer util.CloseWith(tx)

		newList, err := upsertMenuList(tx, 0, app, menuList, oldListOfApp)
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

	weaver.byApplications[app] = newInGroup
	weaver.menuList = generateMenuTree(weaver.byApplications)

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
	return IsSubset(allItems, subset)
}

func IsSubset(allItems, subset []toolbox.Menu) bool {
	for _, item := range subset {
		raw := toolbox.SearchMenuInTree(allItems, item.Name)
		if raw == nil || !toolbox.IsSameMenu(item, *raw) {
			return false
		}
	}
	return true
}

func generateMenuTree(byApps map[string]map[string]*Menu) []toolbox.Menu {
	byID := map[int64]*Menu{}
	for _, menuList := range byApps {
		for _, menu := range menuList {
			if menu == nil {
				continue
			}

			byID[menu.ID] = menu
		}
	}

	byContainerName := map[string]*Menu{}
	for _, menu := range byID {
		if !toolbox.IsMenuContainer(&menu.Menu) {
			continue
		}

		if old, ok := byContainerName[menu.Name]; ok && old != nil {
			var buf bytes.Buffer
			buf.WriteString("容器 ")
			buf.WriteString(menu.Name)
			buf.WriteString(" 已存在")

			buf.WriteString(", 之前的应用为 ")
			buf.WriteString(old.Application)
			if old.ParentID != 0 {
				if parent := byID[old.ParentID]; parent != nil {
					buf.WriteString(", 父节点为 ")
					buf.WriteString(parent.Name)
				}
			}
			buf.WriteString(", 当前的应用为 ")
			buf.WriteString(menu.Application)
			if menu.ParentID != 0 {
				if parent := byID[menu.ParentID]; parent != nil {
					buf.WriteString(", 父节点为 ")
					buf.WriteString(parent.Name)
				}
			}

			panic(buf.String())
		}
	}

	topMenuList := make([]*Menu, 0, 16)
	for _, menu := range byID {
		if menu.ParentID == 0 {
			if menu.Category == "" {
				topMenuList = append(topMenuList, menu)
				continue
			}

			if menu.Category == toolbox.MenuCategoryLinkto {
				container := byContainerName[menu.Name]
				if container == nil {
					panic(errors.New("应用 " + menu.Application + " 的菜单容器 " + menu.Name + " 没有找到"))
				}
				switch container.Category {
				case toolbox.MenuCategoryChildrenContainer:
					container.Container = append(container.Container, menu.Container...)
				case toolbox.MenuCategoryInlineContainer:
					if container.ParentID != 0 {
						parent := byID[container.ParentID]
						if parent != nil {
							parent.Container = append(parent.Container, menu.Container...)
							break
						}
					}
					container.Container = append(container.Container, menu.Container...)
				default:
					panic(errors.New("应用 " + container.Application + " 的菜单容器 " + container.Name + " 的 category 不可识别"))
				}
			}

			panic(errors.New("菜单 " + menu.Name + " 的 category 不可识别"))
		}

		parent := byID[menu.ParentID]
		if parent == nil {
			panic(fmt.Errorf("菜单(%d:%s) 找不到父节点 %d", menu.ID, menu.Name, menu.ParentID))
		}
		parent.Container = append(parent.Container, menu)
	}

	results := make([]toolbox.Menu, 0, len(topMenuList))
	for _, menu := range topMenuList {
		results = append(results, menu.Menu)
	}
	return results
}

func sortMenuList(list []*Menu) {
	if len(list) == 0 {
		return
	}

	sort.Slice(list, func(a, b int) bool {
		return list[a].Application < list[b].Application ||
			(list[a].Application == list[b].Application &&
				list[a].Seqence < list[b].Seqence)
	})

	for _, menu := range list {
		sortMenuList(menu.Container)
	}
}

func copyToMenuList(list []*Menu) []toolbox.Menu {
	if len(list) == 0 {
		return nil
	}

	results := make([]toolbox.Menu, 0, len(list))
	for _, menu := range list {
		menu.Children = copyToMenuList(menu.Container)
		results = append(results, menu.Menu)
	}
	return results
}
