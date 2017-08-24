package toolbox

// Menu 表示一个菜单
type Menu struct {
	Name       string
	Permission string
	Title      string
	URL        string
	Icon       string

	Children []Menu
}

// IsActive 判断这个菜单是否是展开的
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
