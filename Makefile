IMAGE_NAME=jaxmef/datapipe
IMAGE_VERSION=v0.0.1

test:
	go test ./...

lint:
	golangci-lint run ./...

docker:
	docker build -t $(IMAGE_NAME):$(IMAGE_VERSION) .
