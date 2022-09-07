# Vending Machine API

## Requirements

- Go >= 1.19

## Build docker image

```
VERSION=1.2 && docker build -t vending-machine:${VERSION} --build-arg VERSION=${VERSION} .
```

## Generate code coverage report

```shell
go test -coverprofile=cover.out ./pkg/...
go tool cover -html=cover.out -o cover.html
```