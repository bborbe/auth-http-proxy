FROM golang:1.10 AS build
COPY . /go/src/github.com/bborbe/auth-http-proxy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /auth-http-proxy ./src/github.com/bborbe/auth-http-proxy
CMD ["/bin/bash"]

FROM scratch
MAINTAINER Benjamin Borbe <bborbe@rocketnews.de>
COPY --from=build /auth-http-proxy /auth-http-proxy
COPY files/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ENTRYPOINT ["/auth-http-proxy"]
