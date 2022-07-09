NAME := kube-gen
PKG := github.com/kylemcc/kube-gen/cmd/kube-gen
REGISTRIES := kylemcc ghcr.io/kylemcc/kube-gen
CGO_ENABLED := 0

include build.mk
