package service

import (
	"fmt"
	"game/global"
	"game/lib"
	"github.com/liangmanlin/gootp/db"
	"github.com/liangmanlin/gootp/kernel"
)

type as struct {
	AgentID  int32
	ServerID int32
}

type AccountState struct {
	Dirty map[as]bool
	IDMap map[as]int32
}

var AccountActor = &kernel.Actor{
	Init: func(ctx *kernel.Context, pid *kernel.Pid, args ...interface{}) interface{} {
		// 查询数据库
		rs := lib.GameDB.SyncSelect(kernel.Call,global.TABLE_ROLE_ID_INDEX,1)
		m := make(map[as]int32,len(rs))
		for _,v := range rs{
			v2 := v.(*global.RoleIDIndex)
			m[as{v2.AgentID,v2.ServerID}] = v2.Index
		}
		state := AccountState{IDMap: m,Dirty: make(map[as]bool)}
		kernel.SendAfterForever(pid,60*kernel.Millisecond,kernel.Loop{})
		return &state
	},
	HandleCast: func(ctx *kernel.Context, msg interface{}) {
		switch msg.(type) {
		case kernel.Loop:
			dump(ctx.State.(*AccountState))
		}
	},
	HandleCall: func(ctx *kernel.Context, request interface{}) interface{} {
		switch r := request.(type) {
		case *global.CreateRole: // 创建角色
			return createRole(ctx.State.(*AccountState),r)
		}
		return nil
	},
	Terminate: func(ctx *kernel.Context, reason *kernel.Terminate) {

	},
	ErrorHandler: func(ctx *kernel.Context, err interface{}) bool {
		return true
	},
}

func createRole(state *AccountState,m *global.CreateRole) interface{} {
	rs := lib.GameDB.ModSelect(global.TABLE_ROLE_BASE, []string{"RoleID"}, fmt.Sprintf("Account=%s and AgentID=%s and ServerID=%s",
		db.Encode(m.Account), db.Encode(m.AgentID), db.Encode(m.ServerID)))
	roles := make([]int64, 0, len(rs))
	for _, v := range rs {
		roles = append(roles, v[0].(int64))
	}
	if len(roles) >= 3 {
		return &global.PMsg{MsgID: 32}
	}
	roleID := makeRoleID(state,m.AgentID,m.ServerID)
	Attr := &global.PRoleAttr{RoleID: roleID}
	Base := &global.PRoleBase{RoleID: roleID,AgentID: m.AgentID,ServerID: m.ServerID,HeroType: m.HeroType}
	// 由于这个变量不会被修改了所以不需要拷贝
	_,err := lib.GameDB.ModInsert(global.TABLE_ROLE_ATTR,Attr)
	if err != nil {
		return lib.NewPMsg(32)
	}
	_,err = lib.GameDB.ModInsert(global.TABLE_ROLE_BASE,Base)
	if err != nil {
		return lib.NewPMsg(32)
	}
	lib.GameDB.ModUpdateFields(global.TABLE_ACCOUNT,[]string{"LastRole"},[]interface{}{roleID},
		fmt.Sprintf("Account=%s and AgengID=%d and ServerID=%d",db.Encode(m.Account),m.AgentID,m.ServerID))

	return &global.CreateRoleResult{RoleID: roleID,Roles: roles}
}

func makeRoleID(state *AccountState,agentID,serverID int32) int64  {
	key := as{AgentID: agentID,ServerID: serverID}
	if i,ok := state.IDMap[key];ok{
		state.Dirty[key] = true
		i++
		state.IDMap[key] = i
		return global.ROLEID_AGENT*int64(agentID)+global.ROLEID_SERVER*int64(serverID)+int64(i)
	}
	state.IDMap[key] = 1
	lib.GameDB.ModInsert(global.TABLE_ROLE_ID_INDEX,
		&global.RoleIDIndex{AgentID: agentID,ServerID: serverID,Index: 1})
	return global.ROLEID_AGENT*int64(agentID)+global.ROLEID_SERVER*int64(serverID)+1
}

func dump(state *AccountState)  {
	for k := range state.Dirty {
		lib.GameDB.SyncUpdate(global.TABLE_ROLE_ID_INDEX,1,
			&global.RoleIDIndex{AgentID: k.AgentID,ServerID: k.ServerID,Index: state.IDMap[k]})
		delete(state.Dirty,k)
	}
}
