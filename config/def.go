package config

import (
	"encoding/json"
	"sync/atomic"
	"unsafe"
)
// 自动生成，请勿随便修改
func init()  {
	//init-ptr-start
	PtrMap["Goods"] = Goods
	PtrMap["BossWorld"] = BossWorld
	PtrMap["Effect"] = Effect
	PtrMap["Server"] = Server
}

type KV struct {
	Key int
	Value int
}

//Effect-start
type DefEffect struct {
	Id string
	Scale int32
	Prefab string
	Type int32
	S []*KV
}
type effectConfig struct{
	m map[string]int
	arr []DefEffect
}
var Effect = &effectConfig{}
func (E effectConfig)Get(key...interface{})*DefEffect{
	return &E.arr[E.m[sliceToString(key)]]
}
func (E effectConfig)load(path string)  {
	c := &effectConfig{m:make(map[string]int)}
	if err:= json.Unmarshal(readFile("Effect",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Effect)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (E effectConfig)All()[]DefEffect {
	return E.arr
}
//Effect-end
//BossWorld-start
type DefBossWorld struct {
	Type_id int32
	Map_id int32
	Dec_lv int32
	Refresh_time int32
	X int32
	Y int32
	Tire int32
	Sp_drop []int
}
type bossWorldConfig struct{
	m map[int32]int
	arr []DefBossWorld
}
var BossWorld = &bossWorldConfig{}
func (B bossWorldConfig)Get(key int32)*DefBossWorld{
	return &B.arr[B.m[key]]
}
func (B bossWorldConfig)load(path string)  {
	c := &bossWorldConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("BossWorld",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Type_id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&BossWorld)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (B bossWorldConfig)All()[]DefBossWorld {
	return B.arr
}
//BossWorld-end
//Goods-start
type DefGoods struct {
	Type_id int32
	Name string
	Color int32
	Pile_num int32
}
type goodsConfig struct{
	m map[int32]int
	arr []DefGoods
}
var Goods = &goodsConfig{}
func (G goodsConfig)Get(key int32)*DefGoods{
	return &G.arr[G.m[key]]
}
func (G goodsConfig)load(path string)  {
	c := &goodsConfig{m:make(map[int32]int)}
	if err:= json.Unmarshal(readFile("Goods",path), &c.arr);err != nil{panic(err)}
	for i:=0;i<len(c.arr);i++{
		c.m[c.arr[i].Type_id] = i
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Goods)),*(*unsafe.Pointer)(unsafe.Pointer(&c)))
}
func (G goodsConfig)All()[]DefGoods {
	return G.arr
}
//Goods-end
