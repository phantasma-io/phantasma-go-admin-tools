package disasm

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

var methodTable *orderedmap.OrderedMap[string, int] = nil
var methodTableProtocol uint = 0

var knownContracts []string

func registerMissingContractMethods(client rpc.PhantasmaRPC, contractName string) {
	contract, err := client.GetContract(contractName, "main")
	if err != nil {
		panic(err)
	}
	for _, m := range contract.Methods {
		methodName := contractName + "." + m.Name
		argNumber := len(m.Parameters)
		fmt.Println("Registering contract method " + methodName + " with " + strconv.Itoa(argNumber) + " parameters")
		methodTable.Set(methodName, argNumber)
	}
}

func PrepareMethodsRegistry(script []byte, protocol uint, debugLogging bool, clients []rpc.PhantasmaRPC) {
	if methodTable == nil || methodTableProtocol != protocol {
		methodTableProtocol = protocol
		methodTable = GetDefaultMethods(methodTableProtocol)

		// Collecting known contract names
		knownContracts = []string{}
		for pair := methodTable.Oldest(); pair != nil; pair = pair.Next() {
			s := strings.Split(pair.Key, ".")
			if len(s) > 1 {
				knownContracts = append(knownContracts, s[0])
			} else {
				knownContracts = append(knownContracts, pair.Key)
			}
		}
	}

	// fmt.Println("knownContracts: " + strconv.Itoa(len(knownContracts)) + " / protocol: " + strconv.Itoa(int(protocol)))

	var contractNames = ExtractContractNames(script, debugLogging)
	// fmt.Println("contractNames: " + strconv.Itoa(len(contractNames)))
	for _, contractName := range contractNames {
		if !slices.Contains(knownContracts, contractName) {
			// fmt.Println("missing contract: " + contractName)
			registerMissingContractMethods(clients[0], contractName)
			knownContracts = append(knownContracts, contractName)
		}
	}
}
