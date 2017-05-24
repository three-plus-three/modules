package types

type ClassSpec struct {
	Super        string      `json:"super,omitempty" yaml:"super,omitempty"`
	Name         string      `json:"name" yaml:"name"`
	IsAbstractly bool        `json:"abstract,omitempty" yaml:"abstract,omitempty"`
	Keys         [][]string  `json:"keys,omitempty" yaml:"keys,omitempty"`
	Fields       []FieldSpec `json:"fields,omitempty" yaml:"fields,omitempty"`
}

type FieldSpec struct {
	Name         string           `json:"name" ymal:"name"`
	Description  string           `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string           `json:"type" yaml:"type"`
	Collection   bool             `json:"is_array,omitempty" yaml:"is_array,omitempty"`
	IsEmbedded   bool             `json:"embedded,omitempty" yaml:"embedded,omitempty"`
	IsRequired   bool             `json:"required,omitempty" yaml:"required,omitempty"`
	IsReadOnly   bool             `json:"readonly,omitempty" yaml:"readonly,omitempty"`
	IsUniquely   bool             `json:"unique,omitempty" yaml:"unique,omitempty"`
	DefaultValue string           `json:"default,omitempty" yaml:"default,omitempty"`
	Unit         string           `json:"unit,omitempty" yaml:"unit,omitempty"`
	Restrictions *RestrictionSpec `json:"restrictions,omitempty" yaml:"restrictions,omitempty"`
	Annotations  []interface{}    `json:"annotations,omitempty" yaml:"annotations,omitempty"`
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
