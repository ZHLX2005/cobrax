# 适配器模式 (Adapter Pattern)

## 概述

适配器模式是一种结构型设计模式，它允许接口不兼容的类一起工作。适配器模式通过包装一个对象（通常称为被适配者）并将其转换为另一个接口。在 Cobra-X 中，适配器模式主要用于桥接不同的渲染系统和命令接口。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **渲染器适配**: [tui/renderer.go](../tui/renderer.go) - `Renderer` 接口
2. **命令包装**: [cobra/command.go](../cobra/command.go) - `wrapCommand()` 方法
3. **类型转换**: [cobra/command.go](../cobra/command.go) - `Command` 结构嵌入
4. **装饰器适配**: [cobra/decorate.go](../cobra/decorate.go) - `Enhance()` 函数

## 代码实现分析

### 1. 渲染器接口适配

```go
// tui/renderer.go:7-33
type Renderer interface {
    RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error)
    RenderFlagForm(formTitle string, flags []FlagItem) (values map[string]string, err error)
    RenderConfirmation(title, message string) (confirmed bool, err error)
    RenderHelp(title, content string) error
    Cleanup() error
}
```

**适配器模式要素**：
- **目标接口**: `Renderer` 接口
- **适配器**: `DefaultRenderer` 实现
- **被适配者**: BubbleTea 框架
- **客户端**: `Command` 结构

### 2. 默认渲染器适配器

```go
// tui/default_renderer.go:13-18
type DefaultRenderer struct {
    theme    *style.Theme
    programs []*tea.Program
}

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

### 3. BubbleTea 框架适配

```go
// tui/default_renderer.go:32-59
func (r *DefaultRenderer) RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error) {
    // 获取终端尺寸
    width, height := getTerminalSize()

    // 创建菜单模型（BubbleTea 的 Model）
    model := newMenuModel(options, r.theme, width, height)

    // 创建并运行 BubbleTea 程序
    p := tea.NewProgram(model, tea.WithAltScreen())
    r.programs = append(r.programs, p)

    result, err := p.Run()
    if err != nil {
        return -1, fmt.Errorf("failed to run menu: %w", err)
    }

    menuResult, ok := result.(*menuModel)
    if !ok {
        return -1, fmt.Errorf("unexpected result type")
    }

    if menuResult.cancelled {
        return -1, nil
    }

    return menuResult.cursor, nil
}
```

**适配过程**：
1. 将 Cobra-X 的数据结构转换为 BubbleTea 的 Model
2. 启动 BubbleTea 程序
3. 将 BubbleTea 的结果转换回 Cobra-X 的格式

## 命令类型适配

### 1. 命令包装器

```go
// cobra/command.go:94-103
func (c *Command) wrapCommand(cmd *spf13cobra.Command) *Command {
    // 由于嵌入的关系，我们无法直接做类型断言
    // 检查命令是否已经被我们包装过
    // 这里简化处理：总是创建新的包装
    return &Command{
        Command:   cmd,
        tuiConfig: c.tuiConfig,
    }
}
```

### 2. 双向适配

```go
// 添加 cobra.Command 子命令
func (c *Command) AddCommand(cmds ...*Command) {
    for _, cmd := range cmds {
        c.Command.AddCommand(cmd.Command)
    }
}

// 添加原始 spf13/cobra 命令
func (c *Command) AddSpf13Command(cmds ...*spf13cobra.Command) {
    c.Command.AddCommand(cmds...)
}
```

**双向适配**：
- `AddCommand`: 接受 `*Command` 类型
- `AddSpf13Command`: 接受 `*spf13cobra.Command` 类型
- 两者都可以无缝使用

### 3. 执行函数适配

```go
// cobra/command.go:74-82
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

**函数签名适配**：
- 用户提供的函数使用 `*Command` 类型
- 适配后转换为 `*spf13cobra.Command` 类型
- 执行时再包装回 `*Command` 类型

## Flag 类型适配

### 1. Flag 类型识别

```go
// cobra/command.go:321-400
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
    var items []tui.FlagItem
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

        if current.Parent() == nil {
            break
        }
        current = c.wrapCommand(current.Parent())
    }

    return items
}
```

**类型适配**：
- 将 pflag 的类型映射到 TUI 的类型
- 统一不同来源的 flags
- 追踪 flag 的来源命令

### 2. Flag 值应用

```go
// cobra/command.go:402-431
func (c *Command) applyFlagValues(cmd *Command, values map[string]string) error {
    current := cmd
    for current != nil {
        for name, value := range values {
            // 先尝试查找 LocalFlag
            flag := current.LocalFlags().Lookup(name)
            if flag == nil {
                // 如果没有找到，尝试查找 PersistentFlag
                flag = current.PersistentFlags().Lookup(name)
            }
            if flag != nil {
                if err := flag.Value.Set(value); err != nil {
                    return fmt.Errorf("failed to set flag %s: %w", name, err)
                }
                flag.Changed = true
            }
        }

        if current.Parent() == nil {
            break
        }
        current = c.wrapCommand(current.Parent())
    }

    return nil
}
```

## 数据结构适配

### 1. 命令项转换

```go
// cobra/command_tree.go:10-43
func BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
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
    for _, child := range children {
        childItem := BuildCommandTree(child, currentPath)
        if childItem != nil {
            item.Children = append(item.Children, childItem)
        }
    }

    return item
}
```

**结构转换**：
- `spf13cobra.Command` → `tui.CommandItem`
- 递归处理子命令
- 提取相关属性

### 2. 菜单项适配

```go
// 从 CommandItem 构建 MenuItem
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

// 从原始命令构建菜单项
func buildMenuItemsFromCommands(cmds []*spf13cobra.Command) []MenuItem {
    items := make([]MenuItem, 0, len(cmds))
    for _, cmd := range cmds {
        items = append(items, MenuItem{
            ID:          cmd.Name(),
            Label:       cmd.Name(),
            Description: cmd.Short,
        })
    }
    return items
}
```

## 自定义适配器实现

### 1. 自定义渲染器适配器

```go
// 适配其他 UI 框架
type CustomUIAdapter struct {
    ui *CustomUIFramework
}

func (a *CustomUIAdapter) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    // 将 Cobra-X 的 MenuItem 转换为 CustomUI 的选项
    uiItems := make([]CustomUIOption, len(items))
    for i, item := range items {
        uiItems[i] = CustomUIOption{
            Text: item.Label,
            Desc: item.Description,
        }
    }

    // 调用 CustomUI 框架
    selected, err := a.ui.ShowMenu(title, uiItems)
    if err != nil {
        return -1, err
    }

    return selected, nil
}

func (a *CustomUIAdapter) RenderFlagForm(title string, flags []FlagItem) (map[string]string, error) {
    // 类似的适配逻辑
    return nil, nil
}

func (a *CustomUIAdapter) RenderConfirmation(title, message string) (bool, error) {
    return a.ui.ShowConfirm(title, message), nil
}

func (a *CustomUIAdapter) RenderHelp(title, content string) error {
    return a.ui.ShowHelp(title, content)
}

func (a *CustomUIAdapter) Cleanup() error {
    return a.ui.Cleanup()
}
```

### 2. 使用自定义适配器

```go
func main() {
    // 创建自定义 UI 适配器
    customUI := NewCustomUIFramework()
    adapter := &CustomUIAdapter{ui: customUI}

    // 使用适配器
    rootCmd := cobra.NewCommand("myapp",
        cobra.WithTUIConfig(&cobra.TUIConfig{
            Renderer: adapter,
        }),
    )

    rootCmd.Execute()
}
```

## 多层适配

### 1. 渲染器链

```go
type RendererChain struct {
    adapters []Renderer
}

func (c *RendererChain) AddAdapter(adapter Renderer) {
    c.adapters = append(c.adapters, adapter)
}

func (c *RendererChain) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    // 尝试每个适配器，直到成功
    for _, adapter := range c.adapters {
        index, err := adapter.RenderCommandMenu(title, items)
        if err == nil {
            return index, nil
        }
    }
    return -1, errors.New("all adapters failed")
}
```

### 2. 回退适配器

```go
type FallbackRenderer struct {
    primary   Renderer
    secondary Renderer
}

func (f *FallbackRenderer) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    // 尝试主适配器
    index, err := f.primary.RenderCommandMenu(title, items)
    if err == nil {
        return index, nil
    }

    // 主适配器失败，使用次适配器
    log.Printf("Primary renderer failed: %v, using fallback", err)
    return f.secondary.RenderCommandMenu(title, items)
}
```

## 适配器模式的优势

### 1. 接口统一

```go
// 不同的 UI 框架通过统一接口使用
var renderer Renderer

// 可以是 BubbleTea
renderer = NewDefaultRenderer(theme)

// 可以是自定义框架
renderer = &CustomUIAdapter{ui: customUI}

// 使用方式完全相同
index, err := renderer.RenderCommandMenu("Title", items)
```

### 2. 复用现有代码

```go
// 无需修改现有的 cobra 命令
// 只需添加适配器层
originalCmd := &spf13cobra.Command{Use: "myapp"}
enhancedCmd := cobrax.Enhance(originalCmd)

// 现有代码继续工作
originalCmd.Execute()  // CLI 模式
enhancedCmd.Execute()  // 支持 TUI
```

### 3. 解耦

```go
// TUI 模块不直接依赖 cobra 模块
// 通过接口适配实现解耦
type Renderer interface {
    // 不依赖 cobra.Command
    RenderCommandMenu(title string, options []MenuItem) (int, error)
}
```

### 4. 易于测试

```go
// 测试时使用模拟适配器
type MockRenderer struct {
    selected int
    err      error
}

func (m *MockRenderer) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    return m.selected, m.err
}

func TestCommand(t *testing.T) {
    mock := &MockRenderer{selected: 0}
    cmd := setupCommandWithMockRenderer(mock)
    // 测试...
}
```

## 高级技巧

### 1. 动态适配器选择

```go
type AdaptiveRenderer struct {
    adapters map[string]Renderer
    current  Renderer
}

func (a *AdaptiveRenderer) SelectAdapter(condition string) {
    if adapter, ok := a.adapters[condition]; ok {
        a.current = adapter
    }
}

func (a *AdaptiveRenderer) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    if a.current == nil {
        return -1, errors.New("no adapter selected")
    }
    return a.current.RenderCommandMenu(title, items)
}
```

### 2. 适配器装饰

```go
type LoggingRenderer struct {
    Renderer
    logger *log.Logger
}

func (l *LoggingRenderer) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    l.logger.Printf("Rendering menu: %s", title)
    start := time.Now()

    index, err := l.Renderer.RenderCommandMenu(title, items)

    duration := time.Since(start)
    l.logger.Printf("Menu rendered in %v, selected: %d", duration, index)

    return index, err
}
```

### 3. 类型安全适配

```go
// 使用泛型提供类型安全的适配器
type TypedAdapter[T any] struct {
    adapter func(T) (int, error)
}

func (t *TypedAdapter[T]) Adapt(value T) (int, error) {
    return t.adapter(value)
}

// 使用
stringAdapter := &TypedAdapter[string]{
    adapter: func(s string) (int, error) {
        // 处理字符串
        return 0, nil
    },
}
```

## 最佳实践

### 1. 单一职责

```go
// 每个适配器只负责一个接口转换
type BubbleTeaAdapter struct {
    // 只负责 BubbleTea 适配
}

type CustomUIAdapter struct {
    // 只负责 CustomUI 适配
}
```

### 2. 显式错误处理

```go
func (a *Adapter) Adapt(data interface{}) (Result, error) {
    // 明确检查适配失败
    if !a.canAdapt(data) {
        return Result{}, fmt.Errorf("cannot adapt type %T", data)
    }
    // 执行适配...
}
```

### 3. 文档化适配关系

```go
// Renderer 适配 Cobra-X 的渲染接口到不同的 UI 框架
//
// 支持的适配器：
// - DefaultRenderer: BubbleTea 框架
// - CustomUIAdapter: 自定义 UI 框架
//
// 使用示例：
//   renderer := NewDefaultRenderer(theme)
type Renderer interface {
    // ...
}
```

## 潜在问题和解决方案

### 问题 1: 适配器链过长

**问题**: 多层适配可能导致性能损失

**解决方案**:
```go
// 直接实现，避免多层适配
type DirectRenderer struct {
    // 直接实现，不经过适配层
}

func (d *DirectRenderer) RenderCommandMenu(title string, items []MenuItem) (int, error) {
    // 直接实现逻辑
}
```

### 问题 2: 类型信息丢失

**问题**: 适配过程中可能丢失类型信息

**解决方案**:
```go
// 保留原始类型
type PreservedAdapter struct {
    original interface{}
}

func (p *PreservedAdapter) Adapt(data interface{}) interface{} {
    p.original = data
    // 适配逻辑...
}

func (p *PreservedAdapter) GetOriginal() interface{} {
    return p.original
}
```

### 问题 3: 适配器版本兼容

**问题**: 被适配的接口可能变化

**解决方案**:
```go
// 使用适配器工厂
type AdapterFactory interface {
    CreateAdapter(version string) (Adapter, error)
}

type VersionedAdapterFactory struct{}

func (v *VersionedAdapterFactory) CreateAdapter(version string) (Adapter, error) {
    switch version {
    case "v1":
        return &V1Adapter{}, nil
    case "v2":
        return &V2Adapter{}, nil
    default:
        return nil, fmt.Errorf("unsupported version: %s", version)
    }
}
```

## 总结

适配器模式在 Cobra-X 中扮演着关键角色：

1. **接口统一**: 将不同的 UI 框架适配到统一接口
2. **类型转换**: 在不同命令类型之间转换
3. **数据适配**: 转换数据结构以适应不同需求
4. **向后兼容**: 保持与原版 cobra 的兼容性
5. **扩展性**: 支持自定义适配器

适配器模式使得 Cobra-X 能够无缝集成多种技术和框架，同时保持代码的清晰和可维护性。这种设计让用户可以选择最适合自己需求的渲染方式，也使得框架能够适应未来的变化。适配器模式是实现系统集成和代码复用的优秀实践。
