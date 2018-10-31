FROM golang:1.11.1-alpine3.8 as builder

RUN apk add --no-cache git

ENV GO111MODULE=on
ENV PACKAGE github.com/mopsalarm/go-remote-config
WORKDIR $GOPATH/src/$PACKAGE/

COPY go.mod go.sum ./
RUN go mod download

ENV CGO_ENABLED=0

COPY . .
RUN go build -v -ldflags="-s -w" -o /go-remote-config .


FROM alpine:3.8
RUN apk add --no-cache tzdata ca-certificates
EXPOSE 3000

COPY --from=builder /go-remote-config /

ENTRYPOINT ["/go-remote-config", "--config=/rules/config.json"]
