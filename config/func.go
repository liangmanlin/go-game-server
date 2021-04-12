package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync/atomic"
	"unsafe"
)

func sliceToString(keys []interface{}) string {
	if len(keys) == 1 {
		switch v2 := keys[0].(type) {
		case string:
			return v2
		case int8, int16, int, int32, int64, uint8, uint16, uint32, uint64:
			return fmt.Sprintf("%d", v2)
		}
		return ""
	}
	var l []string
	for _, v := range keys {
		switch v2 := v.(type) {
		case string:
			l = append(l, v2)
		case int8, int16, int, int32, int64, uint8, uint16, uint32, uint64:
			l = append(l, fmt.Sprintf("%d", v2))
		}
	}
	return fmt.Sprintf("{%s}", strings.Join(l, ","))
}

// 加载所有配置
func LoadAll(path string) {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, f := range fs {
		if !f.IsDir() {
			ext := filepath.Ext(f.Name())
			if ext == ".json" {
				baseName := strings.TrimSuffix(f.Name(), ext)
				if p, ok := PtrMap[baseName]; ok {
					p.load(path)
				}
			}
		} else {
			LoadAll(path + "/" + f.Name())
		}
	}
}

func Load(name, path string) bool {
	if f, ok := PtrMap[name]; ok {
		f.load(path)
		return true
	}
	return false
}

func readFile(name, path string) []byte {
	file := fmt.Sprintf("%s/%s.json", path, name)
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return buf
}

var PtrMap = make(map[string]configFunc)

type configFunc interface {
	load(path string)
}

type ServerMapType map[string]interface{}

var Server ServerMapType

func (S ServerMapType) Get(key string) interface{} {
	return S[key]
}
func (S ServerMapType) GetListStr(key string, idx int) string {
	if c, ok := S[key]; ok {
		return c.([]interface{})[idx].(string)
	}
	return ""
}
func (S ServerMapType) GetListInt(key string, idx int) int {
	if c, ok := S[key]; ok {
		return int(c.([]interface{})[idx].(float64))
	}
	return 0
}
func (S ServerMapType) GetStr(key string) string {
	if c, ok := S[key]; ok {
		return c.(string)
	}
	return ""
}
func (S ServerMapType) GetInt(key string) int {
	if c, ok := S[key]; ok {
		return int(c.(float64))
	}
	return 0
}
func (S ServerMapType) load(path string) {
	s := make(map[string]interface{})
	if err := json.Unmarshal(readFile("Server", path), &s); err != nil {
		panic(err)
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&Server)), *(*unsafe.Pointer)(unsafe.Pointer(&s)))
}
