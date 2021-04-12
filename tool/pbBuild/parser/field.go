package parser

// 判断是否基础类型
func (f *FieldType)IsBase() bool {
	switch f.Type {
	case PTBool,PTInt8,PTInt16,PTInt32,PTInt64,PTUInt16,PTFloat32,PTFloat64,PTString:
		return true
	}
	return false
}

func (f *FieldType)IsArray() bool {
	return f.Type == PTSlice
}

func (f *FieldType)IsMap() bool {
	return f.Type == PTMap
}
