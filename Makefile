install:
	go mod vendor

lint:
	gofmt -w -s .

test:
	go test -v ./...

build:
	go build -o bin/main main.go

docker-build:
	docker build -t docker_stats_exporter .

docker-run:
	docker run --name docker_stats_exporter -v /var/run/docker.sock:/var/run/docker.sock -v /usr/bin/docker:/usr/bin/docker docker_stats_exporter

start:
	go run main.go	