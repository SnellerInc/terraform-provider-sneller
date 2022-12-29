TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=sneller.io
NAMESPACE=edu
NAME=sneller
BINARY=terraform-provider-${NAME}
VERSION=0.1
OS_ARCH=linux_amd64

default: install

build:
	go build -o ${BINARY}

release:
	goreleaser release --rm-dist --snapshot --skip-publish  --skip-sign

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

debug: install
	ln -fs "`pwd`/__debug_bin" ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/${BINARY}
	cd test && rm -rf .terraform .terraform.lock.hcl
	cd test && terraform init
	cd test && TF_REATTACH_PROVIDERS='{"sneller.io/edu/sneller":{"Protocol":"grpc","ProtocolVersion":5,"Pid":10504,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin515156416"}}}' terraform apply

test: 
	go test -i -count=1 -parallel=4 ./...

testacc:
	TF_ACC=1 go test -count=1 -parallel=1 -timeout 10m -v ./...; \
