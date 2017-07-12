package gameserver

import (
	"log"
	"math/rand"
    "encoding/json"
)

const ARENA_SIZE = 6 // Размер арены - сколько на сколь ячеек

type Arena struct {
	Platforms [ARENA_SIZE][ARENA_SIZE]*Platform `json:"platforms"`
}

func NewArena(infos []*PlatformInfo) *Arena {
	arena := &Arena{}

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
			platform := makePlatform(battlePlatforms, arena, x, y)
			arena.Platforms[y][x] = platform
			log.Printf("Made platform %dx%d\n", y, x)
		}
	}

	return arena
}

func (arena *Arena) ToJsonData() ([]byte, error)  {
    return json.Marshal(arena)
}

// TODO: ???
func makePlatform(infos []*PlatformInfo, arena *Arena, x, y int16) *Platform {
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
			exitCoord[0] = arena.Platforms[y-1][x].ExitCoord[2]
		} else {
			exitCoord[0] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[0] = -1
	}
	// east
	if x < ARENA_SIZE-1 {
		if arena.Platforms[y][x+1] != nil {
			exitCoord[1] = arena.Platforms[y][x+1].ExitCoord[3]
		} else {
			exitCoord[1] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[1] = -1
	}
	// south
	if y < ARENA_SIZE-1 {
		if arena.Platforms[y+1][x] != nil {
			exitCoord[2] = arena.Platforms[y+1][x].ExitCoord[0]
		} else {
			exitCoord[2] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[2] = -1
	}
	// west
	if x > 0 {
		if arena.Platforms[y][x-1] != nil {
			exitCoord[3] = arena.Platforms[y][x-1].ExitCoord[1]
		} else {
			exitCoord[3] = int16(rand.Int()%((PLATFORM_SIDE_SIZE-6-5)/3)*3 + 3 + 1)
		}
	} else {
		exitCoord[3] = -1
	}

	platform := NewPlatform(info, x, y, exitCoord, false)
	return platform
}
