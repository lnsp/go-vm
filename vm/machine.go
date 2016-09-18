package vm

import (
	"encoding/binary"
	"errors"
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
	Memory
	Next        uint16
	Flag        uint16
	Command     uint16
	Args        [MAX_CMD_ARGS]uint16
	KeepRunning bool
	Debug       bool
}

func New() *Machine {
	return &Machine{
		Memory: NewMemory(int(MAX_MEMORY) + 1),
	}
}

func (machine *Machine) EnableDebug(debug bool) {
	machine.Debug = debug
}

func (machine *Machine) Boot(code []byte) error {
	err := machine.initialize()
	if err != nil {
		return err
	}
	err = machine.program(code)
	if err != nil {
		return err
	}
	err = machine.run()
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) interrupt(code, reason uint16) error {
	err := machine.Store(INTERRUPT, code)
	if err != nil {
		return err
	}
	pointer, err := machine.Load(reason)
	if err != nil {
		return err
	}
	err = machine.Store(CODE_POINTER, pointer)
	if err != nil {
		return err
	}
	err = machine.run()
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) push(value uint16) error {
	stackItem, err := machine.Load(STACK_POINTER)
	if err != nil {
		return err
	}
	if stackItem > STACK_MAX-WORD_SIZE {
		machine.interrupt(IR_OVERFLOW_STACK, IR_OVERFLOW)
		return errors.New("stack overflow")
	}
	nextItem := stackItem + WORD_SIZE
	err = machine.Store(nextItem, value)
	if err != nil {
		return err
	}
	machine.Store(STACK_POINTER, nextItem)
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) pop() (uint16, error) {
	pointer, err := machine.Load(STACK_POINTER)
	if err != nil {
		return 0, err
	}
	value, err := machine.Load(pointer)
	if err != nil {
		return 0, err
	}
	err = machine.Store(pointer, 0)
	if err != nil {
		return 0, err
	}
	if pointer > STACK_BASE {
		machine.Store(STACK_POINTER, pointer-WORD_SIZE)
	}
	return value, nil
}

func (machine *Machine) fetchWord() (uint16, error) {
	pointer, err := machine.Load(CODE_POINTER)
	if err != nil {
		return 0, err
	}
	if pointer > MAX_MEMORY-WORD_SIZE {
		machine.interrupt(IR_OVERFLOW_CODE, IR_OVERFLOW)
		return 0, errors.New("out of memory")
	}
	err = machine.Store(CODE_POINTER, pointer+WORD_SIZE)
	if err != nil {
		return 0, err
	}
	word, err := machine.Load(pointer)
	if err != nil {
		return 0, err
	}
	return word, nil
}

func (machine *Machine) handle() error {
	switch machine.Command {
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
		machine.PerformPush()
	case CMD_POP:
		machine.PerformPop()
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
		machine.PerformCall()
	case CMD_RET:
		machine.PerformReturn()
	case CMD_HLT:
		machine.Halt()
	}
	return nil
}

func (machine *Machine) parseState() error {
	machine.Flag = machine.Next & FLAG_MASK
	machine.Command = machine.Next & CMD_MASK

	var err error
	maxFlags := FlagSize[machine.Flag]
	for i := 0; i < maxFlags; i++ {
		machine.Args[i], err = machine.fetchWord()
		if err != nil {
			return err
		}
	}

	if machine.Debug {
		fmt.Printf("%4.4X %X\n", machine.Next, machine.Args)
	}

	return nil
}

func (machine *Machine) iterate() error {
	var err error
	if machine.Debug {
		pointer, err := machine.Load(CODE_POINTER)
		if err != nil {
			return err
		}
		fmt.Printf("%4.4X: ", pointer)
	}

	machine.Next, err = machine.fetchWord()
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) run() error {
	err := machine.iterate()
	if err != nil {
		return err
	}

	for machine.KeepRunning {
		machine.parseState()
		err = machine.handle()
		if err != nil {
			return err
		}
		err = machine.iterate()
		if err != nil {
			return err
		}
	}
	return nil
}

func (machine *Machine) program(code []byte) error {
	size := len(code)

	for i := 0; i < size; i++ {
		err := machine.StoreByte(CODE_BASE+uint16(i), code[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (machine *Machine) initialize() error {
	// Load base values
	err := machine.Store(CODE_POINTER, CODE_BASE)
	if err != nil {
		return err
	}
	err = machine.Store(STACK_POINTER, STACK_BASE)
	if err != nil {
		return err
	}
	err = machine.Store(CODE_BASE, CMD_HLT)
	if err != nil {
		return err
	}
	err = machine.Store(INTERRUPT, MAX_MEMORY)
	if err != nil {
		return err
	}

	// create graphics
	err = machine.Store(OUT_MODE, OUT_MODE_TERM)
	if err != nil {
		return err
	}
	pointer := OUT_COLORS
	for _, color := range BaseColors {
		err = machine.Store(pointer, color)
		if err != nil {
			return err
		}
		pointer += 2
	}

	machine.KeepRunning = true
	return nil
}

func (machine Machine) String() string {
	return machine.dumpSegment(0)
}

func (machine *Machine) dumpSegment(seg int) string {
	start := uint16(seg * 256)
	dump := fmt.Sprintf("SEGMENT %4.4X - %4.4X\n-------------------", start, start+0xFF)
	for i := start; i <= start+0xFF; i += WORD_SIZE {
		if i%16 == 0 {
			dump += "\n"
		}
		word, err := machine.Load(i)
		if err != nil {
			dump += "err"
			return dump
		}
		dump += fmt.Sprintf("%-5.4X", word)
	}
	return dump
}
