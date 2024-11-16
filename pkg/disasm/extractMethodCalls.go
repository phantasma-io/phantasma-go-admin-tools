package disasm

import (
	"strings"

	"github.com/phantasma-io/phantasma-go-admin-tools/pkg/disasm/types"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/phantasma-io/phantasma-go/pkg/vm"
)

type DisasmMethodCall struct {
	ContractName string
	MethodName   string
	Arguments    []vm.VMObject
}

func (call *DisasmMethodCall) ToString(useNewlines bool) string {
	var sb strings.Builder
	sb.WriteString(call.ContractName)
	sb.WriteString(".")
	sb.WriteString(call.MethodName)
	sb.WriteString("(")
	for i := 0; i < len(call.Arguments); i++ {
		if i > 0 {
			sb.WriteByte(',')
			if useNewlines {
				sb.WriteByte('\n')
			}
		}

		var arg = call.Arguments[i]
		sb.WriteString(arg.String()) // TODO in original code it was using ToString()
	}
	sb.WriteString(")")
	return sb.String()
}

func (call *DisasmMethodCall) String() string {
	return call.ToString(true)
}

func PopArgs(contract, method string, stack types.Stack[vm.VMObject], methodArgumentCountTable *orderedmap.OrderedMap[string, int]) []vm.VMObject {

	var key = method
	if contract != "" {
		key = contract + "." + method
	}

	p := methodArgumentCountTable.GetPair(key)
	if p != nil {
		var argCount = p.Value
		var result = make([]vm.VMObject, argCount, argCount)
		for i := 0; i < argCount; i++ {
			result[i] = stack.Pop()
		}
		return result
	} else {
		panic("Cannot disassemble method arguments => " + key)
	}
}

func ExtractMethodCallsWithDisasm(disassembler *Disassembler, methodArgumentCountTable *orderedmap.OrderedMap[string, int]) []DisasmMethodCall {
	var instructions = disassembler.Instructions
	// fmt.Printf("len(instructions): %d\n", len(instructions))
	var result = []DisasmMethodCall{}

	index := 0
	var regs = make([]vm.VMObject, 16, 16)
	var stack = types.Stack[vm.VMObject]{}
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

		case vm.PUSH:
			{
				var src = instruction.Args[0].(byte)
				var val = regs[src]

				var temp = vm.VMObject{}
				temp.Copy(&val)
				stack.Push(temp)
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
				v := stack.Pop()
				var methodName = v.AsString()
				var args = PopArgs(contractName, methodName, stack, methodArgumentCountTable)
				result = append(result, DisasmMethodCall{MethodName: methodName, ContractName: contractName, Arguments: args})
				break
			}

		case vm.EXTCALL:
			{
				var src = instruction.Args[0].(byte)
				var methodName = regs[src].AsString()
				var args = PopArgs("", methodName, stack, methodArgumentCountTable)
				result = append(result, DisasmMethodCall{MethodName: methodName, ContractName: "", Arguments: args})
				break
			}
		}

		index++
	}

	return result
}

func ExtractMethodCalls(script []byte, methodArgumentCountTable *orderedmap.OrderedMap[string, int], debugLogging bool) ([]DisasmMethodCall, uint) {
	var disassembler = NewDisassembler(script, debugLogging)
	return ExtractMethodCallsWithDisasm(disassembler, methodArgumentCountTable), disassembler.instructionPointer
}
