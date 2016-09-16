package main

func (machine *Machine) PerformSimpleArithmetic(base func(uint16) uint16, carry func(int) int) {
	var value1, result, zeroFlag, carryFlag uint16
	value1 = machine.load(machine.Args[0])
	result = base(value1)
	carryResult := carry(int(value1))
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	machine.store(ZERO_FLAG, zeroFlag)
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	machine.store(CARRY_FLAG, carryFlag)
	machine.store(machine.Args[0], result)
}

func (machine *Machine) PerformSimpleLogic(base func(uint16) uint16) {
	var value1, zeroFlag, result uint16
	value1 = machine.load(machine.Args[0])
	result = base(value1)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	machine.store(ZERO_FLAG, zeroFlag)
	machine.store(CARRY_FLAG, 0)
	machine.store(machine.Args[0], result)
}
func (machine *Machine) PerformLogic(base func(uint16, uint16) uint16) {
	var value1, value2, zeroFlag, result uint16
	value1 = machine.load(machine.Args[0])
	if value2 = machine.Args[1]; machine.Flag == FLAG_RR {
		value2 = machine.load(machine.Args[1])
	}
	result = base(value1, value2)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	machine.store(ZERO_FLAG, zeroFlag)
	machine.store(CARRY_FLAG, 0)
	machine.store(machine.Args[0], result)
}

func (machine *Machine) PerformArithmetic(base func(uint16, uint16) uint16, carry func(int, int) int) {
	var value1, value2, result, zeroFlag, carryFlag uint16
	value1 = machine.load(machine.Args[0])
	if value2 = machine.Args[1]; machine.Flag == FLAG_RR {
		value2 = machine.load(machine.Args[1])
	}
	result = base(value1, value2)
	carryResult := carry(int(value1), int(value2))
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	machine.store(ZERO_FLAG, zeroFlag)
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	machine.store(CARRY_FLAG, carryFlag)
	machine.store(machine.Args[0], result)
}
