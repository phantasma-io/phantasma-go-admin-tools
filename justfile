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

build-rocksdb-fedora:
    sudo dnf install gflags snappy snappy-devel zlib zlib-devel bzip2 bzip2-devel lz4-devel libzstd-devel
    sudo sh scripts/rocksdb-build/build.sh

build-rocksdb-debian:
    sudo apt install libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev
    sudo sh scripts/rocksdb-build/build.sh

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
    sh run-account-analyzer.sh

run-rocksdb-storage-reader: build-rocksdb-storage-reader
    go run ./cmd/rocksdb-storage-reader -p {{ STORAGE }} --dump-addresses --output-format=json -f=chain.main