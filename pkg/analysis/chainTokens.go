package analysis

import (
	"fmt"

	"github.com/phantasma-io/phantasma-go/pkg/rpc"
	"github.com/phantasma-io/phantasma-go/pkg/rpc/response"
)

var chainTokens []response.TokenResult

func InitChainTokens(client rpc.PhantasmaRPC) int {
	chainTokens, _ = client.GetTokens(false)
	return len(chainTokens)
}

func PrintTokens() {
	for _, t := range chainTokens {
		fmt.Println(t.Symbol, "flags:", t.Flags)
	}
}

func GetChainToken(symbol string) response.TokenResult {
	for _, t := range chainTokens {
		if t.Symbol == symbol {
			return t
		}
	}

	panic("Token not found")
}
