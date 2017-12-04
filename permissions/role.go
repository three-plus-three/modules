package permissions

import (
	"time"

	"github.com/revel/revel"
	"github.com/three-plus-three/modules/web_ext"
)

type Role struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name unique notnull"`
	Description string    `json:"description,omitempty" xorm:"description"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (role *Role) IsBuiltin() bool {
	return role.Name == web_ext.RoleSuper ||
		role.Name == web_ext.RoleAdministrator ||
		role.Name == web_ext.RoleVisitor ||
		role.Name == web_ext.RoleGuest
}

func (role *Role) TableName() string {
	return "hengwei_roles"
}

func (role *Role) Validate(validation *revel.Validation) bool {
	validation.Required(role.Name).Key("role.Name")
	return validation.HasErrors()
}

func KeyForRoles(key string) string {
	switch key {
	case "id":
		return "role.ID"
	case "name":
		return "role.Name"
	case "description":
		return "role.Description"
	case "created_at":
		return "role.CreatedAt"
	case "updated_at":
		return "role.UpdatedAt"
	}
	return key
}
