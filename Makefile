GO ?= go
GO_PACKAGE := github.com/hummerd/app
CMD := app
OUT := bin/app
GIT_BRANCH := $$(git rev-parse --abbrev-ref HEAD)
GIT_REV := $$(git rev-parse HEAD)
TEST_PACKAGE = `go list ./... | grep -v /vendor/`


build:
	$(info building...)
	go build -ldflags "-X ${GO_PACKAGE}/internal/config.Version=$(GIT_REV)" -o $(OUT) ./internal/cmd
.PHONY: build

hooks:
	git config core.hooksParh .githooks
.PHONY: hooks

build.linux: export GOOS = linux
build.linux: export GOARCH = amd64
build.linux: export OUT = bin/$(GOOS)_$(GOARCH)/$(CMD)
build.linux: build
.PHONY: build.linux

build.docker:
	docker run --rm  -v $(PWD):/go/src/$(GO_PACKAGE) -w /go/src/$(GO_PACKAGE) go_build_env /bin/sh -c "make build.linux"
.PHONY: build.docker

check:
	gometalinter ./... --vendor --disable-all --enable=vet --enable=golint --exclude=comment --enable=staticcheck --enable=errcheck --enable=vetshadow

run: build 
	$(info running...)
	$(OUT)
.PHONY: run

test:
	$(info testing...)
	@$(GO) test -race -timeout 60s ${TEST_PACKAGE}
.PHONY: test

test.coverage:
	$(info testing...)
	@$(GO) test ${TEST_PACKAGE} -race -timeout 60s -coverprofile cover.out || exit 1
	go tool cover -func cover.out
.PHONY: test.coverage

lint:
	$(info linting...)
	gometalinter internal/... --exclude=vendor --exclude="should have comment or be unexported" --cyclo-over=20
.PHONY: lint

