# Vending Machine API

[![Go Report Card](https://goreportcard.com/badge/github.com/mehiX/vending-machine-api)](https://goreportcard.com/report/github.com/mehiX/vending-machine-api)

## Requirements

- Go >= 1.19

## Setup

```
cp .env.tmpl .env
```

Edit the values in `.env` to match your environment.

## Run 

```
# check available command line options (provided defaults should work)
go run ./cmd/server/... -h

# run with defaults
go run ./cmd/server/...

# run on different port
go run ./cmd/server/... -l 127.0.0.1:9999
```

## Build and run with Docker

```
export VERSION=1.4 

docker build -t vending-machine:${VERSION} --build-arg VERSION=${VERSION} .

docker run --rm -d --env-file .env -p 7777:80 vending-machine:${VERSION}
```

## Generate code coverage report

```shell
go test -coverprofile=cover.out ./pkg/...
go tool cover -html=cover.out -o cover.html
```