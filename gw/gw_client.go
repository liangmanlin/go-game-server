package gw

import (
	"game/global"
	"github.com/liangmanlin/gootp/gate"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
	"unsafe"
)

var TcpClientActor = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) unsafe.Pointer {
		t := global.TcpClientState{}
		t.Coder = args[0].(*pb.Coder)
		t.Conn = args[1].(*gate.Conn)
		kernel.ErrorLog("connect")
		return unsafe.Pointer(&t)
	},
	HandleCast: func(context *kernel.Context, msg interface{}) {
		t := (*global.TcpClientState)(context.State)
		switch m := msg.(type) {
		case bool:
			doLogin(t, context)
		case []byte:
			sendBuf(t,m)
		case *gate.TcpError:
			context.Exit("normal")
		case int:
			context.Exit("normal")
		case *global.TcpReConnectGW:
			doReConnect(t,context,m)
		case gate.Pack:
			switch proto := m.Proto.(type) {
			case *global.LoginTosLogin:
				doLoginLogin(t,context,proto)
			case *global.LoginTosSelect:
				doLoginSelect(t,context,proto)
			case *global.LoginTosCreateRole:
				doLoginCreate(t,context,proto)
			default:
				// 只有极少数情况会走到这里
				if t.Player != nil {
					context.Cast(t.Player,m)
				}
			}
		default:
			kernel.ErrorLog("un handle msg: %#v", m)
		}
	},
	HandleCall: func(context *kernel.Context, request interface{}) interface{} {
		return nil
	},
	Terminate: func(context *kernel.Context, reason *kernel.Terminate) {
		t := (*global.TcpClientState)(context.State)
		if t.Conn != nil {
			t.Conn.Close()
		}
		kernel.ErrorLog("exit reason:%s", reason.Reason)
	},
	ErrorHandler: func(context *kernel.Context, err interface{}) bool {
		return false
	},
}

func doLogin(t *global.TcpClientState, context *kernel.Context) {
	t.Conn.SetHead(2)
	err, buf := t.Conn.Recv(0, 5)
	if err != nil {
		kernel.ErrorLog("handshake error %s", err.Error())
		context.Exit(kernel.ExitReasonNormal)
		return
	}
	_, proto := pb.GetCoder(1).Decode(buf)
	pt := proto.(*global.LoginTosConnect)
	if !pt.IsReconnect {
		token := RandomToken() // 生成一个唯一key,作为重连的标志
		InsertToken(token, context.Self())
		rs := &global.LoginTocConnect{Succ: true, IsReconnect: false, Token: token}
		if t.Conn.SendBufHead(t.Coder.Encode(rs, 2)) != nil {
			context.Exit(kernel.ExitReasonNormal)
			return
		}
		// 异步接收
		t.Conn.StartReaderDecode(context.Self(),pb.GetCoder(1).Decode)
	}else{
		pid := TokenToPid(pt.Token)
		if pid != nil {
			// 重连，转移到目标进程
			context.Cast(pid,&global.TcpReConnectGW{Conn: t.Conn,RecID: pt.RecID,Token: pt.Token})
			rs := &global.LoginTocConnect{Succ: true, IsReconnect: true, Token: pt.Token}
			_ = t.Conn.SendBufHead(t.Coder.Encode(rs, 2))
			t.Conn = nil
		}else{
			rs := &global.LoginTocConnect{Succ: false, IsReconnect: true, Token: ""}
			_ = t.Conn.SendBufHead(t.Coder.Encode(rs, 2))
		}
		context.Exit(kernel.ExitReasonNormal)
	}
}

func SendPack(state *global.TcpClientState,pack interface{})  {
	buf := state.Coder.Encode(pack,2)
	sendBuf(state,buf)
}

func sendBuf(state *global.TcpClientState,buf []byte)  {
	state.SendID++
	state.Cache[state.SendID % CacheSize] = buf
	if state.Conn != nil {
		if state.Conn.SendBufHead(buf) != nil {
			state.Conn.Close()
			state.Conn = nil
		}
	}
}