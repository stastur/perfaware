package main

import (
	"errors"
	"fmt"
	"os"
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
	{"add", []Bits{Const(6, 0b100000), S_FLAG, W_FLAG, MOD, Const(3, 0b000), DISP, DATA}},
	{"add", []Bits{Const(7, 0b0000010), W_FLAG, DATA}},
}

func main() {
	filePath := os.Args[1]
	buff, err := os.ReadFile(filePath)
	if err != nil {
		println("Error reading file.")
		panic(err)
	}

	// for _, b := range buff {
	// 	fmt.Printf("%b ", b)
	// }
	// fmt.Println()

	fmt.Println("bits 16")
	index := 0
	for {
		index, err = DecodeInstruction(index, buff)
		if err != nil {
			fmt.Println(";", err)
			break
		}
	}
}

func DecodeInstruction(startingAt int, buff []byte) (int, error) {
	currentByteIndex := startingAt

	if currentByteIndex >= len(buff) {
		return -1, errors.New("EOF")
	}

	for _, bp := range Blueprints {
		bitsLeft := 8
		currentByte := buff[currentByteIndex]
		bytesRead := 1

		var bitsSet uint32 = 0
		bits := make([]uint32, Bits_Count)

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
				constBits &= byte(0xff >> (8 - part.BitCount))

				if constBits != part.Value {
					isValid = false
					break
				}
			}

			bitsSet |= 1 << part.Type
			bitsLeft -= part.BitCount

			mask := ^byte(0xff << part.BitCount)
			if part.Value != 0 {
				bits[part.Type] = uint32(part.Value)
			} else {
				bits[part.Type] = uint32((currentByte >> bitsLeft) & mask)
			}
		}

		if !isValid {
			continue
		}

		mod := bits[Bits_Mod]
		rm := bits[Bits_Rm]
		w := bits[Bits_W] == 1
		s := bits[Bits_S] == 1

		readFromBuff := func(exists bool, wide bool, signExtended bool) uint32 {
			if !exists {
				return 0
			}

			if wide {
				lo := uint32(buff[currentByteIndex+bytesRead+0])
				hi := uint32(buff[currentByteIndex+bytesRead+1])
				bytesRead += 2

				return hi<<8 | lo
			} else {
				lo := buff[currentByteIndex+bytesRead]
				bytesRead += 1

				if signExtended {
					return uint32(int8(lo))
				}

				return uint32(lo)
			}
		}

		hasDirectAddress := mod == 0b00 && rm == 0b110
		hasDisp := isTypeSet(bitsSet, Bits_HasDisp) &&
			(mod == 0b10 || mod == 0b01 || hasDirectAddress)
		hasData := isTypeSet(bitsSet, Bits_HasData)

		bits[Bits_HasDisp] = readFromBuff(hasDisp, mod == 0b10 || hasDirectAddress, w)
		bits[Bits_HasData] = readFromBuff(hasData, w, s)

		var operandLeft interface{}

		if isTypeSet(bitsSet, Bits_Mod) {
			operandLeft = DecodeMemory(rm, mod, w, bits[Bits_HasDisp])
		} else if isTypeSet(bitsSet, Bits_HasData) {
			operandLeft = fmt.Sprintf("%d", bits[Bits_HasData])
		} else if isTypeSet(bitsSet, Bits_HasAddr) {
			operandLeft = fmt.Sprintf("[%d]", readFromBuff(true, w, false))
		}

		var operandRight interface{}

		if isTypeSet(bitsSet, Bits_Reg) {
			operandRight = DecodeReg(bits[Bits_Reg], w)
		} else if isTypeSet(bitsSet, Bits_HasData) {
			var size string
			if w {
				size = "word"
			} else {
				size = "byte"
			}

			operandRight = fmt.Sprintf("%s %d", size, bits[Bits_HasData])
		}

		if bits[Bits_D] == 1 || (isTypeSet(bitsSet, Bits_E) && bits[Bits_E] == 0) {
			operandLeft, operandRight = operandRight, operandLeft
		}

		fmt.Printf("%s %s, %s\n", bp.Name, operandLeft, operandRight)
		currentByteIndex += bytesRead

		return currentByteIndex, nil
	}

	return -1, errors.New("No command")
}

func DecodeMemory(rm uint32, mod uint32, wide bool, disp uint32) interface{} {
	switch mod {
	case 0b00:
		if rm == 0b110 {
			return fmt.Sprintf("[%d]", disp)
		}
		return eac(rm, mod, wide, disp)

	case 0b01:
		return eac(rm, mod, wide, disp)

	case 0b10:
		return eac(rm, mod, wide, disp)

	case 0b11:
		return DecodeReg(rm, wide)
	}

	return ""
}

func eac(rm uint32, mod uint32, wide bool, disp uint32) string {
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

	return fmt.Sprintf("[%s%+d]", regs[rm], int16(disp))
}

func DecodeReg(reg uint32, wide bool) string {
	regs := [][2]string{
		{"al", "ax"},
		{"cl", "cx"},
		{"dl", "dx"},
		{"bl", "bx"},
		{"ah", "sp"},
		{"ch", "bp"},
		{"dh", "si"},
		{"bh", "di"},
	}

	if wide {
		return regs[reg][1]
	} else {
		return regs[reg][0]
	}
}

func isTypeSet(flags uint32, bitsType BitsType) bool {
	bit := uint32(1 << bitsType)
	return flags&bit == bit
}
