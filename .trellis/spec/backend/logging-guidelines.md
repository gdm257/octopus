# Logging Guidelines

> Log levels, format, what to log.

---

## Overview

统一用 `github.com/charmbracelet/log` 的包级函数（`log.Infof` / `log.Warnf` / `log.Errorf` / `log.Debugf`），printf 风格，不建 logger 实例、不结构化字段。级别由配置 `log.level` 控制，在 `cmd/start.go` 的 `PreRun` 里 `log.SetLevel` 生效。HTTP 访问日志直接复用 `gin.Logger()`（`internal/server/middleware/logger.go`）。

---

## Levels

| 级别 | 用途 | 实例 |
|------|------|------|
| `Debugf` | 任务生命周期与耗时，正常运转细节 | `log.Debugf("task %s registered with interval %v", ...)`（`internal/task/task.go`）、`log.Debugf("stats save db task finished, save time: %s", time.Since(startTime))` |
| `Infof` | 启动配置、一次性初始化 | `log.Infof("Using config file: %s", ...)`、`log.Infof("initial user: admin,password: admin")` |
| `Warnf` | 可恢复的失败、降级、重试 | `log.Warnf("failed to update price info: %v", err)`（`internal/task/init.go`） |
| `Errorf` | 操作失败且函数将放弃/返回 | `log.Errorf("stats save db error: %v", err)`、`log.Errorf("http server listen and serve error: %v", err)` |

判断标准：失败被当场吞掉继续跑用 `Warnf`；失败向上返回终止流程的，由顶层调用方 `Errorf`，不要每层重复打。

---

## Format

- 消息英文小写开头，动词起句：`"failed to xxx: %v"`；错误对象永远用 `%v`/`%w` 带上。
- **需要用户在控制台注意到的迁移提示用中文**（`log.Warnf("原 gemini 渠道已按 ... 迁移, 请手动 ...")`，`internal/db/migrate/011.go`）。
- 不打请求级日志（Gin 自带 access log）；不打敏感值（API key、凭据）。
- 定时任务惯例：开始/完成各一条 `Debugf`，完成条带 `time.Since(startTime)` 耗时（`internal/op/stats.go:43-46`、`internal/price/price.go:39-42`）。

## Anti-patterns

- 不要 `log.Fatal`（会绕过 `shutdown.Register` 的清理钩子，统计可能丢失）；顶层失败走 `Errorf` + return。
- 不要 `fmt.Println` / 标准 `log` 包。
- 不要在 op 层每次读写都打日志，缓存命中的常规路径保持静默。
