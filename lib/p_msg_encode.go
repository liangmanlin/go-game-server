package lib

import (
	"game/global"
	"github.com/liangmanlin/gootp/gate/pb"
)

func NewPMsg(msgID int32) *global.PMsg {
	return &global.PMsg{MsgID: msgID}
}

func NewPMsgParam(msgID int32, param ...[]byte) *global.PMsg {
	return &global.PMsg{MsgID: msgID, Bin: appendPMsgByte(param)}
}

func appendPMsgByte(param [][]byte) []byte {
	if len(param) == 0{
		return nil
	}
	var size int
	for _,v := range param{
		size += len(v)
	}
	buf := make([]byte,2, size+2)
	for _, v := range param {
		buf = append(buf, v...)
	}
	pb.WriteSize(buf, 0, size)
	return buf
}

func M_Num(n int32) []byte {
	buf := make([]byte,1,1+4)
	buf[0] = 1
	pb.WriteIn32(buf,n,1)
	return buf
}

func M_String(str string) []byte {
	size := len(str)
	buf := make([]byte,1+2+size)
	buf[0] = 2
	pb.WriteString(buf,str,1)
	return buf
}

func M_Role(roleID int64,roleName string) []byte {
	size := len(roleName)
	buf := make([]byte,1+8+2+size)
	buf[0] = 3
	pb.WriteInt64(buf,roleID,1)
	pb.WriteString(buf,roleName,9)
	return buf
}
