package permissions

import (
	"time"

	"github.com/revel/revel"
)

type User struct {
	ID          int64                  `json:"id" xorm:"id pk autoincr"`
	Name        string                 `json:"name" xorm:"name notnull"`
	Password    string                 `json:"password,omitempty" xorm:"password"`
	Description string                 `json:"description,omitempty" xorm:"description"`
	Attributes  map[string]interface{} `json:"attributes" xorm:"attributes jsonb"`
	Source      string                 `json:"source,omitempty" xorm:"-"`
	CreatedAt   time.Time              `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time              `json:"updated_at,omitempty" xorm:"updated_at updated"`
	Roles       []Role                 `json:"roles" xorm:"-"`
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
	case "attibutes":
		return "user.Attibutes"
	case "company":
		return "user.Attibutes.Company"
	case "number":
		return "user.Attibutes.Number"
	case "address":
		return "user.Attibutes.Address"
	case "created_at":
		return "user.CreatedAt"
	case "updated_at":
		return "user.UpdatedAt"
	}
	return key
}

type UserAndUserGroup struct {
	ID       int64  `json:"id" xorm:"id pk autoincr"`
	UserName string `json:"user_name" xorm:"user_name"`
	GroupID  int64  `json:"group_id" xorm:"group_id notnull"`
}

func (userAndUserGroup *UserAndUserGroup) TableName() string {
	return "hengwei_user_and_user_groups"
}

type UserGroup struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name notnull"`
	Description string    `json:"description" xorm:"description"`
	ParentID    int64     `json:"parent_id" xorm:"parent_id"`
	Users       []User    `json:"users" xorm:"-"`
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
		return "userGroup.ParentId"
	case "created_at":
		return "userGroup.CreatedAt"
	case "updated_at":
		return "userGroup.UpdatedAt"
	}
	return key
}
