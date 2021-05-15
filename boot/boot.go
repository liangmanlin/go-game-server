package boot

import (
	"github.com/liangmanlin/gootp/kernel"
	"log"
)

type Boot struct {
	Svr  *kernel.Actor
	Name string
}

func StartService(sup *kernel.Pid, bootList []Boot) {
	for _, s := range bootList {
		child := &kernel.SupChild{ChildType: kernel.SupChildTypeWorker, Name: s.Name, Svr: s.Svr, ReStart: true}
		err, _ := kernel.SupStartChild(sup, child)
		if err != nil {
			log.Panic(err)
		}else{
			kernel.ErrorLog("%s started",s.Name)
		}
	}
}
