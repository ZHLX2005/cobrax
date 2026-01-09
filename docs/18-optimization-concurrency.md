# 并发处理优化方案

## 概述

Cobra-X 目前主要采用串行处理方式，但在处理大量命令、复杂 UI 渲染和 I/O 操作时，可以通过并发处理显著提升性能。本文档分析当前实现并提出并发优化方案。

## 当前实现分析

### 串行处理瓶颈

```go
// 串行构建命令树
func BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
    // 递归处理，每个子命令串行处理
    children := getAvailableCommands(cmd.Commands())
    for _, child := range children {
        childItem := BuildCommandTree(child, currentPath)  // 串行
        if childItem != nil {
            item.Children = append(item.Children, childItem)
        }
    }
    return item
}

// 串行收集 flags
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem
    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            // 串行处理每个 flag
            item := createFlagItem(flag, current.Name())
            items = append(items, item)
        })
        current = c.wrapCommand(current.Parent())
    }
    return items
}
```

### 性能问题

1. **CPU 利用率低**: 单核处理，多核闲置
2. **阻塞等待**: I/O 操作阻塞整个流程
3. **无并行计算**: 独立的计算无法并行执行
4. **响应延迟**: UI 更新等待耗时操作完成

## 优化方案

### 方案 1: 并行命令树构建

#### 实现思路

使用 worker pool 并行构建独立的子树。

```go
// 并行构建命令树
type ParallelTreeBuilder struct {
    workers   int
    semaphore chan struct{}
    wg        sync.WaitGroup
}

func NewParallelTreeBuilder(workers int) *ParallelTreeBuilder {
    return &ParallelTreeBuilder{
        workers:   workers,
        semaphore: make(chan struct{}, workers),
    }
}

func (b *ParallelTreeBuilder) BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
    currentPath := path
    if path != "" {
        currentPath = path + " " + cmd.Name()
    } else {
        currentPath = cmd.Name()
    }

    isRunnable := cmd.Run != nil || cmd.RunE != nil

    item := &tui.CommandItem{
        ID:         cmd.Name(),
        Name:       cmd.Name(),
        Use:        cmd.Use,
        Short:      cmd.Short,
        Long:       cmd.Long,
        IsRunnable: isRunnable,
        Children:   make([]*tui.CommandItem, 0),
    }

    children := getAvailableCommands(cmd.Commands())
    if len(children) == 0 {
        return item
    }

    // 并行处理子节点
    resultChan := make(chan *tui.CommandItem, len(children))
    errChan := make(chan error, len(children))

    for _, child := range children {
        b.wg.Add(1)
        go func(c *spf13cobra.Command, p string) {
            defer b.wg.Done()
            b.semaphore <- struct{}{}        // 获取信号量
            defer func() { <-b.semaphore }() // 释放信号量

            childItem, err := b.buildWithRecover(c, p)
            if err != nil {
                errChan <- err
                return
            }
            resultChan <- childItem
        }(child, currentPath)
    }

    // 等待所有 goroutine 完成
    b.wg.Wait()
    close(resultChan)
    close(errChan)

    // 收集结果
    var errors []error
    for childItem := range resultChan {
        if childItem != nil {
            item.Children = append(item.Children, childItem)
        }
    }

    for err := range errChan {
        errors = append(errors, err)
    }

    if len(errors) > 0 {
        log.Printf("BuildCommandTree errors: %v", errors)
    }

    return item
}

func (b *ParallelTreeBuilder) buildWithRecover(cmd *spf13cobra.Command, path string) (*tui.CommandItem, error) {
    defer func() {
        if r := recover(); r != nil {
            log.Printf("Recovered in buildWithRecover: %v", r)
        }
    }()
    return b.BuildCommandTree(cmd, path), nil
}
```

**增强版 - 错误处理和重试**：

```go
type BuildTask struct {
    Command *spf13cobra.Command
    Path    string
    Retries int
}

func (b *ParallelTreeBuilder) BuildCommandTreeWithRetry(cmd *spf13cobra.Command, path string, maxRetries int) *tui.CommandItem {
    // ... 前面的代码相同

    for _, child := range children {
        task := BuildTask{
            Command: child,
            Path:    currentPath,
            Retries: maxRetries,
        }

        b.wg.Add(1)
        go func(t BuildTask) {
            defer b.wg.Done()
            b.semaphore <- struct{}{}
            defer func() { <-b.semaphore }()

            var lastErr error
            for i := 0; i <= t.Retries; i++ {
                childItem, err := b.buildWithRecover(t.Command, t.Path)
                if err == nil {
                    resultChan <- childItem
                    return
                }
                lastErr = err
                if i < t.Retries {
                    time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
                }
            }

            if lastErr != nil {
                errChan <- fmt.Errorf("after %d retries: %w", t.Retries, lastErr)
            }
        }(task)
    }

    // ... 其余代码
}
```

### 方案 2: 并行 Flag 收集

#### 实现思路

并行处理不同来源的 flags。

```go
// 并行收集 flags
func (c *Command) collectFlagItemsParallel(cmd *Command) []tui.FlagItem {
    // 收集所有需要处理的命令
    commands := make([]*spf13cobra.Command, 0)
    current := cmd
    for current != nil {
        commands = append(commands, current)
        if current.Parent() == nil {
            break
        }
        current = c.wrapCommand(current.Parent())
    }

    // 并行处理每个命令的 flags
    resultChan := make(chan []tui.FlagItem, len(commands))
    var wg sync.WaitGroup

    for _, current := range commands {
        wg.Add(1)
        go func(c *spf13cobra.Command) {
            defer wg.Done()

            var items []tui.FlagItem
            seen := make(map[string]bool)

            // 处理 LocalFlags
            c.LocalFlags().VisitAll(func(flag *pflag.Flag) {
                if shouldSkipFlag(flag.Name) || seen[flag.Name] {
                    return
                }

                item := createFlagItem(flag, c.Name())
                items = append(items, item)
                seen[flag.Name] = true
            })

            // 处理 PersistentFlags
            c.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
                if shouldSkipFlag(flag.Name) || seen[flag.Name] {
                    return
                }

                item := createFlagItem(flag, c.Name())
                items = append(items, item)
                seen[flag.Name] = true
            })

            resultChan <- items
        }(current)
    }

    // 等待所有 goroutine 完成
    wg.Wait()
    close(resultChan)

    // 合并结果并去重
    allItems := make([]tui.FlagItem, 0)
    seen := make(map[string]bool)

    for items := range resultChan {
        for _, item := range items {
            if !seen[item.Name] {
                allItems = append(allItems, item)
                seen[item.Name] = true
            }
        }
    }

    return allItems
}

func shouldSkipFlag(name string) bool {
    return strings.HasPrefix(name, "tui-") || name == "tui"
}

func createFlagItem(flag *pflag.Flag, sourceCommand string) tui.FlagItem {
    item := tui.FlagItem{
        Name:         flag.Name,
        ShortName:    flag.Shorthand,
        Description:  flag.Usage,
        DefaultValue: flag.DefValue,
        CurrentValue: flag.DefValue,
        Required:     false,
        SourceCommand: sourceCommand,
    }

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

    return item
}
```

### 方案 3: 异步 UI 更新

#### 实现思路

使用 channel 进行 UI 组件间通信。

```go
// 异步 UI 更新模型
type AsyncUIModel struct {
    // 核心数据
    items      []MenuItem
    cursor     int
    selected   bool

    // 异步更新
    updateChan chan tea.Msg
    commandCh  chan func() tea.Msg

    // 同步
    mutex sync.RWMutex
}

func NewAsyncUIModel(items []MenuItem) *AsyncUIModel {
    return &AsyncUIModel{
        items:      items,
        cursor:     0,
        updateChan: make(chan tea.Msg, 100),
        commandCh:  make(chan func() tea.Msg, 10),
    }
}

func (m *AsyncUIModel) Init() tea.Cmd {
    // 启动后台处理 goroutine
    return func() tea.Msg {
        go m.backgroundProcessor()
        return nil
    }
}

func (m *AsyncUIModel) backgroundProcessor() {
    for {
        select {
        case msg := <-m.updateChan:
            // 发送消息到 BubbleTea
            // 这需要在适当的地方处理
        case fn := <-m.commandCh:
            if fn != nil {
                msg := fn()
                // 发送消息
            }
        }
    }
}

func (m *AsyncUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        // 处理按键
        // ...

    case AsyncUpdateMsg:
        // 处理异步更新
        // ...
    }

    return m, nil
}

// 异步更新消息
type AsyncUpdateMsg struct {
    Update func(*AsyncUIModel)
}

// 触发异步更新
func (m *AsyncUIModel) TriggerUpdate(fn func(*AsyncUIModel)) tea.Cmd {
    return func() tea.Msg {
        m.updateChan <- AsyncUpdateMsg{Update: fn}
        return nil
    }
}
```

**实际应用 - 异步加载菜单项**：

```go
type LazyMenuModel struct {
    items        []MenuItem
    loading      bool
    cursor       int
    theme        *style.Theme
    width        int
    height       int

    // 异步加载
    loadChan     chan []MenuItem
    itemsLoader  func() ([]MenuItem, error)
    error        error
}

func (m *LazyMenuModel) Init() tea.Cmd {
    m.loading = true

    // 启动异步加载
    return func() tea.Msg {
        items, err := m.itemsLoader()
        if err != nil {
            return LoadFailedMsg{Error: err}
        }
        return LoadCompletedMsg{Items: items}
    }
}

func (m *LazyMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case LoadCompletedMsg:
        m.items = msg.Items
        m.loading = false

    case LoadFailedMsg:
        m.error = msg.Error
        m.loading = false

    // ... 其他消息处理
    }

    return m, nil
}

type LoadCompletedMsg struct {
    Items []MenuItem
}

type LoadFailedMsg struct {
    Error error
}
```

### 方案 4: 并行搜索

#### 实现思路

并行执行多个搜索查询。

```go
// 并行搜索实现
type ParallelSearch struct {
    workers int
}

func NewParallelSearch(workers int) *ParallelSearch {
    return &ParallelSearch{workers: workers}
}

func (s *ParallelSearch) Search(items []MenuItem, queries []string) []MenuItem {
    if len(queries) == 0 {
        return items
    }

    // 为每个查询启动一个 goroutine
    resultChan := make(chan []MenuItem, len(queries))
    var wg sync.WaitGroup

    for _, query := range queries {
        wg.Add(1)
        go func(q string) {
            defer wg.Done()

            var results []MenuItem
            for _, item := range items {
                if s.matchItem(item, q) {
                    results = append(results, item)
                }
            }

            resultChan <- results
        }(query)
    }

    // 等待所有搜索完成
    wg.Wait()
    close(resultChan)

    // 合并结果并去重
    seen := make(map[string]bool)
    var finalResults []MenuItem

    for results := range resultChan {
        for _, item := range results {
            if !seen[item.ID] {
                seen[item.ID] = true
                finalResults = append(finalResults, item)
            }
        }
    }

    return finalResults
}

func (s *ParallelSearch) matchItem(item MenuItem, query string) bool {
    query = strings.ToLower(query)

    // 匹配标签
    if strings.Contains(strings.ToLower(item.Label), query) {
        return true
    }

    // 匹配描述
    if strings.Contains(strings.ToLower(item.Description), query) {
        return true
    }

    // 匹配 ID
    if strings.Contains(strings.ToLower(item.ID), query) {
        return true
    }

    return false
}
```

### 方案 5: 并发安全的数据结构

#### 实现思路

使用并发安全的数据结构保护共享状态。

```go
// 并发安全的菜单项列表
type SafeMenuItems struct {
    items []MenuItem
    mutex sync.RWMutex
}

func NewSafeMenuItems() *SafeMenuItems {
    return &SafeMenuItems{
        items: make([]MenuItem, 0),
    }
}

func (s *SafeMenuItems) Add(item MenuItem) {
    s.mutex.Lock()
    defer s.mutex.Unlock()
    s.items = append(s.items, item)
}

func (s *SafeMenuItems) Get(index int) (MenuItem, bool) {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    if index < 0 || index >= len(s.items) {
        return MenuItem{}, false
    }
    return s.items[index], true
}

func (s *SafeMenuItems) List() []MenuItem {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    // 返回副本
    result := make([]MenuItem, len(s.items))
    copy(result, s.items)
    return result
}

func (s *SafeMenuItems) Filter(predicate func(MenuItem) bool) []MenuItem {
    s.mutex.RLock()
    defer s.mutex.RUnlock()

    result := make([]MenuItem, 0)
    for _, item := range s.items {
        if predicate(item) {
            result = append(result, item)
        }
    }
    return result
}

// 并发安全的搜索索引
type SearchIndex struct {
    index map[string][]int  // 关键词到菜单项索引的映射
    mutex sync.RWMutex
}

func NewSearchIndex() *SearchIndex {
    return &SearchIndex{
        index: make(map[string][]int),
    }
}

func (idx *SearchIndex) Build(items []MenuItem) {
    idx.mutex.Lock()
    defer idx.mutex.Unlock()

    // 清空旧索引
    idx.index = make(map[string][]int)

    // 构建新索引
    for i, item := range items {
        // 索引标签
        words := strings.Fields(strings.ToLower(item.Label))
        for _, word := range words {
            idx.index[word] = append(idx.index[word], i)
        }

        // 索引描述
        words = strings.Fields(strings.ToLower(item.Description))
        for _, word := range words {
            idx.index[word] = append(idx.index[word], i)
        }
    }
}

func (idx *SearchIndex) Search(query string) []int {
    idx.mutex.RLock()
    defer idx.mutex.RUnlock()

    query = strings.ToLower(query)
    words := strings.Fields(query)

    if len(words) == 0 {
        return nil
    }

    // 获取第一个词的索引
    result, ok := idx.index[words[0]]
    if !ok {
        return nil
    }

    // 对多个词进行交集运算
    for _, word := range words[1:] {
        indices, ok := idx.index[word]
        if !ok {
            return nil
        }
        result = intersect(result, indices)
        if len(result) == 0 {
            return nil
        }
    }

    return result
}

func intersect(a, b []int) []int {
    result := make([]int, 0)
    i, j := 0, 0

    for i < len(a) && j < len(b) {
        if a[i] == b[j] {
            result = append(result, a[i])
            i++
            j++
        } else if a[i] < b[j] {
            i++
        } else {
            j++
        }
    }

    return result
}
```

## 并发模式选择

### 1. Worker Pool 模式

```go
// 通用 worker pool
type WorkerPool struct {
    tasks   chan Task
    results chan Result
    workers int
    wg      sync.WaitGroup
}

type Task func() (interface{}, error)
type Result struct {
    Value interface{}
    Error error
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        tasks:   make(chan Task, 1000),
        results: make(chan Result, 1000),
        workers: workers,
    }
}

func (p *WorkerPool) Start() {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker()
    }
}

func (p *WorkerPool) worker() {
    defer p.wg.Done()

    for task := range p.tasks {
        value, err := task()
        p.results <- Result{Value: value, Error: err}
    }
}

func (p *WorkerPool) Submit(task Task) {
    p.tasks <- task
}

func (p *WorkerPool) Wait() {
    close(p.tasks)
    p.wg.Wait()
    close(p.results)
}

func (p *WorkerPool) GetResults() []Result {
    var results []Result
    for result := range p.results {
        results = append(results, result)
    }
    return results
}
```

### 2. Pipeline 模式

```go
// Pipeline 处理
type Pipeline struct {
    stages []StageFunc
}

type StageFunc func(interface{}) (interface{}, error)

func NewPipeline() *Pipeline {
    return &Pipeline{
        stages: make([]StageFunc, 0),
    }
}

func (p *Pipeline) AddStage(stage StageFunc) *Pipeline {
    p.stages = append(p.stages, stage)
    return p
}

func (p *Pipeline) Process(input interface{}) (interface{}, error) {
    var err error
    current := input

    for _, stage := range p.stages {
        current, err = stage(current)
        if err != nil {
            return nil, fmt.Errorf("pipeline stage failed: %w", err)
        }
    }

    return current, nil
}

// 并行 pipeline
func (p *Pipeline) ProcessParallel(inputs []interface{}) ([]interface{}, []error) {
    results := make([]interface{}, len(inputs))
    errors := make([]error, len(inputs))
    var wg sync.WaitGroup

    for i, input := range inputs {
        wg.Add(1)
        go func(idx int, data interface{}) {
            defer wg.Done()

            result, err := p.Process(data)
            results[idx] = result
            errors[idx] = err
        }(i, input)
    }

    wg.Wait()
    return results, errors
}
```

## 并发安全最佳实践

### 1. 避免死锁

```go
// 好的做法 - 使用 defer 确保释放锁
func (m *Model) Update(data Data) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    // 处理数据
}

// 避免 - 可能忘记解锁
func (m *Model) UpdateBad(data Data) {
    m.mutex.Lock()

    if someCondition {
        return  // 忘记解锁
    }

    m.mutex.Unlock()
}
```

### 2. 减少锁粒度

```go
// 好的做法 - 细粒度锁
type FineGrainedModel struct {
    data1 map[string]string
    data2 map[string]string
    lock1 sync.RWMutex
    lock2 sync.RWMutex
}

func (m *FineGrainedModel) GetData1(key string) string {
    m.lock1.RLock()
    defer m.lock1.RUnlock()
    return m.data1[key]
}

// 避免 - 粗粒度锁
type CoarseModel struct {
    data1 map[string]string
    data2 map[string]string
    lock  sync.RWMutex
}
```

### 3. 使用 channel 通信

```go
// 好的做法 - 使用 channel
func worker(tasks <-chan Task, results chan<- Result) {
    for task := range tasks {
        results <- task()
    }
}

// 避免 - 共享内存 + 锁
type WorkerBad struct {
    tasks   []Task
    results []Result
    mutex   sync.Mutex
}
```

## 性能对比

### 基准测试

```go
func BenchmarkBuildCommandTree(b *testing.B) {
    cmd := createTestTree(100)
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        BuildCommandTree(cmd, "")
    }
}

func BenchmarkBuildCommandTreeParallel(b *testing.B) {
    cmd := createTestTree(100)
    builder := NewParallelTreeBuilder(runtime.NumCPU())
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        builder.BuildCommandTree(cmd, "")
    }
}

func BenchmarkSearch(b *testing.B) {
    items := createTestItems(1000)
    queries := []string{"test", "search", "filter"}
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        searchItems(items, queries)
    }
}

func BenchmarkSearchParallel(b *testing.B) {
    items := createTestItems(1000)
    queries := []string{"test", "search", "filter"}
    searcher := NewParallelSearch(runtime.NumCPU())
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        searcher.Search(items, queries)
    }
}
```

### 预期性能提升

| 操作 | 串行 | 并行 (4核) | 并行 (8核) | 加速比 |
|------|------|-----------|-----------|--------|
| 构建命令树 (100项) | 10ms | 3ms | 2ms | 3-5x |
| 收集 flags (50项) | 2ms | 0.8ms | 0.5ms | 2-4x |
| 搜索 (1000项, 5查询) | 50ms | 15ms | 10ms | 3-5x |
| UI 更新 | 16ms | 5ms | 4ms | 3-4x |

## 推荐实现

### 渐进式并发优化

```go
// 并发优化配置
type ConcurrencyConfig struct {
    EnableParallelBuild    bool
    EnableParallelSearch  bool
    EnableAsyncUI         bool
    WorkerCount           int
    EnableProfiling       bool
}

var DefaultConcurrencyConfig = &ConcurrencyConfig{
    EnableParallelBuild:   true,
    EnableParallelSearch: true,
    EnableAsyncUI:        false,  // 默认关闭，需要更仔细的测试
    WorkerCount:          runtime.NumCPU(),
    EnableProfiling:      false,
}

// 应用并发优化
func ApplyConcurrencyOptimizations(config *ConcurrencyConfig) {
    if config.EnableParallelBuild {
        UseParallelTreeBuilder(config.WorkerCount)
    }

    if config.EnableParallelSearch {
        UseParallelSearch(config.WorkerCount)
    }

    if config.EnableProfiling {
        EnableConcurrencyProfiling()
    }
}
```

## 总结

并发处理优化要点：

1. **并行构建**: 独立子任务并行执行
2. **异步 UI**: 不阻塞 UI 更新
3. **Worker Pool**: 控制 goroutine 数量
4. **并发安全**: 正确使用锁和 channel
5. **渐进式**: 从简单场景开始优化

并发优化可以显著提升性能，但也增加了复杂度。建议从并行构建和搜索开始，在充分测试后引入更复杂的并发模式。始终注意并发安全和正确的错误处理。
