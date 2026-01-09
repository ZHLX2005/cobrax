# 建造者模式 (Builder Pattern)

## 概述

建造者模式是一种创建型设计模式，它允许你分步骤创建复杂对象。在 Go 中，建造者模式通常以"选项模式"（Functional Options Pattern）的形式实现。Cobra-X 大量使用这种模式来构建命令、配置和主题。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **命令构建**: [cobra/command.go](../cobra/command.go) - `CommandOption`
2. **配置构建**: [cobra/tui_config.go](../cobra/tui_config.go) - `TUIOption`
3. **装饰器配置**: [cobra/decorate.go](../cobra/decorate.go) - `EnhanceOption`

## 选项模式 (Functional Options) 实现

### 1. 基础结构

```go
// cobra/command.go:57-64
// CommandOption 命令配置选项
type CommandOption func(*Command)

// WithShort 设置短描述
func WithShort(short string) CommandOption {
    return func(c *Command) {
        c.Command.Short = short
    }
}

// WithLong 设置长描述
func WithLong(long string) CommandOption {
    return func(c *Command) {
        c.Command.Long = long
    }
}
```

### 2. 命令构建实现

```go
// cobra/command.go:31-35
func NewCommand(use string, opts ...CommandOption) *Command {
    return newCommandWithCobra(&spf13cobra.Command{Use: use}, opts...)
}

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

### 3. 复杂选项实现

```go
// cobra/command.go:74-82
// WithRun 设置命令执行函数
func WithRun(fn func(*Command, []string)) CommandOption {
    return func(c *Command) {
        c.Command.Run = func(cmd *spf13cobra.Command, args []string) {
            wrappedCmd := c.wrapCommand(cmd)
            fn(wrappedCmd, args)
        }
    }
}

// cobra/command.go:84-92
// WithRunE 设置带返回值的命令执行函数
func WithRunE(fn func(*Command, []string) error) CommandOption {
    return func(c *Command) {
        c.Command.RunE = func(cmd *spf13cobra.Command, args []string) error {
            wrappedCmd := c.wrapCommand(cmd)
            return fn(wrappedCmd, args)
        }
    }
}
```

**设计亮点**：
- 选项函数返回闭包
- 闭包捕获配置参数
- 延迟应用到目标对象

## TUI 配置构建

### 1. 配置选项定义

```go
// cobra/tui_config.go:81-138
// TUIOption TUI 配置选项函数
type TUIOption func(*TUIConfig)

// WithTUIEnabled 设置是否启用 TUI
func WithTUIEnabled(enabled bool) TUIOption {
    return func(c *TUIConfig) {
        c.Enabled = enabled
    }
}

// WithTUIRenderer 设置自定义渲染器
func WithTUIRenderer(renderer tui.Renderer) TUIOption {
    return func(c *TUIConfig) {
        c.Renderer = renderer
    }
}

// WithTUITheme 设置主题
func WithTUITheme(theme *style.Theme) TUIOption {
    return func(c *TUIConfig) {
        c.Theme = theme
    }
}
```

### 2. 配置构建实现

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

### 3. 默认配置作为基础

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

**设计优势**：
- 提供合理的默认值
- 用户只需覆盖需要修改的部分
- 保证对象始终处于有效状态

## 使用示例

### 1. 简单命令构建

```go
// 最简单的命令
rootCmd := cobra.NewCommand("myapp")

// 添加描述
rootCmd := cobra.NewCommand("myapp",
    cobra.WithShort("My application"),
    cobra.WithLong("A longer description of my application"),
)
```

### 2. 完整功能命令

```go
// 带所有功能的命令
serverCmd := cobra.NewCommand("server",
    cobra.WithShort("Start the server"),
    cobra.WithLong("Start the web server on the specified port"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        port, _ := cmd.Flags().GetInt("port")
        startServer(port)
    }),
)

// 添加 flags
serverCmd.Flags().Int("port", 8080, "Server port")
```

### 3. TUI 配置构建

```go
// 基础配置
config := cobra.NewTUIConfig(
    cobra.WithTUIEnabled(true),
    cobra.WithTUITheme(style.NewTheme("dracula")),
)

// 复杂配置
config := cobra.NewTUIConfig(
    cobra.WithTUIEnabled(true),
    cobra.WithTUIRenderer(&MyCustomRenderer{}),
    cobra.WithTUITheme(style.NewTheme("nord")),
    cobra.WithTUIShowDescription(true),
    cobra.WithTUIShowFlags(true),
    cobra.WithTUIInteractiveMode(cobra.ModeAuto),
    cobra.WithTUIAutoDetect(true),
    cobra.WithTUIConfirmBeforeExecute(true),
)
```

### 4. 组合使用

```go
// 创建带 TUI 的命令
rootCmd := cobra.NewCommand("myapp",
    cobra.WithShort("My application"),
    cobra.WithTUIEnabledOption(true),
    cobra.WithTUIConfig(&cobra.TUIConfig{
        Theme: style.NewTheme("monokai"),
        ShowFlags: true,
    }),
)

// 或者使用配置选项
rootCmd := cobra.NewCommand("myapp",
    cobra.WithShort("My application"),
    cobra.WithTUIConfig(cobra.NewTUIConfig(
        cobra.WithTUIEnabled(true),
        cobra.WithTUITheme(style.NewTheme("dracula")),
    )),
)
```

## 选项模式的优势

### 1. 可读性

```go
// 传统方式 - 参数顺序难以记忆
NewCommand("myapp", "short desc", "long desc", true, nil, runFunc)

// 选项模式 - 清晰明了
NewCommand("myapp",
    WithShort("short desc"),
    WithLong("long desc"),
    WithTUIEnabledOption(true),
    WithRun(runFunc),
)
```

### 2. 可扩展性

```go
// 添加新选项不影响现有代码
func WithTimeout(timeout time.Duration) CommandOption {
    return func(c *Command) {
        c.timeout = timeout
    }
}

// 用户可以选择使用
cmd := NewCommand("myapp", WithTimeout(30*time.Second))
```

### 3. 可选参数

```go
// 只设置需要的参数
cmd := NewCommand("myapp",
    WithShort("My app"),  // 只设置短描述
    // 其他使用默认值
)
```

### 4. 参数顺序无关

```go
// 选项顺序不影响结果
cmd1 := NewCommand("myapp",
    WithShort("desc"),
    WithTUIEnabledOption(true),
)

cmd2 := NewCommand("myapp",
    WithTUIEnabledOption(true),
    WithShort("desc"),
)

// cmd1 和 cmd2 等价
```

## 高级技巧

### 1. 选项组合

```go
// 组合多个选项为单个选项
func WithServerDefaults() CommandOption {
    return func(c *Command) {
        WithShort("Server")(c)
        WithLong("Server command")(c)
        WithTUIEnabledOption(true)(c)
        WithRun(serverRun)(c)
    }
}

// 使用
serverCmd := NewCommand("server", WithServerDefaults())
```

### 2. 条件选项

```go
// 条件应用选项
func WithDebug(debug bool) CommandOption {
    if debug {
        return func(c *Command) {
            c.debug = true
            c.logLevel = "debug"
        }
    }
    return func(c *Command) {} // 空操作
}

// 使用
debugMode := os.Getenv("DEBUG") == "true"
cmd := NewCommand("myapp", WithDebug(debugMode))
```

### 3. 验证选项

```go
// 带验证的选项
func WithPort(port int) CommandOption {
    return func(c *Command) {
        if port < 1 || port > 65535 {
            panic(fmt.Sprintf("invalid port: %d", port))
        }
        c.port = port
    }
}
```

### 4. 选项链

```go
// 支持链式调用的选项
type CommandBuilder struct {
    cmd *Command
}

func (b *CommandBuilder) Short(desc string) *CommandBuilder {
    b.cmd.Command.Short = desc
    return b
}

func (b *CommandBuilder) Long(desc string) *CommandBuilder {
    b.cmd.Command.Long = desc
    return b
}

func (b *CommandBuilder) EnableTUI() *CommandBuilder {
    b.cmd.EnableTUI()
    return b
}

func (b *CommandBuilder) Build() *Command {
    return b.cmd
}

// 使用
cmd := NewCommandBuilder("myapp").
    Short("My app").
    Long("Description").
    EnableTUI().
    Build()
```

### 5. 选项缓存

```go
// 缓存常用配置
var DefaultTUIOptions = []TUIOption{
    WithTUIEnabled(true),
    WithTUITheme(style.DefaultTheme()),
    WithTUIAutoDetect(true),
}

// 使用
config := NewTUIConfig(DefaultTUIOptions...)
```

## 与传统建造者模式对比

### 传统建造者模式（Java 风格）

```java
// 传统建造者
public class Command {
    private String use;
    private String short;
    private String long;

    private Command(Builder builder) {
        this.use = builder.use;
        this.short = builder.short;
        this.long = builder.long;
    }

    public static class Builder {
        private String use;
        private String short;
        private String long;

        public Builder use(String use) {
            this.use = use;
            return this;
        }

        public Builder short(String short) {
            this.short = short;
            return this;
        }

        public Builder long(String long) {
            this.long = long;
            return this;
        }

        public Command build() {
            return new Command(this);
        }
    }
}

// 使用
Command cmd = new Command.Builder()
    .use("myapp")
    .short("My app")
    .long("Description")
    .build();
```

### Go 选项模式

```go
// Go 选项模式
func NewCommand(use string, opts ...CommandOption) *Command {
    cmd := &Command{
        Command: &spf13cobra.Command{Use: use},
        tuiConfig: DefaultTUIConfig(),
        // ...
    }
    for _, opt := range opts {
        opt(cmd)
    }
    return cmd
}

// 使用
cmd := NewCommand("myapp",
    WithShort("My app"),
    WithLong("Description"),
)
```

**对比总结**：

| 特性 | 传统建造者 | 选项模式 |
|------|-----------|----------|
| 代码量 | 较多（需要 Builder 类） | 较少 |
| 类型安全 | 编译时检查 | 编译时检查 |
| 可扩展性 | 需修改 Builder | 添加新函数即可 |
| 默认值 | 需显式设置 | 可提供默认值 |
| 内存分配 | 创建 Builder 对象 | 直接创建目标对象 |

## 最佳实践

### 1. 提供默认值

```go
// 好的做法
func NewCommand(use string, opts ...CommandOption) *Command {
    cmd := &Command{
        Command: &spf13cobra.Command{Use: use},
        tuiConfig: DefaultTUIConfig(),  // 默认配置
        // ...
    }
    // ...
}
```

### 2. 选项命名规范

```go
// 统一使用 With 前缀
func WithShort(short string) CommandOption {}
func WithLong(long string) CommandOption {}
func WithRun(fn func(*Command, []string)) CommandOption {}

// 或使用更具体的前缀
func WithTUIEnabled(enabled bool) CommandOption {}
func WithTUITheme(theme *style.Theme) CommandOption {}
```

### 3. 避免选项冲突

```go
// 通过文档说明选项优先级
// WithTUIConfig 会覆盖其他单独的选项
func WithTUIConfig(config *TUIConfig) CommandOption {
    return func(c *Command) {
        c.tuiConfig = config
    }
}
```

### 4. 支持选项撤销

```go
// 提供撤销选项
func WithoutTUI() CommandOption {
    return func(c *Command) {
        c.DisableTUI()
    }
}
```

## 潜在问题和解决方案

### 问题 1: 选项过多导致混乱

**问题**: 随着功能增加，选项函数可能非常多

**解决方案**:
```go
// 按功能分组
type CommandOptions struct {
    basic    []CommandOption
    tui      []CommandOption
    advanced []CommandOption
}

func WithOptions(opts *CommandOptions) CommandOption {
    return func(c *Command) {
        for _, opt := range opts.basic {
            opt(c)
        }
        // ...
    }
}
```

### 问题 2: 选项依赖

**问题**: 某些选项依赖其他选项

**解决方案**:
```go
// 提供组合选项
func WithTUIEnabled() CommandOption {
    return func(c *Command) {
        WithTUIConfig(&TUIConfig{
            Enabled: true,
            // 设置相关依赖
        })(c)
    }
}
```

### 问题 3: 选项验证时机

**问题**: 何时验证选项配置的有效性

**解决方案**:
```go
// 延迟验证
func (c *Command) Validate() error {
    if c.tuiConfig != nil && c.tuiConfig.Enabled {
        if c.tuiConfig.Theme == nil {
            return errors.New("theme required when TUI enabled")
        }
    }
    return nil
}

// 在构建后验证
cmd := NewCommand("myapp", opts...)
if err := cmd.Validate(); err != nil {
    log.Fatal(err)
}
```

## 总结

建造者模式（选项模式）是 Go 语言中构建复杂对象的优雅方式，在 Cobra-X 中：

1. **命令构建**: 通过选项灵活配置命令
2. **配置构建**: 支持复杂的 TUI 配置
3. **可读性**: 链式调用清晰明了
4. **可扩展性**: 添加新选项无需修改现有代码
5. **默认值**: 提供合理的默认行为

选项模式体现了 Go 语言简洁而强大的哲学，通过函数式编程的方式实现了建造者模式的核心价值，同时避免了传统建造者模式的繁琐。这种模式使得 Cobra-X 的 API 既强大又易用。
