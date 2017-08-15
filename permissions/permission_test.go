package permissions

import (
	"cn/com/hengwei/commons/env_tests"
	"testing"

	fixtures "github.com/AreaHQ/go-fixtures"
	"github.com/three-plus-three/modules/web_ext"
)

func TestHasPermission(t *testing.T) {
	env := env_tests.Clone(nil)

	lifecycle, err := web_ext.NewLifecycle(env)
	if err != nil {
		t.Error(err)
		return
	}
	readUser := InitUser(lifecycle)

	if err := DropTables(lifecycle.ModelEngine); err != nil {
		t.Error(err)
	}

	if err := InitTables(lifecycle.ModelEngine); err != nil {
		t.Error(err)
	}

	if err := fixtures.LoadFiles([]string{
		"fixtures/users.yaml",
		"fixtures/roles.yaml",
		"fixtures/permission_groups.yaml",
		"fixtures/permissions_and_roles.yaml",
		"fixtures/users_and_roles.yaml",
	}, lifecycle.ModelEngine.DB().DB, "postgres"); err != nil {
		t.Error(err)
		return
	}

	u := readUser("t7o")
	if u.HasPermission("p1", CREATE) {
		t.Error("except no p1 create")
	}
	if !u.HasPermission("p11", CREATE) {
		t.Error("except has p11 create")
	}
}
