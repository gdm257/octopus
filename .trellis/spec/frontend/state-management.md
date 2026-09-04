# State Management

> State library, patterns, what goes where.

---

## Overview

状态按归属三分：**服务端数据全部走 TanStack Query**（见 hook-guidelines）；**跨页面 UI 状态用 zustand**（`web/src/stores/`）；**表单等局部状态用 useState + 模块内 state.ts**。没有 Redux、没有数据 Context。

---

## Server State

唯一入口 `apiRequest`（`web/src/api/client.ts`）。全局默认：`staleTime: 60_000`、`refetchOnWindowFocus: false`。401 由 client 统一派发 `api:unauthorized` 事件，页面不各自处理。

## Global UI State (zustand)

`stores/` 按域一个文件（`app.ts` 页面导航、`setting.ts`），模式：

```ts
export const useAppStore = create<AppState>()(
    persist(
        (set, get) => ({ ... }),
        { name: 'nav-storage' }
    )
);
```

- 需要跨会话保留的用 `persist` 中间件并起存储名；纯会话内状态不用。
- store 只放真正全局的少量字段（当前页、主题偏好），页面内部状态**不放**。
- 组件里按字段订阅：`useAppStore((state) => state.currentPage)`，避免整 store 订阅引起重渲染。
- 派生动作写在 store 的 action 里（`setCurrentPage` 里计算 direction），不在组件里算。

## Local / Form State

复杂表单状态抽到模块 `state.ts`（`web/src/components/modules/channel/state.ts`）：定义 `XxxFormState` 类型 + `emptyFormState` 常量 + `fromXxx`（读侧还原）/`toXxx`（提交转换）纯函数，组件用 `useState` 持有。转换函数保持纯，方便提交与探测（`toChannelConfig`）共用。简单状态就地 `useState`。

## Anti-patterns

- 不把 Query 缓存复制进 zustand（两份真相）。
- 不用 Context 传服务端数据；Provider 只做主题/语言/Tooltip（`main.tsx`）。
- 弹窗内重状态（如日志详情）靠"仅打开时挂载子组件"控制（`LogDetail` 的注释），不靠全局状态。
