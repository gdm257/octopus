# Quality Guidelines

> Code review standards, testing requirements.

---

## Overview

代码质量底线：`go build ./...` 与 `go vet ./...` 通过；注释风格与分层约束见各专项 spec。仓库几乎不写测试，提交规范以 CONTRIBUTING.md 为准。

---

## PR Conventions (from CONTRIBUTING.md)

- 一个 PR 只含一个变更主题（一个功能或一个 bugfix），多主题拆分提交。
- **PR 不包含测试文件**——这是仓库明确的贡献约定。
- 允许 AI 辅助，但提交者必须完成人工审查。

## Testing

- 仓库现状：仅 `internal/relay/count_tokens_test.go` 一个单测（表驱动 + httptest），属于纯函数估算逻辑的回归保护。
- 新代码默认不写测试；只有当逻辑是纯函数、易回归且值得固化时才补表驱动单测，风格对齐 `count_tokens_test.go`（`cases := []struct{...}` + `t.Run`）。
- 集成/E2E 测试不写。

## Comment Style

注释**中文**，解释"为什么"而非复述代码；导出函数用 `// FuncName 干什么` 起头（godoc 惯例），关键决策在注释里给出理由：

```go
// ChannelCreate 创建渠道及其凭据, 模型与授权, 返回创建后的完整配置。
// 三者在同一事务内落库: 授权按名称引用两侧, 待凭据与模型拿到主键后由 syncChannelGrants 解析,
// 由此建一个带授权的渠道只需一趟请求。
```

（`internal/op/channel.go`）反例是"// 调用 ChannelCreate"这类复述型注释——不写。

## Review Checklist

1. 分层正确：业务在 `op`，handler 只做绑定与翻译；handler 不直接 `db.GetDB()`。
2. 写操作走缓存穿透：库与缓存同步更新，多表用事务。
3. 错误按 error-handling spec 包装返回；错误文案可作为用户可见 message。
4. 新增可编辑配置字段三处同步：`model` 实体/形状、`op` normalize 与同步、前端 `web/src/api/<resource>.ts` 类型。
5. 迁移可重跑且三库兼容。
6. 落库并暴露给前端的枚举值只增不改。
7. 不引入新依赖前优先用标准库与已有依赖（`samber/lo`、`tidwall/gjson` 等已在用）。

## Common Mistakes

- 更新实体整行 `Save`/`Updates` 覆盖了统计列——必须 `.Select(...)` 点名列（`internal/op/channel.go:104` 的注释说明了原因）。
- 全局配置热加载：`conf.AppConfig` 是启动期一次性的，运行期改配置走 `Setting` 表，别直接改 AppConfig。
- 忘记把新实体加进 `db.AutoMigrate` 列表。
- gin handler 里漏 `return` 造成 `resp.Error` 后继续执行。
