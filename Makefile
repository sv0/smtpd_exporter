.PHONY: build clean test lint dep-update

build:
	export GOFLAGS=-mod=vendor
	CGO_ENABLED=0 gox -mod vendor -ldflags '-extldflags "-static" -X "main.Version=0.1.0"' .

clean:
	rm smtpd_exporter_*

test:
	export GOFLAGS=-mod=vendor
	go test ./...

lint:
	golangci-lint run --enable-all --disable=gochecknoglobals

dep-update:
	go get -u ./...
	go test ./...
	go mod tidy
	go mod vendor
