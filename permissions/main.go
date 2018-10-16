package permissions

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-xorm/xorm"
	cache "github.com/patrickmn/go-cache"
	"github.com/runner-mei/orm"
	"github.com/three-plus-three/modules/concurrency"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
)

func InitUser(engine *xorm.Engine) toolbox.UserManager {
	um := &userManager{
		db:                   &DB{DB: orm.DB{Engine: engine}},
		permissionGroupCache: &GroupCache{},

		userByName: cache.New(5*time.Minute, 10*time.Minute),
		userByID:   cache.New(5*time.Minute, 10*time.Minute),
		groupByID:  cache.New(5*time.Minute, 10*time.Minute),
	}
	um.refresh()

	if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleSuper}).One(&um.superRole); e != nil {
		um.superRole.Name = toolbox.RoleSuper
		log.Println("[warn] role", toolbox.RoleSuper, "isnot found -", e)
	}

	if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleAdministrator}).One(&um.adminRole); e != nil {
		um.adminRole.Name = toolbox.RoleAdministrator
		log.Println("[warn] role", toolbox.RoleAdministrator, "isnot found -", e)
	}

	if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleVisitor}).One(&um.visitorRole); e != nil {
		um.visitorRole.Name = toolbox.RoleVisitor
		log.Println("[warn] role", toolbox.RoleVisitor, "isnot found -", e)
	}

	if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleGuest}).One(&um.guestRole); e != nil {
		um.guestRole.Name = toolbox.RoleGuest
		log.Println("[warn] role", toolbox.RoleGuest, "isnot found -", e)
	}

	return um
}

type userGroup struct {
	um *userManager
	ug UserGroup
}

func (ug *userGroup) ID() int64 {
	return ug.ug.ID
}

func (ug *userGroup) Name() string {
	return ug.ug.Name
}

type userManager struct {
	db                   *DB
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
		um.lastErr.Set(um.permissionGroupCache.refresh(um.db))
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

	var innerList []UserGroup
	err := um.db.UserGroups().Where().All(&innerList)
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
	err := um.db.UserGroups().Where(orm.Cond{"name": groupname}).One(&ug.ug)
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
	err := um.db.UserGroups().ID(groupID).Get(&ug.ug)
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

	var innerList []User
	err := um.db.Users().Where().All(&innerList)
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
	if um.superRole.ID == 0 {
		if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleSuper}).One(&um.superRole); e != nil {
			log.Println("[warn] role", toolbox.RoleSuper, "isnot found -", e)
		} else {
			um.userByID.Flush()
			um.userByName.Flush()
		}
	}
	if um.adminRole.ID == 0 {
		if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleAdministrator}).One(&um.adminRole); e != nil {
			log.Println("[warn] role", toolbox.RoleAdministrator, "isnot found -", e)
		} else {
			um.userByID.Flush()
			um.userByName.Flush()
		}
	}
	if um.visitorRole.ID == 0 {
		if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleVisitor}).One(&um.visitorRole); e != nil {
			log.Println("[warn] role", toolbox.RoleVisitor, "isnot found -", e)
		} else {
			um.userByID.Flush()
			um.userByName.Flush()
		}
	}
	if um.guestRole.ID == 0 {
		if e := um.db.Roles().Where(orm.Cond{"name": toolbox.RoleGuest}).One(&um.guestRole); e != nil {
			log.Println("[warn] role", toolbox.RoleGuest, "isnot found -", e)
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
				return nil, errors.New("user with name is " + userName + " is disabled")
			}
			return u, nil
		}
	}

	um.ensureRoles()

	var u = &user{um: um}
	err := um.db.Users().Where(orm.Cond{"name": userName}).Omit("profiles").One(&u.u)
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
			return nil, errors.Wrap(err, "query user with name is "+userName+"fail")
		}
	}

	if !includeDisabled {
		if u.IsDisabled() {
			return nil, errors.New("user with name is " + userName + " is disabled")
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
	err := um.db.Users().ID(userID).Omit("profiles").Get(&u.u)
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
	condRoles := "exists (select * from " + um.db.UsersAndRoles().Name() + " as users_roles " +
		" where users_roles.role_id = " + um.db.Roles().Name() + ".id and users_roles.user_id = ?)"
	err := um.db.Roles().Where(condRoles, u.ID()).All(&u.roles)
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

	err = um.db.PermissionGroupsAndRoles().Where(orm.Cond{"role_id IN": roleIDs}).All(&u.permissionsAndRoles)
	if err != nil {
		return errors.Wrap(err, "query permissions and roles with user is "+u.Name()+" fail")
	}
	return nil
}

type user struct {
	um                  *userManager
	permissionsAndRoles []PermissionGroupAndRole
	u                   User
	roles               []Role
	roleNames           []string
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

func (u *user) WriteProfile(key, value string) error {
	if err := u.readProfiles(); err != nil {
		return err
	}

	if value == "" {
		if len(u.u.Profiles) == 0 {
			return nil
		}

		updateStr := `UPDATE hengwei_users SET profiles = profiles -$1::text WHERE id = $2`
		_, err := u.um.db.Exec(updateStr, key, u.ID())
		if err != nil {
			return errors.Wrap(err, "WriteProfile")
		}
		delete(u.u.Profiles, key)
		return nil
	}

	updateStr := `UPDATE hengwei_users SET profiles = profiles || jsonb_build_object($1::text, $2::text) WHERE id = $3`
	if len(u.u.Profiles) == 0 {
		updateStr = `UPDATE hengwei_users SET profiles = jsonb_build_object($1::text, $2::text) WHERE id = $3`
	}

	_, err := u.um.db.Exec(updateStr, key, value, u.ID())
	if err != nil {
		return errors.Wrap(err, "WriteProfile")
	}
	u.u.Profiles[key] = value
	return nil
}

func (u *user) readProfiles() error {
	if u.u.Profiles != nil {
		return nil
	}

	var txt []byte //sql.NullString
	err := u.um.db.Engine.DB().DB.QueryRow("SELECT profiles FROM hengwei_users WHERE id = $1", u.ID()).Scan(&txt)
	if err != nil {
		return errors.Wrap(err, "readProfiles")
	}
	if len(txt) != 0 {
		err = json.Unmarshal(txt, &u.u.Profiles)
		if err != nil {
			return errors.Wrap(err, "readProfiles")
		}
	}

	if u.u.Profiles == nil {
		u.u.Profiles = map[string]interface{}{}
	}
	return nil
}

func (u *user) ReadProfile(key string) (interface{}, error) {
	if err := u.readProfiles(); err != nil {
		return nil, err
	}

	return u.u.Profiles[key], nil
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
		log.Println("[permissions] permission group with id is", groupID, "isn't found.")
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
