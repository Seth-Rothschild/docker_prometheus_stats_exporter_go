# About
This is a prometheus exporter for `docker stats`. It serves the information from `docker stats` at the endpoint `/metrics`. 

## Install
Once [go](https://go.dev/doc/install) is installed, you can install the dependencies with `make install`. You can then `make start` to run a server which serves the `/metrics` endpoint.

## Docker
If you'd rather manage your running applications with Docker, we've provided a simple [Dockerfile](./Dockerfile) that contains the executable. Note that in order to run `docker stats`, the container needs access to `docker` which you should only allow for containers you trust. To build the image run `make build` and then `make docker-build`. An example `docker run` command would look like the following:
```
docker run -d \
    --name docker_stats_exporter \
    --restart unless-stopped \
    -e PORT=9200 \
    -p 9200:9200 \
    -v /var/run/docker.sock:/var/run/docker.sock \
    -v /usr/bin/docker:/usr/bin/docker \
    docker_stats_exporter
```