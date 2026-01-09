# 切片数据结构 (Slice Data Structure)

## 概述

切片（Slice）是 Go 语言中最常用的数据结构，它提供了动态数组的实现。在 Cobra-X 中，切片被广泛用于存储动态列表、构建命令树、管理菜单项和收集 flags。

## 在 Cobra-X 中的应用

### 核心应用场景

1. **子命令列表**: 存储命令的子命令
2. **菜单项列表**: 存储可选择的菜单项
3. **Flag 列表**: 存储需要配置的 flags
4. **命令路径**: 存储命令执行的路径
5. **程序管理**: 存储运行的 TUI 程序

## 代码实现分析

### 1. 子命令切片

```go
// cobra/command.go:28
type Command struct {
    *spf13cobra.Command
    tuiConfig *TUIConfig
    tuiFlags  *pflag.FlagSet
    children  []*Command  // 子命令切片
}

// 获取子命令列表
func (c *Command) getChildren(cmd *Command) []*Command {
    spf13Children := cmd.Commands()
    children := make([]*Command, 0, len(spf13Children))  // 预分配容量

    for _, child := range spf13Children {
        if !child.IsAvailableCommand() {
            continue
        }
        children = append(children, c.wrapCommand(child))
    }

    return children
}
```

**切片使用技巧**：
- **预分配容量**: `make([]*Command, 0, len(spf13Children))` 避免动态扩容
- **条件过滤**: 只添加满足条件的元素
- **类型转换**: 包装原始命令类型

### 2. 菜单项切片

```go
// cobra/command.go:262-271
// 构建菜单项
menuItems := make([]tui.MenuItem, 0, len(children))
for _, child := range children {
    menuItems = append(menuItems, tui.MenuItem{
        ID:          child.Use,
        Label:       child.Use,
        Description: child.Short,
        Disabled:    !child.IsAvailableCommand(),
    })
}
```

**切片构建**：
- 预分配容量提高性能
- 结构体字面量初始化
- 条件设置属性

### 3. Flag 列表切片

```go
// cobra/command.go:321-400
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem  // 使用 var 声明，长度为 0
    seen := make(map[string]bool)

    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if strings.HasPrefix(flag.Name, "tui-") || flag.Name == "tui" || seen[flag.Name] {
                return
            }

            item := tui.FlagItem{
                Name:         flag.Name,
                ShortName:    flag.Shorthand,
                Description:  flag.Usage,
                DefaultValue: flag.DefValue,
                CurrentValue: flag.DefValue,
                Required:     false,
                SourceCommand: current.Name(),
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

            items = append(items, item)
            seen[flag.Name] = true
        })

        // 类似处理 PersistentFlags...

        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

**动态切片**：
- 使用 `var` 声明，长度自动增长
- 在循环中追加元素
- 复杂的条件逻辑

### 4. 命令路径切片

```go
// cobra/command.go:253-292
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
    children := c.getChildren(cmd)
    if len(children) == 0 {
        return append(path, cmd), nil  // 追加当前命令到路径
    }

    // ... 构建菜单项...

    selectedIndex, err := renderer.RenderCommandMenu(cmd.Use, menuItems)
    if err != nil {
        return nil, err
    }

    if selectedIndex < 0 {
        return nil, nil
    }

    selectedChild := children[selectedIndex]
    newPath := append(path, cmd)  // 构建新路径
    return c.navigateCommandTree(renderer, selectedChild, newPath)
}
```

**路径切片**：
- 存储命令执行的历史路径
- 使用 `append` 构建新路径
- 递归传递路径信息

### 5. 程序管理切片

```go
// tui/default_renderer.go:13-18
type DefaultRenderer struct {
    theme    *style.Theme
    programs []*tea.Program  // 程序切片
}

func (r *DefaultRenderer) RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error) {
    width, height := getTerminalSize()

    model := newMenuModel(options, r.theme, width, height)

    p := tea.NewProgram(model, tea.WithAltScreen())
    r.programs = append(r.programs, p)  // 追加程序到列表

    result, err := p.Run()
    // ...
}
```

## 切片常见操作

### 1. 过滤切片

```go
// 获取可用的子命令
func getAvailableCommands(cmds []*spf13cobra.Command) []*spf13cobra.Command {
    result := make([]*spf13cobra.Command, 0, len(cmds))

    for _, cmd := range cmds {
        if !cmd.IsAvailableCommand() {
            continue
        }
        if cmd.Hidden {
            continue
        }

        // 过滤补全命令
        if isCompletionCommand(cmd) {
            continue
        }

        // 过滤 help 命令
        if cmd.Name() == "help" {
            continue
        }

        result = append(result, cmd)
    }

    return result
}
```

### 2. 映射切片

```go
// 命令项转换为菜单项
func buildMenuItems(cmds []*CommandItem) []MenuItem {
    items := make([]MenuItem, 0, len(cmds))

    for _, cmd := range cmds {
        items = append(items, MenuItem{
            ID:          cmd.ID,
            Label:       cmd.Name,
            Description: cmd.Short,
        })
    }

    return items
}
```

### 3. 扁平化切片

```go
// 扁平化树形结构
func flattenExecutableCommands(item *CommandItem, path string) []*CommandItem {
    result := make([]*CommandItem, 0)

    // 处理当前节点
    if item.IsRunnable {
        result = append(result, &CommandItem{
            ID:         item.ID,
            Name:       item.Name,
            Use:        path,
            Short:      item.Short,
            Long:       item.Long,
            IsRunnable: true,
            Children:   nil,
        })
    }

    // 递归处理子节点
    for _, child := range item.Children {
        childCommands := flattenExecutableCommands(child, path)
        result = append(result, childCommands...)
    }

    return result
}
```

### 4. 查找切片

```go
// 在命令树中查找指定 ID 的命令
func findCommandByID(root *spf13cobra.Command, id string) *spf13cobra.Command {
    if root.Name() == id {
        return root
    }

    for _, cmd := range root.Commands() {
        if cmd.Name() == id {
            return cmd
        }
        if found := findCommandByID(cmd, id); found != nil {
            return found
        }
    }

    return nil
}

// 线性查找
func FindMenuItem(items []MenuItem, id string) (MenuItem, bool) {
    for _, item := range items {
        if item.ID == id {
            return item, true
        }
    }
    return MenuItem{}, false
}
```

### 5. 排序切片

```go
import "sort"

// 按名称排序菜单项
type ByName []MenuItem

func (b ByName) Len() int           { return len(b) }
func (b ByName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByName) Less(i, j int) bool { return b[i].Label < b[j].Label }

func SortMenuItems(items []MenuItem) {
    sort.Sort(ByName(items))
}

// 使用 sort.Slice
func SortMenuItemsV2(items []MenuItem) {
    sort.Slice(items, func(i, j int) bool {
        return items[i].Label < items[j].Label
    })
}
```

### 6. 去重切片

```go
// 去重切片
func UniqueItems(items []string) []string {
    seen := make(map[string]bool)
    result := make([]string, 0, len(items))

    for _, item := range items {
        if !seen[item] {
            seen[item] = true
            result = append(result, item)
        }
    }

    return result
}
```

## 切片性能优化

### 1. 预分配容量

```go
// 好的做法 - 预分配容量
items := make([]MenuItem, 0, len(commands))
for _, cmd := range commands {
    items = append(items, cmd)
}

// 避免 - 不预分配，多次扩容
var items []MenuItem
for _, cmd := range commands {
    items = append(items, cmd)  // 可能多次扩容
}
```

### 2. 切片重用

```go
// 重用切片底层数组
type Buffer struct {
    buf []byte
}

func (b *Buffer) Write(data []byte) {
    b.buf = append(b.buf[:0], data...)  // 重用底层数组
}
```

### 3. 避免不必要的复制

```go
// 好的做法 - 使用切片
func processItems(items []MenuItem) {
    for _, item := range items {
        // 处理 item
    }
}

// 避免 - 复制到数组
func processItemsBad(items []MenuItem) {
    arr := [100]MenuItem{}  // 固定大小，浪费空间
    copy(arr[:], items)
    for _, item := range arr {
        if item.ID == "" {
            break
        }
        // 处理 item
    }
}
```

## 切片的高级技巧

### 1. 切片作为队列

```go
type Queue struct {
    items []MenuItem
}

func (q *Queue) Enqueue(item MenuItem) {
    q.items = append(q.items, item)
}

func (q *Queue) Dequeue() (MenuItem, bool) {
    if len(q.items) == 0 {
        return MenuItem{}, false
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item, true
}

// 更高效的实现 - 使用循环索引
type CircularQueue struct {
    items []MenuItem
    head  int
    tail  int
    size  int
}

func NewCircularQueue(capacity int) *CircularQueue {
    return &CircularQueue{
        items: make([]MenuItem, capacity),
    }
}

func (q *CircularQueue) Enqueue(item MenuItem) bool {
    if q.size == len(q.items) {
        return false  // 队列已满
    }
    q.items[q.tail] = item
    q.tail = (q.tail + 1) % len(q.items)
    q.size++
    return true
}

func (q *CircularQueue) Dequeue() (MenuItem, bool) {
    if q.size == 0 {
        return MenuItem{}, false
    }
    item := q.items[q.head]
    q.head = (q.head + 1) % len(q.items)
    q.size--
    return item, true
}
```

### 2. 切片作为栈

```go
type Stack struct {
    items []MenuItem
}

func (s *Stack) Push(item MenuItem) {
    s.items = append(s.items, item)
}

func (s *Stack) Pop() (MenuItem, bool) {
    if len(s.items) == 0 {
        return MenuItem{}, false
    }
    index := len(s.items) - 1
    item := s.items[index]
    s.items = s.items[:index]
    return item, true
}

func (s *Stack) Peek() (MenuItem, bool) {
    if len(s.items) == 0 {
        return MenuItem{}, false
    }
    return s.items[len(s.items)-1], true
}
```

### 3. 切片过滤链

```go
// 链式过滤
type FilterFunc func(MenuItem) bool

func FilterItems(items []MenuItem, filters ...FilterFunc) []MenuItem {
    result := make([]MenuItem, 0, len(items))

    for _, item := range items {
        include := true
        for _, filter := range filters {
            if !filter(item) {
                include = false
                break
            }
        }
        if include {
            result = append(result, item)
        }
    }

    return result
}

// 使用
available := FilterItems(items,
    func(item MenuItem) bool { return !item.Disabled },
    func(item MenuItem) bool { return strings.HasPrefix(item.ID, "cmd") },
)
```

### 4. 切片分页

```go
// 分页处理
func Paginate(items []MenuItem, page, pageSize int) []MenuItem {
    if page < 1 {
        page = 1
    }

    start := (page - 1) * pageSize
    if start >= len(items) {
        return []MenuItem{}  // 超出范围
    }

    end := start + pageSize
    if end > len(items) {
        end = len(items)
    }

    return items[start:end]
}

// 计算总页数
func TotalPages(items []MenuItem, pageSize int) int {
    return (len(items) + pageSize - 1) / pageSize
}
```

## 切片的安全使用

### 1. 边界检查

```go
// 好的做法 - 检查边界
if index >= 0 && index < len(items) {
    item := items[index]
    // 使用 item
}

// 避免 - 可能 panic
item := items[index]  // 如果 index 越界会 panic
```

### 2. 切片零值

```go
// nil 切片可以安全使用
var items []MenuItem

// 可以遍历
for _, item := range items {
    // 不会执行
}

// 可以 append
items = append(items, MenuItem{})  // OK

// 可以获取长度
len(items)  // 返回 0

// 但不能访问元素
// items[0]  // panic
```

### 3. 切片拷贝

```go
// 浅拷贝 - 只拷贝切片结构
newSlice := oldSlice  // 两个切片共享底层数组

// 深拷贝 - 拷贝元素
newSlice := make([]MenuItem, len(oldSlice))
copy(newSlice, oldSlice)

// 或使用 append
newSlice := append([]MenuItem{}, oldSlice...)
```

## 切片的常见陷阱

### 陷阱 1: 切片共享底层数组

```go
// 问题 - 修改影响原切片
items := []MenuItem{{ID: "1"}, {ID: "2"}}
subset := items[0:1]
subset[0].ID = "modified"  // 也会修改 items[0]

// 解决 - 使用 copy
subset := make([]MenuItem, 1)
copy(subset, items[0:1])
subset[0].ID = "modified"  // 不影响 items
```

### 陷阱 2: 追加到子切片

```go
// 问题 - 可能覆盖原数据
items := []MenuItem{{ID: "1"}, {ID: "2"}}
subset := items[0:1]
subset = append(subset, MenuItem{ID: "3"})  // 可能覆盖 items[1]

// 解决 - 创建新切片
subset := make([]MenuItem, 1)
copy(subset, items[0:1])
subset = append(subset, MenuItem{ID: "3"})  // 安全
```

### 陷阱 3: 循环中捕获变量

```go
// 问题 - 所有 goroutine 使用相同的变量
for _, item := range items {
    go func() {
        process(item)  // 可能都处理最后一个 item
    }()
}

// 解决 - 传递参数
for _, item := range items {
    go func(item MenuItem) {
        process(item)
    }(item)
}
```

## 最佳实践

### 1. 选择合适的初始容量

```go
// 如果知道大概大小，预分配
items := make([]MenuItem, 0, expectedSize)

// 如果不知道，使用 var
var items []MenuItem
```

### 2. 使用切片表达式避免越界

```go
// 好的做法 - 自动处理边界
items[0:min(10, len(items))]

// 或使用切片
if len(items) > 10 {
    items = items[:10]
}
```

### 3. 删除元素

```go
// 删除索引 i 的元素
items = append(items[:i], items[i+1:]...)

// 删除最后一个元素
items = items[:len(items)-1]

// 删除第一个元素
items = items[1:]
```

## 总结

切片数据结构在 Cobra-X 中：

1. **动态列表**: 存储动态数量的元素
2. **高效追加**: O(1) 平均时间复杂度
3. **灵活操作**: 支持切片、追加、复制等操作
4. **内存共享**: 切片之间可以共享底层数组
5. **类型安全**: 编译时类型检查

切片是 Go 语言中最灵活和常用的数据结构。Cobra-X 充分利用了切片的特性来管理动态列表、构建树形结构和实现各种数据操作。合理使用切片可以大大简化代码并提高性能。
