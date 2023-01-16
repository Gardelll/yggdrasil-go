FROM alpine:latest

LABEL maintainer="Gardel <sunxinao@hotmail.com>"
LABEL "Description"="Go Yggdrasil Server"

ARG TARGETOS
ARG TARGETARCH
RUN mkdir -p /app
COPY "build/yggdrasil-${TARGETOS}-${TARGETARCH}" /app/yggdrasil

EXPOSE 8080
VOLUME /app/data

WORKDIR /app/data
ENTRYPOINT ["/app/yggdrasil"]
