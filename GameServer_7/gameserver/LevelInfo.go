package gameserver

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type LevelInfo struct {
	LevelColor           ColorFloat `json:"levelColor"`
	OmniColor            ColorFloat `json:"omniColor"`
	OmniMult             float64    `json:"omniMult"`
	BgColor              ColorFloat `json:"bgColor"`
	DefaultFogColor      ColorFloat `json:"fogColor"`
	FogNodeColor         ColorFloat `json:"fogNodeColor"`
	FogNodeFogBrightness float64    `json:"fogNodeFogBrightness"`
	LightCoeffMult       float64    `json:"lightMult"`
	TexturesFolderPath   string     `json:"texturesPath"`
	FogNodeTexture       string     `json:"fogTexture"`
	FogNodeHeight        float64    `json:"fogNodeHeight"`
	FogNodeTcScale       float64    `json:"fogNodeTcScale"`
	FogNodeSize          float64    `json:"fogNodeSize"`
	Platforms            []string   `json:"platforms"`
}

func NewLevelsFromReader(reader io.Reader) (map[string]*LevelInfo, error) {
	result := make(map[string]*LevelInfo)
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&result)
	return result, err
}

func NewLevelsFromFile(filePath string) (map[string]*LevelInfo, error) {
	// Загрузка платформ из файла
	f, err := os.Open(filePath)
	if err != nil {
		log.Println(err)
		return make(map[string]*LevelInfo), err
	}
	defer f.Close()

	return NewLevelsFromReader(f)
}
