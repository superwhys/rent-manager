rent-manager-server/
├── main.go
│
├── internal/
│   ├── domain/                    # 领域层（核心业务）
│   │   ├── order/
│   │   │   ├── entity.go          # 实体
│   │   │   ├── value_object.go    # 值对象
│   │   │   ├── repository.go      # 仓储接口
│   │   │   ├── service.go         # 领域服务
│   │   │   ├── event.go           # 领域事件
│   │   │   └── errors.go          # 领域错误定义
│   │   │
│   │   └── user/
│   │       ├── entity.go
│   │       ├── value_object.go
│   │       ├── repository.go
│   │       ├── service.go
│   │       └── errors.go
│   │
│   ├── app/               # 应用层（用例编排）
│   │   ├── order/
│   │   │   ├── dto.go
│   │   │   └── service.go         # 应用服务
│   │   │
│   │   └── user/
│   │       ├── dto.go
│   │       └── service.go
│   │
│   ├── infrastructure/            # 基础设施层（技术实现）
│   │   ├── persistence/
│   │   │   ├── mysql/
│   │   │   │   ├── order_repository.go
│   │   │   │   └── user_repository.go
│   │   │   └── model/
│   │   │       ├── order_po.go    # 持久化对象
│   │   │       └── user_po.go
│   │   │
│   │   ├── cache/
│   │   │   └── redis.go
│   │   │
│   │   ├── mq/
│   │   │   └── kafka.go
│   │   │
│   │   ├── http/
│   │   │   └── client.go
│   │   │
│   │   └── config/
│   │       └── config.go
│   │
│   ├── interfaces/                # 接口适配层（入站）
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── order_handler.go
│   │   │   │   └── user_handler.go
│   │   │   ├── router.go
│   │   │   ├── request/
│   │   │   └── response/
│   │   │
│   │   ├── grpc/
│   │   │   └── handler/
│   │   │
│   │   └── event/
│   │       └── consumer/
│   │
│   └── pkg/                       # internal 内部共享工具（可选）
│       ├── logger/
│       ├── xerr/
│       └── utils/
│
├── pkg/                           # 可对外复用的公共包（谨慎使用）
│   └── ...
│
├── api/                           # OpenAPI / proto 文件
│   ├── openapi/
│   └── proto/
│
├── configs/
│   ├── config.yaml
│   └── ...
│
├── scripts/
│   └── ...
│
├── test/
│   ├── integration/
│   └── e2e/
│
├── go.mod
└── README.md
