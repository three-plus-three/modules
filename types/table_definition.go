package types

import (
	"errors"
	"strconv"
)

type AssocationType int

func (self AssocationType) String() string {
	switch self {
	case BELONGS_TO:
		return "belongs_to"
	case HAS_ONE:
		return "has_one"
	case HAS_MANY:
		return "has_many"
	case HAS_AND_BELONGS_TO_MANY:
		return "has_and_belongs_to_many"
	default:
		return "assocation-" + strconv.Itoa(int(self))
	}
}

const (
	BELONGS_TO              AssocationType = 1
	HAS_MANY                AssocationType = 2
	HAS_ONE                 AssocationType = 3
	HAS_AND_BELONGS_TO_MANY AssocationType = 4
)

type Assocation interface {
	Type() AssocationType
	Target() *ClassDefinition
}

type BelongsTo struct {
	TargetTable *ClassDefinition
	Name        *PropertyDefinition
}

func (self *BelongsTo) Type() AssocationType {
	return BELONGS_TO
}

func (self *BelongsTo) Target() *ClassDefinition {
	return self.TargetTable
}

type HasMany struct {
	TargetTable *ClassDefinition
	ForeignKey  string
	Polymorphic bool
}

func (self *HasMany) Type() AssocationType {
	return HAS_MANY
}

func (self *HasMany) Target() *ClassDefinition {
	return self.TargetTable
}

type HasOne struct {
	TargetTable *ClassDefinition
	ForeignKey  string
	Polymorphic bool
}

func (self *HasOne) Type() AssocationType {
	return HAS_ONE
}

func (self *HasOne) Target() *ClassDefinition {
	return self.TargetTable
}

type HasAndBelongsToMany struct {
	TargetTable *ClassDefinition
	Through     *ClassDefinition
	ForeignKey  string
}

func (self *HasAndBelongsToMany) Type() AssocationType {
	return BELONGS_TO
}

func (self *HasAndBelongsToMany) Target() *ClassDefinition {
	return self.TargetTable
}

//type ClassDefinition ClassDefinition

type TableData struct {
	Id          *PropertyDefinition
	Assocations []Assocation
}

/*
func (self *ClassDefinition) IsInheritanced() bool {
	return (nil != self.Super) || (nil != self.OwnChildren && !self.OwnChildren.IsEmpty())
}

func (self *ClassDefinition) ParentSpec() ClassSpec {
	if nil == self.Super {
		return nil
	}
	return self.Super
}

func (self *ClassDefinition) RootSpec() ClassSpec {
	return self.Root()
}

func (self *ClassDefinition) Root() *ClassDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}

func (self *ClassDefinition) IsAssignableTo(super ClassSpec) bool {
	parent, ok := super.(*ClassDefinition)
	if !ok {
		return false
	}
	return self == parent || self.IsSubclassOf(parent)
}

func (self *ClassDefinition) IsInheritancedFrom(super ClassSpec) bool {
	parent, ok := super.(*ClassDefinition)
	if !ok {
		return false
	}
	return self.IsSubclassOf(parent)
}

func (self *ClassDefinition) IsSubclassOf(super *ClassDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == super {
			return true
		}
	}
	return false
}

func (self *ClassDefinition) IsSingleTableInheritance() bool {
	_, ok := self.Fields["type"]
	if ok {
		return self.IsInheritanced()
	}
	return false
}

func (self *ClassDefinition) HasChildren() bool {
	return (nil != self.OwnChildren && !self.OwnChildren.IsEmpty())
}

func (self *ClassDefinition) FindByUnderscoreName(nm string) *ClassDefinition {
	if self.UnderscoreName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.FindByUnderscoreName(nm)
}

func (self *ClassDefinition) FindByTableName(nm string) *ClassDefinition {
	if self.CollectionName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.FindByTableName(nm)
}

func (self *ClassDefinition) Find(nm string) *ClassDefinition {
	if self.UnderscoreName == nm {
		return self
	}
	if !self.HasChildren() {
		return nil
	}
	return self.Children.Find(nm)
}

func (self *ClassDefinition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("table ")
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
*/

func IsSingleTableInheritance(self *ClassDefinition) bool {
	_, ok := self.Fields["type"]
	if ok {
		return self.IsInheritanced()
	}
	return false
}

func HasChildren(self *ClassDefinition) bool {
	return len(self.Sons) != 0
}

func FindByClassName(self *ClassDefinition, nm string) *ClassDefinition {
	if self.Name == nm {
		return self
	}

	if self.UnderscoreName == nm {
		return self
	}

	for _, son := range self.Children {
		if son.Name == nm {
			return son
		}
		if son.UnderscoreName == nm {
			return son
		}
	}

	return nil
}

func FindByTableName(self *ClassDefinition, nm string) *ClassDefinition {
	if self.CollectionName == nm {
		return self
	}

	for _, son := range self.Sons {
		if child := FindByTableName(son, nm); child != nil {
			return child
		}
	}

	return nil
}

func GetAssocations(self *ClassDefinition) []Assocation {
	tableData, _ := self.Additional.(*TableData)
	if nil == tableData {
		return nil
	}
	return tableData.Assocations
}

func GetAssocation(self *ClassDefinition, target *ClassDefinition,
	foreignKeyOrName string,
	types ...AssocationType) (Assocation, error) {
	assocations := GetAssocationByTargetAndTypes(self, target, types...)
	if 0 == len(assocations) {
		assocations = GetAssocationByTypes(self, types...)
		if nil != target {
			assocations_copy := make([]Assocation, 0, len(assocations))
			for _, assocation := range assocations {
				if assocation.Target().IsAssignableTo(target) {
					assocations_copy = append(assocations_copy, assocation)
				}
			}
			assocations = assocations_copy
		}
		if 0 == len(assocations) {
			return nil, errors.New("table '" + self.UnderscoreName + "' and table '" +
				target.UnderscoreName + "' has not assocations.")
		}
	}

	if 0 == len(foreignKeyOrName) {
		if 1 != len(assocations) {
			return nil, errors.New("table '" + self.UnderscoreName + "' and table '" +
				target.UnderscoreName + "' is ambiguous.")
		}
		return assocations[0], nil
	}

	assocations_by_foreignKey := make([]Assocation, 0, len(assocations))
	for _, assocation := range assocations {
		switch assocation.Type() {
		case HAS_ONE:
			hasOne := assocation.(*HasOne)
			if hasOne.ForeignKey == foreignKeyOrName {
				assocations_by_foreignKey = append(assocations_by_foreignKey, assocation)
			}

		case HAS_MANY:
			hasMany := assocation.(*HasMany)
			if hasMany.ForeignKey == foreignKeyOrName {
				assocations_by_foreignKey = append(assocations_by_foreignKey, assocation)
			}

		case BELONGS_TO:
			belongsTo := assocation.(*BelongsTo)
			if belongsTo.Name.Name == foreignKeyOrName {
				assocations_by_foreignKey = append(assocations_by_foreignKey, assocation)
			}
		default:
			return nil, errors.New("Unsupported Assocation - " + assocation.Type().String())
		}
	}

	if 0 == len(assocations_by_foreignKey) {
		return nil, errors.New("such assocation is not exists.")
	}
	if 1 != len(assocations_by_foreignKey) {
		return nil, errors.New("table '" + self.UnderscoreName + "' and table '" +
			target.UnderscoreName + "' is ambiguous.")
	}
	return assocations_by_foreignKey[0], nil
}

func GetAssocationByTarget(self, cls *ClassDefinition) []Assocation {
	var assocations []Assocation
	tableData, _ := self.Additional.(*TableData)

	if nil != tableData && nil != tableData.Assocations {
		for _, assoc := range tableData.Assocations {
			if cls.IsSubclassOf(assoc.Target()) {
				assocations = append(assocations, assoc)
			}
		}
	}

	if nil == self.Super {
		return assocations
	}

	if nil == assocations {
		return GetAssocationByTarget(self.Super, cls)
	}

	res := GetAssocationByTarget(self.Super, cls)
	if nil != res {
		assocations = append(assocations, res...)
	}

	return assocations
}

func GetAssocationByTypes(self *ClassDefinition, assocationTypes ...AssocationType) []Assocation {
	return GetAssocationByTargetAndTypes(self, nil, assocationTypes...)
}

func GetAssocationByTargetAndTypes(self, cls *ClassDefinition,
	assocationTypes ...AssocationType) []Assocation {
	var assocations []Assocation
	tableData, _ := self.Additional.(*TableData)

	if nil != tableData && nil != tableData.Assocations {
		for _, assoc := range tableData.Assocations {
			found := false
			for _, assocationType := range assocationTypes {
				if assocationType == assoc.Type() {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if nil == cls || cls.IsSubclassOf(assoc.Target()) {
				assocations = append(assocations, assoc)
			}
		}
	}

	if nil == self.Super {
		return assocations
	}

	if nil == assocations {
		return GetAssocationByTargetAndTypes(self.Super, cls, assocationTypes...)
	}

	res := GetAssocationByTargetAndTypes(self.Super, cls, assocationTypes...)
	if nil != res {
		assocations = append(assocations, res...)
	}

	return assocations
}

type TableDefinitions struct {
	underscore2Definitions map[string]*ClassDefinition
	definitions            map[string]*ClassDefinition
	table2definitions      map[string]*ClassDefinition
}

func NewTableDefinitions() *TableDefinitions {
	return &TableDefinitions{underscore2Definitions: make(map[string]*ClassDefinition),
		definitions:       make(map[string]*ClassDefinition),
		table2definitions: make(map[string]*ClassDefinition)}
}

func (self *TableDefinitions) FindByUnderscoreName(nm string) *ClassDefinition {
	return self.underscore2Definitions[nm]
}

func (self *TableDefinitions) FindByTableName(nm string) *ClassDefinition {
	return self.table2definitions[nm]
}

func (self *TableDefinitions) Find(nm string) *ClassDefinition {
	return self.definitions[nm]
}

func stiRoot(cls *ClassDefinition) *ClassDefinition {
	for s := cls; ; s = s.Super {
		if nil == s.Super {
			return s
		}
		if s.Super.CollectionName != cls.CollectionName {
			return s
		}
	}
}

func (self *TableDefinitions) Register(cls *ClassDefinition) {
	self.definitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
	if table, ok := self.table2definitions[cls.CollectionName]; ok {
		if table.IsSubclassOf(cls) {
			self.table2definitions[cls.CollectionName] = cls
		} else if stiRoot(cls) != stiRoot(table) {
			panic("table '" + cls.Name + "' and table '" + table.Name + "' is same with collection name.")
		}
	} else {
		self.table2definitions[cls.CollectionName] = cls
	}
}

func (self *TableDefinitions) Unregister(cls *ClassDefinition) {
	delete(self.definitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
	if !IsSingleTableInheritance(cls) {
		delete(self.table2definitions, cls.CollectionName)
	} else {
		if _, ok := self.table2definitions[cls.CollectionName]; ok {
			if stiRoot(cls) == cls {
				delete(self.table2definitions, cls.CollectionName)
			}
		}
	}
}

func (self *TableDefinitions) All() map[string]*ClassDefinition {
	return self.definitions
}

func (self *TableDefinitions) Len() int {
	return len(self.definitions)
}

func (self *TableDefinitions) IsEmpty() bool {
	return 0 == len(self.definitions)
}

func (self *TableDefinitions) UnderscoreAll() map[string]*ClassDefinition {
	return self.underscore2Definitions
}
