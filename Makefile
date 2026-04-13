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

check-generated:
	$(MAKE) generate
	go mod tidy
	git diff --exit-code --name-only

setup:
	# goimports is called by pkg/render-rules
	env GOFLAGS= go install golang.org/x/tools/cmd/goimports@24a8e95f9d7ae2696f66314da5e50c0d98ccaa90 # v0.43.0
	env GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@ff63afafc529279f454e02f1d060210bd4263951 # v0.7.0

test:
	test -z "$$(gofmt -s -l . | tee /dev/stderr)"
	staticcheck ./...
	go build ./...
	go test -race -v ./...
	go vet ./...

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

build-image: install
ifdef GOBIN
	mkdir -p $(GOBIN)
	cp $(foreach f, $(BINS_IMAGE), $(GOBIN)/$(f)) ./docker/
else
	cp $(foreach f, $(BINS_IMAGE), $(GOPATH)/bin/$(f)) ./docker/
endif
	cd docker && docker build -t ghcr.io/cybozu-go/setup-hw:dev .

.PHONY: all generate check-generate setup test install download-idractools build-image
