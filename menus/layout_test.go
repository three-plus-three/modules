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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid":"monitor",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "after",
    "target":   "home",
    "uid":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "after",
    "target":   "home",
    "inline":   true,
    "uid":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  },
  {
    "category": "location",
    "target": "system.sql_script_manage",
    "location": "after",
    "title": "帮助",
    "inline": true,
    "uid": "nm.system.help",
    "url": "hengwei/internal/doc/"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},

		{
			UID:   "monitor.topology",
			Title: "网络视图",
			URL:   "/hengwei/web/network_topologies/root",
		},
		{
			UID:   "monitor.tour_inspection.inspect_files",
			Title: "巡检文件管理",
			URL:   "/hengwei/web/tour_inspection/inspect_files",
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
					Title: "数据库脚本管理",
					URL:   "/hengwei/web/tools/sql_script_manage",
				},
				{
					UID:   "nm.system.help",
					Title: "帮助",
					URL:   "hengwei/internal/doc/",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "before",
    "target":   "system",
    "uid":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "before",
    "target":   "system",
    "inline":   true,
    "uid":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},

		{
			UID:   "monitor.topology",
			Title: "网络视图",
			URL:   "/hengwei/web/network_topologies/root",
		},
		{
			UID:   "monitor.tour_inspection.inspect_files",
			Title: "巡检文件管理",
			URL:   "/hengwei/web/tour_inspection/inspect_files",
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid": "aaa"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "location",
    "location": "replace",
    "target":   "aaa",
    "uid":       "monitor",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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
    "uid": "home",
    "icon": "fa-home"
  },
  {
    "uid":"aaa",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "uid":"bbb",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "uid":"monitor",
    "title": "运维管理",
    "url": "#",
    "icon": "fa-trello"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog"
  },
  {
    "category": "remove",
    "target":   "aaa",
    "uid":       "remove_aaa",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  },
  {
    "category": "watch",
    "target":   "aaa",
    "location": "bbb",
    "uid":       "watch_aaa",
    "title":    "运维管理",
    "url":      "#",
    "icon":     "fa-trello"
  }]`

	layout, err := readLayout([]byte(layoutText))
	if err != nil {
		t.Error(err)
		return
	}

	var menuList = []toolbox.Menu{
		{
			UID:   "home",
			Title: "首页",
			URL:   "/hengwei/web",
			Icon:  "fa-home",
		},
		{
			UID:   "monitor",
			Title: "运维管理",
			URL:   "#",
			Icon:  "fa-trello",
			Children: []toolbox.Menu{
				{
					UID:   "monitor.topology",
					Title: "网络视图",
					URL:   "/hengwei/web/network_topologies/root",
				},
				{
					UID:   "monitor.tour_inspection.inspect_files",
					Title: "巡检文件管理",
					URL:   "/hengwei/web/tour_inspection/inspect_files",
				},
			},
		},
		{
			UID:   "system",
			Title: "系统管理",
			URL:   "#",
			Icon:  "fa-cog",
			Children: []toolbox.Menu{
				{
					UID:   "system.divider1",
					Title: "divider",
					URL:   "#",
				},
				{
					UID:   "system.sql_script_manage",
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

func TestLayoutInRealWorld(t *testing.T) {
	rawText := `{
  "applications": {
    "mc": [
      {
        "uid": "mc.AutoBackup",
        "title": "备份数据",
        "permission": "mc.AutoBackup",
        "children": [
          {
            "uid": "mc.AutoBackup.Index",
            "title": "备份数据",
            "permission": "mc.AutoBackup",
            "url": "/hengwei/mc_dev/autobackup/index?active="
          }
        ]
      },
      {
        "uid": "mc.NotificationGroups",
        "title": "通知组",
        "permission": "mc.NotificationGroups",
        "children": [
          {
            "uid": "mc.NotificationGroups.Index",
            "title": "通知组",
            "permission": "mc.NotificationGroups",
            "url": "/hengwei/mc_dev/notification_groups"
          }
        ]
      }
    ],
    "um": [
      {
        "uid": "HengweiUsers",
        "title": "用户管理",
        "url": "/hengwei/um/hengweiusers/index",
        "icon": "fa-user"
      },
      {
        "uid": "HengweiUserGroups",
        "title": "用户组管理",
        "url": "/hengwei/um/hengweiusergroups/index?groupId=0",
        "icon": "fa-users"
      },
      {
        "uid": "HengweiRoles",
        "title": "角色管理",
        "url": "/hengwei/um/hengweiroles/index",
        "icon": "fa-user-circle"
      },
      {
        "uid": "HengweiPermissionGroups",
        "title": "权限管理",
        "url": "/hengwei/um/hengweipermissiongroups/index?groupId=0",
        "icon": "fa-archive"
      }
    ],
    "wserver": [
      {
        "uid": "home",
        "title": "首页",
        "url": "/hengwei/web",
        "icon": "fa-home"
      },
      {
        "uid": "system",
        "title": "系统管理",
        "url": "#",
        "icon": "fa-cog",
        "children": [
          {
            "uid": "system.divider1",
            "title": "divider",
            "url": "#"
          },
          {
            "uid": "system.online_users",
            "title": "当前在线用户",
            "url": "/hengwei/web/users/online_users"
          },
          {
            "uid": "system.operation_logs",
            "title": "用户操作日志",
            "url": "/hengwei/web/system/operation_logs"
          },
          {
            "uid": "system.divider2",
            "title": "divider",
            "url": "#"
          },
          {
            "uid": "system.notification_group",
            "title": "通知组",
            "url": "/hengwei/web/notification_groups"
          },
          {
            "uid": "system.notification_rule",
            "title": "通知方式",
            "url": "/hengwei/web/notification_rules"
          },
          {
            "uid": "system.divider3",
            "title": "divider",
            "url": "#"
          },
          {
            "uid": "system.engine_nodes",
            "title": "采集引擎管理",
            "url": "/hengwei/web/system/engine_nodes"
          }
        ]
      }
    ]
  },
  "layout": {
    "default": [
      {
        "category": "",
        "location": "",
        "target": "",
        "inline": false,
        "uid": "home",
        "title": "首页",
        "url": "/hengwei/web",
        "icon": "fa-home"
      },
      {
        "uid": "am.home",
        "title": "资产管理",
        "url": "/hengwei/am",
        "icon": "fa-archive",
        "classes": "special_link"
      },
      {
        "uid": "itsm.home",
        "title": "服务管理",
        "url": "/hengwei/itsm",
        "icon": "fa-asterisk",
        "classes": "special_link"
      },
      {
        "category": "",
        "location": "",
        "target": "",
        "inline": false,
        "uid": "report_management",
        "title": "报表管理",
        "url": "#",
        "icon": "fa-list-alt"
      },
      {
        "category": "",
        "location": "",
        "target": "",
        "inline": false,
        "uid": "system",
        "title": "系统管理",
        "url": "#",
        "icon": "fa-cog"
      },
      {
        "category": "location",
        "location": "replace",
        "target": "system.online_users",
        "inline": true,
        "uid": "app.um",
        "title": "",
        "url": ""
      },
      {
        "category": "location",
        "location": "after",
        "target": "system.notification_group",
        "inline": true,
        "uid": "mc.AutoBackup",
        "title": "备份数据",
        "permission": "mc.AutoBackup",
        "url": ""
      },
      {
        "category": "location",
        "location": "replace",
        "target": "system.notification_group",
        "inline": true,
        "uid": "mc.NotificationGroups",
        "title": "通知组",
        "permission": "mc.NotificationGroups",
        "url": ""
      },
      {
        "category": "watch",
        "location": "system.notification_rule",
        "target": "system.notification_group",
        "inline": false,
        "uid": "remove.notifications",
        "title": "",
        "url": ""
      }
    ]
  },
  "menuList": [
    {
      "uid": "home",
      "title": "首页",
      "url": "/hengwei/web",
      "icon": "fa-home"
    },
    {
      "uid": "monitor",
      "title": "运维管理",
      "url": "#",
      "icon": "fa-trello",
      "children": [
        {
          "uid": "monitor.topology",
          "title": "网络视图",
          "url": "/hengwei/web/network_topologies/root"
        },
        {
          "uid": "monitor.computer_room",
          "title": "机房视图",
          "url": "/hengwei/web/room_topologies/first"
        },
        {
          "uid": "monitor.biz_view_list",
          "title": "业务视图",
          "url": "/hengwei/web/biz_management/biz_views"
        },
        {
          "uid": "monitor.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "monitor.view_panel",
          "title": "监控总览",
          "url": "/hengwei/web/integrated_monitor"
        },
        {
          "uid": "monitor.biz_browse",
          "title": "业务总览",
          "url": "/hengwei/web/biz_management/biz_overview"
        },
        {
          "uid": "monitor.services_browse",
          "title": "服务总览",
          "url": "/hengwei/web/services/tree_view"
        },
        {
          "uid": "monitor.device_browse",
          "title": "设备总览",
          "url": "/hengwei/web/network_devices/device_browse"
        },
        {
          "uid": "monitor.divider2",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "monitor.tour_inspection.current_task",
          "title": "当前巡检任务",
          "url": "/hengwei/web/tour_inspection/inspect_tasks/current_tasks"
        },
        {
          "uid": "monitor.tour_inspection.error_view",
          "title": "任务异常一览",
          "url": "/hengwei/web/tour_inspection/inspect_tasks/task_exceptions"
        },
        {
          "uid": "monitor.tour_inspection.error_confirm",
          "title": "巡检异常处理",
          "url": "/hengwei/web/tour_inspection/inspect_tasks/all_exceptions"
        },
        {
          "uid": "monitor.tour_inspection.query",
          "title": "巡检结果查询",
          "url": "/hengwei/web/tour_inspection/query"
        },
        {
          "uid": "monitor.tour_inspection.report",
          "title": "巡检报告查看",
          "url": "#"
        },
        {
          "uid": "monitor.divider",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "monitor.tour_inspection.inspect_task",
          "title": "巡检任务管理",
          "url": "/hengwei/web/tour_inspection/inspect_tasks"
        },
        {
          "uid": "monitor.tour_inspection.inspect_point",
          "title": "巡检点管理",
          "url": "/hengwei/web/tour_inspection/inspect_points"
        },
        {
          "uid": "monitor.tour_inspection.inspect_files",
          "title": "巡检文件管理",
          "url": "/hengwei/web/tour_inspection/inspect_files"
        }
      ]
    },
    {
      "uid": "monitor_management",
      "title": "监控管理",
      "url": "#",
      "icon": "fa-file-video-o",
      "children": [
        {
          "uid": "monitor.all_view",
          "title": "实时监控",
          "url": "/hengwei/web/monitor_browse"
        },
        {
          "uid": "divider",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "monitor_management.device_monitor",
          "title": "设备负载监控",
          "url": "#",
          "children": [
            {
              "uid": "monitor_management.device_cpu_monitor",
              "title": "CPU利用率",
              "url": "#"
            },
            {
              "uid": "monitor_management.device_mem_monitor",
              "title": "内存利用率",
              "url": "#"
            }
          ]
        },
        {
          "uid": "monitor_management.line_monitor",
          "title": "线路流量监控",
          "url": "#",
          "children": [
            {
              "uid": "monitor_management.line_octets_monitor",
              "title": "上下行流量",
              "url": "#"
            },
            {
              "uid": "monitor_management.line_pkts_monitor",
              "title": "上下行帧流量",
              "url": "#"
            }
          ]
        },
        {
          "uid": "monitor_management.device_topn",
          "title": "设备负载TOPN",
          "url": "#",
          "children": [
            {
              "uid": "monitor_management.device_cpu_topn",
              "title": "CPU利用率",
              "url": "#"
            },
            {
              "uid": "monitor_management.device_mem_topn",
              "title": "内存利用率",
              "url": "#"
            }
          ]
        },
        {
          "uid": "monitor_management.link_topn",
          "title": "线路流量TOPN",
          "url": "#",
          "children": [
            {
              "uid": "monitor_management.link_octets_topn",
              "title": "流量",
              "url": "#"
            },
            {
              "uid": "monitor_management.link_pkts_topn",
              "title": "帧流量",
              "url": "#"
            }
          ]
        },
        {
          "uid": "monitor_management.port_info",
          "title": "设备端口信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.arp_info",
          "title": "设备ARP表信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.route_info",
          "title": "设备路由表信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.address_info",
          "title": "设备IP地址表信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.tcp_connection",
          "title": "设备TCP连接信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.udp_listen",
          "title": "设备UDP监听信息",
          "url": "#"
        },
        {
          "uid": "monitor_management.device_view",
          "title": "设备运行视图",
          "url": "#"
        },
        {
          "uid": "monitor_management.host_view",
          "title": "主机运行视图",
          "url": "#"
        },
        {
          "uid": "monitor_management.db_view",
          "title": "数据库运行视图",
          "url": "#"
        },
        {
          "uid": "monitor_management.mid_view",
          "title": "中间件运行视图",
          "url": "#"
        }
      ]
    },
    {
      "uid": "alert",
      "title": "告警管理",
      "url": "#",
      "icon": "fa-warning",
      "children": [
        {
          "uid": "alert.current_alert",
          "title": "当前告警",
          "url": "#"
        },
        {
          "uid": "alert.log",
          "title": "告警日志",
          "url": "/hengwei/web/alert_events?source=histories"
        },
        {
          "uid": "alert.device_port_log",
          "title": "设备端口日志",
          "url": "/hengwei/web/snmp_traps/for_devices"
        },
        {
          "uid": "alert.rule",
          "title": "告警规则",
          "url": "/hengwei/web/alarm_rules"
        },
        {
          "uid": "alert.composited_rule",
          "title": "组合告警",
          "url": "/hengwei/web/alarm_composited_rules"
        },
        {
          "uid": "alert.composite_log",
          "title": "组合告警日志",
          "url": "/hengwei/web/composite_alert_events?source=histories"
        },
        {
          "uid": "alert.baseline",
          "title": "基线数据管理",
          "url": "/hengwei/web/baselines"
        },
        {
          "uid": "alert.log_alert",
          "title": "日志告警规则",
          "url": "/hengwei/web/log_alarm_rules"
        },
        {
          "uid": "alert.custom_detect_rule",
          "title": "自定义检测告警",
          "url": "/hengwei/web/custom_inpect_alert_rules"
        },
        {
          "uid": "alert.error_list",
          "title": "规则运行日志",
          "url": "/hengwei/web/alarm_rules/error_list"
        }
      ]
    },
    {
      "uid": "resource",
      "title": "资源管理",
      "url": "#",
      "icon": "fa-hdd-o",
      "children": [
        {
          "uid": "resource.device",
          "title": "设备管理",
          "url": "/hengwei/web/network_devices"
        },
        {
          "uid": "resource.link",
          "title": "线路管理",
          "url": "/hengwei/web/network_links"
        },
        {
          "uid": "resource.composite_port",
          "title": "组合端口管理",
          "url": "/hengwei/web/composite_ports"
        },
        {
          "uid": "resource.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "resource.host",
          "title": "主机管理",
          "url": "/hengwei/web/service_hosts"
        },
        {
          "uid": "resource.database",
          "title": "数据库管理",
          "url": "/hengwei/web/service_databases"
        },
        {
          "uid": "resource.midware",
          "title": "中间件管理",
          "url": "/hengwei/web/service_midwares"
        },
        {
          "uid": "resource.app",
          "title": "标准应用管理",
          "url": "/hengwei/web/std_services"
        },
        {
          "uid": "resource.divider3",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "resource.tag_list",
          "title": "Tag列表",
          "url": "/hengwei/web/mos/tags"
        },
        {
          "uid": "resource.run_time",
          "title": "运行时段设置",
          "url": "/hengwei/web/system_settings/mo_run_time"
        },
        {
          "uid": "resource.divider4",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "resource.import",
          "title": "资源导入",
          "url": "/hengwei/web/system_settings/resource_import"
        },
        {
          "uid": "resource.diagnosis",
          "title": "资源诊断",
          "url": "/hengwei/web/system_settings/resource_diagnosis"
        },
        {
          "uid": "ipDevice.distribution",
          "title": "IP设备分布",
          "url": "/hengwei/web/system_settings/ip_device_distribution"
        },
        {
          "uid": "ipService.distribution",
          "title": "IP服务分布",
          "url": "/hengwei/web/system_settings/ip_service_distribution"
        }
      ]
    },
    {
      "uid": "data_management",
      "title": "数据管理",
      "url": "#",
      "icon": "fa-cubes",
      "children": [
        {
          "uid": "topology.search",
          "title": "生成拓扑图",
          "url": "/hengwei/web/topologies/search_params"
        },
        {
          "uid": "topology.report",
          "title": "拓扑生成报告",
          "url": "/hengwei/web/topologies/search_report"
        },
        {
          "uid": "data.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "notifications.list",
          "title": "通知查询",
          "url": "/hengwei/web/notifications/confirmed"
        },
        {
          "uid": "sms.list",
          "title": "短信浏览",
          "url": "/hengwei/web/sms_management/sms_view"
        },
        {
          "uid": "data.divider2",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "harvest.syslog_query",
          "title": "Syslog查询",
          "url": "/hengwei/web/syslogs"
        },
        {
          "uid": "harvest.trap_query",
          "title": "SnmpTrap查询",
          "url": "/hengwei/web/snmp_traps"
        },
        {
          "uid": "data.divider4",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "ip_address.list",
          "title": "地址簿总览",
          "url": "/hengwei/web/ip_addresses"
        },
        {
          "uid": "ip_address.comparison",
          "title": "地址簿对照表",
          "url": "/hengwei/web/ip_addresses/comparisons"
        },
        {
          "uid": "ip_address.setting",
          "title": "地址簿采集设置",
          "url": "/hengwei/web/ip_addresses/sampling_settings"
        },
        {
          "uid": "ip_address.trace_log",
          "title": "地址簿运行日志",
          "url": "/hengwei/web/ip_addresses/trace_log"
        }
      ]
    },
    {
      "uid": "vm_storage",
      "title": "虚拟化&存储",
      "url": "#",
      "icon": "fa-vimeo-square",
      "children": [
        {
          "uid": "vm_storage.vmware_view",
          "title": "虚拟化视图",
          "url": "/hengwei/web/vmware_topologies/default?active=vm_storage&pm=vm_storage.vmware_view"
        },
        {
          "uid": "vm_storage.object_browse",
          "title": "监控一览",
          "url": "/hengwei/web/vmware_vobjects/object_tree/browse"
        },
        {
          "uid": "vm_storage.v_center",
          "title": "虚拟中心管理",
          "url": "/hengwei/web/vmware_vcenters"
        },
        {
          "uid": "vm_storage.esxi",
          "title": "ESX/ESXI主机管理",
          "url": "/hengwei/web/vmware_vcenters/esxi_hosts"
        },
        {
          "uid": "vm_storage.vm",
          "title": "虚拟机管理",
          "url": "/hengwei/web/vmware_vcenters/virtual_machines"
        },
        {
          "uid": "vm_storage.data_store",
          "title": "数据存储管理",
          "url": "/hengwei/web/vmware_vcenters/data_stores"
        },
        {
          "uid": "vm_storage.raid_browse",
          "title": "存储设备一览",
          "url": "/hengwei/web/san_hosts"
        },
        {
          "uid": "vm_storage.raid_capacity",
          "title": "存储容量使用",
          "url": "/hengwei/web/san_hosts/capacity_usage"
        },
        {
          "uid": "vm_storage.raid_topn",
          "title": "存储性能排行",
          "url": "/hengwei/web/san_hosts/perf_topn"
        }
      ]
    },
    {
      "uid": "data_flow",
      "title": "流量分析",
      "url": "#",
      "icon": "fa-filter",
      "children": [
        {
          "uid": "data_flow.live_view",
          "title": "实时流量分析",
          "url": "/hengwei/web/data_flow/live_view"
        },
        {
          "uid": "data_flow.biz_management",
          "title": "业务管理",
          "url": "/hengwei/web/data_flow/bz_objects"
        },
        {
          "uid": "data_flow.biz_analysis",
          "title": "业务流量查看",
          "url": "/hengwei/web/data_flow/biz_live_view"
        }
      ]
    },
    {
      "uid": "common_tools",
      "title": "常用工具",
      "url": "#",
      "icon": "fa-wrench",
      "children": [
        {
          "uid": "common_tools.archive_rule",
          "title": "配置备份规则",
          "url": "/hengwei/web/config_management/archive_rules"
        },
        {
          "uid": "common_tools.archive_report",
          "title": "设备配置查看",
          "url": "/hengwei/web/config_management/archive_report"
        },
        {
          "uid": "common_tools.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "common_tools.device_conf_template",
          "title": "设备配置模板",
          "url": "/hengwei/web/config_management/conf_tempates"
        },
        {
          "uid": "common_tools.tftp_script",
          "title": "TFTP脚本管理",
          "url": "/hengwei/web/config_management/tftp_scripts"
        },
        {
          "uid": "common_tools.divider2",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "common_tools.ping_test",
          "title": "Ping测试",
          "url": "#"
        },
        {
          "uid": "common_tools.snmp_test",
          "title": "Snmp连接测试",
          "url": "#"
        },
        {
          "uid": "common_tools.trace_route",
          "title": "Trace Route",
          "url": "#"
        },
        {
          "uid": "common_tools.dig_test",
          "title": "Dig",
          "url": "#"
        },
        {
          "uid": "common_tools.divider3",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "common_tools.web_console",
          "title": "远程控制台",
          "url": "/hengwei/web/tools/web_console"
        },
        {
          "uid": "common_tools.telnet",
          "title": "Telnet",
          "url": "/hengwei/web/tools/remote_terminal?remote_host=&protocol=telnet"
        },
        {
          "uid": "common_tools.ssh",
          "title": "SSH",
          "url": "/hengwei/web/tools/remote_terminal?remote_host=&protocol=ssh"
        },
        {
          "uid": "common_tools.divider4",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "common_tools.cim",
          "title": "Cim浏览",
          "url": "/hengwei/web/tools/cim_navigator"
        },
        {
          "uid": "common_tools.divider5",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "common_tools.error_list",
          "title": "颜色规则日志",
          "url": "/hengwei/web/system/redis_command_errors"
        }
      ]
    },
    {
      "uid": "report_management",
      "title": "报表管理",
      "url": "#",
      "icon": "fa-list-alt",
      "children": [
        {
          "uid": "report_management.history_rule",
          "title": "历史记录规则",
          "url": "/hengwei/web/histories"
        },
        {
          "uid": "report_management.history_analysis",
          "title": "记录数据查询",
          "url": "/hengwei/web/history_records"
        },
        {
          "uid": "report_management.history_flux",
          "title": "流量数据查询",
          "url": "/hengwei/web/history_records/flux"
        },
        {
          "uid": "report_management.history_error_list",
          "title": "规则运行日志",
          "url": "/hengwei/web/histories/error_list"
        },
        {
          "uid": "report_management.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "report_management.report_rule",
          "title": "报表规则管理",
          "url": "/hengwei/web/report_management/rules"
        },
        {
          "uid": "report_management.report_browse",
          "title": "生成报表查看",
          "url": "/hengwei/web/report_management/prints"
        },
        {
          "uid": "report_management.custom_kpi",
          "title": "自定义统计数据",
          "url": "/hengwei/web/report_management/custom_kpis"
        },
        {
          "uid": "report_management.divider2",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "report_management.default_report_rule",
          "title": "预置报表管理",
          "url": "/hengwei/web/report_management/default_rules"
        },
        {
          "uid": "report_management.default_report_browse",
          "title": "预置报表查看",
          "url": "/hengwei/web/report_management/prints?list_default=true"
        }
      ]
    },
    {
      "uid": "system",
      "title": "系统管理",
      "url": "#",
      "icon": "fa-cog",
      "children": [
        {
          "uid": "system.divider1",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "system.online_users",
          "title": "当前在线用户",
          "url": "/hengwei/web/users/online_users"
        },
        {
          "uid": "system.operation_logs",
          "title": "用户操作日志",
          "url": "/hengwei/web/system/operation_logs"
        },
        {
          "uid": "system.divider2",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "mc.NotificationGroups.Index",
          "title": "通知组",
          "permission": "mc.NotificationGroups",
          "url": "/hengwei/mc_dev/notification_groups"
        },
        {
          "uid": "system.divider3",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "system.engine_nodes",
          "title": "采集引擎管理",
          "url": "/hengwei/web/system/engine_nodes"
        },
        {
          "uid": "system.divider4",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "system.mail_server",
          "title": "服务器设置",
          "url": "/hengwei/web/system_settings/mail_server"
        },
        {
          "uid": "system.divider5",
          "title": "divider",
          "url": "#"
        },
        {
          "uid": "system.image_manage",
          "title": "图元管理",
          "url": "/hengwei/web/custom_resources/image_view?path=images%2Ftopology"
        },
        {
          "uid": "system.script_manage",
          "title": "脚本管理",
          "url": "/hengwei/web/tools/script_manage"
        },
        {
          "uid": "system.sql_script_manage",
          "title": "数据库脚本管理",
          "url": "/hengwei/web/tools/sql_script_manage"
        }
      ]
    }
  ]
}`

	//
	//
	//
	//
	//
	///
	//
	//
	//
	menusText := `
[
  {
    "uid": "home",
    "title": "首页",
    "url": "/hengwei/web",
    "icon": "fa-home"
  },
  {
    "uid": "am.home",
    "title": "资产管理",
    "url": "/hengwei/am",
    "icon": "fa-archive",
    "classes": "special_link"
  },
  {
    "uid": "itsm.home",
    "title": "服务管理",
    "url": "/hengwei/itsm",
    "icon": "fa-asterisk",
    "classes": "special_link"
  },
  {
    "uid": "report_management",
    "title": "报表管理",
    "url": "#",
    "icon": "fa-list-alt"
  },
  {
    "uid": "system",
    "title": "系统管理",
    "url": "#",
    "icon": "fa-cog",
    "children": [
      {
        "uid": "system.divider1",
        "title": "divider",
        "url": "#"
      },
      {
        "uid": "HengweiUsers",
        "title": "用户管理",
        "url": "/hengwei/um/hengweiusers/index",
        "icon": "fa-user"
      },
      {
        "uid": "HengweiUserGroups",
        "title": "用户组管理",
        "url": "/hengwei/um/hengweiusergroups/index?groupId=0",
        "icon": "fa-users"
      },
      {
        "uid": "HengweiRoles",
        "title": "角色管理",
        "url": "/hengwei/um/hengweiroles/index",
        "icon": "fa-user-circle"
      },
      {
        "uid": "HengweiPermissionGroups",
        "title": "权限管理",
        "url": "/hengwei/um/hengweipermissiongroups/index?groupId=0",
        "icon": "fa-archive"
      },
      {
        "uid": "system.operation_logs",
        "title": "用户操作日志",
        "url": "/hengwei/web/system/operation_logs"
      },
      {
        "uid": "system.divider2",
        "title": "divider",
        "url": "#"
      },
      {
        "uid": "mc.NotificationGroups.Index",
        "title": "通知组",
        "permission": "mc.NotificationGroups",
        "url": "/hengwei/mc_dev/notification_groups"
      },
      {
        "uid": "mc.AutoBackup.Index",
        "title": "备份数据",
        "permission": "mc.AutoBackup",
        "url": "/hengwei/mc_dev/autobackup/index?active="
      },
      {
        "uid": "system.divider3",
        "title": "divider",
        "url": "#"
      },
      {
        "uid": "system.engine_nodes",
        "title": "采集引擎管理",
        "url": "/hengwei/web/system/engine_nodes"
      }
    ]
  }
]`

	var menuList []toolbox.Menu
	err := json.Unmarshal([]byte(menusText), &menuList)
	if err != nil {
		t.Error(err)
		return
	}

	var value = struct {
		Applications map[string][]toolbox.Menu `json:"applications"`
		Layouts      map[string][]LayoutItem   `json:"layout"`
	}{}
	err = json.Unmarshal([]byte(rawText), &value)
	if err != nil {
		t.Error(err)
		return
	}

	// byApps := map[string][]toolbox.Menu{}
	// for name, app := range value.Applications {
	// 	//t.Log(name, len(app))
	// 	byApps[name] = toMenuTree(app)
	// }

	layout := &layoutImpl{mainLayout: value.Layouts["default"]}
	//t.Log(layout.mainLayout)
	resultList, err := layout.Generate(value.Applications)
	if err != nil {
		t.Error(err)
		return
	}

	if !isSameMenuArray(resultList, menuList) {
		t.Error("结果不同")
		dumpMenus(t, resultList, menuList)
	}

	t.Log(menuList[1].Classes)
}
