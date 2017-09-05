package permissions

import (
	"strings"
	"time"
)

type Group struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Children       []Group  `json:"children,omitempty"`
	PermissionIDs  []string `json:"permissions,omitempty"`
	PermissionTags []string `json:"tags,omitempty"`
}

type Permission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,emitempty"`
	Tags        []string `json:"tags,emitempty"`
}

//过滤后的权限对象
func GetPermissionsByTag(tag string) ([]Permission, error) {
	all, err := GetPermissions()
	var filterPermissions []Permission
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(all); i++ {
		for j := 0; j < len(all[i].Tags); j++ {
			if strings.EqualFold(tag, all[i].Tags[j]) {
				filterPermissions = append(filterPermissions, all[i])
			}
		}
	}
	return filterPermissions, nil
}

//获取所有 tags
func GetPermissionTags() ([]string, error) {
	all, err := GetPermissions()
	if err != nil {
		return nil, err
	}
	tagsByName := map[string]struct{}{}
	for i := 0; i < len(all); i++ {
		for j := 0; j < len(all[i].Tags); j++ {
			tagsByName[all[i].Tags[j]] = struct{}{}
		}
	}
	tags := make([]string, 0, len(tagsByName))
	for key := range tagsByName {
		tags = append(tags, key)
	}

	return tags, nil
}

//获取权限
func GetPermissions() ([]Permission, error) {
	permissionsCache.Invalid()
	var all []Permission = permissionsCache.Get()
	if len(all) != 0 {
		return all, nil
	}
	for _, p := range privoders {
		permissions, err := p.GetPermissions()
		if err != nil {
			return nil, err
		}
		all = append(all, permissions...)
	}
	permissionsCache.Save(all)
	return all, nil
}

//获取权限组
func GetDefaultPermissionGroups() ([]Group, error) {
	var allGroups []Group
	for _, p := range privoders {
		groups, err := p.GetGroups()
		if err != nil {
			return nil, err
		}

		allGroups = appendGroups(allGroups, groups)
	}
	return allGroups, nil
}

//缓存
var permissionsCache PermissionsCache

//缓存
type PermissionsCache struct {
	permissions []Permission
	saveTime    int64
}

//从缓存中获取权限对象
func (cache *PermissionsCache) Get() []Permission {
	return cache.permissions
}

//缓存过期
func (cache *PermissionsCache) Invalid() {
	if (time.Now().Unix() - cache.saveTime) > 60*10 {
		cache.permissions = nil
	}
}

//将权限对象存入缓存中
func (cache *PermissionsCache) Save(permissions []Permission) {
	cache.saveTime = time.Now().Unix()
	cache.permissions = permissions
}

//注册方法
func RegisterPermissions(privoder PermissionProvider) {
	if privoder == nil {
		panic("provider is nil")
	}
	privoders = append(privoders, privoder)
}

var privoders []PermissionProvider

type PermissionProvider interface {
	Name() string
	GetPermissions() ([]Permission, error)
	GetGroups() ([]Group, error)
}

type PermissionProviderFunc struct {
	ProviderName string
	Permissions  func() ([]Permission, error)
	Groups       func() ([]Group, error)
}

func (f PermissionProviderFunc) Name() string {
	return f.ProviderName
}

func (f PermissionProviderFunc) GetPermissions() ([]Permission, error) {
	if f.Permissions == nil {
		return nil, nil
	}
	return f.Permissions()
}

func (f PermissionProviderFunc) GetGroups() ([]Group, error) {
	if f.Groups == nil {
		return nil, nil
	}
	return f.Groups()
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
