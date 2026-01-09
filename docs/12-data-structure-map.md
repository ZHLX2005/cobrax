# 映射数据结构 (Map Data Structure)

## 概述

映射（Map）是 Go 语言中最重要的数据结构之一，它提供了键值对存储和快速查找能力。在 Cobra-X 中，映射被广泛用于配置管理、去重、值存储和元数据管理。

## 在 Cobra-X 中的应用

### 核心应用场景

1. **Flag 值存储**: 存储用户配置的参数值
2. **去重处理**: 避免重复处理相同的元素
3. **元数据存储**: 存储命令的附加信息
4. **缓存实现**: 缓存计算结果
5. **主题配置**: 存储主题相关的配置

## 代码实现分析

### 1. Flag 值映射

```go
// cobra/decorate.go:437-512
func configureFlags(renderer tui.Renderer, cmd *spf13cobra.Command) (map[string]string, error) {
    var items []tui.FlagItem
    seen := make(map[string]bool)  // 用于去重

    // 遍历当前命令及其所有父命令，聚合所有 flags
    current := cmd
    for current != nil {
        // 收集 LocalFlags
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if isTUIFlag(flag.Name) || seen[flag.Name] {
                return
            }

            item := tui.FlagItem{
                Name:         flag.Name,
                Description:  flag.Usage,
                DefaultValue: flag.DefValue,
                CurrentValue: flag.DefValue,
                SourceCommand: current.Name(),
                // ...
            }

            items = append(items, item)
            seen[flag.Name] = true  // 标记为已处理
        })

        current = current.Parent()
    }

    // 渲染表单并返回值映射
    return renderer.RenderFlagForm("Configure: "+cmd.Name(), items)
}
```

**映射用途**：
- **去重**: 防止同一个 flag 被多次添加
- **快速查找**: O(1) 时间复杂度检查是否已处理

### 2. Flag 值应用

```go
// cobra/decorate.go:514-532
func applyFlagValues(cmd *spf13cobra.Command, values map[string]string) {
    // 遍历当前命令及其所有父命令，应用对应的 flag 值
    current := cmd
    for current != nil {
        for name, value := range values {
            // 先尝试查找 LocalFlag
            flag := current.LocalFlags().Lookup(name)
            if flag == nil {
                // 如果没有找到，尝试查找 PersistentFlag
                flag = current.PersistentFlags().Lookup(name)
            }
            if flag != nil {
                flag.Value.Set(value)
                flag.Changed = true
            }
        }
        current = current.Parent()
    }
}
```

**映射操作**：
- 遍历所有键值对
- 根据键查找对应的 flag
- 应用值到 flag

### 3. 命令元数据映射

```go
// cobra/decorate.go:47-58
func Enhance(cmd *spf13cobra.Command, opts ...EnhanceOption) *spf13cobra.Command {
    if cmd == nil {
        return nil
    }

    config := &EnhanceConfig{
        TUIConfig: DefaultTUIConfig(),
    }

    // 应用选项
    for _, opt := range opts {
        opt(config)
    }

    // 存储 TUI 配置到命令的 Annotations 中
    if cmd.Annotations == nil {
        cmd.Annotations = make(map[string]string)
    }
    cmd.Annotations["tui.enabled"] = "true"
    if config.TUIConfig != nil {
        if config.TUIConfig.Theme != nil {
            cmd.Annotations["tui.theme"] = config.TUIConfig.Theme.Name
        }
    }

    // 添加 TUI flags
    addTUIFlags(cmd)

    // 包装 PreRun/E 以拦截执行
    wrapExecute(cmd, config)

    return cmd
}
```

**Annotations 映射**：
- 存储命令的元数据
- 不影响原有命令结构
- 支持动态添加属性

### 4. 表单值映射

```go
// tui/default_renderer.go:283-298
func newFormModel(items []FlagItem, theme *style.Theme, width, height int) *formModel {
    // 使用映射初始化值
    values := make(map[string]string)
    for _, item := range items {
        values[item.Name] = item.DefaultValue
    }

    return &formModel{
        items:  items,
        cursor: 0,
        values: values,  // 存储表单值
        theme:  theme,
        width:  width,
        height: height,
    }
}
```

## 映射的常见操作

### 1. 分组映射

```go
// tui/default_renderer.go:421-434
// 按来源命令分组表单项
itemsBySource := make(map[string][]*FlagItem)
var sourceOrder []string // 保持分组顺序

for i, item := range m.items {
    source := item.SourceCommand
    if source == "" {
        source = "Global"
    }
    if _, ok := itemsBySource[source]; !ok {
        itemsBySource[source] = make([]*FlagItem, 0)
        sourceOrder = append(sourceOrder, source)
    }
    itemsBySource[source] = append(itemsBySource[source], &m.items[i])
}
```

**分组映射特点**：
- 键是分组标识
- 值是该组的元素列表
- 额外维护顺序信息

### 2. 去重映射

```go
// cobra/command.go:321-400
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem
    seen := make(map[string]bool)  // 去重映射

    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if strings.HasPrefix(flag.Name, "tui-") || flag.Name == "tui" || seen[flag.Name] {
                return  // 跳过已处理的 flag
            }

            // 处理 flag...
            items = append(items, item)
            seen[flag.Name] = true  // 标记为已处理
        })

        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

### 3. 计数映射

```go
// 统计命令使用频率
type CommandStats struct {
    counts map[string]int
    mutex  sync.RWMutex
}

func (s *CommandStats) Record(cmd string) {
    s.mutex.Lock()
    defer s.mutex.Unlock()

    if s.counts == nil {
        s.counts = make(map[string]int)
    }
    s.counts[cmd]++
}

func (s *CommandStats) Get(cmd string) int {
    s.mutex.RLock()
    defer s.mutex.RUnlock()
    return s.counts[cmd]
}

func (s *CommandStats) Top(n int) []string {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    type entry struct {
        cmd   string
        count int
    }

    entries := make([]entry, 0, len(s.counts))
    for cmd, count := range s.counts {
        entries = append(entries, entry{cmd, count})
    }

    sort.Slice(entries, func(i, j int) bool {
        return entries[i].count > entries[j].count
    })

    result := make([]string, 0, n)
    for i := 0; i < n && i < len(entries); i++ {
        result = append(result, entries[i].cmd)
    }

    return result
}
```

### 4. 缓存映射

```go
// 主题缓存
type ThemeCache struct {
    themes map[string]*style.Theme
    mutex  sync.RWMutex
}

func (c *ThemeCache) Get(name string) (*style.Theme, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()

    theme, ok := c.themes[name]
    return theme, ok
}

func (c *ThemeCache) Set(name string, theme *style.Theme) {
    c.mutex.Lock()
    defer c.mutex.Unlock()

    if c.themes == nil {
        c.themes = make(map[string]*style.Theme)
    }
    c.themes[name] = theme
}
```

## 高级映射技巧

### 1. 分层映射

```go
// 分层配置系统
type LayeredConfig struct {
    layers []map[string]interface{}
}

func NewLayeredConfig() *LayeredConfig {
    return &LayeredConfig{
        layers: make([]map[string]interface{}, 0),
    }
}

func (c *LayeredConfig) PushLayer(layer map[string]interface{}) {
    c.layers = append(c.layers, layer)
}

func (c *LayeredConfig) Get(key string) (interface{}, bool) {
    // 从上到下查找
    for i := len(c.layers) - 1; i >= 0; i-- {
        if value, ok := c.layers[i][key]; ok {
            return value, true
        }
    }
    return nil, false
}

func (c *LayeredConfig) Set(key string, value interface{}) {
    // 设置到最上层
    if len(c.layers) == 0 {
        c.PushLayer(make(map[string]interface{}))
    }
    c.layers[len(c.layers)-1][key] = value
}
```

### 2. 观察者映射

```go
// 带变化通知的映射
type ObservableMap struct {
    data     map[string]interface{}
    observers map[string][]func(interface{})
    mutex    sync.RWMutex
}

func NewObservableMap() *ObservableMap {
    return &ObservableMap{
        data:     make(map[string]interface{}),
        observers: make(map[string][]func(interface{})),
    }
}

func (m *ObservableMap) Set(key string, value interface{}) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.data[key] = value

    // 通知观察者
    if observers, ok := m.observers[key]; ok {
        for _, observer := range observers {
            go observer(value)
        }
    }
}

func (m *ObservableMap) Subscribe(key string, observer func(interface{})) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.observers[key] = append(m.observers[key], observer)
}

func (m *ObservableMap) Get(key string) (interface{}, bool) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    value, ok := m.data[key]
    return value, ok
}
```

### 3. 过期映射

```go
// 带过期时间的缓存
type ExpiringMap struct {
    data     map[string]*cacheEntry
    mutex    sync.RWMutex
}

type cacheEntry struct {
    value      interface{}
    expiration time.Time
}

func NewExpiringMap() *ExpiringMap {
    return &ExpiringMap{
        data: make(map[string]*cacheEntry),
    }
}

func (m *ExpiringMap) Set(key string, value interface{}, ttl time.Duration) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.data[key] = &cacheEntry{
        value:      value,
        expiration: time.Now().Add(ttl),
    }
}

func (m *ExpiringMap) Get(key string) (interface{}, bool) {
    m.mutex.RLock()
    entry, ok := m.data[key]
    m.mutex.RUnlock()

    if !ok {
        return nil, false
    }

    // 检查是否过期
    if time.Now().After(entry.expiration) {
        m.mutex.Lock()
        delete(m.data, key)
        m.mutex.Unlock()
        return nil, false
    }

    return entry.value, true
}

func (m *ExpiringMap) Cleanup() {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    now := time.Now()
    for key, entry := range m.data {
        if now.After(entry.expiration) {
            delete(m.data, key)
        }
    }
}
```

### 4. 并发安全映射

```go
// 使用 sync.Map 实现并发安全
type SafeCache struct {
    data sync.Map
}

func (c *SafeCache) Get(key string) (interface{}, bool) {
    return c.data.Load(key)
}

func (c *SafeCache) Set(key string, value interface{}) {
    c.data.Store(key, value)
}

func (c *SafeCache) Delete(key string) {
    c.data.Delete(key)
}

func (c *SafeCache) Range(fn func(key string, value interface{}) bool) {
    c.data.Range(func(key, value interface{}) bool {
        return fn(key.(string), value)
    })
}
```

## 映射的性能考虑

### 1. 预分配容量

```go
// 好的做法 - 预分配容量
seen := make(map[string]bool, len(flags))

// 避免 - 动态扩容
seen := make(map[string]bool)
```

### 2. 避免大对象作为键

```go
// 好的做法 - 使用简单类型作为键
values := make(map[string]string)

// 避免 - 使用复杂结构作为键
type ComplexKey struct {
    a int
    b string
    c []byte
}
values := make(map[ComplexKey]string)  // 性能较差
```

### 3. 值类型选择

```go
// 使用指针减少复制
items := make(map[string]*FlagItem)  // 好
// vs
items := make(map[string]FlagItem)    // 可能有额外复制
```

## 映射的最佳实践

### 1. 零值检查

```go
// 检查键是否存在
if value, ok := m.data[key]; ok {
    // 键存在
} else {
    // 键不存在
}
```

### 2. 删除安全

```go
// 安全删除
delete(m.data, key)  // 即使 key 不存在也不会 panic

// 检查后再删除（不必要）
if _, ok := m.data[key]; ok {
    delete(m.data, key)
}
```

### 3. 迭代顺序

```go
// Go 1.0+ 的 map 迭代顺序是随机的
for key, value := range m.data {
    // 每次运行顺序可能不同
}

// 如果需要有序，使用额外的切片
type OrderedMap struct {
    keys   []string
    values map[string]interface{}
}

func (m *OrderedMap) Set(key string, value interface{}) {
    if _, exists := m.values[key]; !exists {
        m.keys = append(m.keys, key)
    }
    m.values[key] = value
}

func (m *OrderedMap) Range(fn func(key string, value interface{})) {
    for _, key := range m.keys {
        fn(key, m.values[key])
    }
}
```

## 常见问题

### 问题 1: 并发访问 panic

```go
// 错误 - 多个 goroutine 同时访问
var m = make(map[string]string)

go func() {
    m["key1"] = "value1"  // 可能 panic
}()

go func() {
    value := m["key1"]  // 可能 panic
}()

// 正确 - 使用 sync.Map 或加锁
var m = struct {
    sync.RWMutex
    data map[string]string
}{
    data: make(map[string]string),
}

go func() {
    m.Lock()
    m.data["key1"] = "value1"
    m.Unlock()
}()

go func() {
    m.RLock()
    value := m.data["key1"]
    m.RUnlock()
    _ = value
}()
```

### 问题 2: 映射作为函数参数

```go
// 传递映射时传递的是引用
func modifyMap(m map[string]string) {
    m["key"] = "value"  // 会修改原映射
}

// 如果不想修改原映射，先复制
func copyMap(m map[string]string) map[string]string {
    copy := make(map[string]string, len(m))
    for k, v := range m {
        copy[k] = v
    }
    return copy
}
```

### 问题 3: nil 映射

```go
// nil 映射可以读取，但不能写入
var m map[string]string

_ = m["key"]  // OK，返回零值

m["key"] = "value"  // panic!

// 始终初始化映射
m := make(map[string]string)
```

## 总结

映射数据结构在 Cobra-X 中：

1. **配置管理**: 存储和管理各种配置值
2. **去重处理**: 避免重复处理相同元素
3. **快速查找**: O(1) 时间复杂度的查找
4. **元数据存储**: 存储附加信息而不破坏原有结构
5. **灵活分组**: 支持按需分组和组织数据

映射是 Go 语言中最强大的数据结构之一，Cobra-X 充分利用了映射的特性来实现高效的数据管理和快速查找。合理使用映射可以大大简化代码并提高性能。
