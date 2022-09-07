# Vending Machine API

## Requirements

- Go >= 1.19

## Build docker image

```
VERSION=1.2 && docker build -t vending-machine:${VERSION} --build-arg VERSION=${VERSION} .
```