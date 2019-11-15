// Please don't edit this file!
package usermodels

import (
	"context"
	"database/sql"
	"errors"
	"reflect"
	"strings"

	gobatis "github.com/runner-mei/GoBatis"
)

func init() {
	gobatis.Init(func(ctx *gobatis.InitContext) error {
		{ //// UserDao.CreateUser
			if _, exists := ctx.Statements["UserDao.CreateUser"]; !exists {
				sqlStr, err := gobatis.GenerateInsertSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{"user"},
					[]reflect.Type{
						reflect.TypeOf((*User)(nil)),
					}, false)
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.CreateUser error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.CreateUser",
					gobatis.StatementTypeInsert,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.CreateUser"] = stmt
			}
		}
		{ //// UserDao.DisableUser
			if _, exists := ctx.Statements["UserDao.DisableUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("UPDATE ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&User{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString("(user_id, role_id)\r\n       SET disabled = true WHERE id=#{id}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.DisableUser",
					gobatis.StatementTypeUpdate,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.DisableUser"] = stmt
			}
		}
		{ //// UserDao.EnableUser
			if _, exists := ctx.Statements["UserDao.EnableUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("UPDATE ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&User{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString("(user_id, role_id)\r\n       SET disabled = false WHERE id=#{id}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.EnableUser",
					gobatis.StatementTypeUpdate,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.EnableUser"] = stmt
			}
		}
		{ //// UserDao.UpdateUser
			if _, exists := ctx.Statements["UserDao.UpdateUser"]; !exists {
				sqlStr, err := gobatis.GenerateUpdateSQL(ctx.Dialect, ctx.Mapper,
					"user.", reflect.TypeOf(&User{}),
					[]string{
						"id",
					},
					[]reflect.Type{
						reflect.TypeOf(new(int64)).Elem(),
					})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.UpdateUser error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.UpdateUser",
					gobatis.StatementTypeUpdate,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.UpdateUser"] = stmt
			}
		}
		{ //// UserDao.AddRoleToUser
			if _, exists := ctx.Statements["UserDao.AddRoleToUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("INSERT INTO ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserAndRole{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString("(user_id, role_id)\r\n       VALUES(#{userid}, #{roleid})\r\n       ON CONFLICT (user_id, role_id)\r\n       DO UPDATE SET user_id=EXCLUDED.user_id, role_id=EXCLUDED.role_id")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.AddRoleToUser",
					gobatis.StatementTypeInsert,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.AddRoleToUser"] = stmt
			}
		}
		{ //// UserDao.RemoveRoleFromUser
			if _, exists := ctx.Statements["UserDao.RemoveRoleFromUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("DELETE FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserAndRole{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" WHERE user_id = #{userid} and role_id = #{roleid}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.RemoveRoleFromUser",
					gobatis.StatementTypeDelete,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.RemoveRoleFromUser"] = stmt
			}
		}
		return nil
	})
}

func NewUserDao(ref gobatis.SqlSession, userQueryer UserQueryer) UserDao {
	if ref == nil {
		panic(errors.New("param 'ref' is nil"))
	}
	if reference, ok := ref.(*gobatis.Reference); ok {
		if reference.SqlSession == nil {
			panic(errors.New("param 'ref.SqlSession' is nil"))
		}
	} else if valueReference, ok := ref.(gobatis.Reference); ok {
		if valueReference.SqlSession == nil {
			panic(errors.New("param 'ref.SqlSession' is nil"))
		}
	}
	return &UserDaoImpl{session: ref,
		UserQueryer: userQueryer}
}

type UserDaoImpl struct {
	UserQueryer
	session gobatis.SqlSession
}

func (impl *UserDaoImpl) CreateUser(ctx context.Context, user *User) (int64, error) {
	return impl.session.Insert(ctx, "UserDao.CreateUser",
		[]string{
			"user",
		},
		[]interface{}{
			user,
		})
}

func (impl *UserDaoImpl) DisableUser(ctx context.Context, id int64) error {
	_, err := impl.session.Update(ctx, "UserDao.DisableUser",
		[]string{
			"id",
		},
		[]interface{}{
			id,
		})
	return err
}

func (impl *UserDaoImpl) EnableUser(ctx context.Context, id int64) error {
	_, err := impl.session.Update(ctx, "UserDao.EnableUser",
		[]string{
			"id",
		},
		[]interface{}{
			id,
		})
	return err
}

func (impl *UserDaoImpl) UpdateUser(ctx context.Context, id int64, user *User) (int64, error) {
	return impl.session.Update(ctx, "UserDao.UpdateUser",
		[]string{
			"id",
			"user",
		},
		[]interface{}{
			id,
			user,
		})
}

func (impl *UserDaoImpl) AddRoleToUser(ctx context.Context, userid int64, roleid int64) error {
	_, err := impl.session.Insert(ctx, "UserDao.AddRoleToUser",
		[]string{
			"userid",
			"roleid",
		},
		[]interface{}{
			userid,
			roleid,
		},
		true)
	return err
}

func (impl *UserDaoImpl) RemoveRoleFromUser(ctx context.Context, userid int64, roleid int64) error {
	_, err := impl.session.Delete(ctx, "UserDao.RemoveRoleFromUser",
		[]string{
			"userid",
			"roleid",
		},
		[]interface{}{
			userid,
			roleid,
		})
	return err
}

func init() {
	gobatis.Init(func(ctx *gobatis.InitContext) error {
		{ //// UserQueryer.GetRoleByName
			if _, exists := ctx.Statements["UserQueryer.GetRoleByName"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&Role{}),
					[]string{
						"name",
					},
					[]reflect.Type{
						reflect.TypeOf(new(string)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetRoleByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetRoleByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetRoleByName"] = stmt
			}
		}
		{ //// UserQueryer.GetUserByID
			if _, exists := ctx.Statements["UserQueryer.GetUserByID"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{
						"id",
					},
					[]reflect.Type{
						reflect.TypeOf(new(int64)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUserByID error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUserByID",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUserByID"] = stmt
			}
		}
		{ //// UserQueryer.GetUserByName
			if _, exists := ctx.Statements["UserQueryer.GetUserByName"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{
						"name",
					},
					[]reflect.Type{
						reflect.TypeOf(new(string)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUserByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUserByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUserByName"] = stmt
			}
		}
		{ //// UserQueryer.GetUsergroupByID
			if _, exists := ctx.Statements["UserQueryer.GetUsergroupByID"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&UserGroup{}),
					[]string{
						"id",
					},
					[]reflect.Type{
						reflect.TypeOf(new(int64)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUsergroupByID error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUsergroupByID",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUsergroupByID"] = stmt
			}
		}
		{ //// UserQueryer.GetUsergroupByName
			if _, exists := ctx.Statements["UserQueryer.GetUsergroupByName"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&UserGroup{}),
					[]string{
						"name",
					},
					[]reflect.Type{
						reflect.TypeOf(new(string)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUsergroupByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUsergroupByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUsergroupByName"] = stmt
			}
		}
		{ //// UserQueryer.GetUsers
			if _, exists := ctx.Statements["UserQueryer.GetUsers"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUsers error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUsers",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUsers"] = stmt
			}
		}
		{ //// UserQueryer.GetUsergroups
			if _, exists := ctx.Statements["UserQueryer.GetUsergroups"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&UserGroup{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetUsergroups error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUsergroups",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUsergroups"] = stmt
			}
		}
		{ //// UserQueryer.GetRolesByUser
			if _, exists := ctx.Statements["UserQueryer.GetRolesByUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("SELECT * FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&Role{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" AS ")
				sb.WriteString("roles")
				sb.WriteString(" WHERE\r\n  exists (select * from ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserAndRole{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" as users_roles\r\n     where users_roles.role_id = roles.id and users_roles.user_id = #{userID})")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetRolesByUser",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetRolesByUser"] = stmt
			}
		}
		{ //// UserQueryer.GetUserByGroup
			if _, exists := ctx.Statements["UserQueryer.GetUserByGroup"]; !exists {
				var sb strings.Builder
				sb.WriteString("SELECT * FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&User{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" AS ")
				sb.WriteString("users")
				sb.WriteString(" WHERE\r\n  exists (select * from ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserAndUserGroup{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" as u2g\r\n     where u2g.user_id = users.id and u2g.group_id = #{groupID})")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetUserByGroup",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetUserByGroup"] = stmt
			}
		}
		{ //// UserQueryer.GetGroupIDsByUser
			if _, exists := ctx.Statements["UserQueryer.GetGroupIDsByUser"]; !exists {
				var sb strings.Builder
				sb.WriteString("SELECT group_id FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserAndUserGroup{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" AS ")
				sb.WriteString("u2g")
				sb.WriteString(" WHERE user_id = #{userID}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetGroupIDsByUser",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetGroupIDsByUser"] = stmt
			}
		}
		{ //// UserQueryer.GetPermissionAndRoles
			if _, exists := ctx.Statements["UserQueryer.GetPermissionAndRoles"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&PermissionGroupAndRole{}),
					[]string{
						"roleIDs",
					},
					[]reflect.Type{
						reflect.TypeOf([]int64{}),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetPermissionAndRoles error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetPermissionAndRoles",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetPermissionAndRoles"] = stmt
			}
		}
		{ //// UserQueryer.ReadProfile
			if _, exists := ctx.Statements["UserQueryer.ReadProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("SELECT value FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" WHERE id = #{userID} AND name = #{name}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.ReadProfile",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.ReadProfile"] = stmt
			}
		}
		{ //// UserQueryer.WriteProfile
			if _, exists := ctx.Statements["UserQueryer.WriteProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("INSERT INTO ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" (id, name, value) VALUES(#{userID}, #{name}, #{value})\r\n     ON CONFLICT (id, name) DO UPDATE SET value = excluded.value")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.WriteProfile",
					gobatis.StatementTypeUpdate,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.WriteProfile"] = stmt
			}
		}
		{ //// UserQueryer.DeleteProfile
			if _, exists := ctx.Statements["UserQueryer.DeleteProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("DELETE FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" WHERE id=#{userID} AND name=#{name}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.DeleteProfile",
					gobatis.StatementTypeDelete,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.DeleteProfile"] = stmt
			}
		}
		{ //// UserQueryer.GetPermissions
			if _, exists := ctx.Statements["UserQueryer.GetPermissions"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&Permissions{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetPermissions error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetPermissions",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetPermissions"] = stmt
			}
		}
		{ //// UserQueryer.GetPermissionAndGroups
			if _, exists := ctx.Statements["UserQueryer.GetPermissionAndGroups"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&PermissionAndGroup{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserQueryer.GetPermissionAndGroups error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserQueryer.GetPermissionAndGroups",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserQueryer.GetPermissionAndGroups"] = stmt
			}
		}
		return nil
	})
}

func NewUserQueryer(ref gobatis.SqlSession) UserQueryer {
	if ref == nil {
		panic(errors.New("param 'ref' is nil"))
	}
	if reference, ok := ref.(*gobatis.Reference); ok {
		if reference.SqlSession == nil {
			panic(errors.New("param 'ref.SqlSession' is nil"))
		}
	} else if valueReference, ok := ref.(gobatis.Reference); ok {
		if valueReference.SqlSession == nil {
			panic(errors.New("param 'ref.SqlSession' is nil"))
		}
	}
	return &UserQueryerImpl{session: ref}
}

type UserQueryerImpl struct {
	session gobatis.SqlSession
}

func (impl *UserQueryerImpl) GetRoleByName(ctx context.Context, name string) func(*Role) error {
	result := impl.session.SelectOne(ctx, "UserQueryer.GetRoleByName",
		[]string{
			"name",
		},
		[]interface{}{
			name,
		})
	return func(value *Role) error {
		return result.Scan(value)
	}
}

func (impl *UserQueryerImpl) GetUserByID(ctx context.Context, id int64) func(*User) error {
	result := impl.session.SelectOne(ctx, "UserQueryer.GetUserByID",
		[]string{
			"id",
		},
		[]interface{}{
			id,
		})
	return func(value *User) error {
		return result.Scan(value)
	}
}

func (impl *UserQueryerImpl) GetUserByName(ctx context.Context, name string) func(*User) error {
	result := impl.session.SelectOne(ctx, "UserQueryer.GetUserByName",
		[]string{
			"name",
		},
		[]interface{}{
			name,
		})
	return func(value *User) error {
		return result.Scan(value)
	}
}

func (impl *UserQueryerImpl) GetUsergroupByID(ctx context.Context, id int64) func(*UserGroup) error {
	result := impl.session.SelectOne(ctx, "UserQueryer.GetUsergroupByID",
		[]string{
			"id",
		},
		[]interface{}{
			id,
		})
	return func(value *UserGroup) error {
		return result.Scan(value)
	}
}

func (impl *UserQueryerImpl) GetUsergroupByName(ctx context.Context, name string) func(*UserGroup) error {
	result := impl.session.SelectOne(ctx, "UserQueryer.GetUsergroupByName",
		[]string{
			"name",
		},
		[]interface{}{
			name,
		})
	return func(value *UserGroup) error {
		return result.Scan(value)
	}
}

func (impl *UserQueryerImpl) GetUsers(ctx context.Context) ([]User, error) {
	var instances []User
	results := impl.session.Select(ctx, "UserQueryer.GetUsers",
		[]string{},
		[]interface{}{})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetUsergroups(ctx context.Context) ([]UserGroup, error) {
	var instances []UserGroup
	results := impl.session.Select(ctx, "UserQueryer.GetUsergroups",
		[]string{},
		[]interface{}{})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetRolesByUser(ctx context.Context, userID int64) ([]Role, error) {
	var instances []Role
	results := impl.session.Select(ctx, "UserQueryer.GetRolesByUser",
		[]string{
			"userID",
		},
		[]interface{}{
			userID,
		})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetUserByGroup(ctx context.Context, groupID int64) ([]User, error) {
	var instances []User
	results := impl.session.Select(ctx, "UserQueryer.GetUserByGroup",
		[]string{
			"groupID",
		},
		[]interface{}{
			groupID,
		})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetGroupIDsByUser(ctx context.Context, userID int64) ([]int64, error) {
	var instances []int64
	results := impl.session.Select(ctx, "UserQueryer.GetGroupIDsByUser",
		[]string{
			"userID",
		},
		[]interface{}{
			userID,
		})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetPermissionAndRoles(ctx context.Context, roleIDs []int64) ([]PermissionGroupAndRole, error) {
	var instances []PermissionGroupAndRole
	results := impl.session.Select(ctx, "UserQueryer.GetPermissionAndRoles",
		[]string{
			"roleIDs",
		},
		[]interface{}{
			roleIDs,
		})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) ReadProfile(ctx context.Context, userID int64, name string) (string, error) {
	var instance string
	var nullable gobatis.Nullable
	nullable.Value = &instance

	err := impl.session.SelectOne(ctx, "UserQueryer.ReadProfile",
		[]string{
			"userID",
			"name",
		},
		[]interface{}{
			userID,
			name,
		}).Scan(&nullable)
	if err != nil {
		return "", err
	}
	if !nullable.Valid {
		return "", sql.ErrNoRows
	}

	return instance, nil
}

func (impl *UserQueryerImpl) WriteProfile(ctx context.Context, userID int64, name string, value string) error {
	_, err := impl.session.Update(ctx, "UserQueryer.WriteProfile",
		[]string{
			"userID",
			"name",
			"value",
		},
		[]interface{}{
			userID,
			name,
			value,
		})
	return err
}

func (impl *UserQueryerImpl) DeleteProfile(ctx context.Context, userID int64, name string) (int64, error) {
	return impl.session.Delete(ctx, "UserQueryer.DeleteProfile",
		[]string{
			"userID",
			"name",
		},
		[]interface{}{
			userID,
			name,
		})
}

func (impl *UserQueryerImpl) GetPermissions(ctx context.Context) ([]Permissions, error) {
	var instances []Permissions
	results := impl.session.Select(ctx, "UserQueryer.GetPermissions",
		[]string{},
		[]interface{}{})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetPermissionAndGroups(ctx context.Context) ([]PermissionAndGroup, error) {
	var instances []PermissionAndGroup
	results := impl.session.Select(ctx, "UserQueryer.GetPermissionAndGroups",
		[]string{},
		[]interface{}{})
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}
