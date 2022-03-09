FROM golang:1.16.3 AS builder

ENV GOPRIVATE=github.com/blinkops

WORKDIR /go/src/github.com/blinkops/blink-http

COPY go.mod go.sum ./
RUN go mod download
COPY .. .

RUN go build -tags netgo -ldflags '-w -extldflags "-static"' -o /go/bin/blink-http ./cmd/main

FROM alpine:3.14.0 AS plugin

WORKDIR /blink-http
COPY --from=builder /go/bin/blink-http .
COPY config.yaml plugin.yaml ./
COPY actions ./actions
COPY plugins/wiz/actions ./plugins/wiz/actions
COPY plugins/aws-sso/actions ./plugins/aws-sso/actions
COPY plugins/bitbucket/actions ./plugins/bitbucket/actions
COPY plugins/crowdstrike/actions ./plugins/crowdstrike/actions
COPY plugins/datadog/actions ./plugins/datadog/actions
COPY plugins/ermetic/actions ./plugins/ermetic/actions

# Expose the gRPC port.
EXPOSE 1337

RUN chmod a+x blink-http

ENTRYPOINT ./blink-http
