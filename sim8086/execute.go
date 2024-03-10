package main

import "fmt"

type Memory []byte
type Registers []int16

func ExecuteIntruction(inst Instruction, registers Registers, memory Memory) {
	dest := inst.Operands[0]
	source := inst.Operands[1]

	var left int16
	var right int16

	if dest != nil {
		left = GetOperandValue(dest, registers, memory)
	}
	if source != nil {
		right = GetOperandValue(source, registers, memory)
	}

	switch inst.Op {
	case "mov":
		SetOperandValue(dest, right, registers, memory)

	case "add":
		value := left + right
		SetOperandValue(dest, value, registers, memory)
		UpdateFlagsRegister(value, registers)

	case "sub":
		value := left - right
		SetOperandValue(dest, value, registers, memory)
		UpdateFlagsRegister(value, registers)

	case "cmp":
		value := left - right
		UpdateFlagsRegister(value, registers)

	case "jne":
		isZero := registers[RI_flags]&(1<<RF_zero) == (1 << RF_zero)
		if !isZero {
			registers[RI_ip] += int16(int8(right))
		}
	}

	if dest != nil {
		before := left
		after := GetOperandValue(dest, registers, memory)

		fmt.Printf("; %s 0x%04x->0x%04x\n", dest.String(), before, after)
	}

	return
}

func GetRegisterValue(operand OperandRegister, registers Registers) int16 {
	// only wide registers
	// TODO: support _h _l registers
	return registers[operand.Index]
}

func GetOperandValue(operand Operand, registers Registers, memory Memory) int16 {
	switch op := operand.(type) {
	case OperandImmediate:
		return int16(op.Value)

	case OperandRegister:
		return GetRegisterValue(op, registers)

	case OperandDirectAddress:
		lo := memory[op]
		hi := memory[op+1]
		return (int16(hi) << 8) | int16(lo)

	case OperandEffectiveAddress:
		ea := EvalEffectiveAddress(op, registers, memory)
		if op.Wide {
			lo := memory[ea]
			hi := memory[ea+1]
			return (int16(hi) << 8) | int16(lo)
		}

		return int16(memory[ea])
	}

	return 0
}

func SetOperandValue(operand Operand, value int16, registers Registers, memory Memory) {
	switch op := operand.(type) {
	case OperandRegister:
		// only wide registers
		registers[op.Index] = value

	case OperandDirectAddress:
		memory[op] = byte(0x0f & value)
		memory[op+1] = byte(value >> 8)

	case OperandEffectiveAddress:
		ea := EvalEffectiveAddress(op, registers, memory)
		memory[ea] = byte(0xff & value)
		memory[ea+1] = byte(value >> 8)
	}
}

func EvalEffectiveAddress(op OperandEffectiveAddress, registers Registers, memory Memory) int16 {
	var address int16

	bx := OperandRegister{RI_b, 0, 2}
	bp := OperandRegister{RI_bp, 0, 2}
	si := OperandRegister{RI_si, 0, 2}
	di := OperandRegister{RI_di, 0, 2}

	switch op.Base {
	case "bx+si":
		address = GetRegisterValue(bx, registers) + GetRegisterValue(si, registers)
	case "bx+di":
		address = GetRegisterValue(bx, registers) + GetRegisterValue(di, registers)
	case "bp+si":
		address = GetRegisterValue(bp, registers) + GetRegisterValue(si, registers)
	case "bp+di":
		address = GetRegisterValue(bp, registers) + GetRegisterValue(di, registers)
	case "si":
		address = GetRegisterValue(si, registers)
	case "di":
		address = GetRegisterValue(di, registers)
	case "bp":
		address = GetRegisterValue(bp, registers)
	case "bx":
		address = GetRegisterValue(bx, registers)
	}

	return address + op.Disp
}

func PrintFlags(flags int16) {
	strFlags := [RF_Count]string{
		RF_zero: "Z",
		RF_sign: "S",
	}

	fmt.Print("; Flags: ")
	for i, f := range strFlags {
		if flags&(1<<i) == (1 << i) {
			fmt.Print(f)
		}
	}
	fmt.Println()
}

func UpdateFlagsRegister(value int16, registers Registers) {
	registers[RI_flags] &= 0
	registers[RI_flags] |= BoolToInt(value == 0) << RF_zero
	registers[RI_flags] |= BoolToInt(value < 0) << RF_sign

	PrintFlags(registers[RI_flags])
}

func BoolToInt(value bool) int16 {
	if value {
		return 1
	}

	return 0
}

func (registers Registers) Print() {
	printOrder := []RegisterIndex{RI_a, RI_b, RI_c, RI_d, RI_sp, RI_bp, RI_si, RI_di, RI_ip}

	fmt.Println("; Registers")
	for _, idx := range printOrder {
		v := registers[idx]
		if v == 0 {
			continue
		}
		reg := OperandRegister{idx, 0, 2}
		fmt.Printf(";   %s: 0x%04x (%d)\n", reg, v, v)
	}

	PrintFlags(registers[RI_flags])
}
