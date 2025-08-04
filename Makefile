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

check-generate:
	$(MAKE) generate
	go mod tidy
	git diff --exit-code --name-only

setup:
	env GOFLAGS= go install golang.org/x/tools/cmd/goimports@latest
	env GOFLAGS= go install honnef.co/go/tools/cmd/staticcheck@latest

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
	## Must change the URL to the latest version of iDRAC Tools.
	## Please see https://www.dell.com/support/home/ja-jp/drivers/driversdetails?driverid=mfv7t&msockid=2f4827c6031868db216b3232026069ad
	curl 'https://dl.dell.com/FOLDER12638439M/1/Dell-iDRACTools-Web-LX-11.3.0.0-795_A00.tar.gz?uid=a19a6035-6a13-48b9-1fd5-4587c4944a96&fn=Dell-iDRACTools-Web-LX-11.3.0.0-795_A00.tar.gz' \
		-H 'user-agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36' \
		--output idrac-tools.tar.gz
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
