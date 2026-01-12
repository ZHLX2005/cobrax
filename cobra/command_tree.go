package cobra

import (
	"strings"

	spf13cobra "github.com/spf13/cobra"
)

// CommandNode 命令树节点
type CommandNode struct {
	ID         string
	Name       string
	Use        string
	Short      string
	Long       string
	IsRunnable bool
	Children   []*CommandNode
}

// BuildCommandTree 从 cobra 命令构建命令树结构
func BuildCommandTree(cmd *spf13cobra.Command, path string) *CommandNode {
	// 构建当前命令路径
	currentPath := path
	if path != "" {
		currentPath = path + " " + cmd.Name()
	} else {
		currentPath = cmd.Name()
	}

	// 判断命令是否可执行（有 Run 或 RunE）
	isRunnable := cmd.Run != nil || cmd.RunE != nil

	node := &CommandNode{
		ID:         cmd.Name(),
		Name:       cmd.Name(),
		Use:        cmd.Use,
		Short:      cmd.Short,
		Long:       cmd.Long,
		IsRunnable: isRunnable,
		Children:   make([]*CommandNode, 0),
	}

	// 获取可用的子命令
	children := getAvailableCommands(cmd.Commands())
	for _, child := range children {
		childNode := BuildCommandTree(child, currentPath)
		if childNode != nil {
			node.Children = append(node.Children, childNode)
		}
	}

	return node
}

// GetExecutableCommands 获取所有可执行的命令（扁平化列表）
func GetExecutableCommands(cmd *spf13cobra.Command) []*CommandNode {
	root := BuildCommandTree(cmd, "")
	// 如果根命令有子命令，则只返回子命令中的可执行命令
	if len(root.Children) > 0 {
		var result []*CommandNode
		for _, child := range root.Children {
			result = append(result, flattenExecutableCommands(child, "")...)
		}
		return result
	}
	// 如果根命令没有子命令，则返回根命令本身
	return flattenExecutableCommands(root, "")
}

// flattenExecutableCommands 递归扁平化获取所有可执行命令
func flattenExecutableCommands(item *CommandNode, path string) []*CommandNode {
	result := make([]*CommandNode, 0)

	// 构建当前路径
	currentPath := path
	if path != "" {
		currentPath = path + " " + item.Use
	} else {
		currentPath = item.Use
	}

	// 如果是可执行命令，添加到结果
	if item.IsRunnable {
		// 简化路径显示（去除根命令前缀）
		displayPath := currentPath
		parts := strings.Fields(displayPath)
		if len(parts) > 1 {
			displayPath = strings.Join(parts[1:], " ")
		}

		result = append(result, &CommandNode{
			ID:         item.ID,
			Name:       item.Name,
			Use:        displayPath,
			Short:      item.Short,
			Long:       item.Long,
			IsRunnable: true,
			Children:   nil,
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
		pathParts = append([]string{current.Name()}, pathParts...)
		current = current.Parent()
	}

	if len(pathParts) == 0 {
		return cmd.Name()
	}

	return strings.Join(pathParts, " ")
}

// getAvailableCommands 获取可用的命令
func getAvailableCommands(cmds []*spf13cobra.Command) []*spf13cobra.Command {
	var result []*spf13cobra.Command
	for _, cmd := range cmds {
		if !cmd.IsAvailableCommand() {
			continue
		}
		if cmd.Hidden {
			continue
		}
		// 过滤掉 completion 命令
		if isCompletionCommand(cmd) {
			continue
		}
		// 过滤掉 help 命令
		if cmd.Name() == "help" {
			continue
		}
		result = append(result, cmd)
	}
	return result
}

// isCompletionCommand 检查是否是 completion 相关命令
func isCompletionCommand(cmd *spf13cobra.Command) bool {
	name := cmd.Name()
	completionShells := []string{"bash", "fish", "powershell", "zsh", "pwsh"}
	for _, shell := range completionShells {
		if name == shell {
			return true
		}
	}
	if name == "completion" {
		return true
	}
	if cmd.Annotations != nil {
		if cmd.Annotations["command"] == "completion" {
			return true
		}
	}
	return false
}
