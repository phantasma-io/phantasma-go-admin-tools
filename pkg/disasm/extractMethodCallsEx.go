package disasm

import (
	"github.com/phantasma-io/phantasma-go/pkg/rpc"
)

func getCallFullName(call DisasmMethodCall) string {
	if call.ContractName != "" {
		return call.ContractName + "." + call.MethodName
	} else {
		return call.MethodName
	}
}

func ExtractMethodCallsEx(script []byte, protocol uint, debugLogging bool, calls map[string]uint, clients []rpc.PhantasmaRPC) uint {
	PrepareMethodsRegistry(script, protocol, debugLogging, clients)

	disasm, offset := ExtractMethodCalls(script, methodTable, debugLogging)
	// fmt.Println("len disasm", len(disasm))

	for _, call := range disasm {
		n := getCallFullName(call)
		calls[n] = calls[n] + 1
	}

	return offset
}
