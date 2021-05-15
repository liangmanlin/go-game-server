package player

import (
	"fmt"
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
	"reflect"
	"runtime/debug"
	"strings"
)

func checkTransaction(player *global.Player) {
	if !player.IsTransaction {
		panic(&global.PMsg{})
	}
}

func Transaction(player *global.Player, f func() interface{}) (rs TResult) {
	if player.IsTransaction {
		panic(fmt.Errorf("Transaction nesting "))
	}
	player.IsTransaction = true
	defer transactionCatch(player, &rs)
	rs.Result = f()
	rs.OK = true
	commit(player)
	return
}

func Backup(player *global.Player, backupKey global.BackupKey, key, value interface{}) {
	if _, ok := player.BackupMap[backupKey]; !ok {
		data := global.BackData{Key: key, Value: value}
		player.Backup = append(player.Backup, data)
		player.BackupMap[backupKey] = len(player.Backup) - 1
	}
}

func Abort(m interface{}) {
	panic(m)
}

func AddDBQueue(player *global.Player, fun func(player *global.Player,op int32,arg interface{}), op int32, arg interface{}) {
	player.DBQueue = append(player.DBQueue, global.DBQueue{OP: op, Fun: fun, Arg: arg})
}

func transactionCatch(player *global.Player, rs *TResult) {
	player.IsTransaction = false
	if p := recover(); p != nil {
		// 捕捉到错误
		rs.OK = false
		switch p.(type) {
		case *global.PMsg:
			rs.Result = p
		default:
			kernel.ErrorLog("catch error:%s,Stack:%s", p, debug.Stack())
			rs.Result = &global.PMsg{MsgID: 1}
		}
		// 回滚
		rollback(player)
	}
}

func commit(player *global.Player)  {
	if len(player.DBQueue) > 0 {
		for _,v := range player.DBQueue {
			v.Fun(player,v.OP,v.Arg)
		}
		player.DBQueue = nil
	}
	if len(player.BackupMap) > 0 {
		player.BackupMap = make(map[global.BackupKey]int)
	}
	if len(player.Backup) > 0{
		player.Backup = nil
	}
}

func rollback(player *global.Player) {
	if len(player.DBQueue) > 0 {
		player.DBQueue = nil
	}
	if len(player.BackupMap) > 0 {
		player.BackupMap = make(map[global.BackupKey]int)
	}
	// 先进后出，从尾部开始回滚
	if l := len(player.Backup); l > 0 {
		for i := l - 1; i >= 0; i-- {
			d := player.Backup[i]
			switch k := d.Key.(type) {
			case string:
				// 如果是字符串，是可以直接赋值的类型，用反射处理
				rollbackValue(player, k, d.Value)
			default:
				// 余下的一定是函数
				rollbackFun(player, d)
			}
		}
	}
}

func rollbackValue(player *global.Player, k string, v interface{}) {
	arr := strings.Split(k, ".")
	vrt := reflect.ValueOf(v)
	root := reflect.ValueOf(player).Elem()
	l := len(arr) - 1
	for i := 0; i < l; i++ {
		root = root.FieldByName(arr[i])
		if root.Kind() == reflect.Ptr{
			root = root.Elem()
		}
	}
	vf := root.FieldByName(arr[l])
	vf.Set(vrt)
}

func rollbackFun(player *global.Player,data global.BackData)  {
	fun := reflect.ValueOf(data.Key)
	fun.Call([]reflect.Value{reflect.ValueOf(player),reflect.ValueOf(data.Value)})
}
