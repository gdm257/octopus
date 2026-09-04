# 实施计划: 渠道禁用阻断分组选路

## 执行清单(按序)

1. **op 判定与校验** — `internal/op/channel.go`
   - 新增 `ChannelGrantDisabled(id int) bool`(签名与分支见 design.md 改动一)。
   - `ChannelGrantGet` 补 `channel.Enabled` 校验, 错误文案 `channel %d is disabled`; 同步更新其 docstring(渠道禁用与凭据停用并列)。
2. **选路过滤** — `internal/relay/route.go`
   - `pickGroupItem` 增加 `skip func(model.GroupItem) bool` 参数; 手动分支命中指定成员且被跳过时返回零值; 故障转移遍历以 skip 开头 `continue`(先于冷却与探测判断)。
   - 重写 `pickGroupItem` 顶部过时注释(现描述"渠道禁用由调用方发现并上报失败", 与实现不符)。
3. **终态拒绝** — `internal/relay/handler.go`
   - 选路处构造 `skipDisabled` 闭包(调 `op.ChannelGrantDisabled`); 零值命中 `allItemsSkipped` 时 `markFailed` + `rejectRequest`, 错误文案 `model %q unavailable: all channels are disabled`。
   - 新增 `allItemsSkipped(group, skip) bool`(模式感知, 见 design.md 改动三)。
4. **单元测试**
   - `internal/relay/route_test.go`: pick 谓词三分支 + `allItemsSkipped` 矩阵(手动指定成员/故障转移全跳/空组 false)。
   - `internal/op/channel_disabled_test.go`(或并入既有命名习惯): 缓存 Set 构造引用链, 覆盖 `ChannelGrantDisabled` 与 `ChannelGrantGet` 渠道禁用分支。
5. **注释一致性巡检**: `internal/model/channel.go:57` 注释("禁用后不参与选路但保留统计")实现后变真, 保留; 确认无其他与渠道禁用矛盾的注释。

## 验证命令

```bash
go build ./...
go vet ./...
go test ./...
```

## 冒烟验证(手动, 覆盖 AC1/AC2/AC3/AC6)

1. 启动服务, 建渠道 A(启用)与渠道 B, 建分组引用 A、B 各一授权(故障转移, A 优先)。
2. 请求 `/v1/chat/completions` → 由 A 承接; 管理界面禁用 A → 再次请求由 B 承接(AC1)。
3. 禁用 B → 请求立即返回 400 协议错误体, message 含 `all channels are disabled`(AC2)。
4. 切手动模式指定 B 的成员 → 同样立即 400, 不回退 A(AC3)。
5. 重新启用 B → 请求即刻恢复(AC6)。

AC4/AC5 由单元测试覆盖(空组等待、凭据停用不报错); AC7 由代码审查确认(`GroupListModel` 不受本改动影响)。

## 风险文件与回滚点

- `internal/relay/route.go`: `pickGroupItem` 签名变更, 唯一调用点在 `internal/relay/handler.go:99`, 编译期兜底。
- `internal/relay/handler.go`: 选路循环是核心路径, 改动限于零值分支内, 不触碰上游调用与流式提交。
- 回滚: 整任务单 commit, revert 即净。

## 启动前检查

- [ ] prd.md / design.md / implement.md 经用户评审
- [ ] `uvx trellis-runtime task start 09-04-channel-disable-routing` 之后才动产品代码
