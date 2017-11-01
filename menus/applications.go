package menus

import (
	"database/sql"
	"strings"
	"time"

	"github.com/three-plus-three/modules/environment"
	"github.com/three-plus-three/modules/errors"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/urlutil"
	"github.com/three-plus-three/modules/util"
)

var IgnoreListOfProducts = []string{"mc"}

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

func ReadProducts(env *environment.Environment, db *sql.DB, ignoreList []string) ([]toolbox.Menu, error) {
	list, err := ReadProductsFromDB(db, ignoreList)
	if err != nil {
		return nil, err
	}

	names := env.Config.StringWithDefault("applications.names", "")
	if names == "" {
		return list, nil
	}
	return SortBy(list, strings.Split(names, ",")), nil
}

func ReadProductsFromDB(db *sql.DB, ignoreList []string) ([]toolbox.Menu, error) {
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

		found := false
		for _, list := range [][]string{ignoreList, IgnoreListOfProducts} {
			for _, nm := range list {
				if nm == name {
					found = true
					break
				}
			}

			if found {
				break
			}
		}

		if found {
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

// ProductsWrap 增加从数据库中读菜单的功能
func ProductsWrap(env *environment.Environment, applicationID environment.ENV_PROXY_TYPE, db *sql.DB, cb Callback) Callback {
	var cachedValue CachedValue
	cachedValue.MaxAge = 5 * 60
	ignoreList := []string{env.GetServiceConfig(applicationID).Name}
	return func() ([]toolbox.Menu, error) {
		value := cachedValue.Get()
		if value == nil {
			var err error
			value, err = ReadProducts(env, db, ignoreList)
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

func UpdateProduct(env *environment.Environment,
	applicationID environment.ENV_PROXY_TYPE,
	version, title, icon, classes string, db *sql.DB) error {

	so := env.GetServiceConfig(applicationID)
	url := urlutil.Join(env.DaemonUrlPath, so.Name)
	if applicationID == environment.ENV_WSERVER_PROXY_ID {
		url = env.DaemonUrlPath
	}

	var count int64
	err := db.QueryRow("select count(*) from tpt_products where name = $1", so.Name).Scan(&count)
	if err != nil {
		return errors.Wrap(err, "UpdateProduct")
	}

	now := time.Now()
	if count == 0 {
		_, err = db.Exec("INSERT INTO tpt_products (name, version, url, icon, title, classes, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
			so.Name, version, url, icon, title, classes, now, now)
	} else {
		_, err = db.Exec("UPDATE tpt_products SET version=$1, url=$2, icon=$3, title=$4, classes=$5, updated_at=$6 WHERE name=$7",
			version, url, icon, title, classes, now, so.Name)
	}
	if err != nil {
		return errors.Wrap(err, "UpdateProduct")
	}
	return nil
}
