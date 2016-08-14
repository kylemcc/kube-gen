PACKAGES:=$(shell go list ./...)
BUILD_TIME:=`date -u '+%Y-%m-%dT%H:%M:%S'`
REVISION:=`git rev-parse HEAD`
LDFLAGS=-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.revision=$(REVISION)
VERSION?=$(REVISION)

all: kube-gen

clean:
	rm -rf dist/

kube-gen: need-fmt vet test
	go install -ldflags "$(LDFLAGS)" ./cmd/kube-gen

dist: kube-gen
	echo "Building version $(VERSION) / revision $(REVISION)"
	gox -os='!windows !plan9' -output="dist/{{.OS}}-{{.Arch}}/kube-gen" -ldflags "$(LDFLAGS)" ./...

vet:
	go vet ./...

test:
	go test -cover $(PACKAGES)

need-fmt:
	if [ -n "$(shell gofmt -l .)" ]; then \
		echo "Please run go fmt on the following files:"; \
		gofmt -l .; \
		exit 1; \
	fi

.PHONY: kube-gen clean vet test
.SILENT:
