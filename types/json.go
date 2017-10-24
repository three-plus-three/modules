package types

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/three-plus-three/modules/as"
)

type ClassSpec struct {
	Super        string                 `json:"super,omitempty" yaml:"super,omitempty"`
	Name         string                 `json:"name" yaml:"name"`
	IsAbstractly bool                   `json:"abstract,omitempty" yaml:"abstract,omitempty"`
	Keys         [][]string             `json:"keys,omitempty" yaml:"keys,omitempty"`
	Fields       []FieldSpec            `json:"fields,omitempty" yaml:"fields,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type FieldSpec struct {
	Name         string                 `json:"name" ymal:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string                 `json:"type" yaml:"type"`
	Collection   bool                   `json:"is_array,omitempty" yaml:"is_array,omitempty"`
	IsEmbedded   bool                   `json:"embedded,omitempty" yaml:"embedded,omitempty"`
	IsRequired   bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	IsReadOnly   bool                   `json:"readonly,omitempty" yaml:"readonly,omitempty"`
	IsUniquely   bool                   `json:"unique,omitempty" yaml:"unique,omitempty"`
	DefaultValue string                 `json:"default,omitempty" yaml:"default,omitempty"`
	Unit         string                 `json:"unit,omitempty" yaml:"unit,omitempty"`
	Restrictions *RestrictionSpec       `json:"restrictions,omitempty" yaml:"restrictions,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
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

func (p *FieldSpec) IsMultipleChoices() bool {
	if p.Annotations == nil {
		return false
	}

	return as.BoolWithDefault(p.Annotations["multiple"], false)
}

func (p *FieldSpec) HasChoices() bool {
	if p.Restrictions == nil {
		if p.Annotations == nil {
			return false
		}
		source := as.StringWithDefault(p.Annotations["enumerationSource"], "")
		return source != ""
	}
	return len(p.Restrictions.Enumerations) > 0
}

func (p *FieldSpec) ToChoices() interface{} {
	if p.Annotations != nil {
		source := as.StringWithDefault(p.Annotations["enumerationSource"], "")
		if source != "" {
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
	}

	if p.Restrictions == nil || len(p.Restrictions.Enumerations) == 0 {
		return [][2]string{}
	}

	choices := make([][2]string, 0, len(p.Restrictions.Enumerations))
	for _, value := range p.Restrictions.Enumerations {
		choices = append(choices, [2]string{value, value})
	}
	return choices
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
	if p.IsReadOnly {
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
	if p.IsEmbedded {
		Embedded = "true"
	}

	var Collection string
	if p.Collection {
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
		DefaultValue: p.DefaultValue}

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
	Read(args string) (interface{}, error)
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

func (ep *enumerationProvidersImpl) Read(typ, args string) (interface{}, error) {
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

func (hp *HTTPProvider) Read(urlStr string) (interface{}, error) {
	if strings.Contains(urlStr, "{{") {
		t, err := template.New("main").Parse(urlStr)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		err = t.Execute(&buf, hp.args)
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
