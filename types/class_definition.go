package types

import (
	"bytes"
)

type Annotation struct {
	Name  string
	Value interface{}
}

func mergeAnnotations(a, b []Annotation) []Annotation {
	res := make([]Annotation, 0, len(a)+len(b))
	return append(append(res, a...), b...)
}

func isEmbeddedAnnotation(ann Annotation) bool {
	return "embedded" == ann.Name
}

var EmbeddedAnnotation = Annotation{Name: "embedded"}

type CollectionType int

const (
	COLLECTION_UNKNOWN CollectionType = 0
	COLLECTION_ARRAY   CollectionType = 1
	COLLECTION_SET     CollectionType = 2
)

func (t CollectionType) IsArray() bool {
	return t == COLLECTION_ARRAY
}

func (t CollectionType) IsSet() bool {
	return t == COLLECTION_SET
}

func (t CollectionType) IsCollection() bool {
	return t == COLLECTION_SET || t == COLLECTION_ARRAY
}

type PropertyDefinition struct {
	Name         string
	Type         TypeDefinition
	Collection   CollectionType
	IsRequired   bool
	IsReadOnly   bool
	IsUniquely   bool
	Restrictions []Validator
	DefaultValue interface{}
	Annotations  []Annotation
}

func (self *PropertyDefinition) IsEmbedded() bool {
	for _, ann := range self.Annotations {
		if isEmbeddedAnnotation(ann) {
			return true
		}
	}
	return false
}

func (self *PropertyDefinition) IsSerial() bool {
	return "id" == self.Name
}

func (self *PropertyDefinition) IsPrimaryKey() bool {
	return "id" == self.Name
}

func (p *PropertyDefinition) PName() string {
	return p.Name
}
func (p *PropertyDefinition) TypeSpec() TypeDefinition {
	return p.Type
}
func (p *PropertyDefinition) CollectionType() CollectionType {
	return p.Collection
}
func (p *PropertyDefinition) Required() bool {
	return p.IsRequired
}
func (p *PropertyDefinition) ReadOnly() bool {
	return p.IsReadOnly
}
func (p *PropertyDefinition) Uniquely() bool {
	return p.IsUniquely
}
func (p *PropertyDefinition) Validators() []Validator {
	return p.Restrictions
}
func (p *PropertyDefinition) Default() interface{} {
	return p.DefaultValue
}

type KeyDefinition []*PropertyDefinition

type ClassDefinition struct {
	Name           string
	UnderscoreName string
	CollectionName string
	IsAbstractly   bool
	Keys           []KeyDefinition
	OwnFields      map[string]*PropertyDefinition
	Fields         map[string]*PropertyDefinition

	Super    *ClassDefinition
	Sons     []*ClassDefinition
	Children []*ClassDefinition

	Additional interface{}
}

func (self *ClassDefinition) CName() string { // CamelName
	return self.Name
}
func (self *ClassDefinition) UName() string { // UnderscoreName
	return self.UnderscoreName
}
func (self *ClassDefinition) Abstractly() bool {
	return self.IsAbstractly
}

func (self *ClassDefinition) GetKeys() []KeyDefinition {
	ret := make([]KeyDefinition, 0, len(self.Keys))
	for _, fields := range self.Keys {
		key := make([]*PropertyDefinition, 0, len(fields))
		for _, field := range fields {
			key = append(key, field)
		}
		ret = append(ret, key)
	}
	return ret
}

func (self *ClassDefinition) GetOwnProperties() []*PropertyDefinition {
	ret := make([]*PropertyDefinition, 0, len(self.OwnFields))
	for _, v := range self.OwnFields {
		ret = append(ret, v)
	}
	return ret
}

func (self *ClassDefinition) GetProperties() []*PropertyDefinition {
	ret := make([]*PropertyDefinition, 0, len(self.Fields))
	for _, v := range self.Fields {
		ret = append(ret, v)
	}
	return ret
}

func (self *ClassDefinition) GetProperty(nm string) *PropertyDefinition {
	column := self.Fields[nm]
	return column
}

func (self *ClassDefinition) GetOwnProperty(nm string) *PropertyDefinition {
	column := self.OwnFields[nm]
	return column
}

func (self *ClassDefinition) IsInheritanced() bool {
	return (nil != self.Super) || (nil != self.Sons && 0 != len(self.Sons))
}

func (self *ClassDefinition) ParentSpec() *ClassDefinition {
	return self.Super
}
func (self *ClassDefinition) RootSpec() *ClassDefinition {
	return self.Root()
}
func (self *ClassDefinition) Root() *ClassDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}
func (self *ClassDefinition) IsAssignableTo(super *ClassDefinition) bool {
	return self == super || self.IsSubclassOf(super)
}
func (self *ClassDefinition) IsInheritancedFrom(super *ClassDefinition) bool {
	return self.IsSubclassOf(super)
}
func (self *ClassDefinition) IsSubclassOf(cls *ClassDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == cls {
			return true
		}
	}
	return false
}

func (self *ClassDefinition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("class ")
	buffer.WriteString(self.Name)
	if nil != self.Super {
		buffer.WriteString(" < ")
		buffer.WriteString(self.Super.Name)
		buffer.WriteString(" { ")
	} else {
		buffer.WriteString(" { ")
	}
	if nil != self.OwnFields && 0 != len(self.OwnFields) {
		for _, pr := range self.OwnFields {
			buffer.WriteString(pr.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString(" }")
	return buffer.String()
}

type ClassDefinitions struct {
	underscore2Definitions map[string]*ClassDefinition
	clsDefinitions         map[string]*ClassDefinition
	//tableDefinitions       map[string]*ClassDefinition
}

func (self *ClassDefinitions) FindByUnderscoreName(nm string) *ClassDefinition {
	return self.underscore2Definitions[nm]
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	return self.clsDefinitions[nm]
}

// func (self *ClassDefinitions) FindByTableName(nm string) *ClassDefinition {
// 	return self.tableDefinitions[nm]
// }

func (self *ClassDefinitions) Register(cls *ClassDefinition) {
	self.clsDefinitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
	//self.tableDefinitions[cls.CollectionName] = cls
}

func (self *ClassDefinitions) Unregister(cls *ClassDefinition) {
	delete(self.clsDefinitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
	//delete(self.tableDefinitions, cls.CollectionName)
}

func (self *ClassDefinitions) All() map[string]*ClassDefinition {
	return self.clsDefinitions
}

func MakeClassDefinitions(capacity int) *ClassDefinitions {
	return &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, capacity),
		underscore2Definitions: make(map[string]*ClassDefinition, capacity),
		//tableDefinitions:       make(map[string]*ClassDefinition, capacity),
	}
}

func FindByName(sons []*ClassDefinition, name string) *ClassDefinition {
	for _, son := range sons {
		if son.Name == name {
			return son
		}
		if son.UnderscoreName == name {
			return son
		}
		if son.CollectionName == name {
			return son
		}
	}
	return nil
}
