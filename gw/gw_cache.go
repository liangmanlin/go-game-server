package gw

import (
	"github.com/liangmanlin/gootp/kernel"
	"sync"
)

var CacheSize int64= 300

var tokenMap sync.Map

func InsertToken(token string, pid *kernel.Pid) {
	tokenMap.Store(token, pid)
}

func TokenToPid(token string) *kernel.Pid {
	if pid, ok := tokenMap.Load(token); ok {
		return pid.(*kernel.Pid)
	}
	return nil
}
