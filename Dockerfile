FROM golang:1.22 AS builder
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace

# 构建配置
ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOPROXY="https://goproxy.cn,direct"

# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY . .

# Build
# the GOARCH has not a default value to allow the binary be built according to the host where the command
# was called. For example, if we call make docker-build in a local env which has the Apple Silicon M1 SO
# the docker BUILDPLATFORM arg will be linux/arm64 when for Apple x86 it will be linux/amd64. Therefore,
# by leaving it empty we can ensure that the container and binary shipped on it will have the same platform.
RUN make build GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH}

FROM debian:stable-slim

# 垃圾回收模式调整 https://ms2008.github.io/2019/06/30/golang-madvfree/
ENV GODEBUG="madvdontneed=1"

# tls 更新
RUN apt-get update && apt-get install -y --no-install-recommends \
		ca-certificates  \
        netbase \
        && rm -rf /var/lib/apt/lists/ \
        && apt-get autoremove -y && apt-get autoclean -y

COPY --from=builder /workspace/bin/codo-notice /app/

WORKDIR /app

EXPOSE 8000
EXPOSE 9000
VOLUME /data/conf

CMD ["./codo-notice", "-conf", "/data/conf/config.yaml"]
