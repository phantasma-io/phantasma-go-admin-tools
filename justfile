STORAGE := "~/github/phantasma/phantasma-storage-data-legacy"


[group('doc')]
guide:
    cat README.md | less

[group('build')]
build-account-analyzer:
    mkdir -p bin
    go build -o bin ./cmd/account-analyzer

build-rocksdb-storage-reader:
    mkdir -p bin
    CGO_CFLAGS="-I/usr/local/include" \
    CGO_LDFLAGS="-L/usr/local/lib -lrocksdb -lstdc++ -lm -lz -lsnappy -llz4 -lzstd -lbz2" \
    go build -o bin ./cmd/rocksdb-storage-reader

build: build-account-analyzer build-rocksdb-storage-reader

clean:
    go clean -cache
    go clean -testcache
    rm -f bin/account-analyzer
    rm -f bin/rocksdb-storage-reader

[group('test')]
test:
    go test ./...

test-clean:
    go clean -testcache
    go test ./...

[group('run')]
run-account-analyzer: build-account-analyzer
    go run ./cmd/account-analyzer -i --order=asc --show-fungible

run-rocksdb-storage-reader: build-rocksdb-storage-reader
    go run ./cmd/rocksdb-storage-reader -p {{ STORAGE }} --dump-addresses --output-format=json -f=chain.main