# Cobra-X 软件设计规划

## 项目概述
cobra-x 是对 [spf13/cobra](https://github.com/spf13/cobra) 的增强版本，保持 API 完全兼容的同时，增加了 TUI（终端用户界面）面板交互功能。

## 核心目标
1. **100% API 兼容** - 所有现有 cobra 代码无需修改即可使用
2. **TUI 自动生成** - 根据命令树自动生成交互式面板
3. **可扩展性** - 支持自定义面板渲染和交互逻辑
4. **零配置默认体验** - 开箱即用的默认面板样式

---

## 项目结构

```
github.com/ZHLX2005/cobra/
├── cobra/                    # 核心 cobra 兼容层
│   ├── command.go           # Command 结构（继承自原始 cobra）
│   ├── flag.go              # Flag 处理
│   ├── args.go              # 参数解析
│   └── ...
├── tui/                     # TUI 面板系统
│   ├── renderer.go          # 面板渲染器接口
│   ├── default_renderer.go  # 默认渲染器实现
│   ├── interaction.go       # 交互逻辑处理
│   ├── widgets/             # UI 组件
│   │   ├── menu.go          # 菜单组件
│   │   ├── form.go          # 表单组件（flags 输入）
│   │   ├── tree.go          # 命令树组件
│   │   └── help.go          # 帮助组件
│   └── style/               # 样式系统
│       ├── theme.go         # 主题定义
│       └── colors.go        # 颜色配置
├── ext/                     # 扩展功能
│   ├── tui_command.go       # TUI 命令扩展
│   └── panel_builder.go     # 面板构建器
├── examples/                # 示例代码
│   └── basic/
│       └── main.go
├── test/                    # 测试
│   └── ...
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

## 核心设计

### 1. API 兼容性设计

#### 1.1 类型别名与嵌入
```go
package cobra

// Command 保持与原始 cobra.Command 完全兼容
type Command struct {
    *spf13cobra.Command  // 嵌入原始 Command
    tuiConfig            *TUIConfig  // TUI 配置（扩展字段）
    panelBuilder         PanelBuilder // 自定义面板构建器
}

// 所有原始方法自动可用
// 新增方法不影响原有功能
```

#### 1.2 扩展配置结构
```go
// TUIConfig TUI 面板配置
type TUIConfig struct {
    Enabled          bool              // 是否启用 TUI
    Renderer         Renderer          // 自定义渲染器
    Theme            *Theme            // 主题配置
    ShowDescription  bool              // 是否显示命令描述
    ShowFlags        bool              // 是否显示 flags 面板
    InteractiveMode  InteractiveMode   // 交互模式
}

// InteractiveMode 交互模式枚举
type InteractiveMode int

const (
    ModeAuto      InteractiveMode = iota // 自动模式
    ModeTUI                              // 强制 TUI
    ModeCLI                              // 强制 CLI
)
```

### 2. TUI 面板系统架构

#### 2.1 渲染器接口
```go
package tui

// Renderer 面板渲染器接口
type Renderer interface {
    // RenderCommandMenu 渲染命令菜单面板
    RenderCommandMenu(cmd *cobra.Command, options []MenuItem) (selected *cobra.Command, err error)

    // RenderFlagForm 渲染 flag 输入表单
    RenderFlagForm(cmd *cobra.Command, flags []FlagItem) (values map[string]string, err error)

    // RenderConfirmation 渲染确认面板
    RenderConfirmation(title, message string) (confirmed bool, err error)

    // RenderHelp 渲染帮助面板
    RenderHelp(cmd *cobra.Command) error
}

// DefaultRenderer 默认渲染器实现
type DefaultRenderer struct {
    theme *style.Theme
    // ...
}
```

#### 2.2 面板构建器
```go
package ext

// PanelBuilder 面板构建器接口
type PanelBuilder interface {
    // BuildMenu 构建菜单面板
    BuildMenu(cmd *cobra.Command) *MenuPanel

    // BuildFlagForm 构建表单面板
    BuildFlagForm(cmd *cobra.Command) *FormPanel

    // Validate 验证构建的面板
    Validate() error
}

// DefaultPanelBuilder 默认构建器
type DefaultPanelBuilder struct {
    // ...
}
```

### 3. 交互流程设计

#### 3.1 TUI 启动流程
```
┌─────────────────────────────────────────────────────────────┐
│                         用户执行命令                          │
│                    ./myapp --tui 或 ./myapp                  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    检测交互模式                              │
│  • --tui flag 存在？                                         │
│  • 环境变量 COBRA_TUI=true？                                 │
│  • 命令配置了 TUI 启用？                                     │
└─────────────────────────────────────────────────────────────┘
                              │
                    ┌─────────┴─────────┐
                    │                   │
                    ▼                   ▼
              [启用 TUI]            [传统 CLI]
                    │                   │
                    ▼                   │
┌─────────────────────────────────────────────────────────────┐
│                   命令树菜单面板                             │
│  • 显示当前命令的所有子命令                                  │
│  • 显示命令描述                                              │
│  • 支持键盘导航和搜索                                        │
└─────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   选择子命令 / 执行                          │
│  • 如果是叶子命令：进入 flag 配置面板                        │
│  • 如果有子命令：递归显示子菜单                              │
└─────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   Flag 配置面板                              │
│  • 显示所有 flags                                            │
│  • 支持编辑、选择、使用默认值                                │
│  • 实时验证                                                  │
└─────────────────────────────────────────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────┐
│                   确认与执行                                 │
│  • 显示完整命令预览                                          │
│  • 确认后执行                                                │
└─────────────────────────────────────────────────────────────┘
```

#### 3.2 默认面板类型

**1. 命令菜单面板 (CommandMenu)**
```
┌──────────────────────────────────────────────┐
│              MyApp v1.0.0                    │
├──────────────────────────────────────────────┤
│  Select a command:                           │
│                                              │
│  ▶ server     Start the server              │
│    client     Start the client              │
│    config     Manage configuration          │
│    help       Show help                     │
│                                              │
│                                              │
│  [↑↓ Navigate] [Enter Select] [Esc Quit]    │
└──────────────────────────────────────────────┘
```

**2. Flag 表单面板 (FlagForm)**
```
┌──────────────────────────────────────────────┐
│         Configure: server start              │
├──────────────────────────────────────────────┤
│                                              │
│  Port:                      [8080    ]       │
│  Host:                      [0.0.0.0 ]       │
│  TLS:                       [✓]              │
│  Worker count:              [4       ]       │
│  Config file:               [/path...]       │
│                                              │
│                                              │
│  [Tab Next] [Enter Save] [Esc Cancel]       │
└──────────────────────────────────────────────┘
```

**3. 确认面板 (Confirmation)**
```
┌──────────────────────────────────────────────┐
│              Confirm Execution               │
├──────────────────────────────────────────────┤
│                                              │
│  Command to execute:                         │
│  myapp server start --port=8080 --tls        │
│                                              │
│  Proceed?                                    │
│                                              │
│  [Yes]  [No]                                 │
└──────────────────────────────────────────────┘
```

### 4. 样式系统

#### 4.1 主题配置
```go
package style

type Theme struct {
    // 颜色配置
    Colors ColorScheme

    // 布局配置
    Layout LayoutConfig

    // 字符样式
    Styles StyleConfig
}

type ColorScheme struct {
    Primary   string
    Secondary string
    Success   string
    Warning   string
    Error     string
    Muted     string
    // ...
}
```

#### 4.2 预设主题
```go
const (
    ThemeDefault  Theme = "default"  // 默认蓝色主题
    ThemeDark     Theme = "dark"     // 暗色主题
    ThemeLight    Theme = "light"    // 亮色主题
    ThemeMinimal  Theme = "minimal"  // 极简主题
)
```

### 5. 扩展点设计

#### 5.1 自定义渲染器
```go
// 用户可以实现自己的渲染器
type MyCustomRenderer struct {
    // 自定义字段
}

func (r *MyCustomRenderer) RenderCommandMenu(cmd *cobra.Command, options []MenuItem) (*cobra.Command, error) {
    // 自定义渲染逻辑
    // 使用 bubbletea, lipgloss, 或其他 TUI 库
}

// 使用自定义渲染器
cmd.SetTUIConfig(&cobra.TUIConfig{
    Renderer: &MyCustomRenderer{},
})
```

#### 5.2 自定义面板组件
```go
// PanelComponent 面板组件接口
type PanelComponent interface {
    Render(screen *Screen) error
    Handle(event Event) error
}

// 用户可以自定义组件
type MyCustomComponent struct {
    // ...
}

func (c *MyCustomComponent) Render(screen *Screen) error {
    // 自定义渲染
}
```

### 6. 命令扩展

#### 6.1 TUI 相关 Flag
```go
// 自动添加到所有命令
var TUIFlags = []Flag{
    &Flag{
        Name:  "tui",
        Short: "t",
        Usage: "Launch TUI interface",
    },
    &Flag{
        Name:  "tui-theme",
        Usage: "TUI theme (default, dark, light, minimal)",
    },
}
```

#### 6.2 TUI 模式检测
```go
func (c *Command) shouldUseTUI() bool {
    // 1. 检查 --tui flag
    if c.Flags().Changed("tui") {
        return true
    }
    // 2. 检查环境变量
    if os.Getenv("COBRA_TUI") == "true" {
        return true
    }
    // 3. 检查配置
    if c.tuiConfig != nil && c.tuiConfig.Enabled {
        return true
    }
    return false
}
```

---

## 兼容性保证

### 1. 导入路径兼容
```go
// 用户代码可以继续使用
import "github.com/spf13/cobra"

// 或者切换到增强版
import "github.com/ZHLX2005/cobra"

// API 完全相同
```

### 2. 渐进式增强
```go
// 现有代码无需修改即可运行
cmd := &cobra.Command{
    Use:   "server",
    Short: "Start the server",
    Run: func(cmd *cobra.Command, args []string) {
        // 原有逻辑
    },
}

// 可选：添加 TUI 支持
cmd.EnableTUI()  // 新增方法，不影响原有功能
```

### 3. 回退机制
```bash
# 如果 TUI 初始化失败，自动回退到传统 CLI 模式
# 错误信息会被记录但不阻止程序执行
```

---

## 实现步骤

### Phase 1: 基础结构 (Step 1-3)
1. 初始化项目结构和 go.mod
2. 创建核心类型定义
3. 实现 API 兼容层

### Phase 2: TUI 核心 (Step 4-7)
4. 实现渲染器接口
5. 创建默认渲染器
6. 实现交互逻辑
7. 添加样式系统

### Phase 3: 面板组件 (Step 8-10)
8. 实现命令菜单组件
9. 实现 flag 表单组件
10. 实现确认和帮助组件

### Phase 4: 集成与扩展 (Step 11-13)
11. 集成 TUI 到命令执行流程
12. 实现自定义面板接口
13. 添加主题系统

### Phase 5: 完善与测试 (Step 14-15)
14. 编写示例代码
15. 添加测试和文档

---

## 依赖库选择

### TUI 框架
- **推荐**: [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
  - 纯 Go 实现，性能优秀
  - Elm 架构，易于理解和扩展
  - 活跃的社区支持

- **样式**: [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)
  - 强大的终端样式 API
  - 支持颜色、布局、边框等

- **可选替代**:
  - [rivo/tview](https://github.com/rivo/tview) - 功能更丰富
  - [gdamore/tcell](https://github.com/gdamore/tcell) - 底层终端控制

```go
// go.mod 依赖
require (
    github.com/spf13/cobra v1.8.1
    github.com/charmbracelet/bubbletea v0.26.6
    github.com/charmbracelet/lipgloss v0.10.0
    github.com/inconshreveable/mousetrap v1.1.0 // 继承自 cobra
    github.com/spf13/pflag v1.0.5              // 继承自 cobra
)
```

---

## 使用示例

### 基础用法
```go
package main

import (
    "github.com/ZHLX2005/cobra"
)

func main() {
    rootCmd := &cobra.Command{
        Use:   "myapp",
        Short: "My application",
    }

    serverCmd := &cobra.Command{
        Use:   "server",
        Short: "Start the server",
        Run: func(cmd *cobra.Command, args []string) {
            // 业务逻辑
        },
    }

    // 添加 flags
    serverCmd.Flags().String("port", "8080", "Server port")
    serverCmd.Flags().Bool("tls", false, "Enable TLS")

    rootCmd.AddCommand(serverCmd)

    // 启用 TUI（可选）
    rootCmd.EnableTUI()

    rootCmd.Execute()
}
```

### 启动方式
```bash
# 自动检测：如果是交互式终端且无参数，显示 TUI
./myapp

# 强制使用 TUI
./myapp --tui

# 传统 CLI 模式
./myapp server --port=9090

# 设置环境变量
export COBRA_TUI=true
./myapp
```

---

## 关键技术决策

### 1. 为什么嵌入原始 cobra.Command？
- 最大化兼容性
- 零代码迁移成本
- 继承所有原始功能

### 2. 为什么使用接口设计？
- 支持用户自定义渲染
- 便于测试和 mock
- 解耦核心逻辑和 UI 实现

### 3. 为什么提供默认渲染器？
- 开箱即用的体验
- 作为自定义渲染的参考实现
- �盖 80% 的使用场景

---

## 未来扩展

1. **多语言支持** - 国际化命令描述
2. **配置持久化** - 保存常用的命令配置
3. **命令历史** - TUI 内的命令历史记录
4. **插件系统** - 支持第三方面板插件
5. **远程命令** - 通过 TUI 执行远程命令

---

## 总结

cobra-x 的核心设计理念是：
- **兼容优先** - 100% 保持原始 API
- **渐进增强** - TUI 作为可选功能
- **可扩展** - 清晰的扩展点
- **开箱即用** - 默认实现覆盖大部分场景
