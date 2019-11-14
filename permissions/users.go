package permissions

import (
	"github.com/three-plus-three/modules/users"
)

const (
	UserNormal   = users.UserNormal
	ItsmReporter = users.ItsmReporter
)

type OnlineUser = users.OnlineUser
type User = users.User
type UserProfile = users.UserProfile
type UserAndUserGroup = users.UserAndUserGroup
type UserGroup = users.UserGroup
type Role = users.Role
type UserAndRole = users.UserAndRole

func KeyForUsersAndRoles(key string) string {
	switch key {
	case "id":
		return "userAndRole.ID"
	case "user_id":
		return "userAndRole.UserID"
	case "role_id":
		return "userAndRole.RoleID"
	}
	return key
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
