install:
	go mod vendor

lint:
	gofmt -w -s .

test:
	go test -v ./...

build:
	go build -o bin/main main.go

start:
	go run main.go	