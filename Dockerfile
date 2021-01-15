# ---- Build ----
FROM golang:1.15 as builder
WORKDIR /go/src/github.com/flyem-services/
# fetch source from github tag
ARG VERSION=master
RUN git clone --depth 1 --branch ${VERSION} https://github.com/janelia-flyem/flyem-services.git .
# run go get
RUN go get -d -v ./
# run go build
RUN CGO_ENABLED=0 GOOS=linux go build .

# ---- cert Build ----
FROM ubuntu:20.04 as certbuilder
RUN apt-get -y update && apt-get -y install openssl
WORKDIR /opt/certs
RUN openssl req \
    -newkey rsa:4096 -nodes -sha256 -keyout key.pem \
    -x509 -days 365 -out cert.pem \
    -subj /CN=\*.janelia.org

# ---- Release ----
FROM alpine:3.12.3
MAINTAINER flyem project team
LABEL maintainer="plazas@janelia.hhmi.org"
ARG VERSION=master
LABEL version=${VERSION}
RUN apk add --no-cache bash
COPY --from=builder /go/src/github.com/flyem-services/flyem-services /app/
COPY --from=builder /go/src/github.com/flyem-services/swaggerdocs /app/swaggerdocs/
COPY --from=builder /go/src/github.com/flyem-services/config.json.example /app/config/config.json

COPY --from=certbuilder /opt/certs /app/certs/
WORKDIR /app
CMD ./flyem-services -port 15000 -proxy-port 443 /app/config/config.json

