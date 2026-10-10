# 租序 · 租赁管理工作台

基于当前 Go 后端的 Vue 3 + TypeScript 租赁工作台。使用 Element Plus、系统字体、浅色分层布局和适量透明材质；支持桌面与移动端、键盘操作及减少动态效果偏好。

## 启动与验证

- Node.js 20.19+（20.x）或 >= 22.12.0；运行原生 TypeScript 测试建议 Node.js 24。
- pnpm 9.15.4（由 `package.json` 的 `packageManager` 字段固定版本）。

```bash
cd rent-manager/web
corepack enable
pnpm install --frozen-lockfile
pnpm dev
pnpm typecheck
pnpm test
pnpm build
pnpm preview
```

开发时 `/api` 默认代理到 `http://127.0.0.1:8080`。可通过 `.env.local` 的 `API_PROXY_TARGET` 修改代理目标；前端默认使用同源 `/api`，部署到其他 API 地址时设置 `VITE_API_BASE_URL`。

## 嵌入 Go 二进制

在仓库根目录执行：

```bash
make build
./rent-manager-server --configFile ./config.local
```

`make build` 会安装锁定版本的前端依赖、构建 `web/dist`，再通过 Go `embed` 将产物编入二进制。访问 [http://localhost:8080/](http://localhost:8080/) 即可打开前端，同源 `/api` 由 Go 服务处理。Go 静态资源 Handler 支持 history 路由回退，直接访问或刷新 `/login`、`/dashboard` 等页面会返回 `index.html`。

仅构建前端时，在仓库根目录运行 `make build-web`。`web/dist` 不提交到版本库，因此首次执行 `go run`、`go test` 或 `go build` 前需要先完成这一步；前端修改后需重新构建静态资源和 Go 二进制。

部署只需二进制及原有服务配置，并准备 MySQL、Redis（邮箱功能还需要 SMTP），无需额外部署静态资源或安装 Node.js。配置说明见 [根目录 README](../README.md)。

## 登录

仅支持邮箱注册与登录。`/register` 需要邮箱验证码和 8–128 字符的密码，注册成功后自动进入工作台；`/login` 可选择密码或邮箱验证码登录已有账号。验证码发出后显示重发倒计时，密码与验证码只保留在页面内存中。所有会话通过 `/auth/session` 验证，令牌仅保存在当前标签页的 `sessionStorage`。退出登录会调用服务端撤销会话。

接口为 `POST /auth/email/code`、`POST /auth/register`、`POST /auth/password-login`、`POST /auth/email-login`。请先配置后端 `auth.smtp`，否则发码会明确提示不可用；前端不模拟发送成功。

## 已接入功能

| 页面 | 功能与接口 |
| --- | --- |
| 经营概览 | `/home` 的收入、房源/房间汇总与待办；真实房源列表 |
| 房源管理 | `/units` 的自有/获授权楼栋、搜索、新增、编辑、删除、默认收费项目、抄表日 |
| 房间管理 | 单元下房间列表与分页、房型/出租状态筛选、创建、编辑、删除、收费项目、同所有者楼栋迁移 |
| 账单管理 | 单元账单按月/状态筛选与分页、单房或整栋生成、查看明细、确认收款、取消 |
| 抄表管理 | 按月查看单元汇总、缺漏筛选、房间读数录入、差值或直接计费 |
| 角色与授权 | 角色管理、权限组增量绑定/解绑、单元授权与撤销、全局权限目录及管理员写操作 |
| 设置 | 计费项新增/修改、房型管理、账号键值配置 |

租客档案、合同、历史报表、备份恢复尚无对应后端接口；保留页面入口并显示暂未开放，不生成模拟业务数据或假保存结果。

## 契约与边界

- API 层统一处理 `{ code, data, message }` 响应，包括 HTTP 200 中的业务失败。仅认证失效清除会话，权限不足保留登录态。
- authkit 账号字段为 `ID`、`Username`、`Email`，账号 ID 是 64 位小写十六进制字符串。
- 房间 `is_active=true` 表示空置，`false` 表示已出租。金额字段按分传输、按元呈现；计费项单价可含小数分。
- 房源列表的 `monthly_rent_amount` 当前由后端占位计算，前端不展示它。概览收入来自 `/home` 的已收款账单汇总。
- 后端没有当前超级管理员或有效权限查询接口。普通用户可以浏览目录；全局写操作入口明确标注仅限超级管理员，权限最终由后端校验。
- 角色与权限组展示当前已绑定的清单；“添加 / 移除”仅选择本次变更项，成功后刷新绑定关系。
- 房间表单按当前楼栋请求房型与收费项目目录，获授权账号需具备该楼栋的对应查看权限。共享定义仍由所有者维护，下级可在授权楼栋选用。
- 楼栋列表按权限省略详细资料与统计；页面保留楼栋入口，未授权字段显示为不可查看，不作为零值统计。
- 账号配置保留后端键值契约，不将任意键冒充已生效的提醒或自动化开关。
- 分页列表无总数；使用上一页/下一页，不把本页数据计为全量统计。

## 目录

```text
src/
  api/                  # 请求边界和按业务划分的接口
  components/
    common/             # 图标、加载/空/失败状态、标题
    layout/             # 共享导航与工作台布局
    units/ rooms/ ...   # 表单、列表和业务面板
  routers/index.ts      # 唯一路由实例与会话守卫
  stores/               # 当前会话、房源列表与选择
  types/                # 真实后端 DTO
  utils/                # 格式转换、输入校验
  views/                # 每个页面路由独立的 View
  style.css             # 共享设计变量与基础样式
```

`tests/*.test.mjs` 使用 Node 原生测试运行器覆盖请求错误、会话边界和业务输入校验。浏览器验证使用独立接口夹具，正式应用始终请求真实 API；没有自动启用的演示账号或模拟数据回退。真实邮箱投递与数据库写入仍需使用配置好的后端和有效会话进行端到端验收。
