package cobra

import (
	"fmt"
	"os"

	spf13cobra "github.com/spf13/cobra"
)

// Command 是 cobra-x 的核心命令结构
// 它嵌入 spf13cobra.Command 以保持完全的 API 兼容性
type Command struct {
	*spf13cobra.Command

	// treeConfig 树形展示配置
	treeConfig *TreeConfig
}

// NewCommand 创建一个新的命令
func NewCommand(use string, opts ...CommandOption) *Command {
	return newCommandWithCobra(&spf13cobra.Command{Use: use}, opts...)
}

// newCommandWithCobra 从现有的 cobra 命令创建
func newCommandWithCobra(cobraCmd *spf13cobra.Command, opts ...CommandOption) *Command {
	cmd := &Command{
		Command:    cobraCmd,
		treeConfig: &TreeConfig{Theme: DefaultTreeTheme()},
	}

	// 应用选项
	for _, opt := range opts {
		opt(cmd)
	}

	// 初始化 tree flags
	cmd.initTreeFlags()

	return cmd
}

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

// WithRun 设置命令执行函数
func WithRun(fn func(*Command, []string)) CommandOption {
	return func(c *Command) {
		c.Command.Run = func(cmd *spf13cobra.Command, args []string) {
			wrappedCmd := c.wrapCommand(cmd)
			fn(wrappedCmd, args)
		}
	}
}

// WithRunE 设置带返回值的命令执行函数
func WithRunE(fn func(*Command, []string) error) CommandOption {
	return func(c *Command) {
		c.Command.RunE = func(cmd *spf13cobra.Command, args []string) error {
			wrappedCmd := c.wrapCommand(cmd)
			return fn(wrappedCmd, args)
		}
	}
}

// wrapCommand 包装 spf13cobra.Command 为 cobra.Command
func (c *Command) wrapCommand(cmd *spf13cobra.Command) *Command {
	return &Command{
		Command:    cmd,
		treeConfig: c.treeConfig,
	}
}

// WithTreeTheme 设置树形展示主题
func WithTreeTheme(theme *TreeTheme) CommandOption {
	return func(c *Command) {
		c.treeConfig = &TreeConfig{Theme: theme}
	}
}

// initTreeFlags 初始化 tree 相关的 flags
func (c *Command) initTreeFlags() {
	// 添加 tree flags 到主命令的 flag set
	c.Flags().Bool("tree", false, "Display command tree")
	c.Flags().String("tree-theme", "default", "Tree theme (default, dracula, nord, monokai, light)")
	c.Flags().Bool("tree-flags", false, "Show flags in tree view")
	c.Flags().Bool("tree-long", true, "Show long descriptions in tree view")
}

// Execute 执行命令
func (c *Command) Execute() error {
	// 保存原始的帮助函数
	oldHelpFunc := c.HelpFunc()

	// 设置新的帮助函数来检查 --tree flag
	c.SetHelpFunc(func(command *spf13cobra.Command, strs []string) {
		// 检查是否显示树形视图
		if c.shouldShowTree() {
			c.showTree()
			os.Exit(0)
		}
		// 否则调用原始帮助函数
		if oldHelpFunc != nil {
			oldHelpFunc(command, strs)
		}
	})

	// 设置 PersistentPreRun 来处理有 Run 函数的命令
	oldPersistentPreRunE := c.PersistentPreRunE
	c.PersistentPreRunE = func(cmd *spf13cobra.Command, args []string) error {
		if c.shouldShowTree() {
			c.showTree()
			os.Exit(0)
		}
		if oldPersistentPreRunE != nil {
			return oldPersistentPreRunE(cmd, args)
		}
		return nil
	}

	// 使用传统 CLI 模式
	return c.Command.Execute()
}

// ExecuteE 执行命令并返回错误
func (c *Command) ExecuteE() error {
	return c.Execute()
}

// shouldShowTree 判断是否应该显示树形视图
func (c *Command) shouldShowTree() bool {
	// 检查 --tree flag
	if treeFlag, err := c.Flags().GetBool("tree"); err == nil && treeFlag {
		return true
	}

	// 检查环境变量
	if os.Getenv("COBRA_TREE") == "true" {
		return true
	}

	return false
}

// showTree 显示命令树
func (c *Command) showTree() {
	// 获取配置
	config := c.getTreeConfig()

	// 显示扁平化的命令树
	output := DisplayFlatTree(c, config)
	fmt.Println(output)
}

// getTreeConfig 获取树形配置
func (c *Command) getTreeConfig() *TreeConfig {
	config := &TreeConfig{
		Theme:     c.treeConfig.Theme,
		ShowFlags: false,
		ShowLong:  true,
	}

	// 从 flags 读取配置
	if showFlags, err := c.Flags().GetBool("tree-flags"); err == nil {
		config.ShowFlags = showFlags
	}

	if showLong, err := c.Flags().GetBool("tree-long"); err == nil {
		config.ShowLong = showLong
	}

	// 从 flags 读取主题名称
	if themeName, err := c.Flags().GetString("tree-theme"); err == nil {
		config.Theme = GetTreeThemeByName(themeName)
	}

	return config
}

// SetTreeTheme 设置树形展示主题
func (c *Command) SetTreeTheme(theme *TreeTheme) {
	if c.treeConfig == nil {
		c.treeConfig = &TreeConfig{}
	}
	c.treeConfig.Theme = theme
}

// GetTreeConfig 获取树形配置
func (c *Command) GetTreeConfig() *TreeConfig {
	return c.treeConfig
}

// AddCommand 添加子命令
func (c *Command) AddCommand(cmds ...*Command) {
	for _, cmd := range cmds {
		c.Command.AddCommand(cmd.Command)
	}
}

// AddSpf13Command 添加原始 spf13/cobra 命令
func (c *Command) AddSpf13Command(cmds ...*spf13cobra.Command) {
	c.Command.AddCommand(cmds...)
}

// ==================== 兼容性方法（保持 API 兼容） ====================

// EnableTUI 启用 TUI 模式（已废弃，无实际作用）
func (c *Command) EnableTUI() {
	// 不再支持 TUI 交互模式
}

// DisableTUI 禁用 TUI 模式（已废弃，无实际作用）
func (c *Command) DisableTUI() {
	// 不再支持 TUI 交互模式
}

// SetTUIConfig 设置 TUI 配置（已废弃，无实际作用）
func (c *Command) SetTUIConfig(config interface{}) {
	// 不再支持 TUI 交互模式
}

// GetTUIConfig 获取 TUI 配置（已废弃，返回 nil）
func (c *Command) GetTUIConfig() interface{} {
	return nil
}

// SetTUIRenderer 设置自定义渲染器（已废弃，无实际作用）
func (c *Command) SetTUIRenderer(renderer interface{}) {
	// 不再支持 TUI 交互模式
}

// SetPanelBuilder 设置面板构建器（已废弃，无实际作用）
func (c *Command) SetPanelBuilder(builder interface{}) {
	// 不再支持 TUI 交互模式
}

// isInteractiveTerminal 检测是否为交互式终端（保留用于兼容性）
func (c *Command) isInteractiveTerminal() bool {
	// 检查 stdout 是否为终端
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	// 检查是否为字符设备
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	// 检查是否有 stdin
	stdinFi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (stdinFi.Mode() & os.ModeCharDevice) != 0
}

// executeCommand 执行命令（内部使用）
func (c *Command) executeCommand(cmd *Command) error {
	if cmd.RunE != nil {
		return cmd.RunE(cmd.Command, cmd.Flags().Args())
	}

	if cmd.Run != nil {
		cmd.Run(cmd.Command, cmd.Flags().Args())
		return nil
	}

	// 没有执行函数，显示帮助
	return cmd.Help()
}

// getChildren 获取子命令列表（内部使用）
func (c *Command) getChildren(cmd *Command) []*Command {
	spf13Children := cmd.Commands()
	children := make([]*Command, 0, len(spf13Children))

	for _, child := range spf13Children {
		if !child.IsAvailableCommand() {
			continue
		}
		children = append(children, c.wrapCommand(child))
	}

	return children
}

// ==================== 装饰器模式支持 ====================

// 装饰器相关的代码在 decorate.go 中定义
