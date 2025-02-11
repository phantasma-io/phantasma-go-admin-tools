STORAGE=~/github/phantasma/phantasma-storage-data

OUT=../output/
json_format () {
  jq . $1 > $1_f
  mv $1_f $1
}

sh build-rocksdb-storage-reader.sh

mkdir -p output

cd ./bin

./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-block-hashes --output-format=json --output=$OUT/block_heights.json
json_format $OUT/block_heights.json

./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-txes --block-heigts-json=$OUT/block_heights.json --output-format=json --output=$OUT/txes.json >$OUT/txes.log
json_format $OUT/txes.json

./rocksdb-storage-reader -p $STORAGE -d -f=chain.main --dump-blocks --block-heigts-json=$OUT/block_heights.json --output-format=json --output=$OUT/blocks.json
json_format $OUT/blocks.json
