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

type asyncInterrupt struct {
	Identifier, Reason uint16
}

type Machine struct {
	Memory
	next        uint16
	flag        uint16
	command     uint16
	args        [MAX_CMD_ARGS]uint16
	keepRunning bool
	debug       bool
	display     Display
	irQueue     chan asyncInterrupt
}

type machineError struct {
	prefix string
	source error
}

func (me machineError) Error() string {
	return me.prefix + ": " + me.source.Error()
}

func New() *Machine {
	return &Machine{
		Memory:  NewMemory(int(MAX_MEMORY) + 1),
		display: TextDisplay{},
	}
}

func (machine *Machine) EnableDebug(debug bool) {
	machine.debug = debug
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
	err = machine.dispose()
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) dispose() error {
	machine.display.Close()
	return nil
}

func (machine *Machine) Interrupt(code, reason uint16) {
	machine.irQueue <- asyncInterrupt{code, reason}
}

func interruptError(sub error) error {
	return &machineError{"interrupt", sub}
}

func (machine *Machine) updateInterrupts() error {
	var code, reason uint16

	// Fetch latest interrupt
	select {
	case ir, ok := <-machine.irQueue:
		if !ok {
			return interruptError(errors.New("queue closed"))
		}
		code = ir.Identifier
		reason = ir.Reason
	default:
		return nil
	}

	// Store active code pointer on stack
	current, err := machine.Load(CODE_POINTER)
	if err != nil {
		return interruptError(err)
	}
	err = machine.push(current)
	if err != nil {
		return interruptError(err)
	}
	// Store interrupt code in register
	err = machine.Store(INTERRUPT, code)
	if err != nil {
		return interruptError(err)
	}
	// Load interrupt handler
	pointer, err := machine.Load(reason)
	if err != nil {
		return interruptError(err)
	}
	// Jump to interrupt handler
	err = machine.Store(CODE_POINTER, pointer)
	if err != nil {
		return interruptError(err)
	}
	return nil
}

func stackError(sub error) error {
	return &machineError{"stack", sub}
}

func (machine *Machine) push(value uint16) error {
	stackItem, err := machine.Load(STACK_POINTER)
	if err != nil {
		return stackError(err)
	}
	if stackItem > STACK_MAX-WORD_SIZE {
		machine.Interrupt(IR_OVERFLOW_STACK, IR_OVERFLOW)
		return stackError(errors.New("memory overflow"))
	}
	nextItem := stackItem + WORD_SIZE
	err = machine.Store(nextItem, value)
	if err != nil {
		return stackError(err)
	}
	machine.Store(STACK_POINTER, nextItem)
	if err != nil {
		return stackError(err)
	}
	return nil
}

func (machine *Machine) pop() (uint16, error) {
	var value uint16

	pointer, err := machine.Load(STACK_POINTER)
	if err != nil {
		return value, stackError(err)
	}
	value, err = machine.Load(pointer)
	if err != nil {
		return value, stackError(err)
	}
	err = machine.Store(pointer, 0)
	if err != nil {
		return value, stackError(err)
	}
	if pointer <= STACK_BASE {
		return value, nil
	}
	err = machine.Store(STACK_POINTER, pointer-WORD_SIZE)
	if err != nil {
		return value, stackError(err)
	}
	return value, nil
}

func (machine *Machine) fetchWord() (uint16, error) {
	pointer, err := machine.Load(CODE_POINTER)
	if err != nil {
		return 0, err
	}
	if pointer > MAX_MEMORY-WORD_SIZE {
		machine.Interrupt(IR_OVERFLOW_CODE, IR_OVERFLOW)
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
	var err error
	switch machine.command {
	case CMD_ADD:
		err = machine.PerformArithmetic(func(a, b int) int { return a + b })
	case CMD_SUB:
		err = machine.PerformArithmetic(func(a, b int) int { return a - b })
	case CMD_MUL:
		err = machine.PerformArithmetic(func(a, b int) int { return a * b })
	case CMD_DIV:
		err = machine.PerformArithmetic(func(a, b int) int { return a / b })
	case CMD_INC:
		err = machine.PerformSimpleArithmetic(func(a int) int { return a + 1 })
	case CMD_DEC:
		err = machine.PerformSimpleArithmetic(func(a int) int { return a - 1 })
	case CMD_AND:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return a & b })
	case CMD_OR:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return a | b })
	case CMD_XOR:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return a ^ b })
	case CMD_NOT:
		err = machine.PerformSimpleLogic(func(a uint16) uint16 { return a &^ 0xFFFF })
	case CMD_SHL:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return a << b })
	case CMD_SHR:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return a >> b })
	case CMD_MOV:
		err = machine.PerformMove()
	case CMD_PUSH:
		err = machine.PerformPush()
	case CMD_POP:
		err = machine.PerformPop()
	case CMD_CMP:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a == b) })
	case CMD_CNT:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a != b) })
	case CMD_LGE:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a >= b) })
	case CMD_SME:
		err = machine.PerformLogic(func(a, b uint16) uint16 { return toUint16(a <= b) })
	case CMD_JIF:
		err = machine.PerformJump(false)
	case CMD_JMP:
		err = machine.PerformJump(true)
	case CMD_CALL:
		err = machine.PerformCall()
	case CMD_RET:
		err = machine.PerformReturn()
	case CMD_HLT:
		machine.Halt()
	}
	return err
}

func (machine *Machine) parseState() error {
	machine.flag = machine.next & FLAG_MASK
	machine.command = machine.next & CMD_MASK

	var err error
	maxFlags := FlagSize[machine.flag]
	for i := 0; i < maxFlags; i++ {
		machine.args[i], err = machine.fetchWord()
		if err != nil {
			return err
		}
	}

	if machine.debug {
		fmt.Printf("%4.4X %X\n", machine.next, machine.args)
	}

	return nil
}

func iterationError(sub error) error {
	return &machineError{"iterate", sub}
}
func (machine *Machine) iterate() error {
	var err error
	if machine.debug {
		pointer, err := machine.Load(CODE_POINTER)
		if err != nil {
			return iterationError(err)
		}
		fmt.Printf("%4.4X: ", pointer)
	}

	machine.next, err = machine.fetchWord()
	if err != nil {
		return iterationError(err)
	}
	return nil
}

func runtimeError(sub error) error {
	return &machineError{"runtime", sub}
}

func (machine *Machine) run() error {
	err := machine.iterate()
	if err != nil {
		return runtimeError(err)
	}

	for machine.keepRunning {
		machine.parseState()
		err = machine.handle()
		if err != nil {
			return runtimeError(err)
		}
		err = machine.updateInterrupts()
		if err != nil {
			return runtimeError(err)
		}
		err = machine.iterate()
		if err != nil {
			return runtimeError(err)
		}
		machine.display.Draw(80, 24, machine.Segment(OUT_CHARS, OUT_MODE))
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
	machine.display.Init()
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

	machine.keepRunning = true
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
