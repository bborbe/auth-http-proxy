REGISTRY ?= docker.io
IMAGE ?= bborbe/auth-http-proxy
VERSION ?= latest

deps:
	go get -u github.com/golang/lint/golint
	go get -u github.com/kisielk/errcheck
	go get -u golang.org/x/tools/cmd/goimports

build:
	docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) -f Dockerfile .

upload:
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)

clean:
	docker rmi $(REGISTRY)/$(IMAGE):$(VERSION) || true

precommit: ensure format test check
	@echo "ready to commit"

ensure:
	go mod tidy
	go mod vendor

format:
	@go get golang.org/x/tools/cmd/goimports
	@find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	@find . -type f -name '*.go' -not -path './vendor/*' -exec goimports -w "{}" +

test:
	go test -cover -race $(shell go list ./... | grep -v /vendor/)

check: lint vet errcheck

lint:
	@go get github.com/golang/lint/golint
	@golint -min_confidence 1 $(shell go list ./... | grep -v /vendor/)

vet:
	@go vet $(shell go list ./... | grep -v /vendor/)

errcheck:
	@go get github.com/kisielk/errcheck
	@errcheck -ignore '(Close|Write|Fprint)' $(shell go list ./... | grep -v /vendor/)
