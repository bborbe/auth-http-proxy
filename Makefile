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
	find . -type f -name '*.go' -not -path './vendor/*' -exec gofmt -w "{}" +
	go run -mod=mod github.com/incu6us/goimports-reviser/v3 -project-name github.com/bborbe/auth-http-proxy -format -excludes vendor ./...
	find . -type d -name vendor -prune -o -type f -name '*.go' -print0 | xargs -0 -n 10 go run -mod=mod github.com/segmentio/golines --max-len=100 -w

generate:
	rm -rf mocks avro
	go generate -mod=mod ./...

.PHONY: test
test:
	go test -mod=mod -p=$${GO_TEST_PARALLEL:-1} -cover -race $(shell go list -mod=mod ./... | grep -v /vendor/)

check: vet errcheck vulncheck osv-scanner gosec trivy

vet:
	go vet -mod=mod $(shell go list -mod=mod ./... | grep -v /vendor/)

errcheck:
	go run -mod=mod github.com/kisielk/errcheck -ignore '(Close|Write|Fprint)' $(shell go list -mod=mod ./... | grep -v /vendor/)

vulncheck:
	go run -mod=mod golang.org/x/vuln/cmd/govulncheck $(shell go list -mod=mod ./... | grep -v /vendor/)

osv-scanner:
	@if [ -f .osv-scanner.toml ]; then \
		echo "Using .osv-scanner.toml"; \
		go run -mod=mod github.com/google/osv-scanner/v2/cmd/osv-scanner --config .osv-scanner.toml --recursive .; \
	else \
		echo "No config found, running default scan"; \
		go run -mod=mod github.com/google/osv-scanner/v2/cmd/osv-scanner --recursive .; \
	fi

gosec:
	go run -mod=mod github.com/securego/gosec/v2/cmd/gosec -exclude=G104 ./...

trivy:
	trivy fs --scanners vuln,secret --quiet --no-progress --disable-telemetry --exit-code 1 .

addlicense:
	go run -mod=mod github.com/google/addlicense -c "Benjamin Borbe" -y $$(date +'%Y') -l bsd $$(find . -name "*.go" -not -path './vendor/*')

.PHONY: build
build:
	@tags=""; \
	for i in $(VERSIONS); do \
		tags="$$tags -t $(REGISTRY)/$(IMAGE):$$i"; \
	done; \
	echo "docker build --rm=true $$tags ."; \
	DOCKER_BUILDKIT=1 docker build --rm=true --platform=linux/amd64 $$tags .

.PHONY: upload
upload:
	@for i in $(VERSIONS); do \
		echo "docker push $(REGISTRY)/$(IMAGE):$$i"; \
		docker push $(REGISTRY)/$(IMAGE):$$i; \
	done

.PHONY: clean
clean:
	@for i in $(VERSIONS); do \
		echo "docker rmi $(REGISTRY)/$(IMAGE):$$i"; \
		docker rmi $(REGISTRY)/$(IMAGE):$$i || true; \
	done

.PHONY: buca
buca: build upload clean
