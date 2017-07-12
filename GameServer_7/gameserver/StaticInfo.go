package gameserver

import "log"

type StaticInfo struct {
	Platforms map[string]*PlatformInfo
	Levels    map[string]*LevelInfo
}

func NewStaticInfo() (*StaticInfo, error) {
	// Load platforms
	platforms, err := NewPlatformsFromFile("data/platforms.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Load levels
	levels, err := NewLevelsFromFile("data/level_graphics.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	staticInfo := &StaticInfo{
		Platforms: platforms,
		Levels:    levels,
	}
	return staticInfo, nil
}
