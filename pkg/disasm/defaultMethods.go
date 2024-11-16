package disasm

import (
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

func GetDefaultMethods(protocol uint) *orderedmap.OrderedMap[string, int] {
	var table = orderedmap.New[string, int]()
	// TODO fix ordering
	table.Set("Address()", 1)

	table.Set("account.Migrate", 2)
	table.Set("account.RegisterName", 2)
	table.Set("account.RegisterScript", 2)
	table.Set("account.UnregisterName", 1)

	table.Set("consensus.InitPoll", 8)
	table.Set("consensus.SingleVote", 3)
	table.Set("consensus.HasConsensus", 2)
	table.Set("consensus.GetRank", 2)

	table.Set("Data.Get", 3)
	table.Set("Data.Set", 2)

	table.Set("gas.AllowGas", 4)
	table.Set("gas.ApplyInflation", 1)
	table.Set("gas.FixInflationTiming", 2)
	table.Set("gas.SpendGas", 1)

	table.Set("governance.CreateValue", 4)
	table.Set("governance.SetValue", 3)

	table.Set("Map.Clear", 1)

	table.Set("market.SellToken", 6)
	table.Set("market.BuyToken", 3)
	table.Set("market.CancelSale", 2)
	table.Set("market.EditAuction", 9)
	table.Set("market.ListToken", 12)
	table.Set("market.BidToken", 6)

	table.Set("Nexus.BeginInit", 1)
	table.Set("Nexus.EndInit", 1)
	table.Set("Nexus.CreateOrganization", 4)
	table.Set("Nexus.CreateToken", 7)

	table.Set("Runtime.DeployContract", 4)
	table.Set("Runtime.Log", 1)
	table.Set("Runtime.Notify", 3)
	table.Set("Runtime.IsWitness", 1)
	table.Set("Runtime.IsTrigger", 0)
	table.Set("Runtime.TransferBalance", 3)
	table.Set("Runtime.MintTokens", 4)
	table.Set("Runtime.BurnTokens", 3)
	table.Set("Runtime.SwapTokens", 5)
	table.Set("Runtime.TransferTokens", 4)
	table.Set("Runtime.TransferToken", 4)
	table.Set("Runtime.MintToken", 4)
	table.Set("Runtime.BurnToken", 3)
	table.Set("Runtime.InfuseToken", 5)
	table.Set("Runtime.WriteToken", 4)
	table.Set("Runtime.Version", 0)

	if protocol < 14 {
		table.Set("Runtime.UpgradeContract", 3)
	} else {
		table.Set("Runtime.UpgradeContract", 4)
	}

	table.Set("Organization.AddMember", 3)
	table.Set("Organization.RemoveMember", 3)

	table.Set("stake.Migrate", 2)
	table.Set("stake.MasterClaim", 1)
	table.Set("stake.Stake", 2)
	table.Set("stake.Unstake", 2)
	table.Set("stake.Claim", 2)
	table.Set("stake.AddProxy", 3)
	table.Set("stake.RemoveProxy", 2)

	table.Set("storage.CreateFile", 5)
	table.Set("storage.UploadData", 6)
	table.Set("storage.UploadFile", 7)
	table.Set("storage.DeleteFile", 2)
	table.Set("storage.SetForeignSpace", 2)

	table.Set("swap.GetRate", 3)
	table.Set("swap.DepositTokens", 3)
	table.Set("swap.SwapFee", 3)
	table.Set("swap.SwapReverse", 4)
	table.Set("swap.SwapFiat", 4)
	table.Set("swap.SwapTokens", 4)
	table.Set("swap.SwapTokensV2", 4)

	table.Set("validator.SetValidator", 3)

	return table
}
