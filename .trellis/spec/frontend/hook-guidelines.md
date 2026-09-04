# Hook Guidelines

> Custom hook naming, patterns.

---

## Overview

hook 分三类落位：**API 查询 hook**（在 `web/src/api/<resource>.ts`）、**通用 DOM hook**（在 `web/src/hooks/`）、**模块私有 hook**（就近放模块目录）。命名一律 `useXxx`，导出 hook 本身。

---

## API Hooks (TanStack Query)

所有服务端数据经 React Query，hook 与类型同文件放在 `api/<resource>.ts`，模式固定（`web/src/api/channel.ts`）：

```ts
export function useChannelDetail(id?: number) {
    return useQuery({
        queryKey: ['channels', 'detail', id],
        queryFn: () => apiRequest<ChannelDetail>(`/api/v1/channel/detail/${id}`),
        enabled: id !== undefined,
        refetchOnMount: 'always',
    });
}
```

- **queryKey 规范**：`[资源复数, 方面, 参数?]`，如 `['channels', 'stats']`、`['channels', 'detail', id]`。
- **queryOptions 工厂**：被页面查询与启动预取（`app.tsx` 的 `fetchQuery`）两处共享的定义放 `api/queries.ts`；仅单页使用的可留在资源文件内导出（`channelGrantListQueryOptions`）。
- **mutation hook** 命名 `useCreateChannel` / `useUpdateChannel` / `useDeleteChannel`；`onSuccess` 里 `queryClient.invalidateQueries` 失效所有受影响的 key（渠道变更要连带失效 models 与 groups）。
- 统计类查询加 `refetchInterval: 30000`、`refetchOnMount: 'always'`；展示格式化放 `select` 里做（`channelStatsFormattedQueryOptions`）。
- 导出的每个 hook 写中文 JSDoc + `@example`（`useCreateChannel` 的示例见 `api/channel.ts:198`）。
- 有条件查询用 `enabled` 参数，不在调用方判空。

## 通用 Hooks

`hooks/` 只放与业务无关、可全局复用的 DOM/浏览器 hook（现有 `useClickOutside.tsx`：默认导出、泛型 `RefObject<T>`、事件监听在 `useEffect` 内成对注册/移除）。新通用 hook 照此风格；用第三方库前先看 `@uidotdev/usehooks` 是否已带。

## 模块私有 Hooks

只在模块内使用的 hook 就近定义在模块文件里，不导出到 `hooks/`。

---

## Examples

- 查询 + 变更全集：`web/src/api/channel.ts`
- 共享 queryOptions：`web/src/api/queries.ts`
- 通用 hook：`web/src/hooks/useClickOutside.tsx`
