GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)

ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst cmd\,bin\bash.exe,$(dir $(shell where git))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find api -name *.proto")
	# 查找api/protos目录下所有的子目录
    API_PROTO_DIRS := $(shell find api/protos -type d -mindepth 1 -maxdepth 1)
else
	INTERNAL_PROTO_FILES=$(shell find internal -name *.proto)
	API_PROTO_FILES=$(shell find api -name *.proto)
	# 查找api/protos目录下所有的子目录
    API_PROTO_DIRS := $(shell find api/protos -type d -mindepth 1 -maxdepth 1)
endif

.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.8
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/google/wire/cmd/wire@latest
	go install go.uber.org/mock/mockgen@v0.6.0

.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       $(INTERNAL_PROTO_FILES)

.PHONY: api
# generate api proto
api:
	@echo "Generating code for all proto directories..."
	@for dir in $(API_PROTO_DIRS); do \
		dirname=$$(basename $$dir); \
		if [ ! -d "./api/$$dirname" ]; then \
			echo "Creating directory ./api/$$dirname..."; \
			mkdir -p ./api/$$dirname; \
		fi; \
		echo "Processing $$dirname..."; \
		protoc --proto_path=./api/protos/$$dirname \
		       --proto_path=./third_party \
		       --go_out=paths=source_relative:./api/$$dirname \
		       --go-errors_out=paths=source_relative:./api/$$dirname \
		       --go-http_out=paths=source_relative:./api/$$dirname \
		       --go-grpc_out=paths=source_relative:./api/$$dirname \
		       --openapi_out=fq_schema_naming=true,default_response=false:./api/$$dirname \
		       --validate_out=paths=source_relative,lang=go:./api/$$dirname \
		       $$dir/*.proto; \
	done
	@echo "Code generation completed!"

.PHONY: build
# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/ ./...

dao:
	go run internal/scripts/dao-gen/*.go

.PHONY: generate
# generate
generate:
	go mod tidy
	go generate ./...
	wire ./...

.PHONY: all
# generate all
all:
	make api;
	make config;
	make generate;

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
