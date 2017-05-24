package types

func IsTable(cls *ClassDefinition) bool {
	_, ok := cls.Additional.(*TableData)
	return ok
}

func IsDynamic(spec *ClassDefinition) bool {
	return spec.IsAssignableTo(DynamicClass)
}

var (
	DynamicClass = &ClassDefinition{
		Name:           "dynamic",
		UnderscoreName: "dynamic",
		IsAbstractly:   true,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	IntegerClass = &ClassDefinition{
		Name:           "Integer",
		UnderscoreName: "integer",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	BigIntegerClass = &ClassDefinition{
		Name:           "BigInteger",
		UnderscoreName: "biginteger",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	StringClass = &ClassDefinition{
		Name:           "String",
		UnderscoreName: "string",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	DecimalClass = &ClassDefinition{
		Name:           "Decimal",
		UnderscoreName: "decimal",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	DatetimeClass = &ClassDefinition{
		Name:           "Datetime",
		UnderscoreName: "datetime",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	ObjectIdClass = &ClassDefinition{
		Name:           "ObjectId",
		UnderscoreName: "objectId",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	IpAddressClass = &ClassDefinition{
		Name:           "IPAddress",
		UnderscoreName: "ipAddress",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}

	PhysicalAddressClass = &ClassDefinition{
		Name:           "PhysicalAddress",
		UnderscoreName: "physicalAddress",
		IsAbstractly:   false,
		OwnFields:      map[string]*PropertyDefinition{},
		Fields:         map[string]*PropertyDefinition{},
	}
)
