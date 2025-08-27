package main

import (
	"fmt"
	"os"
)

const DEBUG = false

var RegMap = map[uint8]string{
	0:  "al",
	1:  "cl",
	2:  "dl",
	3:  "bl",
	4:  "ah",
	5:  "ch",
	6:  "dh",
	7:  "bh",
	8:  "ax",
	9:  "cx",
	10: "dx",
	11: "bx",
	12: "sp",
	13: "bp",
	14: "si",
	15: "di",
}

var EffAddrBaseMap = map[uint8]string{
	0: "bx + si",
	1: "bx + di",
	2: "bp + si",
	3: "bp + di",
	4: "si",
	5: "di",
	6: "bp",
	7: "bx",
}

var ArithmeticVariantLookup = map[uint8]string{
	0: "add",
	1: "NOT_SUPPORTED",
	2: "NOT_SUPPORTED",
	3: "NOT_SUPPORTED",
	4: "NOT_SUPPORTED",
	5: "sub",
	6: "NOT_SUPPORTED",
	7: "cmp",
}

func resolveOther(start_idx int, data []byte, mod byte, rm byte, w bool) (string, bool, uint8) {
	var other string
	var bytes_consumed uint8
	eff_addr_base := EffAddrBaseMap[uint8(rm)]
	isEffectiveAddr := true
	switch mod {
	case 0:
		if rm == 6 {
			disp_lo := data[start_idx+2]
			disp_hi := data[start_idx+3]
			disp := (uint16(disp_hi) << 8) | uint16(disp_lo)
			other = fmt.Sprintf("[%d]", disp)
			bytes_consumed = 4
		} else {
			other = fmt.Sprintf("[%s]", eff_addr_base)
			bytes_consumed = 2
		}
	case 1:
		disp_lo := data[start_idx+2]
		other = fmt.Sprintf("[%s + %d]", eff_addr_base, disp_lo)
		bytes_consumed = 3
	case 2:
		disp_lo := data[start_idx+2]
		disp_hi := data[start_idx+3]
		disp := (uint16(disp_hi) << 8) | uint16(disp_lo)
		other = fmt.Sprintf("[%s + %d]", eff_addr_base, disp)
		bytes_consumed = 4
	case 3:
		isEffectiveAddr = false
		if w {
			other = RegMap[uint8(rm)+8]
		} else {
			other = RegMap[uint8(rm)]
		}
		bytes_consumed = 2
	default:
		panic("MOD should never have this value")
	}
	return other, isEffectiveAddr, uint8(bytes_consumed)
}

func getImmediate(start_idx int, data []byte, wide bool) (int16, uint8) {
	if wide {
		return int16(uint16(data[start_idx+1])<<8 | uint16(data[start_idx])), 2
	} else {
		return int16(int8(data[start_idx])), 1
	}
}

func acceptRegMemMov(start_idx int, data []byte) (string, uint) {
	d := data[start_idx]&2 > 0
	w := data[start_idx]&1 > 0
	b2 := data[start_idx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111
	reg := (b2 & 0b00111000) >> 3

	var reg_str string
	var other string
	var bytes_consumed uint8
	if w {
		reg_str = RegMap[uint8(reg)+8]
	} else {
		reg_str = RegMap[uint8(reg)]
	}
	other, _, bytes_consumed = resolveOther(start_idx, data, mod, rm, w)
	var src, dst string
	if d {
		src = other
		dst = reg_str
	} else {
		src = reg_str
		dst = other
	}
	asm := fmt.Sprintf("mov %s,%s", dst, src)
	return asm, uint(bytes_consumed)
}

func acceptImmediateToRegMov(start_idx int, data []byte) (string, uint) {
	w := data[start_idx]&8 > 0
	reg := data[start_idx] & 0b00000111

	immediate, _ := getImmediate(start_idx+1, data, w)

	if w {
		reg_str := RegMap[uint8(reg)+8]
		return fmt.Sprintf("mov %s, %d", reg_str, immediate), 3
	} else {
		reg_str := RegMap[uint8(reg)]
		return fmt.Sprintf("mov %s, %d", reg_str, immediate), 2
	}
}

func acceptArithmeticOpRegMem(start_idx int, data []byte) (string, uint) {
	d := data[start_idx]&2 > 0
	w := data[start_idx]&1 > 0
	b2 := data[start_idx+1]
	mod := (b2 & 0b11000000) >> 6
	reg := (b2 & 0b00111000) >> 3
	rm := b2 & 0b00000111

	var reg_str string
	var other string
	var bytes_consumed uint8
	if w {
		reg_str = RegMap[uint8(reg)+8]
	} else {
		reg_str = RegMap[uint8(reg)]
	}
	other, _, bytes_consumed = resolveOther(start_idx, data, mod, rm, w)
	var src, dst string
	if d {
		src = other
		dst = reg_str
	} else {
		src = reg_str
		dst = other
	}
	op := uint8(data[start_idx] & 0b00111000 >> 3)
	asm := fmt.Sprintf("%s %s, %s", ArithmeticVariantLookup[op], dst, src)
	return asm, uint(bytes_consumed)
}

func acceptArithmeticOpRegMemImmediate(start_idx int, data []byte) (string, uint) {
	s := data[start_idx]&2 > 0
	w := data[start_idx]&1 > 0
	b2 := data[start_idx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111

	var other string
	var bytes_consumed uint8
	var isEffectiveAddr bool
	other, isEffectiveAddr, bytes_consumed = resolveOther(start_idx, data, mod, rm, w)
	wideImmediate := w && !s
	immediate, im_bytes := getImmediate(start_idx+int(bytes_consumed), data, wideImmediate)
	bytes_consumed += im_bytes
	op := uint8(data[start_idx+1] & 0b00111000 >> 3)
	clarifier := ""
	if w {
		if isEffectiveAddr {
			clarifier = " word"
		}
	} else {
		if isEffectiveAddr {
			clarifier = " byte"
		}
	}
	asm := fmt.Sprintf("%s%s %s, %d", ArithmeticVariantLookup[op], clarifier, other, immediate)
	return asm, uint(bytes_consumed)
}

func acceptArithmeticOpImmediateAcc(start_idx int, data []byte) (string, uint) {
	w := data[start_idx]&1 > 0
	immediate, _ := getImmediate(start_idx+1, data, w)
	op := uint8(data[start_idx] & 0b00111000 >> 3)
	var bytes_consumed uint
	var reg string
	if w {
		reg = "ax"
		bytes_consumed = 3
	} else {
		reg = "al"
		bytes_consumed = 2
	}
	asm := fmt.Sprintf("%s %s, %d", ArithmeticVariantLookup[op], reg, immediate)
	return asm, bytes_consumed
}

func acceptCondJump(start_idx int, data []byte, jump_variant string) (string, uint) {
	displacement := int8(data[start_idx+1])
	asm := fmt.Sprintf("%s %d", jump_variant, displacement)
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
		var bytes_consumed uint
		if (0b10001000^data[i])&0b11111100 == 0 {
			asm, bytes_consumed = acceptRegMemMov(i, data)
		} else if (0b10110000^data[i])&0b11110000 == 0 {
			asm, bytes_consumed = acceptImmediateToRegMov(i, data)
		} else if (0b00000000^data[i])&0b11111100 == 0 ||
			(0b00101000^data[i])&0b11111100 == 0 ||
			(0b00111000^data[i])&0b11111100 == 0 {
			asm, bytes_consumed = acceptArithmeticOpRegMem(i, data)
		} else if (0b10000000^data[i])&0b11111100 == 0 {
			asm, bytes_consumed = acceptArithmeticOpRegMemImmediate(i, data)
		} else if (0b00000100^data[i])&0b11111110 == 0 ||
			(0b00101100^data[i])&0b11111110 == 0 ||
			(0b00111100^data[i])&0b11111110 == 0 {
			asm, bytes_consumed = acceptArithmeticOpImmediateAcc(i, data)
		} else if data[i] == 0b01110100 {
			asm, bytes_consumed = acceptCondJump(i, data, "je")
		} else if data[i] == 0b01111100 {
			asm, bytes_consumed = acceptCondJump(i, data, "jl")
		} else if data[i] == 0b01111110 {
			asm, bytes_consumed = acceptCondJump(i, data, "jle")
		} else if data[i] == 0b01110010 {
			asm, bytes_consumed = acceptCondJump(i, data, "jb")
		} else if data[i] == 0b01110110 {
			asm, bytes_consumed = acceptCondJump(i, data, "jbe")
		} else if data[i] == 0b01111010 {
			asm, bytes_consumed = acceptCondJump(i, data, "jp")
		} else if data[i] == 0b01110000 {
			asm, bytes_consumed = acceptCondJump(i, data, "jo")
		} else if data[i] == 0b01111000 {
			asm, bytes_consumed = acceptCondJump(i, data, "js")
		} else if data[i] == 0b01110101 {
			asm, bytes_consumed = acceptCondJump(i, data, "jnz")
		} else if data[i] == 0b01111101 {
			asm, bytes_consumed = acceptCondJump(i, data, "jnl")
		} else if data[i] == 0b01111111 {
			asm, bytes_consumed = acceptCondJump(i, data, "jg")
		} else if data[i] == 0b01110011 {
			asm, bytes_consumed = acceptCondJump(i, data, "jnb")
		} else if data[i] == 0b01110111 {
			asm, bytes_consumed = acceptCondJump(i, data, "ja")
		} else if data[i] == 0b01111011 {
			asm, bytes_consumed = acceptCondJump(i, data, "jnp")
		} else if data[i] == 0b01110001 {
			asm, bytes_consumed = acceptCondJump(i, data, "jno")
		} else if data[i] == 0b01111001 {
			asm, bytes_consumed = acceptCondJump(i, data, "jns")
		} else if data[i] == 0b11100010 {
			asm, bytes_consumed = acceptCondJump(i, data, "loop")
		} else if data[i] == 0b11100001 {
			asm, bytes_consumed = acceptCondJump(i, data, "loopz")
		} else if data[i] == 0b11100000 {
			asm, bytes_consumed = acceptCondJump(i, data, "loopnz")
		} else if data[i] == 0b11100011 {
			asm, bytes_consumed = acceptCondJump(i, data, "jcxz")
		} else {
			fmt.Printf("%b\n", data[i])
			panic("That wasn't supposed to happen")
		}
		if DEBUG {
			for j, b := range data[i : i+int(bytes_consumed)] {
				if j > 0 {
					fmt.Print(" ")
				}
				fmt.Printf("%08b", b)
			}
			fmt.Printf(": %s (%d)\n", asm, bytes_consumed)
		} else {
			fmt.Println(asm)
		}
		i += int(bytes_consumed)
	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run part1.go <filename>")
		return
	}
	filename := os.Args[1]
	decode(filename)
}
