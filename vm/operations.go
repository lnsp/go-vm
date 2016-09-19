package vm

func (machine *Machine) Halt() {
	machine.keepRunning = false
}
func (machine *Machine) PerformPush() error {
	var value uint16
	var err error
	switch machine.flag {
	case FLAG_I:
		value = machine.args[0]
	case FLAG_R:
		value, err = machine.Load(machine.args[0])
		if err != nil {
			return err
		}
	}
	err = machine.push(value)
	if err != nil {
		return err
	}
	return nil
}
func (machine *Machine) PerformPop() error {
	value, err := machine.pop()
	if err != nil {
		return err
	}
	err = machine.Store(machine.args[0], value)
	if err != nil {
		return err
	}
	return nil
}
func (machine *Machine) PerformCall() error {
	var value uint16
	var err error
	switch machine.flag {
	case FLAG_I:
		value = machine.args[0]
	case FLAG_R:
		value, err = machine.Load(machine.args[0])
		if err != nil {
			return err
		}
	}
	current, err := machine.Load(CODE_POINTER)
	if err != nil {
		return err
	}
	err = machine.push(current)
	if err != nil {
		return err
	}
	err = machine.Store(CODE_POINTER, value)
	if err != nil {
		return err
	}
	return nil
}
func (machine *Machine) PerformReturn() error {
	value, err := machine.pop()
	if err != nil {
		return err
	}
	err = machine.Store(CODE_POINTER, value)
	if err != nil {
		return err
	}
	return nil
}
func (machine *Machine) PerformSimpleArithmetic(carry func(int) int) error {
	var value1, result, zeroFlag, carryFlag uint16
	value1, err := machine.Load(machine.args[0])
	if err != nil {
		return err
	}
	carryResult := carry(int(value1))
	result = uint16(carryResult)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	err = machine.Store(ZERO_FLAG, zeroFlag)
	if err != nil {
		return err
	}
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	err = machine.Store(CARRY_FLAG, carryFlag)
	if err != nil {
		return err
	}
	err = machine.Store(machine.args[0], result)
	if err != nil {
		return err
	}

	return nil
}

func (machine *Machine) PerformSimpleLogic(base func(uint16) uint16) error {
	var value1, zeroFlag, result uint16
	value1, err := machine.Load(machine.args[0])
	if err != nil {
		return err
	}
	result = base(value1)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}

	err = machine.Store(ZERO_FLAG, zeroFlag)
	if err != nil {
		return err
	}
	err = machine.Store(CARRY_FLAG, 0)
	if err != nil {
		return err
	}
	err = machine.Store(machine.args[0], result)
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) PerformLogic(base func(uint16, uint16) uint16) error {
	var value1, value2, zeroFlag, result uint16
	value1, err := machine.Load(machine.args[0])
	if err != nil {
		return err
	}
	if value2 = machine.args[1]; machine.flag == FLAG_RR {
		value2, err = machine.Load(machine.args[1])
		if err != nil {
			return err
		}
	}
	result = base(value1, value2)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	err = machine.Store(ZERO_FLAG, zeroFlag)
	if err != nil {
		return err
	}
	err = machine.Store(CARRY_FLAG, 0)
	if err != nil {
		return err
	}
	err = machine.Store(machine.args[0], result)
	if err != nil {
		return err
	}
	return nil
}

func (machine *Machine) PerformArithmetic(carry func(int, int) int) error {
	var value1, value2, result, zeroFlag, carryFlag uint16
	value1, err := machine.Load(machine.args[0])
	if err != nil {
		return err
	}
	if value2 = machine.args[1]; machine.flag == FLAG_RR {
		value2, err = machine.Load(machine.args[1])
		if err != nil {
			return err
		}
	}
	carryResult := carry(int(value1), int(value2))
	result = uint16(carryResult)
	zeroFlag = 0
	if result == 0 {
		zeroFlag = 1
	}
	err = machine.Store(ZERO_FLAG, zeroFlag)
	if err != nil {
		return err
	}
	carryFlag = 0
	if int(result) != carryResult {
		carryFlag = 1
	}
	err = machine.Store(CARRY_FLAG, carryFlag)
	if err != nil {
		return err
	}
	err = machine.Store(machine.args[0], result)
	if err != nil {
		return err
	}
	return nil
}
func (machine *Machine) PerformJump(jumpAlways bool) error {
	var value uint16
	var err error
	switch machine.flag {
	case FLAG_I:
		value = machine.args[0]
	case FLAG_R:
		value, err = machine.Load(machine.args[0])
		if err != nil {
			return err
		}
	}
	zeroFlag, err := machine.Load(ZERO_FLAG)
	if err != nil {
		return err
	}
	if jumpAlways || zeroFlag == 1 {
		err = machine.Store(CODE_POINTER, value)
		if err != nil {
			return err
		}
	}
	return nil
}
func (machine *Machine) PerformMove() error {
	var value, target uint16
	var err error
	switch machine.flag {
	case FLAG_RA:
		value, err = machine.Load(machine.args[0])
		if err != nil {
			return err
		}
		target, err = machine.Load(machine.args[1])
		if err != nil {
			return err
		}
	case FLAG_RR:
		value, err = machine.Load(machine.args[0])
		if err != nil {
			return err
		}
		target = machine.args[1]
		if err != nil {
			return err
		}
	case FLAG_AA:
		pointer, err := machine.Load(machine.args[0])
		if err != nil {
			return err
		}
		value, err = machine.Load(pointer)
		if err != nil {
			return err
		}
		target, err = machine.Load(machine.args[1])
		if err != nil {
			return err
		}
	case FLAG_AR:
		pointer, err := machine.Load(machine.args[0])
		if err != nil {
			return err
		}
		value, err = machine.Load(pointer)
		if err != nil {
			return err
		}
		target = machine.args[1]
	case FLAG_IA:
		value = machine.args[0]
		target, err = machine.Load(machine.args[1])
		if err != nil {
			return err
		}
	case FLAG_IR:
		value = machine.args[0]
		target = machine.args[1]
	}
	err = machine.Store(target, value)
	if err != nil {
		return err
	}
	return nil
}
