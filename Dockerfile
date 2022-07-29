FROM debian:bullseye
COPY gpsd-exporter /usr/bin/gpsd-exporter
ENTRYPOINT ["/usr/bin/gpsd-exporter"]