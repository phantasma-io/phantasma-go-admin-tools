cd ./bin
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data --list-contents -f=chain.main #--limit=10
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data --column-family=Platform
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -f=chain.main --list-keys-with-unknown-base-keys
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -f=chain.main --list-keys-with-unknown-sub-keys --base-key=.governance._constraintMap
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -l -f=chain.main --base-key=.balances
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -l -f=chain.main --base-key=.account._addressMap
#./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -f=chain.main -i
# >1 2>&1

./rocksdb-storage-reader -p ~/github/phantasma/phantasma-storage-data -l -f=chain.main --base-key=GHOST.serie9997 # --list-keys-with-unknown-sub-keys
