package ds

import (
	"database/sql"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-xorm/xorm"
	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/ds/models"
	merrors "github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/types"
)

type RecordVersion struct {
	ID        int64     `json:"id,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// ManagedObject 代表一个管理对象, 注意它是一个不可变对象，任何人不要试图修改它
type ManagedObject struct {
	cache *MoCache
	models.Object
	Type  *types.ClassDefinition
	Value interface{}
}

// GetName 返回对象的名称
func (mo *ManagedObject) GetName() string {
	if mo.Value == nil {
		return mo.Name
	}
	switch o := mo.Value.(type) {
	case NetworkDevice:
		return o.DisplayName
	}
	return mo.Name
}

// GetDBObject 返回对象的数据库模型
func (mo *ManagedObject) GetDBObject() interface{} {
	if mo.Value == nil {
		return mo.Object
	}
	switch o := mo.Value.(type) {
	case *NetworkDevice:
		return o.NetworkDevice
	case *NetworkLink:
		return o.NetworkLink
	}
	return mo.Value
}

// RecordVersion 记录的版本号
func (mo *ManagedObject) RecordVersion() RecordVersion {
	return RecordVersion{ID: mo.ID, UpdatedAt: mo.UpdatedAt}
}

// NetworkDevice 代表一个网络管理对象，任何人不要试图修改它
type NetworkDevice struct {
	mo   *ManagedObject
	Type *types.ClassDefinition
	models.NetworkDevice

	snmpParams atomic.Value
}

// accessParams 返回访问参数
func (dev *NetworkDevice) accessParams(t string) (*models.AccessParams, error) {
	var sp []*models.AccessParams
	collection := orm.New(func() interface{} {
		return &models.AccessParams{}
	})(dev.mo.cache.Engine)

	err := collection.Where(orm.Cond{"managed_object_id": dev.mo.ID, "type": t}).All(&sp)
	if err != nil {
		return nil, err
	}
	if len(sp) == 0 {
		return nil, errors.New(t + " is empty.")
	}
	if len(sp) != 1 {
		return nil, errors.New(t + " is muli choices.")
	}
	return sp[0], nil
}

// SnmpParams 返回 SNMP 访问参数
func (dev *NetworkDevice) SnmpParams() (*models.SnmpParams, error) {
	o := dev.snmpParams.Load()
	if nil != o {
		if nd, ok := o.(*models.SnmpParams); ok {
			return nd, nil
		}
	}
	var sp []*models.SnmpParams
	collection := orm.New(func() interface{} {
		return &models.SnmpParams{}
	})(dev.mo.cache.Engine)

	err := collection.Where(orm.Cond{"managed_object_id": dev.mo.ID}).All(&sp)
	if err != nil {
		if strings.Contains(err.Error(), `does not exist`) ||
			strings.Contains(err.Error(), `不存在`) {
			ap, err := dev.accessParams("snmp_param")
			if err != nil {
				return nil, err
			}
			return ap.ToSnmpParams()
		}
		return nil, err
	}
	if len(sp) == 0 {
		return nil, errors.New("SnmpParams is empty.")
	}
	dev.snmpParams.Store(sp[0])
	return sp[0], nil
}

// NetworkLink 代表一个线路，任何人不要试图修改它
type NetworkLink struct {
	mo   *ManagedObject
	Type *types.ClassDefinition
	models.NetworkLink

	from, to atomic.Value
}

// From 返回线路入端的设备
func (link *NetworkLink) From() (*NetworkDevice, error) {
	o := link.from.Load()
	if nil != o {
		if nd, ok := o.(*NetworkDevice); ok {
			return nd, nil
		}
	}

	nd, err := link.mo.cache.GetNetworkDevice(link.FromDevice)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, merrors.NotFound(link.FromDevice)
		}
		return nil, err
	}

	link.from.Store(nd)
	return nd, nil
}

// To 返回线路另一端的设备
func (link *NetworkLink) To() (*NetworkDevice, error) {
	o := link.to.Load()
	if nil != o {
		if nd, ok := o.(*NetworkDevice); ok {
			return nd, nil
		}
	}

	nd, err := link.mo.cache.GetNetworkDevice(link.ToDevice)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, merrors.NotFound(link.ToDevice)
		}
		return nil, err
	}

	link.to.Store(nd)
	return nd, nil
}

// MoCache 管理对角的缓存
type MoCache struct {
	Definitions    *types.TableDefinitions
	managedElement *types.ClassDefinition
	managedObject  *types.ClassDefinition
	managedLink    *types.ClassDefinition
	networkDevice  *types.ClassDefinition
	lock           sync.RWMutex
	values         map[int64]*ManagedObject
	all_devices    []*NetworkDevice
	all_links      []*NetworkLink
	Engine         *xorm.Engine
}

func (cache *MoCache) Init(engine *xorm.Engine, definitions *types.TableDefinitions) error {
	if definitions == nil {
		return errors.New("definitions is nil")
	}
	cache.Engine = engine
	cache.Definitions = definitions

	cache.managedElement = definitions.FindByUnderscoreName("managed_element")
	if nil == cache.managedElement {
		return errors.New("class 'ManagedElement' is not found")
	}

	cache.managedObject = definitions.FindByUnderscoreName("managed_object")
	if nil == cache.managedObject {
		return errors.New("type 'ManagedObject' isn't found")
	}

	cache.networkDevice = definitions.FindByUnderscoreName("network_device")
	if nil == cache.networkDevice {
		return errors.New("type 'NetworkDevice' isn't found")
	}
	cache.managedLink = definitions.FindByUnderscoreName("network_link")
	if nil == cache.managedLink {
		return errors.New("type 'NetworkLink' isn't found")
	}
	return nil
}

func (db *MoCache) Objects() *orm.Collection {
	return orm.New(func() interface{} {
		return &models.Object{}
	})(db.Engine)
}

func (db *MoCache) NetworkDevices() *orm.Collection {
	return orm.New(func() interface{} {
		return &models.NetworkDevice{}
	})(db.Engine)
}

func (db *MoCache) NetworkLinks() *orm.Collection {
	return orm.New(func() interface{} {
		return &models.NetworkLink{}
	})(db.Engine)
}

// Refresh 刷新绶存，确保内存与数据库中的数据一致。
func (cache *MoCache) Refresh() error {
	cache.lock.RLock()
	isEmpty := len(cache.values) == 0
	cache.lock.RUnlock()
	if isEmpty {
		log.Println("[mo_cache] cache is empty, skip refresh.")
		return nil
	}

	rows, err := cache.Engine.DB().Query("select id, updated_at from " + models.Objects.TableName())
	if err != nil {
		if err == sql.ErrNoRows {
			cache.lock.Lock()
			cache.values = nil
			cache.all_devices = nil
			cache.all_links = nil
			cache.lock.Unlock()

			log.Println("[mo_cache] database is empty, clear cache.")
			return nil
		}
		return errors.New("GetSnapshots:" + err.Error())
	}
	defer rows.Close()

	moCopies := map[int64]*ManagedObject{}
	cache.lock.RLock()
	for k, v := range cache.values {
		moCopies[k] = v
	}
	cache.lock.RUnlock()

	//var created []int64
	var updated []int64
	//var deleted []int64

	for rows.Next() {
		var version RecordVersion
		if err := rows.Scan(&version.ID, &version.UpdatedAt); err != nil {
			return errors.New("ReadSnapshots:" + err.Error())
		}

		if o, ok := moCopies[version.ID]; ok {
			if !o.UpdatedAt.Equal(version.UpdatedAt) {
				updated = append(updated, version.ID)
			}
			delete(moCopies, version.ID)
		} // else {
		//	created = append(created, version.ID)
		// }
	}
	cache.lock.Lock()
	for id := range moCopies {
		delete(cache.values, id)
	}
	for _, id := range updated {
		delete(cache.values, id)
	}

	cache.all_devices = nil
	cache.all_links = nil
	cache.lock.Unlock()

	log.Println("[mo_cache] update", len(updated), ", delete", len(moCopies))
	return nil
}

func findSpec(definitions *types.TableDefinitions, name string, defSpec *types.ClassDefinition) *types.ClassDefinition {
	typeSpec := definitions.FindByUnderscoreName(name)
	if typeSpec != nil {
		return typeSpec
	}

	typeSpec = definitions.Find(name)
	if typeSpec != nil {
		return typeSpec
	}

	return defSpec
}

func (cache *MoCache) toNetworkDevice(obj *models.Object) (*ManagedObject, error) {
	var nd models.NetworkDevice
	err := cache.NetworkDevices().Id(obj.ID).Get(&nd)
	if err != nil {
		return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(obj.ID, 10) + ":" + obj.Name + ") fail, " + err.Error())
	}

	typeSpec := findSpec(cache.Definitions, nd.Type, cache.networkDevice)
	if typeSpec == nil {
		return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(nd.ID, 10) + ":" + nd.DisplayName + ") fail, type(" + nd.Type + ") is unknown.")
	}

	mo := &ManagedObject{
		cache:  cache,
		Object: *obj,
		Type:   typeSpec,
	}
	mo.Value = &NetworkDevice{mo: mo,
		NetworkDevice: nd,
		Type:          typeSpec}
	return mo, nil
}

func (cache *MoCache) toNetworkDeviceFrom(nd *models.NetworkDevice) (*ManagedObject, error) {
	typeSpec := findSpec(cache.Definitions, nd.Type, cache.networkDevice)
	if typeSpec == nil {
		return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(nd.ID, 10) + ":" + nd.DisplayName + ") fail, type(" + nd.Type + ") is unknown.")
	}

	mo := &ManagedObject{
		cache: cache,
		Object: models.Object{
			Table:     models.NetworkDevices.TableName(),
			ID:        nd.ID,
			Type:      nd.Type,
			Name:      nd.Name,
			UpdatedAt: nd.UpdatedAt,
			CreatedAt: nd.CreatedAt,
		},
		Type: typeSpec,
	}
	mo.Value = &NetworkDevice{mo: mo,
		NetworkDevice: *nd,
		Type:          typeSpec}
	return mo, nil
}

func (cache *MoCache) toNetworkLink(obj *models.Object) (*ManagedObject, error) {
	var nl models.NetworkLink
	err := cache.NetworkLinks().Id(obj.ID).Get(&nl)
	if err != nil {
		return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(obj.ID, 10) + ":" + obj.Name + ") fail, " + err.Error())
	}

	// typeSpec := cache.lifecycle.Definitions.FindByUnderscoreName(nd.Type)
	// if typeSpec == nil {
	// 	typeSpec = cache.lifecycle.Definitions.Find(nd.Type)
	// 	if typeSpec == nil {
	// 		return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(nd.ID, 10) + ":" + nd.DisplayName + ") fail, type is unknown.")
	// 	}
	// }
	typeSpec := cache.managedLink
	mo := &ManagedObject{
		cache:  cache,
		Object: *obj,
		Type:   typeSpec,
	}
	mo.Value = &NetworkLink{mo: mo,
		NetworkLink: nl,
		Type:        typeSpec}
	return mo, nil
}

func (cache *MoCache) toManagedObject(obj *models.Object) (*ManagedObject, error) {
	switch obj.Table {
	case models.NetworkDevices.TableName():
		return cache.toNetworkDevice(obj)
	case models.NetworkLinks.TableName():
		return cache.toNetworkLink(obj)
	default:
		typeSpec := cache.Definitions.FindByTableName(obj.Table)
		if typeSpec == nil {
			return nil, errors.New("toManagedObject: load mo(" + strconv.FormatInt(obj.ID, 10) + ":" + obj.Name + ") fail, type is unknown.")
		}
		return &ManagedObject{
			cache:  cache,
			Object: *obj,
			Type:   typeSpec,
		}, nil
	}
}

// Get 获取一个指定 ID 的管理对象，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) Get(moID int64) (*ManagedObject, error) {
	cache.lock.RLock()
	if cache.values != nil {
		if old, ok := cache.values[moID]; ok {
			cache.lock.RUnlock()
			return old, nil
		}
	}
	cache.lock.RUnlock()

	var obj models.Object
	err := cache.Objects().Id(moID).Get(&obj)
	if err != nil {
		if err == orm.ErrNotFound {
			return nil, merrors.NotFound(moID)
		}
		return nil, err
	}
	mo, err := cache.toManagedObject(&obj)
	if err != nil {
		return nil, err
	}
	cache.lock.Lock()
	if cache.values == nil {
		cache.values = map[int64]*ManagedObject{moID: mo}
	} else {
		cache.values[moID] = mo
	}
	cache.lock.Unlock()

	return mo, nil
}

func (cache *MoCache) get(obj *models.Object) (*ManagedObject, error) {
	cache.lock.RLock()
	if cache.values != nil {
		if old, ok := cache.values[obj.ID]; ok {
			cache.lock.RUnlock()
			return old, nil
		}
	}
	cache.lock.RUnlock()

	mo, err := cache.toManagedObject(obj)
	if err != nil {
		return nil, err
	}
	cache.lock.Lock()
	if cache.values == nil {
		cache.values = map[int64]*ManagedObject{obj.ID: mo}
	} else {
		cache.values[obj.ID] = mo
	}
	cache.lock.Unlock()

	return mo, nil
}

// GetNetworkDevice 获取一个指定 ID 的网络管理对象，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) GetNetworkDevice(moID int64) (*NetworkDevice, error) {
	mo, err := cache.Get(moID)
	if err != nil {
		return nil, err
	}
	if nil == mo {
		return nil, merrors.NotFound(moID)
	}
	nd, ok := mo.Value.(*NetworkDevice)
	if !ok {
		return nil, errors.New("GetNetworkDevice: load mo(" + strconv.FormatInt(mo.ID, 10) + ":" + mo.Name + ") fail, type(" + mo.Type.UName() + ") isn't network device.")
	}
	return nd, nil
}

func (cache *MoCache) searchNetworkDeviceByAddress(domain, address string) ([]models.NetworkDevice, error) {
	query := cache.NetworkDevices().Where(orm.Cond{"address": address})
	query = query.Or(orm.Cond{
		"exists (SELECT 1 FROM tpt_network_addresses WHERE managed_object_id = " +
			models.NetworkDevices.TableName() +
			".id AND address = ?)": address,
	})
	if "" != domain {
		query = query.And(orm.Cond{"exists (SELECT 1 FROM tpt_domains WHERE id = " +
			models.NetworkDevices.TableName() +
			".domain_id AND name = ?)": domain})
	}

	var devices []models.NetworkDevice
	err := query.All(&devices)
	if err != nil {
		return nil, err
	}
	return devices, nil
}

// GetNetworkDeviceByAddress 获取一个指定 ID 的网络管理对象，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) GetNetworkDeviceByAddress(domain, address string) (*NetworkDevice, error) {
	devices, err := cache.searchNetworkDeviceByAddress(domain, address)
	if err != nil {
		return nil, err
	}
	if len(devices) == 0 {
		return nil, nil
	}
	if len(devices) != 1 {
		return nil, errors.New("GetNetworkDeviceByAddress: muti choice is find.")
	}
	return cache.loadOrCreateNetworkDevice(&devices[0])
}

// SearchNetworkDeviceByName 获取一个指定 ID 的网络管理对象，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) SearchNetworkDeviceByName(name string) ([]*NetworkDevice, error) {
	var devices []models.NetworkDevice
	err := cache.NetworkDevices().Where(orm.Cond{"full_name like": name}).
		Or(orm.Cond{"name like": name}).
		Or(orm.Cond{"zh_name like": name}).
		All(&devices)
	if err != nil {
		return nil, err
	}
	// if len(devices) == 0 {
	// 	filter = models.NetworkDeviceModel.C.ZHNAME.LIKE(name)
	// 	builder = models.NetworkDeviceModel.Where(filter).Select()
	// 	devices, err := models.NetworkDeviceModel.QueryWith(cache.lifecycle.DbRunner, builder)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	if len(devices) == 0 {
	// 		return nil, nil
	// 	}
	// }
	return cache.convertNetworkDevices(devices)
}

func (cache *MoCache) convertNetworkDevices(devices []models.NetworkDevice) ([]*NetworkDevice, error) {
	results := make([]*NetworkDevice, 0, len(devices))
	for _, dev := range devices {
		nd, err := cache.loadOrCreateNetworkDevice(&dev)
		if err != nil {
			return nil, err
		}
		results = append(results, nd)
	}
	return results, nil
}

// SearchNetworkDevices 按指定的字符在设备的名称或地址字段中查找，注意名称是模糊查找
func (cache *MoCache) SearchNetworkDevices(name string) ([]*NetworkDevice, error) {
	if ip := net.ParseIP(name); nil != ip {
		devices, err := cache.searchNetworkDeviceByAddress("", name)
		if err != nil {
			return nil, err
		}
		if len(devices) == 0 {
			return nil, nil
		}
		return cache.convertNetworkDevices(devices)
	}
	return cache.SearchNetworkDeviceByName(name)
}

func (cache *MoCache) loadOrCreateNetworkDevice(nd *models.NetworkDevice) (*NetworkDevice, error) {
	cache.lock.RLock()
	if cache.values != nil {
		if old, ok := cache.values[nd.ID]; ok {
			cache.lock.RUnlock()

			nd, ok := old.Value.(*NetworkDevice)
			if !ok {
				return nil, errors.New("loadOrCreateNetworkDevice: load mo(" + strconv.FormatInt(nd.ID, 10) +
					":" + nd.Name + ") fail, type(" + nd.Type.UName() + ") isn't network device.")
			}
			return nd, nil
		}
	}
	cache.lock.RUnlock()

	mo, err := cache.toNetworkDeviceFrom(nd)
	if err != nil {
		return nil, err
	}
	cache.lock.Lock()
	if cache.values == nil {
		cache.values = map[int64]*ManagedObject{mo.ID: mo}
	} else {
		cache.values[mo.ID] = mo
	}
	cache.lock.Unlock()

	return mo.Value.(*NetworkDevice), nil
}

// ListNetworkDevices 列出所有的网络管理对象，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) ListNetworkDevices() ([]*NetworkDevice, error) {
	cache.lock.RLock()
	all := cache.all_devices
	cache.lock.RUnlock()
	if nil != all {
		return all, nil
	}

	var moList []models.Object
	err := cache.Objects().Where(orm.Cond{"table_name": models.NetworkDevices}).All(&moList)
	if nil != err {
		return nil, err
	}

	var devices = make([]*NetworkDevice, 0, len(moList))
	for _, mo := range moList {
		mo, err := cache.get(&mo)
		if err != nil {
			return nil, err
		}

		nd, ok := mo.Value.(*NetworkDevice)
		if !ok {
			return nil, errors.New("ListNetworkDevices: load mo(" + strconv.FormatInt(mo.ID, 10) + ":" + mo.Name + ") fail, type(" + mo.Type.UName() + ") isn't network device.")
		}
		devices = append(devices, nd)
	}

	cache.lock.Lock()
	cache.all_devices = devices
	cache.lock.Unlock()
	return devices, nil
}

// GetNetworkLink 获取一个指定 ID 的网络线路，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) GetNetworkLink(moID int64) (*NetworkLink, error) {
	mo, err := cache.Get(moID)
	if err != nil {
		return nil, err
	}
	if mo == nil {
		return nil, merrors.NotFound(moID)
	}

	nd, ok := mo.Value.(*NetworkLink)
	if !ok {
		return nil, errors.New("GetNetworkLink: load mo(" + strconv.FormatInt(mo.ID, 10) + ":" + mo.Name + ") fail, type(" + mo.Type.UName() + ") isn't network link.")
	}
	return nd, nil
}

// ListNetworkLinks 列出所有的网络线路，如果管理对象没有被加功到内存那么立即加载, 注意它是一个不可变对象，任何人不要试图修改它
func (cache *MoCache) ListNetworkLinks() ([]*NetworkLink, error) {
	cache.lock.RLock()
	all := cache.all_links
	cache.lock.RUnlock()
	if nil != all {
		return all, nil
	}

	var moList []models.Object
	err := cache.Objects().Where(orm.Cond{"table_name": models.NetworkLinks.TableName()}).All(&moList)
	if nil != err {
		return nil, err
	}

	var links = make([]*NetworkLink, 0, len(moList))
	for _, mo := range moList {
		mo, err := cache.get(&mo)
		if err != nil {
			return nil, err
		}

		nd, ok := mo.Value.(*NetworkLink)
		if !ok {
			return nil, errors.New("ListNetworkLinks: load mo(" + strconv.FormatInt(mo.ID, 10) + ":" + mo.Name + ") fail, type(" + mo.Type.UName() + ") isn't network link.")
		}
		links = append(links, nd)
	}

	cache.lock.Lock()
	cache.all_links = links
	cache.lock.Unlock()
	return links, nil
}
