package users

import (
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/runner-mei/log"
	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/urlutil"
)

// 常用的错误
var (
	ErrUnauthorized       = errors.NewApplicationError(http.StatusUnauthorized, "user is unauthorized")
	ErrCacheInvalid       = errors.New("permission cache is invald")
	ErrTagNotFound        = errors.New("permission tag is not found")
	ErrPermissionNotFound = errors.New("permission is not found")
	ErrAlreadyClosed      = errors.New("server is closed")
)

// Group 缺省组信息
type Group struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Children       []Group  `json:"children,omitempty"`
	PermissionIDs  []string `json:"permissions,omitempty"`
	PermissionTags []string `json:"tags,omitempty"`
}

// Permission 缺省权限对象
type Permission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,emitempty"`
	Tags        []string `json:"tags,emitempty"`
}

// Tag 标签对象
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,emitempty"`

	Children []Tag `json:"children,omitempty"`
}

// GetPermissionsByTag 过滤后的权限对象
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

// GetPermissionTags 获取所有 tags
func GetPermissionTags() ([]Tag, error) {
	return permissionsCache.PermissionTags()
}

// GetPermissionTagByID 获取指定的 tag
func GetPermissionTagByID(id string) (*Tag, error) {
	return permissionsCache.GetPermissionTagByID(id)
}

// GetPermissions 获取权限
func GetPermissions() ([]Permission, error) {
	return permissionsCache.Permissions()
}

// GetPermissionByID 获取指定的权限对象
func GetPermissionByID(id string) (*Permission, error) {
	return permissionsCache.GetPermissionByID(id)
}

// GetDefaultPermissionGroups 获取权限组
func GetDefaultPermissionGroups() ([]Group, error) {
	return permissionsCache.PermissionGroups()
}

func WhenChanged(cb func()) {
	permissionsCache.WhenChanged(cb)
}

//缓存
var permissionsCache permissionCacheImpl

//缓存
type permissionCacheData struct {
	permissions    []Permission
	tags           []Tag
	groups         []Group
	tagByID        map[string]*Tag
	permissionByID map[string]*Permission
	saveTime       int64
}

type permissionCacheImpl struct {
	value atomic.Value

	mu          sync.Mutex
	privoders   map[string]PermissionProvider
	isLoading   int32
	changedFunc func()
	logger      log.Logger
}

// func (cache *permissionCacheImpl) tryRead() *permissionCacheData {
// 	o := cache.value.Load()
// 	if o == nil {
// 		return nil
// 	}
//
// 	d, ok := o.(*permissionCacheData)
// 	if !ok {
// 		return nil
// 	}
// 	if (d.saveTime + 60) < time.Now().Unix() {
// 		return nil
// 	}
// 	return d
// }

func (cache *permissionCacheImpl) register(group string, privoder PermissionProvider) {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	if cache.privoders == nil {
		cache.privoders = map[string]PermissionProvider{}
	}

	cache.privoders[group] = privoder
}

func (cache *permissionCacheImpl) WhenChanged(cb func()) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	cache.changedFunc = cb
}

func (cache *permissionCacheImpl) changed() {
	cache.mu.Lock()
	cb := cache.changedFunc
	cache.mu.Unlock()
	if cb != nil {
		cb()
	}
}

func (cache *permissionCacheImpl) load() (*permissionCacheData, error) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	var all PermissionData
	for _, p := range cache.privoders {
		data, err := p.Read()
		if err != nil {
			return nil, err
		}
		appendPermissionData(&all, data)
	}
	return cache.Save(all.Permissions, all.Tags, all.Groups), nil
}

func (cache *permissionCacheImpl) data() (*permissionCacheData, error) {
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
				if _, err := cache.load(); err != nil {
					if cache.logger != nil {
						cache.logger.Warn("load permissions to cache is fail", log.Error(err))
					}
				}
			}()
		}
	}
	return d, nil
}

// Permissions 从缓存中获取权限对象
func (cache *permissionCacheImpl) Permissions() ([]Permission, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.permissions, nil
}

// GetPermissionByID 按 ID 从缓存中获取权限对象
func (cache *permissionCacheImpl) GetPermissionByID(id string) (*Permission, error) {
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

// PermissionTags 从缓存中获取权限对象
func (cache *permissionCacheImpl) PermissionTags() ([]Tag, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.tags, nil
}

// GetPermissionTagByID 按 ID 从缓存中获取权限对象
func (cache *permissionCacheImpl) GetPermissionTagByID(id string) (*Tag, error) {
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

// PermissionGroups 从缓存中获取权限对象
func (cache *permissionCacheImpl) PermissionGroups() ([]Group, error) {
	d, err := cache.data()
	if err != nil {
		return nil, err
	}

	return d.groups, nil
}

//缓存过期
func (cache *permissionCacheImpl) Invalid() {
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
func (cache *permissionCacheImpl) Save(permissions []Permission, tags []Tag, groups []Group) *permissionCacheData {
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

// PermissionData 用于返回缺省权限对象
type PermissionData struct {
	Permissions []Permission `json:"permissions"`
	Groups      []Group      `json:"groups"`
	Tags        []Tag        `json:"tags"`
}

// PermissionProvider 缺省权限对象的提供者
type PermissionProvider interface {
	Read() (*PermissionData, error)
}

// PermissionProviderFunc 缺省权限对象的提供者
type PermissionProviderFunc func() (*PermissionData, error)

func (f PermissionProviderFunc) Read() (*PermissionData, error) {
	if f == nil {
		return nil, nil
	}
	return f()
}

// Register 注册本 App 的权限信息
func Register(env *environment.Environment, serviceID environment.ENV_PROXY_TYPE, mode string, privoder PermissionProvider) Client {
	if mode == "" {
		mode = "apart" // revel.Config.StringDefault("hengwei.perm.mode", "apart")
	}
	logger := log.New(os.Stderr)
	srvOpt := env.GetServiceConfig(serviceID)
	client := Connect(env,
		serviceID,
		Callback(func() (*PermissionData, error) {
			return privoder.Read()
		}),
		mode,
		PermissionEventName,
		urlutil.Join(env.DaemonUrlPath, "/perm/"),
		logger)

	// lifecycleData.OnClosing(client)

	client.WhenChanged(func() {
		permissionsCache.load()
		permissionsCache.changed()
	})

	permissionsCache.logger = logger
	permissionsCache.register(srvOpt.Name, client)
	return client
}
