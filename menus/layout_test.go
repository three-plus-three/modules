package menus

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	"github.com/three-plus-three/modules/toolbox"
)

func TestLayoutSimple(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id":"monitor",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutInsertAfter(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "after",
    "target":   "home",
    "id":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutInsertAfterInline(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "after",
    "target":   "home",
    "inline":   true,
    "id":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},

		{
			ID:    "monitor.topology",
			Title: "网络视图",
			URL:   "/hengwei/web/network_topologies/root",
		},
		{
			ID:    "monitor.tour_inspection.inspect_files",
			Title: "巡检文件管理",
			URL:   "/hengwei/web/tour_inspection/inspect_files",
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}
	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutInsertBefore(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "before",
    "target":   "system",
    "id":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutInsertBeforeInline(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "before",
    "target":   "system",
    "inline":   true,
    "id":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},

		{
			ID:    "monitor.topology",
			Title: "网络视图",
			URL:   "/hengwei/web/network_topologies/root",
		},
		{
			ID:    "monitor.tour_inspection.inspect_files",
			Title: "巡检文件管理",
			URL:   "/hengwei/web/tour_inspection/inspect_files",
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutReplace(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id": "aaa"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "replace",
    "target":   "aaa",
    "id":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func TestLayoutRemove(t *testing.T) {
	layoutText := `[{
    "id": "home",
    "icon": "fa-home"
  },
  {
    "id":"aaa",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "id":"bbb",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "id":"monitor",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "id": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "remove",
    "target":   "aaa",
    "id":       "remove_aaa",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  },
  {
    "category": "watch",
    "target":   "aaa",
    "location": "bbb",
    "id":       "watch_aaa",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout(strings.NewReader(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			ID:    "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			ID:    "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					ID:    "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					ID:    "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			ID:    "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					ID:    "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					ID:    "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
			},
		},
	}

	resultList, err := layout.Generate(map[string][]toolbox.Menu{"aaa": menuList})
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}
}

func dumpMenus(t *testing.T, resultList, menuList []toolbox.Menu) {
	bs1, _ := json.MarshalIndent(resultList, "", "  ")
	bs2, _ := json.MarshalIndent(menuList, "", "  ")

	results := difflib.Diff(strings.Split(string(bs2), "\n"),
		strings.Split(string(bs1), "\n"))
	if 0 != len(results) {
		for _, rec := range results {
			t.Error(rec.String())
		}
	}
}
