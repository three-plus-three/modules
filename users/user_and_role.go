package users

type UserAndRole struct {
	ID     int64 `json:"id" xorm:"id pk autoincr"`
	UserID int64 `json:"user_id" xorm:"user_id unique(user_role)"`
	RoleID int64 `json:"role_id" xorm:"role_id unique(user_role) notnull"`
}

func (userAndRole *UserAndRole) TableName() string {
	return "hengwei_users_and_roles"
}

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
