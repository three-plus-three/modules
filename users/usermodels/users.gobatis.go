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
		{ //// UserDao.GetUserByGroup
			if _, exists := ctx.Statements["UserDao.GetUserByGroup"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{
						"groupID",
					},
					[]reflect.Type{
						reflect.TypeOf(new(int64)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUserByGroup error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUserByGroup",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUserByGroup"] = stmt
			}
		}
		{ //// UserDao.GetGroupIDsByUser
			if _, exists := ctx.Statements["UserDao.GetGroupIDsByUser"]; !exists {
				return errors.New("sql 'UserDao.GetGroupIDsByUser' error : statement not found - Generate SQL fail: recordType is unknown")
			}
		}
		{ //// UserDao.GetUsers
			if _, exists := ctx.Statements["UserDao.GetUsers"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&User{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUsers error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUsers",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUsers"] = stmt
			}
		}
		{ //// UserDao.ReadProfile
			if _, exists := ctx.Statements["UserDao.ReadProfile"]; !exists {
				return errors.New("sql 'UserDao.ReadProfile' error : statement not found - Generate SQL fail: recordType is unknown")
			}
		}
		{ //// UserDao.DeleteProfile
			if _, exists := ctx.Statements["UserDao.DeleteProfile"]; !exists {
				return errors.New("sql 'UserDao.DeleteProfile' error : statement not found - Generate SQL fail: recordType is unknown")
			}
		}
		{ //// UserDao.GetUserByID
			if _, exists := ctx.Statements["UserDao.GetUserByID"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUserByID error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUserByID",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUserByID"] = stmt
			}
		}
		{ //// UserDao.GetPermissionAndGroups
			if _, exists := ctx.Statements["UserDao.GetPermissionAndGroups"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&PermissionAndGroup{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetPermissionAndGroups error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetPermissionAndGroups",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetPermissionAndGroups"] = stmt
			}
		}
		{ //// UserDao.WriteProfile
			if _, exists := ctx.Statements["UserDao.WriteProfile"]; !exists {
				return errors.New("sql 'UserDao.WriteProfile' error : statement not found - Generate SQL fail: recordType is unknown")
			}
		}
		{ //// UserDao.GetRolesByUser
			if _, exists := ctx.Statements["UserDao.GetRolesByUser"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&Role{}),
					[]string{
						"userID",
					},
					[]reflect.Type{
						reflect.TypeOf(new(int64)).Elem(),
					},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetRolesByUser error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetRolesByUser",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetRolesByUser"] = stmt
			}
		}
		{ //// UserDao.GetRoleByName
			if _, exists := ctx.Statements["UserDao.GetRoleByName"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetRoleByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetRoleByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetRoleByName"] = stmt
			}
		}
		{ //// UserDao.GetUsergroupByName
			if _, exists := ctx.Statements["UserDao.GetUsergroupByName"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUsergroupByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUsergroupByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUsergroupByName"] = stmt
			}
		}
		{ //// UserDao.GetUserByName
			if _, exists := ctx.Statements["UserDao.GetUserByName"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUserByName error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUserByName",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUserByName"] = stmt
			}
		}
		{ //// UserDao.GetUsergroupByID
			if _, exists := ctx.Statements["UserDao.GetUsergroupByID"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUsergroupByID error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUsergroupByID",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUsergroupByID"] = stmt
			}
		}
		{ //// UserDao.GetPermissionAndRoles
			if _, exists := ctx.Statements["UserDao.GetPermissionAndRoles"]; !exists {
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
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetPermissionAndRoles error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetPermissionAndRoles",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetPermissionAndRoles"] = stmt
			}
		}
		{ //// UserDao.GetUsergroups
			if _, exists := ctx.Statements["UserDao.GetUsergroups"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&UserGroup{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetUsergroups error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetUsergroups",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetUsergroups"] = stmt
			}
		}
		{ //// UserDao.GetPermissions
			if _, exists := ctx.Statements["UserDao.GetPermissions"]; !exists {
				sqlStr, err := gobatis.GenerateSelectSQL(ctx.Dialect, ctx.Mapper,
					reflect.TypeOf(&Permissions{}),
					[]string{},
					[]reflect.Type{},
					[]gobatis.Filter{})
				if err != nil {
					return gobatis.ErrForGenerateStmt(err, "generate UserDao.GetPermissions error")
				}
				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetPermissions",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetPermissions"] = stmt
			}
		}
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

func NewUserDao(ref gobatis.SqlSession) UserDao {
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
	return &UserDaoImpl{session: ref}
}

type UserDaoImpl struct {
	session gobatis.SqlSession
}

func (impl *UserDaoImpl) GetUserByGroup(groupID int64) ([]User, error) {
	var instances []User
	results := impl.session.Select(context.Background(), "UserDao.GetUserByGroup",
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

func (impl *UserDaoImpl) GetGroupIDsByUser(userID int64) ([]int64, error) {
	var instances []int64
	results := impl.session.Select(context.Background(), "UserDao.GetGroupIDsByUser",
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

func (impl *UserDaoImpl) GetUsers() ([]User, error) {
	var instances []User
	results := impl.session.Select(context.Background(), "UserDao.GetUsers", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserDaoImpl) ReadProfile(userID int64, name string) (string, error) {
	var instance string
	var nullable gobatis.Nullable
	nullable.Value = &instance

	err := impl.session.SelectOne(context.Background(), "UserDao.ReadProfile",
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

func (impl *UserDaoImpl) DeleteProfile(userID int64, name string) (int64, error) {
	return impl.session.Delete(context.Background(), "UserDao.DeleteProfile",
		[]string{
			"userID",
			"name",
		},
		[]interface{}{
			userID,
			name,
		})
}

func (impl *UserDaoImpl) GetUserByID(id int64) func(*User) error {
	result := impl.session.SelectOne(context.Background(), "UserDao.GetUserByID",
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

func (impl *UserDaoImpl) GetPermissionAndGroups() ([]PermissionAndGroup, error) {
	var instances []PermissionAndGroup
	results := impl.session.Select(context.Background(), "UserDao.GetPermissionAndGroups", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserDaoImpl) WriteProfile(userID int64, name string, value string) error {
	_, err := impl.session.Update(context.Background(), "UserDao.WriteProfile",
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

func (impl *UserDaoImpl) GetRolesByUser(userID int64) ([]Role, error) {
	var instances []Role
	results := impl.session.Select(context.Background(), "UserDao.GetRolesByUser",
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

func (impl *UserDaoImpl) GetRoleByName(name string) func(*Role) error {
	result := impl.session.SelectOne(context.Background(), "UserDao.GetRoleByName",
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

func (impl *UserDaoImpl) GetUsergroupByName(name string) func(*UserGroup) error {
	result := impl.session.SelectOne(context.Background(), "UserDao.GetUsergroupByName",
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

func (impl *UserDaoImpl) GetUserByName(name string) func(*User) error {
	result := impl.session.SelectOne(context.Background(), "UserDao.GetUserByName",
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

func (impl *UserDaoImpl) GetUsergroupByID(id int64) func(*UserGroup) error {
	result := impl.session.SelectOne(context.Background(), "UserDao.GetUsergroupByID",
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

func (impl *UserDaoImpl) GetPermissionAndRoles(roleIDs []int64) ([]PermissionGroupAndRole, error) {
	var instances []PermissionGroupAndRole
	results := impl.session.Select(context.Background(), "UserDao.GetPermissionAndRoles",
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

func (impl *UserDaoImpl) GetUsergroups() ([]UserGroup, error) {
	var instances []UserGroup
	results := impl.session.Select(context.Background(), "UserDao.GetUsergroups", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserDaoImpl) GetPermissions() ([]Permissions, error) {
	var instances []Permissions
	results := impl.session.Select(context.Background(), "UserDao.GetPermissions", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
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

func (impl *UserQueryerImpl) GetRoleByName(name string) func(*Role) error {
	result := impl.session.SelectOne(context.Background(), "UserQueryer.GetRoleByName",
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

func (impl *UserQueryerImpl) GetUserByID(id int64) func(*User) error {
	result := impl.session.SelectOne(context.Background(), "UserQueryer.GetUserByID",
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

func (impl *UserQueryerImpl) GetUserByName(name string) func(*User) error {
	result := impl.session.SelectOne(context.Background(), "UserQueryer.GetUserByName",
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

func (impl *UserQueryerImpl) GetUsergroupByID(id int64) func(*UserGroup) error {
	result := impl.session.SelectOne(context.Background(), "UserQueryer.GetUsergroupByID",
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

func (impl *UserQueryerImpl) GetUsergroupByName(name string) func(*UserGroup) error {
	result := impl.session.SelectOne(context.Background(), "UserQueryer.GetUsergroupByName",
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

func (impl *UserQueryerImpl) GetUsers() ([]User, error) {
	var instances []User
	results := impl.session.Select(context.Background(), "UserQueryer.GetUsers", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetUsergroups() ([]UserGroup, error) {
	var instances []UserGroup
	results := impl.session.Select(context.Background(), "UserQueryer.GetUsergroups", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetRolesByUser(userID int64) ([]Role, error) {
	var instances []Role
	results := impl.session.Select(context.Background(), "UserQueryer.GetRolesByUser",
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

func (impl *UserQueryerImpl) GetUserByGroup(groupID int64) ([]User, error) {
	var instances []User
	results := impl.session.Select(context.Background(), "UserQueryer.GetUserByGroup",
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

func (impl *UserQueryerImpl) GetGroupIDsByUser(userID int64) ([]int64, error) {
	var instances []int64
	results := impl.session.Select(context.Background(), "UserQueryer.GetGroupIDsByUser",
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

func (impl *UserQueryerImpl) GetPermissionAndRoles(roleIDs []int64) ([]PermissionGroupAndRole, error) {
	var instances []PermissionGroupAndRole
	results := impl.session.Select(context.Background(), "UserQueryer.GetPermissionAndRoles",
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

func (impl *UserQueryerImpl) ReadProfile(userID int64, name string) (string, error) {
	var instance string
	var nullable gobatis.Nullable
	nullable.Value = &instance

	err := impl.session.SelectOne(context.Background(), "UserQueryer.ReadProfile",
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

func (impl *UserQueryerImpl) WriteProfile(userID int64, name string, value string) error {
	_, err := impl.session.Update(context.Background(), "UserQueryer.WriteProfile",
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

func (impl *UserQueryerImpl) DeleteProfile(userID int64, name string) (int64, error) {
	return impl.session.Delete(context.Background(), "UserQueryer.DeleteProfile",
		[]string{
			"userID",
			"name",
		},
		[]interface{}{
			userID,
			name,
		})
}

func (impl *UserQueryerImpl) GetPermissions() ([]Permissions, error) {
	var instances []Permissions
	results := impl.session.Select(context.Background(), "UserQueryer.GetPermissions", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}

func (impl *UserQueryerImpl) GetPermissionAndGroups() ([]PermissionAndGroup, error) {
	var instances []PermissionAndGroup
	results := impl.session.Select(context.Background(), "UserQueryer.GetPermissionAndGroups", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
}
