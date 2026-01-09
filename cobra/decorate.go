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

// EnhanceOption 装饰器选项
type EnhanceOption func(*EnhanceConfig)

// EnhanceConfig 增强配置
type EnhanceConfig struct {
	TUIConfig *TUIConfig
}

// Enhance 装饰器函数 - 增强原始 cobra.Command
// 保持原有的命令结构不变，只添加 TUI 功能
//
// 使用示例：
//
//	rootCmd := cobrax.Enhance(rootCmd,
//	    cobrax.WithEnhanceTUIEnabled(true),
//	    cobrax.WithEnhanceTheme("dracula"),
//	)
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
	if config.TUIConfig != nil {
		// 将配置序列化存储（简化版：只存储关键配置）
		if config.TUIConfig.Theme != nil {
			cmd.Annotations["tui.theme"] = config.TUIConfig.Theme.Name
		}
	}

	// 添加 TUI flags
	addTUIFlags(cmd)

	// 包装 PreRun/E 以拦截执行
	wrapExecute(cmd, config)

	return cmd
}

// addTUIFlags 添加 TUI 相关的 flags
func addTUIFlags(cmd *spf13cobra.Command) {
	// 只添加到根命令，避免重复添加
	if cmd.Flags().Lookup("tui") == nil {
		cmd.Flags().Bool("tui", false, "Launch TUI interface")
		cmd.Flags().String("tui-theme", "default", "TUI theme")
		cmd.Flags().Bool("tui-confirm", true, "Show confirmation before execution")
		cmd.Flags().Bool("tui-flags", true, "Show flag configuration panel")
	}
	// 添加内部标记用于追踪 TUI 执行状态
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations["tui.executed"] = "false"
}

// wrapExecute 包装执行逻辑
func wrapExecute(cmd *spf13cobra.Command, config *EnhanceConfig) {
	// 保存原有的执行函数
	originalPersistentPreRunE := cmd.PersistentPreRunE
	originalPreRunE := cmd.PreRunE
	originalRunE := cmd.RunE
	originalRun := cmd.Run

	// 包装 HelpFunc - Check for TUI mode before showing help
	// This is needed because cobra shows help without calling Run for commands with subcommands
	cmd.SetHelpFunc(func(c *spf13cobra.Command, s []string) {
		// Check if TUI should be used
		if shouldUseTUIForCommand(c, config) && c.Annotations["tui.executed"] == "false" {
			c.Annotations["tui.executed"] = "true"
			if err := executeTUIForCommand(c, config); err != nil {
				printError(err)
			}
			return
		}
		// Call original HelpFunc - use the command's method directly
		// The signature may vary by cobra version, so we check what the original function expects
		cmd.HelpFunc() // Use default help
	})

	// 包装 PersistentPreRunE - This runs before any command, including root
	// Check for TUI mode here since cobra may not call Run for commands with subcommands
	cmd.PersistentPreRunE = func(c *spf13cobra.Command, args []string) error {
		// Execute original PersistentPreRunE
		if originalPersistentPreRunE != nil {
			return originalPersistentPreRunE(c, args)
		}
		return nil
	}

	cmd.PreRunE = func(c *spf13cobra.Command, args []string) error {
		if originalPreRunE != nil {
			return originalPreRunE(c, args)
		}
		return nil
	}

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

	// 包装 Run
	if originalRun != nil {
		cmd.Run = func(c *spf13cobra.Command, args []string) {
			// 检查是否需要启动 TUI（只执行一次）
			if shouldUseTUIForCommand(c, config) && c.Annotations["tui.executed"] == "false" {
				c.Annotations["tui.executed"] = "true"
				if err := executeTUIForCommand(c, config); err != nil {
					// TUI 模式出错，打印错误但不退出
					printError(err)
				}
				return
			}
			originalRun(c, args)
		}
	} else if originalRunE == nil {
		// 如果命令既没有 Run 也没有 RunE，添加一个占位 Run 函数
		// 这样 cobra 才会执行命令，从而触发 PersistentPreRunE
		// 这对于有子命令的根命令特别重要
		cmd.Run = func(c *spf13cobra.Command, args []string) {
			// 检查是否需要启动 TUI（只执行一次）
			if shouldUseTUIForCommand(c, config) && c.Annotations["tui.executed"] == "false" {
				c.Annotations["tui.executed"] = "true"
				if err := executeTUIForCommand(c, config); err != nil {
					printError(err)
				}
				return
			}
			// 如果没有启用 TUI，显示帮助（默认行为）
			c.Help()
		}
	}
}

// shouldUseTUIForCommand 判断是否应该使用 TUI
func shouldUseTUIForCommand(cmd *spf13cobra.Command, config *EnhanceConfig) bool {
	// 检查 --tui flag
	if tuiFlag, err := cmd.Flags().GetBool("tui"); err == nil && tuiFlag {
		return true
	}

	// 检查环境变量
	if os.Getenv("COBRA_TUI") == "true" {
		return true
	}

	if config.TUIConfig != nil && config.TUIConfig.Enabled {
		if config.TUIConfig.InteractiveMode == ModeTUI {
			return true
		}
		if config.TUIConfig.InteractiveMode == ModeAuto && checkIsInteractiveTerminal() {
			return true
		}
	}

	return false
}

// executeTUIForCommand 执行 TUI
func executeTUIForCommand(cmd *spf13cobra.Command, config *EnhanceConfig) error {
	// 获取主题
	theme := getThemeForCommand(cmd, config)

	// 获取渲染器
	renderer := getRendererForCommand(theme)
	defer renderer.Cleanup()

	// 导航并执行命令
	return navigateAndExecute(renderer, cmd, config)
}

// navigateAndExecute 导航命令树并执行（支持扁平化视图）
func navigateAndExecute(renderer tui.Renderer, cmd *spf13cobra.Command, config *EnhanceConfig) error {
	// 获取所有可执行命令（扁平化列表）
	executableCommands := GetExecutableCommands(cmd)

	// 如果只有一个可执行命令且是当前命令，直接执行
	if len(executableCommands) == 1 && executableCommands[0].ID == cmd.Name() && (cmd.Run != nil || cmd.RunE != nil) {
		return executeLeafCommand(renderer, cmd, config)
	}

	// 如果有多个可执行命令，显示扁平化菜单
	if len(executableCommands) > 0 {
		// 构建菜单项，显示完整路径
		menuItems := make([]tui.MenuItem, 0, len(executableCommands))
		for _, execCmd := range executableCommands {
			// 构建显示路径（去掉根命令名称）
			displayPath := strings.TrimPrefix(execCmd.Use, cmd.Name()+" ")
			if displayPath == execCmd.Use {
				displayPath = execCmd.Name
			}

			menuItems = append(menuItems, tui.MenuItem{
				ID:          execCmd.ID,
				Label:       displayPath,
				Description: execCmd.Short,
			})
		}

		// 显示菜单让用户选择
		selectedIndex, err := renderer.RenderCommandMenu(cmd.Name()+" Commands", menuItems)
		if err != nil {
			return err
		}

		if selectedIndex < 0 {
			return nil // 用户取消
		}

		// 根据选择的 ID 查找对应的命令
		selectedID := executableCommands[selectedIndex].ID
		selectedCmd := findCommandByID(cmd, selectedID)

		if selectedCmd == nil {
			// 如果找不到命令，尝试通过路径查找
			pathParts := strings.Fields(menuItems[selectedIndex].Label)
			selectedCmd = FindCommandByPath(cmd, strings.Join(pathParts, " "))
		}

		if selectedCmd != nil {
			return executeLeafCommand(renderer, selectedCmd, config)
		}
	}

	// 如果没有可执行的子命令，尝试执行当前命令
	if cmd.Run != nil || cmd.RunE != nil {
		return executeLeafCommand(renderer, cmd, config)
	}

	// 显示帮助
	return cmd.Help()
}

// executeLeafCommand 执行叶子命令
func executeLeafCommand(renderer tui.Renderer, cmd *spf13cobra.Command, config *EnhanceConfig) error {
	// 配置 flags
	if config.TUIConfig != nil && config.TUIConfig.ShowFlags {
		flagValues, err := configureFlags(renderer, cmd)
		if err != nil {
			return err
		}

		// 应用 flag 值
		applyFlagValues(cmd, flagValues)
	}

	// 确认执行
	if config.TUIConfig != nil && config.TUIConfig.ConfirmBeforeExecute {
		confirmed, err := renderer.RenderConfirmation(
			"Confirm",
			buildCommandPreview(cmd),
		)
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	// 执行命令
	return executeOriginalCommand(cmd)
}

// findCommandByID 在命令树中查找指定 ID 的命令
func findCommandByID(root *spf13cobra.Command, id string) *spf13cobra.Command {
	if root.Name() == id {
		return root
	}

	for _, cmd := range root.Commands() {
		if cmd.Name() == id {
			return cmd
		}
		if found := findCommandByID(cmd, id); found != nil {
			return found
		}
	}

	return nil
}

// ============================================================
// 装饰器选项函数
// ============================================================

// WithEnhanceTUIEnabled 启用 TUI
func WithEnhanceTUIEnabled(enabled bool) EnhanceOption {
	return func(c *EnhanceConfig) {
		if c.TUIConfig == nil {
			c.TUIConfig = DefaultTUIConfig()
		}
		c.TUIConfig.Enabled = enabled
	}
}

// WithEnhanceTheme 设置主题
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

// WithEnhanceTUIConfig 完整的 TUI 配置
func WithEnhanceTUIConfig(cfg *TUIConfig) EnhanceOption {
	return func(c *EnhanceConfig) {
		c.TUIConfig = cfg
	}
}

// WithEnhanceTUIConfirm 设置是否确认
func WithEnhanceTUIConfirm(confirm bool) EnhanceOption {
	return func(c *EnhanceConfig) {
		if c.TUIConfig == nil {
			c.TUIConfig = DefaultTUIConfig()
		}
		c.TUIConfig.ConfirmBeforeExecute = confirm
	}
}

// ============================================================
// 辅助函数
// ============================================================

func getAvailableCommands(cmds []*spf13cobra.Command) []*spf13cobra.Command {
	var result []*spf13cobra.Command
	for _, cmd := range cmds {
		if !cmd.IsAvailableCommand() {
			continue
		}
		if cmd.Hidden {
			continue
		}

		// Filter out completion commands (bash, fish, powershell, zsh, etc.)
		// These are generated by cobra and should not appear in TUI menu
		if isCompletionCommand(cmd) {
			continue
		}

		// Filter out help command
		if cmd.Name() == "help" {
			continue
		}

		result = append(result, cmd)
	}
	return result
}

// isCompletionCommand checks if a command is a completion-related command
func isCompletionCommand(cmd *spf13cobra.Command) bool {
	name := cmd.Name()
	// Check for standard shell completion commands
	completionShells := []string{"bash", "fish", "powershell", "zsh", "pwsh"}
	for _, shell := range completionShells {
		if name == shell {
			return true
		}
	}
	// Check if it's the completion command itself
	if name == "completion" {
		return true
	}
	// Check if command has "completion" annotation or group
	if cmd.Annotations != nil {
		if cmd.Annotations["command"] == "completion" {
			return true
		}
	}
	return false
}

func buildMenuItems(cmds []*spf13cobra.Command) []tui.MenuItem {
	items := make([]tui.MenuItem, 0, len(cmds))
	for _, cmd := range cmds {
		items = append(items, tui.MenuItem{
			ID:          cmd.Name(),
			Label:       cmd.Name(),
			Description: cmd.Short,
		})
	}
	return items
}

func configureFlags(renderer tui.Renderer, cmd *spf13cobra.Command) (map[string]string, error) {
	flags := cmd.LocalFlags()
	var items []tui.FlagItem

	flags.VisitAll(func(flag *pflag.Flag) {
		if isTUIFlag(flag.Name) {
			return
		}
		items = append(items, tui.FlagItem{
			Name:         flag.Name,
			Description:  flag.Usage,
			DefaultValue: flag.DefValue,
			CurrentValue: flag.DefValue,
		})
	})

	if len(items) == 0 {
		return nil, nil
	}

	return renderer.RenderFlagForm("Configure: "+cmd.Name(), items)
}

func applyFlagValues(cmd *spf13cobra.Command, values map[string]string) {
	for name, value := range values {
		flag := cmd.LocalFlags().Lookup(name)
		if flag != nil {
			flag.Value.Set(value)
			flag.Changed = true
		}
	}
}

func buildCommandPreview(cmd *spf13cobra.Command) string {
	var parts []string

	// 构建命令路径
	parts = append(parts, cmd.Name())

	// 添加 flags
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed && !isTUIFlag(flag.Name) {
			key := fmt.Sprintf("--%s", flag.Name)
			if flag.Value.Type() == "bool" {
				if flag.Value.String() == "true" {
					parts = append(parts, key)
				}
			} else {
				parts = append(parts, fmt.Sprintf("%s=%s", key, flag.Value.String()))
			}
		}
	})

	return strings.Join(parts, " ")
}

func executeOriginalCommand(cmd *spf13cobra.Command) error {
	// Get the root command
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	// Build the command path (e.g., "gocmds note")
	cmdPath := GetCommandFullPath(cmd)

	// Build args: split command path and add any remaining args
	args := strings.Fields(cmdPath)
	// Add any additional args from flags
	if remainingArgs := cmd.Flags().Args(); len(remainingArgs) > 0 {
		args = append(args, remainingArgs...)
	}

	// Set the args for execution and execute through root command
	rootCmd.SetArgs(args)
	return rootCmd.Execute()
}

func getThemeForCommand(cmd *spf13cobra.Command, config *EnhanceConfig) *style.Theme {
	// 优先使用配置的主题
	if config.TUIConfig != nil && config.TUIConfig.Theme != nil {
		return config.TUIConfig.Theme
	}

	// 从 flag 读取
	if themeName, err := cmd.Flags().GetString("tui-theme"); err == nil {
		return style.NewTheme(themeName)
	}

	return style.DefaultTheme()
}

func getRendererForCommand(theme *style.Theme) tui.Renderer {
	return tui.NewDefaultRenderer(theme)
}

func isTUIFlag(name string) bool {
	return name == "tui" ||
		name == "tui-theme" ||
		name == "tui-confirm" ||
		name == "tui-flags"
}

func checkIsInteractiveTerminal() bool {
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

func printError(err error) {
	// 简单的错误打印
	println("Error:", err.Error())
}
