package permissions

import "github.com/revel/revel"

type UserAndRole struct {
	ID     int64 `json:"id" xorm:"id pk autoincr"`
	UserID int64 `json:"user_id" xorm:"user_id"`
	RoleID int64 `json:"role_id" xorm:"role_id notnull"`
}

func (userAndRole *UserAndRole) TableName() string {
	return "hengwei_users_and_roles"
}

func (userAndRole *UserAndRole) Validate(validation *revel.Validation) bool {
	validation.Required(userAndRole.UserID).Key("userAndRole.UserID")
	validation.Required(userAndRole.RoleID).Key("userAndRole.RoleID")
	return validation.HasErrors()
}

func KeyForUsersAndRoles(key string) string {
	switch key {
	case "id":
		return "userAndRole.ID"
	case "user_name":
		return "userAndRole.UserID"
	case "role_id":
		return "userAndRole.RoleID"
	}
	return key
}
