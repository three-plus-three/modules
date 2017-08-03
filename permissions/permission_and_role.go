package permissions

import (
	"time"

	"github.com/revel/revel"
)

type PermissionGroup struct {
	ID          int64        `json:"id" xorm:"id pk autoincr"`
	Name        string       `json:"name" xorm:"name unique notnull"`
	Permissions []Permission `json:"permissions" xorm:"-"`
	Description string       `json:"description,omitempty" xorm:"description"`
	ParentID    int64        `json:"parent_id,omitempty" xorm:"parent_id"`
	Operation   string       `json:"operation" xorm:"-"`
	CreatedAt   time.Time    `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty" xorm:"updated_at updated"`
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
	ID           int64 `json:"id" xorm:"id pk autoincr"`
	GroupID      int64 `json:"group_id" xorm:"group_id notnull"`
	PermissionID int64 `json:"permission_id" xorm:"permission_id notnull"`
}

func (pag *PermissionAndGroup) TableName() string {
	return "hengwei_permissions_and_groups"
}

const Create = "create"
const Delete = "delete"
const Update = "update"
const Query = "query"

type PermissionGroupAndRole struct {
	ID         int64 `json:"id" xorm:"id pk autoincr"`
	GroupID    int64 `json:"group_id" xorm:"group_id notnull"`
	RoleID     int64 `json:"role_id" xorm:"role_id notnull"`
	AddTable   bool  `json:"add_table,omitempty" xorm:"add_table"`
	DelTable   bool  `json:"del_table,omitempty" xorm:"del_table"`
	EditTable  bool  `json:"edit_table,omitempty" xorm:"edit_table"`
	QueryTable bool  `json:"query_table,omitempty" xorm:"query_table"`
}

func (gap *PermissionGroupAndRole) TableName() string {
	return "hengwei_permission_group_and_roles"
}
