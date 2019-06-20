package permission_ext

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/revel/revel"
	"github.com/three-plus-three/modules/netutil"
	"github.com/three-plus-three/modules/permissions"
)

func ValidatePermissionGroup(pg *permissions.PermissionGroup, validation *revel.Validation) bool {
	validation.Required(pg.Name).Key("permissionGroup.Name")
	validation.MaxSize(pg.Description, 2000).Key("permissionGroup.Description")
	return validation.HasErrors()
}

func ValidateRole(role *permissions.Role, validation *revel.Validation) bool {
	validation.Required(role.Name).Key("role.Name")
	return validation.HasErrors()
}

func ValidateUserAndRole(userAndRole *permissions.UserAndRole, validation *revel.Validation) bool {
	validation.Required(userAndRole.UserID).Key("userAndRole.UserID")
	validation.Required(userAndRole.RoleID).Key("userAndRole.RoleID")
	return validation.HasErrors()
}

func ValidateUser(user *permissions.User, validation *revel.Validation) bool {
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

func ValidateUserGroup(userGroup *permissions.UserGroup, validation *revel.Validation) bool {
	validation.Required(userGroup.Name).Key("userGroup.Name")
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

		fields := bytes.Split(bs, []byte(","))
		for _, field := range fields {
			if len(field) == 0 {
				continue
			}
			field = bytes.TrimSpace(field)
			if len(field) == 0 {
				continue
			}

			ipList = append(ipList, string(field))
		}
	}
	if err := scanner.Err(); err != nil {
		panic(netutil.ErrInvalidIPRange)
	}
	return ipList
}
