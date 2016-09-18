; a script testing calling and returning
JMP main
add:
	ADD AX BX
	RET
main:
	MOV 3 AX
	MOV 4 BX
	CALL add
	HLT
