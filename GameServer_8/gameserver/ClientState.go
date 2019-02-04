package gameserver

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const CLIENT_STATE_MAGIC_NUMBER uint8 = 1

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
	CLIENT_STATUS_WIN     = 2
)

const (
	CLIENT_TYPE_LEFT  = 0
	CLIENT_TYPE_RIGHT = 1
)

// ClienState ... Client state structure
type ClientState struct {
	ID     uint32 `json:"id"`
	Type   uint8  `json:"t"`
	Y      int16  `json:"y"`
	Height int16  `json:"h"`
	Status int8   `json:"st"`
}

func IsClientStateData(rawData []byte) bool {
	reader := bytes.NewReader(rawData)
	var magicNumber uint8 = 0
	err := binary.Read(reader, binary.BigEndian, &(magicNumber))
	if err != nil {
		return false
	}
	if magicNumber != CLIENT_STATE_MAGIC_NUMBER {
		return false
	}
	return true
}

func NewClientState(rawData []byte) (ClientState, error) {
	reader := bytes.NewReader(rawData)

	// MagicNumber
	var magicNumber uint8 = 0
	err := binary.Read(reader, binary.BigEndian, &(magicNumber))
	if err != nil {
		return ClientState{}, err
	}
	if magicNumber != CLIENT_STATE_MAGIC_NUMBER {
		return ClientState{}, errors.New("Wrong magic number for client state")
	}

	// State object
	newState := ClientState{}

	// ID
	err = binary.Read(reader, binary.BigEndian, &(newState.ID))
	if err != nil {
		return ClientState{}, err
	}
	// Type
	err = binary.Read(reader, binary.BigEndian, &(newState.Type))
	if err != nil {
		return ClientState{}, err
	}
	// Y
	err = binary.Read(reader, binary.BigEndian, &(newState.Y))
	if err != nil {
		return ClientState{}, err
	}
	// Height
	err = binary.Read(reader, binary.BigEndian, &(newState.Height))
	if err != nil {
		return ClientState{}, err
	}
	// Status
	err = binary.Read(reader, binary.BigEndian, &(newState.Status))
	if err != nil {
		return ClientState{}, err
	}
	return newState, nil
}

func (client *ClientState) ConvertToBytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	// MagicNumber
	err := binary.Write(buffer, binary.BigEndian, CLIENT_STATE_MAGIC_NUMBER)
	if err != nil {
		return []byte{}, err
	}
	// ID
	err = binary.Write(buffer, binary.BigEndian, client.ID)
	if err != nil {
		return []byte{}, err
	}
	// Type
	err = binary.Write(buffer, binary.BigEndian, client.Type)
	if err != nil {
		return []byte{}, err
	}
	// Y
	err = binary.Write(buffer, binary.BigEndian, client.Y)
	if err != nil {
		return []byte{}, err
	}
	// Height
	err = binary.Write(buffer, binary.BigEndian, client.Height)
	if err != nil {
		return []byte{}, err
	}
	// Status
	err = binary.Write(buffer, binary.BigEndian, client.Status)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}
