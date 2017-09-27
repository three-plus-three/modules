package toolbox

// 菜单的分类
const (
	MenuCategoryChildrenContainer = "children_container"
	MenuCategoryInlineContainer   = "inline_container"
	MenuCategoryLinkto            = "link_to"

	MenuDivider = "divider"
)

func IsMenuContainer(menu *Menu) bool {
	return menu.Category == MenuCategoryChildrenContainer ||
		menu.Category == MenuCategoryInlineContainer
}

// Menu 表示一个菜单
type Menu struct {
	Category   string `json:"category,omitempty" xorm:"category"`
	Name       string `json:"name" xorm:"name unique notnull"`
	Title      string `json:"title" xorm:"title notnull"`
	Permission string `json:"permission,omitempty" xorm:"permission"`
	License    string `json:"license,omitempty" xorm:"license"`
	URL        string `json:"url" xorm:"url"`
	Icon       string `json:"icon,omitempty" xorm:"icon"`

	Children []Menu `json:"children,omitempty" xorm:"-"`
}

// TableName 用于 xorm 的表名
func (menu *Menu) TableName() string {
	return "tpt_menus"
}

// IsActiveWith 判断这个菜单是否是展开的
func (menu Menu) IsActiveWith(ctx map[string]interface{}) bool {
	o := ctx["active"]
	if o == nil {
		o = ctx["controller"]
		if o == nil {
			return false
		}
	}

	name, ok := o.(string)
	if !ok {
		return false
	}
	return menu.IsActive(name)
}

// IsActive 判断这个菜单是否是展开的
func (menu Menu) IsActive(name string) bool {
	if name == menu.Name {
		return true
	}

	for _, child := range menu.Children {
		if child.IsActive(name) {
			return true
		}
	}
	return false
}

// Fail 产生一个 panic
func (menu Menu) Fail() interface{} {
	panic("菜单的级数太多了，最多只支持 3 级")
}

// SearchMenuInTree 在菜单树中查找指定的菜单
func SearchMenuInTree(allList []Menu, name string) *Menu {
	for idx := range allList {
		if allList[idx].Name == name {
			return &allList[idx]
		}
		found := SearchMenuInTree(allList[idx].Children, name)
		if found != nil {
			return found
		}
	}

	return nil
}

// IsSameMenuArray 判断两个菜单列表是否相等
func IsSameMenuArray(newList, oldList []Menu) bool {
	if len(newList) != len(oldList) {
		return false
	}

	for idx, newMenu := range newList {
		if !IsSameMenu(newMenu, oldList[idx]) {
			return false
		}
	}
	return true
}

// IsSameMenu 判断两个菜单是否相等
func IsSameMenu(newMenu, oldMenu Menu) bool {
	if newMenu.Name != oldMenu.Name {
		return false
	}
	if newMenu.Title != oldMenu.Title {
		return false
	}
	if newMenu.Permission != oldMenu.Permission {
		return false
	}
	if newMenu.URL != oldMenu.URL {
		return false
	}
	if newMenu.Icon != oldMenu.Icon {
		return false
	}
	return IsSameMenuArray(newMenu.Children, oldMenu.Children)
}

// MergeMenus 合并菜单列表
func MergeMenus(allList, newList []Menu) []Menu {
	for _, menu := range newList {
		foundIdx := -1
		for idx := range allList {
			if allList[idx].Name == menu.Name {
				foundIdx = idx
			}
		}
		if foundIdx < 0 {
			allList = append(allList, menu)
		} else {
			MergeMenuWithNoChildren(&allList[foundIdx], &menu)
			allList[foundIdx].Children = MergeMenus(allList[foundIdx].Children, menu.Children)
		}
	}
	return allList
}

// MergeMenuWithNoChildren 合并菜单，但子菜单不合并
func MergeMenuWithNoChildren(to, from *Menu) {
	if to.Category == "" {
		to.Category = from.Category
	}
	if to.Title == "" {
		to.Title = from.Title
	}
	if to.Permission == "" {
		to.Permission = from.Permission
	}
	if to.License == "" {
		to.License = from.License
	}
	if to.URL == "" || to.URL == "#" {
		to.URL = from.URL
	}
	if to.Icon == "" {
		to.Icon = from.Icon
	}
}

// Remove 从列表中删除指定的菜单
func Remove(menuList []Menu, name string) []Menu {
	offset := 0
	for i := 0; i < len(menuList); i++ {
		if menuList[i].Name == name {
			continue
		}
		offset++
	}

	for i := 0; i < offset; i++ {
		menuList[i].Children = Remove(menuList[i].Children, name)
	}

	return menuList[:offset]
}
