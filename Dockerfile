FROM debian:12-slim

LABEL maintainer="Gardel <sunxinao@hotmail.com>"
LABEL "Description"="Go Yggdrasil Server"

RUN apt-get update && apt-get install -y ca-certificates
RUN update-ca-certificates
ARG TARGETOS
ARG TARGETARCH
RUN mkdir -p /app
COPY "build/yggdrasil-${TARGETOS}-${TARGETARCH}" /app/yggdrasil

EXPOSE 8080
VOLUME /app/data
COPY assets /app/data/assets/
COPY templates /app/data/templates/

WORKDIR /app/data
ENTRYPOINT ["/app/yggdrasil"]
