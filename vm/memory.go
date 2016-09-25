package vm

import (
	"encoding/binary"
	"fmt"
)

// Memory is a virtual representation of a RAM.
type Memory interface {
	Load(addr uint16) (uint16, error)
	Store(addr, value uint16) error
	StoreByte(addr uint16, value byte) error
	Segment(from, to uint16) []byte
	Convert(value uint16) []byte
	InRange(addr uint16) bool
}

// OutOfRangeError is thrown if a address is out of memory range.
type OutOfRangeError struct {
	Address uint16
}

func (err OutOfRangeError) Error() string {
	return fmt.Sprintf("0x%4.4X is out of memory range", err.Address)
}

// randomAccessMemory is a basic byte storage.
type randomAccessMemory []byte

// InRange checks if the given address is in memory range.
func (memory randomAccessMemory) InRange(addr uint16) bool {
	return int(addr) < len(memory)
}

// Load fetches a word from memory.
func (memory randomAccessMemory) Load(addr uint16) (uint16, error) {
	if !memory.InRange(addr) {
		return 0, &OutOfRangeError{addr}
	}
	return binary.BigEndian.Uint16(memory[addr : addr+2]), nil
}

// Store puts a word into memory.
func (memory randomAccessMemory) Store(addr, value uint16) error {
	if !memory.InRange(addr) {
		return &OutOfRangeError{addr}
	}
	binary.BigEndian.PutUint16(memory[addr:addr+2], value)
	return nil
}

// StoreByte puts a byte into memory.
func (memory randomAccessMemory) StoreByte(addr uint16, value byte) error {
	if !memory.InRange(addr) {
		return &OutOfRangeError{addr}
	}
	memory[addr] = value
	return nil
}

// Segment returns a memory segment.
func (memory randomAccessMemory) Segment(from, to uint16) []byte {
	return memory[from:to]
}

// Convert converts a word into a slice of bytes.
func (randomAccessMemory) Convert(value uint16) []byte {
	data := make([]byte, 2)
	binary.BigEndian.PutUint16(data, value)
	return data
}

// NewMemory creates a new memory with a given range.
func NewMemory(size int) Memory {
	instance := make(randomAccessMemory, size)
	return &instance
}
