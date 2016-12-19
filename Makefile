all: test install run
install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install bin/auth_http_proxy_server/*.go
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
runledisserver:
	ledis-server \
	-addr=localhost:5555 \
	-databases=1
runldapserver:
	docker run \
	-p 389:389 -p 636:636 \
	-e LDAP_SECRET='S3CR3T' \
	-e LDAP_SUFFIX='dc=example,dc=com' \
	-e LDAP_ROOTDN='cn=root,dc=example,dc=com' \
	bborbe/openldap:latest
runauthserver:
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
runhttpdumpserver:
	debug_server \
	-logtostderr \
	-v=2 \
	-port=7777
runauth:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-port=8888 \
	-basic-auth-realm=TestAuth \
	-target-address=localhost:7777 \
	-kind=basic \
	-verifier=auth \
	-auth-url=http://localhost:6666 \
	-auth-application-name=auth \
	-auth-application-password=test123
runfile:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-port=8888 \
	-basic-auth-realm=TestAuth \
	-target-address=localhost:7777 \
	-kind=basic \
	-verifier=file \
	-file-users=sample_users
runldap:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-port=8888 \
	-kind=basic \
	-basic-auth-realm=TestAuth \
	-target-address=localhost:7777 \
	-verifier=ldap \
	-ldap-host="localhost" \
	-ldap-port=389 \
	-ldap-use-ssl=false \
	-ldap-skip-tls=true \
	-ldap-bind-dn="cn=root,dc=example,dc=com" \
	-ldap-bind-password="S3CR3T" \
	-ldap-base-dn="dc=example,dc=com" \
	-ldap-user-filter="(uid=%s)" \
	-ldap-group-filter="(member=uid=%s,ou=users,dc=example,dc=com)" \
	-ldap-user-dn="ou=users" \
	-ldap-group-dn="ou=groups" \
	-ldap-user-field="uid" \
	-ldap-group-field="ou" \
	-required-groups="admins"
runauthconfig:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-config=sample_config_auth.json
runfileconfig:
	auth_http_proxy_server \
	-logtostderr \
	-v=2 \
	-config=sample_config_file.json
runhtml:
	auth_http_proxy_server \
	-logtostderr \
	-v=4 \
	-port=8888 \
	-target-address=localhost:7777 \
	-target-healthz-url=http://localhost:7777 \
	-verifier=file \
	-file-users=sample_users \
	-kind=html \
	-secret=AES256Key-32Characters1234567890
runbasic:
	auth_http_proxy_server \
	-logtostderr \
	-v=4 \
	-port=8888 \
	-target-address=localhost:7777 \
	-target-healthz-url=http://localhost:7777 \
	-verifier=file \
	-file-users=sample_users \
	-kind=basic \
	-basic-auth-realm="Test Auth"
run: runhtml
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
