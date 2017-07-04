package gameserver

import (
	"bytes"
	"encoding/binary"
	"errors"
)

const CLIENT_COMMAND_MAGIC_NUMBER uint8 = 3

const (
	CLIENT_COMMAND_TYPE_MOVE  uint8 = 0
	CLIENT_COMMAND_TYPE_SHOOT uint8 = 1
)

type ClientCommand struct {
	ID    uint32
	X     int16
	Y     int16
	Angle float32
	Type  uint8
}

func IsClientCommandData(data []byte) bool {
	reader := bytes.NewReader(data)
	var magicNumber uint8 = 0
	err := binary.Read(reader, binary.BigEndian, &(magicNumber))
	if err != nil {
		return false
	}
	if magicNumber != CLIENT_COMMAND_MAGIC_NUMBER {
		return false
	}
	return true
}

func NewClientCommand(data []byte) (ClientCommand, error) {
	reader := bytes.NewReader(data)

	// MagicNumber
	var magicNumber uint8 = 0
	err := binary.Read(reader, binary.BigEndian, &(magicNumber))
	if err != nil {
		return ClientCommand{}, err
	}
	if magicNumber != CLIENT_COMMAND_MAGIC_NUMBER {
		return ClientCommand{}, errors.New("Wrong magic number for client command")
	}

	// State object
	command := ClientCommand{}

	// ID
	err = binary.Read(reader, binary.BigEndian, &(command.ID))
	if err != nil {
		return ClientCommand{}, err
	}
	// X
	err = binary.Read(reader, binary.BigEndian, &(command.X))
	if err != nil {
		return ClientCommand{}, err
	}
	// Y
	err = binary.Read(reader, binary.BigEndian, &(command.Y))
	if err != nil {
		return ClientCommand{}, err
	}
	// Angle
	err = binary.Read(reader, binary.BigEndian, &(command.Angle))
	if err != nil {
		return ClientCommand{}, err
	}
	// Type
	err = binary.Read(reader, binary.BigEndian, &(command.Type))
	if err != nil {
		return ClientCommand{}, err
	}

	return command, nil
}
