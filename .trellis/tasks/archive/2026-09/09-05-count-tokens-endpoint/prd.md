# count_tokens 本地估算端点

## Goal

提供 `POST /v1/messages/count_tokens` 端点, 在本地估算 Anthropic Messages 请求的输入 token 数并按协议返回 `{"input_tokens": N}`, 不选择渠道、不请求上游、不产生真实用量。

背景: Claude Code 等客户端依赖该端点统计上下文占用; 端点缺失时客户端退化为对每个上下文条目各发送一次 `max_tokens=1` 的真实推理请求, 放大延迟与费用。

## Background

- 补录说明: 该需求的实现已由 commit 3c885529 完成(`internal/relay/count_tokens.go` + 测试 + 路由注册), 本任务为补建 planning 记录, 无待写代码。
- 路由注册于 `internal/server/handlers/relay.go` 的 init, 紧邻既有 `/messages` 路由。
- 估算规则(见 `internal/relay/count_tokens.go`):
  - 递归累计请求正文中全部字符串值, 按 3.5 字符/token 换算(官方近似 3.5~4, 取下界使估算略偏高), 向上取整。
  - JSON 键、数字、布尔值不参与统计。
  - `type` 为 `image` / `document` 的内容块按固定 1600 token 估算, base64 数据不计入。
- 不校验模型对应的分组是否存在(本地估算与渠道无关)。

## Requirements

- R1 端点接受 Anthropic Messages 请求体, 返回 `{"input_tokens": N}`。
- R2 非法 JSON 请求体返回 400。
- R3 估算纯本地完成, 不触碰渠道选择与上游用量。

## Acceptance Criteria

- [x] AC1 合法请求体返回 200 与 `input_tokens` 字段(`internal/relay/count_tokens_test.go::TestCountTokensHandler`)。
- [x] AC2 文本内容按字符密度换算, 空对象计 0, 数字/布尔不计(`TestEstimateTokens`)。
- [x] AC3 image / document 块的 base64 数据不计入, 按固定成本估算(`TestEstimateTokens`)。
- [x] AC4 非法 JSON 返回 400(`TestCountTokensHandler`)。

## Out of Scope

- 不做精确 tokenizer 级计数(字符密度近似即可)。
- 不校验模型名/分组存在性。
- 不支持非 Anthropic 格式的请求体。
