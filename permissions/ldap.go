package permissions

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/three-plus-three/modules/environment"
	ldap "gopkg.in/ldap.v3"
)

func IsConnectError(err error) bool {
	if ldapErr, ok := err.(*ldap.Error); ok {
		if opErr, ok := ldapErr.Err.(*net.OpError); ok && opErr.Op == "dial" {
			return true
		}
	}
	return false
}

func IsPasswordError(err error) bool {
	if ldapErr, ok := err.(*ldap.Error); ok {
		return ldapErr.ResultCode == ldap.LDAPResultInvalidCredentials
	}
	return false
}

func IsInsufficientAccessRightsError(err error) bool {
	if ldapErr, ok := err.(*ldap.Error); ok {
		return ldapErr.ResultCode == ldap.LDAPResultInsufficientAccessRights ||
			ldapErr.ResultCode == ldap.LDAPResultInappropriateAuthentication
	}
	return false
}

type ldapConfig struct {
	Address    string
	EnableTLS  bool
	BaseDN     string
	UserFilter string
	RoleFilter string
	UserFormat string
}

//读取AD的配置
func readLDAPConfig(env *environment.Environment) (ldapConfig, error) {
	ldapServer := env.Config.StringWithDefault("users.ldap_address", "")
	ldapTLS := env.Config.BoolWithDefault("users.ldap_tls", false)

	ldapFilter := env.Config.StringWithDefault("users.ldap_filter", "(&(objectClass=organizationalPerson))")
	ldapRoleFilter := env.Config.StringWithDefault("users.ldap_role_filter", "(&(objectClass=group))")
	ldapDN := env.Config.StringWithDefault("users.ldap_base_dn", "")
	ldapUserFormat := env.Config.StringWithDefault("users.ldap_user_format", "")
	if ldapUserFormat == "" {
		if ldapDN != "" {
			ldapUserFormat = "cn=%s," + ldapDN
		} else {
			ldapUserFormat = "%s"
		}
	}
	return ldapConfig{
		Address:    ldapServer,
		EnableTLS:  ldapTLS,
		BaseDN:     ldapDN,
		UserFilter: ldapFilter,
		RoleFilter: ldapRoleFilter,
		UserFormat: ldapUserFormat,
	}, nil
}

func ReadUserFromLDAP(env *environment.Environment, username, password string, fields map[string]string) ([]User, error) {
	cfg, err := readLDAPConfig(env)
	if err != nil {
		return nil, err
	}

	//连接活动目录
	l, err := ldap.Dial("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	if cfg.EnableTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}) // nolint
		if err != nil {
			return nil, err
		}
	}

	err = l.Bind(fmt.Sprintf(cfg.UserFormat, username), password)
	if err != nil {
		return nil, err
	}

	ldapRolesFieldName := env.Config.StringWithDefault("users.ldap_roles", "memberOf")
	exceptedRole := env.Config.StringWithDefault("users.ldap_login_role", "")

	//获取数据
	sr, err := l.Search(ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		cfg.UserFilter,
		[]string{},
		nil,
	))
	if err != nil {
		return nil, err
	}

	var users = make([]User, 0, len(sr.Entries))
	for i := 0; i < len(sr.Entries); i++ {
		var roles []string
		var rawRoles []string
		if ldapRolesFieldName != "" {
			roleValues := sr.Entries[i].GetAttributeValues(ldapRolesFieldName)
			roles = make([]string, 0, len(roleValues))
			for _, roleName := range roleValues {
				dn, err := ldap.ParseDN(roleName)
				if err != nil {
					roles = append(roles, roleName)
					continue
				}

				if len(dn.RDNs) == 0 || len(dn.RDNs[0].Attributes) == 0 {
					continue
				}
				roles = append(roles, dn.RDNs[0].Attributes[0].Value)
			}
			rawRoles = roleValues
		}

		if exceptedRole != "" {
			found := false
			for _, role := range roles {
				if role == exceptedRole {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		attributes := map[string]interface{}{
			"roles":     roles,
			"raw_roles": rawRoles}
		for _, attr := range sr.Entries[i].Attributes {
			if newFieldName, ok := fields[attr.Name]; ok {
				attributes[newFieldName] = attr.Values
			}
		}
		attributes["roles"] = roles
		attributes["raw_roles"] = rawRoles

		users = append(users, User{
			Name:        sr.Entries[i].GetAttributeValue("name"),
			Description: sr.Entries[i].GetAttributeValue("description"),
			Attributes:  attributes,
			Source:      "ldap",
			CreatedAt:   convertToTime(strings.Split(sr.Entries[i].GetAttributeValue("whenCreated"), ".")[0]),
			UpdatedAt:   convertToTime(strings.Split(sr.Entries[i].GetAttributeValue("whenChanged"), ".")[0]),
		})
	}
	return users, nil
}

func ReadRolesFromLDAP(env *environment.Environment, username, password string) (map[string]string, error) {
	cfg, err := readLDAPConfig(env)
	if err != nil {
		return nil, err
	}

	//连接活动目录
	l, err := ldap.Dial("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	if cfg.EnableTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}) // nolint
		if err != nil {
			return nil, err
		}
	}

	err = l.Bind(fmt.Sprintf(cfg.UserFormat, username), password)
	if err != nil {
		return nil, err
	}

	//获取数据
	sr, err := l.Search(ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		cfg.RoleFilter,
		[]string{},
		nil,
	))
	if err != nil {
		return nil, err
	}

	roles := map[string]string{}
	for i := 0; i < len(sr.Entries); i++ {
		// sr.Entries[i].PrettyPrint(2)
		roles[sr.Entries[i].GetAttributeValue("name")] = sr.Entries[i].GetAttributeValue("description")
	}
	return roles, nil
}

func ReadUserFieldsFromLDAP(env *environment.Environment, username, password string) (map[string]string, error) {
	cfg, err := readLDAPConfig(env)
	if err != nil {
		return nil, err
	}

	//连接活动目录
	l, err := ldap.Dial("tcp", cfg.Address)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	if cfg.EnableTLS {
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true}) // nolint
		if err != nil {
			return nil, err
		}
	}

	err = l.Bind(fmt.Sprintf(cfg.UserFormat, username), password)
	if err != nil {
		return nil, err
	}

	//获取数据
	sr, err := l.Search(ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		cfg.UserFilter,
		[]string{},
		nil,
	))
	if err != nil {
		return nil, err
	}

	fields := map[string]string{}
	for i := 0; i < len(sr.Entries); i++ {
		for _, attr := range sr.Entries[i].Attributes {
			fields[attr.Name] = attr.Name
		}
	}
	return fields, nil
}

func convertToTime(str string) time.Time {
	if len(str) > 12 {
		return time.Time{}
	}
	theTime, _ := time.ParseInLocation("20060102150405", str, time.Local)
	return theTime
}
