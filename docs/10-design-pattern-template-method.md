# 模板方法模式 (Template Method Pattern)

## 概述

模板方法模式是一种行为型设计模式，它在父类中定义了一个算法的框架，允许子类在不改变算法结构的情况下重写算法的特定步骤。在 Cobra-X 中，模板方法模式主要用于定义命令执行的标准化流程。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **TUI 执行流程**: [cobra/command.go](../cobra/command.go) - `executeTUI()`
2. **命令导航**: [cobra/command.go](../cobra/command.go) - `navigateCommandTree()`
3. **装饰器执行**: [cobra/decorate.go](../cobra/decorate.go) - `navigateAndExecute()`
4. **渲染流程**: [tui/default_renderer.go](../tui/default_renderer.go) - BubbleTea 的 Model 接口

## 代码实现分析

### 1. TUI 执行模板方法

```go
// cobra/command.go:205-251
func (c *Command) executeTUI() error {
    // 1. 获取渲染器（初始化步骤）
    renderer := c.getRenderer()
    defer renderer.Cleanup()

    // 2. 导航命令树（可扩展步骤）
    selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})
    if err != nil {
        return err
    }

    if len(selectedPath) == 0 {
        return nil  // 用户取消
    }

    // 3. 获取最终选中的命令
    selectedCmd := selectedPath[len(selectedPath)-1]

    // 4. 配置 flags（可扩展步骤）
    if c.tuiConfig.ShowFlags {
        flagValues, err := c.configureFlags(renderer, selectedCmd)
        if err != nil {
            return err
        }

        if err := c.applyFlagValues(selectedCmd, flagValues); err != nil {
            return err
        }
    }

    // 5. 确认执行（可扩展步骤）
    if c.tuiConfig.ConfirmBeforeExecute {
        confirmed, err := c.confirmExecution(renderer, selectedPath)
        if err != nil {
            return err
        }
        if !confirmed {
            return nil
        }
    }

    // 6. 执行命令（不可变的核心步骤）
    return c.executeCommand(selectedCmd)
}
```

**模板方法要素**：
- **模板方法**: `executeTUI()` 定义执行流程
- **原语操作**: `navigateCommandTree()`, `configureFlags()` 等具体步骤
- **钩子方法**: `confirmExecution()` 可以被覆盖
- **固定步骤**: `executeCommand()` 是核心执行逻辑

### 2. 命令导航模板

```go
// cobra/command.go:253-292
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
    // 1. 获取子命令
    children := c.getChildren(cmd)

    // 2. 检查是否为叶子节点
    if len(children) == 0 {
        return append(path, cmd), nil
    }

    // 3. 构建菜单项
    menuItems := make([]tui.MenuItem, 0, len(children))
    for _, child := range children {
        menuItems = append(menuItems, tui.MenuItem{
            ID:          child.Use,
            Label:       child.Use,
            Description: child.Short,
            Disabled:    !child.IsAvailableCommand(),
        })
    }

    // 4. 渲染菜单
    selectedIndex, err := renderer.RenderCommandMenu(cmd.Use, menuItems)
    if err != nil {
        return nil, err
    }

    if selectedIndex < 0 {
        return nil, nil  // 用户取消
    }

    // 5. 递归处理选中的命令
    selectedChild := children[selectedIndex]
    newPath := append(path, cmd)
    return c.navigateCommandTree(renderer, selectedChild, newPath)
}
```

**递归模板**：
- 基础情况：叶子节点
- 递归情况：有子节点时继续导航
- 统一的返回格式

### 3. Flag 配置模板

```go
// cobra/command.go:309-319
func (c *Command) configureFlags(renderer tui.Renderer, cmd *Command) (map[string]string, error) {
    // 1. 收集 flag 项
    flagItems := c.collectFlagItems(cmd)

    if len(flagItems) == 0 {
        return nil, nil
    }

    // 2. 渲染 flag 表单
    return renderer.RenderFlagForm("Configure: "+cmd.Use, flagItems)
}
```

**两步模板**：
1. 收集数据（可扩展）
2. 渲染表单（委托给渲染器）

### 4. BubbleTea 模板方法

```go
// tui/default_renderer.go:142-169
type menuModel struct {
    items           []MenuItem
    cursor          int
    cancelled       bool
    theme           *style.Theme
    width           int
    height          int
    quitting        bool
    showDescription bool
}

// Init - 初始化步骤
func (m *menuModel) Init() tea.Cmd {
    return nil
}

// Update - 更新步骤（核心模板方法）
func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        return m.handleKeyMsg(msg)
    case tea.WindowSizeMsg:
        return m.handleWindowSizeMsg(msg)
    }
    return m, nil
}

// View - 渲染步骤
func (m *menuModel) View() string {
    // 渲染逻辑...
}
```

**Elm 架构模板**：
- `Init()`: 初始化
- `Update()`: 处理消息
- `View()`: 渲染视图

## 钩子方法

### 1. 可选的确认步骤

```go
// cobra/command.go:433-442
func (c *Command) confirmExecution(renderer tui.Renderer, path []*Command) (bool, error) {
    // 构建命令字符串
    cmdString := c.buildCommandString(path)

    // 渲染确认对话框
    return renderer.RenderConfirmation(
        "Confirm Execution",
        fmt.Sprintf("Command to execute:\n\n  %s", cmdString),
    )
}
```

**钩子特点**：
- 可以通过配置启用/禁用
- 不影响主流程
- 可以被覆盖

### 2. 自定义执行前处理

```go
// 允许用户添加自定义的执行前逻辑
type Command struct {
    *spf13cobra.Command
    // ...
    preExecuteHook func(*Command) error
}

func (c *Command) executeTUI() error {
    // 获取渲染器
    renderer := c.getRenderer()
    defer renderer.Cleanup()

    // 调用前置钩子
    if c.preExecuteHook != nil {
        if err := c.preExecuteHook(c); err != nil {
            return err
        }
    }

    // 继续标准流程...
    selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})
    // ...
}
```

### 3. 自定义 Flag 收集

```go
// 允许自定义 flag 收集逻辑
type FlagCollector func(*Command) []tui.FlagItem

type Command struct {
    // ...
    flagCollector FlagCollector
}

func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    // 如果有自定义收集器，使用它
    if c.flagCollector != nil {
        return c.flagCollector(cmd)
    }

    // 否则使用默认逻辑
    return c.defaultCollectFlagItems(cmd)
}
```

## 模板方法的变体

### 1. 装饰器版本的执行模板

```go
// cobra/decorate.go:208-278
func navigateAndExecute(renderer tui.Renderer, cmd *spf13cobra.Command, config *EnhanceConfig) error {
    // 1. 获取所有可执行命令（扁平化列表）
    executableCommands := GetExecutableCommands(cmd)

    // 2. 如果只有一个可执行命令且是当前命令，直接执行
    if len(executableCommands) == 1 && executableCommands[0].ID == cmd.Name() && (cmd.Run != nil || cmd.RunE != nil) {
        return executeLeafCommand(renderer, cmd, config)
    }

    // 3. 如果有多个可执行命令，显示扁平化菜单
    if len(executableCommands) > 0 {
        // 构建菜单项
        menuItems := make([]tui.MenuItem, 0, len(executableCommands))
        for _, execCmd := range executableCommands {
            // ... 构建逻辑
        }

        // 4. 显示菜单让用户选择
        selectedIndex, err := renderer.RenderCommandMenu(cmd.Name()+" Commands", menuItems)
        if err != nil {
            return err
        }

        if selectedIndex < 0 {
            return nil  // 用户取消
        }

        // 5. 执行选中的命令
        // ... 查找和执行逻辑
    }

    // 6. 如果没有可执行的子命令，尝试执行当前命令
    if cmd.Run != nil || cmd.RunE != nil {
        return executeLeafCommand(renderer, cmd, config)
    }

    // 7. 显示帮助
    return cmd.Help()
}
```

**与主模板的区别**：
- 使用扁平化视图而非树形导航
- 适用于装饰器模式
- 更简化的执行流程

### 2. 执行叶子命令的子模板

```go
// cobra/decorate.go:280-309
func executeLeafCommand(renderer tui.Renderer, cmd *spf13cobra.Command, config *EnhanceConfig) error {
    // 1. 配置 flags
    if config.TUIConfig != nil && config.TUIConfig.ShowFlags {
        flagValues, err := configureFlags(renderer, cmd)
        if err != nil {
            return err
        }

        // 应用 flag 值
        applyFlagValues(cmd, flagValues)
    }

    // 2. 确认执行
    if config.TUIConfig != nil && config.TUIConfig.ConfirmBeforeExecute {
        confirmed, err := renderer.RenderConfirmation(
            "Confirm",
            buildCommandPreview(cmd),
        )
        if err != nil {
            return err
        }
        if !confirmed {
            return nil
        }
    }

    // 3. 执行命令
    return executeOriginalCommand(cmd)
}
```

## 实现自定义模板方法

### 1. 自定义执行流程

```go
// 自定义命令，覆盖执行流程
type CustomCommand struct {
    *cobra.Command
}

func (c *CustomCommand) executeTUI() error {
    // 添加自定义的前置处理
    if err := c.validateEnvironment(); err != nil {
        return err
    }

    // 调用父类的标准流程
    if err := c.Command.executeTUI(); err != nil {
        return err
    }

    // 添加自定义的后置处理
    return c.cleanupAfterExecution()
}
```

### 2. 自定义导航逻辑

```go
// 跳过某些中间命令的导航
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
    children := c.getChildren(cmd)

    // 跳过没有可执行子命令的中间节点
    if len(children) > 0 {
        hasRunnable := false
        for _, child := range children {
            if child.IsAvailableCommand() {
                hasRunnable = true
                break
            }
        }

        if !hasRunnable {
            // 跳过当前节点，直接返回第一个子命令
            return c.navigateCommandTree(renderer, children[0], path)
        }
    }

    // 标准导航逻辑...
}
```

### 3. 自定义 Flag 收集

```go
// 只收集特定类型的 flags
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem
    seen := make(map[string]bool)

    current := cmd
    for current != nil {
        current.LocalFlags().VisitAll(func(flag *pflag.Flag) {
            if shouldSkipFlag(flag) {  // 自定义过滤逻辑
                return
            }

            // 标准收集逻辑...
            item := tui.FlagItem{
                Name:         flag.Name,
                Description:  flag.Usage,
                DefaultValue: flag.DefValue,
                CurrentValue: flag.DefValue,
                SourceCommand: current.Name(),
            }

            // 类型判断...
            switch flag.Value.Type() {
            case "bool":
                item.Type = tui.FlagTypeBool
            // ...
            }

            items = append(items, item)
            seen[flag.Name] = true
        })

        if current.Parent() == nil {
            break
        }
        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

## 模板方法模式的优势

### 1. 算法框架固定

```go
// 执行流程始终保持一致
func (c *Command) executeTUI() error {
    // 1. 获取渲染器
    renderer := c.getRenderer()

    // 2. 导航选择命令
    selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})

    // 3. 配置参数
    flagValues, err := c.configureFlags(renderer, selectedCmd)

    // 4. 确认执行
    confirmed, err := c.confirmExecution(renderer, selectedPath)

    // 5. 执行命令
    return c.executeCommand(selectedCmd)
}
```

### 2. 步骤可扩展

```go
// 用户可以扩展特定步骤
type MyCommand struct {
    *cobra.Command
}

func (m *MyCommand) configureFlags(renderer tui.Renderer, cmd *Command) (map[string]string, error) {
    // 添加自定义的 flag 配置逻辑
    // 例如：验证 flag 值、提供默认值等
}
```

### 3. 代码复用

```go
// 通用逻辑只需实现一次
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    // 适用于所有命令的通用逻辑
}

// 不同命令可以共享相同的执行流程
```

### 4. 易于维护

```go
// 修改流程只需修改模板方法
func (c *Command) executeTUI() error {
    // 在这里添加新的步骤
    // 所有命令都会自动使用新流程
}
```

## 高级技巧

### 1. 带回滚的模板方法

```go
func (c *Command) executeTUIWithRollback() error {
    // 保存初始状态
    initialState := c.captureState()
    defer c.restoreState(initialState)

    // 执行流程
    if err := c.executeTUI(); err != nil {
        // 出错时回滚
        return err
    }

    return nil
}
```

### 2. 带重试的模板方法

```go
func (c *Command) executeTUIWithRetry(maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := c.executeTUI()
        if err == nil {
            return nil
        }

        if i < maxRetries-1 {
            log.Printf("Execution failed (attempt %d/%d): %v", i+1, maxRetries, err)
            time.Sleep(time.Second)
        }
    }
    return errors.New("max retries exceeded")
}
```

### 3. 带超时的模板方法

```go
func (c *Command) executeTUIWithTimeout(timeout time.Duration) error {
    done := make(chan error, 1)

    go func() {
        done <- c.executeTUI()
    }()

    select {
    case err := <-done:
        return err
    case <-time.After(timeout):
        return errors.New("execution timeout")
    }
}
```

## 最佳实践

### 1. 明确标识模板方法

```go
// executeTUI 定义了 TUI 模式的标准执行流程：
// 1. 获取渲染器
// 2. 导航命令树
// 3. 配置 flags
// 4. 确认执行
// 5. 执行命令
func (c *Command) executeTUI() error {
    // ...
}
```

### 2. 合理设计钩子

```go
// 钩子应该有合理的默认行为
func (c *Command) confirmExecution(renderer tui.Renderer, path []*Command) (bool, error) {
    // 默认实现
    if !c.tuiConfig.ConfirmBeforeExecute {
        return true, nil  // 跳过确认
    }
    // 显示确认对话框
}
```

### 3. 保持步骤独立

```go
// 每个步骤应该是独立的
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
    // 不依赖于其他步骤的内部状态
}

func (c *Command) configureFlags(renderer tui.Renderer, cmd *Command) (map[string]string, error) {
    // 独立的 flag 配置逻辑
}
```

## 潜在问题和解决方案

### 问题 1: 模板方法过于复杂

**问题**: 模板方法包含太多步骤，难以理解和维护

**解决方案**:
```go
// 分解为更小的模板方法
func (c *Command) executeTUI() error {
    renderer := c.getRenderer()
    defer renderer.Cleanup()

    // 分解为独立的阶段
    selectedPath, err := c.selectCommand(renderer)
    if err != nil {
        return err
    }

    if err := c.configureCommand(renderer, selectedPath); err != nil {
        return err
    }

    return c.executeSelected(selectedPath)
}

func (c *Command) selectCommand(renderer tui.Renderer) ([]*Command, error) {
    // 专注于命令选择
}

func (c *Command) configureCommand(renderer tui.Renderer, path []*Command) error {
    // 专注于配置
}

func (c *Command) executeSelected(path []*Command) error {
    // 专注于执行
}
```

### 问题 2: 步骤间数据传递复杂

**问题**: 不同步骤间需要传递大量数据

**解决方案**:
```go
// 使用上下文对象
type ExecutionContext struct {
    SelectedPath []*Command
    FlagValues   map[string]string
    Confirmed    bool
}

func (c *Command) executeTUI() error {
    ctx := &ExecutionContext{}

    renderer := c.getRenderer()
    defer renderer.Cleanup()

    if err := c.selectCommand(renderer, ctx); err != nil {
        return err
    }

    if err := c.configureCommand(renderer, ctx); err != nil {
        return err
    }

    return c.executeSelected(ctx)
}
```

### 问题 3: 钩子方法失控

**问题**: 过多的钩子方法导致混乱

**解决方案**:
```go
// 使用中间件模式
type ExecutionMiddleware func(*Command, ExecutionFunc) error

type ExecutionFunc func(*Command) error

func WithLogging(next ExecutionFunc) ExecutionMiddleware {
    return func(cmd *Command, f ExecutionFunc) error {
        log.Printf("Executing: %s", cmd.Use)
        err := f(cmd)
        log.Printf("Completed: %s, error: %v", cmd.Use, err)
        return err
    }
}

func (c *Command) ExecuteWith(middlewares ...ExecutionMiddleware) error {
    var f ExecutionFunc = func(cmd *Command) error {
        return cmd.executeTUI()
    }

    // 应用中间件
    for i := len(middlewares) - 1; i >= 0; i-- {
        middleware := middlewares[i]
        next := f
        f = func(cmd *Command) error {
            return middleware(cmd, next)
        }
    }

    return f(c)
}
```

## 总结

模板方法模式在 Cobra-X 中实现了：

1. **标准化流程**: 定义了命令执行的统一流程
2. **扩展点**: 提供了可扩展的钩子方法
3. **代码复用**: 通用逻辑只需实现一次
4. **一致性**: 所有命令使用相同的执行模式
5. **灵活性**: 允许在保持结构不变的情况下自定义步骤

模板方法模式使得 Cobra-X 能够保持一致的执行流程，同时允许用户根据需要自定义特定步骤。这种设计平衡了标准化和灵活性，是构建复杂系统的优秀实践。通过合理使用模板方法模式，Cobra-X 实现了清晰、可维护且可扩展的代码结构。
