.PHONY: build
build:
	CGO_ENABLED=1 go build -o s3-signer main.go

.PHONY: setup
setup:
	dep ensure