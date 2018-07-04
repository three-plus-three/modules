package menus

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/urlutil"
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
	UID        string `json:"uid" xorm:"uid unique notnull"`
	Title      string `json:"title" xorm:"title notnull"`
	Classes    string `json:"classes,omitempty" xorm:"classes"`
	Permission string `json:"permission,omitempty" xorm:"permission"`
	License    string `json:"license,omitempty" xorm:"license"`
	URL        string `json:"url" xorm:"url"`
	Icon       string `json:"icon,omitempty" xorm:"icon"`

	Children []LayoutItem `json:"children,omitempty" xorm:"-"`
}

func (menu *LayoutItem) toMenu() toolbox.Menu {
	return toolbox.Menu{
		UID:        menu.UID,
		Title:      menu.Title,
		Classes:    menu.Classes,
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

func (layout *layoutImpl) MergeFrom(merge Layout) error {
	layout.mainLayout = append(layout.mainLayout, merge.(*layoutImpl).mainLayout...)
	return nil
}

func (layout *layoutImpl) Stats() interface{} {
	return layout.mainLayout
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
		//for _, appName := range appNames {
		//	menuList := byApps[appName]
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
			c, ok := byID[remains[idx].UID]
			if !ok {
				local = append(local, remains[idx])
				continue
			}

			if c.layout.Title == "" {
				c.layout.Title = remains[idx].Title
			}
			if c.layout.Classes == "" {
				c.layout.Classes = remains[idx].Classes
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
		}

		if len(remains) == len(local) {
			var buf bytes.Buffer
			buf.WriteString("下列菜单不能处理:")
			for _, menu := range local {
				buf.WriteString(menu.UID)
				buf.WriteString("(")
				buf.WriteString(menu.Title)
				buf.WriteString("),")
			}
			buf.Truncate(buf.Len() - 1)
			return nil, errors.New(buf.String())
		}

		remains = local
	}

	var removeList []*container
	var watchList []*container
	var allList []*container

	for _, c := range byID {
		allList = append(allList, c)
	}

	for len(allList) > 0 {
		var local []*container
		for _, c := range allList {
			switch c.layout.Category {
			case "":
				foundIdx := -1
				for idx := range results {
					if results[idx].UID == c.layout.UID {
						foundIdx = idx
						break
					}
				}
				if foundIdx >= 0 {
					from := c.layout.toMenu()
					from.Children = c.items
					mergeMenuRecursive(&results[foundIdx], &from)
				} else {
					local = append(local, c)
				}
			case categoryRemove:
				removeList = append(removeList, c)
			case categoryWatch:
				watchList = append(watchList, c)
			case categoryLocation:

				var found bool

				switch strings.ToLower(strings.TrimSpace(c.layout.Location)) {
				case locationAfter:
					found, results = insertAfter(results, c, c.layout.Inline)
				case locationBefore:
					found, results = insertBefore(results, c, c.layout.Inline)
				case locationReplace:
					found, results = insertBefore(results, c, c.layout.Inline)
					if found {
						removeList = append(removeList, c)
					}
				default:
					return nil, errors.New("菜单 " + c.layout.UID + " 的 location 不正确")
				}

				if !found {
					//log.Println("insertToTree:", "target =", c.layout.Target, ", uid =", c.layout.UID) // spew.Sprint(allList))
					local = append(local, c)
				}
			default:
				return nil, errors.New("菜单 " + c.layout.UID + " 的 category 不正确")
			}
		}

		if len(local) == len(allList) {
			var buf bytes.Buffer
			buf.WriteString("下列 layout 菜单不能处理:")
			for _, menu := range local {
				buf.WriteString(menu.layout.UID)
				buf.WriteString("(")
				buf.WriteString(menu.layout.Title)
				buf.WriteString("),")
			}
			buf.Truncate(buf.Len() - 1)
			return nil, errors.New(buf.String())
		}
		allList = local
	}

	for _, c := range removeList {
		if c.layout.Category == categoryLocation && c.layout.Location == locationReplace {
			results = removeInTree(results, c.layout.Target)
		} else {
			c.layout.forEach(func(menu *LayoutItem) {
				if menu.UID != "" {
					results = removeInTree(results, menu.UID)
				}
				if menu.Target != "" {
					results = removeInTree(results, menu.Target)
				}
			})
			forEach(c.items, func(menu *toolbox.Menu) {
				if menu.UID != "" {
					results = removeInTree(results, menu.UID)
				}
			})
		}
	}

	for _, item := range layout.mainLayout {
		if item.Category == categoryRemove {
			if item.Target != "" {
				results = removeInTree(results, item.Target)
			}

			item.forEach(func(menu *LayoutItem) {
				if menu.Target != "" {
					results = removeInTree(results, menu.Target)
				}
			})
		}
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

		if layout.UID == "" {
			if layout.Category != categoryRemove {
				panic("layout with target = '" + layout.Target + "' and category = '" + layout.Category + "' is invalid, uid is empty")
			}
		} else if _, exists := byID[layout.UID]; !exists {
			byID[layout.UID] = c
		} else {
			panic("layout.UID '" + layout.UID + "' is duplicated")
		}
	}
	return results
}

// Simple 简单布局器
var Simple Layout = &simpleLayout{}

// Layout 菜单布避生成器
type simpleLayout struct {
}

func (layout *simpleLayout) Stats() interface{} {
	return "simple"
}

func (layout *simpleLayout) MergeFrom(Layout) error {
	return errors.New("simpleLayout is unsupported")
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

func ReadLayoutFromDirectory(dirname string, args map[string]interface{}) (Layout, error) {
	files, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, errors.Wrap(err, "read layout from directory fail")
	}

	var mainLayout []LayoutItem
	for _, fi := range files {
		filename := fi.Name()
		if fi.IsDir() ||
			strings.HasPrefix(filename, ".") ||
			!strings.HasSuffix(strings.ToLower(filename), ".json") {
			continue
		}

		layout, err := ReadLayout(filepath.Join(dirname, filename), args)
		if err != nil {
			return nil, err
		}
		items := layout.(*layoutImpl).mainLayout
		if len(items) > 0 {
			mainLayout = append(mainLayout, items...)
		}
	}
	return &layoutImpl{mainLayout: mainLayout}, nil
}

// ReadLayout 载入 layout 文件
func ReadLayout(filename string, args map[string]interface{}) (Layout, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "read layout fail")
	}

	t, err := template.New("default").Funcs(template.FuncMap{
		"join": urlutil.Join,
	}).Parse(string(bs))
	if err != nil {
		return nil, errors.Wrap(err, "parse url template in '"+filename+"' fail")
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, args); err != nil {
		return nil, errors.Wrap(err, "generate layout in '"+filename+"' fail")
	}

	return readLayout(buf.Bytes())
}

func readLayout(in []byte) (Layout, error) {
	data, err := util.HjsonToJSON(in)
	if err != nil {
		return nil, errors.Wrap(err, "read layout fail")
	}

	var mainLayout []LayoutItem
	err = json.Unmarshal(data, &mainLayout)
	if err != nil {
		return nil, errors.Wrap(err, "read layout fail")
	}
	return &layoutImpl{mainLayout: mainLayout}, nil
}

func ReadProductsFromLayout(env *environment.Environment) (*LayoutItem, error) {
	layoutArgs := map[string]interface{}{
		"httpAddress": env.GetServiceConfig(environment.ENV_WSERVER_PROXY_ID).UrlFor(),
		"urlPrefix":   env.DaemonUrlPath,
		"urlRoot":     env.DaemonUrlPath,
	}
	layout, err := ReadLayoutFromDirectory(env.Fs.FromLib("menu_layouts/default"), layoutArgs)
	if err != nil {
		return nil, errors.Wrap(err, "ReadProductsLayout")
	}

	customlayout, err := ReadLayoutFromDirectory(env.Fs.FromDataConfig("menu_layouts/default"), layoutArgs)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, errors.Wrap(err, "ReadProductsLayout")
		}
	}
	if customlayout != nil {
		if err := layout.MergeFrom(customlayout); err != nil {
			return nil, errors.Wrap(err, "ReadProductsLayout")
		}
	}

	for _, item := range layout.(*layoutImpl).mainLayout {
		for item.UID == "app.products" {
			return &item, nil
		}
	}
	return nil, nil
}
