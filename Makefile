BINPATH ?= build

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)

LDFLAGS = -ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: all
all: audit test build

.PHONY: audit
audit:
	go list -json -m all | nancy sleuth --exclude-vulnerability-file ./.nancy-ignore

.PHONY: build
build:
	go build -tags 'production' $(LDFLAGS) -o $(BINPATH)/dp-import-cantabular-dimension-options

.PHONY: debug
debug:
	go build -tags 'debug' $(LDFLAGS) -o $(BINPATH)/dp-import-cantabular-dimension-options
	HUMAN_LOG=1 DEBUG=1 $(BINPATH)/dp-import-cantabular-dimension-options

.PHONY: debug-run
debug-run:
	HUMAN_LOG=1 DEBUG=1 go run -tags 'debug' $(LDFLAGS) main.go

.PHONY: lint
lint:
	exit

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: produce
produce:
	HUMAN_LOG=1 go run cmd/producer/main.go

.PHONY: convey
convey:
	goconvey ./...

.PHONY: test-component
test-component:
	cd features/compose; docker-compose up --abort-on-container-exit
	echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"
