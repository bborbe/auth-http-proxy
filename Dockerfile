FROM golang:1.12.0 AS build
COPY . /go/src/github.com/bborbe/auth-http-proxy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /auth-http-proxy ./src/github.com/bborbe/auth-http-proxy
CMD ["/bin/bash"]

FROM alpine:3.9 as alpine
RUN apk --no-cache add ca-certificates

FROM scratch
MAINTAINER Benjamin Borbe <bborbe@rocketnews.de>
COPY --from=build /auth-http-proxy /auth-http-proxy
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/auth-http-proxy"]
