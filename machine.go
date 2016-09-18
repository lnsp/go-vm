package main

import (
	"encoding/binary"
	"fmt"
)

var (
	ByteOrder = binary.BigEndian
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
	FlagSize = map[uint16]int{
		FLAG_RR:   2,
		FLAG_RI:   2,
		FLAG_RA:   2,
		FLAG_AA:   2,
		FLAG_AR:   2,
		FLAG_IA:   2,
		FLAG_IR:   2,
		FLAG_I:    1,
		FLAG_R:    1,
		FLAG_NONE: 0,
	}
)

type Machine struct {
	ByteOrder   binary.ByteOrder
	Memory      []byte
	NextCmd     uint16
	Flag        uint16
	Cmd         uint16
	Args        [MAX_CMD_ARGS]uint16
	KeepRunning bool
}

func NewMachine() *Machine {
	return &Machine{
		ByteOrder: ByteOrder,
	}
}

func (machine *Machine) Boot(code []byte) {
	machine.initialize()
	machine.program(code)
	machine.run()
}

func (machine *Machine) interrupt(value, kind uint16) {
	machine.store(INTERRUPT, value)
	machine.store(CODE_POINTER, machine.load(kind))
	machine.run()
}

func (machine *Machine) push(value uint16) {
	stack := machine.load(STACK_POINTER)
	if stack >= STACK_MAX-1 {
		machine.interrupt(IR_OVERFLOW_STACK, IR_OVERFLOW)
		return
	}
	machine.store(stack+2, value)
	machine.store(STACK_POINTER, stack+2)
}

func (machine *Machine) pop() uint16 {
	stack := machine.load(STACK_POINTER)
	value := machine.load(stack)
	machine.store(stack, 0)
	if stack > STACK_BASE {
		machine.store(STACK_POINTER, stack-2)
	}
	return value
}

func (machine *Machine) fetchCmd() uint16 {
	code := machine.load(CODE_POINTER)
	if code >= MAX_MEMORY-1 {
		machine.interrupt(IR_OVERFLOW_CODE, IR_OVERFLOW)
		return CMD_HLT
	}
	machine.store(CODE_POINTER, code+2)
	return machine.load(code)
}

func (machine *Machine) handle() {
	var value uint16
	switch machine.Cmd {
	case CMD_ADD:
		machine.PerformArithmetic(func(a, b uint16) uint16 { return a + b }, func(a, b int) int { return a + b })
	case CMD_SUB:
		machine.PerformArithmetic(func(a, b uint16) uint16 { return a - b }, func(a, b int) int { return a - b })
	case CMD_MUL:
		machine.PerformArithmetic(func(a, b uint16) uint16 { return a * b }, func(a, b int) int { return a * b })
	case CMD_DIV:
		machine.PerformArithmetic(func(a, b uint16) uint16 { return a / b }, func(a, b int) int { return a / b })
	case CMD_INC:
		machine.PerformSimpleArithmetic(func(a uint16) uint16 { return a + 1 }, func(a int) int { return a + 1 })
	case CMD_DEC:
		machine.PerformSimpleArithmetic(func(a uint16) uint16 { return a - 1 }, func(a int) int { return a - 1 })
	case CMD_AND:
		machine.PerformLogic(func(a, b uint16) uint16 { return a & b })
	case CMD_OR:
		machine.PerformLogic(func(a, b uint16) uint16 { return a | b })
	case CMD_XOR:
		machine.PerformLogic(func(a, b uint16) uint16 { return a ^ b })
	case CMD_NOT:
		machine.PerformSimpleLogic(func(a uint16) uint16 { return a &^ 0xFFFF })
	case CMD_SHL:
		machine.PerformLogic(func(a, b uint16) uint16 { return a << b })
	case CMD_SHR:
		machine.PerformLogic(func(a, b uint16) uint16 { return a >> b })
	case CMD_MOV:
		machine.PerformMove()
	case CMD_PUSH:
		switch machine.Flag {
		case FLAG_I:
			value = machine.Args[0]
		case FLAG_R:
			value = machine.load(machine.Args[0])
		}
		machine.push(value)
	case CMD_POP:
		value = machine.pop()
		machine.store(machine.Args[0], value)
	case CMD_CMP:
		machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a == b) })
	case CMD_CNT:
		machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a != b) })
	case CMD_LGE:
		machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a >= b) })
	case CMD_SME:
		machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a <= b) })
	case CMD_JIF:
		machine.PerformJump(false)
	case CMD_JMP:
		machine.PerformJump(true)
	case CMD_CALL:
		switch machine.Flag {
		case FLAG_I:
			value = machine.Args[0]
		case FLAG_R:
			value = machine.load(machine.Args[0])
		}
		machine.push(machine.load(CODE_POINTER))
		machine.store(CODE_POINTER, value)
	case CMD_RET:
		value = machine.pop()
		machine.store(CODE_POINTER, value)
	case CMD_HLT:
		machine.KeepRunning = false
	}
}

func (machine *Machine) parseState() {
	machine.Flag = machine.NextCmd & FLAG_MASK
	machine.Cmd = machine.NextCmd & CMD_MASK

	maxFlags := FlagSize[machine.Flag]
	for i := 0; i < maxFlags; i++ {
		machine.Args[i] = machine.fetchCmd()
	}

}

func (machine *Machine) iterate() {
	machine.NextCmd = machine.fetchCmd()
}

func (machine *Machine) run() {
	machine.iterate()
	for machine.KeepRunning {
		machine.parseState()
		machine.handle()
		machine.iterate()
	}
}

func (machine *Machine) program(code []byte) {
	size := len(code)

	for i := 0; i < size; i++ {
		machine.Memory[int(CODE_BASE)+i] = code[i]
	}
}

func (machine *Machine) initialize() {
	machine.Memory = make([]byte, int(MAX_MEMORY)+1)

	// Load base values
	machine.store(CODE_POINTER, CODE_BASE)
	machine.store(STACK_POINTER, STACK_BASE)
	machine.store(CODE_BASE, CMD_HLT)
	machine.store(INTERRUPT, MAX_MEMORY-1)

	// Init graphics
	machine.store(OUT_MODE, OUT_MODE_TERM)
	pointer := OUT_COLORS
	for _, color := range BaseColors {
		machine.store(pointer, color)
		pointer += 2
	}

	machine.KeepRunning = true
}

func (machine *Machine) load(addr uint16) uint16 {
	return machine.ByteOrder.Uint16(machine.Memory[addr : addr+2])
}

func (machine *Machine) store(addr, value uint16) {
	machine.ByteOrder.PutUint16(machine.Memory[addr:addr+2], value)
}

func (machine Machine) String() string {
	return machine.dumpSegment(0)
}

func (machine *Machine) dumpSegment(seg int) string {
	start := seg * 256
	dump := fmt.Sprintf("SEGMENT %4.4X - %4.4X\n-------------------", start, start+0xFF)
	for i := start; i <= start+0xFF; i++ {
		if i%16 == 0 {
			dump += "\n"
		}
		dump += fmt.Sprintf("%-3.2X", machine.Memory[i])
	}
	return dump
}
