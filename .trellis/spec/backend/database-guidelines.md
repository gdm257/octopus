# Database Guidelines

> ORM, migrations, query patterns, naming conventions.

---

## Overview

ORM 用 GORM，支持 SQLite（默认，`github.com/glebarez/sqlite` 纯 Go 驱动）/ MySQL / PostgreSQL 三种后端，连接初始化集中在 `internal/db/db.go`。启动时 `AutoMigrate` 全部实体，复杂变更走版本化迁移。

---

## Models

实体定义在 `internal/model/`，每个字段带 `json` + `gorm` 双标签与中文行尾注释（见 `internal/model/channel.go`）：

```go
type ChannelKeyConfig struct {
    Name    string `json:"name" gorm:"not null;index:idx_channel_key,unique"` // 凭据名称, 界面展示与人工识别用。
    Key     string `json:"key" gorm:"not null"`                               // 上游访问凭据。
    Enabled bool   `json:"enabled" gorm:"default:true"`                       // 是否可用, 禁用后不参与选路但保留统计。
}
```

既有约定：

- **实体与 API 形状共享同一个 `XxxConfig` 结构体**，实体内嵌它，避免三处各写一遍（`Channel` 内嵌 `ChannelConfig`，`internal/model/channel.go:44`）。
- **可编辑配置整体读写**：编辑副本命名为 `XxxDetail`，读取与提交同构，后端按整体替换处理，不按字段增量比对。
- 复合唯一索引用 `index:idx_xxx,unique` 命名标签；外键删除统一 `constraint:OnDelete:CASCADE`。
- 切片字段（如 `CustomHeader`）用 `gorm:"serializer:json"` 存 JSON。
- 协议位等会落库并暴露给前端的枚举值**不可变更**，只允许追加。

---

## Query Patterns

数据访问只发生在 `internal/op/`，handler 不直接碰 `db.GetDB()`。

- **读走进程内缓存**：每个资源一个 `cache.New[K, V]`（`internal/op/channel.go:16`），读操作先查缓存；缓存即读模型，启动时 `InitCache` 全量加载。
- **写穿透**：写库成功后同步更新缓存。多表写入用事务：

```go
if err := db.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&channel).Error; err != nil {
        return fmt.Errorf("failed to create channel: %w", err)
    }
    return syncChannelChildren(tx, channel.ID, detail)
}); err != nil {
    return nil, err
}
```

- **ctx 是最后一个参数**（既有签名即如此，如 `ChannelCreate(detail *model.ChannelDetail, ctx context.Context)`），调用方传 `c.Request.Context()`。
- 更新指定列时用 `.Select("col1", "col2").Updates(...)` 点名列，避免零值字段丢写或整行覆盖统计列（`internal/op/channel.go:104`）。
- 统计类高频写入先进内存，由定时任务批量落库（`internal/op/stats.go`），不要每请求写库。
- 错误一律 `fmt.Errorf("failed to xxx: %w", err)` 包装返回，不 panic。

---

## Migrations

- 新增实体：加进 `internal/db/db.go:55` 的 `AutoMigrate` 列表即可，字段只能加不能改语义。
- 数据改写/删列等 AutoMigrate 覆盖不了的变更：在 `internal/db/migrate/` 新建递增编号文件（当前最大 `012.go`），在 `init()` 里注册：

```go
func init() {
    RegisterBeforeAutoMigration(Migration{
        Version: 13,
        Up:      migrateXxx,
    })
}
```

- 约束表结构的放 Before（如需在 AutoMigrate 前删旧列），数据改写放 After。
- 迁移必须**可重跑**：先判断现状（`hasPhysicalColumn` / 前缀检查），已处理的行跳过（参考 `internal/db/migrate/012.go`）。
- 跨库兼容：SQLite 与 MySQL/PG 行为不同时用 `db.Dialector.Name()` 分支，删列用 `dropColumnIfExists`（`migrate/migrate.go:153`）。
- 新表/新列命名：表名复数 snake_case（GORM 默认），列名 snake_case。

---

## Examples

- 实体+形状共享：`internal/model/channel.go`
- 缓存读写穿透：`internal/op/channel.go`（`ChannelCreate` / `ChannelUpdate` / `channelRefreshCache`）
- 事务 + 子表整体替换：`syncChannelKeys` / `syncChannelModels` / `syncChannelGrants`
