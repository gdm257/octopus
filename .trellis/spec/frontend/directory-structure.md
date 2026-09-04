# Directory Structure

> Component/page/hook organization.

---

## Overview

React 19 + TypeScript + Vite SPA，位于 `web/`，路径别名 `@/ → web/src`。构建产物输出到 `../static/out` 并嵌入 Go 二进制。**没有路由库**：页面切换由 zustand 的 `currentPage` 驱动，配合 `React.lazy` + motion 过渡（`web/src/components/app.tsx`）。

---

## Directory Layout

```
web/src/
  main.tsx                入口，Provider 组合（Theme → QueryClient → Locale → Tooltip）
  components/
    app.tsx               AppContainer：认证分支 + 页面 lazy 加载与预取
    app-shell.tsx         桌面/移动导航布局
    common/               跨模块小组件（CopyButton、IconButton、PageActions…）
    modules/<feature>/    业务模块，一个目录一个页面级功能
    ui/                   shadcn 风格基础组件（button、badge、tabs…），cva 变体
  api/                    数据层：client.ts（fetch 封装）+ 每资源一个文件 + queries.ts（共享 queryOptions）
  hooks/                  通用 DOM hook（useClickOutside）
  lib/                    纯工具（utils.ts 的 cn/格式化、model-icons、page-preload）
  provider/               theme.tsx、locale.tsx
  stores/                 zustand 全局 store（app.ts、setting.ts）
  locales/                use-intl 文案：en.json、zh_hans.json、zh_hant.json
```

---

## Module Organization

`components/modules/<feature>/` 是页面级单元（channel、group、model、log、setting、home、login…），约定：

- `index.tsx` 是模块入口，导出页面组件与顶栏操作组件（如 `Channel` / `ChannelActions`）。
- 同目录放子组件（`Card.tsx`、`Form.tsx`）、表单状态（`state.ts`）与模块工具（`utils.ts`、`probe.ts`）。
- 页面在 `app.tsx` 通过 `lib/page-preload.ts` 的 `pageImports` 懒加载注册。

新组件落位判断：只在单个模块用 → 放模块目录；跨模块复用 → `common/`；与业务无关的视觉原语 → `ui/`（优先 `pnpm dlx shadcn@latest add`，不手写）。

---

## Naming Conventions

- 组件文件 PascalCase（`Item.tsx`、`MemberStatus.tsx`）；非组件文件 camelCase（`state.ts`、`utils.ts`）。
- 模块内组件命名导出（`function Channel()`）；`ui/` 组件带 `XxxVariants` 导出。
- API 文件按资源命名（`api/channel.ts`、`api/apikey.ts`）。
- 事件与 CSS 沿用 kebab/camel 混合时以既有文件为准，别引入新风格。

---

## Examples

- 模块全貌：`web/src/components/modules/channel/`（index + Form + Card + state.ts）
- 页面注册与预取：`web/src/components/app.tsx`、`web/src/lib/page-preload.ts`
- 构建与代理配置：`web/vite.config.ts`（`/api` 代理到 `http://127.0.0.1:8080`）
