package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
)

type Instruction struct {
	Opcode byte
	D      byte
	W      byte
	Mod    byte
	Reg    byte
	Rm     byte
}

var regs = [][2]string{
	{"al", "ax"},
	{"cl", "cx"},
	{"dl", "dx"},
	{"bl", "bx"},
	{"ah", "sp"},
	{"ch", "bp"},
	{"dh", "si"},
	{"bh", "di"},
}

func (i Instruction) Decode() string {
	var op string
	switch i.Opcode {
	case 0b100010:
		op = "mov"
	default:
		op = "invalid"
	}

	left_operand := regs[i.Reg][i.W]
	right_operand := regs[i.Rm][i.W]

	if i.D == 0b0 {
		left_operand, right_operand = right_operand, left_operand
	}

	return fmt.Sprintf("%s %s, %s", op, left_operand, right_operand)
}

func main() {
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	println("Disassembling " + filePath + "...")

	buff := bufio.NewReader(file)
	for {
		var encoded uint16
		err := binary.Read(buff, binary.BigEndian, &encoded)
		if err != nil {
			break
		}

		var instruction Instruction
		instruction.Opcode = byte(encoded >> 10)
		instruction.D = byte((encoded >> 9) & 0b1)
		instruction.W = byte((encoded >> 8) & 0b1)
		instruction.Mod = byte((encoded >> 6) & 0b11)
		instruction.Reg = byte((encoded >> 3) & 0b111)
		instruction.Rm = byte(encoded & 0b111)

		fmt.Println(instruction.Decode())
	}
}
