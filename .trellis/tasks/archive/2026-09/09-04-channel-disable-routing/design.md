# 技术设计: 渠道禁用阻断分组选路

## 边界与不变量

改动收敛在两条既有生产链路上, 不新增生产文件、不改持久化与 API 形状:

- `internal/op`: 渠道状态缓存本就是可用性的唯一事实源(`groupSnapshot` 与 `ChannelGrantCandidates` 同口径, 见 `internal/op/group.go:292`、`internal/op/channel.go:334`), 渠道禁用判定也放这里。
- `internal/relay`: 选路过滤与终态拒绝放这里, 复用"请求级失败"的既有出口(`markFailed` + `rejectRequest`, 见 `internal/relay/handler.go:129-133`)。

核心不变量: **渠道禁用是选路期过滤 + 全员耗尽时立即终态; 凭据停用与引用缺失维持"选路后由 `ChannelGrantGet` 发现并等待"的既有语义, 两者不合并判定。**

## 改动一: op 层渠道禁用判定

`internal/op/channel.go` 新增:

```go
// ChannelGrantDisabled 判断授权所属渠道是否已被禁用。
// 仅供选路过滤: 凭据停用与引用缺失不在判断之列, 它们维持由转发轮次自行发现并等待的既有语义;
// 引用链不完整时返回 false, 把成员留给既有路径处理。
func ChannelGrantDisabled(id int) bool {
    grant, ok := channelGrantCache.Get(id)
    if !ok {
        return false
    }
    channelModel, ok := channelModelCache.Get(grant.ChannelModelID)
    if !ok {
        return false
    }
    channel, ok := channelCache.Get(channelModel.ChannelID)
    if !ok {
        return false
    }
    return !channel.Enabled
}
```

`ChannelGrantGet`(`internal/op/channel.go:291`)在凭据校验旁并列补渠道校验, 兑现其"必然可直接转发"的文档承诺, 并兜住选路与取授权之间的竞态:

```go
if !channel.Enabled {
    return model.ChannelGrant{}, fmt.Errorf("channel %d is disabled", channel.ID)
}
```

## 改动二: 选路过滤

`pickGroupItem`(`internal/relay/route.go:68`)增加跳过谓词参数, 调用点唯一(`internal/relay/handler.go:99`):

```go
func pickGroupItem(group model.Group, skip func(model.GroupItem) bool) model.GroupItem
```

- 手动分支: 命中 `ActiveItemID` 且 `skip(item)` 时返回零值——手动模式不回退其他成员。
- 故障转移分支: 遍历开头 `skip(item)` 即 `continue`, 置于冷却与探测判断之前——被禁渠道的成员不得占用 `CurrentItemID`/`ProbeItemID` 槽位。

同步重写过时注释(`internal/relay/route.go:66-67`): 原注释称"渠道禁用由调用方发现并作为一轮失败上报", 与实现不符, 改为描述新语义。

## 改动三: 全员耗尽的立即终态

`internal/relay/handler.go` 选路处:

```go
skipDisabled := func(item model.GroupItem) bool { return op.ChannelGrantDisabled(item.ChannelGrantID) }
item := pickGroupItem(group, skipDisabled)
if item.ID == 0 {
    if allItemsSkipped(group, skipDisabled) {
        err := fmt.Errorf("model %q unavailable: all channels are disabled", metadata.Model)
        request.markFailed(err, "", nil)
        rejectRequest(c, inbound, err)
        return
    }
    // 既有等待路径不变
}
```

`allItemsSkipped` 为 relay 内的小函数, 模式感知:

- 手动模式: 仅看 `ActiveItemID` 指向的成员(手动模式的可选集就是它)。
- 故障转移模式: 要求 `len(group.Items) > 0` 且全部成员被 skip。

由此覆盖需求矩阵: 全员被禁→立即报错; 空组→等待; 全员冷却或部分可用→等待; 部分被禁+其余凭据停用→被禁者被过滤后按既有语义等待剩余成员。

## 改动四: 复用既有错误出口

`rejectRequest`(`internal/relay/handler.go:352`)固定返回 `400 + invalid_request_error`，全员渠道禁用复用这一出口:

```go
err := fmt.Errorf("model %q unavailable: all channels are disabled", metadata.Model)
request.markFailed(err, "", nil)
rejectRequest(c, inbound, err)
return
```

请求本身虽合法，但该错误与现有 `model not found` 等无法满足的选路请求保持同一 4XX 口径，无需新增错误响应函数或改动既有四处调用。错误文案仍明确区分渠道禁用。

## 数据流

```
请求 → GroupGetByName(原始成员) → pickGroupItem(skip=ChannelGrantDisabled)
     ├─ 选到成员 → ChannelGrantGet(现含渠道校验, 竞态兜底) → 上游
     ├─ 零值 + allItemsSkipped → markFailed + 400 invalid_request_error 终态
     └─ 零值 + 其余原因 → wait(MemberRetryIntervalSeconds) 重试(现状)
```

禁用/启用经 `ChannelEnabled` 同步写渠道缓存(`internal/op/channel.go:223-224`), 选路每轮实时读缓存, 故即时生效, 无需触碰分组缓存。

## 权衡

- **不在 `GroupGetByName` 补可用性**: 其文档明确"不补齐成员的展示字段"(`internal/op/group.go:57`), 且选路只需布尔判定, 补名称是浪费; 单独的轻量判定函数零侵入。
- **谓词只判渠道禁用、不判引用缺失/凭据停用**: 把"不可选"合并判定会顺手改变空组/凭据停用的既有等待语义, 违反 R3; 判定范围最小化是范围控制, 不是疏漏。
- **4XX 客户端可能按错误策略处理**: 该错误与 `model not found` 等选路请求错误保持一致, 比将管理员禁用动作伪装为 5XX 更符合当前 Relay 错误出口与既有语义。
- **前端零改动**: `Available` 口径 `channel.Enabled && channelKey.Enabled` 已存在, 语义自动变准。

## 兼容与回滚

- 无 DB 迁移、无 API 形状变化、无前端变更。
- 回滚 = revert 单个 commit, 无状态残留(全部为进程内行为)。

## 测试策略

- `internal/relay/route_test.go`: `pickGroupItem` 谓词行为(手动指定成员被跳过→零值; 故障转移跳过后选到次优先级; 全员被跳过→零值)与 `allItemsSkipped` 模式感知(手动看指定成员; 故障转移要求非空且全跳过; 空组为 false)。
- `internal/op` 渠道测试: 直接向包内缓存 Set 渠道/模型/凭据/授权构造引用链, 验证 `ChannelGrantDisabled` 各分支(渠道禁用→true; 渠道启用/引用缺失→false)与 `ChannelGrantGet` 渠道禁用报错。
- 纯单元测试, 零外部依赖, 遵循仓库测试约定。
