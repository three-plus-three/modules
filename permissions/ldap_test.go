package permissions

import (
	"flag"
	"net"
	"testing"

	"github.com/three-plus-three/modules/environment/env_tests"
	ldap "gopkg.in/ldap.v3"
)

var testpassword = flag.String("test.ldap_password", "", "")

func TestLDAP(t *testing.T) {
	if *testpassword == "" {
		t.Skip("请设置  ldap 服务的密码")
	}
	env := env_tests.Clone(nil)

	env.Config.Set("users.ldap_address", "192.168.1.151:389")
	env.Config.Set("users.ldap_base_dn", "cn=Users,dc=hengwei,dc=com,dc=cn")
	env.Config.Set("users.ldap_user_format", "%s@hengwei.com.cn")

	data, err := ReadUserFromLDAP(env, "Administrator", *testpassword, nil)
	if err != nil {
		if ldapErr, ok := err.(*ldap.Error); ok {
			if opErr, ok := ldapErr.Err.(*net.OpError); ok && opErr.Op == "dial" {
				t.Skip("skip ldap test, please ldap server is runring")
			}
		}
		t.Errorf("%T", err)
		t.Error(err)
	}

	t.Log(data)

	roles, err := ReadRolesFromLDAP(env, "Administrator", *testpassword)
	if err != nil {
		if ldapErr, ok := err.(*ldap.Error); ok {
			if opErr, ok := ldapErr.Err.(*net.OpError); ok && opErr.Op == "dial" {
				t.Skip("skip ldap test, please ldap server is runring")
			}
		}
		t.Errorf("%T", err)
		t.Error(err)
	}

	t.Log(roles)
}
