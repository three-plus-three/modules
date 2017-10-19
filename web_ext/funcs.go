package web_ext

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strings"

	"github.com/revel/revel"
	"github.com/three-plus-three/forms"
	"github.com/three-plus-three/modules/functions"
	"github.com/three-plus-three/modules/urlutil"
)

func ResourcesURLFor(s ...string) string {
	return urlutil.JoinWith(urlutil.Join(lifecycleData.URLPrefix, "/internal/custom_resources/"), s)
}

func initTemplateFuncs(lifecycle *Lifecycle) {
	revel.TemplateFuncs["assets"] = func(value string) string {
		return urlutil.Join(lifecycle.URLPrefix, "web/assets", value)
	}

	revel.TemplateFuncs["custom_resources_assets"] = func(value string) string {
		return ResourcesURLFor(value)
	}
	revel.TemplateFuncs["mc_assets"] = func(url string) string {
		return urlutil.Join(lifecycle.URLRoot, "web/assets", url)
	}
	revel.TemplateFuncs["tpt_assets"] = func(value string) string {
		return urlutil.Join(lifecycle.URLPrefix, "web/tpt_assets", value)
	}
	revel.TemplateFuncs["default"] = func(value, defvalue interface{}) interface{} {
		if nil == value {
			return defvalue
		}
		if s, ok := value.(string); ok && "" == s {
			return defvalue
		}
		return value
	}

	revel.TemplateFuncs["args"] = func() map[string]interface{} {
		return map[string]interface{}{}
	}

	revel.TemplateFuncs["arg"] = func(n string, v interface{}, args map[string]interface{}) map[string]interface{} {
		args[n] = v

		return args
	}

	revel.TemplateFuncs["list"] = func() []interface{} {
		return []interface{}{}
	}

	revel.TemplateFuncs["startsWith"] = func(s, sep string) bool {
		return strings.Index(s, sep) == 0
	}

	revel.TemplateFuncs["tabItem"] = func(id, label string, active bool, items []interface{}) []interface{} {
		return append(items, map[string]interface{}{"id": id, "label": label, "active": active})
	}

	revel.TemplateFuncs["menuItem"] = func(id string, class string, label string, items []interface{}) []interface{} {
		return append(items, map[string]interface{}{id: id, "class": class, "label": label})
	}

	revel.TemplateFuncs["urlPrefix"] = func(s ...string) string {
		return urlutil.JoinWith(lifecycle.ApplicationContext, s)
	}

	revel.TemplateFuncs["appRoot"] = func(s ...string) string {
		return urlutil.JoinWith(lifecycle.ApplicationRoot, s)
	}

	revel.TemplateFuncs["urlRoot"] = func(s ...string) string {
		return urlutil.JoinWith(lifecycle.URLRoot, s)
	}

	revel.TemplateFuncs["urlParam"] = func(key, value string, urlStr interface{}) string {
		var u *url.URL
		switch v := urlStr.(type) {
		case string:
			var err error
			u, err = url.Parse(v)
			if err != nil {
				panic(errors.New("url '" + v + "' is invalid url: " + err.Error()))
			}
		case url.URL:
			u = &v
		case *url.URL:
			u = v
		case template.URL:
			var err error
			u, err = url.Parse(string(v))
			if err != nil {
				panic(errors.New("url '" + string(v) + "' is invalid url: " + err.Error()))
			}
		default:
			panic(fmt.Errorf("url '[%T] %s' is invalid url", urlStr, urlStr))
		}

		query := u.Query()
		query.Add(key, value)
		u.RawQuery = query.Encode()
		return u.String()
	}

	funcs := functions.HtmlFuncMap()
	for k, v := range funcs {
		if _, ok := revel.TemplateFuncs[k]; !ok {
			revel.TemplateFuncs[k] = v
		}
	}
	for _, alias := range [][2]string{{"sum", "add"},
		{"tostring", "toString"}} {
		if _, ok := revel.TemplateFuncs[alias[0]]; !ok {
			revel.TemplateFuncs[alias[0]] = funcs[alias[1]]
		}
	}

	forms.Init(revel.DevMode, revel.SourcePath, revel.TemplateFuncs)
	for k, v := range forms.FieldFuncs {
		revel.TemplateFuncs[k] = v
	}

	revel.TemplateFuncs["current_user_has_permission"] = func(ctx map[string]interface{}, permissionName string, op ...string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, op)
	}
	revel.TemplateFuncs["current_user_has_new_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{CREATE})
	}
	revel.TemplateFuncs["current_user_has_del_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{DELETE})
	}
	revel.TemplateFuncs["current_user_has_edit_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{UPDATE})
	}
	revel.TemplateFuncs["current_user_has_write_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{CREATE, DELETE, UPDATE})
	}
	revel.TemplateFuncs["current_user_has_query_permission"] = func(ctx map[string]interface{}, permissionName string) bool {
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{QUERY})
	}
	revel.TemplateFuncs["current_user_has_menu"] = func(ctx map[string]interface{}, permissionName string) bool {
		if permissionName == "" {
			return true
		}
		return CurrentUserHasPermission(lifecycle, ctx, permissionName, []string{QUERY})
	}
	revel.TemplateFuncs["user_has_permission"] = func(ctx map[string]interface{}, user, permissionName, op string) bool {
		u := lifecycle.GetUser(user)
		if u != nil {
			return false
		}
		return u.HasPermission(permissionName, op)
	}

	revel.TemplateFuncs["fieldClasses"] = func(isRequired bool, typ string) string {
		classes := []string{}

		if isRequired {
			classes = append(classes, "required")
		}

		if typ == "integer" {
			classes = append(classes, "digits")
		} else if typ == "decimal" {
			classes = append(classes, "number")
		} else if typ == "date" {
			classes = append(classes, "dateISO")
		} else if typ == "ipAddress" {
			classes = append(classes, "ipv4")
		} else if typ == "physicalAddress" {
			classes = append(classes, "macaddress")
		}

		return strings.Join(classes, " ")
	}
}

func CurrentUserHasPermission(lifecycle *Lifecycle, ctx map[string]interface{}, permissionName string, opList []string) bool {
	o := ctx["currentUser"]
	if o == nil {
		return false
	}

	u, ok := o.(User)
	if !ok {
		return false
	}
	for _, op := range opList {
		if u.HasPermission(permissionName, op) {
			return true
		}
	}
	return false
}
