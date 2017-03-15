FROM alpine:3.5

EXPOSE 3000

# install certificates for datadog
RUN apk add --no-cache ca-certificates

COPY go-remote-config /go-remote-config
ENTRYPOINT ["/go-remote-config"]
