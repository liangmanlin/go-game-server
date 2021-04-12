package parser

const (
	PTBool PType = iota
	PTInt8
	PTInt16
	PTInt32
	PTInt64
	PTUInt16
	PTFloat32
	PTFloat64
	PTString
	PTStruct
	PTSlice
	PTMap
)

const (
	PBTos PBType = iota
	PBToc
	PBP
)

type PType = int

type PBType = int

type Proto struct {
	Name    string
	Type    PBType
	Mod     string
	SubMod  string
	ProtoID int
	HeadC   string
	LineC   string
	Fields  []*Field
}

type Field struct {
	Name  string
	FType *FieldType
	LineC string
}

type FieldType struct {
	Type  PType
	Name  string
	Key   PType      //只有map类型才有
	Value *FieldType //只有map|slice类型才有
}
