HOSTNAME=registry.terraform.io
NAMESPACE=dnscaleou
NAME=dnscale
BINARY=terraform-provider-${NAME}
VERSION=1.0.0
OS_ARCH=darwin_arm64

default: build

build:
	go build -o ${BINARY}

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

test:
	go test ./... -v

testacc:
	TF_ACC=1 go test ./... -v $(TESTARGS) -timeout 120m

fmt:
	gofmt -s -w .

lint:
	golangci-lint run ./...

docs:
	go generate ./...

clean:
	rm -f ${BINARY}

.PHONY: build install test testacc fmt lint docs clean
