# These env vars have to be set in the CI
# GITHUB_TOKEN
# DOCKER_HUB_TOKEN

.PHONY: build deps test release clean push image ci-compile build-dir ci-dist dist-dir ci-release version help

PROJECT := rancher-letsencrypt
PLATFORMS := linux
ARCH := amd64
DOCKER_IMAGE := janeczku/$(PROJECT)

VERSION := $(shell cat VERSION)
SHA := $(shell git rev-parse --short HEAD)

all: help

help:
	@echo "make build - build binary in the current environment"
	@echo "make deps - install build dependencies"
	@echo "make vet - run vet & gofmt checks"
	@echo "make test - run tests"
	@echo "make clean - Duh!"
	@echo "make release - tag with version and trigger CI release build"
	@echo "make image - build Docker image"
	@echo "make dockerhub - build and push image to Docker Hub"
	@echo "make version - show app version"

build: build-dir
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 godep go build -ldflags "-X main.Version=$(VERSION) -X main.Git=$(SHA)" -o build/$(PROJECT)-linux-amd64

deps:
	go get github.com/tools/godep
	go get github.com/c4milo/github-release

vet:
	scripts/vet

test:
	godep go test -v ./...

release:
	git tag `cat VERSION`
	git push origin master --tags

clean:
	go clean
	rm -fr ./build
	rm -fr ./dist

dockerhub: image
	@echo "Pushing $(DOCKER_IMAGE):$(VERSION)"
	docker push $(DOCKER_IMAGE):$(VERSION)

image:
	docker build -t $(DOCKER_IMAGE):$(VERSION) -f Dockerfile.dev .

version:
	@echo $(VERSION) $(SHA)

ci-compile: build-dir $(PLATFORMS)

build-dir:
	@rm -rf build && mkdir build

dist-dir:
	@rm -rf dist && mkdir dist

$(PLATFORMS):
	CGO_ENABLED=0 GOOS=$@ GOARCH=$(ARCH) godep go build -ldflags "-X main.Version=$(VERSION) -X main.Git=$(SHA) -w -s" -a -o build/$(PROJECT)-$@-$(ARCH)/$(PROJECT)

ci-dist: ci-compile dist-dir
	$(eval FILES := $(shell ls build))
	@for f in $(FILES); do \
		(cd $(shell pwd)/build/$$f && tar -cvzf ../../dist/$$f.tar.gz *); \
		(cd $(shell pwd)/dist && shasum -a 512 $$f.tar.gz > $$f.sha512); \
		echo $$f; \
	done
	@cp -r $(shell pwd)/dist/* $(CIRCLE_ARTIFACTS)
	ls $(CIRCLE_ARTIFACTS)

ci-release:
	@previous_tag=$$(git describe --abbrev=0 --tags $(VERSION)^); \
	comparison="$$previous_tag..HEAD"; \
	if [ -z "$$previous_tag" ]; then comparison=""; fi; \
	changelog=$$(git log $$comparison --oneline --no-merges --reverse); \
	github-release $(CIRCLE_PROJECT_USERNAME)/$(CIRCLE_PROJECT_REPONAME) $(VERSION) master "**Changelog**<br/>$$changelog" 'dist/*'
