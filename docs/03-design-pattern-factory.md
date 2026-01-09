# 工厂模式 (Factory Pattern)

## 概述

工厂模式是一种创建型设计模式，它提供了一种创建对象的最佳方式，在创建对象时不会对客户端暴露创建逻辑。在 Cobra-X 中，工厂模式广泛用于创建各种组件实例。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **命令创建**: [cobra/command.go](../cobra/command.go) - `NewCommand()`
2. **渲染器创建**: [tui/default_renderer.go](../tui/default_renderer.go) - `NewDefaultRenderer()`
3. **主题创建**: [tui/style/theme.go](../tui/style/theme.go) - `NewTheme()`
4. **配置创建**: [cobra/tui_config.go](../cobra/tui_config.go) - `DefaultTUIConfig()`

## 代码实现分析

### 1. 简单工厂 - 命令创建

```go
// cobra/command.go:31-35
func NewCommand(use string, opts ...CommandOption) *Command {
    return newCommandWithCobra(&spf13cobra.Command{Use: use}, opts...)
}
```

**设计要点**：
- 封装对象创建逻辑
- 提供统一的入口点
- 支持选项模式配置

### 2. 带参数的工厂方法

```go
// cobra/command.go:37-55
func newCommandWithCobra(cobraCmd *spf13cobra.Command, opts ...CommandOption) *Command {
    cmd := &Command{
        Command:     cobraCmd,
        tuiConfig:   DefaultTUIConfig(),
        tuiFlags:    pflag.NewFlagSet("tui", pflag.ContinueOnError),
        children:    make([]*Command, 0),
    }

    // 应用选项
    for _, opt := range opts {
        opt(cmd)
    }

    // 初始化 TUI flags
    cmd.initTUIFlags()

    return cmd
}
```

**工厂模式的优势**：
1. **封装复杂性**: 客户端不需要知道初始化细节
2. **保证一致性**: 所有创建的对象都正确初始化
3. **支持扩展**: 通过选项模式灵活配置

### 3. 抽象工厂 - 主题系统

```go
// tui/style/theme.go:117-135
func NewTheme(name string) *Theme {
    switch name {
    case ThemeDark:
        return darkTheme()
    case ThemeLight:
        return lightTheme()
    case ThemeMinimal:
        return minimalTheme()
    case ThemeDracula:
        return draculaTheme()
    case ThemeNord:
        return nordTheme()
    case ThemeMonokai:
        return monokaiTheme()
    default:
        return defaultTheme()
    }
}
```

**抽象工厂特点**：
- 根据参数创建不同类型的产品
- 隐藏具体实现细节
- 易于添加新的主题类型

### 4. 具体工厂实现

```go
// tui/style/theme.go:203-220
func draculaTheme() *Theme {
    return &Theme{
        Name: ThemeDracula,
        Colors: ColorScheme{
            Primary:    lipgloss.Color("#BD93F9"), // purple
            Secondary:  lipgloss.Color("#6272A4"), // comment
            Success:    lipgloss.Color("#50FA7B"), // green
            Warning:    lipgloss.Color("#F1FA8C"), // yellow
            Error:      lipgloss.Color("#FF5555"), // red
            Muted:      lipgloss.Color("#6272A4"), // comment
            Background: lipgloss.Color("#282A36"), // background
            Foreground: lipgloss.Color("#F8F8F2"), // foreground
        },
        Layout: defaultLayout(),
        Styles: defaultStyles(),
    }
}
```

### 5. 渲染器工厂

```go
// tui/default_renderer.go:20-30
func NewDefaultRenderer(theme *style.Theme) *DefaultRenderer {
    if theme == nil {
        theme = style.DefaultTheme()
    }

    return &DefaultRenderer{
        theme:    theme,
        programs: make([]*tea.Program, 0),
    }
}
```

**工厂模式特点**：
- 处理默认值
- 验证参数
- 初始化必要字段

## 工厂模式与选项模式的结合

### 1. 选项函数定义

```go
// cobra/command.go:57-64
type CommandOption func(*Command)

func WithShort(short string) CommandOption {
    return func(c *Command) {
        c.Command.Short = short
    }
}

func WithLong(long string) CommandOption {
    return func(c *Command) {
        c.Command.Long = long
    }
}

func WithRun(fn func(*Command, []string)) CommandOption {
    return func(c *Command) {
        c.Command.Run = func(cmd *spf13cobra.Command, args []string) {
            wrappedCmd := c.wrapCommand(cmd)
            fn(wrappedCmd, args)
        }
    }
}
```

### 2. 使用示例

```go
rootCmd := cobra.NewCommand("myapp",
    cobra.WithShort("My application"),
    cobra.WithLong("A longer description"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        fmt.Println("Running")
    }),
    cobra.WithTUIEnabledOption(true),
)
```

**设计优势**：
1. **可读性**: 链式调用清晰明了
2. **可选参数**: 只设置需要的参数
3. **扩展性**: 易于添加新的选项
4. **类型安全**: 编译时检查

## 配置工厂 - TUI 配置

### 1. 默认配置工厂

```go
// cobra/tui_config.go:57-69
func DefaultTUIConfig() *TUIConfig {
    return &TUIConfig{
        Enabled:              false,
        Renderer:             nil,
        Theme:                style.DefaultTheme(),
        ShowDescription:      true,
        ShowFlags:            true,
        InteractiveMode:      ModeAuto,
        AutoDetect:           true,
        ConfirmBeforeExecute: true,
    }
}
```

### 2. 带选项的配置工厂

```go
// cobra/tui_config.go:72-79
func NewTUIConfig(opts ...TUIOption) *TUIConfig {
    config := DefaultTUIConfig()
    for _, opt := range opts {
        opt(config)
    }
    return config
}
```

### 3. 配置选项

```go
// cobra/tui_config.go:84-138
type TUIOption func(*TUIConfig)

func WithTUIEnabled(enabled bool) TUIOption {
    return func(c *TUIConfig) {
        c.Enabled = enabled
    }
}

func WithTUIRenderer(renderer tui.Renderer) TUIOption {
    return func(c *TUIConfig) {
        c.Renderer = renderer
    }
}

func WithTUITheme(theme *style.Theme) TUIOption {
    return func(c *TUIConfig) {
        c.Theme = theme
    }
}
```

## 使用场景示例

### 1. 创建带主题的渲染器

```go
// 方式 1: 使用默认主题
renderer := tui.NewDefaultRenderer(nil)

// 方式 2: 使用指定主题
theme := style.NewTheme("dracula")
renderer := tui.NewDefaultRenderer(theme)

// 方式 3: 通过配置
config := cobra.NewTUIConfig(
    cobra.WithTUITheme(style.NewTheme("nord")),
)
renderer := tui.NewDefaultRenderer(config.Theme)
```

### 2. 创建不同类型的命令

```go
// 根命令
rootCmd := cobra.NewCommand("app",
    cobra.WithShort("My app"),
    cobra.WithTUIEnabledOption(true),
)

// 子命令
serverCmd := cobra.NewCommand("server",
    cobra.WithShort("Start server"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        startServer()
    }),
)

// 添加到根命令
rootCmd.AddCommand(serverCmd)
```

## 工厂模式的优势

### 1. 封装创建逻辑

```go
// 不使用工厂 - 客户端需要知道所有细节
cmd := &cobra.Command{
    Command: &spf13cobra.Command{Use: "myapp"},
    tuiConfig: &cobra.TUIConfig{
        Enabled: false,
        Theme: style.NewTheme("default"),
        // ... 很多配置
    },
    tuiFlags: pflag.NewFlagSet("tui", pflag.ContinueOnError),
    // ... 更多初始化
}
cmd.initTUIFlags()

// 使用工厂 - 简洁明了
cmd := cobra.NewCommand("myapp")
```

### 2. 保证对象一致性

```go
// 所有通过工厂创建的对象都正确初始化
func NewDefaultRenderer(theme *style.Theme) *DefaultRenderer {
    if theme == nil {
        theme = style.DefaultTheme()  // 确保主题不为 nil
    }

    return &DefaultRenderer{
        theme:    theme,
        programs: make([]*tea.Program, 0),  // 初始化 slice
    }
}
```

### 3. 易于测试

```go
// 可以轻松创建测试用的对象
func TestCommand(t *testing.T) {
    cmd := cobra.NewCommand("test",
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            // 测试逻辑
        }),
    )
    // 测试...
}
```

### 4. 集中管理创建逻辑

```go
// 所有主题的创建都在 NewTheme 中
// 添加新主题只需修改一处
func NewTheme(name string) *Theme {
    switch name {
    // ... 现有主题
    case "newtheme":
        return newTheme()  // 新主题
    default:
        return defaultTheme()
    }
}
```

## 高级技巧

### 1. 延迟初始化

```go
// 懒加载主题
func (c *Command) getTheme() *style.Theme {
    if c.tuiConfig != nil && c.tuiConfig.Theme != nil {
        return c.tuiConfig.Theme
    }

    // 从 flags 读取主题名称
    themeName := "default"
    if theme, err := c.Flags().GetString("tui-theme"); err == nil {
        themeName = theme
    }

    return style.NewTheme(themeName)
}
```

### 2. 工厂缓存

```go
// 缓存已创建的主题
var themeCache = make(map[string]*style.Theme)

func NewThemeCached(name string) *Theme {
    if theme, ok := themeCache[name]; ok {
        return theme
    }

    theme := NewTheme(name)
    themeCache[name] = theme
    return theme
}
```

### 3. 建造者工厂结合

```go
// 工厂方法返回建造者
func NewCommandBuilder(use string) *CommandBuilder {
    return &CommandBuilder{
        cmd: &Command{
            Command: &spf13cobra.Command{Use: use},
        },
    }
}

// 使用建造者
cmd := cobra.NewCommandBuilder("myapp").
    Short("My app").
    Long("Description").
    EnableTUI(true).
    Theme("dracula").
    Build()
```

## 最佳实践

### 1. 命名约定

```go
// 工厂函数以 New 开头
func NewCommand() *Command {}
func NewDefaultRenderer() *DefaultRenderer {}
func NewTheme() *Theme {}

// 或返回默认值
func DefaultTUIConfig() *TUIConfig {}
func DefaultTheme() *Theme {}
```

### 2. 参数验证

```go
func NewDefaultRenderer(theme *style.Theme) *DefaultRenderer {
    if theme == nil {
        theme = style.DefaultTheme()  // 提供默认值
    }
    // 验证其他参数...
}
```

### 3. 返回接口类型

```go
// 返回接口而非具体类型
func NewRenderer(theme *style.Theme) Renderer {
    return NewDefaultRenderer(theme)
}

// 客户端依赖接口
var renderer Renderer = NewRenderer(theme)
```

## 潜在问题和解决方案

### 问题 1: 工厂函数过多

**问题**: 随着项目增长，工厂函数可能变得过多

**解决方案**:
```go
// 按功能分组
type Factory struct {
    themeFactory    ThemeFactory
    rendererFactory RendererFactory
    commandFactory  CommandFactory
}

factory := NewFactory()
cmd := factory.CreateCommand("myapp")
```

### 问题 2: 参数过多

**问题**: 工厂函数参数过多导致可读性差

**解决方案**:
```go
// 使用选项模式
cmd := cobra.NewCommand("myapp",
    cobra.WithShort("..."),
    cobra.WithLong("..."),
    cobra.WithRun(...),
    // ... 更多选项
)
```

### 问题 3: 循环依赖

**问题**: 工厂之间可能存在循环依赖

**解决方案**:
```go
// 使用依赖注入
func NewCommandFactory(rendererFactory RendererFactory) *CommandFactory {
    return &CommandFactory{
        rendererFactory: rendererFactory,
    }
}
```

## 总结

工厂模式在 Cobra-X 中扮演着重要角色：

1. **简化对象创建**: 封装复杂的初始化逻辑
2. **保证一致性**: 确保所有对象正确初始化
3. **提高可维护性**: 集中管理创建逻辑
4. **支持扩展**: 通过选项模式灵活配置

与选项模式结合使用，工厂模式在 Go 中变得更加优雅和强大，避免了构造函数参数过多的问题，同时保持了类型安全和可读性。
