FROM alpine:3.5

EXPOSE 3000

# install certificates for datadog
RUN apk add --no-cache ca-certificates tzdata

# Create default config in container
RUN mkdir -p /rules && echo '[]' > /rules/config.json

COPY go-remote-config /go-remote-config
ENTRYPOINT ["/go-remote-config", "--config=/rules/config.json"]
