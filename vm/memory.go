package vm

import (
	"encoding/binary"
	"errors"
)

type Memory interface {
	Load(addr uint16) (uint16, error)
	Store(addr, value uint16) error
	StoreByte(addr uint16, value byte) error
	Convert(value uint16) []byte
	Size() uint16
}

type randomAccessMemory []byte

func (memory randomAccessMemory) Size() uint16 {
	return uint16(len(memory))
}
func (memory randomAccessMemory) Load(addr uint16) (uint16, error) {
	if addr >= memory.Size() {
		return 0, errors.New("memory address out of range")
	}
	return binary.BigEndian.Uint16(memory[addr : addr+2]), nil
}

func (memory randomAccessMemory) Store(addr, value uint16) error {
	if addr >= memory.Size() {
		return errors.New("memory address out of range")
	}
	binary.BigEndian.PutUint16(memory[addr:addr+2], value)
	return nil
}

func (memory randomAccessMemory) StoreByte(addr uint16, value byte) error {
	if addr >= memory.Size() {
		return errors.New("memory address out of range")
	}
	memory[addr] = value
	return nil
}

func (randomAccessMemory) Convert(value uint16) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, value)
	return data
}

func NewMemory(size int) Memory {
	instance := make(randomAccessMemory, size)
	return &instance
}
