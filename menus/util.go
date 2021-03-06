package menus

import (
	"github.com/three-plus-three/modules/toolbox"
)

const (
	actInsertAfterInTree = iota
	actInsertBeforeInTree
	actReplaceInTree
)

func insertAfter(allList []toolbox.Menu, c *container, isInline bool) (bool, []toolbox.Menu) {
	return insertToTree(allList, c, isInline, actInsertAfterInTree)
}

func insertBefore(allList []toolbox.Menu, c *container, isInline bool) (bool, []toolbox.Menu) {
	return insertToTree(allList, c, isInline, actInsertBeforeInTree)
}

func replaceInTree(allList []toolbox.Menu, c *container, isInline bool) (bool, []toolbox.Menu) {
	return insertToTree(allList, c, isInline, actReplaceInTree)
}

func watchInTree(allList []toolbox.Menu, c *container, target string) []toolbox.Menu {
	found := SearchMenuInTree(allList, target)
	if found == nil {
		return removeInTree(allList, c.layout.Location)
	}
	return allList
}

func isEmptyURL(u string) bool {
	return u == "" || u == "#"
}

func insertToTree(allList []toolbox.Menu, c *container, isInline bool, act int) (bool, []toolbox.Menu) {
	for idx := range allList {
		if allList[idx].UID == c.layout.Target {
			if isInline {
				var results []toolbox.Menu
				switch act {
				case actInsertAfterInTree:
					if len(c.items) == 0 {
						if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
							c.items = append(c.items, c.layout.toMenu())
						}
					}

					results = make([]toolbox.Menu, len(allList)+len(c.items))
					copy(results, allList[:idx+1])
					copy(results[idx+1:], c.items)
					copy(results[idx+1+len(c.items):], allList[idx+1:])
				case actInsertBeforeInTree:
					if len(c.items) == 0 {
						if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
							c.items = append(c.items, c.layout.toMenu())
						}
					}

					results = make([]toolbox.Menu, len(allList)+len(c.items))
					copy(results, allList[:idx])
					copy(results[idx:], c.items)
					copy(results[idx+len(c.items):], allList[idx:])

				case actReplaceInTree:
					if len(c.items) == 0 {
						if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
							c.items = append(c.items, c.layout.toMenu())
						}
					}

					if len(c.items) > 0 {
						results = make([]toolbox.Menu, len(allList)+len(c.items))
						copy(results, allList[:idx])
						copy(results[idx:], c.items)
						copy(results[idx+len(c.items):], allList[idx:])
					} else {
						results = allList
					}
				default:
					if len(c.items) == 0 {
						if c.layout.Category == toolbox.MenuNull {
							results = removeInTree(allList, c.layout.UID)
						} else {
							results = allList
						}
					} else {
						results = make([]toolbox.Menu, len(allList)+len(c.items)-1)
						copy(results, allList[:idx])
						copy(results[idx:], c.items)
						copy(results[idx+len(c.items):], allList[idx+1:])
					}
				}
				return true, results
			}

			var results []toolbox.Menu
			switch act {
			case actInsertAfterInTree:
				if len(c.items) == 0 {
					if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
						c.items = append(c.items, c.layout.toMenu())
					}
				}
				if len(c.items) != 0 || !isEmptyURL(c.layout.URL) {
					results = make([]toolbox.Menu, len(allList)+1)
					copy(results, allList[:idx+1])
					results[idx+1] = c.layout.toMenu()
					results[idx+1].Children = c.items
					copy(results[idx+2:], allList[idx+1:])
				} else {
					results = allList
					//fmt.Println("==== after", c.layout.UID, c.layout.Target, len(c.items), c.layout.URL)
				}
			case actInsertBeforeInTree:
				if len(c.items) == 0 {
					if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
						c.items = append(c.items, c.layout.toMenu())
					}
				}

				if len(c.items) != 0 || !isEmptyURL(c.layout.URL) {
					results = make([]toolbox.Menu, len(allList)+1)
					copy(results, allList[:idx])
					results[idx] = c.layout.toMenu()
					results[idx].Children = c.items
					copy(results[idx+1:], allList[idx:])
				} else {
					results = allList
				}
			case actReplaceInTree:
				if len(c.items) == 0 {
					if c.layout.Title == toolbox.MenuDivider || (!isEmptyURL(c.layout.URL) && c.layout.Title != "") {
						item := c.layout.toMenu()
						mergeMenuNonrecursive(&item, &allList[idx])
						c.items = append(c.items, item)
					}
				}

				if len(c.items) != 0 || !isEmptyURL(c.layout.URL) {
					item := c.layout.toMenu()
					mergeMenuNonrecursive(&item, &allList[idx])

					results = make([]toolbox.Menu, len(allList)+1)
					copy(results, allList[:idx])
					results[idx] = item
					results[idx].Children = c.items
					copy(results[idx+1:], allList[idx:])
				} else {
					results = allList
				}
			default:
				if len(c.items) == 0 {
					if c.layout.Category == toolbox.MenuNull {
						results = removeInTree(allList, c.layout.UID)
					} else {
						results = allList
					}
				} else {
					results = allList
					results[idx] = c.layout.toMenu()
					results[idx].Children = c.items
				}
			}
			return true, results
		}

		found, children := insertToTree(allList[idx].Children, c, isInline, act)
		allList[idx].Children = children
		if found {
			return true, allList
		}
	}

	return false, allList
}

// SearchMenuInTree 在菜单树中查找指定的菜单
func SearchMenuInTree(allList []toolbox.Menu, uid string) *toolbox.Menu {
	for idx := range allList {
		if allList[idx].UID == uid {
			return &allList[idx]
		}
		found := SearchMenuInTree(allList[idx].Children, uid)
		if found != nil {
			return found
		}
	}

	return nil
}

// isSameMenuArray 判断两个菜单列表是否相等
func isSameMenuArray(newList, oldList []toolbox.Menu) bool {
	if len(newList) != len(oldList) {
		return false
	}

	for idx, newMenu := range newList {
		if !isSameMenu(newMenu, oldList[idx]) {
			return false
		}
	}
	return true
}

// isSameMenu 判断两个菜单是否相等
func isSameMenu(newMenu, oldMenu toolbox.Menu) bool {
	if newMenu.UID != oldMenu.UID {
		return false
	}
	if newMenu.Title != oldMenu.Title {
		return false
	}
	if newMenu.Classes != oldMenu.Classes {
		return false
	}
	if newMenu.Permission != oldMenu.Permission {
		return false
	}
	if newMenu.License != oldMenu.License {
		return false
	}
	if newMenu.URL != oldMenu.URL {
		return false
	}
	if newMenu.Icon != oldMenu.Icon {
		return false
	}
	return isSameMenuArray(newMenu.Children, oldMenu.Children)
}

// mergeMenuArray 合并菜单列表
func mergeMenuArray(allList, newList []toolbox.Menu) []toolbox.Menu {
	for menuIdx := range newList {
		foundIdx := -1
		for idx := range allList {
			if allList[idx].UID == newList[menuIdx].UID {
				foundIdx = idx
			}
		}
		if foundIdx < 0 {
			allList = append(allList, newList[menuIdx])
		} else {
			mergeMenuRecursive(&allList[foundIdx], &newList[menuIdx])
		}
	}
	return allList
}

// mergeMenuRecursive 合并菜单，但子菜单不合并
func mergeMenuRecursive(to, from *toolbox.Menu) {
	mergeMenuNonrecursive(to, from)
	to.Children = mergeMenuArray(to.Children, from.Children)
}

// mergeMenuNonrecursive 合并菜单，但子菜单不合并
func mergeMenuNonrecursive(to, from *toolbox.Menu) {
	if to.UID == "" {
		to.UID = from.UID
	}
	if to.Title == "" {
		to.Title = from.Title
	}
	if to.Classes == "" {
		to.Classes = from.Classes
	}
	if to.Permission == "" {
		to.Permission = from.Permission
	}
	if to.License == "" {
		to.License = from.License
	}
	if isEmptyURL(to.URL) && !isEmptyURL(from.URL) {
		to.URL = from.URL
	}
	if to.Icon == "" {
		to.Icon = from.Icon
	}
}

func mergeLayoutMenuNonrecursive(layout *LayoutItem, from *toolbox.Menu) {
	if layout.Title == "" {
		layout.Title = from.Title
	}
	if layout.Permission == "" {
		layout.Permission = from.Permission
	}
	if layout.License == "" {
		layout.License = from.License
	}
	if isEmptyURL(layout.URL) && !isEmptyURL(from.URL) {
		layout.URL = from.URL
	}
	if layout.Icon == "" {
		layout.Icon = from.Icon
	}
	if layout.Classes == "" {
		layout.Classes = from.Classes
	}
}

// removeInTree 从列表中删除指定的菜单
func removeInTree(menuList []toolbox.Menu, name string) []toolbox.Menu {
	offset := 0
	for i := 0; i < len(menuList); i++ {
		if menuList[i].UID == name {
			continue
		}

		if i != offset {
			menuList[offset] = menuList[i]
		}

		offset++
	}

	for i := 0; i < offset; i++ {
		menuList[i].Children = removeInTree(menuList[i].Children, name)
	}

	return menuList[:offset]
}

func forEach(allList []toolbox.Menu, cb func(menu *toolbox.Menu)) {
	if len(allList) == 0 {
		return
	}

	for idx := range allList {
		cb(&allList[idx])
	}

	for idx := range allList {
		forEach(allList[idx].Children, cb)
	}
}

func ClearDividerFromList(list []toolbox.Menu) []toolbox.Menu {
	if len(list) == 0 {
		return nil
	}

	offset := 0
	prev := true
	for idx := range list {
		list[idx].Children = ClearDividerFromList(list[idx].Children)
		if list[idx].UID == toolbox.MenuDivider || list[idx].Title == toolbox.MenuDivider {
			if prev {
				continue
			}
			prev = true
		} else {
			prev = false
		}

		if idx != offset {
			list[offset] = list[idx]
		}
		offset++
	}

	if prev {
		offset--
	}
	if offset <= 0 {
		return nil
	}
	return list[:offset]
}
