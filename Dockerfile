FROM golang:alpine3.11 as builder

RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/statsd-protobuf
COPY . .
RUN go get -d -v

RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-w -s" -o /statsd_exporter

FROM scratch

COPY --from=builder /statsd_exporter /statsd_exporter

# ARG ARCH="amd64"
# ARG OS="linux"
# FROM quay.io/prometheus/busybox-${OS}-${ARCH}:latest
# LABEL maintainer="The Prometheus Authors <prometheus-developers@googlegroups.com>"

# ARG ARCH="amd64"
# ARG OS="linux"
# COPY .build/${OS}-${ARCH}/statsd_exporter /bin/statsd_exporter

# USER        nobody
EXPOSE      9102 9125 9125/udp
# HEALTHCHECK CMD wget --spider -S "http://localhost:9102/metrics" -T 60 2>&1 || exit 1
ENTRYPOINT  [ "/statsd_exporter" ]
