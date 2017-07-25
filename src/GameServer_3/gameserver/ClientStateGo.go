package gameserver

import (
	"bytes"
	"encoding/binary"
)

// ClienState ... Client state structure
type ClientState struct {
	ID    int32   `json:"id"`
	X     int32   `json:"x"`
	Y     int32   `json:"y"`
	Delta float64 `json:"d"`
}

func NewClientState(rawData []byte) (ClientState, error) {
	reader := bytes.NewReader(rawData)

	newState := ClientState{}
	// ID
	err := binary.Read(reader, binary.BigEndian, &(newState.ID))
	if err != nil {
		return ClientState{}, err
	}
	// X
	err = binary.Read(reader, binary.BigEndian, &(newState.X))
	if err != nil {
		return ClientState{}, err
	}
	// Y
	err = binary.Read(reader, binary.BigEndian, &(newState.Y))
	if err != nil {
		return ClientState{}, err
	}
	// Delta
	err = binary.Read(reader, binary.BigEndian, &(newState.Delta))
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
	// X
	err = binary.Write(buffer, binary.BigEndian, client.X)
	if err != nil {
		return []byte{}, err
	}
	// Y
	err = binary.Write(buffer, binary.BigEndian, client.Y)
	if err != nil {
		return []byte{}, err
	}
	// Delta
	err = binary.Write(buffer, binary.BigEndian, client.Delta)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}

// конвертация в строку
//func (mes *ClienState) String() string {
//	return "User " + string(mes.Id) + " on : " + string(mes.X) + "x" + string(mes.Y)
//}