# 下面是一些自定义的 Makefile
#include ../../../../Makefile

GOHOSTOS:=$(shell go env GOHOSTOS)
GOPATH:=$(shell go env GOPATH)
VERSION=$(shell git describe --tags --always)
GOVERSION := $(shell go env GOVERSION)
GITCOMMIT := $(shell git rev-parse --short HEAD)
BUILT := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

GOOS := $(shell go env GOOS)
GOARCH := $(shell go env GOARCH)
OSARCH := $(GOOS)/$(GOARCH)
ENV:=pre
KPIDAY:=30
PROJECT_NAME:=api
GO_BIN_OUTPUT:=./bin/codo-notice
GO_CODE_DIR:=.


GOLDFLAGS = -X 'codo-notice/meta.Version=${BUIlDTAG}' \
	-X 'codo-notice/meta.GoVersion=${GOVERSION}' \
	-X 'codo-notice/meta.GitCommit=${GITCOMMIT}' \
	-X 'codo-notice/meta.Built=${BUILT}' \
	-X 'codo-notice/meta.ENV=${ENV}' \
	-X 'codo-notice/meta.OsArch=${OSARCH}'
ifeq ($(GOHOSTOS), windows)
	#the `find.exe` is different from `find` in bash/shell.
	#to see https://docs.microsoft.com/en-us/windows-server/administration/windows-commands/find.
	#changed to use git-bash.exe to run find cli or other cli friendly, caused of every developer has a Git.
	#Git_Bash= $(subst cmd\,bin\bash.exe,$(dir $(shell where git)))
	Git_Bash=$(subst \,/,$(subst mingw64\bin\,bin\bash.exe,$(dir $(shell where git |head -n 1))))
	INTERNAL_PROTO_FILES=$(shell $(Git_Bash) -c "find ./internal -name *.proto")
	API_PROTO_FILES=$(shell $(Git_Bash) -c "find ./pb -name *.proto")
else
	INTERNAL_PROTO_FILES=$(shell find ./internal -name *.proto)
	API_PROTO_FILES=$(shell find ./pb -name *.proto)
endif


.PHONY: init
# init env
init:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	go install github.com/google/wire/cmd/wire@latest
	go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest
	go install mvdan.cc/gofumpt@latest
	go install helm.sh/helm/v3/cmd/helm@latest
	CGO_ENABLED=0 go install github.com/envoyproxy/protoc-gen-validate@v1.0.2
	go install github.com/google/gnostic/cmd/protoc-gen-openapi@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.16.0
	go install github.com/favadi/protoc-go-inject-tag@latest
	go install go.uber.org/nilaway/cmd/nilaway@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest
	go install github.com/Ccheers/protoc-gen-go-kratos-http@latest
	go install golang.org/x/tools/cmd/goimports@latest


.PHONY: config
# generate internal proto
config:
	protoc --proto_path=./internal \
	       --proto_path=./third_party \
 	       --go_out=paths=source_relative:./internal \
	       ./internal/conf/conf.proto

.PHONY: lint
# 格式化代码
lint:
	nilaway ./...

.PHONY: fmt
# 格式化代码
fmt:
	gofumpt -l -w .
	goimports -l -w .


.PHONY: gen_dao
# 生成 dao 层，依赖于 hack/config.yaml
gen_dao:
	gf gen dao

.PHONY: changelog
# 生成 changelog
changelog:
	git-chglog -o ./CHANGELOG.md

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-s -w $(GOLDFLAGS)" -o $(GO_BIN_OUTPUT) $(GO_CODE_DIR)

# 查看最近写了多少代码
.PHONY: kpi
kpi:
ifeq ($(shell uname), Darwin)
	git log --since="`date -v -$(KPIDAY)d +"%Y-%m-%d"`" --before="`date +"%Y-%m-%d"`" --author="`git config --get user.name`" --pretty=tformat: --numstat -- . ":(exclude)*.json" ":(exclude)*.yaml" ":(exclude)*_route.go" ":(exclude)*_handler.go" ":(exclude)app/*/cmd/rpc/client/*" ":(exclude)app/*/cmd/*/pb/*" ":(exclude)app/*/cmd/api/internal/handler/*" ":(exclude)app/*/cmd/rpc/internal/server/*" ":(exclude)app/*/model/*" | awk '{ add += $$1; subs += $$2; loc += $$1 - $$2 } END { printf "added lines: %s removed lines : %s total lines: %s\n",add,subs,loc }'
else
	git log --since="`date -d "$(KPIDAY) day ago" +"%Y-%m-%d"`" --before="`date +"%Y-%m-%d"`" --author="`git config --get user.name`" --pretty=tformat: --numstat -- . ":(exclude)*.json" ":(exclude)*.yaml" ":(exclude)*_route.go" ":(exclude)*_handler.go" ":(exclude)app/*/cmd/rpc/client/*" ":(exclude)app/*/cmd/*/pb/*" ":(exclude)app/*/cmd/api/internal/handler/*" ":(exclude)app/*/cmd/rpc/internal/server/*" ":(exclude)app/*/model/*" | awk '{ add += $$1; subs += $$2; loc += $$1 - $$2 } END { printf "added lines: %s removed lines : %s total lines: %s\n",add,subs,loc }'
endif

.PHONY: api
# 生成 api 层，需要在 app/{service}/cmd/rpc/pb 目录执行
api:
	protoc --proto_path=. \
	--proto_path=./third_party \
	--go-kratos-http_out=paths=source_relative:. \
	--go_out=paths=source_relative:. \
	--go-grpc_out=paths=source_relative:. \
    --openapiv2_out=enums_as_ints=true,json_names_for_fields=false,allow_delete_body=true:. \
    --validate_out=lang=go,paths=source_relative:. \
    --openapi_out=naming=proto:. \
	$(API_PROTO_FILES)

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help