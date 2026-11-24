# binaries to be included in the image
BINS_IMAGE = setup-hw monitor-hw collector setup-apply-firmware setup-isoreboot

GENERATED = redfish/rendered_rules.go
GENERATE_SRC = $(shell find redfish/rules)

# CLI tools are managed by aqua. See aqua.yaml.
GOLANGCI_LINT := aqua exec -- golangci-lint

# Dell iDRAC Tools for Linux. See docs/maintenance.md for how to update.
# - Version: 11.4.2.0
# - URL: https://www.dell.com/support/home/en-au/drivers/driversdetails?driverid=2dnrn
IDRAC_TOOLS_URL := https://dl.dell.com/FOLDER14927789M/1/Dell-iDRACTools-Web-LX-11.4.2.0-1990_A00.tar.gz
IDRAC_TOOLS_SHA256 := 17ef224d5cb4b646b87dbb4bff3f4f5af901bc7e3589f0f27142caa0b4a130b0
IDRAC_TOOLS_STAMP := docker/iDRACTools/.sha256-$(IDRAC_TOOLS_SHA256)

.PHONY: all
all:
	@echo "Specify one of these targets:"
	@echo
	@echo "    generate  - generate codes."
	@echo "    lint      - run golangci-lint."
	@echo "    lint-fix  - run golangci-lint and perform fixes."
	@echo "    test      - run signle host tests."
	@echo "    build     - build binaries for the image into docker directory."

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

.PHONY: build
build: generate
	CGO_ENABLED=0 go build -ldflags="-s -w" -o docker/ $(foreach f, $(BINS_IMAGE), ./pkg/$(f))

.PHONY: download-idractools
download-idractools: $(IDRAC_TOOLS_STAMP)

# iDRAC Tools is downloaded again only when the checksum is changed.
$(IDRAC_TOOLS_STAMP):
	curl -fsSL '$(IDRAC_TOOLS_URL)' -H 'user-agent: setup-hw' --output idrac-tools.tar.gz
	echo "$(IDRAC_TOOLS_SHA256)  idrac-tools.tar.gz" | sha256sum -c
	rm -rf docker/iDRACTools
	tar -xzf idrac-tools.tar.gz -C docker
	touch $@

.PHONY: build-image
build-image: build download-idractools
	docker build -t ghcr.io/cybozu-go/setup-hw:dev docker
