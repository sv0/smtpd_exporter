.PHONY: build clean test lint

export GOFLAGS := -mod=vendor

build:
	CGO_ENABLED=0 gox -mod vendor -ldflags '-extldflags "-static" -X "main.Version=0.1.0"' .

clean:
	rm smtpd_exporter_*

test:
	go test ./...

lint:
	golangci-lint run --enable-all --disable=gochecknoglobals
