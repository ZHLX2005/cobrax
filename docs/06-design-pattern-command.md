# 命令模式 (Command Pattern)

## 概述

命令模式是一种行为型设计模式，它将请求封装为对象，从而允许用户用不同的请求对客户进行参数化、对请求排队或记录请求日志，以及支持可撤销的操作。在 Cobra-X 中，命令模式是整个框架的基础，所有 CLI 和 TUI 操作都通过命令对象执行。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **命令定义**: [cobra/command.go](../cobra/command.go) - `Command` 结构
2. **命令执行**: [cobra/command.go](../cobra/command.go) - `Execute()` 方法
3. **装饰器命令**: [cobra/decorate.go](../cobra/decorate.go) - `Enhance()` 函数
4. **命令树**: [cobra/command_tree.go](../cobra/command_tree.go) - 命令结构管理

## 代码实现分析

### 1. 命令对象结构

```go
// cobra/command.go:15-29
type Command struct {
    *spf13cobra.Command  // 嵌入原始命令

    // TUI 配置
    tuiConfig *TUIConfig
    tuiFlags  *pflag.FlagSet
    children  []*Command
}
```

**命令模式要素**：
- **命令接口**: 通过嵌入 `spf13cobra.Command` 实现
- **具体命令**: `Command` 结构
- **接收者**: 命令的操作对象
- **执行方法**: `Execute()` / `Run` / `RunE`

### 2. 命令执行接口

```go
// cobra/command.go:132-147
func (c *Command) Execute() error {
    // 检查是否应该使用 TUI
    if c.shouldUseTUI() {
        return c.executeTUI()
    }

    // 使用传统 CLI 模式
    return c.Command.Execute()
}

func (c *Command) ExecuteE() error {
    return c.Execute()
}
```

**执行策略**：
1. 智能判断执行模式（TUI/CLI）
2. 路由到适当的执行路径
3. 统一的错误处理

### 3. 命令执行函数

```go
// cobra/command.go:74-92
func WithRun(fn func(*Command, []string)) CommandOption {
    return func(c *Command) {
        c.Command.Run = func(cmd *spf13cobra.Command, args []string) {
            wrappedCmd := c.wrapCommand(cmd)
            fn(wrappedCmd, args)
        }
    }
}

func WithRunE(fn func(*Command, []string) error) CommandOption {
    return func(c *Command) {
        c.Command.RunE = func(cmd *spf13cobra.Command, args []string) error {
            wrappedCmd := c.wrapCommand(cmd)
            return fn(wrappedCmd, args)
        }
    }
}
```

**命令封装**：
- 将用户逻辑封装为函数
- 包装为命令对象
- 支持同步和异步执行

### 4. TUI 执行流程

```go
// cobra/command.go:205-251
func (c *Command) executeTUI() error {
    // 获取渲染器
    renderer := c.getRenderer()
    defer renderer.Cleanup()

    // 导航命令树
    selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})
    if err != nil {
        return err
    }

    if len(selectedPath) == 0 {
        return nil  // 用户取消
    }

    // 获取最终选中的命令
    selectedCmd := selectedPath[len(selectedPath)-1]

    // 配置 flags
    if c.tuiConfig.ShowFlags {
        flagValues, err := c.configureFlags(renderer, selectedCmd)
        if err != nil {
            return err
        }
        if err := c.applyFlagValues(selectedCmd, flagValues); err != nil {
            return err
        }
    }

    // 确认执行
    if c.tuiConfig.ConfirmBeforeExecute {
        confirmed, err := c.confirmExecution(renderer, selectedPath)
        if err != nil {
            return err
        }
        if !confirmed {
            return nil
        }
    }

    // 执行命令
    return c.executeCommand(selectedCmd)
}
```

**执行步骤**：
1. 获取渲染器（命令上下文）
2. 导航选择命令
3. 配置参数
4. 确认执行
5. 执行命令

## 命令树结构

### 1. 树形命令构建

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

    // 递归构建子命令
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

### 2. 命令扁平化

```go
// cobra/command_tree.go:45-59
func GetExecutableCommands(cmd *spf13cobra.Command) []*tui.CommandItem {
    root := BuildCommandTree(cmd, "")
    // 如果根命令有子命令，则只返回子命令中的可执行命令
    if len(root.Children) > 0 {
        var result []*tui.CommandItem
        for _, child := range root.Children {
            result = append(result, flattenExecutableCommands(child, "")...)
        }
        return result
    }
    // 如果根命令没有子命令，则返回根命令本身
    return flattenExecutableCommands(root, "")
}
```

### 3. 命令查找

```go
// cobra/command_tree.go:107-134
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

## 命令模式的使用

### 1. 基础命令定义

```go
// 定义根命令
var rootCmd = cobra.NewCommand("myapp",
    cobra.WithShort("My application"),
    cobra.WithLong("A CLI application with TUI support"),
)

// 定义子命令
var serverCmd = cobra.NewCommand("server",
    cobra.WithShort("Start server"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        port, _ := cmd.Flags().GetInt("port")
        fmt.Printf("Starting server on port %d\n", port)
    }),
)

// 添加 flags
serverCmd.Flags().Int("port", 8080, "Server port")

// 组合命令树
rootCmd.AddCommand(serverCmd)
```

### 2. 命令执行

```go
// 直接执行
func main() {
    rootCmd := setupCommand()
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}
```

### 3. 命令链

```go
// 支持命令链式调用
type CommandChain struct {
    commands []*Command
}

func (c *CommandChain) Add(cmd *Command) {
    c.commands = append(c.commands, cmd)
}

func (c *CommandChain) ExecuteAll() error {
    for _, cmd := range c.commands {
        if err := cmd.Execute(); err != nil {
            return err
        }
    }
    return nil
}
```

## 宏命令 (Macro Command)

### 1. 批量命令执行

```go
// 执行多个命令
func executeCommands(cmds []*Command) error {
    for _, cmd := range cmds {
        if err := cmd.Execute(); err != nil {
            return err
        }
    }
    return nil
}
```

### 2. TUI 导航作为宏命令

```go
// cobra/command.go:253-292
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
    // 获取子命令
    children := c.getChildren(cmd)
    if len(children) == 0 {
        // 叶子命令，返回当前路径
        return append(path, cmd), nil
    }

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

    // 渲染菜单
    selectedIndex, err := renderer.RenderCommandMenu(cmd.Use, menuItems)
    if err != nil {
        return nil, err
    }

    if selectedIndex < 0 {
        return nil, nil  // 用户取消
    }

    // 递归处理选中的命令
    selectedChild := children[selectedIndex]
    newPath := append(path, cmd)
    return c.navigateCommandTree(renderer, selectedChild, newPath)
}
```

**宏命令特点**：
- 组合多个子命令
- 递归执行
- 路径追踪

## 可撤销命令

### 1. 命令历史

```go
type CommandHistory struct {
    commands []CommandExecution
    index    int
}

type CommandExecution struct {
    Command *Command
    Args    []string
    Time    time.Time
}

func (h *CommandHistory) Add(cmd *Command, args []string) {
    exec := CommandExecution{
        Command: cmd,
        Args:    args,
        Time:    time.Now(),
    }
    h.commands = append(h.commands, exec)
    h.index = len(h.commands) - 1
}

func (h *CommandHistory) Undo() (*CommandExecution, error) {
    if h.index < 0 {
        return nil, errors.New("no more commands to undo")
    }
    exec := h.commands[h.index]
    h.index--
    return &exec, nil
}

func (h *CommandHistory) Redo() (*CommandExecution, error) {
    if h.index >= len(h.commands)-1 {
        return nil, errors.New("no more commands to redo")
    }
    h.index++
    return h.commands[h.index], nil
}
```

### 2. 带撤销的命令

```go
type ReversibleCommand struct {
    *Command
    undo func() error
}

func (c *ReversibleCommand) Execute() error {
    // 执行命令
    err := c.Command.Execute()
    if err != nil {
        return err
    }

    // 保存撤销函数
    // ...

    return nil
}

func (c *ReversibleCommand) Undo() error {
    if c.undo != nil {
        return c.undo()
    }
    return errors.New("undo not supported")
}
```

## 命令队列

### 1. 异步命令执行

```go
type CommandQueue struct {
    queue chan *Command
    wg    sync.WaitGroup
}

func NewCommandQueue(size int) *CommandQueue {
    return &CommandQueue{
        queue: make(chan *Command, size),
    }
}

func (q *CommandQueue) Start() {
    q.wg.Add(1)
    go func() {
        defer q.wg.Done()
        for cmd := range q.queue {
            cmd.Execute()
        }
    }()
}

func (q *CommandQueue) Enqueue(cmd *Command) {
    q.queue <- cmd
}

func (q *CommandQueue) Stop() {
    close(q.queue)
    q.wg.Wait()
}
```

### 2. 使用命令队列

```go
func main() {
    queue := NewCommandQueue(10)
    queue.Start()

    // 添加命令到队列
    for i := 0; i < 5; i++ {
        cmd := cobra.NewCommand(fmt.Sprintf("task%d", i),
            cobra.WithRun(func(cmd *cobra.Command, args []string) {
                fmt.Printf("Executing task\n")
            }),
        )
        queue.Enqueue(cmd)
    }

    // 等待完成
    queue.Stop()
}
```

## 命令日志

### 1. 执行日志记录

```go
type CommandLogger struct {
    logger *log.Logger
}

func (l *CommandLogger) LogExecution(cmd *Command, args []string, err error) {
    logEntry := map[string]interface{}{
        "command": cmd.Use,
        "args":    args,
        "time":    time.Now().Format(time.RFC3339),
        "success": err == nil,
    }
    if err != nil {
        logEntry["error"] = err.Error()
    }
    l.logger.Printf("%v", logEntry)
}

func (c *Command) ExecuteWithLogger(logger *CommandLogger) error {
    args := c.Flags().Args()
    err := c.Execute()
    logger.LogExecution(c, args, err)
    return err
}
```

## 命令模式的优势

### 1. 操作解耦

```go
// 命令创建者
creator := func() *Command {
    return cobra.NewCommand("task",
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            // 执行逻辑
        }),
    )
}

// 命令执行者
executor := func(cmd *Command) error {
    return cmd.Execute()
}

// 创建和执行解耦
cmd := creator()
executor(cmd)
```

### 2. 可组合性

```go
// 组合多个命令
type CompositeCommand struct {
    commands []*Command
}

func (c *CompositeCommand) Add(cmd *Command) {
    c.commands = append(c.commands, cmd)
}

func (c *CompositeCommand) Execute() error {
    for _, cmd := range c.commands {
        if err := cmd.Execute(); err != nil {
            return err
        }
    }
    return nil
}
```

### 3. 可撤销性

```go
// 支持撤销的操作
type ReversibleAction struct {
    do   func() error
    undo func() error
}

func (a *ReversibleAction) Execute() error {
    return a.do()
}

func (a *ReversibleAction) Undo() error {
    return a.undo()
}
```

### 4. 延迟执行

```go
// 命令可以延迟执行
func scheduleCommand(cmd *Command, delay time.Duration) {
    time.AfterFunc(delay, func() {
        cmd.Execute()
    })
}
```

## 最佳实践

### 1. 命令单一职责

```go
// 好的做法 - 每个命令只做一件事
var startCmd = cobra.NewCommand("start",
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        startService()
    }),
)

var stopCmd = cobra.NewCommand("stop",
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        stopService()
    }),
)

// 避免 - 一个命令做多件事
var controlCmd = cobra.NewCommand("control",
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        if args[0] == "start" {
            startService()
        } else if args[0] == "stop" {
            stopService()
        }
    }),
)
```

### 2. 错误处理

```go
// 使用 RunE 返回错误
var cmd = cobra.NewCommand("process",
    cobra.WithRunE(func(cmd *cobra.Command, args []string) error {
        if err := validate(); err != nil {
            return err
        }
        return process()
    }),
)
```

### 3. 命令文档

```go
// 提供清晰的命令文档
var cmd = cobra.NewCommand("export",
    cobra.WithShort("Export data to file"),
    cobra.WithLong(`Export data from the database to a specified file format.

Supported formats: JSON, CSV, XML

Examples:
  export data.json
  export --format=csv data.csv`),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        // ...
    }),
)
```

## 潜在问题和解决方案

### 问题 1: 命令参数过多

**问题**: 命令需要很多参数

**解决方案**:
```go
// 使用 flags 而非位置参数
var cmd = cobra.NewCommand("serve",
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        port, _ := cmd.Flags().GetInt("port")
        host, _ := cmd.Flags().GetString("host")
        serve(host, port)
    }),
)
cmd.Flags().Int("port", 8080, "Server port")
cmd.Flags().String("host", "localhost", "Server host")
```

### 问题 2: 命令间共享状态

**问题**: 多个命令需要共享状态

**解决方案**:
```go
// 使用预执行钩子
var rootCmd = cobra.NewCommand("app")
rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
    // 初始化共享状态
    return initSharedState()
}
```

### 问题 3: 命令测试

**问题**: 命令执行逻辑难以测试

**解决方案**:
```go
// 分离业务逻辑
func runServer(host string, port int) error {
    // 实际逻辑
}

var serverCmd = cobra.NewCommand("server",
    cobra.WithRunE(func(cmd *cobra.Command, args []string) error {
        host, _ := cmd.Flags().GetString("host")
        port, _ := cmd.Flags().GetInt("port")
        return runServer(host, port)
    }),
)

// 测试
func TestRunServer(t *testing.T) {
    err := runServer("localhost", 8080)
    assert.NoError(t, err)
}
```

## 总结

命令模式在 Cobra-X 中是核心设计模式，它提供了：

1. **操作封装**: 将执行逻辑封装为命令对象
2. **灵活执行**: 支持同步、异步、延迟执行
3. **可组合性**: 支持宏命令和命令链
4. **可扩展性**: 易于添加新命令
5. **统一接口**: 所有命令通过相同接口执行

Cobra-X 通过命令模式实现了强大的 CLI 和 TUI 功能，同时保持了代码的简洁和可维护性。这种设计使得命令的创建、组合、执行都变得非常简单和灵活。
