package menus

import (
	"cn/com/hengwei/commons/env_tests"
	"testing"

	"github.com/go-xorm/xorm"
	"github.com/three-plus-three/modules/hub/engine"
	"github.com/three-plus-three/modules/toolbox"
)

func TestMeneSimple(t *testing.T) {
	env := env_tests.Clone(nil)
	dataDrv, dataURL := env.Db.Models.Url()
	modelEngine, err := xorm.NewEngine(dataDrv, dataURL)
	if err != nil {
		t.Error(err)
		return
	}
	modelEngine.ShowSQL()

	core, _ := engine.NewCore(&engine.Options{})

	for tidx, test := range tests {
		if err := modelEngine.DropTables(&Menu{}); err != nil {
			t.Error(tidx, test.name, err)
			return
		}

		if err := modelEngine.CreateTables(&Menu{}); err != nil {
			t.Error(tidx, test.name, err)
			return
		}

		weaver, err := NewWeaver(core, &DB{Engine: modelEngine})
		if err != nil {
			t.Error(tidx, test.name, err)
			return
		}
		for idx, step := range test.steps {
			if step.isRestart {
				weaver, err = NewWeaver(core, &DB{Engine: modelEngine})
				if err != nil {
					t.Error(tidx, test.name, err)
					return
				}
			}

			if err := weaver.Update(step.app, step.value); err != nil {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", err)
				return
			}

			results, err := weaver.Generate()
			if err != nil {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", err)
				return
			}

			if !toolbox.IsSameMenuArray(results, step.results) {
				t.Error("[", tidx, test.name, "] [", idx, step, "]", "result is diff - ")
				t.Logf("excepted is %#v", step.results)
				t.Logf("actual   is %#v", results)
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

var tests = []struct {
	name  string
	steps []testStep
}{
	{
		name: "a1",
		steps: []testStep{
			{app: "a1_1",
				value: []toolbox.Menu{
					{
						Category: "",
						Name:     "1",
						Title:    "m1",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "2",
						Title:    "m2",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "3",
						Title:    "m3",
						URL:      "#",
						Children: []toolbox.Menu{
							{
								Category: "",
								Name:     "3_1",
								Title:    "m3_1",
								URL:      "#",
							},
						},
					},
				},
				results: []toolbox.Menu{
					{
						Category: "",
						Name:     "1",
						Title:    "m1",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "2",
						Title:    "m2",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "3",
						Title:    "m3",
						URL:      "#",
						Children: []toolbox.Menu{
							{
								Category: "",
								Name:     "3_1",
								Title:    "m3_1",
								URL:      "#",
							},
						},
					},
				},
			},
			{isRestart: true,
				app: "a1_2",
				results: []toolbox.Menu{
					{
						Category: "",
						Name:     "1",
						Title:    "m1",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "2",
						Title:    "m2",
						URL:      "#",
					},
					{
						Category: "",
						Name:     "3",
						Title:    "m3",
						URL:      "#",
						Children: []toolbox.Menu{
							{
								Category: "",
								Name:     "3_1",
								Title:    "m3_1",
								URL:      "#",
							},
						},
					},
				},
			},
		},
	},
}