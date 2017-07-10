package gameserver

import (
	"encoding/json"
	"io"
)

type PlatformInfo struct {
	SymbolName    string           `json:"symbol_name"`        // имя символа
	Width         uint32           `json:"width"`              // ширина
	Height        uint32           `json:"height"`             // высота
	Cells         []int8           `json:"cells,omitempty"`    // ячейки в платформе, размер - ширина X высота
	Exits         [4]int8          `json:"exits"`              // выходы из платформы
	MonstersNames []string         `json:"monsters"`           // имена монстров на платформе
	SpawnMin      uint8            `json:"monsters_spawn_min"` // минимальное количество монстров
	SpawnMax      uint8            `json:"monsters_spawn_max"` // максимальное количество монстров
	Objects       []PlatformObject `json:"objects,omitempty"`
	Blocks        []PlatformObject `json:"blocks,omitempty"`
}

func NewPlatformsFromData(data []byte) (map[string]PlatformInfo, error) {
	result := map[string]PlatformInfo{}
	err := json.Unmarshal(data, &result)
	return result, err
}

func NewPlatformsFromReader(reader io.Reader) (map[string]PlatformInfo, error) {
	result := map[string]PlatformInfo{}
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&result)
	return result, err
}
