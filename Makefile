GOFLAGS = -mod=vendor
export GOFLAGS

STATIK = redfish/statik/statik.go
STATIK_SRC = $(shell find redfish/rules)

all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    statik  - generate statik codes."
	@echo "    test    - run signle host tests."
	@echo "    setup   - install dependencies."

statik: $(STATIK)

$(STATIK): $(STATIK_SRC)
	mkdir -p $(dir $(STATIK))
	go generate ./pkg/...  # this confuses parallel make

test: $(STATIK)
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | tee /dev/stderr)"
	go build ./...
	go test -race -v ./...
	go vet ./...

setup:
	GO111MODULE=off go get -u github.com/rakyll/statik

.PHONY: all statik test setup
