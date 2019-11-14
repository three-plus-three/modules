package users

import (
	"net/http"
	"time"

	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
)

// 常用的错误
var (
	ErrUnauthorized       = errors.NewApplicationError(http.StatusUnauthorized, "user is unauthorized")
	ErrCacheInvalid       = errors.New("permission cache is invald")
	ErrTagNotFound        = errors.New("permission tag is not found")
	ErrPermissionNotFound = errors.New("permission is not found")
	ErrAlreadyClosed      = errors.New("server is closed")
)

// Group 缺省组信息
type Group struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`

	Children       []Group  `json:"children,omitempty"`
	PermissionIDs  []string `json:"permissions,omitempty"`
	PermissionTags []string `json:"tags,omitempty"`
}

// Permission 缺省权限对象
type Permission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,emitempty"`
	Tags        []string `json:"tags,emitempty"`
}

// Tag 标签对象
type Tag struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,emitempty"`

	Children []Tag `json:"children,omitempty"`
}

const PERMISSION_ID = 0
const PERMISSION_TAG = 1

const CREATE = toolbox.CREATE
const DELETE = toolbox.DELETE
const UPDATE = toolbox.UPDATE
const QUERY = toolbox.QUERY

type PermissionGroup struct {
	ID          int64     `json:"id" xorm:"id pk autoincr"`
	Name        string    `json:"name" xorm:"name unique(pname) notnull"`
	Description string    `json:"description,omitempty" xorm:"description"`
	IsDefault   bool      `json:"is_default" xorm:"is_default null"`
	ParentID    int64     `json:"parent_id,omitempty" xorm:"parent_id unique(pname) null"`
	CreatedAt   time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt   time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (pg *PermissionGroup) TableName() string {
	return "hengwei_permission_groups"
}

type PermissionAndGroup struct {
	ID               int64  `json:"id" xorm:"id pk autoincr"`
	GroupID          int64  `json:"group_id" xorm:"group_id notnull"`
	PermissionObject string `json:"permission_object" xorm:"permission_object notnull"`
	Type             int64  `json:"type" xorm:"type notnull"`
}

func (pag *PermissionAndGroup) TableName() string {
	return "hengwei_permissions_and_groups"
}

type PermissionGroupAndRole struct {
	ID              int64 `json:"id" xorm:"id pk autoincr"`
	GroupID         int64 `json:"group_id" xorm:"group_id unique(group_role) notnull"`
	RoleID          int64 `json:"role_id" xorm:"role_id unique(group_role) notnull"`
	CreateOperation bool  `json:"create_operation,omitempty" xorm:"create_operation"`
	DeleteOperation bool  `json:"delete_operation,omitempty" xorm:"delete_operation"`
	UpdateOperation bool  `json:"update_operation,omitempty" xorm:"update_operation"`
	QueryOperation  bool  `json:"query_operation,omitempty" xorm:"query_operation"`
}

func (gap *PermissionGroupAndRole) TableName() string {
	return "hengwei_permission_groups_and_roles"
}

type Permissions struct {
	PermissionGroup `xorm:"extends"`
	PermissionIDs   []string `xorm:"-"`
	PermissionTags  []string `xorm:"-"`
}

type PermGroupCache interface {
	Get(groupID int64) *Permissions
	GetChildren(groupID int64) []*Permissions
	GetPermissionsByTag(tag string) ([]Permission, error)
}
