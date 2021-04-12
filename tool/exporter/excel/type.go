package excel

type Excel struct {
	Name    string
	Header  []*Head
	AllData [][]string
}

type Head struct {
	Name     string
	Type     FType
	TypeName string
	Commit   string
	Rule     Rule
	Index    int
	Child    *Child // toGo专用，起始不会有值
}

type Child struct {
	Type   ChildType
	Key    KeyType
	Value  FType
	Fields []*Field
}
type Field struct {
	Name  string
	Child *Child
}

type ChildType int
type KeyType int

type Rule int
type FType int
