package main

import (
	"fmt"
	"os"
)

var RegMap = map[uint8]string{
	0: "al",
	1: "cl",
	2: "dl",
	3: "bl",
	4: "ah",
	5: "ch",
	6: "dh",
	7: "bh",
	8: "ax",
	9: "cx",
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

func acceptRegMemMov(start_idx int, data []byte) (string, uint) {
	d := data[start_idx] & 2 > 0
	w := data[start_idx] & 1 > 0
	b2 := data[start_idx + 1]
	mod := (b2 & 0b11000000) >> 6
	reg := (b2 & 0b00111000) >> 3
	rm := b2 & 0b00000111

	eff_addr_base := EffAddrBaseMap[uint8(rm)]

	var reg_str string;
	var other string;
	var bytes_consumed uint8;
	if w {
		reg_str = RegMap[uint8(reg) + 8]
	} else {
		reg_str = RegMap[uint8(reg)]
	}
	switch mod {
		case 0:
			if rm == 6 {
				//dir addr
				//TODO
				return "TODO", 4
			} else {
				other = fmt.Sprintf("[%s]", eff_addr_base)
				bytes_consumed = 2
			}
		case 1:
			disp_lo := data[start_idx + 2]
			other = fmt.Sprintf("[%s + %d]", eff_addr_base, disp_lo)
			bytes_consumed = 3
		case 2:
			disp_lo := data[start_idx + 2]
			disp_hi := data[start_idx + 3]
			disp := (uint16(disp_hi) << 8) | uint16(disp_lo)
			other = fmt.Sprintf("[%s + %d]", eff_addr_base, disp)
			bytes_consumed = 4
		case 3:
			if w {
				other = RegMap[uint8(rm) + 8]
			} else {
				other = RegMap[uint8(rm)]
			}
			bytes_consumed = 2
		default:
			panic("MOD should never have this value")
	}
	var src, dst string;
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

func acceptImmediateToReg(start_idx int, data []byte) (string, uint) {
	w := data[start_idx] & 8 > 0
	reg := data[start_idx] & 0b00000111

	data_lo := data[start_idx + 1]
	if w {
		data_hi := data[start_idx + 2]
		data := (uint16(data_hi) << 8) | uint16(data_lo)
		reg_str := RegMap[uint8(reg) + 8]
		return fmt.Sprintf("mov %s, %d", reg_str, data), 3
	} else {
		reg_str := RegMap[uint8(reg)]
		return fmt.Sprintf("mov %s, %d", reg_str, data_lo), 2
	}
}

func decodeMoves(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Printf("bits 16\n\n")
	for i := 0; i < len(data); {
		var asm string;
		var bytes_consumed uint;
		if (0b10001000 ^ data[i]) >> 2 == 0 {
			asm, bytes_consumed = acceptRegMemMov(i, data)
		} else  if (0b10110000 ^ data[i]) >> 4 == 0 {
			asm, bytes_consumed = acceptImmediateToReg(i, data)
		} else {
			fmt.Printf("%b\n", data[i])
			panic("That wasn't supposed to happen")
		}
		fmt.Println(asm)
		i += int(bytes_consumed)
	}

}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run part1.go <filename>")
		return
	}
	filename := os.Args[1]
	decodeMoves(filename)
}
