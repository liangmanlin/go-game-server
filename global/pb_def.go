package global

//--------------------------Login------------------------------------------------------------

// 登录协议
type LoginTosLogin struct { // router:LoginLogin
	Account  string
	AgentID  int32
	ServerID int32
}
type LoginTocLogin struct {
	Succ       bool
	Reason     *PMsg
	Account    string
	LastRoleID int64
	Roles      []*PRole
}

//--------------------------Game------------------------------------------------------------

type GameTosUP struct { // router:LoginLogin

}

type GameTocUP struct {
}

type PMsg struct {
	MsgID int32
	Bin   []byte // 客户端需要根据商定的规则解开一个字符参数
}

type PRole struct {
	RoleID   int64
	RoleName string
	HeroType int32
	Level    int32
	Skin     *PSkin
}

type PSkin struct {
}

type PGoods struct {
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

type PRoleAttr struct {
	RoleID  int64
	Diamond int64
	Gold    int64
}
