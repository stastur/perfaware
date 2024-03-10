package main

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
