package main

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
	FLAG_II   uint16 = 0x0B00
	FLAG_AI   uint16 = 0x0C00
	FLAG_IR   uint16 = 0x0700
	FLAG_I    uint16 = 0x0800
	FLAG_R    uint16 = 0x0900
	FLAG_A    uint16 = 0x0A00
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
	CMD_CALL uint16 = 0x14
	CMD_RET  uint16 = 0x15
	CMD_HLT  uint16 = 0x16

	IR_OVERFLOW_CODE  uint16 = 0x1
	IR_OVERFLOW_STACK uint16 = 0x2
)
