# 树形数据结构 (Tree Data Structure)

## 概述

树形结构是 Cobra-X 中最核心的数据结构，用于表示命令的层级关系。树形结构支持高效的遍历、搜索和操作，是 CLI 工具中组织命令的理想选择。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **命令树**: [cobra/command_tree.go](../cobra/command_tree.go)
2. **树形菜单**: [tui/tree_menu.go](../tui/tree_menu.go)
3. **命令层次**: [cobra/command.go](../cobra/command.go) - 子命令管理

## 树形结构定义

### 1. 命令树节点

```go
// tui/tree_menu.go:42-50
type CommandItem struct {
    ID         string           // 节点唯一标识
    Name       string           // 命令名称
    Use        string           // 命令使用方式
    Short      string           // 短描述
    Long       string           // 长描述
    Children   []*CommandItem   // 子节点列表
    IsRunnable bool            // 是否为叶子节点（可执行）
}
```

**结构特点**：
- **根节点**: 顶层命令
- **内部节点**: 包含子命令的命令
- **叶子节点**: 可执行的命令
- **多叉树**: 每个节点可以有多个子节点

### 2. 树形菜单节点

```go
// tui/tree_menu.go:7-15
type TreeMenuItem struct {
    MenuItem                     // 嵌入菜单项
    Level     int                // 层级深度（0为根）
    Path      string             // 完整路径
    Children  []*TreeMenuItem    // 子项
    Expanded  bool              // 是否展开
    IsLeaf    bool              // 是否为叶子节点
}
```

## 树的构建

### 1. 递归构建算法

```go
// cobra/command_tree.go:10-43
func BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
    // 构建当前节点路径
    currentPath := path
    if path != "" {
        currentPath = path + " " + cmd.Name()
    } else {
        currentPath = cmd.Name()
    }

    // 判断是否为叶子节点
    isRunnable := cmd.Run != nil || cmd.RunE != nil

    // 创建当前节点
    item := &tui.CommandItem{
        ID:         cmd.Name(),
        Name:       cmd.Name(),
        Use:        cmd.Use,
        Short:      cmd.Short,
        Long:       cmd.Long,
        IsRunnable: isRunnable,
        Children:   make([]*tui.CommandItem, 0),
    }

    // 获取可用子命令
    children := getAvailableCommands(cmd.Commands())

    // 递归构建子节点
    for _, child := range children {
        childItem := BuildCommandTree(child, currentPath)
        if childItem != nil {
            item.Children = append(item.Children, childItem)
        }
    }

    return item
}
```

**构建过程**：
1. 深度优先遍历
2. 为每个命令创建节点
3. 递归处理子命令
4. 构建父子关系

### 2. 树形菜单构建

```go
// tui/tree_menu.go:52-89
func buildTree(items []*CommandItem, level int) *TreeMenuItem {
    if len(items) == 0 {
        return nil
    }

    root := &TreeMenuItem{
        Level:    level,
        Children: make([]*TreeMenuItem, 0, len(items)),
    }

    for _, item := range items {
        node := &TreeMenuItem{
            MenuItem: MenuItem{
                ID:          item.ID,
                Label:       item.Name,
                Description: item.Short,
            },
            Level:    level,
            Children: make([]*TreeMenuItem, 0),
            IsLeaf:   item.IsRunnable || len(item.Children) == 0,
        }

        // 递归处理子节点
        if len(item.Children) > 0 {
            for _, child := range item.Children {
                childNode := buildTree([]*CommandItem{child}, level+1)
                if childNode != nil {
                    node.Children = append(node.Children, childNode)
                }
            }
        }

        root.Children = append(root.Children, node)
    }

    return root
}
```

## 树的遍历

### 1. 深度优先遍历 (DFS)

```go
// 扁平化树形结构（DFS 后序遍历）
func flattenExecutableCommands(item *CommandItem, path string) []*CommandItem {
    result := make([]*CommandItem, 0)

    // 构建当前路径
    currentPath := path
    if path != "" {
        currentPath = path + " " + item.Use
    } else {
        currentPath = item.Use
    }

    // 如果是叶子节点，添加到结果
    if item.IsRunnable {
        displayPath := currentPath
        parts := strings.Fields(displayPath)
        if len(parts) > 1 {
            // 只显示子命令部分
            displayPath = strings.Join(parts[1:], " ")
        }

        result = append(result, &CommandItem{
            ID:         item.ID,
            Name:       item.Name,
            Use:        displayPath,
            Short:      item.Short,
            Long:       item.Long,
            IsRunnable: true,
            Children:   nil,
        })
    }

    // 递归处理子节点
    for _, child := range item.Children {
        childCommands := flattenExecutableCommands(child, currentPath)
        result = append(result, childCommands...)
    }

    return result
}
```

### 2. 广度优先遍历 (BFS)

```go
// 广度优先遍历树
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

// 使用示例
BFSTraverse(root, func(item *CommandItem) {
    fmt.Printf("Level: %d, Command: %s\n", item.Level, item.Name)
})
```

### 3. 层序遍历

```go
// 按层级遍历树
func LevelOrderTraverse(root *CommandItem) [][]*CommandItem {
    if root == nil {
        return nil
    }

    result := make([][]*CommandItem, 0)
    currentLevel := []*CommandItem{root}

    for len(currentLevel) > 0 {
        result = append(result, currentLevel)
        nextLevel := make([]*CommandItem, 0)

        for _, node := range currentLevel {
            nextLevel = append(nextLevel, node.Children...)
        }

        currentLevel = nextLevel
    }

    return result
}
```

## 树的搜索

### 1. 路径搜索

```go
// 根据路径查找命令
func FindCommandByPath(root *spf13cobra.Command, path string) *spf13cobra.Command {
    if path == "" {
        return root
    }

    parts := strings.Fields(path)
    current := root

    for _, part := range parts {
        found := false
        for _, cmd := range current.Commands() {
            if cmd.Name() == part && !cmd.Hidden {
                current = cmd
                found = true
                break
            }
        }

        if !found {
            return nil
        }
    }

    return current
}
```

### 2. ID 搜索

```go
// 在命令树中查找指定 ID 的命令
func FindCommandByID(root *CommandItem, id string) *CommandItem {
    if root.ID == id {
        return root
    }

    for _, child := range root.Children {
        if found := FindCommandByID(child, id); found != nil {
            return found
        }
    }

    return nil
}
```

### 3. 模糊搜索

```go
// 过滤树形菜单
func FilterTreeMenu(items []*TreeMenuItem, query string) []*TreeMenuItem {
    if query == "" {
        return items
    }

    query = strings.ToLower(query)
    result := make([]*TreeMenuItem, 0)

    for _, item := range items {
        // 匹配命令名称
        if strings.Contains(strings.ToLower(item.Label), query) {
            result = append(result, item)
            continue
        }

        // 匹配描述
        if strings.Contains(strings.ToLower(item.Description), query) {
            result = append(result, item)
            continue
        }

        // 匹配路径
        if strings.Contains(strings.ToLower(item.Path), query) {
            result = append(result, item)
        }
    }

    return result
}
```

## 树的操作

### 1. 添加子节点

```go
// 添加子命令
func (c *Command) AddCommand(cmds ...*Command) {
    for _, cmd := range cmds {
        c.Command.AddCommand(cmd.Command)
    }
}
```

### 2. 获取树的深度

```go
// 计算树的深度
func GetTreeDepth(root *CommandItem) int {
    if root == nil || len(root.Children) == 0 {
        return 0
    }

    maxDepth := 0
    for _, child := range root.Children {
        depth := GetTreeDepth(child)
        if depth > maxDepth {
            maxDepth = depth
        }
    }

    return maxDepth + 1
}
```

### 3. 获取树的节点数

```go
// 统计树的节点数
func CountNodes(root *CommandItem) int {
    if root == nil {
        return 0
    }

    count := 1
    for _, child := range root.Children {
        count += CountNodes(child)
    }

    return count
}
```

### 4. 获取树的叶子节点

```go
// 获取所有叶子节点
func GetLeafNodes(root *CommandItem) []*CommandItem {
    if root == nil {
        return nil
    }

    result := make([]*CommandItem, 0)

    if len(root.Children) == 0 {
        result = append(result, root)
        return result
    }

    for _, child := range root.Children {
        result = append(result, GetLeafNodes(child)...)
    }

    return result
}
```

## 树的转换

### 1. 树到列表的转换

```go
// 获取所有可执行命令（扁平化列表）
func GetExecutableCommands(cmd *spf13cobra.Command) []*CommandItem {
    root := BuildCommandTree(cmd, "")

    if len(root.Children) > 0 {
        var result []*CommandItem
        for _, child := range root.Children {
            result = append(result, flattenExecutableCommands(child, "")...)
        }
        return result
    }

    return flattenExecutableCommands(root, "")
}
```

### 2. 列表到树的转换

```go
// 从扁平列表构建树
func BuildTreeFromFlat(items []*CommandItem) *CommandItem {
    if len(items) == 0 {
        return nil
    }

    root := &CommandItem{
        ID:       "root",
        Children: make([]*CommandItem, 0),
    }

    for _, item := range items {
        parts := strings.Fields(item.Use)
        current := root

        // 构建路径
        for i, part := range parts {
            var child *CommandItem
            for _, c := range current.Children {
                if c.Name == part {
                    child = c
                    break
                }
            }

            if child == nil {
                isLast := i == len(parts)-1
                child = &CommandItem{
                    ID:         part,
                    Name:       part,
                    Use:        strings.Join(parts[:i+1], " "),
                    IsRunnable: isLast && item.IsRunnable,
                    Children:   make([]*CommandItem, 0),
                }
                current.Children = append(current.Children, child)
            }

            current = child
        }
    }

    return root
}
```

## 树的可视化

### 1. 缩进显示

```go
// 获取树形菜单的显示文本
func GetTreeMenuDisplay(items []*TreeMenuItem, selectedIndex int) []string {
    lines := make([]string, 0, len(items))

    for i, item := range items {
        cursor := " "
        if i == selectedIndex {
            cursor = "▶"
        }

        // 缩进
        indent := strings.Repeat("  ", item.Level)

        // 图标
        icon := "📄"
        if strings.Contains(item.Path, " ") {
            icon = "📁"
        }

        line := cursor + " " + indent + icon + " " + item.Label

        if item.Description != "" && i == selectedIndex {
            line += "\n" + indent + "   └─ " + item.Description
        }

        lines = append(lines, line)
    }

    return lines
}
```

### 2. ASCII 树形显示

```go
// 渲染 ASCII 树
func RenderASCIITree(node *CommandItem, prefix string, isLast bool) string {
    if node == nil {
        return ""
    }

    var sb strings.Builder

    connector := "├── "
    if isLast {
        connector = "└── "
    }

    icon := "📄"
    if !node.IsRunnable {
        icon = "📁"
    }

    sb.WriteString(prefix + connector + icon + " " + node.Name + "\n")

    for i, child := range node.Children {
        isLastChild := i == len(node.Children)-1
        newPrefix := prefix
        if isLast {
            newPrefix += "    "
        } else {
            newPrefix += "│   "
        }
        sb.WriteString(RenderASCIITree(child, newPrefix, isLastChild))
    }

    return sb.String()
}
```

**输出示例**：
```
📁 myapp
├── 📁 server
│   ├── 📄 start
│   ├── 📄 stop
│   └── 📄 status
├── 📁 client
│   ├── 📄 connect
│   └── 📄 disconnect
└── 📄 config
```

## 树的应用场景

### 1. 命令补全

```go
// 根据前缀获取可能的命令
func GetCompletions(root *CommandItem, prefix string) []string {
    parts := strings.Fields(prefix)
    current := root

    // 导航到最后一层
    for i, part := range parts {
        found := false
        for _, child := range current.Children {
            if child.Name == part {
                current = child
                found = true
                break
            }
        }

        if !found && i < len(parts)-1 {
            return nil  // 路径无效
        }
    }

    // 返回子命令名称
    completions := make([]string, 0, len(current.Children))
    for _, child := range current.Children {
        completions = append(completions, child.Name)
    }

    return completions
}
```

### 2. 命令帮助生成

```go
// 生成命令树的帮助文档
func GenerateHelp(root *CommandItem) string {
    var sb strings.Builder

    sb.WriteString("Available commands:\n\n")

    items := flattenExecutableCommands(root, "")
    for _, item := range items {
        sb.WriteString(fmt.Sprintf("  %-20s %s\n", item.Use, item.Short))
    }

    return sb.String()
}
```

### 3. 命令历史

```go
// 追踪命令执行路径
type CommandPath struct {
    commands []*Command
    index    int
}

func (p *CommandPath) Push(cmd *Command) {
    p.commands = append(p.commands[:p.index], cmd)
    p.index++
}

func (p *CommandPath) Back() *Command {
    if p.index > 0 {
        p.index--
        return p.commands[p.index]
    }
    return nil
}

func (p *CommandPath) Forward() *Command {
    if p.index < len(p.commands)-1 {
        p.index++
        return p.commands[p.index]
    }
    return nil
}
```

## 性能优化

### 1. 缓存树结构

```go
type CommandTreeCache struct {
    tree  *CommandItem
    mutex sync.RWMutex
}

func (c *CommandTreeCache) Get(root *spf13cobra.Command) *CommandItem {
    c.mutex.RLock()
    if c.tree != nil {
        c.mutex.RUnlock()
        return c.tree
    }
    c.mutex.RUnlock()

    c.mutex.Lock()
    defer c.mutex.Unlock()

    if c.tree != nil {
        return c.tree
    }

    c.tree = BuildCommandTree(root, "")
    return c.tree
}
```

### 2. 延迟加载子节点

```go
type LazyCommandItem struct {
    *CommandItem
    childrenLoaded bool
    loadChildren   func() []*CommandItem
}

func (item *LazyCommandItem) GetChildren() []*CommandItem {
    if !item.childrenLoaded {
        item.Children = item.loadChildren()
        item.childrenLoaded = true
    }
    return item.Children
}
```

### 3. 索引优化

```go
type CommandIndex struct {
    byID   map[string]*CommandItem
    byPath map[string]*CommandItem
}

func BuildIndex(root *CommandItem) *CommandIndex {
    index := &CommandIndex{
        byID:   make(map[string]*CommandItem),
        byPath: make(map[string]*CommandItem),
    }

    var walk func(*CommandItem, string)
    walk = func(node *CommandItem, path string) {
        if node == nil {
            return
        }

        currentPath := path
        if path != "" {
            currentPath = path + " " + node.Name
        } else {
            currentPath = node.Name
        }

        index.byID[node.ID] = node
        index.byPath[currentPath] = node

        for _, child := range node.Children {
            walk(child, currentPath)
        }
    }

    walk(root, "")
    return index
}
```

## 最佳实践

### 1. 限制树的深度

```go
const MaxTreeDepth = 5

func BuildCommandTreeWithLimit(cmd *spf13cobra.Command, path string, depth int) *CommandItem {
    if depth > MaxTreeDepth {
        return nil
    }

    // 构建逻辑...
    for _, child := range children {
        childItem := BuildCommandTreeWithLimit(child, currentPath, depth+1)
        // ...
    }
}
```

### 2. 过滤无效节点

```go
func getAvailableCommands(cmds []*spf13cobra.Command) []*spf13cobra.Command {
    var result []*spf13cobra.Command
    for _, cmd := range cmds {
        if !cmd.IsAvailableCommand() {
            continue
        }
        if cmd.Hidden {
            continue
        }
        result = append(result, cmd)
    }
    return result
}
```

### 3. 处理循环引用

```go
func ValidateTree(root *CommandItem) error {
    visited := make(map[string]bool)
    return validateNode(root, visited)
}

func validateNode(node *CommandItem, visited map[string]bool) error {
    if node == nil {
        return nil
    }

    if visited[node.ID] {
        return fmt.Errorf("cycle detected: %s", node.ID)
    }

    visited[node.ID] = true

    for _, child := range node.Children {
        if err := validateNode(child, visited); err != nil {
            return err
        }
    }

    delete(visited, node.ID)
    return nil
}
```

## 总结

树形数据结构在 Cobra-X 中：

1. **表示命令层次**: 自然地表示命令的父子关系
2. **支持高效遍历**: DFS/BFS 遍历支持各种操作
3. **灵活的搜索**: 按路径、ID、关键词搜索
4. **易于可视化**: 缩进或 ASCII 树形显示
5. **支持转换**: 树和列表之间自由转换

树形结构是 CLI 工具中最适合的数据结构，它完美地映射了命令的组织方式。Cobra-X 通过高效的树形结构实现，提供了强大的命令管理能力，同时保持了良好的性能和可维护性。
