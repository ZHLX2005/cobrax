# 内存管理优化方案

## 概述

在 Cobra-X 项目中，内存管理是一个重要但常被忽视的优化点。不合理的内存使用会导致 GC 压力增加、内存占用过高和性能下降。本文档分析当前实现并提出内存优化方案。

## 当前内存使用分析

### 内存分配热点

```go
// 问题 1: 大量临时字符串分配
func (c *Command) buildCommandString(path []*Command) string {
    var parts []string
    seen := make(map[string]bool)

    // 每次都创建新的字符串
    for _, cmd := range path {
        parts = append(parts, cmd.Use)
    }

    // 多次字符串拼接
    for _, cmd := range path {
        cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if flag.Changed && flag.Name != "help" && !seen[flag.Name] {
                key := fmt.Sprintf("--%s", flag.Name)  // 新分配
                if flag.Value.Type() == "bool" {
                    if flag.Value.String() == "true" {
                        parts = append(parts, key)
                    }
                } else {
                    // 又一次字符串分配
                    parts = append(parts, fmt.Sprintf("%s=%s", key, flag.Value.String()))
                }
                seen[flag.Name] = true
            }
        })
    }

    return strings.Join(parts, " ")
}
```

### 问题 2: 重复的样式对象创建

```go
// 每次渲染都创建新样式
func (m *menuModel) View() string {
    // 每次都创建新的样式对象
    titleStyle := lipgloss.NewStyle().
        Bold(true).
        Foreground(m.theme.Colors.Primary).
        MarginTop(1).
        MarginBottom(1)

    borderStyle := lipgloss.NewStyle().
        Border(m.theme.Styles.Border).
        BorderForeground(m.theme.Colors.Secondary).
        Padding(m.theme.Layout.Padding[0], m.theme.Layout.Padding[1]).
        Width(m.width - 4)

    // ... 更多样式创建
}
```

### 问题 3: 切片和映射的频繁分配

```go
// 频繁的切片分配
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem
    seen := make(map[string]bool)  // 每次都创建新映射

    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            // 每次都创建新的 FlagItem
            item := tui.FlagItem{
                Name:         flag.Name,
                ShortName:    flag.Shorthand,
                Description:  flag.Usage,
                DefaultValue: flag.DefValue,
                CurrentValue: flag.DefValue,
                Required:     false,
                SourceCommand: current.Name(),
            }
            items = append(items, item)
            seen[flag.Name] = true
        })
        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

## 优化方案

### 方案 1: 对象池化

#### 实现思路

使用 sync.Pool 重用对象，减少 GC 压力。

```go
// FlagItem 对象池
var flagItemPool = sync.Pool{
    New: func() interface{} {
        return &tui.FlagItem{}
    },
}

func getFlagItem() *tui.FlagItem {
    return flagItemPool.Get().(*tui.FlagItem)
}

func putFlagItem(item *tui.FlagItem) {
    // 重置对象
    *item = tui.FlagItem{}
    flagItemPool.Put(item)
}

// 使用对象池的 flag 收集
func (c *Command) collectFlagItemsPooled(cmd *Command) []tui.FlagItem {
    items := make([]tui.FlagItem, 0, 32)  // 预分配容量
    seen := make(map[string]bool, 32)

    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if strings.HasPrefix(flag.Name, "tui-") || flag.Name == "tui" || seen[flag.Name] {
                return
            }

            // 从池中获取对象
            item := getFlagItem()
            item.Name = flag.Name
            item.ShortName = flag.Shorthand
            item.Description = flag.Usage
            item.DefaultValue = flag.DefValue
            item.CurrentValue = flag.DefValue
            item.Required = false
            item.SourceCommand = current.Name()

            // 确定 flag 类型
            switch flag.Value.Type() {
            case "bool":
                item.Type = tui.FlagTypeBool
            case "int", "int32", "int64":
                item.Type = tui.FlagTypeInt
            case "duration":
                item.Type = tui.FlagTypeDuration
            default:
                item.Type = tui.FlagTypeString
            }

            items = append(items, *item)
            // 注意：不归还到池，因为 items 持有引用
            seen[flag.Name] = true
        })

        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

**更复杂的池化 - 批量归还**：

```go
// 批量对象池
type FlagItemBatch struct {
    items []*tui.FlagItem
    count int
}

var flagItemBatchPool = sync.Pool{
    New: func() interface{} {
        return &FlagItemBatch{
            items: make([]*tui.FlagItem, 32),
        }
    },
}

func GetFlagItemBatch(size int) *FlagItemBatch {
    batch := flagItemBatchPool.Get().(*FlagItemBatch)
    if cap(batch.items) < size {
        batch.items = make([]*tui.FlagItem, size)
    }
    batch.count = size
    return batch
}

func (b *FlagItemBatch) Get(index int) *tui.FlagItem {
    if b.items[index] == nil {
        b.items[index] = &tui.FlagItem{}
    }
    return b.items[index]
}

func (b *FlagItemBatch) Return() {
    // 重置所有项
    for i := 0; i < b.count; i++ {
        *b.items[i] = tui.FlagItem{}
    }
    flagItemBatchPool.Put(b)
}
```

### 方案 2: 字符串构建优化

#### 实现思路

使用 strings.Builder 和预分配减少字符串分配。

```go
// 优化的命令字符串构建
func (c *Command) buildCommandStringOptimized(path []*Command) string {
    // 预估容量
    estimatedSize := 0
    for _, cmd := range path {
        estimatedSize += len(cmd.Use) + 1  // +1 for space
    }

    var builder strings.Builder
    builder.Grow(estimatedSize)

    seen := make(map[string]bool, 32)

    // 添加命令路径
    for i, cmd := range path {
        if i > 0 {
            builder.WriteByte(' ')
        }
        builder.WriteString(cmd.Use)
    }

    // 添加 flags
    for _, cmd := range path {
        cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if flag.Changed && flag.Name != "help" && !seen[flag.Name] {
                builder.WriteString(" --")
                builder.WriteString(flag.Name)

                if flag.Value.Type() != "bool" || flag.Value.String() != "true" {
                    builder.WriteByte('=')
                    builder.WriteString(flag.Value.String())
                }

                seen[flag.Name] = true
            }
        })

        // 类似处理 PersistentFlags
    }

    return builder.String()
}
```

**进一步优化 - 重用 builder**：

```go
// Builder 池
var builderPool = sync.Pool{
    New: func() interface{} {
        b := &strings.Builder{}
        b.Grow(512)  // 预分配初始容量
        return b
    },
}

func getBuilder() *strings.Builder {
    return builderPool.Get().(*strings.Builder)
}

func putBuilder(b *strings.Builder) {
    b.Reset()
    builderPool.Put(b)
}

func (c *Command) buildCommandStringWithPool(path []*Command) string {
    builder := getBuilder()
    defer putBuilder(builder)

    // 使用 builder 构建字符串
    // ...

    return builder.String()
}
```

### 方案 3: 享元模式

#### 实现思路

共享相同的不可变对象，减少重复。

```go
// 享元样式工厂
type StyleFlyweightFactory struct {
    styles map[string]lipgloss.Style
    mutex  sync.RWMutex
}

func NewStyleFlyweightFactory() *StyleFlyweightFactory {
    return &StyleFlyweightFactory{
        styles: make(map[string]lipgloss.Style),
    }
}

func (f *StyleFlyweightFactory) GetStyle(key string, builder func() lipgloss.Style) lipgloss.Style {
    f.mutex.RLock()
    if style, ok := f.styles[key]; ok {
        f.mutex.RUnlock()
        return style
    }
    f.mutex.RUnlock()

    f.mutex.Lock()
    defer f.mutex.Unlock()

    // 双重检查
    if style, ok := f.styles[key]; ok {
        return style
    }

    style := builder()
    f.styles[key] = style
    return style
}

// 全局样式工厂
var globalStyleFactory = NewStyleFlyweightFactory()

// 使用享元样式
func (m *menuModel) ViewWithFlyweight() string {
    if m.quitting {
        return ""
    }

    // 使用享元样式
    titleStyle := globalStyleFactory.GetStyle("title", func() lipgloss.Style {
        return lipgloss.NewStyle().
            Bold(true).
            Foreground(m.theme.Colors.Primary).
            MarginTop(1).
            MarginBottom(1)
    })

    borderStyle := globalStyleFactory.GetStyle("border", func() lipgloss.Style {
        return lipgloss.NewStyle().
            Border(m.theme.Styles.Border).
            BorderForeground(m.theme.Colors.Secondary).
            Padding(m.theme.Layout.Padding[0], m.theme.Layout.Padding[1]).
            Width(m.width - 4)
    })

    // ... 使用样式
}
```

### 方案 4: 延迟分配

#### 实现思路

只在需要时分配内存。

```go
// 延迟分配的模型
type LazyFormModel struct {
    items      []FlagItem
    cursor     int
    values     map[string]string  // 延迟初始化
    cancelled  bool
    theme      *style.Theme
    width      int
    height     int
    quitting   bool
    valuesOnce sync.Once
}

func (m *LazyFormModel) getValues() map[string]string {
    m.valuesOnce.Do(func() {
        m.values = make(map[string]string, len(m.items))
        for _, item := range m.items {
            m.values[item.Name] = item.DefaultValue
        }
    })
    return m.values
}

func (m *LazyFormModel) SetValue(name, value string) {
    values := m.getValues()
    values[name] = value
}

func (m *LazyFormModel) GetValue(name string) string {
    values := m.getValues()
    if value, ok := values[name]; ok {
        return value
    }
    // 查找默认值
    for _, item := range m.items {
        if item.Name == name {
            return item.DefaultValue
        }
    }
    return ""
}
```

### 方案 5: 内存复用

#### 实现思路

复用缓冲区和临时对象。

```go
// 缓冲区池
type BufferPool struct {
    pools map[int]*sync.Pool  // 按大小分类的池
    mutex sync.RWMutex
}

func NewBufferPool() *BufferPool {
    return &BufferPool{
        pools: make(map[int]*sync.Pool),
    }
}

func (p *BufferPool) Get(size int) []byte {
    p.mutex.Lock()
    defer p.mutex.Unlock()

    // 找到最接近的池大小
    poolSize := 1 << uint(ceilLog2(size))
    pool, ok := p.pools[poolSize]

    if !ok {
        pool = &sync.Pool{
            New: func() interface{} {
                return make([]byte, poolSize)
            },
        }
        p.pools[poolSize] = pool
    }

    buf := pool.Get().([]byte)
    return buf[:size]  // 返回所需大小
}

func (p *BufferPool) Put(buf []byte) {
    size := cap(buf)
    p.mutex.RLock()
    pool, ok := p.pools[size]
    p.mutex.RUnlock()

    if ok {
        pool.Put(buf)
    }
}

func ceilLog2(n int) int {
    if n <= 1 {
        return 0
    }
    log := 0
    for n > 1 {
        n >>= 1
        log++
    }
    return log
}

// 使用
var globalBufferPool = NewBufferPool()

func readWithBuffer(r io.Reader) ([]byte, error) {
    // 读取部分数据确定大小
    buf := make([]byte, 512)
    n, err := r.Read(buf)
    if err != nil {
        return nil, err
    }

    // 使用池化的缓冲区
    pooledBuf := globalBufferPool.Get(n)
    copy(pooledBuf, buf[:n])
    defer globalBufferPool.Put(pooledBuf)

    return pooledBuf, nil
}
```

## 内存监控

### 内存使用追踪

```go
// 内存监控器
type MemoryMonitor struct {
    lastAlloc   uint64
    peakAlloc   uint64
    sampleCount int
    mutex       sync.RWMutex
}

var globalMemMonitor = &MemoryMonitor{}

func (m *MemoryMonitor) Start() {
    go m.monitor()
}

func (m *MemoryMonitor) monitor() {
    ticker := time.NewTicker(time.Second)
    defer ticker.Stop()

    for range ticker.C {
        var memStats runtime.MemStats
        runtime.ReadMemStats(&memStats)

        m.mutex.Lock()
        currentAlloc := memStats.Alloc

        if m.lastAlloc == 0 {
            m.lastAlloc = currentAlloc
        }

        if currentAlloc > m.peakAlloc {
            m.peakAlloc = currentAlloc
        }

        m.lastAlloc = currentAlloc
        m.sampleCount++
        m.mutex.Unlock()

        // 可选：记录到日志
        log.Printf("Memory: %d MB (Peak: %d MB)",
            currentAlloc/1024/1024,
            m.peakAlloc/1024/1024)
    }
}

func (m *MemoryMonitor) GetStats() (current, peak uint64) {
    m.mutex.RLock()
    defer m.mutex.RUnlock()
    return m.lastAlloc, m.peakAlloc
}

// 内存分析器
func ProfileMemory(duration time.Duration, filename string) error {
    f, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer f.Close()

    if err := pprof.StartCPUProfile(f); err != nil {
        return err
    }

    time.Sleep(duration)

    pprof.StopCPUProfile()
    return nil
}
```

### 内存泄漏检测

```go
// 引用追踪器
type RefTracker struct {
    refs map[string]int
    mutex sync.RWMutex
}

func NewRefTracker() *RefTracker {
    return &RefTracker{
        refs: make(map[string]int),
    }
}

func (t *RefTracker) Track(name string) {
    t.mutex.Lock()
    defer t.mutex.Unlock()
    t.refs[name]++
}

func (t *RefTracker) Untrack(name string) {
    t.mutex.Lock()
    defer t.mutex.Unlock()

    if t.refs[name] > 0 {
        t.refs[name]--
    }
}

func (t *RefTracker) GetRefs(name string) int {
    t.mutex.RLock()
    defer t.mutex.RUnlock()
    return t.refs[name]
}

func (t *RefTracker) ReportLeaks() []string {
    t.mutex.RLock()
    defer t.mutex.RUnlock()

    leaks := make([]string, 0)
    for name, count := range t.refs {
        if count > 0 {
            leaks = append(leaks, fmt.Sprintf("%s: %d refs", name, count))
        }
    }
    return leaks
}
```

## 内存优化最佳实践

### 1. 预分配容量

```go
// 好的做法 - 预分配
items := make([]MenuItem, 0, len(commands))

// 避免 - 动态扩容
var items []MenuItem
for _, cmd := range commands {
    items = append(items, cmd)  // 多次扩容
}
```

### 2. 避免不必要的复制

```go
// 好的做法 - 使用切片
func process(items []MenuItem) {
    for _, item := range items {
        // 处理
    }
}

// 避免 - 复制到数组
func processBad(items []MenuItem) {
    arr := [100]MenuItem{}
    copy(arr[:], items)  // 不必要的复制
}
```

### 3. 及时释放大对象

```go
// 处理完大对象后显式释放
func processLargeData() {
    data := loadLargeData()

    // 处理数据
    processData(data)

    // 显式释放
    data = nil

    // 可选：触发 GC
    runtime.GC()
}
```

### 4. 使用指针减少复制

```go
// 好的做法 - 使用指针
func processItems(items []*MenuItem) {
    for _, item := range items {
        // 处理 item
    }
}

// 避免 - 按值传递
func processItemsBad(items []MenuItem) {
    for _, item := range items {
        // 复制整个结构体
    }
}
```

## 性能对比

### 内存分配测试

```go
func BenchmarkBuildCommandString(b *testing.B) {
    path := createTestPath(10)  // 10 个命令
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        buildCommandString(path)
    }
}

func BenchmarkBuildCommandStringOptimized(b *testing.B) {
    path := createTestPath(10)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        buildCommandStringOptimized(path)
    }
}

func BenchmarkBuildCommandStringWithPool(b *testing.B) {
    path := createTestPath(10)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        buildCommandStringWithPool(path)
    }
}
```

### 预期内存减少

| 场景 | 原始实现 | 优化后 | 减少 |
|------|----------|--------|------|
| 命令字符串构建 | 1024 B | 512 B | 50% |
| 样式对象创建 | 200 B | 20 B | 90% |
| Flag 列表收集 | 2048 B | 1024 B | 50% |
| 表单值存储 | 512 B | 256 B | 50% |

## 推荐实现

### 内存优化配置

```go
// 内存优化配置
type MemoryOptConfig struct {
    UseObjectPool    bool
    UseBuilderPool   bool
    UseStyleFlyweight bool
    UseLazyInit      bool
    MaxCacheSize     int
}

var DefaultMemoryOptConfig = &MemoryOptConfig{
    UseObjectPool:    true,
    UseBuilderPool:   true,
    UseStyleFlyweight: true,
    UseLazyInit:      true,
    MaxCacheSize:     1000,
}

// 应用优化配置
func ApplyMemoryOptimizations(config *MemoryOptConfig) {
    if config.UseObjectPool {
        initObjectPools()
    }

    if config.UseBuilderPool {
        initBuilderPools()
    }

    if config.UseStyleFlyweight {
        initStyleFactory()
    }

    if config.UseLazyInit {
        enableLazyInit()
    }
}
```

## 总结

内存管理优化要点：

1. **对象池化**: 减少 GC 压力
2. **字符串优化**: 使用 Builder 和池
3. **享元模式**: 共享不可变对象
4. **延迟分配**: 按需分配内存
5. **及时释放**: 处理完大对象后释放

通过合理的内存管理，可以显著降低 GC 压力和内存占用，提升应用的整体性能。建议从对象池化和字符串优化开始，逐步引入其他优化技术。
