package gameserver

import (
	"bytes"
	"encoding/binary"
)

// ClienState ... Client state structure
type ClienState struct {
	ID    int32   `json:"id"`
	X     int32   `json:"x"`
	Y     int32   `json:"y"`
	Delta float64 `json:"d"`
}

func NewClientState(rawData []byte) (ClienState, error) {
	reader := bytes.NewReader(rawData)

	newState := ClienState{}
	// ID
	err := binary.Read(reader, binary.BigEndian, &(newState.ID))
	if err != nil {
		return ClienState{}, err
	}
	// X
	err = binary.Read(reader, binary.BigEndian, &(newState.X))
	if err != nil {
		return ClienState{}, err
	}
	// Y
	err = binary.Read(reader, binary.BigEndian, &(newState.X))
	if err != nil {
		return ClienState{}, err
	}
	// Delta
	err = binary.Read(reader, binary.BigEndian, &(newState.Delta))
	if err != nil {
		return ClienState{}, err
	}
	return newState, nil
}

func (client *ClienState) ConvertToBytes() ([]byte, error) {
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
