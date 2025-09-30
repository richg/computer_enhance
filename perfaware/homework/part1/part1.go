package main

import (
	"fmt"
	"os"
)

const DEBUG = false

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

type EffectiveAddress struct {
	 base string
	 disp uint16
	 dispSize uint8
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

type ArithmeticOp struct {
	opType ArithmeticOpType
	variant OperandVariant
	wide bool
	reg Register
	otherReg Register
	otherEffAddr EffectiveAddress
	otherImmediate Immediate
	size uint8
}


func (op ArithmeticOp) String() string {
	var opMnemonic string
	switch op.opType {
		case Add:
			opMnemonic = "add"
		case Sub:
			opMnemonic = "sub"
		case Cmp:
			opMnemonic = "cmp"
		default:
			opMnemonic = "NOT_SUPPORTED"
	}
	switch op.variant {
		case REG_REG:
			return fmt.Sprintf("%s %s, %s", opMnemonic, op.reg.name, op.otherReg.name)
		case REG_MEM:
			return fmt.Sprintf("%s %s, %s", opMnemonic, op.reg.name, op.otherEffAddr.String())
		case MEM_REG:
			return fmt.Sprintf("%s %s, %s", opMnemonic, op.otherEffAddr.String(), op.reg.name)
		case REG_IMM:
			return fmt.Sprintf("%s %s, %d", opMnemonic, op.reg.name, op.otherImmediate.value)
		case MEM_IMM:
			clarifier := ""
			if op.wide {
				clarifier = " word"
			} else {
				clarifier = " byte"
			}
			return fmt.Sprintf("%s%s %s, %d", opMnemonic, clarifier, op.otherEffAddr.String(), op.otherImmediate)
		default:
			panic("Unexpected Error")
	}
}

func (op ArithmeticOp) execute() {
	var operand1 uint16
	var operand2 uint16
	var result uint16
	switch op.variant {
		case REG_REG:
			operand1 = getRegisterValue(op.reg)
			operand2 = getRegisterValue(op.otherReg)
			result = doOp(op.opType, operand1, operand2)
			updateRegister(op.reg, result)
		case REG_MEM:
			operand1 = getRegisterValue(op.reg)
			operand2 = getRegisterValue(op.otherReg)
			result = doOp(op.opType, operand1, operand2)
			updateRegister(op.reg, result)
		case MEM_REG:
			operand1 = getRegisterValue(op.reg)
			operand2 = getRegisterValue(op.otherReg)
			result = doOp(op.opType, operand1, operand2)
		case REG_IMM:
			operand1 = getRegisterValue(op.reg)
			operand2 = op.otherImmediate.value
			result = doOp(op.opType, operand1, operand2)
			updateRegister(op.reg, result)
		case MEM_IMM:
			operand1 = getRegisterValue(op.reg)
			operand2 = op.otherImmediate.value
			result = doOp(op.opType, operand1, operand2)
		default:
			panic("Unexpected Error")
	}
	flagValues.zero = result == 0
	flagValues.sign = result & 0x8000 > 0
}

func doOp(opType ArithmeticOpType, lhs uint16, rhs uint16) uint16 {
	switch opType {
		case Add:
			return lhs + rhs
		case Sub:
			return lhs - rhs
		case Cmp:
			return lhs - rhs
		default:
			panic(fmt.Sprintf("Unsupported opType %x", opType))
	}
}



type JmpOpType = uint8
const (
	JE JmpOpType = iota
	JL
	JLE
	JB
	JBE
	JP
	JO
	JS
	JNZ
	JNL
	JG
	JNB
	JA
	JNP
	JNO
	JNS
	LOOP
	LOOPZ
	LOOPNZ
	JCXZ
)

type JumpOp struct {
	opType JmpOpType
	disp uint16
	size uint8
}

func (op JumpOp) String() string {
	return fmt.Sprintf("%s %d", getJumpOpName(op), op.disp)
}

func (op JumpOp) execute() {
	switch (op.opType) {
	case JNZ:
		if !flagValues.zero {
			// fmt.Printf("IP: %d, %x\n", ip, ip)
			// fmt.Printf("disp: %d\n", int(op.disp))
			ip += uint16(op.disp)
			// fmt.Printf("IP: %d, %x\n", ip, ip)
		}
	default:
		panic("Not implemented")
	}
}


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

var jumpOpName = [20]string {
	"je",
	"jl",
	"jle",
	"jb",
	"jbe",
	"jp",
	"jo",
	"js",
	"jnz",
	"jnl",
	"jg",
	"jnb",
	"ja",
	"jnp",
	"jno",
	"jns",
	"loop",
	"loopz",
	"loopnz",
	"jcxz",
}

func getJumpOpName(op JumpOp) string {
	return jumpOpName[op.opType]
}



// CPU and Memory state

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

var ip uint16;

var memory = [65536]byte{}

func load(memIdx int16) (byte, byte)  {
	return memory[memIdx] , memory[memIdx+1 % 65536]
}

// func store(memIdx int16, , word bool) {
// 	if word {
// 	} else {
// 	}
// }


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


func resolveOther(startIdx uint16, data []byte, mod byte, rm byte, w bool) (Register, EffectiveAddress) {
	var otherReg Register
	var otherEffAddr EffectiveAddress
	effAddrBase := EffAddrBaseMap[rm]
	switch mod {
		case 0:
			if rm == 6 {
				dispLo := data[startIdx+2]
				dispHi := data[startIdx+3]
				disp := (uint16(dispHi) << 8) | uint16(dispLo)
				otherEffAddr = EffectiveAddress{ base: "", disp: disp , dispSize: 2}
			} else {
				otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: 0, dispSize: 0}
			}
		case 1:
			dispLo := uint16(data[startIdx+2])
			otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: dispLo, dispSize: 1}
		case 2:
			dispLo := data[startIdx+2]
			dispHi := data[startIdx+3]
			disp := (uint16(dispHi) << 8) | uint16(dispLo)
			otherEffAddr = EffectiveAddress{ base: effAddrBase, disp: disp, dispSize:2}
		case 3:
			if w {
				otherReg = RegMap[uint8(rm)+8]
			} else {
				otherReg = RegMap[uint8(rm)]
			}
		default:
			panic("MOD should never have this value")
	}
	return otherReg, otherEffAddr
}

func getImmediate(startIdx uint16, data []byte, wide bool) Immediate {
	if wide {
		return Immediate{
			uint16(uint16(data[startIdx+1])<<8 | uint16(data[startIdx])),
			2,
		}
	} else {
		return Immediate{
			uint16(int8(data[startIdx])),
			1,
		}
	}
}

func getOtherString(otherReg Register, otherEffAddr EffectiveAddress) string {
	if otherReg.name ==  "" {
		return otherEffAddr.String()
	} else {
		return otherReg.name
	}
}

type Immediate struct {
	value uint16
	size uint8
}

type OperandVariant = uint8
const (
	REG_REG OperandVariant = iota
	REG_MEM
	REG_IMM
	MEM_REG
	MEM_IMM
)

type MovOp struct {
	variant OperandVariant
	wide bool
	reg Register
	otherReg Register
	effAddr EffectiveAddress
	immediate Immediate
	size uint8
}

func (op MovOp) String() string {
	var src, dst string
	switch (op.variant) {
		case REG_REG:
			dst = op.reg.name
			src = op.otherReg.name
		case REG_MEM:
			dst = op.reg.name
			src = op.effAddr.String()
		case REG_IMM:
			dst = op.reg.name
			src = fmt.Sprintf("%d", op.immediate.value)
		case MEM_REG:
			dst = op.effAddr.String()
			src = op.reg.name
	}
	return fmt.Sprintf("mov %s, %s", src, dst)
}

func (op MovOp) execute() {
	switch (op.variant) {
		case REG_REG:
			updateRegister(op.reg, getRegisterValue(op.otherReg))
		case REG_MEM:
			panic("NOT YET")
		case REG_IMM:
			updateRegister(op.reg, op.immediate.value)
		case MEM_REG:
			panic("NOT YET")
	}
}

func acceptRegMemMov(startIdx uint16, data []byte) MovOp {
	d := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111
	regBits := (b2 & 0b00111000) >> 3

	var reg Register
	if w {
		reg = RegMap[uint8(regBits)+8]
	} else {
		reg = RegMap[uint8(regBits)]
	}
	otherReg, otherEffAddr := resolveOther(startIdx, data, mod, rm, w)
	var op MovOp
	if otherReg.name !=  "" {
		var dst, src Register
		if d {
			dst = reg
			src = otherReg
		} else {
			dst = otherReg
			src = reg
		}
		op = MovOp{
			REG_REG,
			w,
			dst,
			src,
			EffectiveAddress{},
			Immediate{},
			2,
		}
	} else {
		var variant = REG_MEM
		if !d {
			variant = MEM_REG
		}
		op = MovOp{
			variant,
			w,
			reg,
			Register{},
			otherEffAddr,
			Immediate{},
			2 + otherEffAddr.dispSize,
		}
	}
	return op
}

func acceptImmediateToRegMov(startIdx uint16, data []byte) MovOp {
	w := data[startIdx]&8 > 0
	regBits := data[startIdx] & 0b00000111

	immediate := getImmediate(startIdx+1, data, w)

	var reg Register
	if w {
		reg = RegMap[regBits+8]
	} else {
		reg = RegMap[regBits]
	}
	return MovOp{
		REG_IMM,
		w,
		reg,
		Register{},
		EffectiveAddress{},
		immediate,
		immediate.size + 1,
	}
}

func acceptArithmeticOpRegMem(startIdx uint16, data []byte) ArithmeticOp {
	d := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	regBits := (b2 & 0b00111000) >> 3
	rm := b2 & 0b00000111

	var reg Register
	if w {
		reg = RegMap[uint8(regBits)+8]
	} else {
		reg = RegMap[uint8(regBits)]
	}
	opType := uint8(data[startIdx] & 0b00111000 >> 3)
	otherReg, otherEffAddr := resolveOther(startIdx, data, mod, rm, w)
	var op ArithmeticOp
	if otherReg.name !=  "" {
		var dst, src Register
		if d {
			dst = reg
			src = otherReg
		} else {
			dst = otherReg
			src = reg
		}
		op = ArithmeticOp{
			opType,
			REG_REG,
			w,
			dst,
			src,
			EffectiveAddress{},
			Immediate{},
			2,
		}
	} else {
		var variant = REG_MEM
		if !d {
			variant = MEM_REG
		}
		op = ArithmeticOp{
			opType,
			variant,
			w,
			reg,
			Register{},
			otherEffAddr,
			Immediate{},
			2 + otherEffAddr.dispSize,
		}
	}
	return op
}

func acceptArithmeticOpRegMemImmediate(startIdx uint16, data []byte) ArithmeticOp {
	s := data[startIdx]&2 > 0
	w := data[startIdx]&1 > 0
	b2 := data[startIdx+1]
	mod := (b2 & 0b11000000) >> 6
	rm := b2 & 0b00000111

	dstReg, otherEffAddr := resolveOther(startIdx, data, mod, rm, w)
	wideImmediate := w && !s
	immediate := getImmediate(startIdx + 2 + uint16(otherEffAddr.dispSize) , data, wideImmediate)
	opType := uint8(data[startIdx+1] & 0b00111000 >> 3)
	var op ArithmeticOp
	if dstReg.name !=  "" {
		op = ArithmeticOp{
			opType,
			REG_IMM,
			w,
			dstReg,
			Register{},
			EffectiveAddress{},
			immediate,
			immediate.size + 2,
		}
	} else {
		op = ArithmeticOp{
			opType,
			MEM_IMM,
			w,
			Register{},
			Register{},
			otherEffAddr,
			immediate,
			immediate.size + otherEffAddr.dispSize + 2,
		}
	}
	return op
}

func acceptArithmeticOpImmediateAcc(startIdx uint16, data []byte) ArithmeticOp {
	w := data[startIdx]&1 > 0
	fmt.Printf("Byte 1: %b", data[startIdx])
	fmt.Printf("Byte 2: %b", data[startIdx + 1])
	immediate := getImmediate(startIdx+1, data, w)
	opType := uint8(data[startIdx] & 0b00111000 >> 3)
	var reg Register
	regIdx := 0 // Index of the A register
	if w {
		reg = RegMap[regIdx + 8]
	} else {
		reg = RegMap[regIdx]
	}
	return ArithmeticOp{
		opType,
		REG_IMM,
		w,
		reg,
		Register{},
		EffectiveAddress{},
		immediate,
		immediate.size + 1,
	}
}

func acceptCondJump(startIdx uint16, data []byte, opType JmpOpType) JumpOp {
	displacement := uint16(int8(data[startIdx+1]))
	op := JumpOp {
		opType,
		displacement,
		2,
	}
	return op
}

func decode(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println("Error reading file:", err)
		return
	}
	fmt.Printf("bits 16\n\n")
	streamSize := uint16(len(data))
	i := 0
	for ip < streamSize {
		var jmpOp JumpOp
		var movOp MovOp
		var arithmeticOp ArithmeticOp
		if (0b10001000^data[ip])&0b11111100 == 0 {
			movOp = acceptRegMemMov(ip, data)
		} else if (0b10110000^data[ip])&0b11110000 == 0 {
			movOp = acceptImmediateToRegMov(ip, data)
		} else if (0b00000000^data[ip])&0b11111100 == 0 ||
			(0b00101000^data[ip])&0b11111100 == 0 ||
			(0b00111000^data[ip])&0b11111100 == 0 {
			arithmeticOp = acceptArithmeticOpRegMem(ip, data)
		} else if (0b10000000^data[ip])&0b11111100 == 0 {
			arithmeticOp = acceptArithmeticOpRegMemImmediate(ip, data)
		} else if (0b00000100^data[ip])&0b11111110 == 0 ||
			(0b00101100^data[ip])&0b11111110 == 0 ||
			(0b00111100^data[ip])&0b11111110 == 0 {
			arithmeticOp = acceptArithmeticOpImmediateAcc(ip, data)
		} else if data[ip] == 0b01110100 {
			jmpOp = acceptCondJump(ip, data, JE)
		} else if data[ip] == 0b01111100 {
			jmpOp = acceptCondJump(ip, data, JL)
		} else if data[ip] == 0b01111110 {
			jmpOp = acceptCondJump(ip, data, JLE)
		} else if data[ip] == 0b01110010 {
			jmpOp = acceptCondJump(ip, data, JB)
		} else if data[ip] == 0b01110110 {
			jmpOp = acceptCondJump(ip, data, JBE)
		} else if data[ip] == 0b01111010 {
			jmpOp = acceptCondJump(ip, data, JP)
		} else if data[ip] == 0b01110000 {
			jmpOp = acceptCondJump(ip, data, JO)
		} else if data[ip] == 0b01111000 {
			jmpOp = acceptCondJump(ip, data, JS)
		} else if data[ip] == 0b01110101 {
			jmpOp = acceptCondJump(ip, data, JNZ)
		} else if data[ip] == 0b01111101 {
			jmpOp = acceptCondJump(ip, data, JNL)
		} else if data[ip] == 0b01111111 {
			jmpOp = acceptCondJump(ip, data, JG)
		} else if data[ip] == 0b01110011 {
			jmpOp = acceptCondJump(ip, data, JNB)
		} else if data[ip] == 0b01110111 {
			jmpOp = acceptCondJump(ip, data, JA)
		} else if data[ip] == 0b01111011 {
			jmpOp = acceptCondJump(ip, data, JNP)
		} else if data[ip] == 0b01110001 {
			jmpOp = acceptCondJump(ip, data, JNO)
		} else if data[ip] == 0b01111001 {
			jmpOp = acceptCondJump(ip, data, JNS)
		} else if data[ip] == 0b11100010 {
			jmpOp = acceptCondJump(ip, data, LOOP)
		} else if data[ip] == 0b11100001 {
			jmpOp = acceptCondJump(ip, data, LOOPZ)
		} else if data[ip] == 0b11100000 {
			jmpOp = acceptCondJump(ip, data, LOOPNZ)
		} else if data[ip] == 0b11100011 {
			jmpOp = acceptCondJump(ip, data, JCXZ)
		} else {
			fmt.Printf("%b\n", data[ip])
			panic("That wasn't supposed to happen")
		}

		var bytesConsumed uint
		if jmpOp.size != 0 {
			fmt.Println(jmpOp.String())
			bytesConsumed += uint(jmpOp.size)
		} else if movOp.size != 0 {
			fmt.Println(movOp.String())
			bytesConsumed += uint(movOp.size)
		} else if arithmeticOp.size != 0 {
			fmt.Println(arithmeticOp.String())
			bytesConsumed += uint(arithmeticOp.size)
		} else {
			fmt.Printf("%b\n", data[ip])
			panic(fmt.Sprintf("Unmatched op: %b\n", data[ip]))
		}
		ip += uint16(bytesConsumed)
		if jmpOp.size != 0 {
			jmpOp.execute()
		} else if movOp.size != 0 {
			movOp.execute()
		} else if arithmeticOp.size != 0 {
			arithmeticOp.execute()
		} else {
			fmt.Printf("%b\n", data[ip])
			panic(fmt.Sprintf("Unmatched op: %b\n", data[ip]))
		}
		fmt.Printf("IP: %d\n", ip)
		i++
	}

}

func printFlags() {
	if flagValues.sign {
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
	fmt.Printf("      ip: 0x%04X (%d)\n", ip, ip)
	fmt.Printf("   flags: ")
	printFlags()
	fmt.Println()
}
