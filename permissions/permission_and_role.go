package permissions

import (
	"time"

	"github.com/revel/revel"
	"github.com/three-plus-three/modules/web_ext"
)

type PermissionGroup struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name unique notnull"`
	Description string    `json:"description,omitempty" xorm:"description"`
	ParentID    int64     `json:"parent_id,omitempty" xorm:"parent_id"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (pg *PermissionGroup) TableName() string {
	return "hengwei_permission_groups"
}

func (pg *PermissionGroup) Validate(validation *revel.Validation) bool {
	validation.Required(pg.Name).Key("permissionGroup.Name")
	validation.MaxSize(pg.Description, 2000).Key("permissionGroup.Description")
	return validation.HasErrors()
}

func KeyForPermissionsGroups(key string) string {
	switch key {
	case "id":
		return "permissionGroup.ID"
	case "name":
		return "permissionGroup.Name"
	case "description":
		return "permissionGroup.Description"
	case "parent_id":
		return "permissionGroup.ParentId"
	case "operation":
		return "permissionGroup.Operation"
	case "created_at":
		return "permissionGroup.CreatedAt"
	case "updated_at":
		return "permissionGroup.UpdatedAt"
	}
	return key
}

type PermissionAndGroup struct {
	ID           int64  `json:"id" xorm:"id pk autoincr"`
	GroupID      int64  `json:"group_id" xorm:"group_id notnull"`
	PermissionID string `json:"permission_id" xorm:"permission_id notnull"`
}

func (pag *PermissionAndGroup) TableName() string {
	return "hengwei_permissions_and_groups"
}

const CREATE = web_ext.CREATE
const DELETE = web_ext.DELETE
const UPDATE = web_ext.UPDATE
const QUERY = web_ext.QUERY

type PermissionGroupAndRole struct {
	ID              int64 `json:"id" xorm:"id pk autoincr"`
	GroupID         int64 `json:"group_id" xorm:"group_id notnull"`
	RoleID          int64 `json:"role_id" xorm:"role_id notnull"`
	CreateOperation bool  `json:"create_operation,omitempty" xorm:"create_operation"`
	DeleteOperation bool  `json:"delete_operation,omitempty" xorm:"delete_operation"`
	UpdateOperation bool  `json:"update_operation,omitempty" xorm:"update_operation"`
	QueryOperation  bool  `json:"query_operation,omitempty" xorm:"query_operation"`
}

func (gap *PermissionGroupAndRole) TableName() string {
	return "hengwei_permission_groups_and_roles"
}
