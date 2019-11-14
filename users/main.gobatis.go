// Please don't edit this file!
package users

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
		{ //// UserDao.GetRolesByUser
			if _, exists := ctx.Statements["UserDao.GetRolesByUser"]; !exists {
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
		{ //// UserDao.GetUserByGroup
			if _, exists := ctx.Statements["UserDao.GetUserByGroup"]; !exists {
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

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.GetGroupIDsByUser",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.GetGroupIDsByUser"] = stmt
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
		{ //// UserDao.ReadProfile
			if _, exists := ctx.Statements["UserDao.ReadProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("SELECT value FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" WHERE id = #{userID} AND name = #{name}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.ReadProfile",
					gobatis.StatementTypeSelect,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.ReadProfile"] = stmt
			}
		}
		{ //// UserDao.WriteProfile
			if _, exists := ctx.Statements["UserDao.WriteProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("INSERT INTO ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" (id, name, value) VALUES(#{userID}, #{name}, #{value})\r\n     ON CONFLICT (id, name) DO UPDATE SET value = excluded.value")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.WriteProfile",
					gobatis.StatementTypeInsert,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.WriteProfile"] = stmt
			}
		}
		{ //// UserDao.DeleteProfile
			if _, exists := ctx.Statements["UserDao.DeleteProfile"]; !exists {
				var sb strings.Builder
				sb.WriteString("DELETE FROM ")
				if tablename, err := gobatis.ReadTableName(ctx.Mapper, reflect.TypeOf(&UserProfile{})); err != nil {
					return err
				} else {
					sb.WriteString(tablename)
				}
				sb.WriteString(" WHERE id=#{userID} AND name=#{name}")
				sqlStr := sb.String()

				stmt, err := gobatis.NewMapppedStatement(ctx, "UserDao.DeleteProfile",
					gobatis.StatementTypeDelete,
					gobatis.ResultStruct,
					sqlStr)
				if err != nil {
					return err
				}
				ctx.Statements["UserDao.DeleteProfile"] = stmt
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

func (impl *UserDaoImpl) GetUsers() ([]User, error) {
	var instances []User
	results := impl.session.Select(context.Background(), "UserDao.GetUsers", nil, nil)
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

func (impl *UserDaoImpl) WriteProfile(userID int64, name string, value string) error {
	_, err := impl.session.Insert(context.Background(), "UserDao.WriteProfile",
		[]string{
			"userID",
			"name",
			"value",
		},
		[]interface{}{
			userID,
			name,
			value,
		},
		true)
	return err
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

func (impl *UserDaoImpl) GetPermissions() ([]Permissions, error) {
	var instances []Permissions
	results := impl.session.Select(context.Background(), "UserDao.GetPermissions", nil, nil)
	err := results.ScanSlice(&instances)
	if err != nil {
		return nil, err
	}
	return instances, nil
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
