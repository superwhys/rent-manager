# Rent Manager

租赁管理系统后端，基于 Go、Gin 和 GORM + Gen，使用 MySQL 存储业务数据、Redis 执行接口限流。当前提供微信登录、房源与房间管理、月度抄表、租金账单、首页统计以及按房源授权的 RBAC 接口。

本文说明根目录的 Go 后端。`web/` 为独立的 Vue 前端项目，说明见 [web/README.md](web/README.md)。

## 已实现功能

| 模块 | 当前能力 |
| --- | --- |
| 认证 | 微信 `code` 登录，首次登录创建账号，签发 JWT，刷新访问令牌 |
| 账号与设置 | 获取当前账号信息，读取和更新账号配置 |
| 房源单元（Unit） | 创建、查询、更新、删除楼栋/单元，区分自有与被授权房源，配置默认抄表项和抄表日 |
| 房间 | 单元下创建与查询房间，维护房间信息、租金和出租状态 |
| 房型 | 当前账号下的房型增删改查 |
| 抄表项 | 创建、查询和更新抄表项，配置单价、差值计费或直接计费 |
| 月度抄表 | 录入房间月度读数，计算费用，查询房间状态与单元汇总 |
| 租金账单 | 按房间或单元生成月账单，查询账单，标记收款或取消 |
| 首页 | 汇总收入、空房和抄表/收款待办 |
| RBAC | 管理账号私有角色、全局权限与权限组，为指定房源向其他账号授予角色 |

当前实现的边界见文末“当前限制”，前端展示的功能不一定都有后端接口。

## 技术栈

| 技术 | 版本/用途 |
| --- | --- |
| Go | `go.mod` 声明 `1.27.1` |
| Gin | `v1.12.0`，HTTP 路由与中间件 |
| GORM / Gen | `v1.31.2` / `v0.3.29`，独立持久化模型、类型安全查询与数据库迁移 |
| goutils | `v1.3.1`，配置解析、服务启动、请求处理、日志和基础设施配置 |
| MySQL | 业务数据持久化 |
| Redis | 接口限流，使用 `go-redis/v9` 与 `ulule/limiter/v3` |
| JWT | `golang-jwt/jwt/v5`，使用 HS256 签名 |
| Swagger | `swag v1.16.6` 生成接口文档 |

## 本地启动

### 1. 准备环境

需要 Go 1.27.1 或兼容的更高版本，以及可连接的 MySQL、Redis 和微信小程序配置。微信登录需要客户端提供有效的登录 `code`。

在 MySQL 中先创建数据库，例如：

```sql
CREATE DATABASE rent_manager
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;
```

服务启动时会执行 GORM AutoMigrate，创建或更新表结构；数据库账号需要相应权限。启动过程不负责创建数据库，也不自动初始化权限、权限组或角色。

### 2. 配置服务

在项目根目录创建 `config.local`，内容为 JSON。该文件名匹配现有 `.gitignore` 中的 `*.local`，适合保存本地配置。替换示例中的微信配置、密钥和数据库账号后使用：

```json
{
  "wechat": {
    "AppID": "your-wechat-app-id",
    "SecretID": "your-wechat-app-secret",
    "Jscode2SessionApi": "https://api.weixin.qq.com/sns/jscode2session"
  },
  "jwt": {
    "JwtSecret": "replace-with-your-jwt-secret",
    "JwtTimeout": "2h",
    "JwtRefreshTimeout": "168h"
  },
  "admin": {
    "SuperAdminAccountIDs": []
  },
  "mysql": {
    "instance": "127.0.0.1:3306",
    "database": "rent_manager",
    "username": "rent_manager",
    "password": "your-mysql-password",
    "charset": "utf8mb4",
    "pool_size": 10
  },
  "redis": {
    "server": "127.0.0.1:6379",
    "db": 0,
    "username": "",
    "password": "",
    "poolsize": 20
  }
}
```

配置由 `github.com/miebyte/goutils/flags` 解析，通过 `--configFile`（或 `-f`）指定文件。不指定时，依次查找当前目录、`./configs`、用户主目录下的 `config.json`。

| 配置 | 说明 |
| --- | --- |
| `wechat` | `AppID`、`SecretID`、`Jscode2SessionApi` 均不能为空 |
| `jwt` | 密钥不能为空，`JwtTimeout` 必须大于零；时长支持 `2h` 等字符串 |
| `jwt.JwtRefreshTimeout` | 未设置或不大于零时，默认为 `JwtTimeout × 24` |
| `admin.SuperAdminAccountIDs` | 允许写入全局权限和权限组的账号 ID；空列表表示无人拥有该管理权限 |
| `mysql.instance` | MySQL 的 `host:port`；`pool_size` 控制连接池大小 |
| `redis.server` | Redis 的 `host:port`；`poolsize` 为连接池大小 |

### 3. 启动与检查

在项目根目录执行：

```bash
go mod download
go run . --configFile ./config.local
```

也可以先构建二进制：

```bash
go build -o rent-manager-server .
./rent-manager-server --configFile ./config.local
```

当前监听地址在 `main.go` 中固定为 `0.0.0.0:8080`。

| 地址 | 用途 |
| --- | --- |
| `http://localhost:8080/api` | 业务接口前缀 |
| [Swagger UI](http://localhost:8080/swagger/index.html) | 查看接口与请求字段 |
| `http://localhost:8080/swagger/doc.json` | Swagger JSON |
| `http://localhost:8080/health_check` | 服务进程健康检查，成功返回 `ok` |

```bash
curl http://localhost:8080/health_check
```

健康检查由 `cores` 提供，返回进程响应状态，不检查 MySQL、Redis 或微信接口的可用性。当前 `main.go` 使用 `SwaggerRouter(false)`，始终启用 Swagger。

## 接口调用

### 登录与刷新

微信登录无需 JWT。将客户端获取的微信 `code` 发送到：

```bash
curl -X POST http://localhost:8080/api/auth/wechat-login \
  -H 'Content-Type: application/json' \
  -d '{"code":"your-wechat-login-code"}'
```

成功响应：

```json
{
  "code": 0,
  "data": {
    "access_token": "...",
    "refresh_token": "..."
  },
  "message": "success"
}
```

访问业务接口时携带 `Authorization: Bearer <access_token>`：

```bash
curl http://localhost:8080/api/accout/self \
  -H 'Authorization: Bearer your-access-token'
```

注意：当前账号路由实际拼写为 `/accout/self`。

刷新接口使用 `refresh_token` 获取新的访问令牌，返回原刷新令牌，刷新令牌到期后需要重新登录：

```bash
curl -X POST http://localhost:8080/api/auth/refresh-token \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token":"your-refresh-token"}'
```

### 接口概览

下表路径均以 `/api` 为前缀。除 `/auth/*` 外，业务接口均经过 JWT 校验。`:id` 等表示路径参数，详细请求和响应字段见 Swagger 及 `internal/app/<业务>/dto.go`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| POST | `/auth/wechat-login`、`/auth/refresh-token` | 微信登录、刷新令牌 |
| GET | `/accout/self` | 当前账号信息 |
| GET / PUT | `/settings` | 读取、更新账号设置 |
| GET | `/home` | 首页聚合数据 |
| GET / POST | `/units` | 单元列表、创建单元 |
| GET / PUT / DELETE | `/units/:unit_id` | 单元详情、更新、删除 |
| GET / POST | `/units/:unit_id/rooms` | 单元房间列表、创建房间 |
| GET / PUT / DELETE | `/rooms/:room_id` | 房间详情、更新、删除 |
| GET / POST | `/room-types` | 房型列表、创建房型 |
| GET / PUT / DELETE | `/room-types/:room_type_id` | 房型详情、更新、删除 |
| GET / POST | `/meter-items` | 抄表项列表、创建抄表项 |
| GET / PUT | `/meter-items/:meter_item_id` | 抄表项详情、更新 |
| GET / PUT | `/rooms/:room_id/meter-readings` | 查询、保存房间月度抄表 |
| GET | `/units/:unit_id/meter-readings` | 单元月度抄表汇总 |
| GET | `/rooms/:room_id/rent-bills`、`/units/:unit_id/rent-bills` | 房间、单元账单列表 |
| POST | `/rooms/:room_id/rent-bills/generate`、`/units/:unit_id/rent-bills/generate` | 生成月账单 |
| GET | `/rent-bills/:rent_bill_id` | 账单详情 |
| POST | `/rent-bills/:rent_bill_id/paid`、`/rent-bills/:rent_bill_id/cancel` | 标记收款、取消账单 |
| GET | `/roles` | 当前账号角色列表 |
| POST | `/roles/batch` | 批量创建角色 |
| GET / PUT / DELETE | `/roles/:role_id` | 角色详情、更新、删除 |
| POST / DELETE | `/roles/:role_id/permission-groups` | 绑定、移除角色权限组 |
| GET | `/admin/permissions`、`/admin/permission-groups` | 全局权限、权限组列表 |
| POST | `/admin/permissions/batch`、`/admin/permission-groups/batch` | 批量创建全局权限、权限组 |
| POST / DELETE | `/admin/permission-groups/:permission_group_id/permissions` | 绑定、移除权限组中的权限 |
| GET | `/units/:unit_id/grants` | 单元授权列表 |
| POST / DELETE | `/units/:unit_id/accounts/:account_id/roles` | 授予、撤销目标账号的单元角色 |

### 业务约定

- 月度抄表和账单使用 `period_month`，格式为 `YYYY-MM`。
- 房间 `price`、账单 `base_rent`/`meter_fee`/`total_amount` 和抄表 `fee`/`total_fee` 以整数“分”表示；`unit_price` 单位为“分/计量单位”。首页收入相关字段则按元格式化展示。
- 抄表项支持 `DIFFERENCE`（本次与上次读数差值）和 `DIRECT`（本次输入值）计费，费用为用量乘单价后四舍五入到分。
- 差值型抄表首次录入需要基准读数，后续可取此前最新读数；不能录入下降读数或未绑定到房间的抄表项。
- 月账单由基础租金和抄表费用组成，缺失所需抄表记录时不能生成；单元批量生成只处理已出租房间。
- 账单状态为 `UNPAID`、`PAID`、`CANCELLED`。重新生成可更新未付款账单，已付款或已取消的账单会跳过。

### 响应与限流

接口统一返回 `code`、`data`、`message`：`code == 0` 表示成功，失败时 `data` 通常为 `null`，`message` 提供错误说明。

多数业务错误仍返回 HTTP 200，客户端需要同时检查业务 `code`。认证失败通常返回 HTTP 401；触发限流返回 HTTP 429，限流存储异常返回 HTTP 500。

`/api` 接口使用 Redis 限流，当前每个“Authorization 请求头 + 请求路径 + 客户端 IP”组合在 10 秒内允许 10 次请求，包括公开认证接口。

## 权限与数据归属

资源以账号为归属边界。单元、角色、房型、抄表项和账号设置归属账号，房间归属单元，抄表记录和账单归属房间。

- 账号可以管理自己拥有的资源；删除单元及管理单元授权仅限所有者。
- 访问他人单元的业务数据时，服务检查该单元下的授权角色和权限代码。
- 角色属于创建它的账号，可以绑定全局权限组；权限组关联全局权限。
- 普通登录账号可以查询全局权限与权限组，写入操作要求账号 ID 位于 `admin.SuperAdminAccountIDs` 中。

授权关系为：

```text
被授权账号 → UnitAccountRole → 房源所有者的 Role → PermissionGroup → Permission
```

共享房源时，先创建权限与权限组，所有者创建角色并绑定权限组，再将该角色授予目标账号在指定单元使用。权限代码在 [internal/domain/rbac/permissions.go](internal/domain/rbac/permissions.go) 定义，例如 `unit.view`、`room.manage`、`meter_reading.manage`、`rent_bill.manage`。资源权限校验位于业务服务中，按代码匹配。

## 项目结构

```text
.
├── main.go                         # 配置解析、依赖组装、迁移与服务启动
├── config/                         # 微信、JWT、管理员配置
├── api/
│   ├── api.go                      # 路由与 Swagger 入口
│   ├── routers/                    # HTTP Handler 与 Swagger 注解
│   ├── middlewares/                # JWT 校验、Redis 限流
│   ├── response/                   # 统一响应与错误处理
│   └── docs/                       # 生成的 Swagger 文档
├── internal/
│   ├── app/                        # 按业务拆分的用例服务、DTO 与测试
│   ├── domain/                     # 领域模型、业务规则与仓储接口
│   ├── infra/
│   │   ├── auth/jwt/               # JWT 签发与解析
│   │   ├── persistence/
│   │   │   ├── mysql/              # GORM + Gen 仓储实现
│   │   │   ├── models/             # 持久化模型与 AllModels 迁移清单
│   │   │   ├── query/              # Gen 生成的类型安全查询
│   │   │   └── generate/           # 从模型生成 Query 的入口
│   │   ├── cache/                  # 缓存辅助代码
│   │   └── ratelimit/              # Redis 限流器
│   └── xerr/                       # 业务错误与错误码
├── Makefile                        # Gen 与 Swagger 生成入口
└── web/                            # 独立前端项目
```

HTTP Handler 调用应用服务，应用服务编排领域规则与仓储，MySQL 仓储负责持久化。`main.go` 组装这些依赖。

## 开发与验证

在根目录运行：

```bash
# 运行全部测试并编译各包
go test ./...

# 运行应用层、领域层测试
go test ./internal/app/... ./internal/domain/...

# 编译全部包
go build ./...
```

修改 Go 文件后执行 `gofmt -w <files>`。领域与应用测试使用内存仓储替身；仓储测试默认使用纯 Go SQLite，覆盖账号隔离、软删除、关联和事务。MySQL 兼容性测试使用临时数据库验证旧 Ent 表结构升级、数据保留及新建表结构；Redis 和微信登录仍需单独集成验证。

修改 `internal/infra/persistence/models/` 后重新生成 Query（无需连接数据库）：

```bash
go generate ./internal/infra/persistence
```

修改 HTTP 接口或 Swagger 注解后生成文档。先安装当前版本的 `swag`，并确保 Go 的二进制安装目录在 `PATH` 中：

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
go generate ./api
```

`make generate` 会依次生成 Gen Query 和 Swagger。生成结果位于 `internal/infra/persistence/query/` 和 `api/docs/`，不要手工编辑生成文件。数据库模型保留原 Ent 表名、列类型、索引和外键，仓储保留原 `MaxLen` 字节上限（例如房型名称最多 12 字节）；启动时通过 `models.AllModels()` 迁移，删除列或表等破坏性改动需单独提供版本化 SQL。

仓储与迁移的 MySQL 集成测试可显式提供 DSN。测试账号需要创建和删除数据库的权限，每个测试会创建独立临时库并在结束时删除，不修改 DSN 中的现有数据库：

```bash
RENT_MANAGER_TEST_MYSQL_DSN='root@tcp(127.0.0.1:3306)/?parseTime=true' go test ./internal/infra/persistence/mysql
RENT_MANAGER_MYSQL_MIGRATION_DSN='root@tcp(127.0.0.1:3306)/?parseTime=true' go test ./internal/infra/persistence/models -run TestMigrateEntSchema
```

## 当前限制

- 单元列表的 `monthly_rent_amount` 目前来自占位计算，单元详情的预估月收入、房间数和在租数仍有待补充。首页收入按账单月份汇总已收款账单，`monthly_growth` 表示金额差。
- 当前抄表接口为手工录入；存在导入和自动来源枚举，但尚无对应导入或自动抄表接口。
- 账单收款接口用于标记状态，尚未接入在线支付。当前没有独立的租客、合同管理接口。
- 当前 JWT 中间件校验令牌签名和有效期并提取账号 ID，但尚未限制业务接口只接受 `access` 类型令牌。调用方应使用 `access_token` 访问业务接口。
- `SuperAdmin` 和 `SYSTEM_ALLOW_ALL` 是已有常量；当前全局管理权限以配置中的账号 ID 为准，资源权限按具体代码检查。
