# Template

一个轻量级 Go 模板引擎，采用类似 Liquid/Django 的语法，支持变量插值、过滤器、条件判断和循环控制。

## 安装

```sh
go get github.com/kaptinlin/template
```

## 快速开始

### 一步渲染

```go
output, err := template.Render("Hello, {{ name|upper }}!", map[string]any{
    "name": "alice",
})
// output: "Hello, ALICE!"
```

### 编译后复用

对于需要多次渲染的模板，先编译一次，后续复用：

```go
tmpl, err := template.Compile("Hello, {{ name }}!")
if err != nil {
    log.Fatal(err)
}

output, err := tmpl.Render(map[string]any{"name": "World"})
// output: "Hello, World!"
```

### 使用 io.Writer

```go
tmpl, _ := template.Compile("Hello, {{ name }}!")
ctx := template.NewExecutionContext(map[string]any{"name": "World"})
tmpl.Execute(ctx, os.Stdout)
```

## 模板语法

### 变量

用 `{{ }}` 输出变量，支持点号访问嵌套属性：

```
{{ user.name }}
{{ user.address.city }}
{{ items.0 }}
```

详见 [变量文档](docs/variables.md)。

### 过滤器

用管道符 `|` 对变量应用过滤器，支持链式调用和参数传递：

```
{{ name|upper }}
{{ title|truncate:20 }}
{{ name|lower|capitalize }}
{{ price|plus:10|times:2 }}
```

详见 [过滤器文档](docs/filters.md)。

### 条件判断

```
{% if score > 80 %}
    优秀
{% elif score > 60 %}
    及格
{% else %}
    不及格
{% endif %}
```

### 循环

```
{% for item in items %}
    {{ item }}
{% endfor %}

{% for key, value in dict %}
    {{ key }}: {{ value }}
{% endfor %}
```

循环内可使用 `loop` 变量：

| 属性 | 说明 |
|------|------|
| `loop.Index` | 当前索引（从 0 开始） |
| `loop.Revindex` | 反向索引 |
| `loop.First` | 是否第一次迭代 |
| `loop.Last` | 是否最后一次迭代 |
| `loop.Length` | 集合总长度 |

支持 `{% break %}` 和 `{% continue %}` 控制循环流程。

详见 [控制结构文档](docs/control-structure.md)。

### 注释

```
{# 这段内容不会出现在输出中 #}
```

### 表达式

支持以下运算符（按优先级从低到高）：

| 运算符 | 说明 |
|--------|------|
| `or`, `\|\|` | 逻辑或 |
| `and`, `&&` | 逻辑与 |
| `==`, `!=`, `<`, `>`, `<=`, `>=` | 比较 |
| `+`, `-` | 加减 |
| `*`, `/`, `%` | 乘除取模 |
| `not`, `-`, `+` | 一元运算 |

字面量支持：字符串（`"text"` / `'text'`）、数字（`42`、`3.14`）、布尔值（`true` / `false`）、空值（`null`）。

## 内置过滤器

### 字符串

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `default` | 值为空时返回默认值 | `{{ name\|default:"匿名" }}` |
| `upper` | 转大写 | `{{ name\|upper }}` |
| `lower` | 转小写 | `{{ name\|lower }}` |
| `capitalize` | 首字母大写 | `{{ name\|capitalize }}` |
| `titleize` | 每个单词首字母大写 | `{{ title\|titleize }}` |
| `trim` | 去除首尾空白 | `{{ text\|trim }}` |
| `truncate` | 截断到指定长度 | `{{ text\|truncate:20 }}` |
| `truncateWords` | 截断到指定单词数 | `{{ text\|truncateWords:5 }}` |
| `replace` | 替换子串 | `{{ text\|replace:"old","new" }}` |
| `remove` | 移除子串 | `{{ text\|remove:"bad" }}` |
| `append` | 追加字符串 | `{{ name\|append:"!" }}` |
| `prepend` | 前置字符串 | `{{ name\|prepend:"Hi " }}` |
| `split` | 按分隔符拆分 | `{{ csv\|split:"," }}` |
| `length` | 获取长度 | `{{ name\|length }}` |
| `slugify` | 转为 URL 友好格式 | `{{ title\|slugify }}` |
| `camelize` | 转为 camelCase | `{{ name\|camelize }}` |
| `pascalize` | 转为 PascalCase | `{{ name\|pascalize }}` |
| `dasherize` | 转为短横线分隔 | `{{ name\|dasherize }}` |
| `pluralize` | 单复数选择 | `{{ count\|pluralize:"item","items" }}` |
| `ordinalize` | 转为序数词 | `{{ num\|ordinalize }}` |

### 数学

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `plus` | 加法 | `{{ price\|plus:10 }}` |
| `minus` | 减法 | `{{ price\|minus:5 }}` |
| `times` | 乘法 | `{{ price\|times:2 }}` |
| `divide` | 除法 | `{{ total\|divide:3 }}` |
| `modulo` | 取模 | `{{ num\|modulo:2 }}` |
| `abs` | 绝对值 | `{{ num\|abs }}` |
| `round` | 四舍五入 | `{{ pi\|round:2 }}` |
| `floor` | 向下取整 | `{{ num\|floor }}` |
| `ceil` | 向上取整 | `{{ num\|ceil }}` |
| `atLeast` | 确保最小值 | `{{ num\|atLeast:0 }}` |
| `atMost` | 确保最大值 | `{{ num\|atMost:100 }}` |

### 数组

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `join` | 用分隔符连接 | `{{ items\|join:", " }}` |
| `first` | 第一个元素 | `{{ items\|first }}` |
| `last` | 最后一个元素 | `{{ items\|last }}` |
| `size` | 集合长度 | `{{ items\|size }}` |
| `reverse` | 反转顺序 | `{{ items\|reverse }}` |
| `unique` | 去重 | `{{ items\|unique }}` |
| `shuffle` | 随机排列 | `{{ items\|shuffle }}` |
| `random` | 随机取一个 | `{{ items\|random }}` |
| `max` | 最大值 | `{{ scores\|max }}` |
| `min` | 最小值 | `{{ scores\|min }}` |
| `sum` | 求和 | `{{ scores\|sum }}` |
| `average` | 平均值 | `{{ scores\|average }}` |
| `map` | 提取每个元素的指定键 | `{{ users\|map:"name" }}` |

### Map

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `extract` | 按点号路径提取嵌套值 | `{{ data\|extract:"user.name" }}` |

### 日期

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `date` | 格式化日期 | `{{ timestamp\|date:"Y-m-d" }}` |
| `year` | 提取年份 | `{{ timestamp\|year }}` |
| `month` | 提取月份数字 | `{{ timestamp\|month }}` |
| `monthFull` / `month_full` | 月份全名 | `{{ timestamp\|monthFull }}` |
| `day` | 提取日 | `{{ timestamp\|day }}` |
| `week` | ISO 周数 | `{{ timestamp\|week }}` |
| `weekday` | 星期几 | `{{ timestamp\|weekday }}` |
| `timeAgo` / `timeago` | 相对时间 | `{{ timestamp\|timeAgo }}` |

### 数字格式化

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `number` | 数字格式化 | `{{ price\|number:"0.00" }}` |
| `bytes` | 转为可读字节单位 | `{{ fileSize\|bytes }}` |

### 序列化

| 过滤器 | 说明 | 示例 |
|--------|------|------|
| `json` | 序列化为 JSON | `{{ data\|json }}` |

## 扩展

### 自定义过滤器

```go
template.RegisterFilter("repeat", func(value any, args ...string) (any, error) {
    s := fmt.Sprintf("%v", value)
    n := 2
    if len(args) > 0 {
        if parsed, err := strconv.Atoi(args[0]); err == nil {
            n = parsed
        }
    }
    return strings.Repeat(s, n), nil
})

// {{ "ha"|repeat:3 }} -> "hahaha"
```

### 自定义标签

通过实现 `Statement` 接口并调用 `RegisterTag` 注册自定义标签。以下是一个 `{% set %}` 标签的示例：

```go
type SetNode struct {
    VarName    string
    Expression template.Expression
    Line, Col  int
}

func (n *SetNode) Position() (int, int) { return n.Line, n.Col }
func (n *SetNode) String() string       { return fmt.Sprintf("Set(%s)", n.VarName) }
func (n *SetNode) Execute(ctx *template.ExecutionContext, _ io.Writer) error {
    val, err := n.Expression.Evaluate(ctx)
    if err != nil {
        return err
    }
    ctx.Set(n.VarName, val.Interface())
    return nil
}

template.RegisterTag("set", func(doc *template.Parser, start *template.Token, arguments *template.Parser) (template.Statement, error) {
    varToken, err := arguments.ExpectIdentifier()
    if err != nil {
        return nil, arguments.Error("expected variable name after 'set'")
    }
    if arguments.Match(template.TokenSymbol, "=") == nil {
        return nil, arguments.Error("expected '=' after variable name")
    }
    expr, err := arguments.ParseExpression()
    if err != nil {
        return nil, err
    }
    return &SetNode{
        VarName:    varToken.Value,
        Expression: expr,
        Line:       start.Line,
        Col:        start.Col,
    }, nil
})
```

更多示例见 [examples](examples/) 目录。

## Context 构建

```go
// 直接使用 map
output, _ := template.Render(source, map[string]any{
    "name": "Alice",
    "age":  30,
})

// 使用 ContextBuilder（支持 struct 展开）
ctx, err := template.NewContextBuilder().
    KeyValue("name", "Alice").
    Struct(user).
    Build()
output, _ := tmpl.Render(ctx)
```

## 错误报告

所有错误都包含精确的行列位置信息：

```
lexer error at line 1, col 7: unclosed variable tag, expected '}}'
parse error at line 1, col 4: unknown tag: unknown
parse error at line 1, col 19: unexpected EOF, expected one of: [elif else endif]
```

## 架构

详见 [ARCHITECTURE.md](ARCHITECTURE.md)。

## 贡献

欢迎贡献代码，请参阅 [贡献指南](CONTRIBUTING.md)。

## 许可证

MIT License - 详见 [LICENSE](LICENSE)。
