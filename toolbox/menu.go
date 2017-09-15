package toolbox

// Menu 表示一个菜单
type Menu struct {
	Name       string `json:"name"`
	Title      string `json:"title"`
	Permission string `json:"permission,omitempty"`
	URL        string `json:"url"`
	Icon       string `json:"icon,omitempty"`

	Children []Menu `json:"children,omitempty"`
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
