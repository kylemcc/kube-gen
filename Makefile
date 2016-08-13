PACKAGES:=$(shell go list ./...)

all: kube-gen

clean:
	rm -rf dist/

kube-gen: need-fmt vet test
	go install ./cmd/kube-gen

dist: kube-gen
	gox -os='!windows,!plan9' -output="dist2/{{.OS}}_{{.Arch}}/kube-gen" ./...

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
