.PHONY: build clean test lint dep-update

build:
	GOFLAGS=-mod=vendor CGO_ENABLED=0 gox -mod vendor -ldflags '-extldflags "-static" -X "main.Version=0.1.0"' .

clean:
	rm smtpd_exporter_*

test:
	GOFLAGS=-mod=vendor go test ./...

lint:
	golangci-lint run --enable-all --disable=gochecknoglobals,testpackage

dep-update:
	go get -u ./...
	go test ./...
	go mod tidy
	go mod vendor
