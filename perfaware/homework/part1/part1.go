package main

import (
	"fmt"
	"os"
)

const DEBUG = false

var RegMap = [16]Register{
	{
		name: "al",
		size: 8,
		offset: 0,
		fieldEncoding: 0b000,
	},
	{
		name: "cl",
		size: 8,
		offset: 0,
		fieldEncoding: 0b001,
	},
	{
		name: "dl",
		size: 8,
		offset: 0,
		fieldEncoding: 0b010,
	},
	{
		name: "bl",
		size: 8,
		offset: 0,
		fieldEncoding: 0b011,
	},
	{
		name: "ah",
		size: 8,
		offset: 8,
		fieldEncoding: 0b100,
	},
	{
		name: "ch",
		size: 8,
		offset: 8,
		fieldEncoding: 0b101,
	},
	{
		name: "dh",
		size: 8,
		offset: 8,
		fieldEncoding: 0b110,
	},
	{
		name: "bh",
		size: 8,
		offset: 8,
		fieldEncoding: 0b111,
	},
	{
		name: "ax",
		size: 16,
		offset: 0,
		fieldEncoding: 0b000,
	},
	{
		name: "cx",
		size: 16,
		offset: 0,
		fieldEncoding: 0b001,
	},
	{
		name: "dx",
		size: 16,
		offset: 0,
		fieldEncoding: 0b010,
	},
	{
		name: "bx",
		size: 16,
		offset: 0,
		fieldEncoding: 0b011,
	},
	{
		name: "sp",
		size: 16,
		offset: 0,
		fieldEncoding: 0b100,
	},
	{
		name: "bp",
		size: 16,
		offset: 0,
		fieldEncoding: 0b101,
	},
	{
		name: "si",
		size: 16,
		offset: 0,
		fieldEncoding: 0b110,
	},
	{
		name: "di",
		size: 16,
		offset: 0,
		fieldEncoding: 0b111,
	},
}

var EffAddrBaseMap = [8]string{
	"bx + si",
	"bx + di",
	"bp + si",
	"bp + di",
	"si",
	"di",
	"bp",
	"bx",
}

type ArithmeticOpType = uint8
const (
	Add ArithmeticOpType = iota
	NotSupported1
	NotSupported2
	NotSupported3
	NotSupported4
	Sub
	NotSupported5
	Cmp
)

func getArithmeticOpName(op ArithmeticOpType) string {
	switch op {
	case Add:
		return "add"
	case Sub:
		return "sub"
	case Cmp:
		return "cmp"
	default:
		return "NOT_SUPPORTED"
	}
}

 type EffectiveAddress struct {
	 base string
	 disp uint16
 }

 func (ea EffectiveAddress) String() string {
	if ea.base == "" {
		return fmt.Sprintf("[%d]", ea.disp)
	} else if ea.disp == 0 {
		return fmt.Sprintf("[%s]", ea.base)
	} else {
		return fmt.Sprintf("[%s + %d]", ea.base, ea.disp)
	}
}

type Register struct {
	name string
	size uint8
	offset uint8
	fieldEncoding uint8
}

type Flags struct {
	zero bool
	sign bool
}

var registerValues = [8]uint16{
	0, // "ax",
	0, // "cx",
	0, // "dx",
	0, // "bx",
	0, // "sp",
	0, // "bp",
	0, // "si",
	0,  // "di",
}

var flagValues =  Flags {
	false,
	false,
}
func getRegisterValue(reg Register) uint16 {
	if reg.size == 16 {
		return registerValues[reg.fieldEncoding]
	} else {
		return registerValues[reg.fieldEncoding % 4]
	}
}

func updateRegister(reg Register,  value uint16) {
	if reg.size == 16 {
		registerValues[reg.fieldEncoding] =  uint16(value)
	} else {
		current := getRegisterValue(reg)
		var updated uint16
		if reg.offset == 0 {
			updated = (current & uint16(0xFF00)) | uint16(value)
		} else {
			updated = (current & uint16(0x00FF)) | uint16(value << 8)
		}
		registerValues[reg.fieldEncoding] =  updated
	}
}

func executeOp(op ArithmeticOpType, dst Register, other uint16) {
	current := getRegisterValue(dst)
	var result uint16
	switch op {
	case Add:
		result = current + other
		updateRegister(dst, result)
	case Sub:
		result = current - other
		updateRegister(dst, result)
	case Cmp:
		result = current - other
	default:
		panic(fmt.Sprintf("Unsupported op %x", op))
	}
	flagValues.zero = result == 0
	flagValues.sign = result & 0x8000 > 0
	printFlags()
}



func resolveOther(startIdx int, data []byte, mod byte, rm byte, w bool) (Register, EffectiveAddress, uint8) {
	var otherReg Register
	var otherEffAddr EffectiveAddress
	var bytesConsumed uint8
	effAddrBase := EffAddrBaseMap[rm]
	switch mod {
	case 0:
		if rm == 6 {
			dispLo := data[startIdx+2]
			dispHi := data[startIdx+3]
			disp := (uint16(dispHi) << 8) | uint16(dispLo)
			otherEffAddr = EffectiveAddress{ base: "", disp: disp }
			bytesConsumed = 4
		} else {
			otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: 0 }
			bytesConsumed = 2
		}
	case 1:
		dispLo := uint16(data[startIdx+2])
		otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: dispLo }
		bytesConsumed = 3
	case 2:
		dispLo := data[startIdx+2]
		dispHi := data[startIdx+3]
		disp := (uint16(dispHi) << 8) | uint16(dispLo)
		otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: disp }
		bytesConsumed = 4
	case 3:
		if w {
			otherReg = RegMap[uint8(rm)+8]
		} else {
			otherReg = RegMap[uint8(rm)]
		}
		bytesConsumed = 2
	default:
		panic("MOD should never have this value")
	}
	return otherReg, otherEffAddr, bytesConsumed
}

func getImmediate(startIdx int, data []byte, wide bool) (int16, uint8) {
	if wide {
		return int16(uint16(data[startIdx+1])<<8 | uint16(data[startIdx])), 2
	} else {
		return int16(int8(data[startIdx])), 1
	}
}

func getOtherString(otherReg Register, otherEffAddr EffectiveAddress) string {
	if otherReg.name ==  "" {
		return otherEffAddr.String()
	} else {
		return otherReg.name
	}
}

func acceptRegMemMov(startIdx int, data []byte) (string, uint) {
	d := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111
	regBits := (b2 & 0b00111000) >> 3

	var reg Register
	var bytesConsumed uint8
	if w {
		reg = RegMap[uint8(regBits)+8]
	} else {
		reg = RegMap[uint8(regBits)]
	}
	otherReg, otherEffAddr, bytesConsumed := resolveOther(startIdx, data, mod, rm, w)
	var src, dst string
	other := getOtherString(otherReg, otherEffAddr)
	if d {
		src = other
		dst = reg.name
	} else {
		src = reg.name
		dst = other
	}
	if otherReg.name !=  "" {
		var current uint16
		var dst Register
		if d {
			current = getRegisterValue(otherReg)
			dst = reg
		} else {
			current = getRegisterValue(reg)
			dst = otherReg
		}
		updateRegister(dst, current)
	}
	asm := fmt.Sprintf("mov %s,%s", dst, src)
	return asm, uint(bytesConsumed)
}

func acceptImmediateToRegMov(startIdx int, data []byte) (string, uint) {
	w := data[startIdx]&8 > 0
	regBits := data[startIdx] & 0b00000111

	immediate, immediateSize := getImmediate(startIdx+1, data, w)

	var reg Register
	if w {
		reg = RegMap[regBits+8]
	} else {
		reg = RegMap[regBits]
	}
	updateRegister(reg, uint16(immediate))
	return fmt.Sprintf("mov %s, %d", reg.name, immediate), uint(1 + immediateSize)
}

func acceptArithmeticOpRegMem(startIdx int, data []byte) (string, uint) {
	d := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	regBits := (b2 & 0b00111000) >> 3
	rm := b2 & 0b00000111

	var reg Register
	var bytesConsumed uint8
	if w {
		reg = RegMap[uint8(regBits)+8]
	} else {
		reg = RegMap[uint8(regBits)]
	}
	op := uint8(data[startIdx] & 0b00111000 >> 3)
	otherReg, otherEffAddr, bytesConsumed := resolveOther(startIdx, data, mod, rm, w)
	other := getOtherString(otherReg, otherEffAddr)
	var src, dst string
	if d {
		src = other
		dst = reg.name
		if otherReg.name != "" {
			executeOp(op, reg, getRegisterValue(otherReg))
		}
	} else {
		src = reg.name
		dst = other
		if otherReg.name != "" {
			executeOp(op, otherReg, getRegisterValue(reg))
		}
	}
	asm := fmt.Sprintf("%s %s, %s", getArithmeticOpName(op), dst, src)
	return asm, uint(bytesConsumed)
}

func acceptArithmeticOpRegMemImmediate(startIdx int, data []byte) (string, uint) {
	s := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111

	var dst string
	var bytesConsumed uint8
	dstReg, otherEffAddr, bytesConsumed := resolveOther(startIdx, data, mod, rm, w)
	wideImmediate := w && !s
	immediate, imNumBytes := getImmediate(startIdx+int(bytesConsumed), data, wideImmediate)
	bytesConsumed += imNumBytes
	op := uint8(data[startIdx+1] & 0b00111000 >> 3)
	clarifier := ""
	if otherEffAddr.base != "" || otherEffAddr.disp != 0 {
		if w {
			clarifier = " word"
		} else {
			clarifier = " byte"
		}
	} else {
		executeOp(op, dstReg, uint16(immediate))
	}
	fmt.Println(getArithmeticOpName(op))
	asm := fmt.Sprintf("%s%s %s, %d", getArithmeticOpName(op), clarifier, dst, immediate)
	return asm, uint(bytesConsumed)
}
//
// func acceptArithmeticOpImmediateAcc(startIdx int, data []byte) (string, uint) {
// 	w := data[startIdx]&1 > 0
// 	immediate, _ := getImmediate(startIdx+1, data, w)
// 	op := uint8(data[startIdx] & 0b00111000 >> 3)
// 	var bytesConsumed uint
// 	var reg string
// 	if w {
// 		reg = "ax"
// 		bytesConsumed = 3
// 	} else {
// 		reg = "al"
// 		bytesConsumed = 2
// 	}
// 	asm := fmt.Sprintf("%s %s, %d", ArithmeticVariantLookup[op], reg, immediate)
// 	return asm, bytesConsumed
// }

func acceptCondJump(startIdx int, data []byte, jumpVariant string) (string, uint) {
	displacement := int8(data[startIdx+1])
	asm := fmt.Sprintf("%s %d", jumpVariant, displacement)
	return asm, 2
}

func decode(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Printf("bits 16\n\n")
	for i := 0; i < len(data); {
		var asm string
		var bytesConsumed uint
		if (0b10001000^data[i])&0b11111100 == 0 {
			asm, bytesConsumed = acceptRegMemMov(i, data)
		} else if (0b10110000^data[i])&0b11110000 == 0 {
			asm, bytesConsumed = acceptImmediateToRegMov(i, data)
		} else if (0b00000000^data[i])&0b11111100 == 0 ||
			(0b00101000^data[i])&0b11111100 == 0 ||
			(0b00111000^data[i])&0b11111100 == 0 {
			asm, bytesConsumed = acceptArithmeticOpRegMem(i, data)
		} else if (0b10000000^data[i])&0b11111100 == 0 {
			asm, bytesConsumed = acceptArithmeticOpRegMemImmediate(i, data)
		// } else if (0b00000100^data[i])&0b11111110 == 0 ||
		// 	(0b00101100^data[i])&0b11111110 == 0 ||
		// 	(0b00111100^data[i])&0b11111110 == 0 {
		// 	asm, bytesConsumed = acceptArithmeticOpImmediateAcc(i, data)
		} else if data[i] == 0b01110100 {
			asm, bytesConsumed = acceptCondJump(i, data, "je")
		} else if data[i] == 0b01111100 {
			asm, bytesConsumed = acceptCondJump(i, data, "jl")
		} else if data[i] == 0b01111110 {
			asm, bytesConsumed = acceptCondJump(i, data, "jle")
		} else if data[i] == 0b01110010 {
			asm, bytesConsumed = acceptCondJump(i, data, "jb")
		} else if data[i] == 0b01110110 {
			asm, bytesConsumed = acceptCondJump(i, data, "jbe")
		} else if data[i] == 0b01111010 {
			asm, bytesConsumed = acceptCondJump(i, data, "jp")
		} else if data[i] == 0b01110000 {
			asm, bytesConsumed = acceptCondJump(i, data, "jo")
		} else if data[i] == 0b01111000 {
			asm, bytesConsumed = acceptCondJump(i, data, "js")
		} else if data[i] == 0b01110101 {
			asm, bytesConsumed = acceptCondJump(i, data, "jnz")
		} else if data[i] == 0b01111101 {
			asm, bytesConsumed = acceptCondJump(i, data, "jnl")
		} else if data[i] == 0b01111111 {
			asm, bytesConsumed = acceptCondJump(i, data, "jg")
		} else if data[i] == 0b01110011 {
			asm, bytesConsumed = acceptCondJump(i, data, "jnb")
		} else if data[i] == 0b01110111 {
			asm, bytesConsumed = acceptCondJump(i, data, "ja")
		} else if data[i] == 0b01111011 {
			asm, bytesConsumed = acceptCondJump(i, data, "jnp")
		} else if data[i] == 0b01110001 {
			asm, bytesConsumed = acceptCondJump(i, data, "jno")
		} else if data[i] == 0b01111001 {
			asm, bytesConsumed = acceptCondJump(i, data, "jns")
		} else if data[i] == 0b11100010 {
			asm, bytesConsumed = acceptCondJump(i, data, "loop")
		} else if data[i] == 0b11100001 {
			asm, bytesConsumed = acceptCondJump(i, data, "loopz")
		} else if data[i] == 0b11100000 {
			asm, bytesConsumed = acceptCondJump(i, data, "loopnz")
		} else if data[i] == 0b11100011 {
			asm, bytesConsumed = acceptCondJump(i, data, "jcxz")
		} else {
			fmt.Printf("%b\n", data[i])
			panic("That wasn't supposed to happen")
		}
		if DEBUG {
			for j, b := range data[i : i+int(bytesConsumed)] {
				if j > 0 {
					fmt.Print(" ")
				}
				fmt.Printf("%08b", b)
			}
			fmt.Printf(": %s (%d)\n", asm, bytesConsumed)
		} else {
			fmt.Println(asm)
		}
		i += int(bytesConsumed)
	}

}

func printFlags() {
	if flagValues.zero {
		fmt.Printf("S")
	}
	if flagValues.zero {
		fmt.Printf("Z")
	}
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run part1.go <filename>")
		return
	}
	filename := os.Args[1]
	decode(filename)
	fmt.Printf("Final registers:\n")
	if registerValues[0] != 0 {
		fmt.Printf("      ax: 0x%04X (%d)\n", registerValues[0], registerValues[0])
	}
	if registerValues[3] != 0 {
		fmt.Printf("      bx: 0x%04X (%d)\n", registerValues[3], registerValues[3])
	}
	if registerValues[1] != 0 {
		fmt.Printf("      cx: 0x%04X (%d)\n", registerValues[1], registerValues[1])
	}
	if registerValues[2] != 0 {
		fmt.Printf("      dx: 0x%04X (%d)\n", registerValues[2], registerValues[2])
	}
	if registerValues[4] != 0 {
		fmt.Printf("      sp: 0x%04X (%d)\n", registerValues[4], registerValues[4])
	}
	if registerValues[5] != 0 {
		fmt.Printf("      bp: 0x%04X (%d)\n", registerValues[5], registerValues[5])
	}
	if registerValues[6] != 0 {
		fmt.Printf("      si: 0x%04X (%d)\n", registerValues[6], registerValues[6])
	}
	if registerValues[7] != 0 {
		fmt.Printf("      di: 0x%04X (%d)\n", registerValues[7], registerValues[7])
	}
	fmt.Printf("   flags: ")
	printFlags()

}
