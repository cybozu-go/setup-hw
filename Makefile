BIN_PKGS = ./pkg/idrac-passwd-hash ./pkg/setup-hw ./pkg/monitor-hw
GENERATED = redfish/rendered_rules.go
GENERATE_SRC = $(shell find redfish/rules)

GOFLAGS = -mod=vendor
export GOFLAGS

all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    generate  - generate codes."
	@echo "    test      - run signle host tests."
	@echo "    install   - install binaries."

generate: $(GENERATED)

$(GENERATED): $(GENERATE_SRC) pkg/render-rules/main.go
	go generate ./redfish/...

test: generate
	test -z "$$(gofmt -s -l . | grep -v '^vendor' | tee /dev/stderr)"
	test -z "$$(golint $$(go list ./... | grep -v /vendor/) | tee /dev/stderr)"
	go build ./...
	go test -race -v ./...
	go vet ./...

install: generate
ifdef GOBIN
	mkdir -p $(GOBIN)
endif
	GOBIN=$(GOBIN) go install $(BIN_PKGS)

.PHONY: all generate test install
