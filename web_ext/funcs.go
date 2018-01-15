package web_ext

import (
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"strconv"
	"strings"

	"github.com/revel/revel"
	"github.com/runner-mei/orm"
	"github.com/three-plus-three/forms"
	"github.com/three-plus-three/modules/functions"
	"github.com/three-plus-three/modules/toolbox"
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

	revel.TemplateFuncs["list"] = func(args ...interface{}) []interface{} {
		return args
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

	revel.TemplateFuncs["urljoin"] = urlutil.Join

	revel.TemplateFuncs["js"] = func(s string) template.JS {
		return template.JS(s)
	}

	revel.TemplateFuncs["jsstr"] = func(s string) template.JSStr {
		return template.JSStr(s)
	}

	revel.TemplateFuncs["html"] = func(s string) template.HTML {
		return template.HTML(s)
	}

	revel.TemplateFuncs["htmlAttr"] = func(s string) template.HTMLAttr {
		return template.HTMLAttr(s)
	}

	revel.TemplateFuncs["urlParam"] = func(key string, value, urlObject interface{}) string {
		var u *url.URL
		switch v := urlObject.(type) {
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
			panic(fmt.Errorf("url '[%T] %s' is invalid url", urlObject, urlObject))
		}

		query := u.Query()
		query.Add(key, fmt.Sprint(value))
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

var localeMessages = map[string]string{
	"update.record_not_found":     "update.record_not_found",
	"unique_value_already_exists": "",
}

func ErrorToFlash(c *revel.Controller, err error, notFoundKey ...string) {
	if err == orm.ErrNotFound {
		if len(notFoundKey) >= 1 && notFoundKey[0] != "" {
			c.Flash.Error(revel.Message(c.Request.Locale, notFoundKey[0]))
		} else {
			c.Flash.Error(revel.Message(c.Request.Locale, "update.record_not_found"))
		}
	} else {
		if oerr, ok := err.(*orm.Error); ok && len(oerr.Validations) > 0 {
			for _, validation := range oerr.Validations {
				localeMessage := validation.Message
				if key, found := localeMessages[validation.Code]; found {
					if key == "" {
						localeMessage = revel.Message(c.Request.Locale, validation.Code)
					} else {
						localeMessage = revel.Message(c.Request.Locale, key)
					}
				} else {
					localeMessage = revel.Message(c.Request.Locale, validation.Code)
				}

				c.Validation.Error(localeMessage).
					Key(validation.Key)
			}
			c.Validation.Keep()
		} else {
			c.Flash.Error(err.Error())
		}
	}
}

func NewPaginatorWith(c *revel.Controller, pageSize int, total interface{}) *toolbox.Paginator {
	form, _ := c.Request.GetForm()
	if form != nil {
		pageIndex, _ := strconv.Atoi(form.Get("pageIndex"))
		return toolbox.NewPaginatorWith(c.Request.URL, pageIndex, pageSize, total)
	}
	return toolbox.NewPaginatorWith(c.Request.URL, 0, pageSize, total)
}

//var (
//	controllerPtrType = reflect.TypeOf(&revel.Controller{})
//)

// Find the value of the target, starting from val and including embedded types.
// Also, convert between any difference in indirection.
// If the target couldn't be found, the returned Value will have IsValid() == false
// func findTarget(val reflect.Value, target reflect.Type) reflect.Value {
// 	// Look through the embedded types (until we reach the *revel.Controller at the top).
// 	valueQueue := []reflect.Value{val}
// 	for len(valueQueue) > 0 {
// 		val, valueQueue = valueQueue[0], valueQueue[1:]
//
// 		// Check if val is of a similar type to the target type.
// 		if val.Type() == target {
// 			return val
// 		}
// 		if val.Kind() == reflect.Ptr && val.Elem().Type() == target {
// 			return val.Elem()
// 		}
// 		if target.Kind() == reflect.Ptr && target.Elem() == val.Type() {
// 			return val.Addr()
// 		}
//
// 		// Else, add each anonymous field to the queue.
// 		if val.Kind() == reflect.Ptr {
// 			val = val.Elem()
// 		}
//
// 		for i := 0; i < val.NumField(); i++ {
// 			if val.Type().Field(i).Anonymous {
// 				valueQueue = append(valueQueue, val.Field(i))
// 			}
// 		}
// 	}
//
// 	return reflect.Value{}
// }
