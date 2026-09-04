# Quality Guidelines

> Linting, testing, accessibility.

---

## Overview

提交前两条硬线：`pnpm lint`（ESLint 10 flat config）与 `pnpm build`（含 `tsc --noEmit` 类型检查）必须通过。前端不写测试（与仓库"PR 不含测试文件"的贡献约定一致）。

## Lint

配置在 `web/eslint.config.mjs`：`@eslint/js` recommended + `typescript-eslint` recommended + `react-hooks` flat recommended + `react-refresh` vite 预设。

既有豁免不要动：`react-refresh/only-export-components` 的 `allowExportNames`（`buttonVariants`、`useMorphingDialog` 等变体与共享 store 同文件导出是项目约定）。新文件触发该规则时优先拆文件，确实属于既有约定形态的再加豁免名。

## React Compiler

`babel-plugin-react-compiler` 已启用，组件自动 memo 化。因此：

- 遵守 React 规则（hook 顶层调用、不条件调用），编译器才能生效。
- 不为性能手写 `useMemo`/`useCallback`/`memo` 包裹；确有列表级重渲染压力才显式 `memo`（先例：`LogCard`）。

## Accessibility

- 装饰性图标 `aria-hidden="true"`；可按压元素用 `type="button"`。
- 状态用语义属性：选中项 `aria-pressed`，禁用态 `disabled` + `disabled:` 样式变体。
- 图标配色与文字对比走语义 token（`text-muted-foreground`），暗色模式必须可用（`dark:` 变体）。

## i18n 完整性

新增文案三份 locale 同步更新（`en.json`、`zh_hans.json`、`zh_hant.json`），漏一份就是缺陷。文案 key 按页面/模块分命名空间。

## Performance 约定

- 页面懒加载经 `lib/page-preload.ts`，新页面必须注册进去。
- 长列表用 `@tanstack/react-virtual`（`common/VirtualizedGrid.tsx`）。
- 高频刷新（统计、进行中日志）控制刷新范围：只在需要时启动 interval（`log/Item.tsx` 的秒级刷新注释）。

## Common Mistakes

- 在组件里硬编码中文/英文文案而没走 `useTranslations`。
- mutation 成功后漏失效关联 queryKey（渠道变更要连带 groups/models）。
- 引入新 UI 库前未检查 `ui/` 与 radix 是否已覆盖。
- API 类型字段自行 camelCase 化，与后端 snake_case JSON 脱节。
