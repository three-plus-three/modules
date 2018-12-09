package web_ext

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/revel/revel"
	gobatis "github.com/runner-mei/GoBatis"
	"github.com/runner-mei/orm"
	"github.com/three-plus-three/forms"
	"github.com/three-plus-three/modules/functions"
	"github.com/three-plus-three/modules/toolbox"
	"github.com/three-plus-three/modules/types"
	"github.com/three-plus-three/modules/urlutil"
	"github.com/three-plus-three/modules/util"
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
	revel.TemplateFuncs["default"] = func(values ...interface{}) interface{} {
		for _, value := range values[:len(values)-1] {
			if value != nil && !util.IsZero(reflect.ValueOf(value)) {
				return value
			}
		}
		return values[len(values)-1]
	}

	revel.TemplateFuncs["toPtr"] = func(value interface{}) interface{} {
		if value == nil {
			return nil
		}
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Ptr {
			return value
		}
		if rv.CanAddr() {
			return rv.Addr().Interface()
		}
		return value
	}

	revel.TemplateFuncs["indexPtr"] = func(value interface{}, idx int) interface{} {
		if value == nil {
			return nil
		}
		rv := reflect.ValueOf(value)
		target := rv.Index(idx)

		if target.CanAddr() {
			return target.Addr().Interface()
		}
		return target.Interface()
	}

	revel.TemplateFuncs["args"] = func(args ...interface{}) map[string]interface{} {
		if len(args) > 0 {
			ctx, ok := args[0].(map[string]interface{})
			if ok {
				result := map[string]interface{}{}
				for key, value := range ctx {
					result[key] = value
				}
				return result
			}
		}
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

	revel.TemplateFuncs["urlParam"] = func(key string, value, urlObject interface{}) interface{} {
		if value == nil {
			return urlObject
		}
		if s, ok := value.(string); ok && s == "" {
			return urlObject
		}

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

	revel.TemplateFuncs["parseTpl"] = func(tpl string, data interface{}) string {
		t := template.New("main")
		t = t.Funcs(functions.HtmlFuncMap())
		t, err := t.Parse(tpl)

		if err != nil {
			return fmt.Sprintf("模板`%v`解析失败 => `%v`", tpl, err.Error())
		}

		var buf bytes.Buffer
		err = t.Execute(&buf, data)

		if err != nil {
			return fmt.Sprintf("模板`%v`执行失败 => `%v`", tpl, err.Error())
		}

		return buf.String()
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

	revel.TemplateFuncs["msg"] = func(viewArgs map[string]interface{}, message string, args ...interface{}) template.HTML {
		str, ok := viewArgs[revel.CurrentLocaleViewArg].(string)
		if !ok {
			str = revel.Config.StringDefault("i18n.default_language", "zh")
		}
		return template.HTML(revel.MessageFunc(str, message, args...))
	}

	toolbox.InitUserFuncs(lifecycle.UserManager, nil, revel.TemplateFuncs)

	revel.TemplateFuncs["newfield"] = NewField
}

func NewField(ctx interface{}, fieldSpec interface{}) forms.FieldInterface {
	field, ok := fieldSpec.(*types.FieldSpec)
	if !ok {
		fieldValue, ok := fieldSpec.(types.FieldSpec)
		if !ok {
			field = &fieldValue
		} else {
			rv := reflect.ValueOf(fieldSpec)
			if rv.Kind() == reflect.Ptr {
				rv = rv.Elem()
			}
			for i := 0; i < rv.NumField(); i++ {
				f := rv.Field(i)
				if !f.IsValid() {
					continue
				}
				if f.IsNil() {
					continue
				}

				field, ok = f.Interface().(*types.FieldSpec)
				if ok {
					break
				}
				fieldValue, ok := f.Interface().(types.FieldSpec)
				if !ok {
					field = &fieldValue
					break
				}
			}

			if field == nil {
				panic(errors.New("field is unknown type - " + rv.Type().PkgPath() + "." + rv.Type().Name()))
			}
		}
	}

	fieldLabel := field.Label + ":"
	if field.HasChoices() {
		widget := forms.SelectField(ctx, field.Name, fieldLabel, field.ToChoices())
		if field.IsArray {
			return widget.MultipleChoice()
		}
		return widget
	}
	format := field.Type
	if field.Format != "" {
		format = field.Format
	}

	if field.Annotations != nil {
		if inputType := field.Annotations["input_type"]; inputType != nil {
			if s, ok := inputType.(string); ok && s != "" {
				format = s
			}
		}
	}

	switch strings.ToLower(format) {
	case "select":
		widget := forms.SelectField(ctx, field.Name, fieldLabel, field.ToChoices())
		if field.IsArray {
			return widget.MultipleChoice()
		}
		return widget
	case "ip", "ipaddress":
		return forms.IPAddressField(ctx, field.Name, fieldLabel)
	case "email":
		return forms.EmailField(ctx, field.Name, fieldLabel)
	case "area":
		return forms.TextAreaField(ctx, field.Name, fieldLabel, 3, 0)
	case "string", "text":
		if field.Restrictions != nil {
			var isTextArea = false

			for _, value := range [][2]string{
				{"Length", field.Restrictions.Length},
				{"MaxLength", field.Restrictions.MaxLength},
			} {
				if value[1] != "" {
					length, err := strconv.ParseInt(value[1], 10, 64)
					if err != nil {
						panic(errors.New(value[0] + " of field '" + field.Name + "' is invalid - " + value[1]))
					}
					if length > 500 {
						isTextArea = true
					}
				}
			}
			if isTextArea {
				return forms.TextAreaField(ctx, field.Name, fieldLabel, 3, 0)
			}
		}
		return forms.TextField(ctx, field.Name, fieldLabel)
	case "integer", "number", "biginteger", "int", "int64", "uint", "uint64", "float", "float64":
		if field.Restrictions.MinValue != "" && field.Restrictions.MaxValue != "" {
			if format != "float" && format != "float64" {
				minValue, err := strconv.ParseInt(field.Restrictions.MinValue, 10, 64)
				if err != nil {
					panic(errors.New("minValue of field '" + field.Name + "' is invalid - " + field.Restrictions.MinValue))
				}
				maxValue, err := strconv.ParseInt(field.Restrictions.MaxValue, 10, 64)
				if err != nil {
					panic(errors.New("maxValue of field '" + field.Name + "' is invalid - " + field.Restrictions.MaxValue))
				}
				return forms.RangeField(ctx, field.Name, fieldLabel, minValue, maxValue, 1)
			}
		}
		return forms.NumberField(ctx, field.Name, fieldLabel)
	case "boolean", "bool", "checkbox":
		return forms.Checkbox(ctx, field.Name, fieldLabel)
	case "password":
		return forms.PasswordField(ctx, field.Name, fieldLabel)
	case "time":
		return forms.TimeField(ctx, field.Name, fieldLabel)
	case "date":
		return forms.DateField(ctx, field.Name, fieldLabel)
	case "datetime":
		return forms.DatetimeField(ctx, field.Name, fieldLabel)
	default:
		return forms.TextField(ctx, field.Name, fieldLabel)
	}
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
		} else if oerr, ok := err.(*gobatis.Error); ok && len(oerr.Validations) > 0 {
			var localeMessage string

			for _, validation := range oerr.Validations {
				localeMessage = validation.Message
				if key, found := localeMessages[validation.Code]; found {
					if key == "" {
						localeMessage = revel.Message(c.Request.Locale, validation.Code, validation.Message)
					} else {
						localeMessage = revel.Message(c.Request.Locale, key, validation.Message)
					}
				} else {
					localeMessage = revel.Message(c.Request.Locale, validation.Code, validation.Message)
				}

				if len(validation.Columns) > 0 {
					for _, column := range validation.Columns {
						c.Validation.Error(localeMessage).Key(column)
					}
				} else {
					c.Validation.Error(localeMessage)
				}
			}
			c.Validation.Keep()
			if len(oerr.Validations) == 1 && localeMessage != "" {
				c.Flash.Error(localeMessage)
			} else {
				c.Flash.Error(err.Error())
			}
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
