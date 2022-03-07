package global

import (
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/rand"
	"github.com/liangmanlin/gootp/timer"
	"unsafe"
)

type Player struct {
	*kernel.Context
	Transaction
	RoleID         int64
	Conn           gate.Conn
	GWPid          *kernel.Pid
	MapPid         *kernel.Pid
	DirtyMod       map[string]bool
	BagMaxID       int32
	Bag            *BagData
	Attr           *PRoleAttr
	Base           *PRoleBase
	Timer          *timer.Timer
	PersistentTime int64
	Rand           *rand.Rand
	PropData       unsafe.Pointer // 规避循环引用
	Prop           *PProp
	Map            *RoleMap
	HeartTime      int64
}

type Transaction struct {
	IsTransaction bool
	BackupMap     map[BackupKey]int
	Backup        []BackData
	DBQueue       []DBQueue
}

type BagData struct {
	Goods     map[int32]*PGoods
	MaxSize   int32
	Dirty     map[int32]GoodsDirty
	TypeIDMap map[int32][]int32
}

type GoodsDirty struct {
	Type  int32
	Goods *PGoods
}

type DBQueue struct {
	OP  int32
	Fun func(player *Player, op int32, arg interface{})
	Arg interface{}
}

type BackData struct {
	Key   interface{}
	Value interface{}
}

type BackupKey struct {
	BackupID int32
	ID       int32
}

type HandleFunc = func(player *Player, proto interface{})

type KV struct {
	K int32
	V int32
}

type CreateRole struct {
	AgentID  int32
	ServerID int32
	Account  string
	Name     string
	HeroType int32
	Sex      int32
}

type CreateRoleResult struct {
	RoleID int64
	Roles  []int64
}

type TcpReConnect struct {
}

type TcpReConnectGW struct {
	Conn  gate.Conn
	RecID int64
	Token string
}

type Kick struct {
}

type OK struct {
}

type RoleDeadArg struct {
	SrcID   int64
	SrcType int8
}

type RoleUpdateProps struct {
	RoleID     int64
	FightPower int64
	UP         []*PKV
}

type RoleExitMap struct {
	RoleID int64
}

type MapRoleExit struct {
	ExitReturn bool
	X          float32
	Y          float32
}

type MapRoleEnter struct {
	Pid     *kernel.Pid
	MapID   int32
	MapName string
	X       float32
	Y       float32
}

// PMsgParam start---------------------------------------------------------

const (
	param_num byte = iota + 1
	param_string
	param_map_id
	param_role
)

type ParamNum int32

func (p ParamNum) ToBinary() []byte {
	return []byte{param_num, uint8(p >> 24), uint8(p >> 16), uint8(p >> 8), uint8(p)}
}

type ParamString string

func (p ParamString) ToBinary() []byte {
	size := len(p)
	buf := make([]byte, 0, 3+size)
	buf = append(buf, param_string, uint8(size>>8), uint8(size))
	buf = append(buf, p...)
	return buf
}

type ParamMapID int32

func (p ParamMapID) ToBinary() []byte {
	return []byte{param_map_id, uint8(p >> 24), uint8(p >> 16), uint8(p >> 8), uint8(p)}
}

type ParamRole struct {
	RoleID   int64
	RoleName string
}

func (p ParamRole) ToBinary() []byte {
	size := len(p.RoleName)
	buf := make([]byte, 0, 3+8+size)
	roleID := p.RoleID
	buf = append(buf, param_role, uint8(roleID>>56), uint8(roleID>>48), uint8(roleID>>40), uint8(roleID>>32),
		uint8(roleID>>24), uint8(roleID>>16), uint8(roleID>>8), uint8(roleID), uint8(size>>8), uint8(size))
	buf = append(buf, p.RoleName...)
	return buf
}

// PMsgParam end------------------------------------------------------
