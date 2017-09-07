package permissions

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrCacheInvalid = errors.New("permission cache is invald")
var ErrTagNotFound = errors.New("permission tag is not found")
var ErrPermissionNotFound = errors.New("permission is not found")

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

type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,emitempty"`

	Children []Tag `json:"children,omitempty"`
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
			if tag == all[i].Tags[j] {
				filterPermissions = append(filterPermissions, all[i])
			}
		}
	}
	return filterPermissions, nil
}

//获取所有 tags
func GetPermissionTags() ([]Tag, error) {
	return permissionsCache.PermissionTags()
}

func GetPermissionTagByID(id string) (*Tag, error) {
	return permissionsCache.GetPermissionTagByID(id)
}

//获取权限
func GetPermissions() ([]Permission, error) {
	return permissionsCache.Permissions()
}

func GetPermissionByID(id string) (*Permission, error) {
	return permissionsCache.GetPermissionByID(id)
}

//获取权限组
func GetDefaultPermissionGroups() ([]Group, error) {
	return permissionsCache.PermissionGroups()
}

//缓存
var permissionsCache PermissionCache

//缓存
type permissionCacheData struct {
	permissions    []Permission
	tags           []Tag
	groups         []Group
	tagByID        map[string]*Tag
	permissionByID map[string]*Permission
	saveTime       int64
}

type PermissionCache struct {
	value atomic.Value

	mu        sync.Mutex
	isLoading int32
}

func (cache *PermissionCache) dataAsync() *permissionCacheData {
	o := cache.value.Load()
	if o == nil {
		return nil
	}

	d, ok := o.(*permissionCacheData)
	if !ok {
		return nil
	}
	if (d.saveTime + 60) < time.Now().Unix() {
		return nil
	}
	return d
}

func (cache *PermissionCache) load() (*permissionCacheData, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	if d := cache.dataAsync(); d != nil {
		return d, nil
	}

	var all PermissionData
	for _, p := range privoders {
		data, err := p.Get()
		if err != nil {
			return nil, err
		}
		appendPermissionData(&all, data)
	}
	return cache.Save(all.Permissions, all.Tags, all.Groups), nil
}

func (cache *PermissionCache) data() (*permissionCacheData, error) {
	o := cache.value.Load()
	if o == nil {
		return cache.load()
	}

	d, ok := o.(*permissionCacheData)
	if !ok {
		return cache.load()
	}

	if (d.saveTime + 60) < time.Now().Unix() {
		if atomic.CompareAndSwapInt32(&cache.isLoading, 0, 1) {
			go func() {
				defer atomic.StoreInt32(&cache.isLoading, 0)
				cache.load()
			}()
		}
	}
	return d, nil
}

//从缓存中获取权限对象
func (cache *PermissionCache) Permissions() ([]Permission, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.permissions, nil
}

//按 ID 从缓存中获取权限对象
func (cache *PermissionCache) GetPermissionByID(id string) (*Permission, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	perm := d.permissionByID[id]
	if perm == nil {
		return nil, ErrPermissionNotFound
	}
	return perm, nil
}

//从缓存中获取权限对象
func (cache *PermissionCache) PermissionTags() ([]Tag, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.tags, nil
}

//按 ID 从缓存中获取权限对象
func (cache *PermissionCache) GetPermissionTagByID(id string) (*Tag, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	tag := d.tagByID[id]
	if tag == nil {
		return nil, ErrTagNotFound
	}
	return tag, nil
}

//从缓存中获取权限对象
func (cache *PermissionCache) PermissionGroups() ([]Group, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.groups, nil
}

//缓存过期
func (cache *PermissionCache) Invalid() {
	cache.Save(nil, nil, nil)
}

func removeTags(tagsInPermissions map[string]struct{}, tags []Tag) {
	for _, tag := range tags {
		delete(tagsInPermissions, tag.ID)
		removeTags(tagsInPermissions, tag.Children)
	}
}

func addTags(tagByID map[string]*Tag, tags []Tag) {
	for idx := range tags {
		addTags(tagByID, tags[idx].Children)
		tagByID[tags[idx].ID] = &tags[idx]
	}
}

//将权限对象存入缓存中
func (cache *PermissionCache) Save(permissions []Permission, tags []Tag, groups []Group) *permissionCacheData {
	d := &permissionCacheData{
		saveTime:       time.Now().Unix(),
		permissions:    permissions,
		tags:           tags,
		groups:         groups,
		permissionByID: map[string]*Permission{},
		tagByID:        map[string]*Tag{},
	}
	if permissions == nil {
		d.saveTime = 0
	}

	tagsInPermissions := map[string]struct{}{}
	for idx := range d.permissions {
		d.permissionByID[d.permissions[idx].ID] = &d.permissions[idx]

		for _, tag := range d.permissions[idx].Tags {
			tagsInPermissions[tag] = struct{}{}
		}
	}

	removeTags(tagsInPermissions, d.tags)
	for tag := range tagsInPermissions {
		d.tags = append(d.tags, Tag{
			ID:   tag,
			Name: tag,
		})
	}
	addTags(d.tagByID, d.tags)

	cache.value.Store(d)
	return d
}

var privoders = map[string]PermissionProvider{}

//注册方法
func RegisterPermissions(name string, privoder PermissionProvider) {
	if privoder == nil {
		panic("provider is nil")
	}
	if _, ok := privoders[name]; ok {
		panic("privoder '" + name + "' is already exists.")
	}
	privoders[name] = privoder
}

type PermissionData struct {
	Permissions []Permission `json:"permissions"`
	Groups      []Group      `json:"groups"`
	Tags        []Tag        `json:"tags"`
}

type PermissionProvider interface {
	Get() (*PermissionData, error)
}

type PermissionProviderFunc func() (*PermissionData, error)

func (f PermissionProviderFunc) Get() (*PermissionData, error) {
	if f == nil {
		return nil, nil
	}
	return f()
}

func appendPermissionData(all, data *PermissionData) {
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
