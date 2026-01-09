package tui

import (
	// 不依赖 cobra 包，避免循环导入
)

// Renderer TUI 面板渲染器接口
// 定义了 TUI 界面的核心渲染方法
// 用户可以实现此接口来自定义 TUI 的外观和行为
type Renderer interface {
	// RenderCommandMenu 渲染命令菜单面板
	// 显示可用的子命令列表，让用户选择
	// 返回用户选择的命令索引，如果取消则返回 -1
	RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error)

	// RenderFlagForm 渲染 flag 输入表单
	// 显示所有需要配置的 flags，让用户输入或选择值
	// 返回 flag 名称到值的映射
	RenderFlagForm(formTitle string, flags []FlagItem) (values map[string]string, err error)

	// RenderConfirmation 渲染确认面板
	// 显示将要执行的命令，询问用户是否确认
	// 返回用户是否确认
	RenderConfirmation(title, message string) (confirmed bool, err error)

	// RenderHelp 渲染帮助面板
	// 显示命令的帮助信息
	RenderHelp(title, content string) error

	// Cleanup 清理资源
	// 在渲染器不再使用时调用，用于清理终端状态
	Cleanup() error
}

// MenuItem 菜单项
// 用于命令菜单面板中显示一个可选择的命令
type MenuItem struct {
	// ID 命令唯一标识符（如命令名称）
	ID string

	// Label 显示文本
	Label string

	// Description 描述文本
	Description string

	// Disabled 是否禁用此选项
	Disabled bool

	// Metadata 附加元数据
	// 可用于存储自定义信息
	Metadata map[string]interface{}
}

// FlagItem flag 项
// 用于 flag 表单面板中显示一个可配置的 flag
type FlagItem struct {
	// Name flag 名称
	Name string

	// ShortName 短名称（单字符）
	ShortName string

	// Description flag 描述
	Description string

	// DefaultValue 默认值
	DefaultValue string

	// CurrentValue 当前值
	CurrentValue string

	// Required 是否必填
	Required bool

	// Type flag 类型
	Type FlagType

	// SourceCommand 参数来源的命令名称
	SourceCommand string

	// Options 可选值列表
	// 对于枚举类型的 flag，限制可选择的值
	Options []FlagOption

	// Validator 自定义验证器
	// 验证用户输入的值是否有效
	Validator func(value string) error

	// Metadata 附加元数据
	Metadata map[string]interface{}
}

// FlagType flag 类型枚举
type FlagType int

const (
	// FlagTypeString 字符串类型
	FlagTypeString FlagType = iota

	// FlagTypeBool 布尔类型
	FlagTypeBool

	// FlagTypeInt 整数类型
	FlagTypeInt

	// FlagTypeDuration 时间段类型
	FlagTypeDuration

	// FlagTypeEnum 枚举类型
	FlagTypeEnum
)

// FlagOption flag 可选值
// 用于枚举类型的 flag
type FlagOption struct {
	// Value 值
	Value string

	// Label 显示标签
	// 如果为空，使用 Value
	Label string

	// Description 选项描述
	Description string
}

// RenderContext 渲染上下文
// 在渲染过程中传递的上下文信息
type RenderContext struct {
	// Width 终端宽度
	Width int

	// Height 终端高度
	Height int

	// Theme 主题
	// 注意：这里使用 interface{} 来避免循环依赖
	// 实际使用时会转换为具体的 Theme 类型
	Theme interface{}

	// Metadata 附加元数据
	Metadata map[string]interface{}
}

// RenderResult 渲染结果
// 渲染操作的通用返回值
type RenderResult struct {
	// Success 是否成功
	Success bool

	// Cancelled 是否被用户取消
	Cancelled bool

	// Error 错误信息
	Error error

	// Data 附加数据
	Data map[string]interface{}
}
