GOPATH ?= $(shell go env GOPATH)

# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif

GO       := GO111MODULE=on go
GOBUILD  := CGO_ENABLED=0 $(GO) build

PACKAGES  := $$(go list ./...)
FILES     := $$(find . -name "*.go" | grep -vE "vendor")
TOPDIRS   := $$(ls -d */ | grep -vE "vendor")

.PHONY: default test check doc

default:
	$(GOBUILD)

check:
	@echo "gofmt (simplify)"
	@ gofmt -s -l -w $(FILES) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "vet"
	@$(GO) vet -all *.go 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'
	@$(GO) vet -all $(TOPDIRS) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "golint"
	GO111MODULE=off go get github.com/golang/lint/golint
	@ golint -set_exit_status $(PACKAGES)

	@echo "errcheck"
	GO111MODULE=off go get github.com/kisielk/errcheck
	@ errcheck -blank $(PACKAGES) | grep -v "_test\.go" | awk '{print} END{if(NR>0) {exit 1}}'

test: check
	@ log_level=debug $(GO) test -p 3 -cover $(PACKAGES)

doc:
	@mkdir -p doc
	@ $(GO) run main.go --doc
