package permissions

import (
	"testing"

	fixtures "github.com/AreaHQ/go-fixtures"
	"github.com/three-plus-three/modules/environment/env_tests"
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

	permissionProvider := PermissionProviderFunc{
		Permissions: func() ([]Permission, error) {
			allPermissions := []Permission{Permission{"um_1", "1", "2", []string{"um"}},
				Permission{"um_2", "2", "2", []string{"um"}},
				Permission{"as_1", "3", "2", []string{"as"}},
				Permission{"as_2", "4", "2", []string{"as"}}}
			return allPermissions, nil
		}}

	RegisterPermissions(permissionProvider)

	u := readUser("admin")
	if !u.HasPermission("perm_not_exists_in_db", CREATE) {
		t.Error("admin 有任何权限")
	}

	u = readUser("adm")
	if !u.HasPermission("perm_not_exists_in_db", CREATE) {
		t.Error("有 administrator 角色的用户有任何权限")
	}

	if !u.HasPermission("p12", CREATE) {
		t.Error("有 administrator 角色的用户有任何权限")
	}

	u = readUser("viewer")
	if !u.HasPermission("perm_not_exists_in_db", QUERY) {
		t.Error("有 visitor 角色的用户有任何读权限")
	}

	if !u.HasPermission("p12", QUERY) {
		t.Error("有 visitor 角色的用户有任何读权限")
	}

	if u.HasPermission("perm_not_exists_in_db", UPDATE) {
		t.Error("有 visitor 角色的用户没有任何写权限")
	}

	if u.HasPermission("p12", UPDATE) {
		t.Error("有 visitor 角色的用户没有任何写权限")
	}

	u = readUser("t7o")
	if u.HasPermission("p1", CREATE) {
		t.Error("except no p1 create")
	}

	if !u.HasPermission("p12", CREATE) {
		t.Error("except has p11 create")
	}

	// 1个用户有1个角色 关联父子关系的两个权限组 操作相同 其权限相同
	if !u.HasPermission("p11", CREATE) {
		t.Error("1个用户有1个角色 关联父子关系的两个权限组 操作相同 其权限相同")
	}

	//用户有1个角色 关联父子关系的两个权限组  与角色关联的操作不相同  两组权限相同
	if !u.HasPermission("p11", UPDATE) {
		t.Error("用户有1个角色 关联父子关系的两个权限组  与角色关联的操作不相同  两组权限相同")
	}

	//1个用户有1个角色 关联父子关系的两个权限组 操作不相同 其权限不相同 查父组权限
	if !u.HasPermission("p12", UPDATE) {
		t.Error("1个用户有1个角色 关联父子关系的两个权限组 操作不相同 其权限不相同 查父组权限")
	}
	//1个用户有1个角色 关联父子关系的两个权限组 操作不相同 其权限不相同 查子组权限
	if !u.HasPermission("p32", UPDATE) {
		t.Error("1个用户有1个角色 关联父子关系的两个权限组 操作不相同 其权限不相同 查子组权限")
	}
	u = readUser("t71")
	//1个用户有1个角色 关联无上下关系两个权限组 操作不相同 其权限相同
	if !u.HasPermission("p12", UPDATE) {
		t.Error("1个用户有1个角色 关联无上下关系两个权限组 操作不相同 其权限相同")
	}
	u = readUser("t72")
	//1个用户有2个角色  关联同一个权限组  操作不同
	if !u.HasPermission("p22", UPDATE) {
		t.Error("1个用户有2个角色  关联同一个权限组  操作不同")
	}
	//1个用户有2个角色  关联不同权限组  操作不同  两组权限相同
	if !u.HasPermission("p12", UPDATE) {
		t.Error("1个用户有2个角色  关联不同权限组  操作不同  两组权限相同")
	}
	//1个用户有2个角色  关联有父子关系权限组  操作不相同   两组权限相同
	if !u.HasPermission("p11", DELETE) {
		t.Error("1个用户有2个角色  关联有父子关系权限组  操作不相同   两组权限相同")
	}

	u = readUser("t73")
	//1个用户有2个角色  角色1关联父子组  角色2关联父组     操作全不相同   父子组权限不相同 	查询的是父组权限
	if !u.HasPermission("p13", UPDATE) {
		t.Error("1个用户有2个角色  角色1关联父子组  角色2关联父组 操作全不相同   父子组权限不相同 	查询的是父组权限")
	}

	//1个用户有2个角色  角色1关联父子组  角色2关联子组     操作全不相同   父子组权限不相同    查询的是子组权限
	if !u.HasPermission("p32", UPDATE) {
		t.Error("1个用户有2个角色  角色1关联父子组  角色2关联父组 操作全不相同   父子组权限不相同 	查询的是父组权限")
	}

	//1个用户有2个角色  角色1关联父子组  角色2关联父组     操作全不相同   父子组权限相同
	if !u.HasPermission("p11", UPDATE) {
		t.Error("1个用户有2个角色  角色1关联父子组  角色2关联父组 操作全不相同   父子组权限不相同 	查询的是父组权限")
	}

	u = readUser("A1")
	//用户关联1个角色
	if !u.HasPermission("um_1", CREATE) {
		t.Error("权限组与权限的Tags关联")
	}

	if !u.HasPermission("um_2", CREATE) {
		t.Error("权限组与权限的Tags关联")
	}

	//用户关联两个角色  关联同一个权限组  操作不同  tags相同
	if !u.HasPermission("um_1", UPDATE) {
		t.Error("用户关联两个角色  关联同一个权限组  操作不同")
	}

	if !u.HasPermission("um_2", CREATE) {
		t.Error("用户关联两个角色  关联同一个权限组  操作不同")
	}

	//用户关联两个角色  关联父子权限组  操作不同  tags相同
	if !u.HasPermission("um_1", UPDATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作不同  tags相同")
	}

	//用户关联两个角色  关联父子权限组  操作相同  tags相同
	if !u.HasPermission("um_1", CREATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作相同  tags相同")
	}

	//用户关联两个角色  关联父子权限组  操作相同  tags不相同
	if !u.HasPermission("um_1", CREATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作相同  tags相同")
	}

	//用户关联一个角色  关联一个权限组  操作相同  tags不相同
	if !u.HasPermission("as_1", CREATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作相同  tags不相同")
	}

	if !u.HasPermission("as_2", CREATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作相同  tags不相同")
	}

	//用户关联一个角色  关联一个权限组  操作相同  tags不相同
	if !u.HasPermission("um_1", CREATE) {
		t.Error("用户关联两个角色  关联父子权限组  操作相同  tags不相同")
	}

}
