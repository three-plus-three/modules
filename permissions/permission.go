package permissions

import (
	"strings"
	"time"
)

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
	privoders = append(privoders, privoder)
}

var privoders []PermissionProvider

type PermissionProvider interface {
	GetPermissions() ([]Permission, error)
}

type PermissionGetFunc func() ([]Permission, error)

func (f PermissionGetFunc) GetPermissions() ([]Permission, error) {
	return f()
}
