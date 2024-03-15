package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
)

var mode string
var dump string
var filePath string

func init() {
	flag.StringVar(&mode, "mode", "decode", "command mode - [exec, decode, cycles]")
	flag.StringVar(&dump, "dump", "", "file path for memory dump")
	flag.StringVar(&filePath, "path", "", "file path to asm binary")
}

func main() {
	flag.Parse()

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
