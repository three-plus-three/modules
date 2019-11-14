package users

import (
	"time"

	"github.com/three-plus-three/modules/toolbox"
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

func KeyForOnlineUsers(key string) string {
	switch key {
	case "user_id":
		return "onlineUser.UserID"
	case "Uuid":
		return "onlineUser.UUID"
	case "address":
		return "onlineUser.Address"
	case "created_at":
		return "onlineUser.CreatedAt"
	case "updated_at":
		return "onlineUser.UpdatedAt"
	}
	return key
}

const (
	UserNormal   = 0
	ItsmReporter = 1
)

type User struct {
	ID          int64                  `json:"id" xorm:"id pk autoincr"`
	Name        string                 `json:"name" xorm:"name unique notnull"`
	Nickname    string                 `json:"nickname" xorm:"nickname unique notnull"`
	Password    string                 `json:"password,omitempty" xorm:"password null"`
	Description string                 `json:"description,omitempty" xorm:"description null"`
	Attributes  map[string]interface{} `json:"attributes" xorm:"attributes jsonb null"`
	Profiles    map[string]interface{} `json:"profiles" xorm:"profiles jsonb null"`
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

func KeyForUsers(key string) string {
	switch key {
	case "id":
		return "user.ID"
	case "name":
		return "user.Name"
	case "nickname":
		return "user.Nickname"
	case "type":
		return "user.Type"
	case "password":
		return "user.Password"
	case "description":
		return "user.Description"
	case "source":
		return "user.Source"
	case "attibutes":
		return "user.Attibutes"
	case "created_at":
		return "user.CreatedAt"
	case "updated_at":
		return "user.UpdatedAt"
	}
	return key
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

func KeyForUsersAndUserGroups(key string) string {
	switch key {
	case "id":
		return "userAndUserGroup.ID"
	case "user_id":
		return "userAndUserGroup.UserID"
	case "group_id":
		return "userAndUserGroup.GroupID"
	}
	return key
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

func KeyForUserGroups(key string) string {
	switch key {
	case "id":
		return "userGroup.ID"
	case "name":
		return "userGroup.Name"
	case "description":
		return "userGroup.Description"
	case "parent_id":
		return "userGroup.ParentID"
	case "created_at":
		return "userGroup.CreatedAt"
	case "updated_at":
		return "userGroup.UpdatedAt"
	}
	return key
}
