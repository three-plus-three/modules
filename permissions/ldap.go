package permissions

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/three-plus-three/modules/environment"

	"gopkg.in/ldap.v2"
)

type ADConfig struct {
	Address   string
	EnableTLS bool
	Username  string
	Password  string
	BaseDN    string
	Filter    string
}

//读取AD的配置
func readLDAPConfig(env *environment.Environment) (ADConfig, error) {
	ldapServer := env.Config.StringWithDefault("users.ldap_address", "")
	ldapTLS := env.Config.BoolWithDefault("users.ldap_tls", false)

	ldapFilter := env.Config.StringWithDefault("users.ldap_filter", "(&(objectClass=organizationalPerson))")
	ldapUsername := env.Config.StringWithDefault("users.ldap_username", "")
	ldapPassword := env.Config.StringWithDefault("users.ldap_password", "")
	ldapDN := env.Config.StringWithDefault("users.ldap_dn", "")

	return ADConfig{
		Address:   ldapServer,
		EnableTLS: ldapTLS,
		Username:  ldapUsername,
		Password:  ldapPassword,
		BaseDN:    ldapDN,
		Filter:    ldapFilter,
	}, nil
}

func ReadUserFromLDAP(env *environment.Environment) ([]User, error) {
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
		err = l.StartTLS(&tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return nil, err
		}
	}

	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return nil, err
	}

	ldapRolesFieldName := env.Config.StringWithDefault("users.ldap_roles", "memberOf")

	//获取数据
	sr, err := l.Search(ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		cfg.Filter,
		[]string{},
		nil,
	))
	if err != nil {
		return nil, err
	}

	var users = make([]User, 0, len(sr.Entries))
	for i := 0; i < len(sr.Entries); i++ {
		var roles []string
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
				roles = append(roles, fmt.Sprintf("%#v", dn.RDNs[0].Attributes[0].Value))
			}
		}

		users = append(users, User{
			Name:        sr.Entries[i].GetAttributeValue("name"),
			Description: sr.Entries[i].GetAttributeValue("description"),
			Attributes: map[string]interface{}{
				"roles":           roles,
				"streetAddress":   sr.Entries[i].GetAttributeValue("streetAddress"),
				"company":         sr.Entries[i].GetAttributeValue("company"),
				"telephoneNumber": sr.Entries[i].GetAttributeValue("telephoneNumber")},
			Source:    "AD",
			CreatedAt: convertToTime(strings.Split(sr.Entries[i].GetAttributeValue("whenCreated"), ".")[0]),
			UpdatedAt: convertToTime(strings.Split(sr.Entries[i].GetAttributeValue("whenChanged"), ".")[0]),
		})
	}
	return users, nil
}

func convertToTime(str string) time.Time {
	if len(str) > 12 {
		return time.Time{}
	}
	theTime, _ := time.ParseInLocation("20060102150405", str, time.Local)
	return theTime
}
