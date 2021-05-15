package global

import "github.com/liangmanlin/gootp/db"

const (
	TABLE_ROLE_ID_INDEX = "role_id_index"
	TABLE_ACCOUNT       = "account"
	TABLE_BAG           = "bag"
	TABLE_ROLE_ATTR     = "role_attr"
	TABLE_ROLE_BASE     = "role_base"
	TABLE_ROLE_PROP     = "role_prop"
	TABLE_ROLE_MAP      = "role_map"
)

// 数据表需要在这里添加
var DBTables = []*db.TabDef{
	{Name: TABLE_ROLE_ID_INDEX, Pkey: []string{"AgentID", "ServerID"}, DataStruct: &RoleIDIndex{}},
	{Name: TABLE_ACCOUNT, Pkey: []string{"Account", "AgentID", "ServerID"}, DataStruct: &Account{}},
	{Name: TABLE_BAG, Pkey: []string{"RoleID", "ID"}, DataStruct: &Bag{}},
	{Name: TABLE_ROLE_ATTR, Pkey: []string{"RoleID"}, DataStruct: &PRoleAttr{}},
	{Name: TABLE_ROLE_BASE, Pkey: []string{"RoleID"}, DataStruct: &PRoleBase{}},
	{Name: TABLE_ROLE_PROP, Pkey: []string{"RoleID"}, DataStruct: &RoleProp{}},
	{Name: TABLE_ROLE_MAP, Pkey: []string{"RoleID"}, DataStruct: &RoleMap{}},
}

type RoleIDIndex struct {
	AgentID  int32
	ServerID int32
	Index    int32
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

type RoleProp struct {
	RoleID int64
	Prop   *PProp
	CeilFP []PCeilFP
}

type RoleMap struct {
	RoleID  int64
	MapID   int32
	MapName string
	Node    string
	X       int32
	Y       int32
}
