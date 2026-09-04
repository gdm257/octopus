# Type Safety

> TypeScript conventions, type organization.

---

## Overview

`strict: true`，类型检查在构建里（`tsc --noEmit && vite build`）。类型与使用处同文件就近定义；跨文件共享的 API 形状放 `api/<resource>.ts`。

---

## API Types

- **字段名与后端 JSON 完全一致（snake_case 原样保留）**，不做 camelCase 转换（`ChannelDetail` 的 `base_url`、`openai_response_path`，`web/src/api/channel.ts:59`）。
- 每个类型配中文注释说明语义与约束，与后端共享的枚举位值标注"不可变更"：

```ts
// Protocol 是渠道支持的上游线协议位掩码，位值与后端 model.Protocol 一致，不可变更。
// 一条授权可同时支持多个协议，按位或组合；1 << 0 由后端保留待用。
export const Protocol = {
    OpenAIChatCompletion: 1 << 1,
    OpenAIResponse: 1 << 2,
    AnthropicMessage: 1 << 3,
} as const;
```
- 读写同构：编辑表单读取与提交共用同一类型（`ChannelDetail` 创建时 `id: 0`），不另建提交 DTO。
- 统计展示类型命名 `XxxFormatted`，由 `select` 从原始类型映射出来。
- 请求泛型：`apiRequest<T>(path, { method, body })`，响应经信封解包为 `T`。

## Type Patterns

- 数据形状用 `type`（含交叉复用 `StatsMetrics & { ... }`）；对象骨架/store 可用 `interface`。
- 字面量联合表达枚举域：`type Page = 'home' | 'channel' | ...`、`dialect: 'generic'`。
- 状态机用判别联合（`log.status === 'running' | 'committed' | 'failed' | 'canceled'` 分支）。
- 常量对象加 `as const`。
- React 类型从值推导（`React.ComponentProps<"button">`），不重复手写 HTML 属性。
- 后端可能给空数组的字段标注"恒为数组，后端承诺不为 null"，前端不做 null 兜底（`ChannelDetail.keys`）。

## Rules

- 不用 `any`；确需逃生用 `unknown` + 收窄（`catch (cause)` 后 `cause instanceof Error ? cause.message : undefined`）。
- 不开非空断言除了入口 DOM（`document.getElementById('root')!`）。
- 类型断言只在 JSON 解析边界（`JSON.parse(content) as object`）。

---

## Examples

- API 类型全集与注释风格：`web/src/api/channel.ts`
- 判别联合状态机：`web/src/components/modules/log/Item.tsx`
- store 类型：`web/src/stores/app.ts`
