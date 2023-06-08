TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=registry.terraform.io
NAMESPACE=snellerinc
NAME=sneller
BINARY=terraform-provider-${NAME}
VERSION?=$(shell git describe --tags --exact-match | sed 's/^v//')
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test: 
	go test -i -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=1 -timeout 10m -v ./...; \
