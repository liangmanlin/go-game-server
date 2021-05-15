package gw

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/liangmanlin/gootp/gate/pb"
	"github.com/liangmanlin/gootp/rand"
	"strconv"
	"sync/atomic"
)

var idx int32

func init()  {
	idx = rand.New().Random(1,2000)
}

func RandomToken() string {
	id := atomic.AddInt32(&idx,1)
	buf := md5.Sum(strconv.AppendInt(nil,int64(id+1024),10))
	pb.WriteIn32(buf[:],id,0)
	token := hex.EncodeToString(buf[:])
	if _,ok := tokenMap.Load(token);ok{
		return RandomToken()
	}
	return token
}
