; display test
PUSH 65
MOV OCH BX
loop:
	POP AX
	MOV AX [BX]
	INC AX
	INC BX
	INC BX
	PUSH AX
	CMP AX 91
	JIF loop
	HLT

