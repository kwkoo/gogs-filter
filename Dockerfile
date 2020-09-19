FROM golang:1.15.2 as builder

ARG PREFIX=github.com/kwkoo
ARG PACKAGE=gogsfilter
LABEL builder=true
COPY src /go/src/
RUN \
  set -x \
  && \
  cd /go/src/ \
  && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /go/bin/${PACKAGE} .

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

LABEL maintainer="kin.wai.koo@gmail.com"
LABEL builder=false
LABEL org.opencontainers.image.source="https://github.com/kwkoo/gogs-filter"
COPY --from=builder /go/bin/${PACKAGE} /usr/bin/${PACKAGE}

RUN chmod 755 /usr/bin/${PACKAGE}

USER 1001

ENTRYPOINT ["/usr/bin/gogsfilter"]
