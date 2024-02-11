package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type Instruction []byte

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

var effectiveAddress = []string{
	"bx+si",
	"bx+di",
	"bp+si",
	"bp+di",
	"si",
	"di",
	"bp",
	"bx",
}

var aluOps = []string{
	"add",
	"or",
	"adc",
	"sbb",
	"and",
	"sub",
	"xor",
	"cmp",
}

var jumps = []string{
	"jo",
	"jno",
	"jb",
	"jnb",
	"jz",
	"jnz",
	"jbe",
	"ja",
	"js",
	"jns",
	"jp",
	"jnp",
	"jl",
	"jge",
	"jle",
	"jg",
}

var conrolTransfers = []string{
	"loopnz",
	"loopz",
	"loop",
	"jcxz",
}

func DecodeInstruction(buff io.Reader) (string, error) {
	var b1 byte
	err := binary.Read(buff, binary.LittleEndian, &b1)
	if err != nil {
		return "", err
	}
	var op, left, right string

	switch true {
	case b1>>2 == 0b100010:
		op = "mov"

		var b2 byte
		binary.Read(buff, binary.LittleEndian, &b2)

		d := b1 >> 1 & 0b1
		w := b1 & 0b1
		mod := b2 >> 6
		reg := b2 >> 3 & 0b111
		rm := b2 & 0b111

		left = regs[reg][w]
		right = DecodeRm(buff, mod, w, rm)

		if d == 0 {
			left, right = right, left
		}

	case b1>>1 == 0b1100011:
		// immediate to register/memory
		op = "mov"

		var b2 byte
		binary.Read(buff, binary.LittleEndian, &b2)

		w := b1 & 0b1
		mod := b2 >> 6
		rm := b2 & 0b111

		left = DecodeRm(buff, mod, w, rm)

		if w == 0 {
			var b byte
			binary.Read(buff, binary.LittleEndian, &b)
			right = fmt.Sprintf("byte %d", b)
		} else {
			var w uint16
			binary.Read(buff, binary.LittleEndian, &w)
			right = fmt.Sprintf("word %d", w)
		}

	case b1>>4 == 0b1011:
		// immediate to register
		op = "mov"

		w := b1 >> 3 & 0b1
		rm := b1 & 0b111
		left = regs[rm][w]

		if w == 0 {
			var disp8 byte
			binary.Read(buff, binary.LittleEndian, &disp8)
			right = fmt.Sprintf("%d", disp8)
		} else {
			var disp16 uint16
			binary.Read(buff, binary.LittleEndian, &disp16)
			right = fmt.Sprintf("%d", disp16)
		}

	case b1>>2 == 0b101000:
		// memory to accumulator
		op = "mov"
		d := (b1 >> 1 & 0b1) ^ 0b1
		w := b1 & 0b1
		left = regs[0][w]

		if w == 0 {
			var disp8 byte
			binary.Read(buff, binary.LittleEndian, &disp8)
			right = fmt.Sprintf("[%d]", disp8)
		} else {
			var disp16 uint16
			binary.Read(buff, binary.LittleEndian, &disp16)
			right = fmt.Sprintf("[%d]", disp16)
		}

		if d == 0 {
			left, right = right, left
		}

	case b1>>6 == 0 && b1>>2&0b1 == 0:
		op = aluOps[b1>>3&0b111]

		var b2 byte
		binary.Read(buff, binary.LittleEndian, &b2)

		d := b1 >> 1 & 0b1
		w := b1 & 0b1
		mod := b2 >> 6
		reg := b2 >> 3 & 0b111
		rm := b2 & 0b111

		left = regs[reg][w]
		right = DecodeRm(buff, mod, w, rm)

		if d == 0 {
			left, right = right, left
		}

	case b1>>2 == 0b100000:
		var b2 byte
		binary.Read(buff, binary.LittleEndian, &b2)

		s := b1 >> 1 & 0b1
		w := b1 & 0b1
		mod := b2 >> 6
		rm := b2 & 0b111

		op = aluOps[b2>>3&0b111]

		left = DecodeRm(buff, mod, w, rm)
		if w == 0 {
			left = fmt.Sprintf("byte %s", left)
		} else {
			left = fmt.Sprintf("word %s", left)
		}

		if s<<1|w == 0b01 {
			var w uint16
			binary.Read(buff, binary.LittleEndian, &w)
			right = fmt.Sprintf("%d", w)
		} else {
			var b byte
			binary.Read(buff, binary.LittleEndian, &b)
			right = fmt.Sprintf("%d", b)
		}

	case (b1>>1&0b11) == 0b10 && (b1>>6 == 0):
		op = aluOps[b1>>3&0b111]

		w := b1 & 0b1
		left = regs[0][w]

		if w == 0 {
			var disp8 byte
			binary.Read(buff, binary.LittleEndian, &disp8)
			right = fmt.Sprintf("%d", disp8)
		} else {
			var disp16 uint16
			binary.Read(buff, binary.LittleEndian, &disp16)
			right = fmt.Sprintf("%d", disp16)
		}

	case b1>>4 == 0b0111:
		op = jumps[b1&0b1111]
		var disp8 byte
		binary.Read(buff, binary.LittleEndian, &disp8)

		return fmt.Sprintf("%s %d", op, disp8), nil

	case b1>>4 == 0b1110:
		op = conrolTransfers[b1&0b1111]
		var disp8 byte
		binary.Read(buff, binary.LittleEndian, &disp8)

		return fmt.Sprintf("%s %d", op, disp8), nil
	}

	return fmt.Sprintf("%s %s, %s", op, left, right), nil
}

const (
	MemoryModeNoDisp = iota
	MemoryModeDisp8
	MemoryModeDisp16
	RegisterMode
)

func DecodeRm(buff io.Reader, mod byte, w byte, rm byte) string {
	var decoded string

	switch mod {
	case MemoryModeNoDisp:
		if rm == 0b110 {
			var disp16 uint16
			binary.Read(buff, binary.LittleEndian, &disp16)
			decoded = fmt.Sprintf("[%d]", disp16)
		} else {
			decoded = fmt.Sprintf("[%s]", effectiveAddress[rm])
		}

	case MemoryModeDisp8:
		var disp8 byte
		binary.Read(buff, binary.LittleEndian, &disp8)

		if w == 1 {
			extended := int16(int8(disp8))
			decoded = fmt.Sprintf("[%s%+d]", effectiveAddress[rm], extended)
		} else {
			decoded = fmt.Sprintf("[%s%+d]", effectiveAddress[rm], disp8)
		}

	case MemoryModeDisp16:
		var disp16 uint16
		binary.Read(buff, binary.LittleEndian, &disp16)

		if w == 1 {
			decoded = fmt.Sprintf("[%s%+d]", effectiveAddress[rm], int16(disp16))
		} else {
			decoded = fmt.Sprintf("[%s%+d]", effectiveAddress[rm], disp16)
		}

	case RegisterMode:
		decoded = regs[rm][w]
	}

	return decoded
}

func main() {
	filePath := os.Args[1]
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	buff := bufio.NewReader(file)

	fmt.Println("bits 16")
	for {
		instruction, err := DecodeInstruction(buff)
		if err != nil {
			break
		}

		fmt.Println(instruction)
	}
}
