package gw

import (
	"fmt"
	"game/global"
	"game/lib"
	"game/player"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/kernel"
)

func doReConnect(state *global.TcpClientState, ctx *kernel.Context, m *global.TcpReConnectGW) {
	if state.SendID-m.RecID > CacheSize {
		rs := &global.LoginTocConnect{Succ: false, IsReconnect: true}
		_ = state.Conn.SendBufHead(state.Coder.Encode(rs, 2))
		return
	}
	rs := &global.LoginTocConnect{Succ: true, IsReconnect: true, Token: m.Token}
	_ = state.Conn.SendBufHead(state.Coder.Encode(rs, 2))
	// 补包
	syncPack(state, m.RecID)
	// 通知player
	if state.Player != nil {
		ctx.Cast(state.Player, global.TcpReConnect{})
		m.Conn.StartReaderDecode(state.Player, pb.GetCoder(1).Decode)
	}
}

func doLoginLogin(state *global.TcpClientState, ctx *kernel.Context, m *global.LoginTosLogin) {
	t := db.ModSelectRow(db.GameDB, global.TABLE_ACCOUNT, m.Account, m.AgentID, m.ServerID)
	var account *global.Account
	if t == nil {
		account = &global.Account{Account: m.Account, AgentID: m.AgentID, ServerID: m.ServerID}
		db.ModInsert(db.GameDB, global.TABLE_ACCOUNT, account)
	} else {
		account = t.(*global.Account)
	}
	baseList := db.ModSelectAllWhere(db.GameDB, global.TABLE_ROLE_BASE,
		fmt.Sprintf("Account = %s and AgentID = %d and ServerID = %d", db.Encode(m.Account), m.AgentID, m.ServerID))
	var roles []*global.PRole
	for i := range baseList {
		b := baseList[i].(*global.PRoleBase)
		role := &global.PRole{RoleID: b.RoleID, RoleName: b.Name, HeroType: b.HeroType, Level: b.Level, Skin: b.Skin}
		roles = append(roles, role)
	}
	state.Roles = roles
	r := &global.LoginTocLogin{Succ: true, Account: m.Account, LastRoleID: account.LastRole, Roles: roles}
	SendPack(state, r)
	state.Account = m.Account
	state.AgentID = m.AgentID
	state.ServerID = m.ServerID
}

func doLoginSelect(state *global.TcpClientState, ctx *kernel.Context, m *global.LoginTosSelect) {
	for i := range state.Roles {
		if state.Roles[i].RoleID == m.RoleID {
			state.RoleID = m.RoleID
			break
		}
	}
	if state.RoleID == 0 {
		// 错误
		r := &global.LoginTocSelect{Succ: false, Reason: &global.PMsg{MsgID: 1}}
		SendPack(state, r)
		return
	}
	var rl []int64
	for i := range state.Roles{
		rl = append(rl,state.Roles[i].RoleID)
	}
	kickRoles(ctx,rl,state.RoleID)
	playerPid := player.Start(ctx.Self(), state.Conn, state.RoleID)
	if playerPid != nil {
		state.Roles = nil // gc
		state.Player = playerPid
		r := &global.LoginTocSelect{Succ: true, RoleID: state.RoleID}
		SendPack(state, r)
	} else {
		r := &global.LoginTocSelect{Succ: false, Reason: &global.PMsg{MsgID: 1}}
		SendPack(state, r)
	}
}

func doLoginCreate(state *global.TcpClientState, ctx *kernel.Context, m *global.LoginTosCreateRole) {
	if state.RoleID > 0 || state.Account == "" {
		r := &global.LoginTocCreateRole{Succ: false, Reason: &global.PMsg{MsgID: 32}}
		SendPack(state, r)
		return
	}
	ok, rs := ctx.CallName(global.SYS_ACCOUNT_SERVER,
		&global.CreateRole{
			AgentID:  state.AgentID,
			Account:  state.Account,
			ServerID: state.ServerID,
			Name:     m.Name,
			HeroType: m.HeroType,
			Sex:      m.Sex,
		})
	if !ok{
		r := &global.LoginTocCreateRole{Succ: false, Reason: lib.ErrToPMsg(rs)}
		SendPack(state, r)
		return
	}
	switch msg :=rs.(type) {
	case *global.PMsg:
		r := &global.LoginTocCreateRole{Succ: false, Reason: msg}
		SendPack(state, r)
	case *global.CreateRoleResult:
		state.RoleID = msg.RoleID
		kickRoles(ctx,msg.Roles,msg.RoleID)
		playerPid := player.Start(ctx.Self(), state.Conn, state.RoleID)
		if playerPid != nil {
			state.Roles = nil // gc
			state.Player = playerPid
			r := &global.LoginTocCreateRole{Succ: true, RoleID: state.RoleID}
			SendPack(state, r)
		} else {
			r := &global.LoginTocCreateRole{Succ: false, Reason: &global.PMsg{MsgID: 1}}
			SendPack(state, r)
		}
	}
}

func syncPack(state *global.TcpClientState, id int64) {
	id++
	c := state.Conn
	for id <= state.SendID {
		pack := state.Cache[id%CacheSize]
		_ = c.SendBufHead(pack)
		id++
	}
}

func kickRoles(ctx *kernel.Context,roles []int64,roleID int64)  {
	for _,rid := range roles {
		if rid != roleID {
			if pid := lib.GetRolePid(rid);pid != nil{
				ctx.Call(pid,global.Kick{})
			}
		}
	}
}