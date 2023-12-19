# About
This is a prometheus exporter for `docker stats`. It serves the information from `docker stats` at the endpoint `/metrics`. 

## Install
Once [go](https://go.dev/doc/install) is installed, you can install the dependencies with `make install`. You can then `make start` to run a server which serves the `/metrics` endpoint.