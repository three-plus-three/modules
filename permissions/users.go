package permissions

import (
	"time"

	"github.com/revel/revel"
)

type OnlineUser struct {
	UserID    int64     `json:"user_id" xorm:"user_id pk"`
	Address   string    `json:"address" xorm:"address"`
	CreatedAt time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (onlineUser *OnlineUser) TableName() string {
	return "hengwei_online_users"
}

type User struct {
	ID          int64                  `json:"id" xorm:"id pk autoincr"`
	Name        string                 `json:"name" xorm:"name unique notnull"`
	Nickname    string                 `json:"nickname" xorm:"nickname unique notnull"`
	Password    string                 `json:"password,omitempty" xorm:"password null"`
	Description string                 `json:"description,omitempty" xorm:"description"`
	Attributes  map[string]interface{} `json:"attributes" xorm:"attributes jsonb"`
	Source      string                 `json:"source,omitempty" xorm:"source"`
	LockedAt    *time.Time             `json:"locked_at,omitempty" xorm:"locked_at null"`
	CreatedAt   time.Time              `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (user *User) TableName() string {
	return "hengwei_users"
}

func (user *User) Validate(validation *revel.Validation) bool {
	validation.Required(user.Name).Key("user.Name")
	if user.Source != "AD" {
		validation.MinSize(user.Password, 8).Key("user.Password")
		validation.MaxSize(user.Password, 250).Key("user.Password")
	}
	return validation.HasErrors()
}

func KeyForUsers(key string) string {
	switch key {
	case "id":
		return "user.ID"
	case "name":
		return "user.Name"
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
	Description string    `json:"description" xorm:"description"`
	ParentID    int64     `json:"parent_id" xorm:"parent_id"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (userGroup *UserGroup) TableName() string {
	return "hengwei_user_groups"
}

func (userGroup *UserGroup) Validate(validation *revel.Validation) bool {
	validation.Required(userGroup.Name).Key("userGroup.Name")
	return validation.HasErrors()
}

func KeyForUserGroup(key string) string {
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
