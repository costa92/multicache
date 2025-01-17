# MultiCache 使用文档

## 安装
```bash
go get github.com/yourusername/multicache
```

## 快速开始

### 1. 定义模型
```go
type User struct {
    ID     uint    `json:"id"`
    Name   string  `json:"name"`
    Email  string  `json:"email"`
    Orders []Order `json:"orders"`
}

type Order struct {
    ID        uint      `json:"id"`
    UserID    uint      `json:"user_id"`
    Amount    float64   `json:"amount"`
    CreatedAt time.Time `json:"created_at"`
}
```

### 2. 使用 GORM 加载器
```go
// 创建加载器
userLoader := loader.NewGormLoader(db, models.User{}).
    WithPreload("Orders").
    WithCondition("status = ?", "active").
    WithPreloadQuery("Orders", "created_at > ?", time.Now().AddDate(0, -1, 0))

// 加载数据
users, err := userLoader.Load()
if err != nil {
    log.Fatal(err)
}
```

### 3. 使用缓存管理器
```go
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

### 4. 使用查询条件
```go
// 字符串条件
nameCondition := cache.StringFieldCondition[models.User]{
    FieldExtractor: func(u models.User) string { return u.Name },
    Value:         "John",
    Operation:     "contains",
}

// 数值条件
amountCondition := cache.NumberFieldCondition[models.Order, float64]{
    FieldExtractor: func(o models.Order) float64 { return o.Amount },
    Value:         1000,
    Operation:     "gte",
}

// 复合条件
composite := cache.CompositeCondition[models.User]{
    Conditions: []cache.QueryCondition[models.User]{condition1, condition2},
    Operation:  "and",
}
```

## 高级用法

### 1. MongoDB 支持
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

### 2. 组合使用场景
```go
// 不同 TTL 的缓存
activeUserCache := cache.NewCacheManager[models.User](activeUserLoader).
    WithTTL(5 * time.Minute)
premiumUserCache := cache.NewCacheManager[models.User](premiumUserLoader).
    WithTTL(1 * time.Minute)

// 实时查询
recentOrderLoader := loader.NewGormLoader(db, models.Order{}).
    WithCondition("created_at > ?", time.Now().AddDate(0, 0, -1)).
    WithJoinsModel(models.User{}, "orders.user_id", "users.id")
```

## 最佳实践

1. 缓存策略
   - 频繁访问的数据使用较长的 TTL
   - 实时性要求高的数据使用较短的 TTL
   - 关键数据使用复合缓存策略

2. 查询优化
   - 使用适当的预加载减少 N+1 查询
   - 合理使用连接查询
   - 避免过度使用复杂查询

3. 错误处理
   - 始终检查错误返回
   - 实现优雅降级
   - 使用日志记录错误

4. 性能优化
   - 合理设置缓存容量
   - 适当使用调试模式
   - 监控缓存命中率

## 注意事项

1. 线程安全
   - 缓存操作是线程安全的
   - 避免在回调函数中修改共享状态

2. 内存管理
   - 注意设置合适的 TTL
   - 及时清理不需要的缓存
   - 监控内存使用情况

3. 错误处理
   - 处理所有可能的错误情况
   - 实现合适的降级策略
   - 提供有意义的错误信息 