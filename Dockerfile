FROM ubuntu:22.04

COPY ./bin/main /usr/local/bin/docker_stats_exporter

CMD ["/usr/local/bin/docker_stats_exporter"]