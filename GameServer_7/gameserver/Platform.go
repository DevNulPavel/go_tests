package gameserver

import (
	"math/rand"
    "math"
)

const (
	PLATFORM_SIDE_SIZE      = 24 // вся платформа с учетом мостов
	PLATFORM_WORK_SIZE      = 18 // платформа без мостов
	PLATFORM_BLOCK_SIZE_3x3 = 3  // Platform Block Size 3х3
	PLATFORM_BLOCK_SIZE_6x6 = 6  // Platform Block Size 6х6
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
	EnterX   int16
	EnterY   int16
	EnterDir PlatformDir
	// Exit
	ExitCoord [4]int16
	ExitDir   PlatformDir
	// Cells
	Cells []PlatformCellType
	// Items
	Monsters  []PlatformMonster // TODO: ???
	Objects   []PlatformObject  // TODO: ???
	Blocks    []PlatformObject  // TODO: ???
	HaveDecor bool
}

func NewPlatform(info *PlatformInfo, posX, posY int16, exits [4]int16) *Platform {
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
			point := getPortalCoord(dir, exits)

			platform.EnterX = point.X
			platform.EnterY = point.Y
			platform.EnterDir = dir
		}
	}

	// Cells and walls
	createCells(platform)

	return platform
}

func getPortalCoord(dir PlatformDir, exit [4]int16) Point16 {
	switch {
	case (dir == DIR_NORTH) && (exit[DIR_NORTH] != -1):
		return Point16{exit[DIR_NORTH], 0}

	case (dir == DIR_EAST) && (exit[DIR_EAST] != -1):
		return Point16{PLATFORM_SIDE_SIZE - 1, exit[DIR_EAST]}

	case (dir == DIR_SOUTH) && (exit[DIR_SOUTH] != -1):
		return Point16{exit[DIR_SOUTH], PLATFORM_SIDE_SIZE - 1}

	case (dir == DIR_WEST) && (exit[DIR_WEST] != -1):
		return Point16{0, exit[DIR_WEST]}

	default:
		return Point16{-1, -1}
	}

	return Point16{-1, -1}
}

func createCells(platform *Platform) {
	// TODO: разделить??
	switch platform.Info.Type {
	case PLATFORM_INFO_TYPE_BRIDGE:
		makeBridgeCells(platform)

	case PLATFORM_INFO_TYPE_BATTLE:
		makeBattleCells(platform)
	}
}

func makeBridgeCells(platform *Platform) {
	w := platform.Width
	h := platform.Height

	// Blocks
	block3x3 := make([]*PlatformObjectInfo, 0)
	for i := range platform.Info.Blocks {
		obj := &platform.Info.Blocks[i]
		if (obj.Width == PLATFORM_BLOCK_SIZE_3x3) && (obj.Height == PLATFORM_BLOCK_SIZE_3x3) {
			block3x3 = append(block3x3, obj)
		}
	}

	// Info
	cellsInfo := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}
	for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
		cellsInfo[i] = CELL_TYPE_BLOCK
	}

	for y := uint16(0); y < h; y += PLATFORM_BLOCK_SIZE_3x3 {
		for x := uint16(0); x < w; x += PLATFORM_BLOCK_SIZE_3x3 {
			haveBlock := false
			for yy := uint16(0); yy < PLATFORM_BLOCK_SIZE_3x3; yy++ {
				for xx := uint16(0); xx < PLATFORM_BLOCK_SIZE_3x3; xx++ {
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
				platform.Objects = appendObjects(platform.Objects, block3x3,
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
}

func makeBattleCells(platform *Platform) {
	endPoints := make([]Point16, 0)

	w := platform.Width
	h := platform.Height

	// Blocks
    block3x3 := make([]*PlatformObjectInfo, 0)
    for i := range platform.Info.Blocks {
        obj := &platform.Info.Blocks[i]
        if (obj.Width == PLATFORM_BLOCK_SIZE_3x3) && (obj.Height == PLATFORM_BLOCK_SIZE_3x3) {
            block3x3 = append(block3x3, obj)
        }
    }
    /*block1x1 := make([]*PlatformObjectInfo, 0)
    block2x2 := make([]*PlatformObjectInfo, 0)
    for i := range platform.Info.Blocks {
        obj := &platform.Info.Blocks[i]
        if (obj.Width == PLATFORM_BLOCK_SIZE_6x6) && (obj.Height == PLATFORM_BLOCK_SIZE_6x6) {
            block2x2 = append(block2x2, obj)
        } else if (obj.Width == PLATFORM_BLOCK_SIZE_3x3) && (obj.Height == PLATFORM_BLOCK_SIZE_3x3) {
            block1x1 = append(block1x1, obj)
        }
    }*/

	// Info
	cellsInfo := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}
	cellsWalls := [PLATFORM_SIDE_SIZE * PLATFORM_SIDE_SIZE]PlatformCellType{}
	for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
		cellsInfo[i] = CELL_TYPE_BLOCK
		cellsWalls[i] = CELL_TYPE_UNDEF
	}

	// Make cells
	platform.Cells = make([]PlatformCellType, PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE)
	for i := range platform.Cells {
		platform.Cells[i] = CELL_TYPE_BLOCK
	}

	// Make logic
	foundPath := false
	for foundPath == false {
		// Clear arrays
		platform.Blocks = make([]PlatformObject, 0)
		platform.Objects = make([]PlatformObject, 0)

		// Clear info
		for i := 0; i < PLATFORM_SIDE_SIZE*PLATFORM_SIDE_SIZE; i++ {
			cellsInfo[i] = CELL_TYPE_BLOCK
			cellsWalls[i] = CELL_TYPE_UNDEF
		}

		// TODO: ???

        createBlocks6x6(platform)
		createBlocks3x3(platform, []PlatformCellType(cellsInfo), block3x3)

		//createArches();
		//createWalls();

		// заполняем стенами ячейки
		for y := uint16(0); y < PLATFORM_WORK_SIZE; y++ {
			for x := uint16(0); x < PLATFORM_WORK_SIZE; x++ {
				index := uint16(y*w + x)
				if cellsWalls[index] != CELL_TYPE_UNDEF {
					cellsInfo[index] = cellsWalls[index]
				}
			}
		}

		// TODO: ???
		//createPlatformElements();

		start := -1
		foundPath = false

		for i := 1; i < 4; i++ {
			if platform.ExitCoord[i] != -1 {
				if start == -1 {
					start = i
				} else {
					endPoints = make([]Point16, 1)

					endPoint := getPortalCoord(PlatformDir(i), platform.ExitCoord)
					endPoints = append(endPoints, endPoint)

					startPoint := getPortalCoord(PlatformDir(start), platform.ExitCoord)

					// TODO: ???
					//path = _pathManager.findPathOld(startpt, endPoints, false, true, false);
					path := make([]Point16, 0)

					if len(path) > 0 {
						foundPath = true
						break
					}
				}
			}
		}
	}

	for i := 0; i < 4; i++ {
		dir := PlatformDir(i)

		exitPoint := getPortalCoord(dir, platform.ExitCoord)
		if (exitPoint.X == -1) || (exitPoint.Y == -1) {
			continue
		}

		y := exitPoint.Y
		x := exitPoint.X

		if dir == DIR_EAST {
			x -= PLATFORM_BLOCK_SIZE_6x6
		}
		if dir == DIR_SOUTH {
			y -= PLATFORM_BLOCK_SIZE_6x6
		}

		if (dir == DIR_NORTH) || (dir == DIR_SOUTH) {
			cellsInfo[y*int16(w)+(x-2)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x-1)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x+1)] = CELL_TYPE_WALL
			cellsInfo[y*int16(w)+(x+2)] = CELL_TYPE_WALL
		}
		if (dir == DIR_EAST) || (dir == DIR_WEST) {
			cellsInfo[(y-2)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y-1)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y+1)*int16(w)+x] = CELL_TYPE_WALL
			cellsInfo[(y+2)*int16(w)+x] = CELL_TYPE_WALL
		}
	}

	for i := uint16(0); i < w*h; i++ {
		platform.Cells[i] = cellsInfo[i]
	}
}

func createBlocks6x6(platform *Platform) {
	platform.HaveDecor = false

	for y := int16(0); y < PLATFORM_WORK_SIZE; y += PLATFORM_BLOCK_SIZE_6x6 {
		for x := int16(0); x < PLATFORM_WORK_SIZE; x += PLATFORM_BLOCK_SIZE_6x6 {
			curPoint := NewPoint16(x, y).Div(PLATFORM_BLOCK_SIZE_6x6)

			northPoint := getPortalCoord(DIR_NORTH, platform.ExitCoord)
			eastPoint := getPortalCoord(DIR_EAST, platform.ExitCoord)
			southPoint := getPortalCoord(DIR_SOUTH, platform.ExitCoord)
			westPoint := getPortalCoord(DIR_WEST, platform.ExitCoord)

			testPoint1 := northPoint.Div(PLATFORM_BLOCK_SIZE_6x6)
			testPoint2 := NewPoint16(eastPoint.X-PLATFORM_BLOCK_SIZE_6x6, eastPoint.Y).Div(PLATFORM_BLOCK_SIZE_6x6)
			testPoint3 := NewPoint16(southPoint.X, southPoint.Y-PLATFORM_BLOCK_SIZE_6x6).Div(PLATFORM_BLOCK_SIZE_6x6)
			testPoint4 := westPoint.Div(PLATFORM_BLOCK_SIZE_6x6)

			isExit := false
			if (curPoint == testPoint1) || (curPoint == testPoint2) || (curPoint == testPoint3) || (curPoint == testPoint4) {
				isExit = true
			}

			posTest := (y == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3) && (x == PLATFORM_WORK_SIZE/2-PLATFORM_BLOCK_SIZE_3x3)
			if (rand.Int()%2 == 0) || posTest || ((rand.Int() == 0) && isExit) {
				platform.Objects = appendObjects(platform.Objects,
					platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_DECOR],
					x, y, 0, 3)
			}
		}
	}
}

// TODO: Пробрасывается ли указатель в cellInfo??
func createBlocks3x3(platform *Platform, cellInfo []PlatformCellType, block3x3 []*PlatformObjectInfo) {
    edges := make([]Point16, 0)
    for y := int16(0); y < PLATFORM_WORK_SIZE; y += PLATFORM_BLOCK_SIZE_3x3 {
        for x := int16(0); x < PLATFORM_WORK_SIZE; x += PLATFORM_BLOCK_SIZE_3x3 {
            addEdge := false
            addEdge = addEdge || (x == 0)
            addEdge = addEdge || (y == 0)
            addEdge = addEdge || (x == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3)
            addEdge = addEdge || (y == PLATFORM_WORK_SIZE-PLATFORM_BLOCK_SIZE_3x3)

            if addEdge {
                edges = append(edges, NewPoint16(x, y))
            }
        }
    }
    // Перемешивание
    for i := range edges {
        j := rand.Intn(i + 1)
        edges[i], edges[j] = edges[j], edges[i]
    }

    for i := 0; i < 4; i++ {
        dir := PlatformDir(i)
        exit := getPortalCoord(dir, platform.ExitCoord)
        if (exit.X == -1) || (exit.Y == -1) {
            continue
        }
        // TODO: было в оригинале
        exit = exit.Div(3)
        exit = exit.Mul(3)

        if cellInfo[exit.Y*int16(platform.Width) + exit.X] == CELL_TYPE_BLOCK {
            for y := int16(0); y < PLATFORM_BLOCK_SIZE_3x3; y++ {
                for x := int16(0); x < PLATFORM_BLOCK_SIZE_3x3; x++ {
                    cellInfo[(exit.Y+y)*int16(platform.Width) + (exit.X+x)] = CELL_TYPE_SPACE
                }
            }
            if (rand.Int()%3 == 0) && (dir == DIR_EAST || dir == DIR_SOUTH) {
                direction := int8(1)
                if i & 1 {
                    direction = 0
                }
                platform.Blocks = appendObjects(platform.Blocks,
                    platform.Info.ObjectsByType[PLATFORM_OBJ_TYPE_FLOOR],
                    exit.X, exit.Y,
                    direction,
                    3)
            }else{
                direction := int8((exit.X + exit.Y) & 3)
                platform.Blocks = appendObjects(platform.Blocks,
                    block3x3,
                    exit.X, exit.Y,
                    direction,
                    3)
            }
        }
    }

    exit := getPortalCoord(DIR_EAST, platform.ExitCoord)
    if exit.X != -1 {
        edges = append(edges, NewPoint16(exit.X - 2, exit.Y - 1))
    }
    exit = getPortalCoord(DIR_SOUTH, platform.ExitCoord)
    if exit.X != -1 {
        edges = append(edges, NewPoint16(exit.X - 1, exit.Y - 2))
    }

    center := NewPoint16(PLATFORM_WORK_SIZE/2 - PLATFORM_BLOCK_SIZE_3x3, PLATFORM_WORK_SIZE/2 - PLATFORM_BLOCK_SIZE_3x3)

    edgesSize := len(edges)
    for i := 0; i < edgesSize; i++ {
        // Iterate from end
        point := edges[len(edges)-1]
        edges = edges[0:len(edges)-1]

        searchComplete := false
        for searchComplete {
            // Check1
            check1 := false
            check1 = check1 || (rand.Int()%2 == 0)
            check1 = check1 || (point.Y / PLATFORM_BLOCK_SIZE_6x6 == center.Y/PLATFORM_BLOCK_SIZE_6x6)
            check1 = check1 || (point.X >= (PLATFORM_WORK_SIZE - PLATFORM_BLOCK_SIZE_3x3))
            // Check2
            check2 := point.X/PLATFORM_BLOCK_SIZE_6x6 != center.X/PLATFORM_BLOCK_SIZE_6x6
            // Check3
            check3 := point.Y < (PLATFORM_WORK_SIZE - PLATFORM_BLOCK_SIZE_3x3)
            if check1 && check2 && check3 {
                if point.X / PLATFORM_BLOCK_SIZE_6x6 < center.X / PLATFORM_BLOCK_SIZE_6x6 {
                    point.X += PLATFORM_BLOCK_SIZE_3x3
                }else{
                    point.X -= PLATFORM_BLOCK_SIZE_3x3
                }
            } else {
                if point.Y / PLATFORM_BLOCK_SIZE_6x6 < center.Y / PLATFORM_BLOCK_SIZE_6x6 {
                    point.Y += PLATFORM_BLOCK_SIZE_3x3
                }else{
                    point.Y -= PLATFORM_BLOCK_SIZE_3x3
                }
            }

            check4 := point.Div(PLATFORM_BLOCK_SIZE_6x6) != center.Div(PLATFORM_BLOCK_SIZE_6x6)
            check5 := cellInfo[point.Y * int16(platform.Width) + point.X] == CELL_TYPE_BLOCK
            if check4 && check5 {
                for y := int16(0); y < PLATFORM_BLOCK_SIZE_3x3; y++ {
                    for x := int16(0); x < PLATFORM_BLOCK_SIZE_3x3; x++ {
                        cellInfo[(exit.Y+y)*int16(platform.Width) + (exit.X+x)] = CELL_TYPE_SPACE
                    }
                }
            }

            // TODO: ???
            if point.Div(PLATFORM_BLOCK_SIZE_6x6) == center.Div(PLATFORM_BLOCK_SIZE_6x6) {
                searchComplete = true
            }
        }
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
