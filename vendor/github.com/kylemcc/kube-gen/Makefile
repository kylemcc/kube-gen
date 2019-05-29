PACKAGES:=$(shell go list ./... | grep -v vendor)
BUILD_TIME:=`date -u '+%Y-%m-%dT%H:%M:%S'`
REVISION:=`git rev-parse HEAD`
LDFLAGS=-X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME) -X main.revision=$(REVISION)
VERSION?=$(REVISION)

all: kube-gen

clean:
	rm -rf dist/
	rm -rf build/

kube-gen: need-fmt vet test
	go install -ldflags "$(LDFLAGS)" ./cmd/kube-gen

build: clean kube-gen
	echo "Building version $(VERSION) / revision $(REVISION)"
	gox -os='!windows !plan9' -output="build/{{.OS}}-{{.Arch}}/kube-gen" -ldflags "$(LDFLAGS)" $(PACKAGES)

dist: build
	mkdir dist; \
	for build in $(shell ls build/); do \
		tar -czvf dist/kube-gen-$${build}-$(VERSION).tar.gz -C build/$${build} kube-gen; \
	done; \
	cd dist && shasum -a 256 * > sha256sums.txt

vet:
	go vet *.go
	go vet cmd/kube-gen/*.go

test:
	go test -cover $(PACKAGES)

need-fmt:
	if [ -n "$(shell gofmt -l *.go cmd/kube-gen/*.go)" ]; then \
		echo "Please run go fmt on the following files:"; \
		gofmt -l .; \
		exit 1; \
	fi

.PHONY: kube-gen clean vet test
.SILENT:
