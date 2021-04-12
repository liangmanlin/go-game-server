package global

import "github.com/liangmanlin/gootp/db"

// 数据表需要在这里添加
var DBTables = []*db.TabDef{
	{"account",  []string{"Account"},  []string{"AgentID"},&Account{},},
	{"bag",[]string{"RoleID","ID"},nil,&Bag{}},
	{"role_attr",[]string{"RoleID"},nil,&PRoleAttr{}},
}

type Account struct {
	Account     string
	AgentID     int32
	ServerID    int32
	LastRole    int64
	FcmOffline  int32
	LastOffline int32
	BanTime     int32
}

// 需要与Goods保持一致
type Bag struct {
	RoleID     int64
	ID         int32
	Type       int32
	TypeID     int32
	Num        int32
	Bind       bool
	StartTime  int32
	EndTime    int32
	CreateTime int32
}
