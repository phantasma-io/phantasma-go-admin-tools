ADDRESS=
NEXUS=--nexus=mainnet
#NEXUS=--nexus=testnet
#SHOW_FAILED=--show-failed
MODE=--track-account-state\ --use-initial-state
#MODE=--get-initial-state
SYMBOL=--symbol=SOUL
EVENTS=--event-kind=TokenClaim\ --event-kind=TokenStake
#EVENTS=--event-kind=TokenStake

just build-account-analyzer
cd ./bin
./account-analyzer $MODE $SHOW_FAILED $NEXUS $SYMBOL $EVENTS --order=asc --show-fungible --show-nonfungible --address=$ADDRESS
