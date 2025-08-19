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

func decodeMoves(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Printf("bits 16\n\n")
	for i := 0; i < len(data) - 1; i+=2 {
		movByte1 := data[i]
		w := movByte1 & 1 > 0
		d := movByte1 & 2 > 0
		movByte2 := data[i+1]
		// mod part not necessary since it will always be 11 (reg to reg)
		reg := (movByte2 & 0x38) >> 3
		mr := movByte2 & 0x7
		// fmt.Printf("mov byte1: %08b\n", movByte1)
		// fmt.Printf("m: %d\n", w)
		// fmt.Printf("d: %d\n", d)
		// fmt.Printf("mov byte2: %08b\n", movByte2)
		// fmt.Printf("mask: %08b\n", 0x38)
		// fmt.Printf("reg: %08b\n", reg)
		// fmt.Printf("mr: %08b\n", mr)

		var dst string;
		var src string;
		var offset uint8
		if w {
			offset = 8
		} else {
			offset = 0
		}
		// fmt.Printf("offset: %d\n", offset)
		if d {
			dst = RegMap[uint8(reg) + offset]
			src = RegMap[uint8(mr) + offset]
		} else {
			dst = RegMap[uint8(mr) + offset]
			src = RegMap[uint8(reg) + offset]
		}
		fmt.Printf("mov %s,%s\n", dst, src)
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
