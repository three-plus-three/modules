package toolbox

import (
	"bytes"
	"io"
	"strings"
)

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
	Classes    string `json:"classes,omitempty" xorm:"classes"`

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
	if name == menu.UID || strings.HasPrefix(menu.UID, name) {
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
	panic("菜单的级数太多了，最多只支持 3 级 - " + menu.Title + "/" + menu.UID)
}

func FormatMenus(out io.Writer, isIgnore func(name string) bool, menuList []Menu, layer int, indent bool) {
	if isIgnore == nil {
		isIgnore = func(string) bool {
			return false
		}
	}
	if layer > 0 && indent {
		out.Write(bytes.Repeat([]byte("  "), layer))
	}
	out.Write([]byte("[\r\n"))
	layer++
	for idx, menu := range menuList {
		if layer > 0 {
			out.Write(bytes.Repeat([]byte("  "), layer))
		}
		out.Write([]byte("{"))

		needComma := false
		if menu.UID != "" && !isIgnore("uid") {
			io.WriteString(out, `"uid":"`)
			io.WriteString(out, menu.UID)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.Title != "" && !isIgnore("title") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"title":"`)
			io.WriteString(out, menu.Title)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.Permission != "" && !isIgnore("permission") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"permission":"`)
			io.WriteString(out, menu.Permission)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.License != "" && !isIgnore("license") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"license":"`)
			io.WriteString(out, menu.License)
			io.WriteString(out, "\"")
			needComma = true
		}
		if menu.Icon != "" && !isIgnore("icon") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"icon":"`)
			io.WriteString(out, menu.Icon)
			io.WriteString(out, "\"")
		}

		if menu.Classes != "" && !isIgnore("classes") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"classes":"`)
			io.WriteString(out, menu.Classes)
			io.WriteString(out, "\"")
			needComma = true
		}

		if menu.URL != "" && !isIgnore("url") {
			if needComma {
				io.WriteString(out, `,`)
			}
			io.WriteString(out, `"url":"`)
			io.WriteString(out, menu.URL)
			io.WriteString(out, "\"")
			needComma = true
		}

		if len(menu.Children) > 0 && !isIgnore("children") {
			if needComma {
				io.WriteString(out, `,`)
			}

			out.Write([]byte("\r\n"))
			if layer > 0 {
				out.Write(bytes.Repeat([]byte("  "), layer+1))
			}

			io.WriteString(out, `"children":`)
			FormatMenus(out, isIgnore, menu.Children, layer+1, false)
		}

		out.Write([]byte("}"))

		if idx != len(menuList)-1 {
			out.Write([]byte(",\r\n"))
		} else {
			out.Write([]byte("\r\n"))
		}
	}

	if (layer - 1) > 0 {
		out.Write(bytes.Repeat([]byte("  "), layer))
	}
	out.Write([]byte("]"))
}
