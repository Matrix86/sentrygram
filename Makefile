TARGET=sentrygram
LDFLAGS="-s -w"

all: deps build build_client

deps:
	dep ensure

build:
	go build -o ${TARGET} -v -ldflags=${LDFLAGS} cmd/sentrygram/main.go

build_client:
	go build -o ${TARGET}_client -v -ldflags=${LDFLAGS} cmd/sentrygram_client/main.go
clean:
	rm ${TARGET}
	rm ${TARGET}_client

install: build
	mv ${TARGET} $(GOPATH)/bin/
	mv ${TARGET}_client $(GOPATH)/bin/
