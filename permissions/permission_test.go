// nolint
package permissions

import (
	"testing"

	fixtures "github.com/AreaHQ/go-fixtures"
	"github.com/runner-mei/orm"
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

	allPermissions := []Permission{Permission{"um_1", "1", "2", []string{"um"}},
		Permission{"um_2", "2", "2", []string{"um"}},
		Permission{"as_1", "3", "2", []string{"as"}},
		Permission{"as_2", "4", "2", []string{"as"}}}

	RegisterPermissions("um_bultin_test1",
		PermissionProviderFunc(func() (*PermissionData, error) {
			return &PermissionData{
				Permissions: allPermissions,
			}, nil
		}))

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

	u = readUser("t70")
	if !u.HasPermission("p12", CREATE) {
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
	if !u.HasPermission("p12", QUERY) {
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

func TestSaveDefaultPermissionGroups(t *testing.T) {
	env := env_tests.Clone(nil)
	lifecycle, err := web_ext.NewLifecycle(env)
	if err != nil {
		t.Error(err)
		return
	}

	var db = DB{Engine: lifecycle.ModelEngine}
	DropTables(lifecycle.ModelEngine)
	InitTables(lifecycle.ModelEngine)

	var allPermissions = []Permission{Permission{"um_1", "1", "2", []string{"um"}},
		Permission{"um_2", "2", "2", []string{"um"}},
		Permission{"um_3", "3", "2", []string{"um"}}}

	var allGroups = []Group{Group{Name: "分组1", Children: []Group{
		Group{Name: "分组1-1", PermissionIDs: []string{"um_3"}},
		Group{Name: "分组1-2", PermissionIDs: []string{"um_2"}}}}}

	RegisterPermissions("um_bultin1",
		PermissionProviderFunc(func() (*PermissionData, error) {
			return &PermissionData{
				Permissions: allPermissions,
				Groups:      allGroups,
			}, nil
		}))

	//测试是否通过
	err = saveDefaultPermissionGroups(&db, allGroups)
	if err != nil {
		t.Error(err)
	}

	var group1 PermissionGroup
	err = db.PermissionGroups().Where(orm.Cond{"name": "分组1"}).One(&group1)
	if err != nil {
		t.Error(err)
	}
	var group2 PermissionGroup
	err = db.PermissionGroups().Where(orm.Cond{"name": "分组1-1"}).One(&group2)
	if err != nil {
		t.Error(err)
	}

	var group3 PermissionGroup
	err = db.PermissionGroups().Where(orm.Cond{"name": "分组1-2"}).One(&group3)
	if err != nil {
		t.Error(err)
	}

	var groupAndPermission1 PermissionAndGroup
	err = db.PermissionsAndGroups().Where(orm.Cond{"group_id": group2.ID, "permission_object": "um_3"}).One(&groupAndPermission1)
	if err != nil {
		t.Error(err)
	}

	var groupAndPermission2 PermissionAndGroup
	err = db.PermissionsAndGroups().Where(orm.Cond{"group_id": group3.ID, "permission_object": "um_2"}).One(&groupAndPermission2)
	if err != nil {
		t.Error(err)
	}

	//测试删除组
	var allGroupA = []Group{Group{Name: "分组1", Children: []Group{
		Group{Name: "分组1-1", PermissionIDs: []string{"um_3"}}}}}
	err = saveDefaultPermissionGroups(&db, allGroupA)
	if err != nil {
		t.Error(err)
	}

	var group PermissionGroup
	err = db.PermissionGroups().Where(orm.Cond{"name": "分组1-2"}).One(&group)
	if err != nil {
		if err != orm.ErrNotFound {
			t.Error(err)
		}
	}

	if group.Name != "" {
		t.Error("删除组失败")
	}

	//测试增加组
	var allGroupB = []Group{Group{Name: "分组1", Children: []Group{
		Group{Name: "分组1-1", PermissionIDs: []string{"um_3"}},
		Group{Name: "分组1-2", PermissionIDs: []string{"um_2"}}}}}

	err = saveDefaultPermissionGroups(&db, allGroupB)
	if err != nil {
		t.Error(err)
	}
	var groupB PermissionGroup
	err = db.PermissionGroups().Where(orm.Cond{"name": "分组1-2"}).One(&groupB)
	if err != nil {
		t.Error(err)
	}

	if groupB.Name == "" {
		t.Error(err)
	}

	//增加权限
	var allGroupC = []Group{Group{Name: "分组1", Children: []Group{
		Group{Name: "分组1-1", PermissionIDs: []string{"um_3", "um_1"}},
		Group{Name: "分组1-2", PermissionIDs: []string{"um_2"}}}}}

	err = saveDefaultPermissionGroups(&db, allGroupC)
	if err != nil {
		t.Error(err)
	}

	var permissionAndGroupA PermissionAndGroup
	err = db.PermissionsAndGroups().Where(orm.Cond{"permission_object": "um_1"}).One(&permissionAndGroupA)
	if err != nil {
		t.Error(err)
	}

	if permissionAndGroupA.ID == 0 {
		t.Error("添加失败")
	}

	//删除权限
	var allGroupD = []Group{Group{Name: "分组1", Children: []Group{
		Group{Name: "分组1-1", PermissionIDs: []string{"um_3"}},
		Group{Name: "分组1-2", PermissionIDs: []string{"um_2"}}}}}

	err = saveDefaultPermissionGroups(&db, allGroupD)
	if err != nil {
		t.Error(err)
	}

	var permissionAndGroup PermissionAndGroup
	err = db.PermissionsAndGroups().Where(orm.Cond{"permission_object": "um_1"}).One(&permissionAndGroup)
	if err != nil {
		if err != orm.ErrNotFound {
			t.Error(err)
		}
	}

	if permissionAndGroup.PermissionObject != "" {
		t.Error("没有删除权限")
	}
}
