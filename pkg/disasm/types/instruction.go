package types

import "github.com/phantasma-io/phantasma-go/pkg/vm"

type Instruction struct {
	Offset uint
	Opcode vm.Opcode
	Args   []any
}
