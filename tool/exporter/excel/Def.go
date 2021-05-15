package excel

const (
	RuleNone Rule = iota
	RuleCommon
	RuleServer
	RuleClient
)

const (
	FTypeInt FType = iota
	FTypeFloat
	FTypeString
	FTypeKey
	FTypeTerm
	FTypeSlice
)

var ruleMap = map[int]Rule{
	17: RuleCommon,
	40: RuleClient,
	51: RuleServer,
	55: RuleNone,
}

const (
	HeadChildTypeStruct ChildType = iota
	HeadChildTypeSlice
	HeadChildTypeMap
)

const (
	KeyTypeInt KeyType = iota
	KeyTypeString
)


