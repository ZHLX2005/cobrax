package cobra

import (
	"strings"

	spf13cobra "github.com/spf13/cobra"
	"github.com/ZHLX2005/cobrax/tui"
)

// BuildCommandTree 从 cobra 命令构建命令树结构
func BuildCommandTree(cmd *spf13cobra.Command, path string) *tui.CommandItem {
	// 构建当前命令路径
	currentPath := path
	if path != "" {
		currentPath = path + " " + cmd.Name()
	} else {
		currentPath = cmd.Name()
	}

	// 判断命令是否可执行（有 Run 或 RunE）
	isRunnable := cmd.Run != nil || cmd.RunE != nil

	item := &tui.CommandItem{
		ID:         cmd.Name(),
		Name:       cmd.Name(),
		Use:        cmd.Use,
		Short:      cmd.Short,
		Long:       cmd.Long,
		IsRunnable: isRunnable,
		Children:   make([]*tui.CommandItem, 0),
	}

	// 获取可用的子命令
	children := getAvailableCommands(cmd.Commands())
	for _, child := range children {
		childItem := BuildCommandTree(child, currentPath)
		if childItem != nil {
			item.Children = append(item.Children, childItem)
		}
	}

	return item
}

// GetExecutableCommands 获取所有可执行的命令（扁平化列表）
func GetExecutableCommands(cmd *spf13cobra.Command) []*tui.CommandItem {
	root := BuildCommandTree(cmd, "")
	return flattenExecutableCommands(root, "")
}

// flattenExecutableCommands 递归扁平化获取所有可执行命令
func flattenExecutableCommands(item *tui.CommandItem, path string) []*tui.CommandItem {
	result := make([]*tui.CommandItem, 0)

	// 构建当前路径
	currentPath := path
	if path != "" {
		currentPath = path + " " + item.Name
	} else {
		currentPath = item.Name
	}

	// 如果是可执行命令，添加到结果
	if item.IsRunnable {
		result = append(result, &tui.CommandItem{
			ID:         item.ID,
			Name:       item.Name,
			Use:        item.Use,
			Short:      item.Short,
			Long:       item.Long,
			IsRunnable: true,
			Children:   nil, // 扁平化后不需要子节点
		})
	}

	// 递归处理子节点
	for _, child := range item.Children {
		childCommands := flattenExecutableCommands(child, currentPath)
		result = append(result, childCommands...)
	}

	return result
}

// FindCommandByPath 根据路径查找命令
// 例如: "gocmds note" 会查找 note 子命令
func FindCommandByPath(root *spf13cobra.Command, path string) *spf13cobra.Command {
	if path == "" {
		return root
	}

	parts := strings.Fields(path)
	current := root

	for _, part := range parts {
		// 查找子命令
		found := false
		for _, cmd := range current.Commands() {
			if cmd.Name() == part && !cmd.Hidden {
				current = cmd
				found = true
				break
			}
		}

		if !found {
			return nil
		}
	}

	return current
}

// GetCommandFullPath 获取命令的完整路径
func GetCommandFullPath(cmd *spf13cobra.Command) string {
	if cmd == nil {
		return ""
	}

	// 递归获取父级路径
	var pathParts []string
	current := cmd

	for current != nil {
		if current.Parent() != nil {
			pathParts = append([]string{current.Name()}, pathParts...)
		}
		current = current.Parent()
	}

	if len(pathParts) == 0 {
		return cmd.Name()
	}

	return strings.Join(pathParts, " ")
}
