.PHONY: build clean

build:
	CGO_ENABLED=0 gox -mod vendor -ldflags '-extldflags "-static" -X "main.Version=0.0.0"' -os 'linux' -arch 'amd64' -arch '386' .

clean:
	rm smtpd_exporter_*
