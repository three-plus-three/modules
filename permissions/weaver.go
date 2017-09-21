package permissions

import (
	"strconv"
	"sync"

	"github.com/three-plus-three/modules/hub"
	hub_engine "github.com/three-plus-three/modules/hub/engine"
)

//go:generate genny -pkg=permissions -in=../weaver/client.go -out=client-gen.go gen "ValueType=*PermissionData"
//go:generate genny -pkg=permissions -in=../weaver/server.go -out=server-gen.go gen "WeaveType=*PermissionData"

const PermissionEventName = "permissions.changed"

func NewWeaver(core *hub_engine.Core) (Weaver, error) {
	weaver := &memWeaver{core: core}
	return weaver, nil
}

type memWeaver struct {
	core *hub_engine.Core

	mu       sync.RWMutex
	all      PermissionData
	byGroups map[string]*PermissionData
}

func (weaver *memWeaver) Update(app string, data *PermissionData) error {
	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	weaver.byGroups[app] = data

	if len(weaver.all.Groups) > 0 {
		weaver.all.Groups = weaver.all.Groups[:0]
	}
	if len(weaver.all.Permissions) > 0 {
		weaver.all.Permissions = weaver.all.Permissions[:0]
	}
	if len(weaver.all.Tags) > 0 {
		weaver.all.Tags = weaver.all.Tags[:0]
	}
	for _, group := range weaver.byGroups {
		appendPermissionData(&weaver.all, group)
	}
	weaver.core.CreateTopicIfNotExists(PermissionEventName).
		Send(hub.Message([]byte(strconv.Itoa(len(weaver.all.Permissions)))))
	return nil
}

func (weaver *memWeaver) Generate() (*PermissionData, error) {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	return &weaver.all, nil
}

func isSubset(allItems, subset *PermissionData) bool {
	if !containGroups(allItems.Groups, subset.Groups) {
		return false
	}
	if !containPermissions(allItems.Permissions, subset.Permissions) {
		return false
	}
	return containTags(allItems.Tags, subset.Tags)
}

func containGroups(allItems, items []Group) bool {
	if len(allItems) < len(items) {
		return false
	}

	for _, item := range items {
		foundIdx := -1
		for idx, a := range allItems {
			if a.Name == item.Name {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return false
		}

		if allItems[foundIdx].Description != item.Description {
			return false
		}
		if !containsString(allItems[foundIdx].PermissionIDs, item.PermissionIDs) {
			return false
		}
		if !containsString(allItems[foundIdx].PermissionTags, item.PermissionTags) {
			return false
		}
		if !containGroups(allItems[foundIdx].Children, item.Children) {
			return false
		}
	}
	return true
}

func containsString(allItems, items []string) bool {
	if len(allItems) < len(items) {
		return false
	}

	for _, item := range items {
		foundIdx := -1
		for idx, a := range allItems {
			if a == item {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return false
		}
	}
	return true
}

func containPermissions(allItems, items []Permission) bool {
	if len(allItems) < len(items) {
		return false
	}

	for _, item := range items {
		foundIdx := -1
		for idx, a := range allItems {
			if a.ID == item.ID {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return false
		}

		if allItems[foundIdx].Name != item.Name {
			return false
		}
		if allItems[foundIdx].Description != item.Description {
			return false
		}
		if !containsString(allItems[foundIdx].Tags, item.Tags) {
			return false
		}
	}
	return true
}

func containTags(allItems, items []Tag) bool {
	if len(allItems) < len(items) {
		return false
	}

	for _, item := range items {
		foundIdx := -1
		for idx, a := range allItems {
			if a.ID == item.ID {
				foundIdx = idx
				break
			}
		}
		if foundIdx < 0 {
			return false
		}

		if allItems[foundIdx].Name != item.Name {
			return false
		}
		if allItems[foundIdx].Description != item.Description {
			return false
		}
		if !containTags(allItems[foundIdx].Children, item.Children) {
			return false
		}
	}
	return true
}
