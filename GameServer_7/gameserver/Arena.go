package gameserver

import (
	"math/rand"
)

const ARENA_SIZE = 6 // Размер арены - сколько на сколь ячеек

type Arena struct {
	Platforms [ARENA_SIZE][ARENA_SIZE]*Platform
}

func NewArena(infos map[string]*PlatformInfo) *Arena {
	arena := &Arena{}

	for y := int16(0); y < ARENA_SIZE; y++ {
		for x := int16(0); x < ARENA_SIZE; x++ {
            if arena.Platforms[y][x] == nil {
                platform := makePlatform(infos, arena, x, y)
                arena.Platforms[y][x] = platform
            }
		}
	}

	return arena
}

func makePlatform(infos map[string]*PlatformInfo, arena *Arena, x, y int16) *Platform {
	// Дергаем рандомную платформу
	var info *PlatformInfo = nil
	randomIndex := rand.Int() % len(infos)
	i := 0
	for key := range infos {
		i++
		if i == randomIndex {
			info = infos[key]
			break
		}
	}
	if info == nil {
		return nil
	}

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
