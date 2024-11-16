package disasm

import (
	"github.com/phantasma-io/phantasma-go/pkg/vm"
)

func ExtractContractNamesWithDisasm(disassembler *Disassembler) []string {
	var instructions = disassembler.Instructions
	var result []string

	index := 0
	var regs = make([]vm.VMObject, 16, 16)
	for index < len(instructions) {
		var instruction = instructions[index]

		switch instruction.Opcode {
		case vm.LOAD:
			{
				var dst = instruction.Args[0].(byte)
				var _type = instruction.Args[1].(vm.VMType)
				var bytes = instruction.Args[2].([]byte)

				regs[dst] = vm.VMObject{}
				regs[dst].SetValue(bytes, _type)

				break
			}

		case vm.CTX:
			{
				var src = instruction.Args[0].(byte)
				var dst = instruction.Args[1].(byte)

				regs[dst] = vm.VMObject{}
				regs[dst].Copy(&regs[src])
				break
			}

		case vm.SWITCH:
			{
				var src = instruction.Args[0].(byte)

				var contractName = regs[src].AsString()
				result = append(result, contractName)
				break
			}
		}

		index++
	}

	return result // TODO add distinct
}

func ExtractContractNames(script []byte, debugLogging bool) []string {
	var disassembler = NewDisassembler(script, debugLogging)
	return ExtractContractNamesWithDisasm(disassembler)
}
