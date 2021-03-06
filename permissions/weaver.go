package permissions

import (
	"strconv"
	"sync"

	"github.com/three-plus-three/modules/hub"
)

//go:generate genny -pkg=permissions -in=../weaver/client.go -out=client-gen.go gen "ValueType=*PermissionData"
//go:generate genny -pkg=permissions -in=../weaver/server.go -out=server-gen.go gen "WeaveType=*PermissionData"

const PermissionEventName = "permissions.changed"

func NewWeaver(sendEvent func(hub.Message)) (Weaver, error) {
	weaver := &memWeaver{sendEvent: sendEvent,
		byGroups: map[string]*PermissionData{}}
	return weaver, nil
}

type memWeaver struct {
	sendEvent func(hub.Message)

	mu       sync.RWMutex
	all      PermissionData
	byGroups map[string]*PermissionData
}

func (weaver *memWeaver) Stats() interface{} {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	apps := map[string]interface{}{}
	for k, v := range weaver.byGroups {
		apps[k] = v
	}

	return map[string]interface{}{
		"applications": apps,
	}
}

func (weaver *memWeaver) Update(app string, data *PermissionData) error {
	weaver.mu.Lock()
	defer weaver.mu.Unlock()
	if data == nil || (len(data.Groups) == 0 &&
		len(data.Permissions) == 0 &&
		len(data.Tags) == 0) {
		old, ok := weaver.byGroups[app]
		if !ok {
			return nil
		}
		if len(old.Groups) == 0 &&
			len(old.Permissions) == 0 &&
			len(old.Tags) == 0 {
			return nil
		}
		delete(weaver.byGroups, app)

	} else {
		weaver.byGroups[app] = data
	}

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

	event := hub.CreateDataMessage([]byte(strconv.Itoa(len(weaver.all.Permissions))))
	weaver.sendEvent(event)
	return nil
}

func (weaver *memWeaver) Generate(ctx string) (*PermissionData, error) {
	weaver.mu.RLock()
	defer weaver.mu.RUnlock()
	return &weaver.all, nil
}

func isSame(allItems, subset *PermissionData) bool {
	return isSubset(allItems, subset)
}

func IsSubset(allItems, subset *PermissionData) bool {
	return isSubset(allItems, subset)
}

func isSubset(allItems, subset *PermissionData) bool {
	if allItems == nil {
		if subset == nil {
			return true
		}

		return len(subset.Groups) == 0 &&
			len(subset.Permissions) == 0 &&
			len(subset.Tags) == 0
	}

	if subset == nil {
		return true
	}

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

var MergePermissionData = appendPermissionData

func appendPermissionData(all, data *PermissionData) {
	if data == nil {
		return
	}

	if len(data.Permissions) > 0 {
		all.Permissions = append(all.Permissions, data.Permissions...)
	}
	if len(data.Groups) > 0 {
		all.Groups = appendGroups(all.Groups, data.Groups)
	}
	if len(data.Tags) > 0 {
		all.Tags = append(all.Tags, data.Tags...)
	}
}

func appendGroups(allGroups, groups []Group) []Group {
	for _, group := range groups {
		found := false
		for idx := range allGroups {
			if allGroups[idx].Name == group.Name {
				found = true

				allGroups[idx].PermissionIDs = mergeStrings(allGroups[idx].PermissionIDs, group.PermissionIDs)
				allGroups[idx].PermissionTags = mergeStrings(allGroups[idx].PermissionTags, group.PermissionTags)
				allGroups[idx].Children = appendGroups(allGroups[idx].Children, group.Children)
			}
		}
		if !found {
			allGroups = append(allGroups, group)
		}
	}
	return allGroups
}

func mergeStrings(a, b []string) []string {
	for _, s := range b {
		found := false
		for _, v := range a {
			if v == s {
				found = true
			}
		}
		if !found {
			a = append(a, s)
		}
	}
	return a
}
