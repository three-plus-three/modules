package types

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/three-plus-three/modules/errors"
)

var emptyArgs = []interface{}{}

type ClassSpec struct {
	Super        string                 `json:"super,omitempty" yaml:"super,omitempty"`
	Name         string                 `json:"name" yaml:"name"`
	IsAbstractly bool                   `json:"abstract,omitempty" yaml:"abstract,omitempty"`
	Keys         [][]string             `json:"keys,omitempty" yaml:"keys,omitempty"`
	Fields       []FieldSpec            `json:"fields,omitempty" yaml:"fields,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type FieldSpec struct {
	Name           string                 `json:"name" ymal:"name"`
	Label          string                 `json:"label,omitempty" ymal:"label,omitempty"`
	Description    string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Type           string                 `json:"type" yaml:"type"`
	Format         string                 `json:"format" yaml:"format"`
	IsArray        bool                   `json:"is_array,omitempty" yaml:"is_array,omitempty"`
	IsOriginal     bool                   `json:"is_original,omitempty" yaml:"is_original,omitempty"`
	IsRequired     bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	IsReadonly     bool                   `json:"readonly,omitempty" yaml:"readonly,omitempty"`
	IsUniquely     bool                   `json:"unique,omitempty" yaml:"unique,omitempty"`
	Default        string                 `json:"default,omitempty" yaml:"default,omitempty"`
	Unit           string                 `json:"unit,omitempty" yaml:"unit,omitempty"`
	Restrictions   *RestrictionSpec       `json:"restrictions,omitempty" yaml:"restrictions,omitempty"`
	Annotations    map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type RestrictionSpec struct {
	Enumerations []string `json:"enumerations,omitempty" yaml:"enumerations,omitempty"`
	Pattern      string   `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinValue     string   `json:"minValue,omitempty" yaml:"minValue,omitempty"`
	MaxValue     string   `json:"maxValue,omitempty" yaml:"maxValue,omitempty"`
	Length       string   `json:"length,omitempty" yaml:"length,omitempty"`
	MinLength    string   `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	MaxLength    string   `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
}

func (p *FieldSpec) DefaultValue() interface{} {
	if p.Default == "" {
		if p.IsArray {
			return []string{}
		}
		return nil
	}
	return p.Default
}

func (p *FieldSpec) HasChoices() bool {
	if p.Restrictions != nil && len(p.Restrictions.Enumerations) > 0 {
		return true
	}

	if p.Annotations == nil {
		return false
	}

	enumerationSource := p.Annotations["enumerationSource"]
	return enumerationSource != nil
}

func (p *FieldSpec) ToChoices() interface{} {
	if p.Restrictions != nil && len(p.Restrictions.Enumerations) > 0 {
		return p.Restrictions.Enumerations
	}

	if p.Annotations == nil {
		return [][2]string{}
	}

	enumerationSource := p.Annotations["enumerationSource"]
	if enumerationSource == nil {
		return [][2]string{}
	}
	if source, ok := enumerationSource.(string); ok {
		ss := strings.SplitN(source, ",", 2)
		if len(ss) != 2 {
			panic(errors.New("enumerationSource is invalid value - " + source))
		}
		values, err := enumerationProviders.Read(ss[0], ss[1])
		if err != nil {
			panic(errors.New("ToChoices: " + err.Error()))
		}
		return values
	}

	if source, ok := enumerationSource.(map[string]interface{}); ok {
		typ := source["type"]
		if typ == nil {
			panic(errors.New("type of enumerationSource is nil"))
		}

		values, err := enumerationProviders.Read(fmt.Sprint(typ), source)
		if err != nil {
			panic(errors.New("ToChoices: " + err.Error()))
		}
		return values
	}

	return enumerationSource
}

func (p *FieldSpec) CSSClasses() string {
	classes := []string{}

	if p.IsRequired {
		classes = append(classes, "required")
	}

	if p.Type == "integer" {
		classes = append(classes, "digits")
	} else if p.Type == "decimal" {
		classes = append(classes, "number")
	} else if p.Type == "date" {
		classes = append(classes, "dateISO")
	} else if p.Type == "ipAddress" {
		classes = append(classes, "ipv4")
	} else if p.Type == "physicalAddress" {
		classes = append(classes, "macaddress")
	}

	return strings.Join(classes, " ")
}

func (p *FieldSpec) ToXML() *XMLPropertyDefinition {
	var ReadOnly *XMLReadOnly
	if p.IsReadonly {
		ReadOnly = &XMLReadOnly{}
	}
	var Unique *XMLUnique
	if p.IsUniquely {
		Unique = &XMLUnique{}
	}
	var Required *XMLRequired
	if p.IsRequired {
		Required = &XMLRequired{}
	}

	var Embedded = "false"
	if !p.IsOriginal {
		Embedded = "true"
	}

	var Collection string
	if p.IsArray {
		Collection = "array"
	}

	xpd := &XMLPropertyDefinition{
		Name:       p.Name,
		Type:       p.Type,
		Embedded:   Embedded,
		Collection: Collection,

		Unit:         p.Unit,
		ReadOnly:     ReadOnly,
		Unique:       Unique,
		Required:     Required,
		DefaultValue: p.Default}

	if p.Restrictions != nil {
		if len(p.Restrictions.Enumerations) != 0 {
			xpd.Enumerations = &p.Restrictions.Enumerations
		}

		xpd.Pattern = p.Restrictions.Pattern
		xpd.MinValue = p.Restrictions.MinValue
		xpd.MaxValue = p.Restrictions.MaxValue
		xpd.Length = p.Restrictions.Length
		xpd.MinLength = p.Restrictions.MinLength
		xpd.MaxLength = p.Restrictions.MaxLength
	}

	return xpd
}

var enumerationProviders = enumerationProvidersImpl{
	providers: map[string]EnumerationProvider{},
}

// RegisterEnumerationProvider 注册一个新的枚举值提供接口
func RegisterEnumerationProvider(typ string, provider EnumerationProvider) {
	enumerationProviders.Register(typ, provider)
}

// EnumerationProvider 枚举值提供接口
type EnumerationProvider interface {
	Read(args interface{}) (interface{}, error)
}

type enumerationProvidersImpl struct {
	mu        sync.RWMutex
	providers map[string]EnumerationProvider
}

func (ep *enumerationProvidersImpl) Register(typ string, provider EnumerationProvider) {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	if _, ok := ep.providers[typ]; ok {
		panic(errors.New("enumerationSource '" + typ + "' is already exists"))
	}
	ep.providers[typ] = provider
}

func (ep *enumerationProvidersImpl) Read(typ string, args interface{}) (interface{}, error) {
	ep.mu.RLock()
	defer ep.mu.RUnlock()

	provider := ep.providers[typ]
	if provider == nil {
		return nil, errors.New("enumerationSource '" + typ + "' is unsupported")
	}
	return provider.Read(args)
}

// 一个简单的 枚举值 扩展如下
// 1. 第一步先注册扩展接口
// func init() {
// 	provider := &dbProvider{}
// 	types.Register("sql", provider)
// }
//
// type dbProvider struct {}
// func (dp *dbProvider) Read(args string) (interface{}, error)
// 		var choice []forms.InputChoice
// 		err := Lifecycle.DB.Query(args).All(&choice)
// 		return choice, err
// }
//
// 2. 保存数据， 如下
// p.Annotations["enumerationSource"] = "sql," + "select label, value from xxxx"

// HTTPProvider 一个缺省的 http 格式的 枚举值 扩展接口
type HTTPProvider struct {
	args map[string]interface{}
}

func (hp *HTTPProvider) Read(urlObject interface{}) (interface{}, error) {
	args := hp.args
	var urlStr string
	switch v := urlObject.(type) {
	case string:
		urlStr = v
	case map[string]interface{}:
		for _, name := range []string{"value", "url"} {
			value := v[name]
			if value == nil {
				continue
			}
			if s, ok := value.(string); ok {
				urlStr = s
				break
			}
		}

		if args == nil {
			args = map[string]interface{}{}
		}
		for key, value := range v {
			args[key] = value
		}
	default:
		return nil, fmt.Errorf("HTTPProvider: args is unknow type - %T %#v", urlObject, urlObject)
	}

	if strings.Contains(urlStr, "{{") {
		t, err := template.New("main").Parse(urlStr)
		if err != nil {
			return nil, err
		}
		var buf strings.Builder
		err = t.Execute(&buf, args)
		if err != nil {
			return nil, err
		}
		urlStr = buf.String()
	}
	res, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	bs, err := ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		if len(bs) != 0 {
			return nil, errors.New(res.Status + " - " + string(bs))
		}
		return nil, errors.New(res.Status)
	}
	// 因为 forms 中的 select 控件支持 []byte， 所以就不序列化了
	// var choices []forms.InputChoice
	// err = json.Unmarshal(bs, &choices)
	// return  choices, err

	return bs, err
}

// Queryer 数据的查询接口
type Queryer interface {
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
}

// DbProvider 一个缺省的 db 格式的 枚举值 扩展接口
type DbProvider struct {
	DB Queryer
}

func (dp *DbProvider) Read(a interface{}) (interface{}, error) {
	var args []interface{}
	var sqlStr string
	switch v := a.(type) {
	case string:
		sqlStr = v
	case map[string]interface{}:
		for _, name := range []string{"value", "sql"} {
			value := v[name]
			if value == nil {
				continue
			}
			if s, ok := value.(string); ok {
				sqlStr = s
				break
			}
		}

		args, _ = v["arguments"].([]interface{})
	default:
		return nil, fmt.Errorf("HTTPProvider: args is unknow type - %T %#v", a, a)
	}

	if args == nil {
		args = emptyArgs
	}
	opts, err := ReadInputChoices(dp.DB, sqlStr, args...)
	return opts, err
}

// ReadInputChoices 从数据库中读选项
func ReadInputChoices(db Queryer, query string, args ...interface{}) ([][2]string, error) {
	rows, err := db.QueryContext(context.Background(), query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "查询结果失败:sql:"+query)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, errors.Wrap(err, "查询结果失败:sql:"+query)
	}

	var firstValue = true
	if strings.ToLower(columns[0]) != "value" {
		firstValue = false
	}

	var opts [][2]string
	for rows.Next() {
		var id, value sql.NullString
		err = rows.Scan(&id, &value)
		if err != nil {
			return nil, errors.Wrap(err, "查询结果失败:sql:"+query)
		}

		if firstValue {
			opts = append(opts, [2]string{id.String, value.String})
		} else {
			opts = append(opts, [2]string{value.String, id.String})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "查询结果失败:sql:"+query)
	}
	return opts, nil
}
