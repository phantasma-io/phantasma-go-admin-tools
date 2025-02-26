STORAGE=~/github/phantasma/phantasma-storage-data

OUT=../output/
json_format () {
  jq . $1 > $1_f
  mv $1_f $1
}

json_del_key () {
  jq 'del(..|.'$1'?)' $2 > $2_d
  mv $2_d $2
}

sh build-rocksdb-storage-reader.sh

mkdir -p output

cd ./bin

# Getting all addresses and names
./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-addresses --output-format=csv --output=$OUT/addresses.csv
# Getting all token symbols (both fungible and non-fungible)
./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-token-symbols --output-format=csv --output=$OUT/tokens_list.csv

# Getting staking data using addresses.csv
./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-staking-claims --subkeys-csv=$OUT/addresses.csv --output-format=json --output=$OUT/staking_claims.json
json_format $OUT/staking_claims.json

./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-stakes --subkeys-csv=$OUT/addresses.csv --output-format=json --output=$OUT/stakes.json
json_format $OUT/stakes.json

./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-staking-master-age --subkeys-csv=$OUT/addresses.csv --output-format=json --output=$OUT/staking_master_age.json
json_format $OUT/staking_master_age.json

./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-staking-master-claims --subkeys-csv=$OUT/addresses.csv --output-format=json --output=$OUT/staking_master_claims.json
json_format $OUT/staking_master_claims.json

./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-staking-leftovers --subkeys-csv=$OUT/addresses.csv --output-format=json --output=$OUT/staking_leftovers.json
json_format $OUT/staking_leftovers.json

# Getting non-fungible token balances using tokens_list.csv
./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-balances-nft --subkeys-csv=$OUT/tokens_list.csv --output-format=json --output=$OUT/nft_balances.json
json_format $OUT/nft_balances.json

# Getting fungible token balances using tokens_list.csv
./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-balances --subkeys-csv=$OUT/tokens_list.csv --output-format=json --output=$OUT/fungible_balances.json
json_format $OUT/fungible_balances.json

# NFTs data
./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-nfts --nft-balances-json=$OUT/nft_balances.json --output-format=json --output=$OUT/nft_datas.json
json_format $OUT/nft_datas.json

./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-series --nft-balances-json=$OUT/nft_balances.json --output-format=json --output=$OUT/nft_series.json
json_format $OUT/nft_series.json

./rocksdb-storage-reader -p $STORAGE -f=chain.main --dump-token-info --subkeys-csv=$OUT/tokens_list.csv --output-format=json --output=$OUT/token_infos.json
json_format $OUT/token_infos.json

./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-contract-infos  --output-format=json --output=$OUT/contract_infos.json
json_format $OUT/contract_infos.json

curl -X 'GET' 'https://pharpc2.phantasma.info/api/v1/GetTokens?extended=false' -H 'accept: application/json' > $OUT/token_addresses_and_supplies.json
json_format $OUT/token_addresses_and_supplies.json
json_del_key name $OUT/token_addresses_and_supplies.json
json_del_key decimals $OUT/token_addresses_and_supplies.json
json_del_key maxSupply $OUT/token_addresses_and_supplies.json
json_del_key flags $OUT/token_addresses_and_supplies.json
json_del_key script $OUT/token_addresses_and_supplies.json
json_del_key series $OUT/token_addresses_and_supplies.json
json_del_key external $OUT/token_addresses_and_supplies.json
json_del_key price $OUT/token_addresses_and_supplies.json
