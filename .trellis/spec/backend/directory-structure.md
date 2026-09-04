# Directory Structure

> How backend code is organized in this project.

---

## Overview

Go 1.26 单二进制应用，module 路径 `github.com/bestruirui/octopus`。前端构建产物嵌入二进制（`static/out`），入口 `main.go` → `cmd/`（cobra CLI）→ `internal/`。

核心分层（自上而下，只允许上层依赖下层）：

```
cmd (cobra 命令) / main.go
  → internal/server (Gin 路由、handlers、middleware、resp)
    → internal/op (业务逻辑 + 进程内缓存，数据访问唯一入口)
      → internal/model (GORM 实体与 JSON 形状定义)
        → internal/db (GORM 初始化 + migrate)
```

`internal/relay` 是 LLM 转发核心，直接依赖 `internal/op` 与 `internal/model`；`internal/task`、`internal/price`、`internal/update` 是后台任务；`internal/rhttp` 提供共享 HTTP client；`internal/utils/*` 是纯工具包。

---

## Directory Layout

```
cmd/                      cobra 命令（root / start / version）
internal/
  conf/                   viper 配置（config.go、AppConfig 全局变量）
  db/                     GORM 初始化（db.go）
    migrate/              版本化迁移（001-012.go，编号递增）
  model/                  GORM 实体 + API 请求/响应形状
  op/                     业务操作层，带进程内缓存（op/cache.go）
  price/                  模型价格同步
  relay/                  LLM 转发（handler、route、upstream、protocol）
  rhttp/                  共享 http.Client（直连/代理）
  server/
    auth/                 JWT 签发与校验
    handlers/             每资源一个文件，init() 注册路由
    middleware/           Auth / APIKeyAuth / RequireJSON / CORS / Logger
    resp/                 统一响应封装（resp.go）与错误文案（error.go）
    router/               路由注册器（NewGroupRouter / NewRoute）
  task/                   定时任务注册与调度
  update/                 自更新
  utils/                  cache / diff / shutdown / xstrings
```

---

## Module Organization

**按资源竖切**：同一资源在每层都用同名文件，例如 channel：

- `internal/model/channel.go` — 实体与形状
- `internal/op/channel.go` — 业务与缓存
- `internal/server/handlers/channel.go` — 路由与 handler

新增资源的完整路径：model 定义 → op 实现 → handlers 注册。不要把业务逻辑写进 handler，也不要让 handler 直接查 `db.GetDB()`——一切数据读写走 `internal/op`。

**注册靠 `init()`**：路由、迁移、定时任务都在各自文件的 `init()` 里向全局注册表登记（见 `internal/server/handlers/channel.go:25`、`internal/db/migrate/012.go:11`、`internal/task/init.go`）。新增文件放进对应目录即自动生效，无需改中心列表。

**启动顺序**固定在 `cmd/start.go`：Load 配置 → InitDB → op.InitCache → UserInit → server.Start → task.Init/RUN，并通过 `shutdown.Register` 注册逆序关闭钩子。

---

## Naming Conventions

- 包名：单词、全小写（`conf`、`op`、`resp`、`rhttp`）。
- 文件名：资源名单数小写（`channel.go`、`apikey.go`）；迁移文件用三位数字前缀（`013.go`）。
- handler 函数：小写驼峰、包内私有（`createChannel`、`getChannelDetail`）。
- op 函数：`资源 + 动作`、导出（`ChannelCreate`、`ChannelDetailGet`、`LLMBatchCreate`）。
- 模型：`Channel`、`ChannelKey`、`ChannelGrant`；可编辑子集命名为 `XxxConfig`，读写副本命名为 `XxxDetail`。
- 常量与协议位：`ProtocolOpenAIResponse`、`DialectGeneric`（见 `internal/model/channel.go`）。

---

## Examples

- 完整竖切参考：`internal/model/channel.go` + `internal/op/channel.go` + `internal/server/handlers/channel.go`
- 路由注册器：`internal/server/router/router.go`
- 后台任务：`internal/task/task.go`
