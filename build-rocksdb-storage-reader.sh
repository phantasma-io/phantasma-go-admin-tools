CGO_CFLAGS="-I/usr/local/include" \
CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd -lbz2" \
  go build -o bin ./cmd/rocksdb-storage-reader