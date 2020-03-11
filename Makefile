GOPATH ?= $(shell go env GOPATH)

# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif

GO       := GO111MODULE=on go
GOBUILD  := CGO_ENABLED=0 $(GO) build

PACKAGES  := $$(go list ./...)
FILES     := $$(find . -name "*.go")

.PHONY: default test check doc

default:
	$(GOBUILD)

check:
	@echo "gofmt (simplify)"
	@ gofmt -s -l -w $(FILES) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "vet"
	@ $(GO) vet -all $(PACKAGES) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "golint"
	@ $(GO) build golang.org/x/lint/golint
	@ ./golint -set_exit_status $(PACKAGES)

	@echo "errcheck"
	@ $(GO) build github.com/kisielk/errcheck
	@ ./errcheck -blank $(PACKAGES) | grep -v "_test\.go" | awk '{print} END{if(NR>0) {exit 1}}'

test: check
	@ log_level=fatal $(GO) test $(PACKAGES)

doc:
	@mkdir -p doc
	@ $(GO) run main.go --doc
