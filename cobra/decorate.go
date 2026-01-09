package cobra

import (
	"fmt"
	"os"
	"strings"

	spf13cobra "github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ZHLX2005/cobra/tui"
	"github.com/ZHLX2005/cobra/tui/style"
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
}

// wrapExecute 包装执行逻辑
func wrapExecute(cmd *spf13cobra.Command, config *EnhanceConfig) {
	// 保存原有的执行函数
	originalPersistentPreRunE := cmd.PersistentPreRunE
	originalPreRunE := cmd.PreRunE
	originalRunE := cmd.RunE
	originalRun := cmd.Run

	// 包装 PreRunE
	cmd.PersistentPreRunE = func(c *spf13cobra.Command, args []string) error {
		// 检查是否需要启动 TUI
		if shouldUseTUIForCommand(c, config) {
			if err := executeTUIForCommand(c, config); err != nil {
				return err
			}
			// TUI 执行后不需要继续执行原有逻辑
			return nil
		}

		// 否则执行原有的 PreRunE
		if originalPersistentPreRunE != nil {
			return originalPersistentPreRunE(c, args)
		}
		return nil
	}

	cmd.PreRunE = func(c *spf13cobra.Command, args []string) error {
		if shouldUseTUIForCommand(c, config) {
			if err := executeTUIForCommand(c, config); err != nil {
				return err
			}
			return nil
		}

		if originalPreRunE != nil {
			return originalPreRunE(c, args)
		}
		return nil
	}

	// 包装 RunE
	if originalRunE != nil {
		cmd.RunE = func(c *spf13cobra.Command, args []string) error {
			if shouldUseTUIForCommand(c, config) {
				return executeTUIForCommand(c, config)
			}
			return originalRunE(c, args)
		}
	}

	// 包装 Run
	if originalRun != nil {
		cmd.Run = func(c *spf13cobra.Command, args []string) {
			if shouldUseTUIForCommand(c, config) {
				if err := executeTUIForCommand(c, config); err != nil {
					// TUI 模式出错，打印错误但不退出
					printError(err)
				}
				return
			}
			originalRun(c, args)
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

	// 检查配置
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

// navigateAndExecute 导航命令树并执行
func navigateAndExecute(renderer tui.Renderer, cmd *spf13cobra.Command, config *EnhanceConfig) error {
	// 如果有子命令，显示菜单
	children := getAvailableCommands(cmd.Commands())
	if len(children) > 0 {
		// 显示菜单让用户选择
		selectedIndex, err := renderer.RenderCommandMenu(cmd.Name(), buildMenuItems(children))
		if err != nil {
			return err
		}

		if selectedIndex < 0 {
			return nil // 用户取消
		}

		// 递归执行选中的命令
		selectedCmd := children[selectedIndex]
		return navigateAndExecute(renderer, selectedCmd, config)
	}

	// 叶子命令，配置 flags
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
		if !cmd.Hidden {
			result = append(result, cmd)
		}
	}
	return result
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
	if cmd.RunE != nil {
		return cmd.RunE(cmd, cmd.Flags().Args())
	}
	if cmd.Run != nil {
		cmd.Run(cmd, cmd.Flags().Args())
	}
	return nil
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
