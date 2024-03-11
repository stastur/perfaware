package main

func EstimateCycles(op Operand) (cycles int) {
	if ea, ok := op.(OperandEffectiveAddress); ok {
		if ea.Disp != 0 {
			cycles += 4
		}

		switch ea.Base {
		case "bx+si":
			cycles += 7
		case "bp+di":
			cycles += 7
		case "bx+di":
			cycles += 8
		case "bp+si":
			cycles += 8

		default:
			cycles += 5
		}
	}

	if _, ok := op.(OperandDirectAddress); ok {
		cycles = 6
	}

	return
}

func (inst Instruction) EstimateCycles() (cycles int) {
	switch inst.Op {
	case "mov":
		switch left := inst.Operands[0].(type) {
		case OperandRegister:
			switch right := inst.Operands[1].(type) {
			case OperandRegister:
				if right.Index == RI_a {
					cycles = 10
					break
				}
				cycles = 2

			case OperandImmediate:
				cycles = 4

			case OperandEffectiveAddress, OperandDirectAddress:
				cycles = 8 + EstimateCycles(right)
			}

		case OperandEffectiveAddress:
			switch right := inst.Operands[1].(type) {
			case OperandRegister:
				if right.Index == RI_a {
					cycles = 10
					break
				}
				cycles = 9 + EstimateCycles(left)

			case OperandImmediate:
				cycles = 10 + EstimateCycles(left)

			case OperandEffectiveAddress, OperandDirectAddress:
			}
		}

	case "add", "sub":
		switch left := inst.Operands[0].(type) {
		case OperandRegister:
			switch right := inst.Operands[1].(type) {
			case OperandRegister:
				cycles = 3

			case OperandImmediate:
				cycles = 4

			case OperandEffectiveAddress, OperandDirectAddress:
				cycles = 9 + EstimateCycles(right)
			}

		case OperandEffectiveAddress, OperandDirectAddress:
			switch inst.Operands[1].(type) {
			case OperandRegister:
				cycles = 16 + EstimateCycles(left)

			case OperandImmediate:
				cycles = 17 + EstimateCycles(left)

			case OperandEffectiveAddress, OperandDirectAddress:
			}

		}

	case "cmp":
		switch left := inst.Operands[0].(type) {
		case OperandRegister:
			switch right := inst.Operands[1].(type) {
			case OperandRegister:
				cycles = 3

			case OperandImmediate:
				cycles = 4

			case OperandEffectiveAddress, OperandDirectAddress:
				cycles = 9 + EstimateCycles(right)
			}

		case OperandEffectiveAddress, OperandDirectAddress:
			switch inst.Operands[1].(type) {
			case OperandRegister:
				cycles = 9 + EstimateCycles(left)

			case OperandImmediate:
				cycles = 10 + EstimateCycles(left)

			case OperandEffectiveAddress, OperandDirectAddress:
			}

		}
	}

	return
}
