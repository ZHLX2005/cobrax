# 装饰器模式 (Decorator Pattern)

## 概述

装饰器模式是一种结构型设计模式，它允许在不修改原始对象代码的情况下，动态地为对象添加新功能。在 Cobra-X 项目中，装饰器模式是实现零侵入 TUI 功能增强的核心设计。

## 在 Cobra-X 中的应用

### 核心实现位置

- **文件**: [cobra/decorate.go](../cobra/decorate.go)
- **核心函数**: `Enhance()`

### 设计目的

Cobra-X 的核心目标是保持与原版 `spf13/cobra` 的 100% API 兼容性。通过装饰器模式，用户可以为现有的 cobra 命令添加 TUI 功能，而无需修改任何现有代码。

## 代码实现分析

### 1. Enhance 函数 - 装饰器入口

```go
// cobra/decorate.go:32-67
func Enhance(cmd *spf13cobra.Command, opts ...EnhanceOption) *spf13cobra.Command {
    if cmd == nil {
        return nil
    }

    // 创建增强配置
    config := &EnhanceConfig{
        TUIConfig: DefaultTUIConfig(),
    }

    // 应用选项
    for _, opt := range opts {
        opt(config)
    }

    // 存储 TUI 配置到命令的 Annotations 中
    // 这样不会影响原有的命令结构
    if cmd.Annotations == nil {
        cmd.Annotations = make(map[string]string)
    }
    cmd.Annotations["tui.enabled"] = "true"

    // 添加 TUI flags
    addTUIFlags(cmd)

    // 包装 PreRun/E 以拦截执行
    wrapExecute(cmd, config)

    return cmd
}
```

**设计亮点**:
- 使用 `Annotations` 存储配置，不破坏原有结构
- 通过选项函数支持灵活配置
- 返回相同类型，保持 API 兼容性

### 2. 执行逻辑包装

```go
// cobra/decorate.go:86-169
func wrapExecute(cmd *spf13cobra.Command, config *EnhanceConfig) {
    // 保存原有的执行函数
    originalPersistentPreRunE := cmd.PersistentPreRunE
    originalPreRunE := cmd.PreRunE
    originalRunE := cmd.RunE
    originalRun := cmd.Run

    // 包装 RunE
    if originalRunE != nil {
        cmd.RunE = func(c *spf13cobra.Command, args []string) error {
            // 检查是否需要启动 TUI（只执行一次）
            if shouldUseTUIForCommand(c, config) && c.Annotations["tui.executed"] == "false" {
                c.Annotations["tui.executed"] = "true"
                return executeTUIForCommand(c, config)
            }
            return originalRunE(c, args)
        }
    }

    // 类似地包装 Run 和其他生命周期函数...
}
```

**设计要点**:
- 保存原有函数引用
- 在新函数中添加 TUI 逻辑
- 根据条件决定执行路径
- 确保只执行一次（通过 `tui.executed` 标记）

### 3. 装饰器选项模式

```go
// cobra/decorate.go:333-360
type EnhanceOption func(*EnhanceConfig)

func WithEnhanceTUIEnabled(enabled bool) EnhanceOption {
    return func(c *EnhanceConfig) {
        if c.TUIConfig == nil {
            c.TUIConfig = DefaultTUIConfig()
        }
        c.TUIConfig.Enabled = enabled
    }
}

func WithEnhanceTheme(themeName string) EnhanceOption {
    return func(c *EnhanceConfig) {
        if c.TUIConfig == nil {
            c.TUIConfig = DefaultTUIConfig()
        }
        if c.TUIConfig.Theme == nil {
            c.TUIConfig.Theme = style.NewTheme(themeName)
        }
    }
}
```

## 使用示例

### 基础用法

```go
import (
    "github.com/spf13/cobra"
    "github.com/ZHLX2005/cobrax"
)

// 原始命令定义
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Running myapp")
    },
}

// 使用装饰器增强
rootCmd = cobrax.Enhance(rootCmd,
    cobrax.WithEnhanceTUIEnabled(true),
    cobrax.WithEnhanceTheme("dracula"),
)

func main() {
    rootCmd.Execute()
}
```

### 在现有项目中应用

```go
// 无需修改现有的命令定义
// 只需在 main 函数中添加一行即可

func main() {
    // 原有代码
    // rootCmd.Execute()

    // 添加装饰器
    rootCmd = cobrax.Enhance(rootCmd)
    rootCmd.Execute()
}
```

## 设计优势

### 1. 零侵入性

- 不需要修改现有代码
- 保持原有的命令结构不变
- 可以渐进式采用

### 2. 完全兼容

- 返回相同类型 (`*spf13cobra.Command`)
- 所有原有功能保持不变
- 可以与现有工具链无缝集成

### 3. 灵活配置

- 通过选项函数支持各种配置
- 可以选择性启用/禁用功能
- 支持运行时动态切换

### 4. 职责分离

- 装饰器只负责添加功能
- 原有命令保持独立
- 易于维护和扩展

## 与其他模式的配合

### 与策略模式配合

装饰器模式将 TUI 功能添加到命令，而策略模式允许不同的渲染策略：

```go
// 自定义渲染器
type MyCustomRenderer struct{}

func (r *MyCustomRenderer) RenderCommandMenu(title string, items []tui.MenuItem) (int, error) {
    // 自定义实现
}

// 使用装饰器配置自定义策略
rootCmd = cobrax.Enhance(rootCmd,
    cobrax.WithEnhanceTUIConfig(&cobra.TUIConfig{
        Renderer: &MyCustomRenderer{},
    }),
)
```

### 与建造者模式配合

选项函数本身就是建造者模式的一种变体：

```go
// 链式配置
rootCmd = cobrax.Enhance(rootCmd,
    cobrax.WithEnhanceTUIEnabled(true),
    cobrax.WithEnhanceTheme("nord"),
    cobrax.WithEnhanceTUIConfirm(true),
)
```

## 实现细节要点

### 1. 避免重复执行

```go
// 使用 annotation 标记避免重复执行
if c.Annotations["tui.executed"] == "false" {
    c.Annotations["tui.executed"] = "true"
    return executeTUIForCommand(c, config)
}
```

### 2. 生命周期钩子

Cobra-X 在多个生命周期点插入 TUI 逻辑：

- `HelpFunc`: 有子命令的命令触发 TUI
- `PersistentPreRunE`: 父命令的预处理
- `PreRunE`: 当前命令的预处理
- `Run/RunE`: 实际执行逻辑

### 3. 错误处理

```go
// TUI 模式出错时的优雅降级
if err := executeTUIForCommand(c, config); err != nil {
    printError(err)
    return
}
```

## 潜在优化方向

### 1. 性能优化

- 缓存装饰器配置
- 减少反射操作

### 2. 功能扩展

- 支持多层装饰
- 支持装饰器链
- 支持装饰器撤销

### 3. 类型安全

- 使用泛型增强类型安全
- 编译时检查装饰器兼容性

## 总结

装饰器模式是 Cobra-X 项目的核心设计模式，它使得在不修改现有代码的情况下添加 TUI 功能成为可能。这种设计的最大价值在于：

1. **零迁移成本**: 用户只需一行代码即可启用 TUI
2. **向后兼容**: 不影响现有功能和 API
3. **渐进增强**: 可以逐步采用，无需全盘重构

这种设计展示了如何在保持兼容性的同时为成熟框架添加新功能，是软件工程中的优秀实践案例。
