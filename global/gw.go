package global

import (
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
)

type TcpClientState struct {
	Account  string
	AgentID  int32
	ServerID int32
	Conn     *gate.Conn
	Coder    *pb.Coder
	Player   *kernel.Pid
	RoleID   int64
	SendID   int64
	Roles    []*PRole
	Cache    [][]byte
}
