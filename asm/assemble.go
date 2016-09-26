// Package asm provides assembly generators for the govm.
package asm

import (
	"fmt"
	"github.com/lnsp/govm/vm"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

const (
	// No argument given
	ARG_NONE = iota
	// Register argument
	ARG_REGISTER
	// Address argument
	ARG_ADDRESS
	// Immediate argument
	ARG_IMMEDIATE
)

var (
	argMap = map[uint16]map[int]uint16{
		vm.FLAG_NONE: map[int]uint16{
			ARG_REGISTER:  vm.FLAG_R,
			ARG_IMMEDIATE: vm.FLAG_I,
			ARG_ADDRESS:   vm.FLAG_A,
		},
		vm.FLAG_I: map[int]uint16{
			ARG_REGISTER:  vm.FLAG_IR,
			ARG_IMMEDIATE: vm.FLAG_II,
			ARG_ADDRESS:   vm.FLAG_IA,
		},
		vm.FLAG_R: map[int]uint16{
			ARG_REGISTER:  vm.FLAG_RR,
			ARG_IMMEDIATE: vm.FLAG_RI,
			ARG_ADDRESS:   vm.FLAG_RA,
		},
		vm.FLAG_A: map[int]uint16{
			ARG_REGISTER:  vm.FLAG_AR,
			ARG_IMMEDIATE: vm.FLAG_AI,
			ARG_ADDRESS:   vm.FLAG_AA,
		},
	}
	commandMap = map[string]uint16{
		"ADD":  vm.CMD_ADD,
		"SUB":  vm.CMD_SUB,
		"MUL":  vm.CMD_MUL,
		"DIV":  vm.CMD_DIV,
		"INC":  vm.CMD_INC,
		"DEC":  vm.CMD_DEC,
		"AND":  vm.CMD_AND,
		"OR":   vm.CMD_OR,
		"XOR":  vm.CMD_XOR,
		"NOT":  vm.CMD_NOT,
		"SHL":  vm.CMD_SHL,
		"SHR":  vm.CMD_SHR,
		"MOV":  vm.CMD_MOV,
		"PUSH": vm.CMD_PUSH,
		"POP":  vm.CMD_POP,
		"CMP":  vm.CMD_CMP,
		"CNT":  vm.CMD_CNT,
		"LGE":  vm.CMD_LGE,
		"SME":  vm.CMD_SME,
		"JIF":  vm.CMD_JIF,
		"JMP":  vm.CMD_JMP,
		"CALL": vm.CMD_CALL,
		"RET":  vm.CMD_RET,
		"HLT":  vm.CMD_HLT,
	}
	registerMap = map[string]uint16{
		"AX":  vm.REGISTER_AX,
		"BX":  vm.REGISTER_BX,
		"CX":  vm.REGISTER_CX,
		"DX":  vm.REGISTER_DX,
		"IR":  vm.INTERRUPT,
		"IRS": vm.IR_STATE,
		"IRK": vm.IR_KEYBOARD,
		"IRO": vm.IR_OVERFLOW,
		"SB":  vm.STACK_BASE,
		"CP":  vm.CODE_POINTER,
		"SP":  vm.STACK_POINTER,
		"ZF":  vm.ZERO_FLAG,
		"CF":  vm.CARRY_FLAG,
	}
	systemPointers = map[string]uint16{
		"SM":  vm.STACK_MAX,
		"OCH": vm.OUT_CHARS,
		"OCL": vm.OUT_COLORS,
		"CB":  vm.CODE_BASE,
		"OMD": vm.OUT_MODE,
	}
	pointerRegex, _ = regexp.Compile("[a-zA-Z]+")
	numberRegex, _  = regexp.Compile("((0x[0-9a-fA-F]+)|(0[0-7]+)|([0-9]+))")
)

// PointerReference represents a assembler reference to a point in memory.
type PointerReference struct {
	Name      string
	Line, Arg int
}

// Assemble generates bytecode from GOVM ASM.
func Assemble(code string) []byte {
	var references []PointerReference
	var lineBuffer [][]uint16
	var lineDebug []string
	definedPointers := make(map[string]uint16)
	cleanLines := CleanCode(code)

	// Build bytecode as normal
	for i := 0; i < len(cleanLines); i++ {
		active := cleanLines[i]
		if active == "" {
			continue
		}

		// Handle marker
		if strings.HasSuffix(active, ":") {
			definedPointers[strings.TrimRight(active, ":")] = uint16(len(lineBuffer))
			continue
		}

		tokens := strings.Split(active, " ")
		cmd := strings.ToUpper(tokens[0])
		var result []uint16
		switch cmd {
		case "DB":
			if strings.HasPrefix(tokens[1], "\"") {
				result = ParseString(active[3:])
			} else {
				result = []uint16{ParseNumber(active[3:])}
			}
		default:
			data, refs := ParseCommand(tokens, len(lineBuffer))
			result = data
			references = append(references, refs...)
		}
		lineBuffer = append(lineBuffer, result)
		lineDebug = append(lineDebug, active)
	}

	mappedBytes := make([]uint16, 0)
	byteCount := uint16(0)
	for _, line := range lineBuffer {
		mappedBytes = append(mappedBytes, byteCount)
		byteCount += uint16(len(line) * 2)
	}

	for name, ptr := range definedPointers {
		definedPointers[name] = mappedBytes[ptr] + vm.CODE_BASE
	}

	for _, p := range references {
		if _, ok := definedPointers[p.Name]; !ok {
			fmt.Printf("ERROR: Missing pointer %s\n", p.Name)
			return []byte{}
		}
		lineBuffer[p.Line][p.Arg+1] = definedPointers[p.Name]
	}

	for index, command := range lineDebug {
		fmt.Printf("%4.4X %-24s %4.4X\n", mappedBytes[index]+vm.CODE_BASE, command, lineBuffer[index])
	}

	var output []byte
	for _, line := range lineBuffer {
		data := make([]byte, len(line)*2)
		for i, e := range line {
			vm.ByteOrder.PutUint16(data[i*2:i*2+2], e)
		}
		output = append(output, data...)
	}

	return output
}

// ParseCommand parses a specific command and returns a word representation and a slice of pointers.
func ParseCommand(args []string, line int) ([]uint16, []PointerReference) {
	cmdMap, ok := commandMap[args[0]]
	if !ok {
		fmt.Printf("ERROR: Unknown command %s\n", args[0])
		return []uint16{}, []PointerReference{}
	}
	cmd := []uint16{cmdMap}
	flag := vm.FLAG_NONE
	pointers := make([]PointerReference, 0)

	for index, arg := range args[1:] {
		argType := ARG_NONE
		if strings.HasPrefix(arg, "[") {
			arg = strings.Trim(arg, "[]")
			argType = ARG_ADDRESS
		}

		var argValue uint16
		if numberRegex.MatchString(arg) {
			argValue = ParseNumber(arg)
			argType = ARG_IMMEDIATE
		} else if pointerRegex.MatchString(arg) {
			if v, ok := registerMap[arg]; ok {
				if argType == ARG_NONE {
					argType = ARG_REGISTER
				}
				argValue = v
			} else if v, ok := systemPointers[arg]; ok {
				argValue = v
				argType = ARG_IMMEDIATE
			} else {
				pointers = append(pointers, PointerReference{arg, line, index})
				argType = ARG_IMMEDIATE
			}
		}

		cmd = append(cmd, argValue)
		flag = argMap[flag][argType]
	}

	cmd[0] = cmd[0] | flag
	return cmd, pointers
}

// ParseString converts a string into a utf-16 encoded slice of words.
func ParseString(str string) []uint16 {
	str = strings.Trim(str, "\"")
	return utf16.Encode([]rune(str))
}

// ParseNumber converts a hex-, octal- or decimal number into a word.
func ParseNumber(str string) uint16 {
	var result uint16
	if strings.HasPrefix(str, "0x") {
		trimmed := strings.TrimLeft(str, "0x")
		parsed, _ := strconv.ParseUint(trimmed, 16, 16)
		result = uint16(parsed)
	} else if strings.HasPrefix(str, "0") {
		trimmed := strings.TrimLeft(str, "0")
		parsed, _ := strconv.ParseUint(trimmed, 8, 16)
		result = uint16(parsed)
	} else {
		parsed, _ := strconv.ParseUint(str, 10, 16)
		result = uint16(parsed)
	}
	return result
}

// CleanCode remove comments, trim lines etc.
func CleanCode(code string) []string {
	lines := strings.Split(code, "\n")
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, ";") {
			trimmed = ""
		}
		lines[index] = trimmed
	}
	return lines
}
