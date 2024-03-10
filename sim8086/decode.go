package main

import (
	"errors"
	"fmt"
	"strings"
)

func DecodeInstruction(startingAt int, buff []byte) (*Instruction, error) {
	if startingAt >= len(buff) {
		return nil, errors.New("EOF")
	}

	currentByteIndex := startingAt

	for _, bp := range Blueprints {
		var bitsSet uint32
		var bitsLeft int
		var bytesRead int
		var currentByte byte

		bits := make([]uint16, Bits_Count)

		isValid := true
		for _, part := range bp.Bits {
			if bitsLeft == 0 && part.BitCount != 0 {
				currentByte = buff[currentByteIndex+bytesRead]
				bitsLeft = 8
				bytesRead++
			}

			if part.BitCount > bitsLeft {
				isValid = false
				break
			}

			if part.Type == Bits_Literal {
				constBits := currentByte >> (bitsLeft - part.BitCount)
				constBits &= 0xff >> (8 - part.BitCount)

				if constBits != part.Value {
					isValid = false
					break
				}
			}

			bitsSet |= 1 << part.Type
			bitsLeft -= part.BitCount

			mask := ^byte(0xff << part.BitCount)
			if part.Value != 0 {
				bits[part.Type] = uint16(part.Value)
			} else {
				bits[part.Type] = uint16((currentByte >> bitsLeft) & mask)
			}
		}

		if !isValid {
			continue
		}

		mod := bits[Bits_Mod]
		rm := bits[Bits_Rm]
		w := bits[Bits_W] == 1
		s := bits[Bits_S] == 1

		readFromBuff := func(exists bool, wide bool, signExtended bool) uint16 {
			if !exists {
				return 0
			}

			if wide {
				lo := uint16(buff[currentByteIndex+bytesRead+0])
				hi := uint16(buff[currentByteIndex+bytesRead+1])
				bytesRead += 2

				return hi<<8 | lo
			} else {
				lo := buff[currentByteIndex+bytesRead]
				bytesRead += 1

				if signExtended {
					return uint16(int8(lo))
				}

				return uint16(lo)
			}
		}

		hasDirectAddress := mod == 0b00 && rm == 0b110
		hasDisp := isTypeSet(bitsSet, Bits_HasDisp) &&
			(mod == 0b10 || mod == 0b01 || hasDirectAddress)
		hasData := isTypeSet(bitsSet, Bits_HasData)

		bits[Bits_HasDisp] = readFromBuff(hasDisp, mod == 0b10 || hasDirectAddress, w)
		bits[Bits_HasData] = readFromBuff(hasData, w && !s, s)

		var instruction Instruction

		if isTypeSet(bitsSet, Bits_Mod) {
			instruction.Operands[0] = DecodeRm(rm, mod, w, bits[Bits_HasDisp])
		} else if isTypeSet(bitsSet, Bits_HasAddr) {
			instruction.Operands[0] = OperandDirectAddress(readFromBuff(true, w, false))
		}

		if isTypeSet(bitsSet, Bits_Reg) {
			instruction.Operands[1] = DecodeReg(bits[Bits_Reg], w)
		}

		if bits[Bits_D] == 1 || (isTypeSet(bitsSet, Bits_E) && bits[Bits_E] == 0) {
			instruction.Operands[0], instruction.Operands[1] = instruction.Operands[1], instruction.Operands[0]
		}

		if isTypeSet(bitsSet, Bits_HasData) {
			instruction.Operands[1] = OperandImmediate{bits[Bits_HasData], w}
		}

		instruction.Op = bp.Name
		instruction.Size = bytesRead

		return &instruction, nil
	}

	return nil, errors.New("No command")
}

func DecodeRm(rm uint16, mod uint16, wide bool, disp uint16) Operand {
	switch mod {
	case 0b00:
		if rm == 0b110 {
			return OperandDirectAddress(disp)
		}
		return eac(rm, wide, disp)

	case 0b01:
		return eac(rm, wide, disp)

	case 0b10:
		return eac(rm, wide, disp)

	case 0b11:
		return DecodeReg(rm, wide)
	}

	return nil
}

func eac(rm uint16, wide bool, disp uint16) OperandEffectiveAddress {
	regs := []string{
		"bx+si",
		"bx+di",
		"bp+si",
		"bp+di",
		"si",
		"di",
		"bp",
		"bx",
	}

	return OperandEffectiveAddress{regs[rm], int16(disp), wide}
}

type RegisterIndex byte

const (
	RI_a RegisterIndex = iota
	RI_c
	RI_d
	RI_b
	RI_sp
	RI_bp
	RI_si
	RI_di
	RI_ip

	RI_flags

	RI_Count
)

type RegisterFlag int

const (
	RF_zero RegisterFlag = iota
	RF_sign

	RF_Count
)

func DecodeReg(reg uint16, wide bool) OperandRegister {
	regs := [][2]OperandRegister{
		// Might need swapping low and high
		// currently low - no offset, high - 1 offset
		{{RI_a, 0, 1}, {RI_a, 0, 2}},
		{{RI_c, 0, 1}, {RI_c, 0, 2}},
		{{RI_d, 0, 1}, {RI_d, 0, 2}},
		{{RI_b, 0, 1}, {RI_b, 0, 2}},
		{{RI_a, 1, 1}, {RI_sp, 0, 2}},
		{{RI_c, 1, 1}, {RI_bp, 0, 2}},
		{{RI_d, 1, 1}, {RI_si, 0, 2}},
		{{RI_b, 1, 1}, {RI_di, 0, 2}},
	}

	idx := 0
	if wide {
		idx = 1
	}

	return regs[reg][idx]
}

func isTypeSet(flags uint32, bitsType BitsType) bool {
	bit := uint32(1 << bitsType)
	return flags&bit == bit
}

type Operand interface {
	fmt.Stringer
}

type OperandRegister struct {
	Index  RegisterIndex
	Offset int
	Size   int
}

func (reg OperandRegister) String() string {
	var regStrings = [][3]string{
		{"al", "ah", "ax"},
		{"cl", "ch", "cx"},
		{"dl", "dh", "dx"},
		{"bl", "bh", "bx"},
		{"sp", "sp", "sp"},
		{"", "", "bp"},
		{"", "", "si"},
		{"", "", "di"},
		{"", "", "ip"},
	}

	idx := reg.Offset
	if reg.Size == 2 {
		idx = 2
	}

	return regStrings[reg.Index][idx]
}

type OperandImmediate struct {
	Value uint16
	Wide  bool
}

func (imm OperandImmediate) String() string {
	size := "byte"
	if imm.Wide {
		size = "word"
	}

	return fmt.Sprintf("%s %d", size, imm.Value)
}

type OperandDirectAddress int

func (addr OperandDirectAddress) String() string {
	return fmt.Sprintf("[%d]", addr)
}

type OperandEffectiveAddress struct {
	Base string
	Disp int16
	Wide bool
}

func (ea OperandEffectiveAddress) String() string {
	return fmt.Sprintf("[%s%+d]", ea.Base, ea.Disp)
}

type Instruction struct {
	Op       string
	Size     int
	Operands [2]Operand
}

func (inst Instruction) String() string {
	var stringOperands []string
	for _, op := range inst.Operands {
		if op != nil {
			stringOperands = append(stringOperands, op.String())
		}
	}

	return fmt.Sprintf("%s %s", inst.Op, strings.Join(stringOperands, ", "))
}
