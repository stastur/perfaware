package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

func main() {
	flags := make(map[string]string)

	for _, flag := range os.Args[1:] {
		pair := strings.Split(strings.Trim(flag, " -"), "=")

		if len(pair) == 2 {
			flags[pair[0]] = pair[1]
		}
	}

	mode := flags["mode"]
	dump := flags["dump"]
	filePath := flags["path"]

	buff, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println("Error reading file.")
		panic(err)
	}

	fmt.Println("bits 16")

	memory := make(Memory, (2<<15)-1)
	registers := make(Registers, RI_Count)
	cycles := 0

	for int(registers[RI_ip]) < len(buff) {
		instruction, err := DecodeInstruction(int(registers[RI_ip]), buff)

		if err != nil {
			fmt.Println(";", err)
			break
		}

		registers[RI_ip] += int16(instruction.Size)
		cycles += instruction.EstimateCycles()

		fmt.Println(instruction.String())

		if mode == "cycles" {
			fmt.Printf("; cycles +%d = %d\n", instruction.EstimateCycles(), cycles)
		}

		if mode == "exec" {
			ExecuteIntruction(*instruction, registers, memory)
		}
	}

	if mode == "exec" {
		fmt.Println()
		registers.Print()
	}

	if dump != "" {
		file, err := os.Create(dump)
		if err != nil {
			return
		}

		binary.Write(file, binary.BigEndian, memory)
	}
}
