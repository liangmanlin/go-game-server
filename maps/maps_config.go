package maps

import (
	"fmt"
	"game/config"
	"game/global"
	"github.com/liangmanlin/gootp/gutil"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

const (
	MAP_DATA_GRID = iota + 1
	MAP_DATA_JUMP
	MAP_DATA_MONSTER
	MAP_DATA_DYNAMIC_GRID
	MAP_DATA_BORN
	MAP_DATA_COLLECT
)

var allMaps = make(map[int32]*MapConfig, 100)

// 需要保证单线程执行
func loadConfig(mapConfigPath string) {
	cache := make(map[int32]*MapConfig)
	allMap := config.Maps.All()
	for i := range allMap {
		cfg := &allMap[i]
		blockID := cfg.BlockPath
		if c ,ok := cache[blockID];ok{
			mc := *c
			mc.MapID = cfg.Id
			allMaps[cfg.Id] = &mc
			continue
		}
		file := mapConfigPath+"/"+strconv.FormatInt(int64(blockID),10)
		fi, err := os.Stat(file)
		if err != nil || fi.IsDir() {
			log.Panic(fmt.Errorf("map %d block:%d not found",cfg.Id,blockID))
		}
		cache[blockID] = load(file,cfg.Id)
	}
	var maxWidth, maxHeight int32
	for _, v := range allMaps {
		maxWidth = gutil.MaxInt32(v.Width, maxWidth)
		maxHeight = gutil.MaxInt32(v.Height, maxHeight)
	}
	buildAreasCache(maxWidth, maxHeight)
}

func load(file string,mapID int32) *MapConfig {
	buf, err := ioutil.ReadFile(file)
	if err != nil {
		log.Panic(err)
	}
	var idx int
	// TODO 暂行先弄一个老项目的测试
	idx = 3 // 前三位已经废弃
	mapC := &MapConfig{MapID: mapID}
	idx, v := readInt16(idx, buf)
	mapC.Width = int32(v)
	idx, v = readInt16(idx, buf)
	mapC.Height = int32(v)
	mapC.PosList = make([]MapPos, int(mapC.Width+1)*int(mapC.Height+1))
	size := len(buf)
	var t int8
	var bSize int32
	for idx < size {
		idx, t = readInt8(idx, buf)
		switch t {
		case MAP_DATA_GRID:
			idx = readGrid(mapC, idx, buf)
		case MAP_DATA_BORN:
			idx, mapC.BornPos = readBorn(idx, buf)
		case MAP_DATA_MONSTER:
			idx = readMonster(mapC, idx, buf)
		default:
			// 其他暂时不需要
			idx, bSize = readInt32(idx, buf)
			idx += int(bSize)
		}
	}
	allMaps[mapID] = mapC
	return mapC
}

func GetMapConfig(mapID int32) *MapConfig {
	return allMaps[mapID]
}

func readInt32(idx int, buf []byte) (int, int32) {
	v := int32(buf[idx])<<24 + int32(buf[idx+1])<<16 + int32(buf[idx+2])<<8 + int32(buf[idx+3])
	return idx + 4, v
}

func readInt16(idx int, buf []byte) (int, int16) {
	v := int16(buf[idx])<<8 + int16(buf[idx+1])
	return idx + 2, v
}

func readInt8(idx int, buf []byte) (int, int8) {
	return idx + 1, int8(buf[idx])
}

func readGrid(config *MapConfig, idx int, buf []byte) int {
	idx, v := readInt32(idx, buf)
	end := idx + int(v)
	var x, y, g int16
	var t int8
	for idx < end {
		idx, x = readInt16(idx, buf)
		idx, y = readInt16(idx, buf)
		idx, t = readInt8(idx, buf)
		g = 1
		if t > 10 { // 如果是边缘，增加g值，使得寻路尽量避过边缘
			g = 3
		}
		config.PosList[config.GridIdx(int32(x), int32(y))] = MapPos{X: x, Y: y, Type: int16(t) % 10, G: g} // 大于10的是边缘
		config.PosCount++
	}
	return idx
}

func readMonster(config *MapConfig, idx int, buf []byte) int {
	idx, v := readInt32(idx, buf)
	count := v / 8
	config.Monsters = make([]MapConfigMonster, count)
	for i := int32(0); i < count; i++ {
		m := &config.Monsters[i]
		idx, m.X = readInt16(idx, buf)
		idx, m.Y = readInt16(idx, buf)
		idx, m.TypeID = readInt32(idx, buf)
	}
	return idx
}

func readBorn(idx int, buf []byte) (int, Area) {
	idx, _ = readInt32(idx, buf)
	idx, x := readInt16(idx, buf)
	idx, y := readInt16(idx, buf)
	return idx, Area{x, y}
}

func (m *MapConfig) PosWalkAble(pos *global.PPos) bool {
	return m.XYI32WalkAble(gutil.Round(pos.X),gutil.Round(pos.Y))
}

func (m *MapConfig) XYWalkAble(x, y float32) bool {
	return m.XYI32WalkAble(gutil.Round(x), gutil.Round(y))
}

func (m *MapConfig) XYI32WalkAble(x, y int32) bool {
	if x >= 0 && x <= m.Width && y >= 0 && y <= m.Height {
		idx := m.GridIdx(x, y)
		return m.PosList[idx].Type > 0
	}
	return false
}

func (m *MapConfig) XYI32WalkAbleBorder(x, y int32, c int) int {
	if x >= 0 && x <= m.Width && y >= 0 && y <= m.Height {
		idx := m.GridIdx(x, y)
		grid := m.PosList[idx]
		if grid.Type > 0 {
			if grid.G > 1 {
				c++
			}
			return c
		}
		return 100
	}
	return 100
}

func(m *MapConfig) GetWidth() int32 {
	return m.Width
}

func(m *MapConfig) GetHeight() int32 {
	return m.Height
}

func (m *MapConfig)GridType(gridIdx int32) (gridType int16, g int16) {
	grid := m.PosList[gridIdx]
	return grid.Type,grid.G
}

func (m *MapConfig) PosSafe(pos *global.PPos)bool  {
	x := gutil.Round(pos.X)
	y := gutil.Round(pos.Y)
	if x >= 0 && x <= m.Width && y >= 0 && y <= m.Height {
		idx := m.GridIdx(x, y)
		return m.PosList[idx].Type == 2
	}
	return false
}

func (m *MapConfig) GridIdx(x, y int32) int {
	return int(m.Width*y + x)
}

