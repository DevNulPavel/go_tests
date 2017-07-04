package gameserver

import (
    "encoding/binary"
    "bytes"
)

const WORLD_INFO_MAGIC_NUMBER uint8 = 2

type WorldInfo struct {
    SizeX uint16
    SizeY uint16
    ClientsCount uint16
}

func NewWorldInfo() WorldInfo {
    const worldSizeX = 400
    const worldSizeY = 400

    info := WorldInfo{
        SizeX: worldSizeX,
        SizeY: worldSizeY,
        ClientsCount: 0,
    }
    return info
}

func (info *WorldInfo) ConvertToBytes() ([]byte, error) {
    buffer := new(bytes.Buffer)

    // MagicNumber
    err := binary.Write(buffer, binary.BigEndian, WORLD_INFO_MAGIC_NUMBER)
    if err != nil {
        return []byte{}, err
    }
    // SizeX
    err = binary.Write(buffer, binary.BigEndian, info.SizeX)
    if err != nil {
        return []byte{}, err
    }
    // SizeY
    err = binary.Write(buffer, binary.BigEndian, info.SizeY)
    if err != nil {
        return []byte{}, err
    }
    // ClientsCount
    err = binary.Write(buffer, binary.BigEndian, info.ClientsCount)
    if err != nil {
        return []byte{}, err
    }

    return buffer.Bytes(), nil
}