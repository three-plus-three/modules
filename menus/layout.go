package menus

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/util"
)

// 菜单的分类
const (
	categoryNull     = "null"
	categoryLocation = "location"
	categoryRemove   = "remove"
	categoryWatch    = "watch"

	locationAfter   = "after"
	locationBefore  = "before"
	locationReplace = "replace"
)

func isMenu(category string) bool {
	return category == ""
}

// LayoutItem 表示一个菜单
type LayoutItem struct {
	Category   string `json:"category" xorm:"category"`
	Location   string `json:"location" xorm:"location"`
	Target     string `json:"target" xorm:"target"`
	Inline     bool   `json:"inline" xorm:"inline"`
	ID         string `json:"id" xorm:"id unique notnull"`
	Title      string `json:"title" xorm:"title notnull"`
	Permission string `json:"permission,omitempty" xorm:"permission"`
	License    string `json:"license,omitempty" xorm:"license"`
	URL        string `json:"url" xorm:"url"`
	Icon       string `json:"icon,omitempty" xorm:"icon"`

	Children []LayoutItem `json:"children,omitempty" xorm:"-"`
}

func (menu *LayoutItem) toMenu() toolbox.Menu {
	return toolbox.Menu{
		ID:         menu.ID,
		Title:      menu.Title,
		Permission: menu.Permission,
		License:    menu.License,
		URL:        menu.URL,
		Icon:       menu.Icon,
	}
}

func (menu *LayoutItem) forEach(cb func(menu *LayoutItem)) {
	if menu == nil {
		return
	}
	cb(menu)

	if len(menu.Children) == 0 {
		return
	}

	for idx := range menu.Children {
		menu.Children[idx].forEach(cb)
	}
}

type layoutImpl struct {
	mainLayout []LayoutItem
}

type container struct {
	layout *LayoutItem
	items  []toolbox.Menu
}

func (layout *layoutImpl) Generate(byApps map[string][]toolbox.Menu) ([]toolbox.Menu, error) {
	if len(byApps) == 0 {
		return nil, nil
	}

	byID := map[string]*container{}
	results := toToolboxMenus(layout.mainLayout, byID)

	var remains []toolbox.Menu
	for appName, menuList := range byApps {
		c, ok := byID["app."+appName]
		if ok {
			c.items = mergeMenuArray(c.items, menuList)
			continue
		}
		remains = append(remains, menuList...)
	}

	for len(remains) > 0 {
		var local []toolbox.Menu
		for idx := range remains {
			c, ok := byID[remains[idx].ID]
			if ok {
				if c.layout.Title == "" {
					c.layout.Title = remains[idx].Title
				}
				if c.layout.Permission == "" {
					c.layout.Permission = remains[idx].Permission
				}
				if c.layout.License == "" {
					c.layout.License = remains[idx].License
				}
				if c.layout.URL == "" || c.layout.URL == "#" {
					c.layout.URL = remains[idx].URL
				}
				if c.layout.Icon == "" {
					c.layout.Icon = remains[idx].Icon
				}

				c.items = mergeMenuArray(c.items, remains[idx].Children)
				continue
			}

			local = append(local, remains[idx])
		}

		if len(remains) == len(local) {
			var buf bytes.Buffer
			buf.WriteString("下列菜单不能处理:")
			for _, menu := range remains {
				buf.WriteString(menu.ID)
				buf.WriteString("(")
				buf.WriteString(menu.Title)
				buf.WriteString("),")
			}
			buf.Truncate(buf.Len() - 1)
			panic(errors.New(buf.String()))
		}

		remains = local
	}

	var removeList []*container
	var watchList []*container

	for _, c := range byID {
		switch c.layout.Category {
		case "":
			for idx := range results {
				if results[idx].ID == c.layout.ID {
					from := c.layout.toMenu()
					from.Children = c.items
					mergeMenuRecursive(&results[idx], &from)
					break
				}
			}
		case categoryRemove:
			removeList = append(removeList, c)
		case categoryWatch:
			watchList = append(watchList, c)
		case categoryLocation:
			switch strings.ToLower(strings.TrimSpace(c.layout.Location)) {
			case locationAfter:
				results = insertAfter(results, c, c.layout.Inline)
			case locationBefore:
				results = insertBefore(results, c, c.layout.Inline)
			case locationReplace:
				results = replaceInTree(results, c, c.layout.Inline)
			default:
				return nil, errors.New("菜单 " + c.layout.ID + " 的 location 不正确")
			}
		default:
			return nil, errors.New("菜单 " + c.layout.ID + " 的 category 不正确")
		}
	}

	for _, c := range removeList {
		c.layout.forEach(func(menu *LayoutItem) {
			results = removeInTree(results, menu.ID)
			results = removeInTree(results, menu.Target)
		})
		forEach(c.items, func(menu *toolbox.Menu) {
			results = removeInTree(results, menu.ID)
		})
	}
	for _, c := range watchList {
		results = watchInTree(results, c, c.layout.Target)
	}

	return clearDividerFromList(results), nil
}

func toToolboxMenus(mainLayout []LayoutItem, byID map[string]*container) []toolbox.Menu {
	results := make([]toolbox.Menu, 0, len(mainLayout))
	for idx, layout := range mainLayout {
		c := &container{
			layout: &mainLayout[idx],
		}
		if isMenu(layout.Category) {
			results = append(results, layout.toMenu())
			c.items = toToolboxMenus(layout.Children, byID)
			results[len(results)-1].Children =
				mergeMenuArray(results[len(results)-1].Children, c.items)
		}
		byID[layout.ID] = c
	}
	return results
}

// ReadLayout 载入 layout 文件
func ReadLayout(filename string) (Layout, error) {
	in, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read layout fail")
	}
	defer util.CloseWith(in)
	return readLayout(in)
}
func readLayout(in io.Reader) (Layout, error) {
	var mainLayout []LayoutItem
	err := json.NewDecoder(in).Decode(&mainLayout)
	if err != nil {
		return nil, errors.Wrap(err, "read layout fail")
	}
	return &layoutImpl{mainLayout: mainLayout}, nil
}

// Simple 简单布局器
var Simple Layout = &simpleLayout{}

// Layout 菜单布避生成器
type simpleLayout struct {
}

func (layout *simpleLayout) Generate(menuList map[string][]toolbox.Menu) ([]toolbox.Menu, error) {
	if len(menuList) == 0 {
		return nil, nil
	}
	if len(menuList) == 1 {
		for _, a := range menuList {
			return a, nil
		}
	}

	results := make([]toolbox.Menu, 0, len(menuList))
	for _, a := range menuList {
		results = append(results, a...)
	}
	return results, nil
}
