package permissions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/revel/revel"
	"github.com/three-plus-three/modules/netutil"
	"github.com/three-plus-three/modules/web_ext"
)

type OnlineUser struct {
	UserID    int64     `json:"user_id" xorm:"user_id pk"`
	AuthID    string    `json:"auth_id,omitempty" xorm:"auth_id unique"`
	Address   string    `json:"address" xorm:"address"`
	CreatedAt time.Time `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (onlineUser *OnlineUser) TableName() string {
	return "hengwei_online_users"
}

func KeyForOnlineUsers(key string) string {
	switch key {
	case "user_id":
		return "onlineUser.UserID"
	case "auth_id":
		return "onlineUser.AuthID"
	case "address":
		return "onlineUser.Address"
	case "created_at":
		return "onlineUser.CreatedAt"
	case "updated_at":
		return "onlineUser.UpdatedAt"
	}
	return key
}

const (
	UserNormal   = 0
	ItsmReporter = 1
)

type User struct {
	ID          int64                  `json:"id" xorm:"id pk autoincr"`
	Name        string                 `json:"name" xorm:"name unique notnull"`
	Nickname    string                 `json:"nickname" xorm:"nickname unique notnull"`
	Password    string                 `json:"password,omitempty" xorm:"password null"`
	Description string                 `json:"description,omitempty" xorm:"description"`
	Attributes  map[string]interface{} `json:"attributes" xorm:"attributes jsonb"`
	Profiles    map[string]interface{} `json:"profiles" xorm:"profiles jsonb"`
	Source      string                 `json:"source,omitempty" xorm:"source"`
	// Type        int                    `json:"type,omitempty" xorm:"type"`
	Disabled  bool       `json:"disabled,omitempty" xorm:"disabled"`
	LockedAt  *time.Time `json:"locked_at,omitempty" xorm:"locked_at null"`
	CreatedAt time.Time  `json:"created_at,omitempty" xorm:"created_at created"`
	UpdatedAt time.Time  `json:"updated_at,omitempty" xorm:"updated_at updated"`
}

func (user *User) IsDisabled() bool {
	return user.Disabled // || user.Type == ItsmReporter
}

func (user *User) TableName() string {
	return "hengwei_users"
}

func (user *User) IsBuiltin() bool {
	return user.Name == web_ext.UserAdmin ||
		user.Name == web_ext.UserGuest ||
		user.Name == web_ext.UserTPTNetwork
}

func (user *User) IsHidden() bool {
	return user.Name == web_ext.UserTPTNetwork // || user.Type == ItsmReporter
}

func (user *User) Validate(validation *revel.Validation) bool {
	validation.Required(user.Name).Key("user.Name")
	validation.Required(user.Nickname).Key("user.Nickname")
	if user.Source != "ldap" {
		validation.MinSize(user.Password, 8).Key("user.Password")
		validation.MaxSize(user.Password, 250).Key("user.Password")
	}

	o := user.Attributes["white_address_list"]
	if o != nil {
		var ss = toStrings(o)
		if len(ss) != 0 {
			_, err := netutil.ToCheckers(ss)
			if err != nil {
				validation.Error(err.Error()).Key("user.Attributes[white_address_list]")
			}
		}
	}
	return validation.HasErrors()
}

func toStrings(o interface{}) []string {
	if ss, ok := o.([]string); ok {
		return ss
	}

	if ss, ok := o.([]interface{}); ok {
		var ipList []string
		for _, i := range ss {
			ipList = append(ipList, fmt.Sprint(i))
		}
		return ipList
	}

	s, ok := o.(string)
	if !ok {
		bs, ok := o.([]byte)
		if !ok {
			panic(fmt.Errorf("o is unsupport type - %T %s", o, o))
		}
		s = string(bs)
	}

	var ipList []string
	if err := json.Unmarshal([]byte(s), &ipList); err == nil {
		return ipList
	}

	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		bs := scanner.Bytes()
		if len(bs) == 0 {
			continue
		}
		bs = bytes.TrimSpace(bs)
		if len(bs) == 0 {
			continue
		}
		ipList = append(ipList, string(bs))
	}
	if err := scanner.Err(); err != nil {
		panic(netutil.ErrInvalidIPRange)
	}
	return ipList
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

type UserAndUserGroup struct {
	ID      int64 `json:"id" xorm:"id pk autoincr"`
	UserID  int64 `json:"user_id" xorm:"user_id notnull"`
	GroupID int64 `json:"group_id" xorm:"group_id notnull"`
}

func (userAndUserGroup *UserAndUserGroup) TableName() string {
	return "hengwei_users_and_user_groups"
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
