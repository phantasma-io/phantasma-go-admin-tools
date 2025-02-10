STORAGE=~/github/phantasma/phantasma-storage-data

OUT=../output/
json_format () {
  jq . $1 > $1_f
  mv $1_f $1
}

sh build-rocksdb-storage-reader.sh

mkdir -p output

cd ./bin

./rocksdb-storage-reader -p $STORAGE --merge-kcal-leftovers --fungible-balances-json=$OUT/fungible_balances.json --kcal-leftovers-json=$OUT/staking_leftovers.json --output-format=json --output=$OUT/fungible_balances_with_leftovers.json >$OUT/leftovers-merge.log
json_format $OUT/fungible_balances_with_leftovers.json
