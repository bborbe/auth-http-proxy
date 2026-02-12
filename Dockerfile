# syntax=docker/dockerfile:1

FROM golang:1.26.0 AS build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w" -o /main .

FROM alpine:3.21
RUN apk --no-cache add \
    ca-certificates \
    rsync \
    openssh-client \
    tzdata
COPY --from=build /main /main
COPY --from=build /usr/local/go/lib/time/zoneinfo.zip /
ENV ZONEINFO=/zoneinfo.zip
ENTRYPOINT ["/main"]
