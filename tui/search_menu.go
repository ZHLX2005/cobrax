package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ZHLX2005/cobrax/tui/style"
)

// SearchMenuModel 带搜索功能的菜单模型
type SearchMenuModel struct {
	items           []MenuItem      // 所有菜单项
	filteredItems   []MenuItem      // 过滤后的菜单项
	cursor          int             // 当前光标位置
	cancelled       bool            // 是否取消
	quitting        bool            // 是否正在退出
	theme           *style.Theme    // 主题
	width           int             // 终端宽度
	height          int             // 终端高度
	showDescription bool            // 是否显示描述

	// 搜索相关
	searchMode      bool            // 是否处于搜索模式
	searchQuery     string          // 搜索查询字符串
	searchCursor    int             // 搜索输入光标位置
}

// NewSearchMenuModel 创建搜索菜单模型
func NewSearchMenuModel(items []MenuItem, theme *style.Theme, width, height int) *SearchMenuModel {
	return &SearchMenuModel{
		items:           items,
		filteredItems:   items,
		cursor:          0,
		searchQuery:     "",
		searchCursor:    0,
		theme:           theme,
		width:           width,
		height:          height,
		showDescription: true,
		searchMode:      false,
	}
}

// Init 初始化
func (m *SearchMenuModel) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (m *SearchMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		return m.handleNavKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleSearchKey 处理搜索模式的按键
func (m *SearchMenuModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		m.cancelled = true
		return m, tea.Quit

	case "esc":
		// 退出搜索模式
		m.searchMode = false
		m.searchQuery = ""
		m.searchCursor = 0
		m.filteredItems = m.items
		m.cursor = 0

	case "enter":
		// 选择当前项并退出
		if len(m.filteredItems) > 0 {
			m.quitting = true
			return m, tea.Quit
		}

	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.searchCursor--
			if m.searchCursor < 0 {
				m.searchCursor = 0
			}
			m.filterItems()
		}

	case "ctrl+u":
		// 清空搜索
		m.searchQuery = ""
		m.searchCursor = 0
		m.filterItems()

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filteredItems)-1 {
			m.cursor++
		}

	default:
		// 添加搜索字符
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
			m.searchCursor++
			m.filterItems()
		}
	}

	return m, nil
}

// handleNavKey 处理导航模式的按键
func (m *SearchMenuModel) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		m.cancelled = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "j":
		if m.cursor < len(m.filteredItems)-1 {
			m.cursor++
		}

	case "enter", " ":
		if len(m.filteredItems) > 0 {
			item := m.filteredItems[m.cursor]
			if !item.Disabled {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case "/", "ctrl+s":
		// 进入搜索模式
		m.searchMode = true

	case "ctrl+r":
		// 清空过滤
		m.searchQuery = ""
		m.filteredItems = m.items
		m.cursor = 0
	}

	return m, nil
}

// filterItems 根据搜索查询过滤菜单项
func (m *SearchMenuModel) filterItems() {
	if m.searchQuery == "" {
		m.filteredItems = m.items
		m.cursor = 0
		return
	}

	query := strings.ToLower(m.searchQuery)
	m.filteredItems = make([]MenuItem, 0)

	for _, item := range m.items {
		// 匹配标签
		if strings.Contains(strings.ToLower(item.Label), query) {
			m.filteredItems = append(m.filteredItems, item)
			continue
		}

		// 匹配描述
		if strings.Contains(strings.ToLower(item.Description), query) {
			m.filteredItems = append(m.filteredItems, item)
			continue
		}

		// 匹配 ID
		if strings.Contains(strings.ToLower(item.ID), query) {
			m.filteredItems = append(m.filteredItems, item)
		}
	}

	m.cursor = 0
}

// View 渲染视图
func (m *SearchMenuModel) View() string {
	if m.quitting {
		return ""
	}

	// 构建样式
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Colors.Primary).
		MarginTop(1).
		MarginBottom(1)

	borderStyle := lipgloss.NewStyle().
		Border(m.theme.Styles.Border).
		BorderForeground(m.theme.Colors.Secondary).
		Padding(m.theme.Layout.Padding[0], m.theme.Layout.Padding[1]).
		Width(m.width - 4)

	// 构建标题
	titleText := "Select a command"
	if m.searchMode {
		titleText = "Search commands"
	}
	title := titleStyle.Render(titleText)

	// 构建搜索输入框
	var searchInput string
	if m.searchMode {
		searchStyle := lipgloss.NewStyle().
			Foreground(m.theme.Colors.Primary).
			Background(m.theme.Colors.Muted).
			Padding(0, 1)

		prompt := "/"
		if len(m.searchQuery) > 0 {
			prompt = m.searchQuery
		}
		searchInput = searchStyle.Render(prompt + "_")
	} else if m.searchQuery != "" {
		// 显示当前过滤器
		filterStyle := lipgloss.NewStyle().
			Foreground(m.theme.Colors.Success).
			MarginBottom(1)

		searchInput = filterStyle.Render("Filter: " + m.searchQuery + " [Ctrl+R to clear]")
	}

	// 构建菜单项
	var items strings.Builder
	if len(m.filteredItems) == 0 {
		noResultsStyle := lipgloss.NewStyle().
			Foreground(m.theme.Colors.Muted).
			MarginTop(1)

		items.WriteString(noResultsStyle.Render("No matching commands found"))
	} else {
		for i, item := range m.filteredItems {
			cursor := " "
			if i == m.cursor {
				cursor = "▶"
			}

			label := item.Label
			if label == "" {
				label = item.ID
			}

			// 高亮匹配的搜索词
			if m.searchQuery != "" {
				label = m.highlightMatch(label, m.searchQuery)
			}

			text := fmt.Sprintf("%s %s", cursor, label)

			if i == m.cursor {
				text = m.theme.Styles.SelectedStyle.Render(text)
				if item.Description != "" && m.showDescription {
					desc := item.Description
					if m.searchQuery != "" {
						desc = m.highlightMatch(desc, m.searchQuery)
					}
					text += "\n   " + m.theme.Styles.HelpStyle.Render(desc)
				}
			} else if item.Disabled {
				text = m.theme.Styles.DisabledStyle.Render(text)
			}

			items.WriteString(text + "\n")
		}
	}

	// 构建帮助文本
	helpText := m.buildHelpText()

	// 组合内容
	content := title
	if searchInput != "" {
		content += "\n" + searchInput
	}
	content += "\n\n" + items.String() + helpText

	return borderStyle.Render(content)
}

// highlightMatch 高亮匹配的文本
func (m *SearchMenuModel) highlightMatch(text, query string) string {
	if query == "" {
		return text
	}

	index := strings.Index(strings.ToLower(text), strings.ToLower(query))
	if index == -1 {
		return text
	}

	highlightStyle := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Warning).
		Bold(true)

	before := text[:index]
	match := text[index : index+len(query)]
	after := text[index+len(query):]

	return before + highlightStyle.Render(match) + after
}

// buildHelpText 构建帮助文本
func (m *SearchMenuModel) buildHelpText() string {
	if m.searchMode {
		return m.theme.Styles.HelpStyle.Render("\n[Type to search] [Enter Select] [Esc Exit search] [Ctrl+U Clear]")
	}

	help := "\n[↑↓ Navigate] [Enter Select] [/ Search]"
	if m.searchQuery != "" {
		help += " [Ctrl+R Clear filter]"
	} else {
		help += " [Esc Quit]"
	}

	return m.theme.Styles.HelpStyle.Render(help)
}

// GetCursor 获取当前选中的索引
func (m *SearchMenuModel) GetCursor() int {
	return m.cursor
}

// IsCancelled 是否取消
func (m *SearchMenuModel) IsCancelled() bool {
	return m.cancelled
}

// GetFilteredItems 获取过滤后的菜单项
func (m *SearchMenuModel) GetFilteredItems() []MenuItem {
	return m.filteredItems
}
