package main

import (
	"encoding/binary"
	"fmt"
)

const (
	MAX_CMD_ARGS  uint16 = 0x02
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
	FLAG_RR   uint16 = 0x0100
	FLAG_RI   uint16 = 0x0200
	FLAG_RA   uint16 = 0x0300
	FLAG_AA   uint16 = 0x0400
	FLAG_AR   uint16 = 0x0500
	FLAG_IA   uint16 = 0x0600
	FLAG_IR   uint16 = 0x0700
	FLAG_I    uint16 = 0x0800
	FLAG_R    uint16 = 0x0900
	FLAG_NONE uint16 = 0x0000

	CMD_MASK uint16 = 0x00FF
	CMD_ADD  uint16 = 0x01 // R,R - R,I
	CMD_SUB  uint16 = 0x02 // R,R - R,I
	CMD_MUL  uint16 = 0x03 // R,R - R,I
	CMD_DIV  uint16 = 0x04 // R,R - R,I
	CMD_INC  uint16 = 0x05 // R
	CMD_DEC  uint16 = 0x06 // R
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
	CMD_LGE  uint16 = 0x17
	CMD_SME  uint16 = 0x18
	CMD_JIF  uint16 = 0x12 // R - I
	CMD_JMP  uint16 = 0x13 // R - I
	CMD_CLL  uint16 = 0x14
	CMD_RET  uint16 = 0x15
	CMD_HLT  uint16 = 0x16

	IR_OVERFLOW_CODE  uint16 = 0x1
	IR_OVERFLOW_STACK uint16 = 0x2
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
	ByteOrder     = binary.BigEndian
	Memory        []byte
	NextCommand   uint16
	ActiveFlag    uint16
	ActiveCommand uint16
	ActiveArgs    [MAX_CMD_ARGS]uint16
	ShutDown      bool
)

func Boot(code []byte) {
	initialize()
	program(code)
	evaluate()
}

func throwInterrupt(value, kind uint16) {
	store(INTERRUPT, value)
	store(CODE_POINTER, load(kind))
	executeActive()
}

func pushOnStack(value uint16) {
	stack := load(STACK_POINTER)
	if stack >= STACK_MAX-1 {
		throwInterrupt(IR_OVERFLOW_STACK, IR_OVERFLOW)
		return
	}
	store(stack+2, value)
	store(STACK_POINTER, stack+2)
}

func popFromStack() uint16 {
	stack := load(STACK_POINTER)
	value := load(stack)
	store(stack, 0)
	if stack > STACK_BASE {
		store(STACK_POINTER, stack-2)
	}
	return value
}

func loadNextCommand() uint16 {
	code := load(CODE_POINTER)
	if code >= MAX_MEMORY-1 {
		throwInterrupt(IR_OVERFLOW_CODE, IR_OVERFLOW)
		return CMD_HLT
	}
	store(CODE_POINTER, code+2)
	return load(CODE_POINTER)
}

func RunMinArithmetic(base func(uint16) uint16, carry func(int) int) {
	var value1, result, zeroFlag, carryFlag uint16
	value1 = load(ActiveArgs[0])
	result = base(value1)
	carryResult := carry(int(value1))
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	store(ZERO_FLAG, zeroFlag)
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	store(CARRY_FLAG, carryFlag)
	store(ActiveArgs[0], result)
}

func RunMinLogic(base func(uint16) uint16) {
	var value1, zeroFlag, result uint16
	value1 = load(ActiveArgs[0])
	result = base(value1)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	store(ZERO_FLAG, zeroFlag)
	store(CARRY_FLAG, 0)
	store(ActiveArgs[0], result)
}
func RunLogic(base func(uint16, uint16) uint16) {
	var value1, value2, zeroFlag, result uint16
	value1 = load(ActiveArgs[0])
	if value2 = ActiveArgs[1]; ActiveFlag != FLAG_RR {
		value2 = load(ActiveArgs[1])
	}
	result = base(value1, value2)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	store(ZERO_FLAG, zeroFlag)
	store(CARRY_FLAG, 0)
	store(ActiveArgs[0], result)
}

func RunArithmetic(base func(uint16, uint16) uint16, carry func(int, int) int) {
	var value1, value2, result, zeroFlag, carryFlag uint16
	value1 = load(ActiveArgs[0])
	if value2 = ActiveArgs[1]; ActiveFlag != FLAG_RR {
		value2 = load(ActiveArgs[1])
	}
	result = base(value1, value2)
	carryResult := carry(int(value1), int(value2))
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	store(ZERO_FLAG, zeroFlag)
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	store(CARRY_FLAG, carryFlag)
	store(ActiveArgs[0], result)
}

func Btouint(b bool) uint16 {
	if b {
		return 1
	}
	return 0
}
func executeActive() {
	switch ActiveCommand {
	case CMD_ADD:
		RunArithmetic(func(a, b uint16) uint16 { return a + b }, func(a, b int) int { return a + b })
	case CMD_SUB:
		RunArithmetic(func(a, b uint16) uint16 { return a - b }, func(a, b int) int { return a - b })
	case CMD_MUL:
		RunArithmetic(func(a, b uint16) uint16 { return a * b }, func(a, b int) int { return a * b })
	case CMD_DIV:
		RunArithmetic(func(a, b uint16) uint16 { return a / b }, func(a, b int) int { return a / b })
	case CMD_INC:
		RunMinArithmetic(func(a uint16) uint16 { return a + 1 }, func(a int) int { return a + 1 })
	case CMD_DEC:
		RunMinArithmetic(func(a uint16) uint16 { return a - 1 }, func(a int) int { return a - 1 })
	case CMD_AND:
		RunLogic(func(a, b uint16) uint16 { return a & b })
	case CMD_OR:
		RunLogic(func(a, b uint16) uint16 { return a | b })
	case CMD_XOR:
		RunLogic(func(a, b uint16) uint16 { return a ^ b })
	case CMD_NOT:
		RunMinLogic(func(a uint16) uint16 { return a &^ 0xFFFF })
	case CMD_SHL:
		RunLogic(func(a, b uint16) uint16 { return a << b })
	case CMD_SHR:
		RunLogic(func(a, b uint16) uint16 { return a >> b })
	case CMD_MOV:
		var value uint16
		switch ActiveFlag {
		case FLAG_RR, FLAG_RA, FLAG_AA, FLAG_AR:
			value = load(ActiveArgs[0])
		case FLAG_IA, FLAG_IR:
			value = ActiveArgs[0]
		}
		store(ActiveArgs[1], value)
	case CMD_PUSH:
		var value uint16
		switch ActiveFlag {
		case FLAG_I:
			value = ActiveArgs[0]
		case FLAG_R:
			value = ActiveArgs[1]
		}
		pushOnStack(value)
	case CMD_POP:
		value := popFromStack()
		store(ActiveArgs[0], value)
	case CMD_CMP:
		RunLogic(func(a, b uint16) uint16 { return Btouint(a == b) })
	case CMD_CNT:
		RunLogic(func(a, b uint16) uint16 { return Btouint(a != b) })
	case CMD_LGE:
		RunLogic(func(a, b uint16) uint16 { return Btouint(a >= b) })
	case CMD_SME:
		RunLogic(func(a, b uint16) uint16 { return Btouint(a <= b) })
	case CMD_JIF:
		var value uint16
		if ActiveFlag != FLAG_I {
			value = load(ActiveArgs[0])
		} else {
			value = ActiveArgs[0]
		}
		if load(ZERO_FLAG) == 1 {
			store(CODE_POINTER, value)
		}
	case CMD_JMP:
		var value uint16
		if ActiveFlag != FLAG_I {
			value = load(ActiveArgs[0])
		} else {
			value = ActiveArgs[0]
		}
		store(CODE_POINTER, value)
	case CMD_CLL:
		var value uint16
		if ActiveFlag != FLAG_I {
			value = load(ActiveArgs[0])
		} else {
			value = ActiveArgs[0]
		}
		pushOnStack(value)
		store(CODE_POINTER, value)
	case CMD_RET:
		value := popFromStack()
		store(CODE_POINTER, value)
	case CMD_HLT:
		ShutDown = true
	}
}

func evaluate() {
	NextCommand = loadNextCommand()
	for !ShutDown {
		ActiveFlag = NextCommand & FLAG_MASK
		ActiveCommand = NextCommand & CMD_MASK
		switch ActiveFlag {
		case FLAG_RR, FLAG_RI, FLAG_RA, FLAG_AA, FLAG_AR, FLAG_IA, FLAG_IR:
			ActiveArgs[0] = loadNextCommand()
			ActiveArgs[1] = loadNextCommand()
		case FLAG_I, FLAG_R:
			ActiveArgs[0] = loadNextCommand()
		}
		executeActive()
		NextCommand = loadNextCommand()
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
