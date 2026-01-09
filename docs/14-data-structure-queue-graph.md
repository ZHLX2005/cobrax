# 队列与图数据结构 (Queue and Graph Data Structures)

## 第一部分：队列 (Queue)

## 概述

队列是一种先进先出（FIFO）的数据结构。在 Cobra-X 中，队列主要用于管理 TUI 程序实例、处理命令导航和实现事件系统。

## 在 Cobra-X 中的应用

### 1. 程序队列

```go
// tui/default_renderer.go:13-18
type DefaultRenderer struct {
    theme    *style.Theme
    programs []*tea.Program  // 简单的程序队列
}

func (r *DefaultRenderer) RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error) {
    width, height := getTerminalSize()

    model := newMenuModel(options, r.theme, width, height)

    p := tea.NewProgram(model, tea.WithAltScreen())
    r.programs = append(r.programs, p)  // 入队

    result, err := p.Run()
    // ...
}
```

**队列特点**：
- **顺序管理**: 按创建顺序管理程序
- **生命周期追踪**: 追踪所有活动的 TUI 程序
- **清理资源**: 统一管理和清理资源

## 队列的实现

### 1. 基础队列

```go
// 命令队列
type CommandQueue struct {
    commands []*Command
    mutex    sync.Mutex
}

func NewCommandQueue() *CommandQueue {
    return &CommandQueue{
        commands: make([]*Command, 0),
    }
}

// 入队
func (q *CommandQueue) Enqueue(cmd *Command) {
    q.mutex.Lock()
    defer q.mutex.Unlock()
    q.commands = append(q.commands, cmd)
}

// 出队
func (q *CommandQueue) Dequeue() (*Command, bool) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    if len(q.commands) == 0 {
        return nil, false
    }

    cmd := q.commands[0]
    q.commands = q.commands[1:]
    return cmd, true
}

// 查看队首元素
func (q *CommandQueue) Peek() (*Command, bool) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    if len(q.commands) == 0 {
        return nil, false
    }
    return q.commands[0], true
}

// 队列大小
func (q *CommandQueue) Size() int {
    q.mutex.Lock()
    defer q.mutex.Unlock()
    return len(q.commands)
}

// 是否为空
func (q *CommandQueue) IsEmpty() bool {
    q.mutex.Lock()
    defer q.mutex.Unlock()
    return len(q.commands) == 0
}
```

### 2. 循环队列

```go
// 更高效的循环队列实现
type CircularQueue struct {
    items []interface{}
    head  int
    tail  int
    size  int
    mutex sync.Mutex
}

func NewCircularQueue(capacity int) *CircularQueue {
    return &CircularQueue{
        items: make([]interface{}, capacity),
    }
}

func (q *CircularQueue) Enqueue(item interface{}) bool {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    if q.size == len(q.items) {
        return false  // 队列已满
    }

    q.items[q.tail] = item
    q.tail = (q.tail + 1) % len(q.items)
    q.size++
    return true
}

func (q *CircularQueue) Dequeue() (interface{}, bool) {
    q.mutex.Lock()
    defer q.mutex.Unlock()

    if q.size == 0 {
        return nil, false
    }

    item := q.items[q.head]
    q.head = (q.head + 1) % len(q.items)
    q.size--
    return item, true
}
```

### 3. 优先级队列

```go
import (
    "container/heap"
)

// 优先级队列（用于命令优先级管理）
type PriorityQueue struct {
    items []*Command
    mutex sync.Mutex
}

func NewPriorityQueue() *PriorityQueue {
    pq := &PriorityQueue{
        items: make([]*Command, 0),
    }
    heap.Init(pq)
    return pq
}

func (pq *PriorityQueue) Len() int { return len(pq.items) }

func (pq *PriorityQueue) Less(i, j int) bool {
    // 根据命令优先级比较
    return pq.items[i].Priority < pq.items[j].Priority
}

func (pq *PriorityQueue) Swap(i, j int) {
    pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
    pq.mutex.Lock()
    defer pq.mutex.Unlock()
    item := x.(*Command)
    pq.items = append(pq.items, item)
}

func (pq *PriorityQueue) Pop() interface{} {
    pq.mutex.Lock()
    defer pq.mutex.Unlock()

    old := pq.items
    n := len(old)
    item := old[n-1]
    pq.items = old[0 : n-1]
    return item
}

func (pq *PriorityQueue) Enqueue(cmd *Command) {
    heap.Push(pq, cmd)
}

func (pq *PriorityQueue) Dequeue() *Command {
    return heap.Pop(pq).(*Command)
}
```

## 队列的应用场景

### 1. 命令历史队列

```go
// 命令执行历史
type CommandHistory struct {
    history []string
    max     int
    mutex   sync.Mutex
}

func NewCommandHistory(maxSize int) *CommandHistory {
    return &CommandHistory{
        history: make([]string, 0, maxSize),
        max:     maxSize,
    }
}

func (h *CommandHistory) Add(cmd string) {
    h.mutex.Lock()
    defer h.mutex.Unlock()

    h.history = append(h.history, cmd)
    if len(h.history) > h.max {
        h.history = h.history[1:]  // 移除最旧的
    }
}

func (h *CommandHistory) Get(index int) (string, bool) {
    h.mutex.Lock()
    defer h.mutex.Unlock()

    if index < 0 || index >= len(h.history) {
        return "", false
    }
    return h.history[index], true
}

func (h *CommandHistory) List() []string {
    h.mutex.Lock()
    defer h.mutex.Unlock()

    result := make([]string, len(h.history))
    copy(result, h.history)
    return result
}
```

### 2. 异步命令队列

```go
// 异步执行命令的队列
type AsyncCommandQueue struct {
    queue    chan *Command
    workers  int
    wg       sync.WaitGroup
    handler  func(*Command) error
    logger   *log.Logger
}

func NewAsyncCommandQueue(workers int, queueSize int, handler func(*Command) error) *AsyncCommandQueue {
    return &AsyncCommandQueue{
        queue:   make(chan *Command, queueSize),
        workers: workers,
        handler: handler,
    }
}

func (q *AsyncCommandQueue) Start() {
    for i := 0; i < q.workers; i++ {
        q.wg.Add(1)
        go q.worker()
    }
}

func (q *AsyncCommandQueue) worker() {
    defer q.wg.Done()

    for cmd := range q.queue {
        if err := q.handler(cmd); err != nil {
            q.log("Command %s failed: %v", cmd.Use, err)
        } else {
            q.log("Command %s completed", cmd.Use)
        }
    }
}

func (q *AsyncCommandQueue) Enqueue(cmd *Command) error {
    select {
    case q.queue <- cmd:
        return nil
    default:
        return errors.New("queue is full")
    }
}

func (q *AsyncCommandQueue) Stop() {
    close(q.queue)
    q.wg.Wait()
}

func (q *AsyncCommandQueue) log(format string, args ...interface{}) {
    if q.logger != nil {
        q.logger.Printf(format, args...)
    }
}
```

### 3. BFS 遍历队列

```go
// 使用队列实现广度优先遍历
func BFSTraverse(root *CommandItem, fn func(*CommandItem)) {
    if root == nil {
        return
    }

    queue := []*CommandItem{root}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        fn(current)

        // 添加子节点到队列
        queue = append(queue, current.Children...)
    }
}
```

---

## 第二部分：图 (Graph)

## 概述

图是由节点和边组成的数据结构。在 Cobra-X 中，命令之间的依赖关系形成了一个有向图，图结构用于表示复杂的命令依赖和执行关系。

## 在 Cobra-X 中的应用

### 1. 命令依赖图

```go
// 命令依赖关系
type CommandGraph struct {
    nodes map[string]*CommandNode
    edges map[string][]string  // 依赖关系
    mutex sync.RWMutex
}

type CommandNode struct {
    ID          string
    Command     *spf13cobra.Command
    Dependencies []string
    Dependents  []string
}

func NewCommandGraph() *CommandGraph {
    return &CommandGraph{
        nodes: make(map[string]*CommandNode),
        edges: make(map[string][]string),
    }
}

// 添加节点
func (g *CommandGraph) AddNode(cmd *spf13cobra.Command) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    id := cmd.Name()
    if _, exists := g.nodes[id]; !exists {
        g.nodes[id] = &CommandNode{
            ID:      id,
            Command: cmd,
        }
    }
}

// 添加边（依赖关系）
func (g *CommandGraph) AddEdge(from, to string) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    // 确保节点存在
    if _, exists := g.nodes[from]; !exists {
        g.nodes[from] = &CommandNode{ID: from}
    }
    if _, exists := g.nodes[to]; !exists {
        g.nodes[to] = &CommandNode{ID: to}
    }

    // 添加依赖关系
    g.edges[from] = append(g.edges[from], to)

    // 更新节点信息
    fromNode := g.nodes[from]
    toNode := g.nodes[to]

    fromNode.Dependencies = append(fromNode.Dependencies, to)
    toNode.Dependents = append(toNode.Dependents, from)
}
```

### 2. 图的遍历

```go
// DFS 遍历
func (g *CommandGraph) DFS(start string, fn func(*CommandNode)) {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    visited := make(map[string]bool)
    g.dfsHelper(start, visited, fn)
}

func (g *CommandGraph) dfsHelper(nodeID string, visited map[string]bool, fn func(*CommandNode)) {
    if visited[nodeID] {
        return
    }

    visited[nodeID] = true

    if node, ok := g.nodes[nodeID]; ok {
        fn(node)

        // 遍历依赖的节点
        for _, depID := range node.Dependencies {
            g.dfsHelper(depID, visited, fn)
        }
    }
}

// BFS 遍历
func (g *CommandGraph) BFS(start string, fn func(*CommandNode)) {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    visited := make(map[string]bool)
    queue := []string{start}

    for len(queue) > 0 {
        nodeID := queue[0]
        queue = queue[1:]

        if visited[nodeID] {
            continue
        }
        visited[nodeID] = true

        if node, ok := g.nodes[nodeID]; ok {
            fn(node)

            // 添加依赖节点到队列
            for _, depID := range node.Dependencies {
                if !visited[depID] {
                    queue = append(queue, depID)
                }
            }
        }
    }
}
```

### 3. 拓扑排序

```go
// 拓扑排序（用于确定命令执行顺序）
func (g *CommandGraph) TopologicalSort() ([]string, error) {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    // 计算入度
    inDegree := make(map[string]int)
    for _, node := range g.nodes {
        inDegree[node.ID] = 0
    }

    for from, toList := range g.edges {
        for _, to := range toList {
            inDegree[to]++
        }
    }

    // 找到所有入度为 0 的节点
    queue := make([]string, 0)
    for id, degree := range inDegree {
        if degree == 0 {
            queue = append(queue, id)
        }
    }

    // 拓扑排序
    result := make([]string, 0)
    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]
        result = append(result, current)

        // 减少依赖节点的入度
        for _, dependent := range g.nodes[current].Dependents {
            inDegree[dependent]--
            if inDegree[dependent] == 0 {
                queue = append(queue, dependent)
            }
        }
    }

    // 检查是否有环
    if len(result) != len(g.nodes) {
        return nil, errors.New("graph has cycle")
    }

    return result, nil
}
```

### 4. 环检测

```go
// 检测图中是否有环
func (g *CommandGraph) HasCycle() bool {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    for _, node := range g.nodes {
        if !visited[node.ID] {
            if g.hasCycleHelper(node.ID, visited, recStack) {
                return true
            }
        }
    }

    return false
}

func (g *CommandGraph) hasCycleHelper(nodeID string, visited, recStack map[string]bool) bool {
    visited[nodeID] = true
    recStack[nodeID] = true

    node := g.nodes[nodeID]
    for _, depID := range node.Dependencies {
        if !visited[depID] {
            if g.hasCycleHelper(depID, visited, recStack) {
                return true
            }
        } else if recStack[depID] {
            return true
        }
    }

    recStack[nodeID] = false
    return false
}
```

## 图的应用场景

### 1. 命令执行顺序

```go
// 根据依赖关系确定执行顺序
type CommandExecutor struct {
    graph *CommandGraph
}

func (e *CommandExecutor) ExecuteCommands() error {
    // 获取拓扑排序
    order, err := e.graph.TopologicalSort()
    if err != nil {
        return fmt.Errorf("invalid command dependencies: %w", err)
    }

    // 按顺序执行命令
    for _, cmdID := range order {
        node := e.graph.nodes[cmdID]
        if err := e.executeCommand(node.Command); err != nil {
            return fmt.Errorf("command %s failed: %w", cmdID, err)
        }
    }

    return nil
}

func (e *CommandExecutor) executeCommand(cmd *spf13cobra.Command) error {
    // 执行命令逻辑
    return cmd.RunE(cmd, cmd.Flags().Args())
}
```

### 2. 命令链可视化

```go
// 生成命令依赖的 Mermaid 图
func (g *CommandGraph) GenerateMermaidDiagram() string {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    var sb strings.Builder
    sb.WriteString("graph TD\n")

    // 添加节点
    for id, node := range g.nodes {
        sb.WriteString(fmt.Sprintf("  %s[%s]\n", id, id))
    }

    // 添加边
    for from, toList := range g.edges {
        for _, to := range toList {
            sb.WriteString(fmt.Sprintf("  %s --> %s\n", from, to))
        }
    }

    return sb.String()
}

// 输出示例：
// graph TD
//   root[root]
//   server[server]
//   start[start]
//   root --> server
//   server --> start
```

### 3. 最短路径查找

```go
// 查找从一个命令到另一个命令的最短路径
func (g *CommandGraph) ShortestPath(from, to string) ([]string, error) {
    g.mutex.RLock()
    defer g.mutex.RUnlock()

    if _, ok := g.nodes[from]; !ok {
        return nil, fmt.Errorf("node %s not found", from)
    }
    if _, ok := g.nodes[to]; !ok {
        return nil, fmt.Errorf("node %s not found", to)
    }

    // BFS 查找最短路径
    visited := make(map[string]bool)
    parent := make(map[string]string)
    queue := []string{from}
    visited[from] = true

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        if current == to {
            // 构建路径
            path := make([]string, 0)
            node := to
            for node != "" {
                path = append([]string{node}, path...)
                node = parent[node]
            }
            return path, nil
        }

        // 遍历邻居
        for _, neighbor := range g.nodes[current].Dependents {
            if !visited[neighbor] {
                visited[neighbor] = true
                parent[neighbor] = current
                queue = append(queue, neighbor)
            }
        }

        for _, neighbor := range g.nodes[current].Dependencies {
            if !visited[neighbor] {
                visited[neighbor] = true
                parent[neighbor] = current
                queue = append(queue, neighbor)
            }
        }
    }

    return nil, errors.New("no path found")
}
```

## 图的优化

### 1. 邻接表优化

```go
// 使用邻接表优化图存储
type OptimizedGraph struct {
    adjacencyList map[string]*AdjacencyList
    mutex        sync.RWMutex
}

type AdjacencyList struct {
    node         *CommandNode
    neighbors    []*CommandNode
    neighborLock sync.RWMutex
}

func (g *OptimizedGraph) AddEdge(from, to string) {
    g.mutex.Lock()
    defer g.mutex.Unlock()

    // 确保邻接表存在
    if _, ok := g.adjacencyList[from]; !ok {
        g.adjacencyList[from] = &AdjacencyList{
            node:      &CommandNode{ID: from},
            neighbors: make([]*CommandNode, 0),
        }
    }
    if _, ok := g.adjacencyList[to]; !ok {
        g.adjacencyList[to] = &AdjacencyList{
            node:      &CommandNode{ID: to},
            neighbors: make([]*CommandNode, 0),
        }
    }

    // 添加边
    fromList := g.adjacencyList[from]
    toList := g.adjacencyList[to]

    fromList.neighborLock.Lock()
    fromList.neighbors = append(fromList.neighbors, toList.node)
    fromList.neighborLock.Unlock()
}
```

### 2. 图缓存

```go
// 缓存图计算结果
type CachedGraph struct {
    *CommandGraph
    topologyCache []string
    cacheTime     time.Time
    cacheMutex    sync.RWMutex
    cacheDuration time.Duration
}

func (g *CachedGraph) TopologicalSort() ([]string, error) {
    g.cacheMutex.RLock()
    if time.Since(g.cacheTime) < g.cacheDuration {
        result := g.topologyCache
        g.cacheMutex.RUnlock()
        return result, nil
    }
    g.cacheMutex.RUnlock()

    // 重新计算
    result, err := g.CommandGraph.TopologicalSort()
    if err != nil {
        return nil, err
    }

    // 更新缓存
    g.cacheMutex.Lock()
    g.topologyCache = result
    g.cacheTime = time.Now()
    g.cacheMutex.Unlock()

    return result, nil
}
```

## 总结

### 队列

1. **程序管理**: 管理 TUI 程序实例
2. **命令历史**: 追踪命令执行历史
3. **异步处理**: 实现异步命令执行
4. **BFS 遍历**: 支持广度优先遍历
5. **FIFO 特性**: 保证先进先出

### 图

1. **依赖关系**: 表示命令间的依赖
2. **拓扑排序**: 确定执行顺序
3. **环检测**: 避免循环依赖
4. **路径查找**: 查找命令间路径
5. **可视化**: 生成依赖图

队列和图是 Cobra-X 中重要的数据结构，它们分别用于管理执行流程和表示复杂关系。合理使用这些结构可以实现高效的命令管理和依赖处理。
