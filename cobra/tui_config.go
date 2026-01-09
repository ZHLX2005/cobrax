package cobra

import (
	"github.com/ZHLX2005/cobra/tui"
	"github.com/ZHLX2005/cobra/tui/style"
)

// TUIConfig TUI 面板配置
// 用于配置 TUI 界面的行为和外观
type TUIConfig struct {
	// Enabled 是否启用 TUI 模式
	// 可以通过 --tui flag 或环境变量 COBRA_TUI=true 启用
	Enabled bool

	// Renderer 自定义渲染器
	// 如果为 nil，使用默认渲染器
	Renderer tui.Renderer

	// Theme 主题配置
	// 如果为 nil，使用默认主题
	Theme *style.Theme

	// ShowDescription 是否在菜单中显示命令描述
	ShowDescription bool

	// ShowFlags 是否在执行前显示 flag 配置面板
	ShowFlags bool

	// InteractiveMode 交互模式
	InteractiveMode InteractiveMode

	// AutoDetect 是否自动检测终端能力
	// 如果启用，在非交互式终端或终端不支持时会回退到 CLI 模式
	AutoDetect bool

	// ConfirmBeforeExecute 执行命令前是否显示确认面板
	ConfirmBeforeExecute bool
}

// InteractiveMode 交互模式枚举
type InteractiveMode int

const (
	// ModeAuto 自动模式
	// 根据终端能力和参数自动选择 TUI 或 CLI 模式
	ModeAuto InteractiveMode = iota

	// ModeTUI 强制使用 TUI 模式
	// 即使终端不支持也会尝试启动（可能失败）
	ModeTUI

	// ModeCLI 强制使用 CLI 模式
	// 即使其他配置启用了 TUI 也不使用
	ModeCLI
)

// DefaultTUIConfig 返回默认的 TUI 配置
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

// NewTUIConfig 创建新的 TUI 配置
// 支持函数式选项模式
func NewTUIConfig(opts ...TUIOption) *TUIConfig {
	config := DefaultTUIConfig()
	for _, opt := range opts {
		opt(config)
	}
	return config
}

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

// WithTUIShowDescription 设置是否显示命令描述
func WithTUIShowDescription(show bool) TUIOption {
	return func(c *TUIConfig) {
		c.ShowDescription = show
	}
}

// WithTUIShowFlags 设置是否显示 flag 配置面板
func WithTUIShowFlags(show bool) TUIOption {
	return func(c *TUIConfig) {
		c.ShowFlags = show
	}
}

// WithTUIInteractiveMode 设置交互模式
func WithTUIInteractiveMode(mode InteractiveMode) TUIOption {
	return func(c *TUIConfig) {
		c.InteractiveMode = mode
	}
}

// WithTUIAutoDetect 设置是否自动检测终端能力
func WithTUIAutoDetect(detect bool) TUIOption {
	return func(c *TUIConfig) {
		c.AutoDetect = detect
	}
}

// WithTUIConfirmBeforeExecute 设置是否在执行前确认
func WithTUIConfirmBeforeExecute(confirm bool) TUIOption {
	return func(c *TUIConfig) {
		c.ConfirmBeforeExecute = confirm
	}
}
