GOFILES=$(shell find . -name *.go)
REGISTRY=alandiegosantos

TARGET=webserver

all: ${TARGET}

webserver: ${GOFILES}
	go mod vendor
	CGO_ENABLED=0 GOFLAGS="-mod=vendor" go build -o $@ -v

.PHONY: clean
clean:
	@rm -f ${TARGET}


.PHONY: docker
docker: ${GOFILES}
	docker build . -t ${REGISTRY}/${TARGET}:latest