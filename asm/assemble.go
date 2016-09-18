package asm

import (
	"fmt"
	"github.com/lnsp/gvm/vm"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf16"
)

var (
	CommandMap = map[string]uint16{
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
	Registers = map[string]uint16{
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
	SystemPointers = map[string]uint16{
		"SM":  vm.STACK_MAX,
		"OCH": vm.OUT_CHARS,
		"OCL": vm.OUT_COLORS,
		"CB":  vm.CODE_BASE,
		"OMD": vm.OUT_MODE,
	}
	PointerPattern, _ = regexp.Compile("[a-zA-Z]+")
	NumberPattern, _  = regexp.Compile("((0x[0-9a-fA-F]+)|(0[0-7]+)|([0-9]+))")
)

type PointerReference struct {
	Name      string
	Line, Arg int
}

func Assemble(code string) []byte {
	var references []PointerReference
	var lineBuffer [][]uint16
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
			fmt.Println("push", active, "to defined pointers")
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
		fmt.Printf("%s %X\n", cmd, result)
		lineBuffer = append(lineBuffer, result)
	}

	mappedBytes := make([]uint16, 0)
	byteCount := uint16(0)
	for _, line := range lineBuffer {
		mappedBytes = append(mappedBytes, byteCount)
		byteCount += uint16(len(line) * 2)
	}
	fmt.Printf("mapped bytes %d\n", mappedBytes)

	for name, ptr := range definedPointers {
		definedPointers[name] = mappedBytes[ptr] + vm.CODE_BASE
		fmt.Printf("Mapping %s to %X\n", name, definedPointers[name])
	}

	for _, p := range references {
		lineBuffer[p.Line][p.Arg+1] = definedPointers[p.Name]
		fmt.Println("replaced", p.Name)
	}

	var output []byte
	for _, line := range lineBuffer {
		data := make([]byte, len(line)*2)
		for i, e := range line {
			vm.ByteOrder.PutUint16(data[i*2:i*2+2], e)
		}
		fmt.Printf("Converted %X to %X\n", line, data)
		output = append(output, data...)
	}
	return output
}

func ParseCommand(args []string, line int) ([]uint16, []PointerReference) {
	cmdMap := CommandMap[args[0]]
	cmd := []uint16{cmdMap}
	flag := vm.FLAG_NONE
	pointers := make([]PointerReference, 0)

	fmt.Println("parsing ", args, cmdMap)

	for index, arg := range args[1:] {
		var argType string
		if strings.HasPrefix(arg, "[") {
			arg = strings.Trim(arg, "[]")
			argType = "address"
		}

		var argValue uint16
		if NumberPattern.MatchString(arg) {
			argValue = ParseNumber(arg)
			argType = "intermediate"
		} else if PointerPattern.MatchString(arg) {
			if v, ok := Registers[arg]; ok {
				if argType == "" {
					argType = "register"
				}
				argValue = v
			} else {
				if v, ok := SystemPointers[arg]; ok {
					argValue = v
				} else {
					pointers = append(pointers, PointerReference{arg, line, index})
					argType = "intermediate"
				}
			}
		}
		fmt.Println(index, arg, argValue, argType)

		cmd = append(cmd, argValue)
		fmt.Println(cmd)

		switch flag {
		case vm.FLAG_NONE:
			switch argType {
			case "intermediate":
				flag = vm.FLAG_I
			case "register":
				flag = vm.FLAG_R
			case "address":
				flag = vm.FLAG_A
			}
		case vm.FLAG_I:
			switch argType {
			case "intermediate":
				flag = vm.FLAG_II
			case "register":
				flag = vm.FLAG_IR
			case "address":
				flag = vm.FLAG_IA
			}
		case vm.FLAG_R:
			switch argType {
			case "intermediate":
				flag = vm.FLAG_RI
			case "register":
				flag = vm.FLAG_RR
			case "address":
				flag = vm.FLAG_RA
			}
		case vm.FLAG_A:
			switch argType {
			case "intermediate":
				flag = vm.FLAG_AI
			case "register":
				flag = vm.FLAG_AR
			case "address":
				flag = vm.FLAG_AA
			}
		}
	}

	cmd[0] = cmd[0] | flag
	return cmd, pointers
}

func ParseString(str string) []uint16 {
	str = strings.Trim(str, "\"")
	return utf16.Encode([]rune(str))
}

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

// Remove comments, trim lines etc.
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
