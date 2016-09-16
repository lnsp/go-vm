package main

import (
	"encoding/binary"
	"fmt"
)

const (
	MAX_MEMORY    uint16 = 0xFFFF
	CODE_POINTER  uint16 = 0x0000
	STACK_POINTER uint16 = 0x0002
	ZERO_FLAG     uint16 = 0x0004
	CARRY_FLAG    uint16 = 0x0006
	REGISTER_AX   uint16 = 0x0008
	REGISTER_BX   uint16 = 0x000A
	REGISTER_CX   uint16 = 0x000C
	REGISTER_DX   uint16 = 0x000D
	INTERRUPT     uint16 = 0x0010
	IR_STATE      uint16 = 0x0012
	IR_KEYBOARD   uint16 = 0x0014
	IR_OVERFLOW   uint16 = 0x0016
	STACK_BASE    uint16 = 0x0100
	STACK_MAX     uint16 = 0x01FF
	OUT_CHARS     uint16 = 0x1000
	OUT_COLORS    uint16 = 0x1F00
	OUT_MODE      uint16 = 0x1FFE
	OUT_MODE_TERM uint16 = 0x0001
	CODE_BASE     uint16 = 0x2000

	FLAG_MASK uint16 = 0xFF00
	FLAG_RR   uint16 = 0x8000
	FLAG_RI   uint16 = 0x4000
	FLAG_RA   uint16 = 0x2000
	FLAG_AA   uint16 = 0x1000
	FLAG_AR   uint16 = 0x800
	FLAG_IA   uint16 = 0x400
	FLAG_IR   uint16 = 0x200

	CMD_ADD  uint16 = 0x01 // R,R - R,I
	CMD_SUB  uint16 = 0x02 // R,R - R,I
	CMD_MUL  uint16 = 0x03 // R,R - R,I
	CMD_DIV  uint16 = 0x04 // R,R - R,I
	CMD_INC  uint16 = 0x05 // R,R - R,I
	CMD_DEC  uint16 = 0x06 // R,R - R,I
	CMD_AND  uint16 = 0x07 // R,R - R,I
	CMD_OR   uint16 = 0x08 // R,R - R,I
	CMD_XOR  uint16 = 0x09 // R,R - R,I
	CMD_NOT  uint16 = 0x0A // R,R - R,I
	CMD_SHL  uint16 = 0x0B // R,R - R,I
	CMD_SHR  uint16 = 0x0C // R,R - R,I
	CMD_MOV  uint16 = 0x0D // R,R - R,A - A,A - A,R - I,A - I,R
	CMD_PUSH uint16 = 0x0E // R - I
	CMD_POP  uint16 = 0x0F // R
	CMD_CMP  uint16 = 0x10 // R,R - R,I
	CMD_CNT  uint16 = 0x11 // R,R - R,I
	CMD_JIF  uint16 = 0x12 // R - I
	CMD_JMP  uint16 = 0x13 // R - I
	CMD_HLT  uint16 = 0x14
)

var (
	BaseColors = []uint16{
		0x000, // Black
		0xFFF, // White
		0xF00, // Red
		0x0F0, // Green
		0x00F, // Blue
		0xFF0, // Yellow
		0xF0F, // Fuchsia
		0x0FF, // Aqua
	}
	ByteOrder = binary.BigEndian
	Memory    []byte
)

func Boot(code []byte) {
	initialize()
	program(code)
	evaluate()
}

func evaluate() {
	cmd := load(CODE_POINTER)
	for cmd != CMD_HALT {
		iterations := 1
		for i := 0; i < iterations; i++ {
			switch cmd {
			case CMD_ADD:
			case CMD_SUB:
			case CMD_MUL:
			case CMD_DIV:
			case CMD_INC:
			case CMD_DEC:
			case CMD_AND:
			case CMD_OR:
			case CMD_XOR:
			case CMD_NOT:
			case CMD_SHL:
			case CMD_SHR:
			case CMD_MOV:
			case CMD_PUSH:
			case CMD_POP:
			case CMD_CMP:
			case CMD_CNT:
			case CMD_JIF:
			case CMD_JMP:
			case CMD_HLT:
			}
		}
	}
}

func program(code []byte) {
	size := len(code)

	for i := 0; i < size; i++ {
		Memory[int(CODE_BASE)+i] = code[i]
	}
}

func initialize() {
	Memory = make([]byte, int(MAX_MEMORY)+1)

	// Load base values
	store(CODE_POINTER, CODE_BASE)
	store(CODE_BASE, CMD_HLT)
	store(INTERRUPT, MAX_MEMORY-1)

	// Init graphics
	store(OUT_MODE, OUT_MODE_TERM)
	pointer := OUT_COLORS
	for _, color := range BaseColors {
		store(pointer, color)
		pointer += 2
	}
}

func load(addr uint16) uint16 {
	return ByteOrder.Uint16(Memory[addr : addr+2])
}

func store(addr, value uint16) {
	ByteOrder.PutUint16(Memory[addr:addr+2], value)
}

func printSegment(seg int) {
	start := seg * 16
	fmt.Printf("SEGMENT %4.4X - %4.4X\n-------------------", start, start+0xFF)
	for i := start; i <= start+0xFF; i++ {
		if i%16 == 0 {
			fmt.Println()
		}
		fmt.Printf("%-3.2X", Memory[i])
	}
	fmt.Println()
}
