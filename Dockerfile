FROM golang:1.15.8-alpine as build
ENV CGO_ENABLED=0

ARG GOOS=linux
ARG GOARCH=amd64
ARG LDFLAGS

RUN apk add --no-cache git ca-certificates openssh bash

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN GOOS=${GOOS} GOARCH=${GOARCH} go build \
  ${LDFLAGS} \
    ${LDFLAGS} \
  -ldflags " \
  -X 'main.gitCommit=$(git rev-parse --short HEAD)' -X 'main.date=$(date --utc +%F_%T)'" \
  -o "custom-error-pages" .

FROM build as test

ENV CI true

RUN go test ./...

FROM alpine:3.13.1

COPY etc etc
COPY www www
COPY --from=build /app/custom-error-pages /custom-error-pages

ENV ERROR_FILES_PATH /www

CMD ["/custom-error-pages"]
