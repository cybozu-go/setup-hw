# binaries to be included in the image
BINS_IMAGE = setup-hw monitor-hw collector setup-apply-firmware setup-isoreboot

# binaries not to be included in the image
BINS_NOIMAGE = idrac-passwd-hash

BINS = $(BINS_IMAGE) $(BINS_NOIMAGE)
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

setup:
	env GOFLAGS= go install golang.org/x/tools/cmd/goimports@latest
	env GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@latest

test: setup generate
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	go build ./...
	go test -race -v ./...
	go vet ./...

install: generate
ifdef GOBIN
	mkdir -p $(GOBIN)
endif
	GOBIN=$(GOBIN) go install $(foreach f, $(BINS), ./pkg/$(f))

build-image: install
ifdef GOBIN
	mkdir -p $(GOBIN)
	cp $(foreach f, $(BINS_IMAGE), $(GOBIN)/$(f)) ./docker/
else
	cp $(foreach f, $(BINS_IMAGE), $(GOPATH)/bin/$(f)) ./docker/
endif
	cd docker && docker build -t quay.io/cybozu/setup-hw:dev .

.PHONY: all generate test install build-image setup
