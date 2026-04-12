# improve.md

## 目标

这份文档按“苹果哲学”记录 `github.com/kaptinlin/template` 的收敛结果与剩余工作。

这里的苹果哲学，不是视觉风格，而是四条工程原则：

- 少即是多：外部只暴露一个清晰心智模型
- 默认正确：最安全、最常见的路径应该天然成立
- 一致性：同类能力的行为和命名要统一
- 隐藏复杂度：复杂机制留在内部，不把拼装责任推给调用者

当前状态：

- `Engine + Format + Feature` 已经是唯一主设计
- `go test ./...` 当前通过
- 这一轮已经完成一组核心收敛，剩下的是继续精修边界与体验

## 已完成

### 1. `resolved` 已成为真正的缓存身份

已完成内容：

- `Engine.Load` 主缓存改为按 `resolved` 存储
- 增加了 `name -> resolved` 辅助索引
- 不同输入名映射到同一模板时会共享同一个编译结果

收益：

- `Loader` 契约和 `Engine` 实现重新对齐
- alias/template indirection 不再导致重复缓存
- `ChainLoader` 的分层缓存语义闭合了

### 2. 并发加载已具备 in-flight 去重

已完成内容：

- `Engine.Load` 对同一个 `resolved` 增加了 in-flight 合并
- 并发请求会等待同一次编译结果，而不是重复做 IO 和编译

收益：

- 冷启动和缓存失效场景更稳
- 并发访问不再把内部重复工作暴露成性能抖动

### 3. 子上下文继承规则已收敛

已完成内容：

- `NewChildContext` 现在会保留运行时状态
- 新增 `NewIsolatedChildContext`
- `include` 的子上下文派生已统一走固定入口

覆盖的运行时状态包括：

- `engine`
- `autoescape`
- `includeDepth`
- `currentLeaf`

收益：

- 不再依赖调用点手动补齐隐式字段
- include / only 的上下文语义更稳定

### 4. `for` 循环作用域已明确

已完成内容：

- `ForNode.Execute` 退出时会恢复进入循环前的私有绑定
- `loop` 不再泄漏到循环外
- 单变量和双变量绑定都会恢复或清除
- 嵌套循环结束后，父级 `loop` 会正确恢复

收益：

- 循环语义从“实现刚好如此”变成“显式契约”
- 私有上下文不再留下模糊残留状态

### 5. registry 注册语义已经显式化

已完成内容：

- `Registry` / `TagRegistry` 现在都有：
  - `Register`
  - `Replace`
  - `MustRegister`
- engine 内部 override 路径改用 `Replace`
- built-in bootstrap 改用 `MustRegister`

收益：

- 系统内部不再依赖模糊的“可能覆盖 / 可能报错”语义
- 注册意图从调用点即可读出

### 6. loader-backed 错误已稳定带模板来源

已完成内容：

- `Engine.Load` 和命名编译路径会统一补模板名
- 解析类错误会显示为 `template:line:col: message`
- 非解析类错误至少会显示为 `template: ...`
- `errors.Is` / `errors.As` 仍然可以穿透到底层错误

收益：

- 模板作者看到的错误更像产品错误，而不是内部实现错误
- 不影响现有 sentinel / type matching

### 7. 常见 `0:0` 解析错误已减少

已完成内容：

- `Parser.ParseExpression` 在空参数场景下会用调用点锚定位置
- `ExprParser.parseErr` 在读到输入末尾时会回退到上一个 token 的位置
- 空 `if` 条件这类高频错误现在能带非零位置

收益：

- loader-backed 错误的位置信息更可用
- 常见解析失败不再轻易退化为 `0:0`

### 8. benchmark 和并发测试矩阵已补齐关键路径

已完成内容：

- 增加了 `Engine.Load` 命中 / 未命中 benchmark
- 增加了 layout render benchmark
- 增加了 cached parallel render benchmark
- 增加了 alias-name 并发 render 共享 resolved-template 的测试

收益：

- 前面收紧的缓存 / 并发语义有了直接回归保护
- 关键路径不再只能靠直觉优化

### 9. `ContextBuilder.Struct` 已完成一轮保守快路径优化

已完成内容：

- 对普通 struct / 嵌套 plain struct / `[]T` / `map[string]T` 增加了反射快路径
- 对 `omitempty` 做了对齐
- 对 `MarshalJSON` / `MarshalText` 等自定义 JSON 语义保留原回退行为
- 加了对应 benchmark 和语义测试

收益：

- 不改外部 API，也不破坏 JSON 语义
- 热门场景下 `Struct` 成本明显下降

## 当前基线

当前已经有一组可以直接参考的 benchmark 基线：

- `BenchmarkEngineLoadCacheHit`
- `BenchmarkEngineLoadCacheMissAfterReset`
- `BenchmarkEngineRenderLayout`
- `BenchmarkEngineRenderCachedParallel`
- `BenchmarkContextBuilderKeyValue`
- `BenchmarkContextBuilderStructFlat`
- `BenchmarkContextBuilderStructNested`

这些 benchmark 的价值不在于绝对数字，而在于后续优化和回归时有了同一把尺子。

## 剩余项

### 1. 继续补齐“所有错误都有准确位置”

现在模板名和大部分高频位置信息已经补上，但还不能保证所有解析/表达式错误都给出理想的行列。

剩余工作：

- 继续梳理表达式解析尾部错误
- 检查 include / extends / filter arg / subscript 场景的定位是否都稳定
- 用更细粒度测试锁定位置质量

### 2. 补充 chain loader 与依赖关系的更强契约测试

当前缓存与 resolved 语义已经收紧，但还可以继续加：

- 更明确的 chain loader 层隔离测试
- include / extends 在 alias / layered loader 场景下的契约测试

### 3. 评估 `ContextBuilder.Struct` 是否值得继续下钻

当前已经做了有界优化，但不建议立即继续复杂化。

下一步应该先观察：

- benchmark 是否已经足够
- 实际热路径是否真的依赖 builder

如果没有明确收益，就不应继续往更复杂的反射缓存或代码生成方向推进。

### 4. 文档层还可以同步收口

建议后续把实现收敛同步反映到：

- `README.md`
- `ARCHITECTURE.md`
- `CONTRIBUTING.md`

重点不是写更多，而是让文档和当前实现状态一致。

## 建议执行顺序

如果继续按“苹果式收敛”往下做，我建议顺序如下：

1. 继续消灭剩余的 `0:0` 错误定位
2. 补强 chain loader / layout alias 的契约测试
3. 观察 `ContextBuilder.Struct` benchmark，再决定是否继续优化
4. 同步更新 README / architecture / contributing 文档

## 总结

这个项目目前最关键的一步已经完成了：从“功能已经能工作”进入了“内部语义开始真正收紧”的阶段。

已经收紧的点包括：

- 缓存身份
- 并发加载
- 子上下文派生
- 循环作用域
- registry 语义
- 错误来源
- benchmark 基线
- `ContextBuilder.Struct` 热点路径

接下来不需要大扩张，继续做小而稳的收口即可。真正符合苹果哲学的方向，仍然是继续减少内部歧义，而不是继续增加表面能力。
