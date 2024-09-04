STORAGE=~/github/phantasma/phantasma-storage-data-legacy
#STORAGE=~/github/phantasma/phantasma-storage-data/state.db

cd ./bin
#./rocksdb-storage-reader -p $STORAGE
#./rocksdb-storage-reader -p $STORAGE --list-contents -f=chain.main --limit=10000
#./rocksdb-storage-reader -p $STORAGE --column-family=Platform
#./rocksdb-storage-reader -p $STORAGE -f=chain.main --list-keys-with-unknown-base-keys
#./rocksdb-storage-reader -p $STORAGE -f=chain.main --list-keys-with-unknown-sub-keys --base-key=.governance._constraintMap
#./rocksdb-storage-reader -p $STORAGE -l -f=chain.main --base-key=.balances
#./rocksdb-storage-reader -p $STORAGE -l -f=chain.main --base-key=.account._addressMap
#./rocksdb-storage-reader -p $STORAGE -f=chain.main -i
# >1 2>&1

#./rocksdb-storage-reader -p $STORAGE -l -f=chain.main --base-key=GHOST.serie9997 # --list-keys-with-unknown-sub-keys
#./rocksdb-storage-reader -p $STORAGE -l -f=chain.main --base-key=.interop._swapMap --list-unique-sub-keys -v --parse-subkey-as-hash
./rocksdb-storage-reader -p $STORAGE -l -f=chain.main --base-key=.tokens.list
