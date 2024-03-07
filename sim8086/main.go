package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type BitsType int

type Bits struct {
	Type     BitsType
	BitCount int
	Value    byte
}

const (
	Bits_Literal BitsType = iota
	Bits_Mod
	Bits_Reg
	Bits_Rm

	Bits_W
	Bits_D
	Bits_S
	Bits_E // made up flag, opposite to D

	Bits_HasData
	Bits_HasDisp
	Bits_HasAddr

	Bits_Count
)

func Const(size int, value byte) Bits {
	return Bits{
		Type:     Bits_Literal,
		BitCount: size,
		Value:    value,
	}
}

func Implicit(t BitsType, value byte) Bits {
	return Bits{
		Type:     t,
		BitCount: 0,
		Value:    value,
	}
}

var D_FLAG = Bits{Bits_D, 1, 0}
var W_FLAG = Bits{Bits_W, 1, 0}
var S_FLAG = Bits{Bits_S, 1, 0}
var E_FLAG = Bits{Bits_E, 1, 0}

var MOD = Bits{Bits_Mod, 2, 0}
var REG = Bits{Bits_Reg, 3, 0}
var RM = Bits{Bits_Rm, 3, 0}

var DATA = Bits{Bits_HasData, 0, 0}
var ADDR = Bits{Bits_HasAddr, 0, 0}
var DISP = Bits{Bits_HasDisp, 0, 0}

type IstructionBlueprint struct {
	Name string
	Bits []Bits
}

var Blueprints = []IstructionBlueprint{
	{"mov", []Bits{Const(6, 0b100010), D_FLAG, W_FLAG, MOD, REG, RM, DISP}},
	{"mov", []Bits{Const(7, 0b1100011), W_FLAG, MOD, Const(3, 0), RM, DISP, DATA}},
	{"mov", []Bits{Const(4, 0b1011), W_FLAG, REG, DATA, Implicit(Bits_D, 1)}},
	{"mov", []Bits{Const(6, 0b101000), E_FLAG, W_FLAG, ADDR, Implicit(Bits_Reg, 0)}},

	{"add", []Bits{Const(6, 0b000000), D_FLAG, W_FLAG, MOD, REG, RM, DISP}},
	{"add", []Bits{Const(6, 0b100000), S_FLAG, W_FLAG, MOD, Const(3, 0), RM, DISP, DATA}},
	{"add", []Bits{Const(7, 0b0000010), W_FLAG, DATA, Implicit(Bits_Reg, 0), Implicit(Bits_D, 1)}},

	{"sub", []Bits{Const(6, 0b001010), D_FLAG, W_FLAG, MOD, REG, RM, DISP}},
	{"sub", []Bits{Const(6, 0b100000), S_FLAG, W_FLAG, MOD, Const(3, 0b101), RM, DISP, DATA}},
	{"sub", []Bits{Const(7, 0b0010110), W_FLAG, DATA, Implicit(Bits_Reg, 0), Implicit(Bits_D, 1)}},

	{"cmp", []Bits{Const(6, 0b001110), D_FLAG, W_FLAG, MOD, REG, RM, DISP}},
	{"cmp", []Bits{Const(6, 0b100000), S_FLAG, W_FLAG, MOD, Const(3, 0b111), RM, DISP, DATA}},
	{"cmp", []Bits{Const(7, 0b0011110), W_FLAG, DATA, Implicit(Bits_Reg, 0), Implicit(Bits_D, 1)}},

	{"jo", []Bits{Const(4, 0b0111), Const(4, 0), DATA}},
	{"jno", []Bits{Const(4, 0b0111), Const(4, 1), DATA}},
	{"jb", []Bits{Const(4, 0b0111), Const(4, 2), DATA}},
	{"jnb", []Bits{Const(4, 0b0111), Const(4, 3), DATA}},
	{"jz", []Bits{Const(4, 0b0111), Const(4, 4), DATA}},
	{"jne", []Bits{Const(4, 0b0111), Const(4, 5), DATA}},
	{"jbe", []Bits{Const(4, 0b0111), Const(4, 6), DATA}},
	{"ja", []Bits{Const(4, 0b0111), Const(4, 7), DATA}},
	{"js", []Bits{Const(4, 0b0111), Const(4, 8), DATA}},
	{"jns", []Bits{Const(4, 0b0111), Const(4, 9), DATA}},
	{"jp", []Bits{Const(4, 0b0111), Const(4, 10), DATA}},
	{"jnp", []Bits{Const(4, 0b0111), Const(4, 11), DATA}},
	{"jl", []Bits{Const(4, 0b0111), Const(4, 12), DATA}},
	{"jnl", []Bits{Const(4, 0b0111), Const(4, 13), DATA}},
	{"jle", []Bits{Const(4, 0b0111), Const(4, 14), DATA}},
	{"jg", []Bits{Const(4, 0b0111), Const(4, 15), DATA}},

	{"loopnz", []Bits{Const(4, 0b1110), Const(4, 0), DATA}},
	{"loopz", []Bits{Const(4, 0b1110), Const(4, 1), DATA}},
	{"loop", []Bits{Const(4, 0b1110), Const(4, 2), DATA}},
	{"jcxz", []Bits{Const(4, 0b1110), Const(4, 3), DATA}},
}

func main() {
	filePath := os.Args[1]
	buff, err := os.ReadFile(filePath)
	if err != nil {
		println("Error reading file.")
		panic(err)
	}

	fmt.Println("bits 16")

	for index := 0; index < len(buff); {
		instruction, err := DecodeInstruction(index, buff)
		if err != nil {
			fmt.Println(";", err)
			break
		}

		fmt.Println(instruction.String())
		index += instruction.Size
	}
}

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
			instruction.Operands[0] = DirectAddress(readFromBuff(true, w, false))
		}

		if isTypeSet(bitsSet, Bits_Reg) {
			instruction.Operands[1] = DecodeReg(bits[Bits_Reg], w)
		}

		if bits[Bits_D] == 1 || (isTypeSet(bitsSet, Bits_E) && bits[Bits_E] == 0) {
			instruction.Operands[0], instruction.Operands[1] = instruction.Operands[1], instruction.Operands[0]
		}

		if isTypeSet(bitsSet, Bits_HasData) {
			instruction.Operands[1] = Immediate{bits[Bits_HasData], w}
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
			return DirectAddress(disp)
		}
		return eac(rm, mod, wide, disp)

	case 0b01:
		return eac(rm, mod, wide, disp)

	case 0b10:
		return eac(rm, mod, wide, disp)

	case 0b11:
		return DecodeReg(rm, wide)
	}

	return nil
}

func eac(rm uint16, mod uint16, wide bool, disp uint16) EffectiveAddress {
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

	return EffectiveAddress{regs[rm], int16(disp)}
}

func DecodeReg(reg uint16, wide bool) Register {
	regs := [][2]Register{
		{"al", "ax"},
		{"cl", "cx"},
		{"dl", "dx"},
		{"bl", "bx"},
		{"ah", "sp"},
		{"ch", "bp"},
		{"dh", "si"},
		{"bh", "di"},
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

type Register string

func (reg Register) String() string {
	return string(reg)
}

type Immediate struct {
	Value uint16
	Wide  bool
}

func (imm Immediate) String() string {
	size := "byte"
	if imm.Wide {
		size = "word"
	}

	return fmt.Sprintf("%s %d", size, imm.Value)
}

type DirectAddress int

func (addr DirectAddress) String() string {
	return fmt.Sprintf("[%d]", addr)
}

type EffectiveAddress struct {
	Base string
	Disp int16
}

func (ea EffectiveAddress) String() string {
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
