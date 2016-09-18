; testing memory access
JMP main
addr:
	DB 0x2004
main:
	MOV addr AX
	MOV 0x0016 [AX]
	JMP addr
