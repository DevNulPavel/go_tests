package gameserver

type PlatformCellType int8

const (
	CELL_TYPE_UNDEF PlatformCellType = -1 // не определено
	CELL_TYPE_BLOCK PlatformCellType = 0  // непроходимо

	CELL_TYPE_WALK PlatformCellType = 1 << 0 // можно ходить
	CELL_TYPE_PROJ PlatformCellType = 1 << 1 // могут пролетать снаряды
	CELL_TYPE_LOOK PlatformCellType = 1 << 2 // можно смотреть

	CELL_TYPE_PIT   PlatformCellType = CELL_TYPE_PROJ | CELL_TYPE_LOOK                  // яма
	CELL_TYPE_HOLE  PlatformCellType = CELL_TYPE_PROJ | CELL_TYPE_LOOK                  // дыра в стене
	CELL_TYPE_GLASS PlatformCellType = CELL_TYPE_LOOK                                   // стекло
	CELL_TYPE_GRASS PlatformCellType = CELL_TYPE_WALK | CELL_TYPE_PROJ                  // трава
	CELL_TYPE_SPACE PlatformCellType = CELL_TYPE_WALK | CELL_TYPE_PROJ | CELL_TYPE_LOOK // свободное пространство
)
