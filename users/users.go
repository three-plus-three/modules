package users

import (
	"time"

	"github.com/three-plus-three/modules/toolbox"
)

const (
	UserNormal   = 0
	ItsmReporter = 1
)

type OnlineUser struct {
	UserID    int64     `json:"user_id" xorm:"user_id pk"`
	Uuid      string    `json:"uuid,omitempty" xorm:"uuid unique"`
	Address   string    `json:"address" xorm:"address"`
	CreatedAt time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (onlineUser *OnlineUser) TableName() string {
	return "hengwei_online_users"
}

type Role struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name unique notnull"`
	Description string    `json:"description,omitempty" xorm:"description null"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (role *Role) IsBuiltin() bool {
	return role.Name == toolbox.RoleSuper ||
		role.Name == toolbox.RoleAdministrator ||
		role.Name == toolbox.RoleVisitor ||
		role.Name == toolbox.RoleGuest
}

func (role *Role) TableName() string {
	return "hengwei_roles"
}

type UserAndRole struct {
	ID     int64 `json:"id" xorm:"id pk autoincr"`
	UserID int64 `json:"user_id" xorm:"user_id unique(user_role)"`
	RoleID int64 `json:"role_id" xorm:"role_id unique(user_role) notnull"`
}

func (userAndRole *UserAndRole) TableName() string {
	return "hengwei_users_and_roles"
}

type User struct {
	ID          int64                  `json:"id" xorm:"id pk autoincr"`
	Name        string                 `json:"name" xorm:"name unique notnull"`
	Nickname    string                 `json:"nickname" xorm:"nickname unique notnull"`
	Password    string                 `json:"password,omitempty" xorm:"password null"`
	Description string                 `json:"description,omitempty" xorm:"description null"`
	Attributes  map[string]interface{} `json:"attributes" xorm:"attributes jsonb null"`
	Profiles    map[string]string      `json:"profiles" xorm:"profiles jsonb null"`
	Source      string                 `json:"source,omitempty" xorm:"source null"`
	Signature   string                 `json:"signature,omitempty" xorm:"signature null"`
	// Type        int                    `json:"type,omitempty" xorm:"type"`
	Disabled  bool       `json:"disabled,omitempty" xorm:"disabled null"`
	LockedAt  *time.Time `json:"locked_at,omitempty" xorm:"locked_at null"`
	CreatedAt time.Time  `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time  `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (user *User) IsDisabled() bool {
	return user.Disabled // || user.Type == ItsmReporter
}

func (user *User) TableName() string {
	return "hengwei_users"
}

func (user *User) IsBuiltin() bool {
	return user.Name == toolbox.UserAdmin ||
		user.Name == toolbox.UserGuest ||
		user.Name == toolbox.UserTPTNetwork
}

func (user *User) IsHidden() bool {
	return user.Name == toolbox.UserTPTNetwork // || user.Type == ItsmReporter
}

type UserProfile struct {
	ID    int64  `json:"id" xorm:"id pk unique(a)"`
	Name  string `json:"name" xorm:"name pk unique(a) notnull"`
	Value string `json:"value,omitempty" xorm:"value"`
}

func (user *UserProfile) TableName() string {
	return "hengwei_user_profiles"
}

type UserAndUserGroup struct {
	ID      int64 `json:"id" xorm:"id pk autoincr"`
	UserID  int64 `json:"user_id" xorm:"user_id notnull"`
	GroupID int64 `json:"group_id" xorm:"group_id notnull"`
}

func (userAndUserGroup *UserAndUserGroup) TableName() string {
	return "hengwei_users_and_user_groups"
}

type UserGroup struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name notnull"`
	Description string    `json:"description" xorm:"description null"`
	ParentID    int64     `json:"parent_id" xorm:"parent_id"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (userGroup *UserGroup) TableName() string {
	return "hengwei_user_groups"
}

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
