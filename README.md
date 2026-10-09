# Rent Manager

租赁管理系统后端，基于 Go、Gin 和 GORM + Gen，使用 MySQL 存储业务数据、Redis 执行接口限流。当前提供微信登录、房源与房间管理、月度抄表、租金账单、首页统计以及按房源授权的 RBAC 接口。

本文说明根目录的 Go 后端。`web/` 为独立的 Vue 前端项目，说明见 [web/README.md](web/README.md)。

## 已实现功能

| 模块 | 当前能力 |
| --- | --- |
| 认证 | 通过 authkit 使用微信 `code` 登录，首次登录创建 authkit 账号，查询与撤销会话 |
| 账号与设置 | 通过会话获取 authkit 账号，读取和更新本项目的账号配置 |
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
| authkit | `github.com/miebyte/authkit v1.9.0`，账号、微信登录与可撤销的会话令牌 |
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

服务启动时先迁移 authkit 的认证表，再通过 GORM AutoMigrate 创建或更新业务表；数据库账号需要相应权限。启动过程不负责创建数据库，也不自动初始化权限、权限组或角色。

新部署可使用空数据库。对于旧数据库，启动检查一旦发现原 `accounts` 表就会停止，避免把整型账号引用自动转换成无效的 authkit 身份。[账号迁移 SQL](migrations/20261009_authkit_accounts.sql) 仅适用于旧账号及相关业务表均为空的情况，会拒绝非空数据；请先检查脚本再按需手动执行。需要保留历史数据时，必须另行制定数据迁移方案，本项目不建立账号映射。

### 2. 配置服务

在项目根目录创建 `config.local`，内容为 JSON。该文件名匹配现有 `.gitignore` 中的 `*.local`，适合保存本地配置。可复制 [config.example.json](config.example.json)，替换微信配置和数据库账号后使用：

```json
{
  "auth": {
    "wechat": {
      "app_id": "your-wechat-app-id",
      "secret": "your-wechat-app-secret"
    },
    "session_ttl": "720h"
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
| `auth.wechat` | `app_id`、`secret` 均不能为空，不能包含首尾空白 |
| `auth.session_ttl` | 会话有效时长，支持 `720h` 等字符串；零值默认 30 天，不允许负值 |
| `admin.SuperAdminAccountIDs` | 允许写入全局权限和权限组的 authkit 账号 ID 字符串列表，ID 为 64 位小写十六进制；空列表表示无人拥有该管理权限 |
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

### 登录与会话

微信登录无需令牌。将客户端获取的微信 `code` 发送到：

```bash
curl -X POST http://localhost:8080/api/auth/wechat-login \
  -H 'Content-Type: application/json' \
  -d '{"code":"your-wechat-login-code"}'
```

成功响应示例：

```json
{
  "code": 0,
  "data": {
    "account": {
      "ID": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
      "Username": "",
      "Email": ""
    },
    "access_token": "...",
    "expires_at": "2026-11-08T00:00:00Z",
    "created": true
  },
  "message": "success"
}
```

`account` 直接返回 authkit 的账号对象，字段名为 `ID`、`Username`、`Email`；`created` 表示本次登录是否创建了账号。`access_token` 是 authkit 签发的会话令牌，客户端应原样保存并以 `Authorization: Bearer <access_token>` 传入。会话过期后重新进行微信登录。

查询当前会话，响应 `data` 包含 `account` 和 `expires_at`：

```bash
curl http://localhost:8080/api/auth/session \
  -H 'Authorization: Bearer your-access-token'
```

退出登录并撤销当前会话：

```bash
curl -X DELETE http://localhost:8080/api/auth/session \
  -H 'Authorization: Bearer your-access-token'
```

旧的 `/auth/refresh-token` 和 `/accout/self` 接口已移除。账号由 authkit 统一管理，本项目不保留独立账号模型或账号映射。

### 接口概览

下表路径均以 `/api` 为前缀。`POST /auth/wechat-login` 公开，其余接口均校验 authkit 会话令牌。`:id` 等表示路径参数，详细请求和响应字段见 Swagger 及 `internal/app/<业务>/dto.go`。

| 方法 | 路径 | 用途 |
| --- | --- | --- |
| POST | `/auth/wechat-login` | 微信登录 |
| GET / DELETE | `/auth/session` | 查询当前账号与会话有效期、撤销当前会话 |
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

资源以 authkit 账号为归属边界。所有 `account_id`、`granted_by_account_id` 和可空的 `reader_id` 都直接使用 authkit 账号 ID（64 位小写十六进制字符串），单元、角色、房间等业务资源 ID 仍为整数。单元、角色、房型、抄表项和账号设置归属账号，房间归属单元，抄表记录和账单归属房间。

- 账号可以管理自己拥有的资源；删除单元及管理单元授权仅限所有者。
- 访问他人单元的业务数据时，服务检查该单元下的授权角色和权限代码。
- 角色属于创建它的账号，可以绑定全局权限组；权限组关联全局权限。
- 普通登录账号可以查询全局权限与权限组，写入操作要求账号 ID 位于 `admin.SuperAdminAccountIDs` 中。

授权关系为：

```text
被授权账号 → UnitAccountRole → 房源所有者的 Role → PermissionGroup → Permission
```

授予单元角色前，应用服务通过 authkit 的账号仓储校验目标账号是否存在。共享房源时，先创建权限与权限组，所有者创建角色并绑定权限组，再将该角色授予目标账号在指定单元使用。权限代码在 [internal/domain/rbac/permissions.go](internal/domain/rbac/permissions.go) 定义，例如 `unit.view`、`room.manage`、`meter_reading.manage`、`rent_bill.manage`。资源权限校验位于业务服务中，按代码匹配。

## 项目结构

```text
.
├── main.go                         # 配置解析、依赖组装、迁移与服务启动
├── config/                         # authkit 微信登录、会话与管理员配置
├── api/
│   ├── api.go                      # 路由与 Swagger 入口
│   ├── routers/                    # HTTP Handler 与 Swagger 注解
│   ├── middlewares/                # authkit 会话校验、Redis 限流
│   ├── response/                   # 统一响应与错误处理
│   └── docs/                       # 生成的 Swagger 文档
├── internal/
│   ├── app/                        # 按业务拆分的用例服务、DTO 与测试
│   ├── domain/                     # 领域模型、业务规则与仓储接口
│   ├── infra/
│   │   ├── persistence/
│   │   │   ├── mysql/              # GORM + Gen 仓储实现，query/ 保存生成查询
│   │   │   └── models/             # 业务持久化模型与 AllModels 迁移清单
│   │   ├── cache/                  # 缓存辅助代码
│   │   └── ratelimit/              # Redis 限流器
│   └── xerr/                       # 业务错误与错误码
├── cmd/generate/                   # 从业务模型生成 Gen Query
├── Makefile                        # Gen 与 Swagger 生成入口
└── web/                            # 独立前端项目
```

认证 Handler 直接接入 authkit；业务 HTTP Handler 调用应用服务，应用服务编排领域规则与仓储，MySQL 仓储负责业务持久化。`main.go` 组装这些依赖。

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

修改 Go 文件后执行 `gofmt -w <files>`。领域与应用测试使用内存仓储替身；仓储测试默认使用纯 Go SQLite，覆盖账号隔离、软删除、关联和事务。MySQL 仓储测试可使用临时数据库验证实际数据库行为；Redis 和微信登录仍需单独集成验证。

修改 `internal/infra/persistence/models/` 后重新生成 Query（无需连接数据库）：

```bash
go generate ./cmd/generate
```

修改 HTTP 接口或 Swagger 注解后生成文档。先安装当前版本的 `swag`，并确保 Go 的二进制安装目录在 `PATH` 中：

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.6
go generate ./api
```

`make generate` 会依次生成 Gen Query 和 Swagger。生成结果位于 `internal/infra/persistence/mysql/query/` 和 `api/docs/`，不要手工编辑生成文件。业务模型的账号引用使用 `varchar(64)` 和二进制排序规则，抄表人的 `reader_id` 可空；不再关联本项目原 `accounts` 表。authkit 的表由 `authmysql.Models()` 提供，业务表由 `models.AllModels()` 提供。旧整型账号及其业务引用不能通过简单修改列类型变为 authkit 身份；已有数据库需要另行处理历史数据，删除表等破坏性改动需单独提供版本化 SQL。

仓储的 MySQL 集成测试可显式提供 DSN。测试账号需要创建和删除数据库的权限，每个测试会创建独立临时库并在结束时删除，不修改 DSN 中的现有数据库：

```bash
RENT_MANAGER_TEST_MYSQL_DSN='root@tcp(127.0.0.1:3306)/?parseTime=true' go test ./internal/infra/persistence/mysql
```

## 当前限制

- 单元列表的 `monthly_rent_amount` 目前来自占位计算，单元详情的预估月收入、房间数和在租数仍有待补充。首页收入按账单月份汇总已收款账单，`monthly_growth` 表示金额差。
- 当前抄表接口为手工录入；存在导入和自动来源枚举，但尚无对应导入或自动抄表接口。
- 账单收款接口用于标记状态，尚未接入在线支付。当前没有独立的租客、合同管理接口。
- `SuperAdmin` 和 `SYSTEM_ALLOW_ALL` 是已有常量；当前全局管理权限以配置中的账号 ID 为准，资源权限按具体代码检查。
