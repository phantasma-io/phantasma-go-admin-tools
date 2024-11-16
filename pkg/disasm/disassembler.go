package disasm

import (
	"encoding/binary"
	"fmt"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/disasm/types"
	"github.com/phantasma-io/phantasma-go/pkg/vm"
)

type Disassembler struct {
	script             []byte
	Instructions       []types.Instruction
	instructionPointer uint
	DebugLogging       bool
}

func NewDisassembler(script []byte, debugLogging bool) *Disassembler {

	d := Disassembler{script: script, DebugLogging: debugLogging}

	d.Instructions = d.GetInstructions()

	return &d
}

func (ds *Disassembler) Read8() byte {
	if ds.DebugLogging {
		fmt.Printf("Read8(): ds.instructionPointer: %d\n", ds.instructionPointer)
	}
	if ds.instructionPointer >= uint(len(ds.script)) {
		panic("Outside of range")
	}

	result := ds.script[ds.instructionPointer]
	if ds.DebugLogging {
		fmt.Printf("Read8(): result: %x\n", result)
	}
	ds.instructionPointer++
	return result
}

func (ds *Disassembler) Read16() uint16 {
	if ds.DebugLogging {
		fmt.Printf("Read16(): ds.instructionPointer: %d\n", ds.instructionPointer)
	}
	return binary.LittleEndian.Uint16([]byte{ds.Read8(), ds.Read8()})
}

func (ds *Disassembler) Read32() uint32 {
	if ds.DebugLogging {
		fmt.Printf("Read32(): ds.instructionPointer: %d\n", ds.instructionPointer)
	}
	return binary.LittleEndian.Uint32([]byte{ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8()})
}

func (ds *Disassembler) Read64() uint64 {
	if ds.DebugLogging {
		fmt.Printf("Read64(): ds.instructionPointer: %d\n", ds.instructionPointer)
	}
	return binary.LittleEndian.Uint64([]byte{ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8(), ds.Read8()})
}

func (ds *Disassembler) ReadVar(max uint64) uint64 {
	n := ds.Read8()

	var val uint64

	switch n {
	case 0xFD:
		val = uint64(ds.Read16())
		break
	case 0xFE:
		val = uint64(ds.Read32())
		break
	case 0xFF:
		val = ds.Read64()
		break
	default:
		val = uint64(n)
		break
	}

	if val > max {
		panic("Input exceed max")
	}

	return val
}

func (ds *Disassembler) ReadBytes(length int) []byte {

	if ds.DebugLogging {
		fmt.Printf("ReadBytes(): ds.instructionPointer: %d, length: %d\n", ds.instructionPointer, length)
	}

	if ds.instructionPointer+uint(length) >= uint(len(ds.script)) {
		panic("Outside of range")
	}

	result := make([]byte, length, length)
	for i := 0; i < length; i++ {
		result[i] = ds.script[ds.instructionPointer]
		ds.instructionPointer++
	}

	return result
}

func (ds *Disassembler) GetInstructions() []types.Instruction {
	ds.instructionPointer = 0

	result := []types.Instruction{}

	if ds.DebugLogging {
		fmt.Printf("GetInstructions(): len(ds.script): %d\n", len(ds.script))
	}

	for ds.instructionPointer < uint(len(ds.script)) {
		var temp types.Instruction
		temp.Offset = uint(ds.instructionPointer)
		var opByte = ds.Read8()
		temp.Opcode = vm.Opcode(opByte)

		if ds.DebugLogging {
			fmt.Printf("temp.Opcode: %d\n", temp.Opcode)
		}

		switch temp.Opcode {
		case vm.RET:
			temp.Args = []any{}
			result = append(result, temp)
			return result

		// args: byte src_reg, byte dest_reg
		case vm.CTX, vm.MOVE, vm.COPY, vm.SWAP, vm.SIZE, vm.COUNT, vm.SIGN, vm.NOT, vm.NEGATE, vm.ABS, vm.UNPACK, vm.REMOVE:
			src := ds.Read8()
			dst := ds.Read8()

			temp.Args = []any{src, dst}

		// args: byte dst_reg, byte type, var length, var data_bytes
		case vm.LOAD:
			dst := ds.Read8()
			_type := vm.VMType(ds.Read8())
			len := int(ds.ReadVar(0xFFFF))

			bytes := ds.ReadBytes(len)

			temp.Args = []any{dst, _type, bytes}

		case vm.CAST:
			src := ds.Read8()
			dst := ds.Read8()
			_type := vm.VMType(ds.Read8())

			temp.Args = []any{src, dst, _type}

		// args: byte src_reg
		case vm.POP, vm.PUSH, vm.EXTCALL, vm.THROW, vm.CLEAR:
			src := ds.Read8()
			temp.Args = []any{src}

		// args: ushort offset, byte regCount
		case vm.CALL:
			count := ds.Read8()
			ofs := ds.Read16()
			temp.Args = []any{count, ofs}

		// args: ushort offset, byte src_reg
		// NOTE: JMP only has offset arg, not the rest
		case vm.JMP, vm.JMPIF, vm.JMPNOT:
			if temp.Opcode == vm.JMP {
				newPos := ds.Read16()
				temp.Args = []any{newPos}
			} else {
				src := ds.Read8()
				newPos := ds.Read16()
				temp.Args = []any{src, newPos}
			}

		// args: byte src_a_reg, byte src_b_reg, byte dest_reg
		case vm.AND, vm.OR, vm.XOR, vm.CAT, vm.EQUAL, vm.LT, vm.GT, vm.LTE, vm.GTE:
			srcA := ds.Read8()
			srcB := ds.Read8()
			dst := ds.Read8()

			temp.Args = []any{srcA, srcB, dst}

		// args: byte src_reg, byte dest_reg, var length
		case vm.LEFT, vm.RIGHT:
			src := ds.Read8()
			dst := ds.Read8()
			len := uint16(ds.ReadVar(0xFFFF))

			temp.Args = []any{src, dst, len}

		// args: byte src_reg, byte dest_reg, var index, var length
		case vm.RANGE:
			src := ds.Read8()
			dst := ds.Read8()
			index := int(ds.ReadVar(0xFFFF))
			len := int(ds.ReadVar(0xFFFF))

			temp.Args = []any{src, dst, index, len}

		// args: byte reg
		case vm.INC, vm.DEC, vm.SWITCH:
			dst := ds.Read8()
			temp.Args = []any{dst}

		// args: byte src_a_reg, byte src_b_reg, byte dest_reg
		case vm.ADD, vm.SUB, vm.MUL, vm.DIV, vm.MOD, vm.SHR, vm.SHL, vm.MIN, vm.MAX, vm.POW, vm.PUT, vm.GET:
			srcA := ds.Read8()
			srcB := ds.Read8()
			dst := ds.Read8()
			temp.Args = []any{srcA, srcB, dst}

		default:
			temp.Args = []any{}
		}

		result = append(result, temp)
	}

	return result
}
