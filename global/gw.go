package global

import (
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
)

type TcpClientState struct {
	Conn  *gate.Conn
	Coder *pb.Coder
}
