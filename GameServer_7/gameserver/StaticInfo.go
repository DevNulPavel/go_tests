package gameserver

import (
	"io/ioutil"
	"log"
)

type StaticInfo struct {
	Platforms     map[string]*PlatformInfo
	Levels        map[string]*LevelInfo
	TestArenaData []byte
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

	// Test arena
	testArenaData, err := ioutil.ReadFile("data/arenaDump.json")
	if err != nil {
		log.Println(err)
		return nil, err
	}

	staticInfo := &StaticInfo{
		Platforms:     platforms,
		Levels:        levels,
		TestArenaData: testArenaData,
	}
	return staticInfo, nil
}
