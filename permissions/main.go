//go:generate gobatis main.go
package permissions

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	cache "github.com/patrickmn/go-cache"
	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/log"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
)

type UserDao interface {
	// @record_type Role
	GetRoleByName(name string) func(*Role) error
	// @record_type User
	GetUserByID(id int64) func(*User) error
	// @record_type User
	GetUserByName(name string) func(*User) error
	// @record_type UserGroup
	GetUsergroupByID(id int64) func(*UserGroup) error
	// @record_type UserGroup
	GetUsergroupByName(name string) func(*UserGroup) error
	// @record_type User
	GetUsers() ([]User, error)
	// @record_type UserGroup
	GetUsergroups() ([]UserGroup, error)

	// @default SELECT * FROM <tablename type="Role" as="roles" /> WHERE
	//  exists (select * from <tablename type="UserAndRole" /> as users_roles
	//     where users_roles.role_id = roles.id and users_roles.user_id = #{userID})
	GetRolesByUser(userID int64) ([]Role, error)

	// @default SELECT * FROM <tablename type="User" as="users" /> WHERE
	//  exists (select * from <tablename type="UserAndUserGroup" /> as u2g
	//     where u2g.user_id = users.id and u2g.group_id = #{groupID})
	GetUserByGroup(groupID int64) ([]User, error)

	// @default SELECT group_id FROM <tablename type="UserAndUserGroup" as="u2g" /> WHERE user_id = #{userID}
	GetGroupIDsByUser(userID int64) ([]int64, error)

	// @record_type PermissionGroupAndRole
	GetPermissionAndRoles(roleIDs []int64) ([]PermissionGroupAndRole, error)

	// @default SELECT value FROM <tablename type="UserProfile" /> WHERE id = #{userID} AND name = #{name}
	ReadProfile(userID int64, name string) (string, error)

	// @type insert
	// @default INSERT INTO <tablename type="UserProfile" /> (id, name, value) VALUES(#{userID}, #{name}, #{value})
	//     ON CONFLICT (id, name) DO UPDATE SET value = excluded.value
	WriteProfile(userID int64, name, value string) error

	// @default DELETE FROM <tablename type="UserProfile" /> WHERE id=#{userID} AND name=#{name}
	DeleteProfile(userID int64, name string) (int64, error)

	GetPermissions() ([]Permissions, error)
	GetPermissionAndGroups() ([]PermissionAndGroup, error)
}

func InitUser(db *sql.DB, driverName string, logger log.Logger) toolbox.UserManager {
	factory, err := gobatis.New(&gobatis.Config{
		Tracer:     log.NewSQLTracer(logger),
		TagPrefix:  "xorm",
		TagMapper:  gobatis.TagSplitForXORM,
		DriverName: driverName,
		DB:         db,
		XMLPaths: []string{
			"gobatis",
		},
	})
	if err != nil {
		panic(err)
	}
	reference := factory.Reference()
	userDao := NewUserDao(&reference)

	um := &userManager{
		logger:               logger,
		userDao:              userDao,
		permissionGroupCache: &GroupCache{},
		userByName:           cache.New(5*time.Minute, 10*time.Minute),
		userByID:             cache.New(5*time.Minute, 10*time.Minute),
		groupByID:            cache.New(5*time.Minute, 10*time.Minute),
		groupByName:          cache.New(5*time.Minute, 10*time.Minute),
	}
	um.refresh()

	um.ensureRoles()

	return um
}

type userManager struct {
	logger               log.Logger
	userDao              UserDao
	permissionGroupCache *GroupCache
	userByName           *cache.Cache
	userByID             *cache.Cache
	groupByName          *cache.Cache
	groupByID            *cache.Cache
	lastErr              concurrency.ErrorValue

	superRole   Role
	adminRole   Role
	visitorRole Role
	guestRole   Role
}

func (um *userManager) refresh() {
	refresh := func() {
		um.lastErr.Set(um.permissionGroupCache.refresh(um.userDao))
	}
	um.permissionGroupCache.Init(5*time.Minute, refresh)
}

func (um *userManager) groupcacheIt(ug toolbox.UserGroup) {
	um.groupByName.SetDefault(ug.Name(), ug)
	um.groupByID.SetDefault(strconv.FormatInt(ug.ID(), 10), ug)
}

func (um *userManager) Groups(opts ...toolbox.UserOption) ([]toolbox.UserGroup, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	if o, found := um.groupByName.Get("____all____"); found && o != nil {
		if ugArray, ok := o.([]toolbox.UserGroup); ok && ugArray != nil {
			return ugArray, nil
		}
	}

	var innerList, err = um.userDao.GetUsergroups()
	if err != nil {
		return nil, errors.Wrap(err, "query all usergroup fail")
	}

	var ugList = make([]toolbox.UserGroup, 0, len(innerList))
	for idx := range innerList {
		ug := &userGroup{um: um, ug: innerList[idx]}
		ugList = append(ugList, ug)
		um.groupcacheIt(ug)
	}

	um.groupByName.SetDefault("____all____", ugList)
	return ugList, nil
}

func (um *userManager) GroupByName(groupname string, opts ...toolbox.UserOption) (toolbox.UserGroup, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	if o, found := um.groupByName.Get(groupname); found && o != nil {
		if ug, ok := o.(toolbox.UserGroup); ok && ug != nil {
			return ug, nil
		}
	}

	var ug = &userGroup{um: um}
	err := um.userDao.GetUsergroupByName(groupname)(&ug.ug)
	if err != nil {
		return nil, errors.Wrap(err, "query usergroup with name is "+groupname+"fail")
	}
	um.groupcacheIt(ug)
	return ug, nil
}

func (um *userManager) GroupByID(groupID int64, opts ...toolbox.UserOption) (toolbox.UserGroup, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	if o, found := um.groupByID.Get(strconv.FormatInt(groupID, 10)); found && o != nil {
		if ug, ok := o.(toolbox.UserGroup); ok && ug != nil {
			return ug, nil
		}
	}

	var ug = &userGroup{um: um}
	err := um.userDao.GetUsergroupByID(groupID)(&ug.ug)
	if err != nil {
		return nil, errors.Wrap(err, "query usergroup with id is "+fmt.Sprint(groupID)+"fail")
	}
	um.groupcacheIt(ug)
	return ug, nil
}

func (um *userManager) Users(opts ...toolbox.UserOption) ([]toolbox.User, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	var includeDisabled bool
	for _, opt := range opts {
		switch opt.(type) {
		case toolbox.UserIncludeDisabled:
			includeDisabled = true
		}
	}

	if includeDisabled {
		if o, found := um.userByName.Get("____all____"); found && o != nil {
			if ugArray, ok := o.([]toolbox.User); ok && ugArray != nil {
				return ugArray, nil
			}
		}
	} else {
		if o, found := um.userByName.Get("____all_enabled____"); found && o != nil {
			if ugArray, ok := o.([]toolbox.User); ok && ugArray != nil {
				return ugArray, nil
			}
		}
	}

	innerList, err := um.userDao.GetUsers()
	if err != nil {
		return nil, errors.Wrap(err, "query all usergroup fail")
	}

	um.ensureRoles()

	var uList = make([]toolbox.User, 0, len(innerList))
	var enabledList = make([]toolbox.User, 0, len(innerList))

	for idx := range innerList {
		u := &user{um: um, u: innerList[idx]}
		if err := um.load(u); err != nil {
			return nil, err
		}
		uList = append(uList, u)
		if !u.IsDisabled() {
			enabledList = append(enabledList, u)
		}
		um.usercacheIt(u)
	}

	um.userByName.SetDefault("____all____", uList)
	um.userByName.SetDefault("____all_enabled____", enabledList)

	if includeDisabled {
		return uList, nil
	}
	return enabledList, nil
}

func (um *userManager) usercacheIt(u toolbox.User) {
	um.userByName.SetDefault(u.Name(), u)
	um.userByID.SetDefault(strconv.FormatInt(u.ID(), 10), u)
}

func (um *userManager) ensureRoles() {
	for _, data := range []struct {
		role *Role
		name string
	}{
		{role: &um.superRole, name: toolbox.RoleSuper},
		{role: &um.adminRole, name: toolbox.RoleAdministrator},
		{role: &um.visitorRole, name: toolbox.RoleVisitor},
		{role: &um.guestRole, name: toolbox.RoleGuest},
	} {
		if data.role.ID != 0 {
			continue
		}

		err := um.userDao.GetRoleByName(data.name)(data.role)
		if err != nil {
			data.role.Name = data.name
			um.logger.Warn("role isnot found", log.String("role", data.name), log.Error(err))
		} else {
			um.userByID.Flush()
			um.userByName.Flush()
		}
	}
}

func (um *userManager) ByName(userName string, opts ...toolbox.UserOption) (toolbox.User, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	var includeDisabled bool
	for _, opt := range opts {
		switch opt.(type) {
		case toolbox.UserIncludeDisabled:
			includeDisabled = true
		}
	}

	if o, found := um.userByName.Get(userName); found && o != nil {
		if u, ok := o.(toolbox.User); ok && u != nil {
			if includeDisabled {
				return u, nil
			}

			if u.(*user).IsDisabled() {
				return nil, errors.New("user with name is '" + userName + "' is disabled")
			}
			return u, nil
		}
	}

	um.ensureRoles()

	var u = &user{um: um}
	err := um.userDao.GetUserByName(userName)(&u.u)
	if err != nil {
		switch userName {
		case toolbox.UserAdmin:
			u.u.Name = userName
			u.roleNames = []string{toolbox.RoleAdministrator}
			u.roles = []Role{um.adminRole}

			um.usercacheIt(u)
			return u, nil
		case toolbox.UserGuest:
			u.u.Name = userName
			u.roleNames = []string{toolbox.RoleGuest}
			u.roles = []Role{um.guestRole}
			um.usercacheIt(u)
			return u, nil
		default:
			return nil, errors.Wrap(err, "query user with name is '"+userName+"' fail")
		}
	}

	if !includeDisabled {
		if u.IsDisabled() {
			return nil, errors.New("user with name is '" + userName + "' is disabled")
		}
	}

	err = um.load(u)
	if err != nil {
		return nil, err
	}
	um.usercacheIt(u)
	return u, nil
}

func (um *userManager) ByID(userID int64, opts ...toolbox.UserOption) (toolbox.User, error) {
	if e := um.lastErr.Get(); e != nil {
		return nil, e
	}

	var includeDisabled bool
	for _, opt := range opts {
		switch opt.(type) {
		case toolbox.UserIncludeDisabled:
			includeDisabled = true
		}
	}

	if o, found := um.userByID.Get(strconv.FormatInt(userID, 10)); found && o != nil {
		if u, ok := o.(toolbox.User); ok && u != nil {
			if includeDisabled {
				return u, nil
			}

			if u.(*user).IsDisabled() {
				return nil, errors.New("user with name is " + u.Name() + " is disabled")
			}
			return u, nil
		}
	}

	um.ensureRoles()

	var u = &user{um: um}
	err := um.userDao.GetUserByID(userID)(&u.u)
	if err != nil {
		return nil, errors.Wrap(err, "query user with id is "+fmt.Sprint(userID)+"fail")
	}

	if !includeDisabled {
		if u.IsDisabled() {
			return nil, errors.New("user with name is " + u.Name() + " is disabled")
		}
	}

	err = um.load(u)
	if err != nil {
		return nil, err
	}
	um.usercacheIt(u)
	return u, nil
}

func (um *userManager) load(u *user) error {
	var err error
	u.roles, err = um.userDao.GetRolesByUser(u.ID())
	if err != nil {
		return errors.Wrap(err, "query permissions and roles with user is "+u.Name()+" fail")
	}

	u.roleNames = nil
	u.Roles() // 缓存 roleNames

	if um.superRole.ID != 0 {
		for _, role := range u.roles {
			if role.ID == um.superRole.ID {
				return nil
			}
		}
	}

	if um.adminRole.ID != 0 {
		for _, role := range u.roles {
			if role.ID == um.adminRole.ID {
				return nil
			}
		}

		if u.u.Name == toolbox.UserAdmin {
			u.roles = append(u.roles, um.adminRole)

			u.roleNames = nil
			u.Roles() // 缓存 roleNames

			return nil
		}
	}

	var roleIDs = make([]int64, len(u.roles))
	for idx := range u.roles {
		roleIDs[idx] = u.roles[idx].ID
	}

	if len(roleIDs) > 0 {
		u.permissionsAndRoles, err = um.userDao.GetPermissionAndRoles(roleIDs)
		//err = um.db.PermissionGroupsAndRoles().Where(orm.Cond{"role_id IN": roleIDs}).All(&u.permissionsAndRoles)
		if err != nil {
			return errors.Wrap(err, "query permissions and roles with user is "+u.Name()+" fail")
		}
	}

	u.groups, err = um.userDao.GetGroupIDsByUser(u.ID())
	if err != nil {
		return errors.Wrap(err, "query user and usergroup with user is "+u.Name()+" fail")
	}
	return nil
}

type user struct {
	um                  *userManager
	permissionsAndRoles []PermissionGroupAndRole
	u                   User
	roles               []Role
	roleNames           []string
	groups              []int64
	profiles            map[string]string
}

func (u *user) IsDisabled() bool {
	return u.u.IsDisabled()
}

func (u *user) ID() int64 {
	return u.u.ID
}

func (u *user) Name() string {
	return u.u.Name
}

func (u *user) Nickname() string {
	return u.u.Nickname
}

func (u *user) HasAdminRole() bool {
	return u.hasRoleID(u.um.adminRole.ID)
}

func (u *user) HasGuestRole() bool {
	return u.hasRoleID(u.um.guestRole.ID)
}

func (u *user) hasRoleID(id int64) bool {
	for idx := range u.roles {
		if u.roles[idx].ID == id {
			return true
		}
	}
	return false
}

func (u *user) HasRole(role string) bool {
	for _, name := range u.roleNames {
		if name == role {
			return true
		}
	}
	return false
}

func (u *user) IsMemberOf(group int64) bool {
	for _, id := range u.groups {
		if id == group {
			return true
		}
	}
	return false
}

func (u *user) WriteProfile(key, value string) error {
	if value == "" {
		_, err := u.um.userDao.DeleteProfile(u.ID(), key)
		if err != nil {
			return errors.Wrap(err, "DeleteProfile")
		}
		if u.profiles != nil {
			delete(u.profiles, key)
		}
		return nil
	}

	err := u.um.userDao.WriteProfile(u.ID(), key, value)
	if err != nil {
		return errors.Wrap(err, "WriteProfile")
	}

	if u.profiles != nil {
		u.profiles[key] = value
	}
	return nil
}

func (u *user) ReadProfile(key string) (string, error) {
	if u.profiles != nil {
		value, ok := u.profiles[key]
		if ok {
			return value, nil
		}
	}
	value, err := u.um.userDao.ReadProfile(u.ID(), key)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", errors.Wrap(err, "ReadProfile")
	}
	if u.profiles != nil {
		u.profiles[key] = value
	} else {
		u.profiles = map[string]string{key: value}
	}
	return value, nil
}

func (u *user) Roles() []string {
	if len(u.roleNames) != 0 {
		return u.roleNames
	}
	if len(u.roles) == 0 {
		return nil
	}

	roleNames := make([]string, 0, len(u.roles))
	for idx := range u.roles {
		roleNames = append(roleNames, u.roles[idx].Name)
	}

	u.roleNames = roleNames
	return u.roleNames
}

func (u *user) Data(key string) interface{} {
	switch key {
	case "id":
		return u.u.ID
	case "name":
		return u.u.Name
	case "nickname":
		return u.u.Nickname
	case "description":
		return u.u.Description
	case "attributes":
		return u.u.Attributes
	case "source":
		return u.u.Source
	case "created_at":
		return u.u.CreatedAt
	case "updated_at":
		return u.u.UpdatedAt
	default:
		if u.u.Attributes != nil {
			return u.u.Attributes[key]
		}
	}
	return nil
}

func (u *user) HasPermission(permissionID, op string) bool {
	if u.Name() == toolbox.UserAdmin {
		return true
	}

	if u.um.superRole.ID != 0 {
		for _, role := range u.roles {
			if role.ID == u.um.superRole.ID {
				return true
			}
		}
	}

	if u.um.adminRole.ID != 0 {
		for _, role := range u.roles {
			if role.ID == u.um.adminRole.ID {
				return true
			}
		}
	}
	if u.um.visitorRole.ID != 0 && toolbox.QUERY == op {
		for _, role := range u.roles {
			if role.ID == u.um.visitorRole.ID {
				return true
			}
		}
	}

	for _, pr := range u.permissionsAndRoles {
		enableOperation := false
		switch op {
		case toolbox.CREATE:
			enableOperation = pr.CreateOperation
		case toolbox.DELETE:
			enableOperation = pr.DeleteOperation
		case toolbox.UPDATE:
			enableOperation = pr.UpdateOperation
		case toolbox.QUERY:
			enableOperation = pr.QueryOperation
			if pr.CreateOperation ||
				pr.DeleteOperation ||
				pr.UpdateOperation {
				enableOperation = true
			}
		default:
			panic(errors.New("Operation '" + op + "' is unknown"))
		}
		if !enableOperation {
			continue
		}

		if u.hasPermission(pr.GroupID, permissionID) {
			return true
		}
	}
	return false
}

func (u *user) hasPermission(groupID int64, permissionID string) bool {
	group := u.um.permissionGroupCache.Get(groupID)
	if group == nil {
		u.um.logger.Warn("[permissions] permission group isn't found", log.Int64("groupID", groupID))
		return false
	}
	return u.hasPermissionInGroup(group, permissionID)
}

func (u *user) hasPermissionInGroup(group *Permissions, permissionID string) bool {
	// 在本组中查找是不是有这个权限
	for _, id := range group.PermissionIDs {
		if permissionID == id {
			return true
		}
	}

	// 在本组中查找是不是有标签含有这个权限
	for _, tag := range group.PermissionTags {
		permissionList, err := GetPermissionsByTag(tag)
		if err != nil {
			panic(err)
		}

		for _, permission := range permissionList {
			if permissionID == permission.ID {
				return true
			}
		}
	}

	// 在子组中查找这个权限
	children := u.um.permissionGroupCache.GetChildren(group.ID)
	if len(children) != 0 {
		for _, child := range children {
			if u.hasPermissionInGroup(child, permissionID) {
				return true
			}
		}
	}
	return false
}

type userGroup struct {
	um       *userManager
	ug       UserGroup
	children []toolbox.User
}

func (ug *userGroup) ID() int64 {
	return ug.ug.ID
}

func (ug *userGroup) Name() string {
	return ug.ug.Name
}

func (ug *userGroup) Users(opts ...toolbox.UserOption) ([]toolbox.User, error) {
	var includeDisabled bool
	for _, opt := range opts {
		switch opt.(type) {
		case toolbox.UserIncludeDisabled:
			includeDisabled = true
		}
	}

	if ug.children == nil {

		innerList, err := ug.um.userDao.GetUserByGroup(ug.ID())
		if err != nil {
			return nil, errors.Wrap(err, "query all usergroup fail")
		}

		ug.um.ensureRoles()

		var uList = make([]toolbox.User, 0, len(innerList))

		for idx := range innerList {
			u := &user{um: ug.um, u: innerList[idx]}
			if err := ug.um.load(u); err != nil {
				return nil, err
			}
			uList = append(uList, u)
			ug.um.usercacheIt(u)
		}

		ug.children = uList
	}

	if includeDisabled {
		return ug.children, nil
	}

	var enabledList = make([]toolbox.User, 0, len(ug.children))
	for idx := range ug.children {
		if !ug.children[idx].(*user).IsDisabled() {
			enabledList = append(enabledList, ug.children[idx])
		}
	}

	return enabledList, nil
}
