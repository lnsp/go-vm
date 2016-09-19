; a collatz problem in GVM ASM
MOV 6 DX
PUSH 1
JMP main
mod:
	MOV AX CX
	; use int division to calculate mod
	DIV AX BX
	; AX - intdiv(AX, BX) * BX = CX
	MUL AX BX
	SUB CX AX
	RET
main:
	MOV DX BX
	CNT BX 1
	JIF halt
	MOV 2 BX
	MOV DX AX
	CALL mod
	CMP CX 0
	JIF if
else:
	DIV DX 2
	JMP loop
if:
	MUL DX 3
	ADD DX 1
loop:
	; count iterations
	POP CX
	INC CX
	PUSH CX
	JMP main
halt:
	; copy iterations to AX
	MOV CX AX
	HLT
