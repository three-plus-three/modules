package toolbox

// 菜单的分类
const (
	// MenuCategoryChildrenContainer = "children_container"
	// MenuCategoryInlineContainer   = "inline_container"
	// MenuCategoryLinkto            = "link_to"

	MenuDivider = "divider"
	MenuNull    = "null"
)

//
// func IsMenuContainer(menu *Menu) bool {
// 	return menu.Category == MenuCategoryChildrenContainer ||
// 		menu.Category == MenuCategoryInlineContainer
// }

// Menu 表示一个菜单
type Menu struct {
	UID        string `json:"uid,omitempty" xorm:"uid notnull"`
	Title      string `json:"title,omitempty" xorm:"title notnull"`
	Permission string `json:"permission,omitempty" xorm:"permission"`
	License    string `json:"license,omitempty" xorm:"license"`
	URL        string `json:"url,omitempty" xorm:"url"`
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
	if name == menu.UID {
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
