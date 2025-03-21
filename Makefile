REGISTRY ?= docker.io
IMAGE ?= bborbe/auth-http-proxy
VERSION  ?= latest
VERSIONS = $(VERSION)
VERSIONS += $(shell git fetch --tags; git tag -l --points-at HEAD)

default: precommit

install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install *.go

build:
	@tags=""; \
	for i in $(VERSIONS); do \
		tags="$$tags -t $(REGISTRY)/$(IMAGE):$$i"; \
	done; \
	echo "docker build --no-cache --rm=true $$tags ."; \
	docker build --no-cache --rm=true --platform=linux/amd64 $$tags .

clean:
	@for i in $(VERSIONS); do \
		echo "docker rmi $(REGISTRY)/$(IMAGE):$$i"; \
		docker rmi $(REGISTRY)/$(IMAGE):$$i || true; \
	done

upload:
	@for i in $(VERSIONS); do \
		echo "docker push $(REGISTRY)/$(IMAGE):$$i"; \
		docker push $(REGISTRY)/$(IMAGE):$$i; \
	done

versions:
	@for i in $(VERSIONS); do echo $$i; done;


precommit: ensure format generate test check addlicense
	@echo "ready to commit"

ensure:
	go mod verify
	go mod vendor

format:
	go run -mod=vendor github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/auth-http-proxy -format -excludes vendor ./...

generate:
	rm -rf mocks avro
	go generate -mod=vendor ./...

test:
	go test -mod=vendor -p=1 -cover -race $(shell go list -mod=vendor ./... | grep -v /vendor/)

check: errcheck vulncheck

errcheck:
	go run -mod=vendor github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=vendor ./... | grep -v /vendor/)

addlicense:
	go run -mod=vendor github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

vulncheck:
	go run -mod=vendor golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=vendor ./... | grep -v /vendor/)

