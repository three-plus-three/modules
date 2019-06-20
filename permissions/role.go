package permissions

import (
	"time"

	"github.com/three-plus-three/modules/toolbox"
)

type Role struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name unique notnull"`
	Description string    `json:"description,omitempty" xorm:"description"`
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
