# binaries to be included in the image
BINS_IMAGE = setup-hw monitor-hw collector setup-apply-firmware setup-isoreboot

# binaries not to be included in the image
BINS_NOIMAGE = idrac-passwd-hash

BINS = $(BINS_IMAGE) $(BINS_NOIMAGE)
GENERATED = redfish/rendered_rules.go
GENERATE_SRC = $(shell find redfish/rules)

# CLI tools are managed by aqua. See aqua.yaml.
GOLANGCI_LINT := aqua exec -- golangci-lint

.PHONY: all
all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    generate  - generate codes."
	@echo "    lint      - run golangci-lint."
	@echo "    lint-fix  - run golangci-lint and perform fixes."
	@echo "    test      - run signle host tests."
	@echo "    install   - install binaries."

.PHONY: generate
generate: $(GENERATED)

$(GENERATED): $(GENERATE_SRC) pkg/render-rules/main.go
	go generate ./redfish/...

.PHONY: check-generated
check-generated:
	$(MAKE) generate
	go mod tidy
	git diff --exit-code --name-only

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix:
	$(GOLANGCI_LINT) run --fix ./...

.PHONY: test
test:
	go build ./...
	go test -race -v ./...

.PHONY: install
install: generate
ifdef GOBIN
	mkdir -p $(GOBIN)
endif
	GOBIN=$(GOBIN) CGO_ENABLED=0 go install -ldflags="-s -w" $(foreach f, $(BINS), ./pkg/$(f))

.PHONY: download-idractools
download-idractools:
	# Must change the URL to the latest version of iDRAC Tools.
	# Please see https://www.dell.com/support/home/ja-jp/drivers/driversdetails?driverid=2fgym
	curl 'https://dl.dell.com/FOLDER13988164M/1/Dell-iDRACTools-Web-LX-11.4.0.0-1435_A00.tar.gz' \
		-H 'user-agent: setup-hw' \
		--output idrac-tools.tar.gz
	echo "b706d0ac3f09e74a32a9e6dfa883641e1edb0a8d2cbdd85908766f502417c3ca  idrac-tools.tar.gz" | sha256sum -c
	tar -xzf idrac-tools.tar.gz -C docker

.PHONY: build-image
build-image: install
ifdef GOBIN
	mkdir -p $(GOBIN)
	cp $(foreach f, $(BINS_IMAGE), $(GOBIN)/$(f)) ./docker/
else
	cp $(foreach f, $(BINS_IMAGE), $(GOPATH)/bin/$(f)) ./docker/
endif
	cd docker && docker build -t ghcr.io/cybozu-go/setup-hw:dev .
