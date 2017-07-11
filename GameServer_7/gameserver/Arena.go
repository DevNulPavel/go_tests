package gameserver

const ARENA_SIZE = 6 // Размер арены - сколько на сколь ячеек

type Arena struct {
	PlatformsMap    [ARENA_SIZE][ARENA_SIZE]*Platform
	BattlePlatforms []*PlatformInfo // TODO: ??? надо ли ???
	BridgePlatforms []*PlatformInfo // TODO: ??? надо ли ???
}
