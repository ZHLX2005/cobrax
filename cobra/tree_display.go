package cobra

import (
	"fmt"
	"strings"

	spf13cobra "github.com/spf13/cobra"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/pflag"
)

// TreeConfig 树形展示配置
type TreeConfig struct {
	Theme       *TreeTheme
	ShowFlags   bool
	ShowLong    bool
	IndentWidth int
}

// TreeTheme 树形展示主题
type TreeTheme struct {
	// 样式
	RootStyle       lipgloss.Style
	BranchStyle     lipgloss.Style
	LeafStyle       lipgloss.Style
	DescriptionStyle lipgloss.Style
	FlagStyle       lipgloss.Style
	LineStyle       lipgloss.Style
}

// DefaultTreeTheme 返回默认树形主题
func DefaultTreeTheme() *TreeTheme {
	return &TreeTheme{
		RootStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("86")), // blue
		BranchStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("228")), // yellow
		LeafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")), // white
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("242")). // gray
			Faint(true),
		FlagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("159")), // cyan
		LineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")), // dark gray
	}
}

// DraculaTreeTheme 返回 Dracula 主题
func DraculaTreeTheme() *TreeTheme {
	return &TreeTheme{
		RootStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#BD93F9")), // purple
		BranchStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF79C6")), // pink
		LeafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")), // foreground
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")). // comment
			Faint(true),
		FlagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")), // green
		LineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#44475A")), // current line
	}
}

// NordTreeTheme 返回 Nord 主题
func NordTreeTheme() *TreeTheme {
	return &TreeTheme{
		RootStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#88C0D0")), // frost
		BranchStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#EBCB8B")), // yellow
		LeafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D8DEE9")), // snow storm
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4C566A")). // dark
			Faint(true),
		FlagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A3BE8C")), // green
		LineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B4252")), // polar night
	}
}

// MonokaiTreeTheme 返回 Monokai 主题
func MonokaiTreeTheme() *TreeTheme {
	return &TreeTheme{
		RootStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#66D9EF")), // cyan
		BranchStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FD971F")), // orange
		LeafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2")), // foreground
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#75715E")). // comment
			Faint(true),
		FlagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6E22E")), // green
		LineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3E3D32")), // line
	}
}

// LightTreeTheme 返回亮色主题
func LightTreeTheme() *TreeTheme {
	return &TreeTheme{
		RootStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("26")), // blue
		BranchStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("214")), // orange
		LeafStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("16")), // black
		DescriptionStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")). // gray
			Faint(true),
		FlagStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("28")), // green
		LineStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("248")), // light gray
	}
}

// GetTreeThemeByName 根据名称获取主题
func GetTreeThemeByName(name string) *TreeTheme {
	switch name {
	case "dracula":
		return DraculaTreeTheme()
	case "nord":
		return NordTreeTheme()
	case "monokai":
		return MonokaiTreeTheme()
	case "light":
		return LightTreeTheme()
	default:
		return DefaultTreeTheme()
	}
}

// TreeDisplayNode 树形显示节点
type TreeDisplayNode struct {
	Name        string
	Description string
	Path        string
	IsRunnable  bool
	Children    []*TreeDisplayNode
	Flags       []FlagDisplayInfo
}

// FlagDisplayInfo flag 显示信息
type FlagDisplayInfo struct {
	Name         string
	ShortName    string
	Description  string
	DefaultValue string
}

// DisplayTree 显示命令树（树形结构）
func DisplayTree(root *Command, config *TreeConfig) string {
	if config == nil {
		config = &TreeConfig{
			Theme:       DefaultTreeTheme(),
			ShowFlags:   false,
			ShowLong:    false,
			IndentWidth: 4,
		}
	}

	// 构建树形结构
	tree := buildDisplayTree(root, "", 0)

	// 生成显示文本
	var builder strings.Builder
	renderTree(&builder, tree, "", config.Theme, true, 0, config)

	return builder.String()
}

// buildDisplayTree 构建显示树
func buildDisplayTree(cmd *Command, path string, depth int) *TreeDisplayNode {
	// 构建当前节点路径
	currentPath := path
	if path == "" {
		currentPath = cmd.Name()
	} else {
		currentPath = path + " " + cmd.Name()
	}

	node := &TreeDisplayNode{
		Name:        cmd.Name(),
		Description: cmd.Short,
		Path:        currentPath,
		IsRunnable:  cmd.Run != nil || cmd.RunE != nil,
		Children:    make([]*TreeDisplayNode, 0),
	}

	// 递归处理子命令
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() {
			continue
		}
		wrapped := &Command{Command: child}
		childNode := buildDisplayTree(wrapped, currentPath, depth+1)
		node.Children = append(node.Children, childNode)
	}

	return node
}

// renderTree 渲染树形结构
func renderTree(builder *strings.Builder, node *TreeDisplayNode, prefix string, theme *TreeTheme, isLast bool, depth int, config *TreeConfig) {
	// 选择连接符
	connector := "├── "
	if isLast {
		connector = "└── "
	}

	// 确定样式
	var style lipgloss.Style
	if depth == 0 {
		style = theme.RootStyle
	} else if node.IsRunnable {
		style = theme.LeafStyle
	} else {
		style = theme.BranchStyle
	}

	// 渲染节点名称
	nodeLine := prefix + connector + node.Name
	builder.WriteString(style.Render(nodeLine))
	builder.WriteString("\n")

	// 渲染描述
	if node.Description != "" && config.ShowLong {
		descPrefix := prefix
		if isLast {
			descPrefix += "    "
		} else {
			descPrefix += "│   "
		}
		descLine := descPrefix + "└─ " + node.Description
		builder.WriteString(theme.DescriptionStyle.Render(descLine))
		builder.WriteString("\n")
	}

	// 渲染子节点
	for i, child := range node.Children {
		childIsLast := i == len(node.Children)-1
		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += "│   "
		}
		renderTree(builder, child, childPrefix, theme, childIsLast, depth+1, config)
	}
}

// DisplayFlatTree 显示扁平化的命令列表
func DisplayFlatTree(root *Command, config *TreeConfig) string {
	if config == nil {
		config = &TreeConfig{
			Theme:       DefaultTreeTheme(),
			ShowFlags:   false,
			ShowLong:    true,
			IndentWidth: 2,
		}
	}

	// 获取所有命令路径
	commands := getAllCommandPaths(root)

	var builder strings.Builder

	// 标题
	title := fmt.Sprintf("Command Tree (%d commands)\n", len(commands))
	builder.WriteString(config.Theme.RootStyle.Bold(true).Render(title))
	builder.WriteString("\n")

	// 显示每个命令
	for i, cmdInfo := range commands {
		// 命令路径
		pathLine := fmt.Sprintf("%2d. %s", i+1, cmdInfo.path)
		if cmdInfo.isRunnable {
			pathLine += " ✓"
		}
		builder.WriteString(config.Theme.LeafStyle.Render(pathLine))
		builder.WriteString("\n")

		// 描述
		if cmdInfo.short != "" && config.ShowLong {
			descPrefix := strings.Repeat(" ", len(fmt.Sprintf("%2d. ", i+1)))
			descLine := descPrefix + "   " + cmdInfo.short
			builder.WriteString(config.Theme.DescriptionStyle.Render(descLine))
			builder.WriteString("\n")
		}

		// flags
		if config.ShowFlags && cmdInfo.cmd != nil {
			flags := collectFlagsForDisplay(cmdInfo.cmd)
			for _, flag := range flags {
				flagName := "--" + flag.Name
				if flag.ShortName != "" {
					flagName = "-" + flag.ShortName + ", " + flagName
				}
				flagPrefix := strings.Repeat(" ", len(fmt.Sprintf("%2d. ", i+1)))
				flagLine := flagPrefix + "   " + flagName
				builder.WriteString(config.Theme.FlagStyle.Render(flagLine))
				builder.WriteString("\n")
			}
		}
	}

	return builder.String()
}

// cmdInfo 命令信息
type cmdInfo struct {
	path       string
	short      string
	isRunnable bool
	cmd        *Command
}

// getAllCommandPaths 获取所有命令的路径
func getAllCommandPaths(root *Command) []cmdInfo {
	tree := buildDisplayTree(root, "", 0)
	var infos []cmdInfo
	collectPathsWithInfo(tree, "", &infos)
	return infos
}

// collectPathsWithInfo 收集所有路径和信息
func collectPathsWithInfo(node *TreeDisplayNode, prefix string, infos *[]cmdInfo) {
	currentPath := prefix + node.Name
	info := cmdInfo{
		path:       currentPath,
		short:      node.Description,
		isRunnable: node.IsRunnable,
	}
	*infos = append(*infos, info)

	for _, child := range node.Children {
		collectPathsWithInfo(child, currentPath+" ", infos)
	}
}

// collectFlagsForDisplay 收集 flags 用于显示
func collectFlagsForDisplay(cmd *Command) []FlagDisplayInfo {
	var flags []FlagDisplayInfo

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		if strings.HasPrefix(flag.Name, "tree-") || flag.Name == "tree" {
			return
		}

		info := FlagDisplayInfo{
			Name:         flag.Name,
			ShortName:    flag.Shorthand,
			Description:  flag.Usage,
			DefaultValue: flag.DefValue,
		}
		flags = append(flags, info)
	})

	return flags
}

// FindCommandByPathString 根据路径字符串查找命令
func FindCommandByPathString(root *Command, path string) (*Command, error) {
	parts := strings.Fields(path)
	if len(parts) == 0 {
		return root, nil
	}

	current := root.Command
	for i := 1; i < len(parts); i++ {
		found := false
		for _, cmd := range current.Commands() {
			if cmd.Name() == parts[i] {
				current = cmd
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("command not found: %s", parts[i])
		}
	}

	return &Command{Command: current}, nil
}

// GetCommandTreeString 获取命令树字符串（便捷函数）
func GetCommandTreeString(root *spf13cobra.Command, themeName string, showFlags, showLong bool) string {
	cmd := &Command{Command: root}
	config := &TreeConfig{
		Theme:     GetTreeThemeByName(themeName),
		ShowFlags: showFlags,
		ShowLong:  showLong,
	}
	return DisplayFlatTree(cmd, config)
}
