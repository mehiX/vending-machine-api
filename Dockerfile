#build stage
FROM golang:alpine AS builder

ARG VERSION=1.0.2

RUN apk add --no-cache gcc libc-dev git
WORKDIR /go/src/app
COPY . .

# Normally would use `latest` but current version v1.8.5 throws an error at the moment
RUN go install github.com/swaggo/swag/cmd/swag@v1.8.3
RUN swag init -d "./cmd/server" --parseDependency --parseDepth 1
RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /go/bin/app -v ./cmd/server/...
RUN go build -o /go/bin/healthcheck -v ./cmd/healthcheck/...

#final stage
FROM alpine:latest

ARG VERSION=${VERSION}

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder --chown=app /go/bin/app /app
COPY --from=builder --chown=app /go/bin/healthcheck /healthcheck
USER app

ENTRYPOINT ["/app"] 
CMD ["-e", "''", "-l", ":80"]

LABEL Name=VendinMachineAPI Version=${VERSION}

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD /healthcheck http://localhost/health || exit 1