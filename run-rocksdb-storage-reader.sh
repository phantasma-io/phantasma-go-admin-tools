STORAGE=~/github/phantasma/phantasma-storage-data-legacy
#STORAGE=~/github/phantasma/phantasma-storage-data/state.db

cd ./bin
#./rocksdb-storage-reader -p $STORAGE
#./rocksdb-storage-reader -p $STORAGE --dump -f=chain.main --limit=10000
#./rocksdb-storage-reader -p $STORAGE --column-family=Platform
#./rocksdb-storage-reader -p $STORAGE -f=chain.main --list-keys-with-unknown-base-keys
#./rocksdb-storage-reader -p $STORAGE -f=chain.main --list-keys-with-unknown-sub-keys --base-key=.governance._constraintMap
#./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --base-key=.balances
#./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --base-key=.account._addressMap
#./rocksdb-storage-reader -p $STORAGE -f=chain.main -i
# >1 2>&1

#./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --base-key=GHOST.serie9997 # --list-keys-with-unknown-sub-keys
#./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --base-key=.interop._swapMap --list-unique-sub-keys -v --parse-subkey-as-hash
#./rocksdb-storage-reader -p $STORAGE -d --output-format=csv -f=chain.main --base-key=.tokens.list
./rocksdb-storage-reader -p $STORAGE -d --output-format=json -f=chain.main --base-key=.balances --subkeys=SOUL,KCAL,USD,CROWN,NEO,GAS,ETH,MKNI,TTRS,GOATI,GHOST,DANK,GAME,USDT,SEM,USDC,DAI,DYT,LEET,BNB,HOD,SPECC,BRC,SPE,WNDR,NKTR,GM,GFNFT,SNFT,SMNFT,WAGS #--subkeys2=P2K5fzTcq6RdLoc5ehYix4MiHPUX2xBo7micbwFCtJ8Q5si
