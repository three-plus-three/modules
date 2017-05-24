package types

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func GetID(cls *ClassDefinition) *PropertyDefinition {
	return cls.Additional.(*TableData).Id
}

func tryGetID(cls *ClassDefinition) *PropertyDefinition {
	if cls.Additional != nil {
		return nil
	}
	tData, _ := cls.Additional.(*TableData)
	if tData == nil {
		return nil
	}
	return tData.Id
}

func SetID(cls *ClassDefinition, column *PropertyDefinition) {
	tData, _ := cls.Additional.(*TableData)
	if tData == nil {
		tData = &TableData{Id: column}
		cls.Additional = tData
	} else {
		tData.Id = column
	}
}

func MakeIdColumn() *PropertyDefinition {
	return &PropertyDefinition{Name: "id", Type: GetTypeDefinition("objectId"),
		Collection: COLLECTION_UNKNOWN}
}

func makeAssocation(definitions map[string]*ClassDefinition, cls *ClassDefinition,
	t, tName, polymorphic, fKey string) (Assocation, error) {

	target, ok := definitions[tName]
	if !ok {
		return nil, errors.New("'" + tName + "' is not found.")
	}
	// if nil != target.Super {
	// 	*errs = append(*errs, "'"+tName+"' is a sub class.")
	// 	return nil
	// }
	is_polymorphic := false
	if "" == polymorphic {
		is_polymorphic = false
	} else if "true" == polymorphic {
		is_polymorphic = true
	} else {
		return nil, errors.New("attribute 'polymorphic' is unrecorign.")
	}

	if is_polymorphic {
		if "" != fKey {
			return nil, errors.New("'foreignKey' must is not present .")
		}
		fKey = "parent_id"
		pr, ok := target.OwnFields["parent_id"]
		if !ok {
			pr = &PropertyDefinition{Name: "parent_id", Type: GetTypeDefinition("objectId"),
				Collection: COLLECTION_UNKNOWN}

			target.OwnFields["parent_id"] = pr
		} else {
			if "objectId" != pr.Type.Name() {
				return nil, errors.New("'parent_id' is not objectId type")
			}
		}

		pr, ok = target.OwnFields["parent_type"]
		if !ok {
			pr = &PropertyDefinition{Name: "parent_type", Type: GetTypeDefinition("string"),
				Collection: COLLECTION_UNKNOWN}
			target.OwnFields["parent_type"] = pr
		} else {
			if "string" != pr.Type.Name() {
				return nil, errors.New("'parent_type' is reserved and must is a string type")
			}
			if pr.Collection.IsCollection() {
				return nil, errors.New("'parent_type' is reserved and is a collection")
			}
		}
	} else {
		if "" == fKey {
			fKey = Underscore(cls.Name) + "_id"
		}
		pr, ok := target.OwnFields[fKey]
		if !ok {
			pr = &PropertyDefinition{Name: fKey, Type: GetTypeDefinition("objectId"),
				Collection: COLLECTION_UNKNOWN}
			target.OwnFields[fKey] = pr
		} else {
			if "objectId" != pr.Type.Name() {
				return nil, errors.New("'foreignKey' is not objectId type")
			}
			if pr.Collection.IsCollection() {
				return nil, errors.New("'foreignKey' is a collection")
			}
		}
	}
	if "has_many" == t {
		return &HasMany{TargetTable: target, Polymorphic: is_polymorphic, ForeignKey: fKey}, nil
	}
	return &HasOne{TargetTable: target, Polymorphic: is_polymorphic, ForeignKey: fKey}, nil
}

func loadAssocations(definitions map[string]*ClassDefinition, cls *ClassDefinition, xmlDefinition *XMLClassDefinition, errs *[]string) {
	var assocations []Assocation

	if nil != xmlDefinition.BelongsTo && 0 != len(xmlDefinition.BelongsTo) {
		for _, belongs_to := range xmlDefinition.BelongsTo {
			target, ok := definitions[belongs_to.Target]
			if !ok {
				*errs = append(*errs, "belongs_to Target '"+belongs_to.Target+
					"' of class '"+xmlDefinition.Name+"' is not found.")
				continue
			}

			if "" == belongs_to.Name {
				belongs_to.Name = Underscore(belongs_to.Target) + "_id"
			}

			pr, ok := cls.OwnFields[belongs_to.Name]
			if !ok {
				pr = &PropertyDefinition{Name: belongs_to.Name, Type: GetTypeDefinition("objectId"),
					Collection: COLLECTION_UNKNOWN}
				cls.OwnFields[belongs_to.Name] = pr
			}

			assocations = append(assocations, &BelongsTo{TargetTable: target, Name: pr})
		}
	}
	if nil != xmlDefinition.HasMany && 0 != len(xmlDefinition.HasMany) {
		for _, hasMany := range xmlDefinition.HasMany {
			ass, err := makeAssocation(definitions, cls, "has_many", hasMany.Target,
				hasMany.Polymorphic, hasMany.ForeignKey)
			if nil != err {
				*errs = append(*errs, "load has_many '"+hasMany.Target+"' failed, "+err.Error())
			}
			if nil == ass {
				continue
			}

			assocations = append(assocations, ass)
		}
	}
	if nil != xmlDefinition.HasOne && 0 != len(xmlDefinition.HasOne) {
		for _, hasOne := range xmlDefinition.HasOne {
			ass, err := makeAssocation(definitions, cls, "has_one", hasOne.Target,
				"", hasOne.ForeignKey)
			if nil != err {
				*errs = append(*errs, "load has_one '"+hasOne.Target+"' failed, "+err.Error())
			}
			if nil == ass {
				continue
			}

			assocations = append(assocations, ass)
		}
	}
	if nil != xmlDefinition.HasAndBelongsToMany && 0 != len(xmlDefinition.HasAndBelongsToMany) {
		for _, habtm := range xmlDefinition.HasAndBelongsToMany {
			target, ok := definitions[habtm.Target]
			if !ok {
				*errs = append(*errs, "Target '"+habtm.Target+
					"' of has_and_belongs_to_many is not found.")
				continue
			}

			foreignKey := habtm.ForeignKey
			if "" == foreignKey {
				foreignKey = Underscore(cls.Name) + "_id"
			}

			through, ok := definitions[habtm.Through]
			if !ok {
				*errs = append(*errs, "Through '"+habtm.Through+
					"' of has_and_belongs_to_many is not found.")
				continue
			}

			assocations = append(assocations, &HasAndBelongsToMany{TargetTable: target,
				Through: through, ForeignKey: foreignKey})
		}
	}

	if len(assocations) > 0 {
		tableData, _ := cls.Additional.(*TableData)
		if tableData == nil {
			tableData = &TableData{Assocations: assocations}
			cls.Additional = tableData
		} else {
			tableData.Assocations = append(tableData.Assocations, assocations...)
		}
	}
}

func LoadFiles(files []string) (*TableDefinitions, error) {
	var xmlList []XMLClassDefinitions
	for _, nm := range files {
		f, err := ioutil.ReadFile(nm)
		if nil != err {
			return nil, fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
		}

		var xmlDefinition XMLClassDefinitions
		err = xml.Unmarshal(f, &xmlDefinition)
		if nil != err {
			return nil, fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
		}

		if nil == xmlDefinition.Definitions || 0 == len(xmlDefinition.Definitions) {
			return nil, fmt.Errorf("unmarshal xml '%s' error, definitions is empty", nm)
		}
		xmlList = append(xmlList, xmlDefinition)
	}

	res, errList := loadDefinitions(xmlList)
	if 0 != len(errList) {
		//errList = mergeErrors(nil, "load file '"+nm+"' error:", errList)
		return nil, errors.New(strings.Join(errList, "\r\n"))
	}
	return res, nil
}

func LoadTableDefinitions(nm string) (*TableDefinitions, error) {
	f, err := ioutil.ReadFile(nm)
	if nil != err {
		return nil, fmt.Errorf("read file '%s' failed, %s", nm, err.Error())
	}

	var xmlDefinition XMLClassDefinitions
	err = xml.Unmarshal(f, &xmlDefinition)
	if nil != err {
		return nil, fmt.Errorf("unmarshal xml '%s' failed, %s", nm, err.Error())
	}

	if nil == xmlDefinition.Definitions || 0 == len(xmlDefinition.Definitions) {
		return nil, fmt.Errorf("unmarshal xml '%s' error, definitions is empty", nm)
	}

	res, errList := loadDefinitions([]XMLClassDefinitions{xmlDefinition})

	if 0 != len(errList) {
		errList = mergeErrors(nil, "load file '"+nm+"' error:", errList)
		return nil, errors.New(strings.Join(errList, "\r\n"))
	}
	return res, nil
}

func loadDefinitions(xmlList []XMLClassDefinitions) (*TableDefinitions, []string) {
	var mixins = map[string][]*PropertyDefinition{}
	var errList []string

	for _, xmlDefinitionList := range xmlList {
		m, e := loadMixinFieldDefinitions("", xmlDefinitionList.Mixins, make([]string, 0, 10))
		if m != nil {
			for k, v := range m {
				mixins[k] = v
			}
		}
		if len(e) != 0 {
			errList = append(errList, e...)
		}
	}

	definitions := make(map[string]*ClassDefinition)
	// load table definitions and own properties
	for _, xmlDefinitionList := range xmlList {
		for _, xmlDefinition := range xmlDefinitionList.Definitions {
			_, ok := definitions[xmlDefinition.Name]
			if ok {
				errList = append(errList, "table '"+xmlDefinition.Name+
					"' is aleady exists.")
				continue
			}

			cls := &ClassDefinition{Name: xmlDefinition.Name,
				UnderscoreName: Underscore(xmlDefinition.Name),
				CollectionName: "tpt_" + Tableize(xmlDefinition.Name)}

			msgs := LoadOwnFields(xmlDefinition.Properties, cls)
			switch xmlDefinition.Abstract {
			case "true":
				cls.IsAbstractly = true
			case "false", "":
				cls.IsAbstractly = false
			default:
				msgs = append(msgs, "'abstract' value is invalid, it must is 'true' or 'false', actual is '"+xmlDefinition.Abstract+"'")
			}
			if nil != msgs && 0 != len(msgs) {
				errList = mergeErrors(errList, "", msgs)
			}

			if nil != xmlDefinition.Includes && 0 != len(xmlDefinition.Includes) {
				for _, includeMixin := range xmlDefinition.Includes {
					mixin, ok := mixins[includeMixin]
					if !ok {
						msgs = append(msgs, "mixin '"+includeMixin+"' isn't found.")
						continue
					}
					if nil != mixin {
						for _, pr := range mixin {
							if _, found := cls.OwnFields[pr.Name]; found {
								msgs = append(msgs, "property '"+pr.Name+"' is duplicated.")
								continue
							}
							cls.OwnFields[pr.Name] = pr
						}
					}
				}
			}
			for _, pr := range xmlDefinition.Properties {
				if nil != pr.Key {
					column, ok := cls.OwnFields[pr.Name]
					if !ok {
						panic("property '" + pr.Name + "' of '" + xmlDefinition.Name + "' is not found.")
					}

					cls.Keys = append(cls.Keys, []*PropertyDefinition{column})
				}
			}

			for _, combinedKey := range xmlDefinition.CombinedKeys {
				if nil == combinedKey.Names || 0 == len(combinedKey.Names) {
					log.Print("[WARN] '" + xmlDefinition.Name + "' has empty key.")
					continue
				}

				columns := make([]*PropertyDefinition, 0, len(combinedKey.Names))
				for _, nm := range combinedKey.Names {
					column, ok := cls.OwnFields[nm]
					if !ok {
						panic("property '" + nm + "' of '" + xmlDefinition.Name + "' is not found.")
					}
					columns = append(columns, column)
				}
				cls.Keys = append(cls.Keys, columns)
			}

			definitions[cls.Name] = cls
		}
	}

	// load super class
	for _, xmlDefinitionList := range xmlList {
		for _, xmlDefinition := range xmlDefinitionList.Definitions {
			cls, ok := definitions[xmlDefinition.Name]
			if !ok {
				continue
			}
			if "" == xmlDefinition.Base {
				continue
			}
			super, ok := definitions[xmlDefinition.Base]
			if !ok || nil == super {
				errList = append(errList, "Base '"+xmlDefinition.Base+
					"' of class '"+xmlDefinition.Name+"' is not found.")
			} else {
				if 0 == len(cls.Keys) {
					cls.Keys = super.Keys
				}
				cls.Super = super
			}
		}
	}

	// load own assocations
	for _, xmlDefinitionList := range xmlList {
		for _, xmlDefinition := range xmlDefinitionList.Definitions {
			cls, ok := definitions[xmlDefinition.Name]
			if !ok {
				continue
			}

			loadAssocations(definitions, cls, &xmlDefinition, &errList)
		}
	}

	return LoadAfter(definitions, errList)
}

func LoadAfter(definitions map[string]*ClassDefinition, errList []string) (*TableDefinitions, []string) {
	// load the properties of super class
	for _, cls := range definitions {
		if nil != cls.Super {
			for s := cls; ; s = s.Super {
				if s.Super == nil {
					id := tryGetID(s)
					if id == nil {
						id = MakeIdColumn()
						SetID(s, id)
						s.OwnFields[id.Name] = id
					}

					SetID(cls, id)
					break
				}
			}
			continue
		}

		id := tryGetID(cls)
		if id == nil {
			id = MakeIdColumn()
			SetID(cls, id)
		}
		cls.OwnFields[id.Name] = id
	}

	// load the properties of super class
	for _, cls := range definitions {
		loadParentFields(cls, &errList)
	}

	// load the children of super class
	for _, cls := range definitions {
		if nil == cls.Super {
			continue
		}

		cls.Super.Sons = append(cls.Super.Sons, cls)
		for s := cls.Super; nil != s; s = s.Super {
			s.Children = append(s.Children, cls)
		}
	}

	// reset collection name
	for _, cls := range definitions {
		if !IsSingleTableInheritance(cls) {
			for s := cls.Super; nil != s; s = s.Super {
				if IsSingleTableInheritance(s) {
					errList = append(errList, "'"+cls.Name+"' is not simple table inheritance, but parent table '"+s.Name+"' is simple table inheritance")
					break
				}
			}

			//fmt.Printf("%v --> not sti\r\n %v\r\n\r\n", cls.Name, cls.String())
			continue
		}

		last := cls.CollectionName

		for s := cls.Super; nil != s; s = s.Super {
			if !IsSingleTableInheritance(s) {
				break
			}
			last = s.CollectionName
		}
		//fmt.Printf("%v --> %v\r\n", cls.Name, cls.CollectionName)
		cls.CollectionName = last
	}

	// check id is exists.
	for _, cls := range definitions {
		if id, ok := cls.Fields["id"]; !ok || nil == id {
			errList = append(errList, "'"+cls.Name+"' has not 'id'")
		}

		//fmt.Println(cls.Name, cls.UnderscoreName, cls.CollectionName)
	}

	// change collection name
	// for _, cls := range definitions {
	// 	SetCollectionName(self, cls, &errList)
	// }

	// // check hierarchical of type
	// for _, cls := range definitions {
	// 	errList = checkHierarchicalType(self, cls, errList)
	// }

	if 0 != len(errList) {
		return nil, errList
	}

	res := NewTableDefinitions()
	for _, cls := range definitions {
		res.Register(cls)
	}
	return res, nil
}
