package gameserver

import (
	"bytes"
	"encoding/binary"
)

const CLIENT_STATE_MAGIC_NUMBER uint8 = 1

const (
	CLIENT_STATUS_IN_GAME = 0
	CLIENT_STATUS_FAIL    = 1
	CLIENT_STATUS_WIN     = 2
)

// ClienState ... ServerClient state structure
type ServerClientState struct {
	ID     uint32
	Status int8
}

func (client *ServerClientState) ConvertToBytes() ([]byte, error) {
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
	// Status
	err = binary.Write(buffer, binary.BigEndian, client.Status)
	if err != nil {
		return []byte{}, err
	}
	return buffer.Bytes(), nil
}
