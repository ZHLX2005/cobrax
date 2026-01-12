package cobra

import (
	"fmt"
	"os"

	spf13cobra "github.com/spf13/cobra"
)

// EnhanceOption 装饰器选项
type EnhanceOption func(*EnhanceConfig)

// EnhanceConfig 增强配置
type EnhanceConfig struct {
	TreeTheme *TreeTheme
}

// Enhance 装饰器函数 - 增强原始 cobra.Command
// 添加 --tree flag 支持，不修改原有命令结构
//
// 使用示例：
//
//	enhanced := cobrax.Enhance(rootCmd,
//	    cobrax.WithThemeName("dracula"),
//	)
func Enhance(cmd *spf13cobra.Command, opts ...EnhanceOption) *spf13cobra.Command {
	if cmd == nil {
		return nil
	}

	// 创建增强配置
	config := &EnhanceConfig{
		TreeTheme: DefaultTreeTheme(),
	}

	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	// 添加 tree flags
	addTreeFlags(cmd)

	// 包装 PreRun/E 来处理 tree flag
	addTreeHandler(cmd, config)

	return cmd
}

// WithThemeName 设置主题名称
func WithThemeName(themeName string) EnhanceOption {
	return func(c *EnhanceConfig) {
		c.TreeTheme = GetTreeThemeByName(themeName)
	}
}

// WithTreeEnabled 启用树形展示（保留用于兼容性，实际通过 flag 控制）
func WithTreeEnabled(enabled bool) EnhanceOption {
	return func(c *EnhanceConfig) {
		// 树形展示通过 --tree flag 控制
	}
}

// addTreeFlags 添加 tree 相关的 flags
func addTreeFlags(cmd *spf13cobra.Command) {
	// 只添加到根命令，避免重复添加
	if cmd.Flags().Lookup("tree") == nil {
		cmd.Flags().Bool("tree", false, "Display command tree")
		cmd.Flags().String("tree-theme", "default", "Tree theme (default, dracula, nord, monokai, light)")
		cmd.Flags().Bool("tree-flags", false, "Show flags in tree view")
		cmd.Flags().Bool("tree-long", true, "Show long descriptions in tree view")
	}
}

// addTreeHandler 添加 tree 处理器
func addTreeHandler(cmd *spf13cobra.Command, config *EnhanceConfig) {
	// 保存原始的帮助函数
	oldHelpFunc := cmd.HelpFunc()

	// 设置新的帮助函数来检查 --tree flag
	cmd.SetHelpFunc(func(c *spf13cobra.Command, strs []string) {
		// 检查是否需要显示树
		if treeFlag, err := cmd.Flags().GetBool("tree"); err == nil && treeFlag {
			wrappedCmd := &Command{
				Command:    c,
				treeConfig: &TreeConfig{Theme: config.TreeTheme},
			}
			treeConfig := wrappedCmd.getTreeConfig()
			output := DisplayFlatTree(wrappedCmd, treeConfig)
			fmt.Println(output)
			os.Exit(0)
		}
		// 否则调用原始帮助函数
		if oldHelpFunc != nil {
			oldHelpFunc(c, strs)
		}
	})

	// 设置 PersistentPreRun 来处理有子命令的情况
	oldPersistentPreRunE := cmd.PersistentPreRunE
	cmd.PersistentPreRunE = func(c *spf13cobra.Command, args []string) error {
		// 检查是否需要显示树
		if treeFlag, err := cmd.Flags().GetBool("tree"); err == nil && treeFlag {
			wrappedCmd := &Command{
				Command:    c,
				treeConfig: &TreeConfig{Theme: config.TreeTheme},
			}
			treeConfig := wrappedCmd.getTreeConfig()
			output := DisplayFlatTree(wrappedCmd, treeConfig)
			fmt.Println(output)
			os.Exit(0)
		}
		if oldPersistentPreRunE != nil {
			return oldPersistentPreRunE(c, args)
		}
		return nil
	}
}
