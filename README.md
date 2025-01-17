# MultiCache

MultiCache 是一个高性能的 Go 语言本地缓存库，支持多数据源、关联查询和灵活的缓存策略。

## 特性

- 支持多种数据源
  - GORM (MySQL, PostgreSQL, SQLite 等)
  - MongoDB
- 灵活的查询条件
  - 条件查询
  - 关联查询
  - 预加载
  - 聚合查询
- 高性能缓存
  - 可配置的 TTL
  - 自动过期
  - 线程安全
  - 内存高效
- 类型安全
  - 基于泛型
  - 编译时类型检查

## 安装

```bash
go get github.com/costa92/multicache
```

## 快速开始

### 1. 使用 GORM 加载器

```go
import (
    "github.com/costa92/multicache/loader"
    "github.com/costa92/multicache/models"
)

// 创建加载器
userLoader := loader.NewGormLoader(db, models.User{}).
    WithPreload("Orders").
    WithCondition("status = ?", "active")

// 加载数据
users, err := userLoader.Load()
if err != nil {
    log.Fatal(err)
}
```

### 2. 使用缓存管理器

```go
import "github.com/costa92/multicache/cache"

// 创建缓存
userCache := cache.NewCacheManager[models.User](userLoader).
    WithTTL(5 * time.Minute)

// 初始化缓存
if err := userCache.Refresh(); err != nil {
    log.Fatal(err)
}

// 从缓存获取数据
user, err := userCache.Get(1)
if err != nil {
    log.Printf("User not found: %v", err)
}
```

### 3. MongoDB 支持

```go
// 创建 MongoDB 加载器
orderLoader := loader.NewMongoLoader[models.Order](ctx, coll).
    WithAggregate(mongo.Pipeline{
        bson.D{{Key: "$match", Value: bson.M{"amount": bson.M{"$gt": 1000}}}},
        bson.D{{Key: "$sort", Value: bson.M{"amount": -1}}},
    })

// 创建缓存
orderCache := cache.NewRelatedCacheManager[models.Order](orderLoader, 1*time.Minute)
```

## 文档

- [需求文档](docs/requirements.md)
- [使用文档](docs/usage.md)

## 主要组件

1. 数据加载器
   - `GormLoader`: GORM 数据库加载器
   - `MongoLoader`: MongoDB 数据库加载器

2. 缓存管理器
   - `CacheManager`: 基础缓存管理器
   - `RelatedCacheManager`: 关联数据缓存管理器

3. 查询条件
   - `StringFieldCondition`: 字符串字段条件
   - `NumberFieldCondition`: 数值字段条件
   - `CompositeCondition`: 复合条件

## 示例

查看 [examples](examples/) 目录获取更多使用示例：
- 基本 GORM 查询
- MongoDB 聚合查询
- 缓存使用场景
- 复杂查询示例

## 性能优化

- 使用适当的预加载减少 N+1 查询
- 合理设置缓存 TTL
- 启用调试模式排查性能问题
- 监控缓存命中率

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License

## 作者

costa92

## 版本历史

- v0.1.0 - 初始版本
  - 支持 GORM 和 MongoDB
  - 基本缓存功能
  - 查询条件支持
