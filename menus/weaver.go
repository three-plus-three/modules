package menus

import (
	"strconv"
	"sync"

	"github.com/three-plus-three/modules/hub"
	hub_engine "github.com/three-plus-three/modules/hub/engine"
	"github.com/three-plus-three/modules/toolbox"
)

type menuWeaver struct {
	core *hub_engine.Core

	mu       sync.RWMutex
	menuList []toolbox.Menu
}

func (weaver *menuWeaver) Update(group string, data interface{}) error {
	// if len(data) != 0 {
	// 	err = srv.weaver.Update(group, data)
	// 	if err != nil {
	// 		http.Error(w, err.Error(), http.StatusBadRequest)
	// 		return
	// 	}
	// }
	menuList := data.([]toolbox.Menu)
	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.menuList = menuList

	weaver.core.CreateTopicIfNotExists("menus.changed").
		Send(hub.Message([]byte(strconv.Itoa(len(menuList)))))
	return nil
}

func (weaver *menuWeaver) Generate() (interface{}, error) {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	return weaver.menuList, nil
}
