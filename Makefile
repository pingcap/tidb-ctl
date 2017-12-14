GOPATH ?= $(shell go env GOPATH)

# Ensure GOPATH is set before running build process.
ifeq "$(GOPATH)" ""
  $(error Please set the environment variable GOPATH before running `make`)
endif

PACKAGES  := $$(go list ./...)
FILES     := $$(find . -name "*.go" | grep -vE "vendor")
TOPDIRS   := $$(ls -d */ | grep -vE "vendor")

.PHONY: default test check doc

default:
	go build

check:
	@echo "gofmt (simplify)"
	@ gofmt -s -l -w $(FILES) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "vet"
	@ go tool vet -all -shadow *.go 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'
	@ go tool vet -all -shadow $(TOPDIRS) 2>&1 | awk '{print} END{if(NR>0) {exit 1}}'

	@echo "golint"
	go get github.com/golang/lint/golint
	@ golint -set_exit_status $(PACKAGES)

	@echo "errcheck"
	go get github.com/kisielk/errcheck
	@ errcheck -blank $(PACKAGES) | grep -v "_test\.go" | awk '{print} END{if(NR>0) {exit 1}}'

test: check
	@ log_level=debug go test -p 3 -cover $(PACKAGES)

doc:
	@mkdir -p doc
	@ go run main.go -H127.0.0.1 -P1234 --doc