package menus

import (
	"cn/com/hengwei/commons/env_tests"
	"testing"

	"github.com/go-xorm/xorm"
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

	for tidx, test := range tests {
		if err := modelEngine.DropTables(&Menu{}); err != nil {
			t.Error(tidx, test.name, err)
			return
		}

		if err := modelEngine.CreateTables(&Menu{}); err != nil {
			t.Error(tidx, test.name, err)
			return
		}

		weaver, err := NewWeaver(nil, &DB{Engine: modelEngine})
		if err != nil {
			t.Error(tidx, test.name, err)
			return
		}
		for idx, step := range test.steps {
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

var tests = []struct {
	name  string
	steps []struct {
		name    string
		app     string
		value   []toolbox.Menu
		results []toolbox.Menu
	}
}{}
