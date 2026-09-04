# Component Guidelines

> Component patterns, props conventions.

---

## Overview

只写函数组件，命名导出。样式全部 Tailwind v4 工具类内联；组合类名用 `cn()`（`web/src/lib/utils.ts`，clsx + tailwind-merge）。React Compiler 已启用（`vite.config.ts` 的 babel 插件），不要手写 `useMemo`/`useCallback` 做 memoization 补偿，列表项性能敏感处仍显式 `memo`（`LogCard`，`modules/log/Item.tsx:468`）。

---

## Component Patterns

- 小组件：单文件内多个私有函数组件，导出最小集合（`modules/log/Item.tsx` 里 `LogMetrics`/`JsonContent`/`LogDetail` 私有，只导出 `LogCard`）。
- 基础组件（`ui/`）：cva 定义变体，`data-slot`/`data-variant` 属性钩子，支持 `asChild` 的用 radix `Slot`——照抄 `ui/button.tsx` 的形状，不要自创结构。
- 弹窗用 `ui/morphing-dialog.tsx` 的 MorphingDialog 家族；确认/提示用 `sonner` 的 `toast.success` / `toast.error`，错误带 `description: cause.message`。
- 图标 `lucide-react`，装饰性图标加 `aria-hidden="true"`。

## Props Conventions

- 简单 props 直接内联解构类型：`function JsonContent({ content, fallbackText }: { content: string | object | undefined; fallbackText: string })`。
- 复用形状先起名（`interface ObservedRound` / `type ChannelFormState`）再引用；数据形状优先 `type`，store/对象骨架可用 `interface`。
- 布尔变体用字面量联合：`variant: 'card' | 'footer'`。
- props 传数据 hook 结果（`log` 对象）而不是零散字段，与 API 类型对齐。

## Styling

- 暗色模式用 `dark:` 变体；语义 token（`text-muted-foreground`、`bg-card`、`border-border`、`text-destructive`）优先于具体色值。
- 图标尺寸 `size-3.5`/`size-4`，数字用 `tabular-nums`，截断 `truncate` + `min-w-0`。
- 条件类名 `cn("base", cond ? "a" : "b")`。

## i18n

所有界面文案走 `useTranslations('namespace.key')` + `locales/*.json`，三份 locale 同步新增；**不在组件里硬编码中英文文案**。带值文案 `t('retryIndex', { index: round.round })`。

## Comments

中文注释解释意图与边界，不复述代码：

```tsx
// 仅在请求进行中或弹窗打开时走秒级刷新, 避免已完成日志持续触发重渲染。
```

缩进 4 空格（`ui/` 的 shadcn 生成文件保持其原样 2 空格）。

---

## Examples

- 私有子组件拆分与 memo：`web/src/components/modules/log/Item.tsx`
- 表单组件 + 状态模块配合：`modules/channel/Form.tsx` + `modules/channel/state.ts`
- ui 变体组件范式：`web/src/components/ui/button.tsx`
