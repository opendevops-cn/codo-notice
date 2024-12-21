# 项目模板

# 目录结构
```text
.
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── etc # 配置文件
│   └── config.yaml
├── go.mod
├── go.sum
├── internal
│   ├── biz # 核心业务层
│   │   ├── README.md
│   │   ├── biz.go
│   │   └── greeter.go
│   ├── conf # 配置定义层
│   │   ├── conf.pb.go
│   │   └── conf.proto
│   ├── impl # biz 实现层
│   │   ├── README.md
│   │   └── repo
│   │       ├── data.go
│   │       └── greeter.go
│   ├── dep # 第三方依赖
│   │   ├── data.go
│   │   ├── otel.go
│   │   └── provider.go
│   ├── server # 运输服务层
│   │   ├── grpc.go
│   │   ├── http.go
│   │   └── server.go
│   └── service # 协议转换层
│       ├── README.md
│       ├── greeter.go
│       └── service.go
├── main.go # 函数入口
├── openapi.yaml # openaiv3 接口文档
├── pb # protobuf 定义
│   ├── greeter.pb.go
│   ├── greeter.proto
│   ├── greeter_grpc.pb.go
│   └── greeter_http.pb.go
├── third_party # proto 第三方依赖
│   ├── README.md
│   ├── google
│   │   ├── api
│   │   └── protobuf
│   ├── openapi
│   │   └── v3
│   └── validate
│       ├── README.md
│       └── validate.proto
├── wire.go # wire 文件
└── wire_gen.go
```