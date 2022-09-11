# Vending Machine API

[![Build Status](https://github.com/mehiX/vending-machine-api/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/features/actions)
[![codecov](https://codecov.io/gh/mehiX/vending-machine-api/branch/main/graph/badge.svg?token=DR9TYNBWAK)](https://codecov.io/gh/mehiX/vending-machine-api)
[![Go Report Card](https://goreportcard.com/badge/github.com/mehiX/vending-machine-api)](https://goreportcard.com/report/github.com/mehiX/vending-machine-api)
[![Release](https://img.shields.io/github/release/mehiX/vending-machine-api.svg?style=flat-square)](https://github.com/mehiX/vending-machine-api/releases)

## Requirements

- Go >= 1.18

## Setup and dependencies

```
cp .env.tmpl .env
```

Edit the values in `.env` to match your environment.

```
# install go-swagger cli
go install github.com/swaggo/swag/cmd/swag@v1.8.3

# generate swagger docs
go generate ./...
```

## Run 

```
# check available command line options (provided defaults should work)
go run ./cmd/server/... -h

# run with defaults
go run ./cmd/server/...

# specify the listening address on the command line
go run ./cmd/server/... -l 127.0.0.1:9999
```

## Build and run with Docker

```
docker login ghcr.io

docker-compose up -d --pull always && docker-compose logs -f vm
```

To test local changes add the `--build` flag:

```
docker-compose up -d --build && docker-compose logs -f vm
```

## Generate code coverage report

```shell
go test -coverprofile=cover.out ./internal/app/...
go tool cover -html=cover.out -o cover.html
```