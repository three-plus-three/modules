package menus

import (
	"database/sql"
	"strings"
	"time"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/util"
)

func SortBy(list []toolbox.Menu, names []string) []toolbox.Menu {
	if len(list) == 0 {
		return list
	}

	offset := 0
	for _, name := range names {
		foundIdx := -1
		for idx := range list {
			if list[idx].UID == name {
				foundIdx = idx
			}
		}
		if foundIdx < 0 {
			continue
		}

		if foundIdx != offset {
			tmp := list[offset]
			list[offset] = list[foundIdx]
			list[foundIdx] = tmp
		}
		offset++
	}
	return list
}

func ReadApplications(env *environment.Environment, db *sql.DB) ([]toolbox.Menu, error) {
	list, err := ReadApplicationsFromDB(db)
	if err != nil {
		return nil, err
	}

	names := env.Config.StringWithDefault("applications.names", "")
	if names == "" {
		return list, nil
	}
	return SortBy(list, strings.Split(names, ",")), nil
}

func ReadApplicationsFromDB(db *sql.DB) ([]toolbox.Menu, error) {
	var id int64
	var address sql.NullString
	var name string
	var url sql.NullString
	var icon sql.NullString
	var title string
	var classes sql.NullString

	rows, err := db.Query("select id, address, name, url, icon, title, classes from tpt_products")
	if err != nil {
		return nil, err
	}
	defer util.CloseWith(rows)

	var menuList []toolbox.Menu
	for rows.Next() {
		err = rows.Scan(&id, &address, &name, &url, &icon, &title, &classes)
		if err != nil {
			return nil, err
		}

		if !url.Valid {
			continue
		}

		menuList = append(menuList, toolbox.Menu{
			UID:   "product-" + name,
			Title: title,
			// Permission: "product." + name,
			// License    string `json:"license,omitempty" xorm:"license"`
			URL:     url.String,
			Icon:    icon.String,
			Classes: classes.String + " special_link",
		})
	}
	return menuList, rows.Err()
}

// ApplicationsWrap 增加从数据库中读菜单的功能
func ApplicationsWrap(env *environment.Environment, db *sql.DB, cb Callback) Callback {
	var cachedValue CachedValue
	cachedValue.MaxAge = 5 * 60
	return func() ([]toolbox.Menu, error) {
		value := cachedValue.Get()
		if value == nil {
			var err error
			value, err = ReadApplications(env, db)
			if err != nil {
				return nil, err
			}
			cachedValue.Set(value, time.Now())
		}

		if len(value) == 0 {
			return cb()
		}

		value2, err := cb()
		if err != nil {
			return nil, err
		}

		return append(value, value2...), nil
	}
}
