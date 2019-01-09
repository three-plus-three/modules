package menus

import (
	"cn/com/hengwei/commons/env_tests"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aryann/difflib"
	"github.com/three-plus-three/modules/hub/engine"
	"github.com/three-plus-three/modules/toolbox"
)

func TestMenuSimple(t *testing.T) {
	env := env_tests.Clone(nil)
	// dataDrv, dataURL := env.Db.Models.Url()
	// modelEngine, err := xorm.NewEngine(dataDrv, dataURL)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// modelEngine.ShowSQL()

	core, _ := engine.NewCore(&engine.Options{})

	for tidx, test := range tests {
		// if err := modelEngine.DropTables(&Menu{}); err != nil {
		// 	t.Error(tidx, test.name, err)
		// 	return
		// }

		// if err := modelEngine.CreateTables(&Menu{}); err != nil {
		// 	t.Error(tidx, test.name, err)
		// 	return
		// }

		logger := log.New(os.Stderr, "[menu] ", log.LstdFlags)
		weaver, err := NewWeaver(logger, env, core, nil, test.layout, nil, nil)
		if err != nil {
			t.Error(tidx, test.name, err)
			return
		}
		for idx, step := range test.steps {
			if step.isRestart {
				weaver, err = NewWeaver(logger, env, core, nil, test.layout, nil, nil)
				if err != nil {
					t.Error(tidx, test.name, err)
					return
				}
			}

			if err := weaver.Update(step.app, step.value); err != nil {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", err)
				return
			}

			results, err := weaver.Generate("")
			if err != nil {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", err)
				return
			}

			if !isSameMenuArray(results, step.results) {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", "result is diff - ")
				t.Logf("excepted is %#v", step.results)
				t.Logf("actual   is %#v", results)

				bs, _ := json.MarshalIndent(weaver.Stats(), "", "  ")
				t.Log(string(bs))
			}
		}
	}
}

type testStep struct {
	isRestart bool
	name      string
	app       string
	value     []toolbox.Menu
	results   []toolbox.Menu
}

var menuTest1 = []toolbox.Menu{
	{
		Title: "m1",
		URL:   "#",
	},
	{
		Title: "m2",
		URL:   "#",
	},
	{
		Title: "m3",
		URL:   "#",
		Children: []toolbox.Menu{
			{
				Title: "m3_1",
				URL:   "#",
			},
		},
	},
}

var test2 = []testStep{
	{app: "a1_1",
		value: []toolbox.Menu{
			{
				UID:   "1",
				Title: "m1",
				URL:   "#",
			},
			{
				UID:   "2",
				Title: "m2",
				URL:   "#",
			},
			{
				UID:   "3",
				Title: "m3",
				URL:   "#",
				Children: []toolbox.Menu{
					{
						UID:   "3_1",
						Title: "m3_1",
						URL:   "#",
					},
				},
			},
		},
		results: []toolbox.Menu{
			{
				UID:   "1",
				Title: "m1",
				URL:   "#",
			},
			{
				UID:   "2",
				Title: "m2",
				URL:   "#",
			},
			{
				UID:   "3",
				Title: "m3",
				URL:   "#",
				Children: []toolbox.Menu{
					{
						UID:   "3_1",
						Title: "m3_1",
						URL:   "#",
					},
				},
			},
		},
	},
	{isRestart: true,
		app: "a1_2",
		results: []toolbox.Menu{
			{
				UID:   "1",
				Title: "m1",
				URL:   "#",
			},
			{
				UID:   "2",
				Title: "m2",
				URL:   "#",
			},
			{
				UID:   "3",
				Title: "m3",
				URL:   "#",
				Children: []toolbox.Menu{
					{
						UID:   "3_1",
						Title: "m3_1",
						URL:   "#",
					},
				},
			},
		},
	},
}

var tests = []struct {
	layout Layout
	name   string
	steps  []testStep
}{
	{
		name:   "a1",
		layout: &simpleLayout{},
		steps: []testStep{
			{app: "a1_1",
				value:   menuTest1,
				results: menuTest1,
			},
			{app: "a1_1",
				value:   menuTest1,
				results: menuTest1,
			},
		},
	},
	{
		name:   "a2",
		layout: &simpleLayout{},
		steps:  test2,
	},
}

func TestMenuSimple2(t *testing.T) {
	env := env_tests.Clone(nil)
	// dataDrv, dataURL := env.Db.Models.Url()
	// modelEngine, err := xorm.NewEngine(dataDrv, dataURL)
	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
	// modelEngine.ShowSQL()

	for _, test := range []struct {
		filename string
		layout   string
	}{
		{filename: "tests\\test1.json", layout: "am"},
		{filename: "tests\\test2.json"},
		{filename: "tests\\test3.json", layout: "vpn_management"},
	} {
		t.Run(test.filename, func(t *testing.T) {
			core, _ := engine.NewCore(&engine.Options{})

			logger := log.New(os.Stderr, "[menu] ", log.LstdFlags)
			weaver := &menuWeaver{Logger: logger, env: env, core: core, layouts: nil}

			data, err := ioutil.ReadFile(test.filename)
			if err != nil {
				t.Error(err)
				return
			}

			var udata = &struct {
				Apps    map[string][]toolbox.Menu `json:"applications"`
				Layouts map[string][]LayoutItem   `json:"layout"`
				Results []toolbox.Menu            `json:"menuList"`
			}{}
			err = json.Unmarshal(data, &udata)
			if err != nil {
				t.Error(err)
				return
			}

			weaver.byApplications = udata.Apps
			weaver.layouts = map[string]Layout{}
			for key, value := range udata.Layouts {
				weaver.layouts[key] = &layoutImpl{value}
			}

			weaver.layout = weaver.layouts["default"]

			if err := weaver.Init(); err != nil {
				t.Error(err)
				return
			}
			weaver.byApplications = udata.Apps
			weaver.menuList = nil
			weaver.menuListByLayout = map[string][]toolbox.Menu{}

			results, err := weaver.Generate(test.layout)
			if err != nil {
				t.Error(err)
				return
			}

			if !isSameMenuArray(results, udata.Results) {
				t.Error("not same")
				var s1, s2 strings.Builder
				toolbox.FormatMenus(&s1, nil, udata.Results, 0, true)
				toolbox.FormatMenus(&s2, nil, results, 0, true)

				seq1 := strings.Split(s1.String(), "\n")
				seq2 := strings.Split(s2.String(), "\n")
				records := difflib.Diff(seq1, seq2)
				for idx := range records {
					t.Log(records[idx])
				}

				// toolbox.FormatMenus(os.Stdout, nil, results, 0, true)

				return
			}
		})
	}
}
