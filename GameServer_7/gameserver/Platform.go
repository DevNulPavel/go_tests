package gameserver

import (
	"math/rand"
)

const (
	PLATFORM_SIDE_SIZE      = 24                          // вся платформа с учетом мостов
	PLATFORM_WORK_SIZE      = 18                          // платформа без мостов
	PLATFORM_BLOCK_SIZE_1x1 = 3                           // Platform Block Size 1х1
	PLATFORM_BLOCK_SIZE_2x2 = PLATFORM_BLOCK_SIZE_1x1 * 2 // Platform Block Size 2х2
)

type PlatformMonster struct {
	Name string
	X    int16
	Y    int16
}

type PlatformObject struct {
	Id  string
	X   int16
	Y   int16
	Rot int8
}

type Platform struct {
	Info *PlatformInfo
	// Pos and size
	PosX   int16
	PosY   int16
	Width  uint16
	Height uint16
	// Enter
	EnterX   int8
	EnterY   int8
	EnterDir PlatformDir
	// Exit
	ExitCoord [4]int8
	ExitDir   PlatformDir
	// Cells
	Cells []PlatformCellType
	// Items
	Monsters []PlatformMonster // TODO: ???
	Objects  []PlatformObject  // TODO: ???
	Blocks   []PlatformObject  // TODO: ???
}

func NewPlatform(info *PlatformInfo, posX, posY int16, exits [4]int8) *Platform {
	platform := &Platform{}

	// Info
	platform.Info = info

	// Pos and size
	platform.PosX = posX
	platform.PosY = posY
	platform.Width = info.Width
	platform.Height = info.Height

	// Exit and enter
	platform.ExitCoord = exits
	for i := range platform.ExitCoord {
		if platform.ExitCoord[i] != -1 {
			dir := PlatformDir(i)
			x, y := getExitCoord(dir, exits)

			platform.EnterX = x
			platform.EnterY = y
			platform.EnterDir = dir
		}
	}

	// Cells and walls
	createCells(platform)

	return platform
}

func getExitCoord(dir PlatformDir, exit [4]int8) (x, y int8) {
	switch {
	case (dir == DIR_NORTH) && (exit[DIR_NORTH] != -1):
		return exit[DIR_NORTH], 0

	case (dir == DIR_EAST) && (exit[DIR_EAST] != -1):
		return PLATFORM_SIDE_SIZE - 1, exit[DIR_EAST]

	case (dir == DIR_SOUTH) && (exit[DIR_SOUTH] != -1):
		return exit[DIR_SOUTH], PLATFORM_SIDE_SIZE - 1

	case (dir == DIR_WEST) && (exit[DIR_WEST] != -1):
		return 0, exit[DIR_WEST]

	default:
		return -1, -1
	}

	return -1, -1
}

func createCells(platform *Platform) {
	w := platform.Width
	h := platform.Height
	ww := PLATFORM_WORK_SIZE
	hh := PLATFORM_WORK_SIZE

	// Blocks
	block1x1 := make([]*PlatformObjectInfo, 0)
	block2x2 := make([]*PlatformObjectInfo, 0)
	for i := range platform.Info.Blocks {
		obj := &platform.Info.Blocks[i]
		if (obj.Width == PLATFORM_BLOCK_SIZE_2x2) && (obj.Height == PLATFORM_BLOCK_SIZE_2x2) {
			block2x2 = append(block2x2, obj)
		} else if (obj.Width == PLATFORM_BLOCK_SIZE_1x1) && (obj.Height == PLATFORM_BLOCK_SIZE_1x1) {
			block1x1 = append(block1x1, obj)
		}
	}

	cellsInfo := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}
	cellsWalls := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}

	for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
		cellsInfo[i] = CELL_TYPE_BLOCK
		cellsWalls[i] = CELL_TYPE_UNDEF
	}

	switch platform.Info.Type {
	case PLATFORM_INFO_TYPE_BRIDGE:
		for y := uint16(0); y < h; y += PLATFORM_BLOCK_SIZE_1x1 {
			for x := uint16(0); x < w; x += PLATFORM_BLOCK_SIZE_1x1 {
				haveBlock := false
				for yy := uint16(0); yy < PLATFORM_BLOCK_SIZE_1x1; yy++ {
					for xx := uint16(0); xx < PLATFORM_BLOCK_SIZE_1x1; xx++ {
						// Index
						cellsIndex := (y+yy)*w + (x + xx)
						infoCellValue := platform.Info.Cells[cellsIndex]
						// Update cells
						cellsInfo[cellsIndex] = infoCellValue
						if infoCellValue != CELL_TYPE_BLOCK {
							haveBlock = true
						}
					}
				}

				if haveBlock {
					platform.Objects = appendObjects(platform.Objects, block1x1,
						int16(x), int16(y),
						int8((x+y)&3), 3)
				}
			}
		}

		// Make cells
		cellsCount := w * h
		platform.Cells = make([]PlatformCellType, cellsCount)
		for i := range platform.Cells {
			platform.Cells[i] = CELL_TYPE_BLOCK
		}
		for i := uint16(0); i < cellsCount; i++ {
			platform.Cells[i] = cellsInfo[i]
		}
	case PLATFORM_INFO_TYPE_BATTLE:

	}
}

func appendObjects(container []PlatformObject, objects []*PlatformObjectInfo, x, y int16, rot int8, size int16) []PlatformObject {
	if len(objects) == 0 {
		return container
	}

	var selectedItem *PlatformObjectInfo = nil

	// Random probability
	sumProb := 0
	for i := range objects {
		sumProb += int(objects[i].Probability * 100)
	}
	randVal := rand.Int() % sumProb

	// Select random item
	variant := 0
	for i := range objects {
		selected := (variant <= randVal) && (variant+int(objects[i].Probability*100) > randVal)
		if selected {
			selectedItem = objects[i]
			break
		} else {
			variant += int(objects[i].Probability * 100)
		}
	}

	if selectedItem == nil {
		return container
	}

	// Max size
	maxInt16 := func(a, b int16) int16 {
		if a > b {
			return a
		}
		return b
	}
	size = maxInt16(maxInt16(selectedItem.Width, selectedItem.Height), size)

	// Position
	if rot == 1 {
		y += size
	} else if rot == 2 {
		x += size
		y += size
	} else if rot == 3 {
		x += size
	}

	// Append
	object := PlatformObject{
		Id:  selectedItem.Id,
		X:   x,
		Y:   y,
		Rot: rot,
	}

	container = append(container, object)
	return container
}
