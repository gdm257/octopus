# Error Handling

> How errors are caught, logged, and returned.

---

## Overview

统一响应封装在 `internal/server/resp/`：所有管理端接口返回 `{code, message, data}` 信封（`resp/resp.go`），通用错误文案集中在 `resp/error.go`。错误以普通 `error` 值逐层向上返回，不 panic、不用自定义错误类型树。

---

## Response Envelope

```go
resp.Success(c, data)                          // 200 + {code:200, message:"success", data:...}
resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON) // {code:400, message:...}
```

- 成功无数据时 `resp.Success(c, nil)`。
- `Error` 内部已 `AbortWithStatusJSON`，handler 里 `return` 即可，不需要额外 `c.Abort()`。

## 错误文案

通用文案用 `resp/error.go` 里的常量（`ErrBadRequest`、`ErrInvalidJSON`、`ErrInvalidParam`、`ErrResourceNotFound`、`ErrUnauthorized`、`ErrDatabase` 等）；资源特定校验失败直接把 op 层返回的错误文本给出去（`err.Error()`），如 `"channel name is required"`。文案面向终端用户，保持英文小写句式。

## Handler 模式

handler 只做绑定、调用 op、翻译错误码（`internal/server/handlers/channel.go`）：

```go
var req model.ChannelDetail
if err := c.ShouldBindJSON(&req); err != nil {
    resp.Error(c, http.StatusBadRequest, resp.ErrInvalidJSON)
    return
}
channel, err := op.ChannelCreate(&req, c.Request.Context())
if err != nil {
    resp.Error(c, http.StatusInternalServerError, err.Error())
    return
}
resp.Success(c, channel)
```

状态码约定：

- JSON 绑定失败、参数非法 → `400` + `ErrInvalidJSON` / `ErrInvalidParam`
- 未认证 → `401` + `ErrUnauthorized`（middleware 已处理，见 `middleware/auth.go`）
- 资源不存在 → `404`
- 上游/网关类失败 → `502`，带上游原文（`fetchModel` 中 `fmt.Sprintf("openai: %v; anthropic: %v", ...)`）
- 其余 op 层错误 → `500` + `err.Error()`

## Error Wrapping

- 错误消息全小写开头，动作开头式：`fmt.Errorf("failed to create channel: %w", err)`，用 `%w` 保留链。
- op 层返回的校验类错误（`fmt.Errorf("channel not found")`）不带包装，直接作为用户可见文案。
- middleware 拒绝请求：`resp.Error(...)` 后必须 `c.Abort()`。

## Relay 链路例外

`internal/relay` 面向 LLM 客户端，错误格式必须遵循 OpenAI/Anthropic 协议错误体，不走 `resp` 信封（见 `internal/relay/handler.go`）；上游错误要拦截转换（Upstream Error Shielding），不能把网关内部错误裸传给 agent。

---

## Examples

- 信封与常量：`internal/server/resp/resp.go`、`internal/server/resp/error.go`
- handler 错误翻译全集：`internal/server/handlers/channel.go`
- middleware 拒绝 + Abort：`internal/server/middleware/auth.go`
