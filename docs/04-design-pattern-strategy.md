# 策略模式 (Strategy Pattern)

## 概述

策略模式是一种行为型设计模式，它定义了一系列算法，将每个算法封装起来，并使它们可以互换。策略模式让算法独立于使用它的客户端而变化。在 Cobra-X 中，策略模式主要用于实现可插拔的渲染器和主题系统。

## 在 Cobra-X 中的应用

### 核心实现位置

1. **渲染器接口**: [tui/renderer.go](../tui/renderer.go) - `Renderer` 接口
2. **默认实现**: [tui/default_renderer.go](../tui/default_renderer.go) - `DefaultRenderer`
3. **主题系统**: [tui/style/theme.go](../tui/style/theme.go) - `Theme` 结构
4. **交互模式**: [cobra/tui_config.go](../cobra/tui_config.go) - `InteractiveMode` 枚举

## 代码实现分析

### 1. 渲染器策略接口

```go
// tui/renderer.go:7-33
type Renderer interface {
    // RenderCommandMenu 渲染命令菜单面板
    RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error)

    // RenderFlagForm 渲染 flag 输入表单
    RenderFlagForm(formTitle string, flags []FlagItem) (values map[string]string, err error)

    // RenderConfirmation 渲染确认面板
    RenderConfirmation(title, message string) (confirmed bool, err error)

    // RenderHelp 渲染帮助面板
    RenderHelp(title, content string) error

    // Cleanup 清理资源
    Cleanup() error
}
```

**策略模式要素**：
- **策略接口**: `Renderer` 定义了渲染行为的抽象
- **具体策略**: `DefaultRenderer` 是默认实现
- **上下文**: `Command` 使用渲染器执行渲染操作

### 2. 默认策略实现

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

### 3. 策略使用 - 命令执行

```go
// cobra/command.go:205-211
func (c *Command) executeTUI() error {
    // 获取渲染器（策略）
    renderer := c.getRenderer()
    defer renderer.Cleanup()

    // 使用渲染器执行操作
    selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})
    // ...
}
```

### 4. 策略选择逻辑

```go
// cobra/command.go:520-528
func (c *Command) getRenderer() tui.Renderer {
    if c.tuiConfig != nil && c.tuiConfig.Renderer != nil {
        // 使用自定义渲染器策略
        return c.tuiConfig.Renderer
    }

    // 使用默认渲染器策略
    return tui.NewDefaultRenderer(c.getTheme())
}
```

**策略切换机制**：
- 优先使用用户配置的自定义渲染器
- 回退到默认渲染器
- 支持运行时切换

## 主题策略

### 1. 主题作为样式策略

```go
// tui/style/theme.go:7-21
type Theme struct {
    Name   string
    Colors ColorScheme
    Layout LayoutConfig
    Styles StyleConfig
}
```

主题是一种策略，定义了：
- **颜色策略**: 如何着色界面元素
- **布局策略**: 如何排列界面元素
- **样式策略**: 如何装饰界面元素

### 2. 主题工厂（策略创建）

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

### 3. 主题选择策略

```go
// cobra/command.go:505-518
func (c *Command) getTheme() *style.Theme {
    // 1. 优先使用配置的主题
    if c.tuiConfig != nil && c.tuiConfig.Theme != nil {
        return c.tuiConfig.Theme
    }

    // 2. 从 flags 读取主题名称
    themeName := "default"
    if theme, err := c.Flags().GetString("tui-theme"); err == nil {
        themeName = theme
    }

    // 3. 返回对应主题
    return style.NewTheme(themeName)
}
```

**优先级策略**：
1. 显式配置的主题
2. 命令行参数指定的主题
3. 默认主题

## 交互模式策略

### 1. 模式枚举

```go
// cobra/tui_config.go:40-55
type InteractiveMode int

const (
    // ModeAuto 自动模式 - 根据终端能力自动选择
    ModeAuto InteractiveMode = iota

    // ModeTUI 强制使用 TUI 模式
    ModeTUI

    // ModeCLI 强制使用 CLI 模式
    ModeCLI
)
```

### 2. 模式选择策略

```go
// cobra/command.go:149-181
func (c *Command) shouldUseTUI() bool {
    // 1. 检查强制 CLI 模式
    if c.tuiConfig != nil && c.tuiConfig.InteractiveMode == ModeCLI {
        return false
    }

    // 2. 检查强制 TUI 模式
    if c.tuiConfig != nil && c.tuiConfig.InteractiveMode == ModeTUI {
        return true
    }

    // 3. 检查 --tui flag
    if tuiFlag, err := c.Flags().GetBool("tui"); err == nil && tuiFlag {
        return true
    }

    // 4. 检查环境变量
    if os.Getenv("COBRA_TUI") == "true" {
        return true
    }

    // 5. 检查配置和终端能力
    if c.tuiConfig != nil && c.tuiConfig.Enabled {
        if c.tuiConfig.AutoDetect {
            return c.isInteractiveTerminal()
        }
        return true
    }

    return false
}
```

**决策策略**：
1. 最先检查强制模式（配置优先）
2. 然后检查命令行参数（用户输入）
3. 最后检查环境变量和终端能力

## 自定义策略实现

### 1. 自定义渲染器

```go
// 实现自定义渲染策略
type CustomRenderer struct {
    // 自定义字段
    useColor bool
    language string
}

func (r *CustomRenderer) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    // 自定义实现 - 例如使用不同的 UI 库
    // 或者支持国际化
    return 0, nil
}

func (r *CustomRenderer) RenderFlagForm(title string, flags []tui.FlagItem) (map[string]string, error) {
    // 自定义表单实现
    return nil, nil
}

func (r *CustomRenderer) RenderConfirmation(title, message string) (bool, error) {
    // 自定义确认实现
    return true, nil
}

func (r *CustomRenderer) RenderHelp(title, content string) error {
    // 自定义帮助实现
    return nil
}

func (r *CustomRenderer) Cleanup() error {
    // 清理资源
    return nil
}
```

### 2. 使用自定义策略

```go
// 方式 1: 通过配置
rootCmd := cobra.NewCommand("myapp",
    cobra.WithTUIConfig(&cobra.TUIConfig{
        Renderer: &CustomRenderer{
            useColor: true,
            language: "zh-CN",
        },
    }),
)

// 方式 2: 通过选项
rootCmd.SetTUIRenderer(&CustomRenderer{
    useColor: true,
    language: "en-US",
})
```

### 3. 自定义主题

```go
// 创建自定义主题策略
myTheme := &style.Theme{
    Name: "custom",
    Colors: style.ColorScheme{
        Primary:   lipgloss.Color("#FF6B6B"),
        Secondary: lipgloss.Color("#4ECDC4"),
        Success:   lipgloss.Color("#95E1D3"),
        Warning:   lipgloss.Color("#FFEAA7"),
        Error:     lipgloss.Color("#FF6B6B"),
        Muted:     lipgloss.Color("#DFE6E9"),
        Background: lipgloss.Color("#2D3436"),
        Foreground: lipgloss.Color("#DFE6E9"),
    },
    Layout: style.defaultLayout(),
    Styles: style.defaultStyles(),
}

// 使用自定义主题
rootCmd := cobra.NewCommand("myapp",
    cobra.WithTUIConfig(&cobra.TUIConfig{
        Theme: myTheme,
    }),
)
```

## 策略模式的优势

### 1. 算法族封装

将不同的渲染算法封装成独立的策略：

```go
// 不同渲染策略
var strategies = map[string]tui.Renderer{
    "default": tui.NewDefaultRenderer(style.DefaultTheme()),
    "simple":  &SimpleRenderer{},
    "advanced": &AdvancedRenderer{},
    "web":     &WebRenderer{},
}

// 运行时选择
renderer := strategies[selectedStrategy]
```

### 2. 运行时切换

```go
// 根据用户选择动态切换
switch userPreference {
case "dark":
    cmd.SetTUIConfig(&cobra.TUIConfig{
        Theme: style.NewTheme("dark"),
    })
case "light":
    cmd.SetTUIConfig(&cobra.TUIConfig{
        Theme: style.NewTheme("light"),
    })
}
```

### 3. 独立测试

每个策略可以独立测试：

```go
func TestCustomRenderer(t *testing.T) {
    renderer := &CustomRenderer{}
    items := []tui.MenuItem{
        {ID: "1", Label: "Item 1"},
    }

    index, err := renderer.RenderCommandMenu("Test", items)
    assert.NoError(t, err)
    assert.Equal(t, 0, index)
}
```

### 4. 扩展性

添加新策略无需修改现有代码：

```go
// 新增渲染器
type WebRenderer struct{}

func (r *WebRenderer) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    // 实现 Web UI 渲染
}

// 无需修改现有代码，直接使用
renderer := &WebRenderer{}
```

## 策略模式的实现技巧

### 1. 策略注册表

```go
type RendererRegistry struct {
    renderers map[string]tui.Renderer
}

func NewRendererRegistry() *RendererRegistry {
    return &RendererRegistry{
        renderers: make(map[string]tui.Renderer),
    }
}

func (r *RendererRegistry) Register(name string, renderer tui.Renderer) {
    r.renderers[name] = renderer
}

func (r *RendererRegistry) Get(name string) (tui.Renderer, bool) {
    renderer, ok := r.renderers[name]
    return renderer, ok
}

// 使用
registry := NewRendererRegistry()
registry.Register("default", tui.NewDefaultRenderer(nil))
registry.Register("custom", &CustomRenderer{})
```

### 2. 策略链

```go
// 支持多个策略组合
type RendererChain struct {
    renderers []tui.Renderer
}

func (c *RendererChain) AddRenderer(renderer tui.Renderer) {
    c.renderers = append(c.renderers, renderer)
}

func (c *RendererChain) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    // 尝试每个渲染器，直到成功
    for _, renderer := range c.renderers {
        index, err := renderer.RenderCommandMenu(title, items)
        if err == nil {
            return index, nil
        }
    }
    return -1, errors.New("all renderers failed")
}
```

### 3. 策略缓存

```go
type CachedRenderer struct {
    renderer  tui.Renderer
    cache     map[string]interface{}
    cacheLock sync.RWMutex
}

func (c *CachedRenderer) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    // 检查缓存
    cacheKey := fmt.Sprintf("%s-%v", title, items)
    c.cacheLock.RLock()
    if cached, ok := c.cache[cacheKey]; ok {
        c.cacheLock.RUnlock()
        return cached.(int), nil
    }
    c.cacheLock.RUnlock()

    // 调用实际渲染器
    index, err := c.renderer.RenderCommandMenu(title, items)
    if err == nil {
        c.cacheLock.Lock()
        c.cache[cacheKey] = index
        c.cacheLock.Unlock()
    }

    return index, err
}
```

## 与其他模式的配合

### 1. 与工厂模式配合

```go
// 工厂创建策略
func NewRenderer(strategy string) tui.Renderer {
    switch strategy {
    case "default":
        return tui.NewDefaultRenderer(nil)
    case "custom":
        return &CustomRenderer{}
    default:
        return tui.NewDefaultRenderer(nil)
    }
}
```

### 2. 与装饰器模式配合

```go
// 装饰器包装策略
type LoggingRenderer struct {
    renderer tui.Renderer
    logger   *log.Logger
}

func (l *LoggingRenderer) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    l.logger.Printf("Rendering menu: %s", title)
    index, err := l.renderer.RenderCommandMenu(title, items)
    l.logger.Printf("Menu rendered, selected: %d", index)
    return index, err
}
```

### 3. 与单例模式配合

```go
// 主题通常使用单例
var (
    defaultThemeOnce sync.Once
    defaultTheme     *style.Theme
)

func DefaultTheme() *style.Theme {
    defaultThemeOnce.Do(func() {
        defaultTheme = defaultTheme()
    })
    return defaultTheme
}
```

## 最佳实践

### 1. 策略接口设计

```go
// 好的设计 - 接口精简
type Renderer interface {
    RenderCommandMenu(title string, items []MenuItem) (int, error)
    RenderFlagForm(title string, flags []FlagItem) (map[string]string, error)
    // ...
}

// 避免 - 接口过于庞大
type Renderer interface {
    // 50+ 个方法...
}
```

### 2. 提供默认实现

```go
// 总是提供合理的默认实现
func (c *Command) getRenderer() tui.Renderer {
    if c.tuiConfig != nil && c.tuiConfig.Renderer != nil {
        return c.tuiConfig.Renderer
    }
    return tui.NewDefaultRenderer(c.getTheme())
}
```

### 3. 策略验证

```go
// 验证策略可用性
func ValidateRenderer(renderer tui.Renderer) error {
    if renderer == nil {
        return errors.New("renderer cannot be nil")
    }
    // 其他验证...
    return nil
}
```

## 潜在问题和解决方案

### 问题 1: 策略选择复杂

**问题**: 客户端需要知道选择哪个策略

**解决方案**:
```go
// 提供智能选择
func SelectRenderer(config Config) tui.Renderer {
    if config.UseWeb {
        return &WebRenderer{}
    }
    if config.HighContrast {
        return &HighContrastRenderer{}
    }
    return tui.NewDefaultRenderer(nil)
}
```

### 问题 2: 策略间通信

**问题**: 策略之间需要共享状态

**解决方案**:
```go
// 使用上下文对象
type RenderContext struct {
    Width   int
    Height  int
    Theme   *style.Theme
    Shared  map[string]interface{}
}
```

### 问题 3: 策略切换开销

**问题**: 频繁切换策略可能有性能开销

**解决方案**:
```go
// 策略池
type RendererPool struct {
    pool map[string]tui.Renderer
    sync.RWMutex
}

func (p *RendererPool) Get(name string) tui.Renderer {
    p.RLock()
    if renderer, ok := p.pool[name]; ok {
        p.RUnlock()
        return renderer
    }
    p.RUnlock()

    p.Lock()
    defer p.Unlock()
    // 创建并缓存...
}
```

## 总结

策略模式在 Cobra-X 中实现了：

1. **渲染器策略**: 支持不同的 UI 渲染方式
2. **主题策略**: 支持多种视觉样式
3. **交互模式策略**: 支持不同的交互方式
4. **可扩展性**: 用户可以实现自定义策略

策略模式使得 Cobra-X 具有极强的灵活性和可扩展性，用户可以根据需要选择或实现适合自己的渲染策略，而无需修改核心代码。这种设计体现了"对扩展开放，对修改关闭"的开闭原则。
