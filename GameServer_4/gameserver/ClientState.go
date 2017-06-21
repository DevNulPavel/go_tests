package gameserver

import (
	"bytes"
	"encoding/binary"
)

const (
	CLIENT_STATUS_NOT_IN_GAME = 0
	CLIENT_STATUS_IN_GAME     = 1
	CLIENT_STATUS_FAIL        = 2
	CLIENT_STATUS_WIN         = 3
)

// ClienState ... Client state structure
type ClientState struct {
	ID     int32 `json:"id"`
	Y      int32 `json:"y"`
	Status int8  `json:"st"`
}

func NewClientState(rawData []byte) (ClientState, error) {
	reader := bytes.NewReader(rawData)

	newState := ClientState{}
	// ID
	err := binary.Read(reader, binary.BigEndian, &(newState.ID))
	if err != nil {
		return ClientState{}, err
	}
	// Y
	err = binary.Read(reader, binary.BigEndian, &(newState.Y))
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
	// ID
	err := binary.Write(buffer, binary.BigEndian, client.ID)
	if err != nil {
		return []byte{}, err
	}
	// Y
	err = binary.Write(buffer, binary.BigEndian, client.Y)
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
