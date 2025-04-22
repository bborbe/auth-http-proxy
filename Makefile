REGISTRY ?= docker.io
IMAGE ?= bborbe/auth-http-proxy
VERSION  ?= latest
VERSIONS = $(VERSION)
VERSIONS += $(shell git fetch --tags; git tag -l --points-at HEAD)

default: precommit

precommit: ensure format generate test check addlicense
	@echo "ready to commit"

ensure:
	go mod tidy
	go mod verify
	rm -rf vendor

format:
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/auth-http-proxy -format -excludes vendor ./...

generate:
	rm -rf mocks avro
	go generate -mod=mod ./...

test:
	go test -mod=mod -p=1 -cover -race $(shell go list -mod=mod ./... | grep -v /vendor/)

check: errcheck vulncheck

errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/)

addlicense:
	go run -mod=mod github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)

install:
	GOBIN=$(GOPATH)/bin GO15VENDOREXPERIMENT=1 go install *.go

versions:
	@for i in $(VERSIONS); do echo $$i; done;

.PHONY: build
build:
	go mod vendor
	@tags=""; \
	for i in $(VERSIONS); do \
		tags="$$tags -t $(REGISTRY)/$(IMAGE):$$i"; \
	done; \
	echo "docker build --no-cache --rm=true $$tags ."; \
	docker build --no-cache --rm=true --platform=linux/amd64 $$tags .

.PHONY: clean
clean:
	@for i in $(VERSIONS); do \
		echo "docker rmi $(REGISTRY)/$(IMAGE):$$i"; \
		docker rmi $(REGISTRY)/$(IMAGE):$$i || true; \
	done
	rm -rf vendor

.PHONY: upload
upload:
	@for i in $(VERSIONS); do \
		echo "docker push $(REGISTRY)/$(IMAGE):$$i"; \
		docker push $(REGISTRY)/$(IMAGE):$$i; \
	done


.PHONY: apply
apply:
	@for i in $(DIRS); do \
		cd $$i; \
		echo "apply $${i}"; \
		make apply; \
		cd ..; \
	done

.PHONY: buca
buca: build upload clean apply
