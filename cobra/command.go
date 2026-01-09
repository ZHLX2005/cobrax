package cobra

import (
	"fmt"
	"os"
	"strings"

	spf13cobra "github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ZHLX2005/cobrax/tui"
	"github.com/ZHLX2005/cobrax/tui/style"
)

// Command 是 cobra-x 的核心命令结构
// 它嵌入 spf13cobra.Command 以保持完全的 API 兼容性
// 所有原始 cobra 的方法都可用，同时扩展了 TUI 功能
type Command struct {
	*spf13cobra.Command

	// tuiConfig TUI 配置
	tuiConfig *TUIConfig

	// tuiFlags TUI 相关的 flags
	tuiFlags *pflag.FlagSet

	// children 子命令缓存（用于 TUI 导航）
	children []*Command
}

// NewCommand 创建一个新的命令
// 这是创建 cobra-x 命令的推荐方式
func NewCommand(use string, opts ...CommandOption) *Command {
	return newCommandWithCobra(&spf13cobra.Command{Use: use}, opts...)
}

// newCommandWithCobra 从现有的 cobra 命令创建
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
	// 由于嵌入的关系，我们无法直接做类型断言
	// 检查命令是否已经被我们包装过（通过检查字段是否存在）
	// 这里简化处理：总是创建新的包装
	return &Command{
		Command:   cmd,
		tuiConfig: c.tuiConfig,
	}
}

// WithTUIConfig 设置 TUI 配置
func WithTUIConfig(config *TUIConfig) CommandOption {
	return func(c *Command) {
		c.tuiConfig = config
	}
}

// WithTUIEnabledOption 启用 TUI（选项函数版本）
func WithTUIEnabledOption(enabled bool) CommandOption {
	return func(c *Command) {
		if c.tuiConfig == nil {
			c.tuiConfig = DefaultTUIConfig()
		}
		c.tuiConfig.Enabled = enabled
	}
}

// initTUIFlags 初始化 TUI 相关的 flags
func (c *Command) initTUIFlags() {
	// 添加 TUI flags 到主命令的 flag set
	// 注意：不使用短选项以避免与用户 flags 冲突
	c.Flags().Bool("tui", false, "Launch TUI interface")
	c.Flags().String("tui-theme", "default", "TUI theme (default, dark, light, minimal, dracula, nord, monokai)")
	c.Flags().Bool("tui-confirm", true, "Show confirmation before executing command")
	c.Flags().Bool("tui-flags", true, "Show flag configuration panel")
}

// Execute 执行命令
// 这是命令执行的入口点，会根据配置选择 TUI 或 CLI 模式
func (c *Command) Execute() error {
	// 检查是否应该使用 TUI
	if c.shouldUseTUI() {
		return c.executeTUI()
	}

	// 使用传统 CLI 模式
	return c.Command.Execute()
}

// ExecuteE 执行命令并返回错误
func (c *Command) ExecuteE() error {
	return c.Execute()
}

// shouldUseTUI 判断是否应该使用 TUI 模式
func (c *Command) shouldUseTUI() bool {
	// 检查强制 CLI 模式
	if c.tuiConfig != nil && c.tuiConfig.InteractiveMode == ModeCLI {
		return false
	}

	// 检查强制 TUI 模式
	if c.tuiConfig != nil && c.tuiConfig.InteractiveMode == ModeTUI {
		return true
	}

	// 检查 --tui flag
	if tuiFlag, err := c.Flags().GetBool("tui"); err == nil && tuiFlag {
		return true
	}

	// 检查环境变量
	if os.Getenv("COBRA_TUI") == "true" {
		return true
	}

	// 检查配置是否启用
	if c.tuiConfig != nil && c.tuiConfig.Enabled {
		// 自动检测终端能力
		if c.tuiConfig.AutoDetect {
			return c.isInteractiveTerminal()
		}
		return true
	}

	return false
}

// isInteractiveTerminal 检测是否为交互式终端
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

// executeTUI 使用 TUI 模式执行命令
func (c *Command) executeTUI() error {
	// 获取渲染器
	renderer := c.getRenderer()
	defer renderer.Cleanup()

	// 从根命令开始导航
	selectedPath, err := c.navigateCommandTree(renderer, c, []*Command{})
	if err != nil {
		return err
	}

	if len(selectedPath) == 0 {
		// 用户取消
		return nil
	}

	// 获取最终选中的命令
	selectedCmd := selectedPath[len(selectedPath)-1]

	// 配置 flags
	if c.tuiConfig.ShowFlags {
		flagValues, err := c.configureFlags(renderer, selectedCmd)
		if err != nil {
			return err
		}

		// 应用 flag 值
		if err := c.applyFlagValues(selectedCmd, flagValues); err != nil {
			return err
		}
	}

	// 确认执行
	if c.tuiConfig.ConfirmBeforeExecute {
		confirmed, err := c.confirmExecution(renderer, selectedPath)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	// 执行命令
	return c.executeCommand(selectedCmd)
}

// navigateCommandTree 导航命令树
func (c *Command) navigateCommandTree(renderer tui.Renderer, cmd *Command, path []*Command) ([]*Command, error) {
	// 获取子命令
	children := c.getChildren(cmd)
	if len(children) == 0 {
		// 叶子命令，返回当前路径
		return append(path, cmd), nil
	}

	// 构建菜单项
	menuItems := make([]tui.MenuItem, 0, len(children))
	for _, child := range children {
		menuItems = append(menuItems, tui.MenuItem{
			ID:          child.Use,
			Label:       child.Use,
			Description: child.Short,
			Disabled:    !child.IsAvailableCommand(),
		})
	}

	// 如果没有可用的子命令，返回当前命令
	if len(menuItems) == 0 {
		return append(path, cmd), nil
	}

	// 渲染菜单
	selectedIndex, err := renderer.RenderCommandMenu(cmd.Use, menuItems)
	if err != nil {
		return nil, err
	}

	if selectedIndex < 0 {
		return nil, nil // 用户取消
	}

	// 递归处理选中的命令
	selectedChild := children[selectedIndex]
	newPath := append(path, cmd)
	return c.navigateCommandTree(renderer, selectedChild, newPath)
}

// getChildren 获取子命令列表
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

// configureFlags 配置 flags
func (c *Command) configureFlags(renderer tui.Renderer, cmd *Command) (map[string]string, error) {
	flagItems := c.collectFlagItems(cmd)

	if len(flagItems) == 0 {
		return nil, nil
	}

	// 渲染 flag 表单
	return renderer.RenderFlagForm("Configure: "+cmd.Use, flagItems)
}

// collectFlagItems 收集 flag 项（包括所有父命令的 flags）
func (c *Command) collectFlagItems(cmd *Command) []tui.FlagItem {
	var items []tui.FlagItem
	seen := make(map[string]bool) // 用于跟踪已添加的 flags，避免重复

	// 遍历当前命令及其所有父命令，聚合所有 flags
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

		if current.Parent() == nil {
			break
		}
		current = c.wrapCommand(current.Parent())
	}

	return items
}

// applyFlagValues 应用 flag 值（包括所有父命令的 flags）
func (c *Command) applyFlagValues(cmd *Command, values map[string]string) error {
	// 遍历当前命令及其所有父命令，应用对应的 flag 值
	current := cmd
	for current != nil {
		for name, value := range values {
			flag := current.LocalFlags().Lookup(name)
			if flag != nil {
				if err := flag.Value.Set(value); err != nil {
					return fmt.Errorf("failed to set flag %s: %w", name, err)
				}

				// 标记为已改变
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

// confirmExecution 确认执行
func (c *Command) confirmExecution(renderer tui.Renderer, path []*Command) (bool, error) {
	// 构建命令字符串
	cmdString := c.buildCommandString(path)

	return renderer.RenderConfirmation(
		"Confirm Execution",
		fmt.Sprintf("Command to execute:\n\n  %s", cmdString),
	)
}

// buildCommandString 构建命令字符串
func (c *Command) buildCommandString(path []*Command) string {
	var parts []string
	seen := make(map[string]bool) // 用于跟踪已添加的 flags，避免重复

	// 添加命令路径
	for _, cmd := range path {
		parts = append(parts, cmd.Use)
	}

	// 遍历所有命令（包括所有父命令），添加所有变更后的 flags
	for _, cmd := range path {
		cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
			if flag.Changed && flag.Name != "help" && !seen[flag.Name] {
				key := fmt.Sprintf("--%s", flag.Name)
				if flag.Value.Type() == "bool" {
					if flag.Value.String() == "true" {
						parts = append(parts, key)
					}
				} else {
					parts = append(parts, fmt.Sprintf("%s=%s", key, flag.Value.String()))
				}
				seen[flag.Name] = true
			}
		})
	}

	return strings.Join(parts, " ")
}

// executeCommand 执行命令
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

// getTheme 获取主题
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

// getRenderer 获取渲染器
func (c *Command) getRenderer() tui.Renderer {
	if c.tuiConfig != nil && c.tuiConfig.Renderer != nil {
		return c.tuiConfig.Renderer
	}

	// 返回默认渲染器
	return tui.NewDefaultRenderer(c.getTheme())
}

// EnableTUI 启用 TUI 模式
func (c *Command) EnableTUI() {
	if c.tuiConfig == nil {
		c.tuiConfig = DefaultTUIConfig()
	}
	c.tuiConfig.Enabled = true
}

// DisableTUI 禁用 TUI 模式
func (c *Command) DisableTUI() {
	if c.tuiConfig == nil {
		c.tuiConfig = DefaultTUIConfig()
	}
	c.tuiConfig.Enabled = false
}

// SetTUIConfig 设置 TUI 配置
func (c *Command) SetTUIConfig(config *TUIConfig) {
	c.tuiConfig = config
}

// GetTUIConfig 获取 TUI 配置
func (c *Command) GetTUIConfig() *TUIConfig {
	return c.tuiConfig
}

// SetTUIRenderer 设置自定义渲染器
func (c *Command) SetTUIRenderer(renderer tui.Renderer) {
	if c.tuiConfig == nil {
		c.tuiConfig = DefaultTUIConfig()
	}
	c.tuiConfig.Renderer = renderer
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
