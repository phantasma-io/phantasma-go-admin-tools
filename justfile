[group('doc')]
guide:
    cat README.md | less

[group('build')]
build:
    go build ./...

build-verbose:
    go build -v ./...

clean:
    go clean -cache
    go clean -testcache

[group('test')]
test:
    go test ./...

test-clean:
    go clean -testcache
    go test ./...

[group('run')]
run-account-analyzer:
    go run ./cmd/account-analyzer -i --order=asc --show-fungible
