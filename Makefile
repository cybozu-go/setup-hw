BIN_PKGS = ./pkg/idrac-passwd-hash ./pkg/setup-hw ./pkg/monitor-hw ./pkg/collector
GENERATED = redfish/rendered_rules.go
GENERATE_SRC = $(shell find redfish/rules)

all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    generate  - generate codes."
	@echo "    test      - run signle host tests."
	@echo "    install   - install binaries."

generate: $(GENERATED)

$(GENERATED): $(GENERATE_SRC) pkg/render-rules/main.go
	go generate ./redfish/...

test: generate staticcheck
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	go build ./...
	go test -race -v ./...
	go vet ./...

staticcheck:
	if ! which staticcheck >/dev/null; then \
		cd /tmp; env GOFLAGS= GO111MODULE=on go get honnef.co/go/tools/cmd/staticcheck; \
	fi

install: generate
ifdef GOBIN
	mkdir -p $(GOBIN)
endif
	GOBIN=$(GOBIN) go install $(BIN_PKGS)

build-image: install
ifdef GOBIN
	mkdir -p $(GOBIN)
	cp $(GOBIN)/setup-hw $(GOBIN)/monitor-hw $(GOBIN)/collector ./docker/
else
	cp $(GOPATH)/bin/setup-hw $(GOPATH)/bin/monitor-hw $(GOPATH)/bin/collector ./docker/
endif
	cd docker && docker build -t quay.io/cybozu/setup-hw:dev .

.PHONY: all generate test install build-image staticcheck
