# 组合模式 (Composite Pattern)

## 概述

组合模式是一种结构型设计模式，它允许你将对象组合成树形结构来表示"部分-整体"的层次结构。组合模式能让客户以一致的方式处理个别对象以及对象组合。在 Cobra-X 中，命令树是组合模式的典型应用。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **命令树构建**: [cobra/command_tree.go](../cobra/command_tree.go) - `BuildCommandTree()`
2. **树形菜单**: [tui/tree_menu.go](../tui/tree_menu.go) - `TreeMenuItem`
3. **命令层次**: [cobra/command.go](../cobra/command.go) - 子命令管理
4. **扁平化**: [cobra/command_tree.go](../cobra/command_tree.go) - `GetExecutableCommands()`

## 代码实现分析

### 1. 树形结构定义

```go
// tui/tree_menu.go:42-50
type CommandItem struct {
    ID         string
    Name       string
    Use        string
    Short      string
    Long       string
    Children   []*CommandItem  // 子节点列表
    IsRunnable bool            // 是否可执行（叶子节点）
}
```

**组合模式要素**：
- **组件接口**: `CommandItem`（统一的结构）
- **叶子节点**: `IsRunnable == true` 的命令
- **复合节点**: 有 `Children` 的命令
- **统一操作**: 递归处理所有节点

### 2. 树形结构构建

```go
// cobra/command_tree.go:10-43
func BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
    // 构建当前命令路径
    currentPath := path
    if path != "" {
        currentPath = path + " " + cmd.Name()
    } else {
        currentPath = cmd.Name()
    }

    // 判断命令是否可执行
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

    // 获取可用的子命令
    children := getAvailableCommands(cmd.Commands())
    for _, child := range children {
        childItem := BuildCommandTree(child, currentPath)
        if childItem != nil {
            item.Children = append(item.Children, childItem)
        }
    }

    return item
}
```

**递归构建**：
- 为每个命令创建节点
- 递归处理子命令
- 构建树形层次结构

### 3. 树形菜单结构

```go
// tui/tree_menu.go:7-15
type TreeMenuItem struct {
    MenuItem
    Level     int              // 层级深度（0为根）
    Path      string           // 完整路径
    Children  []*TreeMenuItem  // 子项
    Expanded  bool             // 是否展开
    IsLeaf    bool             // 是否为叶子节点
}
```

### 4. 树形菜单构建

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

## 组合模式的核心操作

### 1. 递归遍历

```go
// 扁平化树形结构
func flattenTree(node *TreeMenuItem, level int, path string) []*TreeMenuItem {
    if node == nil {
        return nil
    }

    result := make([]*TreeMenuItem, 0)

    // 构建当前节点路径
    currentPath := path
    if node.Label != "" {
        if currentPath != "" {
            currentPath += " " + node.Label
        } else {
            currentPath = node.Label
        }
    }

    // 如果是叶子节点，添加到结果
    if node.IsLeaf && node.Label != "" {
        item := &TreeMenuItem{
            MenuItem: MenuItem{
                ID:          node.ID,
                Label:       node.Label,
                Description: node.Description,
            },
            Level:    level,
            Path:     currentPath,
            IsLeaf:   true,
        }
        result = append(result, item)
    }

    // 递归处理子节点
    for _, child := range node.Children {
        childItems := flattenTree(child, level+1, currentPath)
        result = append(result, childItems...)
    }

    return result
}
```

**递归模式**：
- 处理当前节点
- 递归处理子节点
- 合并结果

### 2. 树形搜索

```go
// 在命令树中查找指定路径的命令
func FindCommandByPath(root *spf13cobra.Command, path string) *spf13cobra.Command {
    if path == "" {
        return root
    }

    parts := strings.Fields(path)
    current := root

    for _, part := range parts {
        // 查找子命令
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

### 3. 树形过滤

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

## 树形结构的操作

### 1. 添加子节点

```go
// cobra/command.go:564-569
func (c *Command) AddCommand(cmds ...*Command) {
    for _, cmd := range cmds {
        c.Command.AddCommand(cmd.Command)
    }
}
```

### 2. 获取子命令

```go
// cobra/command.go:294-307
func (c *Command) getChildren(cmd *Command) []*Command {
    spf13Children := cmd.Commands()
    children := make([]*Command, 0, len(spf13Children))

    for _, child := range spf13Children {
        if !child.IsAvailableCommand() {
            continue
        }
        children = append(children, c.wrapCommand(child))
    }

    return children
}
```

### 3. 获取完整路径

```go
// cobra/command_tree.go:136-156
func GetCommandFullPath(cmd *spf13cobra.Command) string {
    if cmd == nil {
        return ""
    }

    // 递归获取父级路径
    var pathParts []string
    current := cmd

    for current != nil {
        pathParts = append([]string{current.Name()}, pathParts...)
        current = current.Parent()
    }

    if len(pathParts) == 0 {
        return cmd.Name()
    }

    return strings.Join(pathParts, " ")
}
```

## 扁平化操作

### 1. 树到列表的转换

```go
// 获取所有可执行命令（扁平化列表）
func GetExecutableCommands(cmd *spf13cobra.Command) []*tui.CommandItem {
    root := BuildCommandTree(cmd, "")

    // 如果根命令有子命令，则只返回子命令中的可执行命令
    if len(root.Children) > 0 {
        var result []*tui.CommandItem
        for _, child := range root.Children {
            // 不传递根命令名称
            result = append(result, flattenExecutableCommands(child, "")...)
        }
        return result
    }

    // 如果根命令没有子命令，则返回根命令本身
    return flattenExecutableCommands(root, "")
}
```

**扁平化优势**：
- 简化用户界面
- 便于搜索和过滤
- 适合命令补全

### 2. 列表到树的转换

```go
// 从扁平列表构建树
func BuildTreeFromFlat(items []*CommandItem) *CommandItem {
    if len(items) == 0 {
        return nil
    }

    // 创建根节点
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

## 树形结构的可视化

### 1. 带缩进的显示

```go
// 获取树形菜单的显示文本（带缩进和图标）
func GetTreeMenuDisplay(items []*TreeMenuItem, selectedIndex int) []string {
    lines := make([]string, 0, len(items))

    for i, item := range items {
        // 选择指示器
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

        // 构建显示文本
        line := cursor + " " + indent + icon + " " + item.Label

        // 如果有描述，添加到下一行
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
func RenderASCIITree(node *CommandItem, prefix string, isLast bool) string {
    if node == nil {
        return ""
    }

    var sb strings.Builder

    // 当前节点
    connector := "├── "
    if isLast {
        connector = "└── "
    }

    icon := "📄"
    if !node.IsRunnable {
        icon = "📁"
    }

    sb.WriteString(prefix + connector + icon + " " + node.Name + "\n")

    // 子节点
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

## 组合模式的优势

### 1. 统一接口

```go
// 统一处理单个命令和命令组
func ExecuteCommand(cmd *Command) error {
    return cmd.Execute()
}

func ExecuteCommandGroup(cmds []*Command) error {
    for _, cmd := range cmds {
        if err := ExecuteCommand(cmd); err != nil {
            return err
        }
    }
    return nil
}
```

### 2. 递归操作

```go
// 对整棵树执行操作
func WalkTree(root *CommandItem, fn func(*CommandItem)) {
    if root == nil {
        return
    }

    fn(root)

    for _, child := range root.Children {
        WalkTree(child, fn)
    }
}

// 使用
WalkTree(root, func(item *CommandItem) {
    fmt.Println(item.Use)
})
```

### 3. 层次遍历

```go
// 广度优先遍历
func BFSTraverse(root *CommandItem, fn func(*CommandItem)) {
    if root == nil {
        return
    }

    queue := []*CommandItem{root}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        fn(current)

        queue = append(queue, current.Children...)
    }
}
```

## 实际应用示例

### 1. 命令树结构

```
myapp (根命令)
├── server (服务器命令组)
│   ├── start (启动服务器)
│   ├── stop (停止服务器)
│   └── restart (重启服务器)
├── client (客户端命令组)
│   ├── connect (连接服务器)
│   ├── disconnect (断开连接)
│   └── status (查看状态)
└── config (配置命令)
    ├── get (获取配置)
    ├── set (设置配置)
    └── list (列出所有配置)
```

### 2. 代码实现

```go
func setupCommands() *cobra.Command {
    rootCmd := cobra.NewCommand("myapp",
        cobra.WithShort("My application"),
    )

    // 服务器命令组
    serverCmd := cobra.NewCommand("server",
        cobra.WithShort("Server commands"),
    )

    startCmd := cobra.NewCommand("start",
        cobra.WithShort("Start server"),
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            startServer()
        }),
    )

    stopCmd := cobra.NewCommand("stop",
        cobra.WithShort("Stop server"),
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            stopServer()
        }),
    )

    serverCmd.AddCommand(startCmd, stopCmd)

    // 客户端命令组
    clientCmd := cobra.NewCommand("client",
        cobra.WithShort("Client commands"),
    )

    connectCmd := cobra.NewCommand("connect",
        cobra.WithShort("Connect to server"),
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            connect()
        }),
    )

    clientCmd.AddCommand(connectCmd)

    // 组装命令树
    rootCmd.AddCommand(serverCmd, clientCmd)

    return rootCmd
}
```

## 高级技巧

### 1. 树形结构缓存

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

    // 双重检查
    if c.tree != nil {
        return c.tree
    }

    c.tree = BuildCommandTree(root, "")
    return c.tree
}

func (c *CommandTreeCache) Invalidate() {
    c.mutex.Lock()
    defer c.mutex.Unlock()
    c.tree = nil
}
```

### 2. 树形结构比较

```go
func CompareTrees(oldTree, newTree *CommandItem) []string {
    changes := make([]string, 0)

    // 找出新增的命令
    newCommands := findMissingCommands(oldTree, newTree)
    for _, cmd := range newCommands {
        changes = append(changes, fmt.Sprintf("Added: %s", cmd.Use))
    }

    // 找出删除的命令
    deletedCommands := findMissingCommands(newTree, oldTree)
    for _, cmd := range deletedCommands {
        changes = append(changes, fmt.Sprintf("Deleted: %s", cmd.Use))
    }

    return changes
}
```

### 3. 树形结构验证

```go
func ValidateTree(root *CommandItem) error {
    if root == nil {
        return errors.New("tree is nil")
    }

    // 检查循环引用
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

## 最佳实践

### 1. 合理的深度

```go
// 限制树的深度
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

### 2. 过滤隐藏节点

```go
// 获取可用的子命令
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

### 3. 延迟加载

```go
// 延迟加载子命令
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

## 潜在问题和解决方案

### 问题 1: 树的深度过大

**问题**: 命令层级过深导致导航困难

**解决方案**:
```go
// 使用扁平化视图
func GetExecutableCommands(cmd *spf13cobra.Command) []*CommandItem {
    // 返回扁平化列表，不保留层级结构
    return flattenAllCommands(BuildCommandTree(cmd, ""))
}
```

### 问题 2: 树的重复遍历

**问题**: 多次遍历同一棵树影响性能

**解决方案**:
```go
// 使用缓存
type TreeCache struct {
    cache map[string]*CommandItem
    mutex sync.RWMutex
}

func (c *TreeCache) GetOrCreate(key string, builder func() *CommandItem) *CommandItem {
    c.mutex.RLock()
    if tree, ok := c.cache[key]; ok {
        c.mutex.RUnlock()
        return tree
    }
    c.mutex.RUnlock()

    c.mutex.Lock()
    defer c.mutex.Unlock()

    if tree, ok := c.cache[key]; ok {
        return tree
    }

    tree := builder()
    c.cache[key] = tree
    return tree
}
```

### 问题 3: 树的修改同步

**问题**: 树结构修改后需要同步更新视图

**解决方案**:
```go
// 使用观察者模式
type ObservableTree struct {
    *CommandItem
    observers []func(*CommandItem)
    mutex     sync.RWMutex
}

func (t *ObservableTree) AddChild(child *CommandItem) {
    t.mutex.Lock()
    t.Children = append(t.Children, child)
    t.mutex.Unlock()

    t.notifyObservers()
}

func (t *ObservableTree) notifyObservers() {
    t.mutex.RLock()
    observers := t.observers
    t.mutex.RUnlock()

    for _, observer := range observers {
        observer(t)
    }
}
```

## 总结

组合模式在 Cobra-X 中实现了优雅的命令层次结构：

1. **树形结构**: 命令组织成树形层次
2. **统一接口**: 一致的方式处理单个命令和命令组
3. **递归操作**: 简化树形结构的遍历和操作
4. **灵活视图**: 支持树形和扁平化视图
5. **易于扩展**: 添加新命令只需添加到树中

组合模式使得 Cobra-X 能够处理复杂的命令层次结构，同时保持代码的简洁和可维护性。这种设计让用户可以自然地组织命令，也使得框架能够提供多种导航和执行方式。树形结构是 CLI 工具的天然选择，组合模式为这种结构提供了完美的实现方案。
