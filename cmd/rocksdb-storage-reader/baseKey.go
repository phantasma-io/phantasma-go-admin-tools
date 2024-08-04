package main

type BaseKey string

const (
	Abi                   BaseKey = "abi"
	AccountAddressMap     BaseKey = ".account._addressMap"
	AccountNameMap        BaseKey = ".account._nameMap"
	AddressTxHashMap      BaseKey = ".adblmp"
	Balances              BaseKey = ".balances"
	Blocks                BaseKey = ".blocks"
	Burned                BaseKey = ".burned"
	ChainAddress          BaseKey = ".chain.addr"
	ChainArchives         BaseKey = ".chain.archives"
	ChainName             BaseKey = ".chain.name"
	ChainOrg              BaseKey = ".chain.org"
	ChainsList            BaseKey = ".chains.list"
	Contracts             BaseKey = "contracts"
	ExchangeOtcBook       BaseKey = ".exchange._otcBook"
	GasAllowanceMap       BaseKey = ".gas._allowanceMap"
	GasAllowanceTargets   BaseKey = ".gas._allowanceTargets"
	GasInflationReady     BaseKey = ".gas._inflationReady"
	GasLastInflationDate  BaseKey = ".gas._lastInflationDate"
	GasRewardAccum        BaseKey = ".gas._rewardAccum"
	Governance            BaseKey = ".governance"
	Ids                   BaseKey = ".ids"
	InteropHistoryMap     BaseKey = ".interop._historyMap"
	InteropPlatformHashes BaseKey = ".interop._platformHashes"
	InteropSwapMap        BaseKey = ".interop._swapMap"
	InteropWithdraws      BaseKey = ".interop._withdraws"
	Height                BaseKey = ".height"
	Market                BaseKey = ".market"
	Moonjar               BaseKey = "moonjar"
	Name                  BaseKey = "name"
	Nexus                 BaseKey = ".nexus"
	Org                   BaseKey = ".org"
	Owner                 BaseKey = "owner"
	Ownership             BaseKey = ".ownership"
	Pharming              BaseKey = "pharming"
	Platforms             BaseKey = ".platforms"
	Sale                  BaseKey = ".sale"
	Script                BaseKey = "script"
	Series                BaseKey = "series"
	Slots                 BaseKey = "slots"
	Stake                 BaseKey = ".stake"
	Storage               BaseKey = ".storage"
	SwapAddress           BaseKey = ".swapaddr"
	SwapMap               BaseKey = ".swapmap"
	Token                 BaseKey = ".token"
	TokensList            BaseKey = ".tokens.list"
	Txs                   BaseKey = ".txs"
	TxBlockMap            BaseKey = ".txblmp"
	Uuid                  BaseKey = "_uid"
	Validator             BaseKey = ".validator"

	// Contracts
	ContractBnb   BaseKey = "BNB"
	ContractBrc   BaseKey = "BRC"
	ContractCrown BaseKey = "CROWN"
	ContractDank  BaseKey = "DANK"
	ContractDai   BaseKey = "DAI"
	ContractDyt   BaseKey = "DYT"
	ContractEth   BaseKey = "ETH"
	ContractGame  BaseKey = "GAME"
	ContractGas   BaseKey = "GAS"
	ContractGhost BaseKey = "GHOST"
	ContractGfnft BaseKey = "GFNFT"
	ContractGm    BaseKey = "GM"
	ContractGoati BaseKey = "GOATI"
	ContractHod   BaseKey = "HOD"
	ContractKcal  BaseKey = "KCAL"
	ContractLeet  BaseKey = "LEET"
	ContractMkni  BaseKey = "MKNI"
	ContractNeo   BaseKey = "NEO"
	ContractNktr  BaseKey = "NKTR"
	ContractSnft  BaseKey = "SNFT"
	ContractSem   BaseKey = "SEM"
	ContractSoul  BaseKey = "SOUL"
	ContractSpe   BaseKey = "SPE"
	ContractSpecc BaseKey = "SPECC"
	ContractSmnft BaseKey = "SMNFT"
	ContractTtrs  BaseKey = "TTRS"
	ContractUsdc  BaseKey = "USDC"
	ContractUsdt  BaseKey = "USDT"
	ContractWags  BaseKey = "WAGS"
	ContractWndr  BaseKey = "WNDR"
)

var knownBaseKeys []BaseKey = []BaseKey{
	Abi,
	AccountAddressMap,
	AccountNameMap,
	AddressTxHashMap,
	Balances,
	Blocks,
	Burned,
	ChainAddress,
	ChainArchives,
	ChainName,
	ChainOrg,
	ChainsList,
	Contracts,
	ExchangeOtcBook,
	GasAllowanceMap,
	GasAllowanceTargets,
	GasInflationReady,
	GasLastInflationDate,
	GasRewardAccum,
	Governance,
	Ids,
	InteropHistoryMap,
	InteropPlatformHashes,
	InteropSwapMap,
	InteropWithdraws,
	Height,
	Market,
	Moonjar,
	Name,
	Nexus,
	Org,
	Owner,
	Ownership,
	Pharming,
	Platforms,
	Sale,
	Script,
	Series,
	Slots,
	Stake,
	Storage,
	SwapAddress,
	SwapMap,
	Token,
	TokensList,
	Txs,
	TxBlockMap,
	Uuid,
	Validator,

	// Contracts
	ContractBnb,
	ContractBrc,
	ContractCrown,
	ContractDank,
	ContractDai,
	ContractDyt,
	ContractEth,
	ContractGame,
	ContractGas,
	ContractGhost,
	ContractGfnft,
	ContractGm,
	ContractGoati,
	ContractHod,
	ContractKcal,
	ContractLeet,
	ContractMkni,
	ContractNeo,
	ContractNktr,
	ContractSnft,
	ContractSem,
	ContractSoul,
	ContractSpe,
	ContractSpecc,
	ContractSmnft,
	ContractTtrs,
	ContractUsdc,
	ContractUsdt,
	ContractWags,
	ContractWndr}

func (b BaseKey) Bytes() []byte {
	return []byte(b)
}

func (b BaseKey) String() string {
	return string(b)
}

func GetBytesForKnownBaseKeys() map[BaseKey][]byte {
	var knownBaseKeysBytes map[BaseKey][]byte = make(map[BaseKey][]byte, len(knownBaseKeys))
	for _, p := range knownBaseKeys {
		knownBaseKeysBytes[p] = p.Bytes()
	}
	return knownBaseKeysBytes
}

var KnowSubKeys map[BaseKey][]string = map[BaseKey][]string{
	Balances: {".BNB",
		".DANK",
		".DYT",
		".ETH",
		".GAS",
		".GM",
		".GOATI",
		".HOD",
		".KCAL",
		".MKNI",
		".NEO",
		".NKTR",
		".SPE",
		".SOUL",
		".USDT",
		".WAGS",
		".WNDR"},
	Burned: {".BRC",
		".CROWN",
		".DANK",
		".DYT",
		".GAME",
		".GFNFT",
		".GHOST",
		".GOATI",
		".KCAL",
		".SEM",
		".SMNFT",
		".SNFT",
		".SPECC",
		".TTRS",
		".WAGS"},
	Ids: {".BRC",
		".CROWN",
		".GAME",
		".GHOST",
		".GFNFT",
		".LEET",
		".SEM",
		".SMNFT",
		".SNFT",
		".SPECC",
		".TTRS"}}

func GetBytesForKnownSubKeys(baseKey BaseKey, addBaseKey bool) map[string][]byte {
	if KnowSubKeys[baseKey] == nil {
		return nil
	}

	var knownSubKeysBytes map[string][]byte = make(map[string][]byte, len(KnowSubKeys[baseKey]))
	for _, p := range KnowSubKeys[baseKey] {
		if addBaseKey {
			knownSubKeysBytes[p] = []byte(string(baseKey) + p)
		} else {
			knownSubKeysBytes[p] = []byte(p)
		}
	}
	return knownSubKeysBytes
}
