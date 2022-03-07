package lib

import (
	"fmt"
	"game/config"
	"game/global"
	"github.com/liangmanlin/gootp/kernel"
	"github.com/liangmanlin/gootp/node"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

// 公共库，这里的代码不能依赖其他包，仅仅可以依赖 global 包

var roleIDMap sync.Map

func IsRoleOnline(roleID int64) bool {
	_, ok := roleIDMap.Load(roleID)
	return ok
}

func GetRolePid(roleID int64) *kernel.Pid {
	if pid, ok := roleIDMap.Load(roleID); ok {
		return pid.(*kernel.Pid)
	}
	return nil
}

func SetRolePid(roleID int64, pid *kernel.Pid) {
	roleIDMap.Store(roleID, pid)
}

func DelRolePid(roleID int64) {
	roleIDMap.Delete(roleID)
}

func GetPlayerName(roleID int64) string {
	return "player_" + strconv.FormatInt(roleID, 10)
}

func ErrToPMsg(err interface{}) *global.PMsg {
	switch r := err.(type) {
	case *global.PMsg:
		return r
	default:
		_, file, line, ok := runtime.Caller(1)
		if !ok {
			file = "???"
			line = 0
		} else {
			file = filepath.Base(file)
		}
		kernel.ErrorLog("[%s:%d] error:%#v,", file, line, err)
		return &global.PMsg{MsgID: 1}
	}
}

var _serverRoot string

func GetServerRoot() string {
	return _serverRoot
}

func GetServerID() int32 {
	return int32(config.Server.GetListInt("server_id",0))
}

func NormalMapName(mapID int32) string {
	return "normal_" + strconv.FormatInt(int64(mapID), 10)
}

var mapNamePid sync.Map

func RegisterMap(name string, pid *kernel.Pid) {
	mapNamePid.Store(name, pid)
}
func UnRegisterMap(name string) {
	mapNamePid.Delete(name)
}

func GetMapPid(name string) *kernel.Pid {
	if pid, ok := mapNamePid.Load(name); ok {
		return pid.(*kernel.Pid)
	}
	return nil
}

func XYLessThan(x, y, x2, y2, dist float32) bool {
	x2 = x2 - x
	y2 = y2 - y
	return x2*x2+y2*y2 <= dist
}

func PosLessThan(pos1, pos2 *global.PPos, dist float32) bool {
	return XYLessThan(pos1.X, pos1.Y, pos2.X, pos2.Y, dist)
}

func CalcDir(pos1, pos2 *global.PPos) int32 {
	r := math.Atan2(float64(pos2.Y-pos1.Y), float64(pos2.X-pos1.X))
	dir := int32(math.Round(r * 180 / math.Pi))
	if dir < 0 {
		dir += 360
	}
	return dir
}

func Dir(dx, dy float64) int16 {
	r := math.Atan2(dy, dx)
	dir := int16(math.Round(r * 180 / math.Pi))
	if dir < 0 {
		dir += 360
	}
	return dir
}

func DirToTan(dir int32) float64 {
	return float64(dir) / 180 * math.Pi
}

// TODO 使用unsafe会更快
func ToProp(from interface{}) *global.PProp {
	ft := reflect.ValueOf(from)
	if ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}
	size := ft.NumField()
	rs := global.PProp{}
	ptr := reflect.New(reflect.TypeOf(rs))
	rt := ptr.Elem()
	for i := 1; i < size; i++ {
		v := rt.Field(i - 1)
		d := ft.Field(i)
		v.Set(d)
	}
	return ptr.Interface().(*global.PProp)
}

// 利用反射，获取一个比较好看的数据
// 性能一般，慎用
func ToString(proto interface{}) string {
	rf := reflect.ValueOf(proto)
	return toString(rf)
}

func toString(rf reflect.Value) string {
	if rf.Kind() == reflect.Ptr {
		if rf.IsZero() {
			return "nil"
		}
		rf = rf.Elem()
	}
	switch rf.Kind() {
	case reflect.Struct:
		return structToString(rf)
	case reflect.Slice, reflect.Array:
		return sliceToString(rf)
	case reflect.Map:
		return mapToString(rf)
	case reflect.String:
		return rf.String()
	default:
		return fmt.Sprintf("%v", rf.Interface())
	}
}

func structToString(rf reflect.Value) string {
	size := rf.NumField()
	tf := rf.Type()
	var sl []string
	var vs string
	for i := 0; i < size; i++ {
		f := rf.Field(i)
		vs = toString(f)
		sl = append(sl, tf.Field(i).Name+":"+vs)
	}
	return rf.Type().Name() + "{" + strings.Join(sl, ",") + "}"
}

func sliceToString(rf reflect.Value) string {
	size := rf.Len()
	var sl []string
	for i := 0; i < size; i++ {
		sl = append(sl, toString(rf.Index(i)))
	}
	return "[" + strings.Join(sl, ",") + "]"
}

func mapToString(rf reflect.Value) string {
	ranges := rf.MapRange()
	var sl []string
	for ranges.Next() {
		k := ranges.Key()
		v := ranges.Value()
		sl = append(sl, toString(k)+":"+toString(v))
	}
	return "{" + strings.Join(sl, ",") + "}"
}

func CastMap(mapPid *kernel.Pid, mod int32, msg interface{}) {
	// 用于检测消息是否注册,正式环境不检查
	if global.DEBUG && global.CHECK_NODE_PROTO {
		rType := reflect.TypeOf(msg)
		if !node.IsProtoDef(rType) {
			kernel.ErrorLog("node proto %s not defined", rType.Name())
			return
		}
	}
	if mapPid != nil {
		kernel.Cast(mapPid, &kernel.KMsg{ModID: mod, Msg: msg})
	}
}

// 代码一样的
func CastPlayer(playerPid *kernel.Pid, mod int32, msg interface{}) {
	// 用于检测消息是否注册,正式环境不检查
	if global.DEBUG && global.CHECK_NODE_PROTO {
		rType := reflect.TypeOf(msg)
		if !node.IsProtoDef(rType) {
			kernel.ErrorLog("node proto %s not defined", rType.Name())
			return
		}
	}
	kernel.Cast(playerPid, &kernel.KMsg{ModID: mod, Msg: msg})
}

func init() {
	_serverRoot = filepath.Dir(filepath.Dir(os.Args[0]))
}
