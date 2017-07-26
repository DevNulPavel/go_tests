package gameserver

import (
	"encoding/json"
	"log"
	"math/rand"
)

const ARENA_SIZE = 2 // Размер арены - сколько на сколь ячеек

type ArenaModel struct {
	Type      string                            `json:"type"`
	Platforms [ARENA_SIZE][ARENA_SIZE]*Platform `json:"platforms"`
}

func NewArenaModel(infos []*PlatformInfo) ArenaModel {
	arena := ArenaModel{}

	arena.Type = "ArenaInfo"

	// Platforms
	bridgePlatforms := make([]*PlatformInfo, 0)
	battlePlatforms := make([]*PlatformInfo, 0)
	for _, value := range infos {
		switch value.Type {
		case PLATFORM_INFO_TYPE_BRIDGE:
			bridgePlatforms = append(bridgePlatforms, value)
		case PLATFORM_INFO_TYPE_BATTLE:
			battlePlatforms = append(battlePlatforms, value)
		}
	}

	for y := int16(0); y < ARENA_SIZE; y++ {
		for x := int16(0); x < ARENA_SIZE; x++ {
			platform := makePlatform(battlePlatforms, &arena, x, y)
			arena.Platforms[y][x] = platform
			log.Printf("Made platform %dx%d\n", y, x)
		}
	}

	return arena
}

func (arena *ArenaModel) ToBytes() ([]byte, error) {
	jsonData, err := json.Marshal(arena)
	if err != nil {
		return []byte{}, err
	}
	return jsonData, nil
}

// TODO: ???
func makePlatform(infos []*PlatformInfo, arena *ArenaModel, x, y int16) *Platform {
	if len(infos) == 0 {
		return nil
	}

	// Дергаем рандомную платформу
	randomIndex := rand.Int() % len(infos)
	info := infos[randomIndex]

	exitCoord := [4]int16{}

	// north
	if y > 0 {
		if arena.Platforms[y-1][x] != nil {
			exitCoord[DIR_NORTH] = arena.Platforms[y-1][x].ExitCoord[DIR_SOUTH]
		} else {
			exitCoord[DIR_NORTH] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[DIR_NORTH] = -1
	}
	// east
	if x < ARENA_SIZE-1 {
		if arena.Platforms[y][x+1] != nil {
			exitCoord[DIR_EAST] = arena.Platforms[y][x+1].ExitCoord[DIR_WEST]
		} else {
			exitCoord[DIR_EAST] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[DIR_EAST] = -1
	}
	// south
	if y < ARENA_SIZE-1 {
		if arena.Platforms[y+1][x] != nil {
			exitCoord[DIR_SOUTH] = arena.Platforms[y+1][x].ExitCoord[DIR_NORTH]
		} else {
			exitCoord[DIR_SOUTH] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[DIR_SOUTH] = -1
	}
	// west
	if x > 0 {
		if arena.Platforms[y][x-1] != nil {
			exitCoord[DIR_WEST] = arena.Platforms[y][x-1].ExitCoord[DIR_EAST]
		} else {
			exitCoord[DIR_WEST] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[DIR_WEST] = -1
	}

	platform := NewPlatform(info, x * PLATFORM_SIDE_SIZE, y * PLATFORM_SIDE_SIZE, exitCoord, false)
	return platform
}
