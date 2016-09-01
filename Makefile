install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install bin/auth_http_proxy_server/auth_http_proxy_server.go
test:
	GO15VENDOREXPERIMENT=1 go test -cover `glide novendor`
vet:
	go tool vet .
	go tool vet --shadow .
lint:
	golint -min_confidence 1 ./...
errcheck:
	errcheck -ignore '(Close|Write)' ./...
check: lint vet errcheck
runledis:
	ledis-server \
	-addr=localhost:5555 \
	-databases=1
runauth:
	auth_server \
	-logtostderr \
	-v=2 \
	-port=6666 \
	-prefix=/auth \
	-ledisdb-address=localhost:5555 \
	-auth-application-password=test123
runfileserver:
	file_server \
	-logtostderr \
	-v=2 \
	-port=7777 \
	-root=/tmp
runwithauth:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-port=8888 \
	-basic-auth-realm=TestAuth \
	-target-address=localhost:7777 \
	-verifier=auth \
	-auth-url=http://localhost:6666 \
	-auth-application-name=auth \
	-auth-application-password=test123
runwithfile:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-port=8888 \
	-basic-auth-realm=TestAuth \
	-target-address=localhost:7777 \
	-verifier=file \
	-file-users=sample_users
open:
	open http://localhost:8888/
format:
	find . -name "*.go" -exec gofmt -w "{}" \;
	goimports -w=true .
prepare:
	go get -u github.com/bborbe/server/bin/file_server
	go get -u github.com/bborbe/auth/bin/auth_server
	go get -u github.com/siddontang/ledisdb/cmd/ledis-server
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/Masterminds/glide
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	glide install
update:
	glide up
clean:
	rm -rf var vendor target node_modules
