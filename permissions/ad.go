package permissions

import (
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
}

//读取AD的配置
func readLDAPConfig(env *environment.Environment) (ADConfig, error) {
	ldapServer := env.Config.StringWithDefault("users.ldap_address", "")
	ldapTLS := env.Config.BoolWithDefault("users.ldap_tls", false)

	ldapUsername := env.Config.StringWithDefault("users.ldap_username", "")
	ldapPassword := env.Config.StringWithDefault("users.ldap_password", "")
	ldapDN := env.Config.StringWithDefault("users.ldap_dn", "")

	return ADConfig{
		Address:   ldapServer,
		EnableTLS: ldapTLS,
		Username:  ldapUsername,
		Password:  ldapPassword,
		BaseDN:    ldapDN,
	}, nil
}

func ReadUserFromLDAP(env *environment.Environment) ([]User, error) {
	var users []User
	cfg, err := readLDAPConfig(env)
	if err != nil {
		return nil, err
	}

	//连接活动目录
	l, err := ldap.Dial("tcp", cfg.Address)
	if err != nil {
		return users, err
	}
	defer l.Close()

	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return users, err
	}

	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(objectClass=organizationalPerson))",
		[]string{},
		nil,
	)

	//获取数据
	sr, err := l.Search(searchRequest)
	if err != nil {
		return users, err
	}

	for i := 0; i < len(sr.Entries); i++ {
		users = append(users, User{
			Name:        sr.Entries[i].GetAttributeValue("name"),
			Description: sr.Entries[i].GetAttributeValue("description"),
			Attributes: map[string]interface{}{
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
