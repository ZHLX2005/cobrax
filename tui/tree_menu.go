package tui

import (
	"strings"
)

// TreeMenuItem 树形菜单项
type TreeMenuItem struct {
	MenuItem
	Level     int              // 层级深度（0为根）
	Path      string           // 完整路径
	Children  []*TreeMenuItem  // 子项
	Expanded  bool             // 是否展开（用于树形视图）
	IsLeaf    bool             // 是否为叶子节点（可执行命令）
}

// TreeMenuData 树形菜单数据
type TreeMenuData struct {
	Items    []*TreeMenuItem  // 所有菜单项（扁平化）
	FlatMode bool             // 是否使用扁平模式
}

// BuildTreeMenu 从命令列表构建树形菜单
func BuildTreeMenu(commands []*CommandItem) *TreeMenuData {
	data := &TreeMenuData{
		Items:    make([]*TreeMenuItem, 0),
		FlatMode: true, // 默认使用扁平模式，显示所有可执行命令
	}

	// 构建树形结构
	root := buildTree(commands, 0)

	// 扁平化处理，收集所有叶子节点
	if root != nil {
		data.Items = flattenTree(root, 0, "")
	}

	return data
}

// CommandItem 命令项（从外部传入）
type CommandItem struct {
	ID         string
	Name       string
	Use        string
	Short      string
	Long       string
	Children   []*CommandItem
	IsRunnable bool // 是否可执行（有 Run 或 RunE）
}

// buildTree 递归构建树形结构
func buildTree(items []*CommandItem, level int) *TreeMenuItem {
	if len(items) == 0 {
		return nil
	}

	root := &TreeMenuItem{
		Level:    level,
		Children: make([]*TreeMenuItem, 0, len(items)),
	}

	for _, item := range items {
		node := &TreeMenuItem{
			MenuItem: MenuItem{
				ID:          item.ID,
				Label:       item.Name,
				Description: item.Short,
			},
			Level:    level,
			Children: make([]*TreeMenuItem, 0),
			IsLeaf:   item.IsRunnable || len(item.Children) == 0,
		}

		// 递归处理子节点
		if len(item.Children) > 0 {
			for _, child := range item.Children {
				childNode := buildTree([]*CommandItem{child}, level+1)
				if childNode != nil {
					node.Children = append(node.Children, childNode)
				}
			}
		}

		root.Children = append(root.Children, node)
	}

	return root
}

// flattenTree 将树形结构扁平化，只收集可执行节点
func flattenTree(node *TreeMenuItem, level int, path string) []*TreeMenuItem {
	if node == nil {
		return nil
	}

	result := make([]*TreeMenuItem, 0)

	// 构建当前节点路径
	currentPath := path
	if node.Label != "" {
		if currentPath != "" {
			currentPath += " " + node.Label
		} else {
			currentPath = node.Label
		}
	}

	// 如果是叶子节点（可执行命令），添加到结果中
	if node.IsLeaf && node.Label != "" {
		item := &TreeMenuItem{
			MenuItem: MenuItem{
				ID:          node.ID,
				Label:       node.Label,
				Description: node.Description,
			},
			Level:    level,
			Path:     currentPath,
			IsLeaf:   true,
		}
		result = append(result, item)
	}

	// 递归处理子节点
	for _, child := range node.Children {
		childItems := flattenTree(child, level+1, currentPath)
		result = append(result, childItems...)
	}

	return result
}

// FilterTreeMenu 过滤树形菜单
func FilterTreeMenu(items []*TreeMenuItem, query string) []*TreeMenuItem {
	if query == "" {
		return items
	}

	query = strings.ToLower(query)
	result := make([]*TreeMenuItem, 0)

	for _, item := range items {
		// 匹配命令名称
		if strings.Contains(strings.ToLower(item.Label), query) {
			result = append(result, item)
			continue
		}

		// 匹配描述
		if strings.Contains(strings.ToLower(item.Description), query) {
			result = append(result, item)
			continue
		}

		// 匹配路径
		if strings.Contains(strings.ToLower(item.Path), query) {
			result = append(result, item)
		}
	}

	return result
}

// GetTreeMenuDisplay 获取树形菜单的显示文本（带缩进和图标）
func GetTreeMenuDisplay(items []*TreeMenuItem, selectedIndex int) []string {
	lines := make([]string, 0, len(items))

	for i, item := range items {
		// 选择指示器
		cursor := " "
		if i == selectedIndex {
			cursor = "▶"
		}

		// 缩进
		indent := strings.Repeat("  ", item.Level)

		// 图标
		icon := "📄"
		if strings.Contains(item.Path, " ") {
			icon = "📁"
		}

		// 构建显示文本
		line := cursor + " " + indent + icon + " " + item.Label

		// 如果有描述，添加到下一行
		if item.Description != "" && i == selectedIndex {
			line += "\n" + indent + "   └─ " + item.Description
		}

		lines = append(lines, line)
	}

	return lines
}
