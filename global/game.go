package global

import (
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/rand"
	"github.com/liangmanlin/gootp/timer"
)

type Player struct {
	RoleID         int64
	Conn           *gate.Conn
	GWPid          *kernel.Pid
	DirtyMod       map[string]bool
	Transaction
	BagMaxID       int32
	Bag            *BagData
	Attr           *PRoleAttr
	Timer          *timer.Timer
	PersistentTime int64
	Rand			*rand.Rand
}

type Transaction struct {
	IsTransaction  bool
	BackupMap      map[BackupKey]int
	Backup         []BackData
	DBQueue        []DBQueue
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

type HandleFunc = func(ctx *kernel.Context, player *Player, proto interface{})

type KV struct {
	K int32
	V int32
}

type Loop struct {
}
