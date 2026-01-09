package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ZHLX2005/cobrax/tui/style"
)

// DefaultRenderer 默认 TUI 渲染器
// 使用 bubbletea 实现交互式终端界面
type DefaultRenderer struct {
	theme    *style.Theme
	programs []*tea.Program
}

// NewDefaultRenderer 创建默认渲染器
func NewDefaultRenderer(theme *style.Theme) *DefaultRenderer {
	if theme == nil {
		theme = style.DefaultTheme()
	}

	return &DefaultRenderer{
		theme:    theme,
		programs: make([]*tea.Program, 0),
	}
}

// RenderCommandMenu 渲染命令菜单面板
func (r *DefaultRenderer) RenderCommandMenu(menuTitle string, options []MenuItem) (selectedIndex int, err error) {
	// 获取终端尺寸
	width, height := getTerminalSize()

	// 创建菜单模型
	model := newMenuModel(options, r.theme, width, height)

	// 创建并运行程序
	p := tea.NewProgram(model, tea.WithAltScreen())
	r.programs = append(r.programs, p)

	result, err := p.Run()
	if err != nil {
		return -1, fmt.Errorf("failed to run menu: %w", err)
	}

	menuResult, ok := result.(*menuModel)
	if !ok {
		return -1, fmt.Errorf("unexpected result type")
	}

	if menuResult.cancelled {
		return -1, nil
	}

	return menuResult.cursor, nil
}

// RenderFlagForm 渲染 flag 输入表单
func (r *DefaultRenderer) RenderFlagForm(formTitle string, flags []FlagItem) (map[string]string, error) {
	if len(flags) == 0 {
		return nil, nil
	}

	// 获取终端尺寸
	width, height := getTerminalSize()

	// 创建表单模型
	model := newFormModel(flags, r.theme, width, height)

	// 创建并运行程序
	p := tea.NewProgram(model, tea.WithAltScreen())
	r.programs = append(r.programs, p)

	result, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("failed to run form: %w", err)
	}

	formResult, ok := result.(*formModel)
	if !ok {
		return nil, fmt.Errorf("unexpected result type")
	}

	if formResult.cancelled {
		return nil, nil
	}

	return formResult.getValues(), nil
}

// RenderConfirmation 渲染确认面板
func (r *DefaultRenderer) RenderConfirmation(title, message string) (bool, error) {
	// 获取终端尺寸
	width, height := getTerminalSize()

	// 创建确认模型
	model := newConfirmModel(title, message, r.theme, width, height)

	// 创建并运行程序
	p := tea.NewProgram(model, tea.WithAltScreen())
	r.programs = append(r.programs, p)

	result, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("failed to run confirmation: %w", err)
	}

	confirmResult, ok := result.(*confirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected result type")
	}

	return confirmResult.confirmed, nil
}

// RenderHelp 渲染帮助面板
func (r *DefaultRenderer) RenderHelp(title, content string) error {
	// TODO: 实现帮助面板
	return nil
}

// Cleanup 清理资源
func (r *DefaultRenderer) Cleanup() error {
	// bubbletea 会自动清理，这里添加未来可能需要的清理逻辑
	return nil
}

// getTerminalSize 获取终端尺寸
func getTerminalSize() (int, int) {
	// 简单实现：使用默认尺寸
	// 实际应用中可以使用更精确的方法
	return 80, 24
}

// ============================================================================
// 菜单模型
// ============================================================================

// menuModel 菜单模型
type menuModel struct {
	items           []MenuItem
	cursor          int
	cancelled       bool
	theme           *style.Theme
	width           int
	height          int
	quitting        bool
	showDescription bool
}

// newMenuModel 创建菜单模型
func newMenuModel(items []MenuItem, theme *style.Theme, width, height int) *menuModel {
	return &menuModel{
		items:           items,
		cursor:          0,
		theme:           theme,
		width:           width,
		height:          height,
		showDescription: true,
	}
}

// Init 初始化
func (m *menuModel) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (m *menuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
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
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}

		case "enter", " ":
			item := m.items[m.cursor]
			if !item.Disabled {
				m.quitting = true
				return m, tea.Quit
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View 渲染视图
func (m *menuModel) View() string {
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
	title := titleStyle.Render("Select a command:")

	// 构建菜单项
	var items strings.Builder
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = "▶"
		}

		label := item.Label
		if label == "" {
			label = item.ID
		}

		text := fmt.Sprintf("%s %s", cursor, label)

		if i == m.cursor {
			text = m.theme.Styles.SelectedStyle.Render(text)
			if item.Description != "" && m.showDescription {
				text += "\n   " + m.theme.Styles.HelpStyle.Render(item.Description)
			}
		} else if item.Disabled {
			text = m.theme.Styles.DisabledStyle.Render(text)
		}

		items.WriteString(text + "\n")
	}

	// 构建帮助文本
	helpText := m.theme.Styles.HelpStyle.Render("\n[↑↓ Navigate] [Enter Select] [Esc/Quit]")

	// 组合内容
	content := title + "\n" + items.String() + helpText

	return borderStyle.Render(content)
}

// ============================================================================
// 表单模型
// ============================================================================

// formModel 表单模型
type formModel struct {
	items        []FlagItem
	cursor       int
	values       map[string]string
	cancelled    bool
	theme        *style.Theme
	width        int
	height       int
	quitting     bool
	editMode     bool
	editBuffer   string
}

// newFormModel 创建表单模型
func newFormModel(items []FlagItem, theme *style.Theme, width, height int) *formModel {
	values := make(map[string]string)
	for _, item := range items {
		values[item.Name] = item.DefaultValue
	}

	return &formModel{
		items:  items,
		cursor: 0,
		values: values,
		theme:  theme,
		width:  width,
		height: height,
	}
}

// Init 初始化
func (m *formModel) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (m *formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.editMode {
			return m.handleEditKey(msg)
		}
		return m.handleNavKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// handleNavKey 处理导航按键
func (m *formModel) handleNavKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q", "esc":
		m.quitting = true
		m.cancelled = true
		return m, tea.Quit

	case "up", "shift+tab":
		if m.cursor > 0 {
			m.cursor--
		}

	case "down", "tab":
		if m.cursor < len(m.items)-1 {
			m.cursor++
		}

	case "enter", " ":
		m.quitting = true
		return m, tea.Quit

	case "left", "right":
		// 切换 bool 值
		item := m.items[m.cursor]
		if item.Type == FlagTypeBool {
			// 切换 true/false
			if m.values[item.Name] == "true" {
				m.values[item.Name] = "false"
			} else {
				m.values[item.Name] = "true"
			}
		}

	case "e", "r":
		// 进入编辑模式
		item := m.items[m.cursor]
		if item.Type != FlagTypeBool {
			m.editMode = true
			m.editBuffer = m.values[item.Name]
		}
	}

	return m, nil
}

// handleEditKey 处理编辑按键
func (m *formModel) handleEditKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	item := m.items[m.cursor]

	switch msg.String() {
	case "enter":
		// 保存并退出编辑模式
		m.values[item.Name] = m.editBuffer
		m.editMode = false
		return m, nil

	case "esc":
		// 取消编辑
		m.editMode = false
		return m, nil

	case "backspace":
		if len(m.editBuffer) > 0 {
			m.editBuffer = m.editBuffer[:len(m.editBuffer)-1]
		}

	default:
		// 添加字符
		if len(msg.String()) == 1 {
			m.editBuffer += msg.String()
		}
	}

	return m, nil
}

// View 渲染视图
func (m *formModel) View() string {
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
	title := titleStyle.Render("Configure flags:")

	// 构建表单项
	var items strings.Builder
	for i, item := range m.items {
		cursor := " "
		if i == m.cursor {
			cursor = "▶"
		}

		var valueDisplay string
		if m.editMode && i == m.cursor {
			valueDisplay = m.editBuffer + "_"
		} else {
			valueDisplay = m.values[item.Name]
		}

		text := fmt.Sprintf("%s %s: [%s]", cursor, item.Name, valueDisplay)

		if i == m.cursor {
			text = m.theme.Styles.SelectedStyle.Render(text)
			if item.Description != "" {
				text += "\n   " + m.theme.Styles.HelpStyle.Render(item.Description)
			}
			// 对于 bool 类型，显示特殊提示
			if item.Type == FlagTypeBool {
				text += "\n   " + m.theme.Styles.HelpStyle.Render("← → Toggle value")
			}
		}

		items.WriteString(text + "\n")
	}

	// 构建帮助文本 - 动态根据当前 flag 类型显示不同的帮助
	currentItem := m.items[m.cursor]
	helpText := "[↑↓/Tab Navigate] [Enter/Space Save&Quit] "
	if currentItem.Type == FlagTypeBool {
		helpText += "[← → Toggle] "
	} else {
		helpText += "[E Edit] "
	}
	helpText += "[Esc Cancel]"
	helpText = m.theme.Styles.HelpStyle.Render("\n" + helpText)

	// 组合内容
	content := title + "\n" + items.String() + helpText

	return borderStyle.Render(content)
}

// getValues 获取所有值
func (m *formModel) getValues() map[string]string {
	return m.values
}

// ============================================================================
// 确认模型
// ============================================================================

// confirmModel 确认模型
type confirmModel struct {
	title     string
	message   string
	confirmed bool
	cancelled bool
	theme     *style.Theme
	width     int
	height    int
	cursor    int // 0 = yes, 1 = no
	quitting  bool
}

// newConfirmModel 创建确认模型
func newConfirmModel(title, message string, theme *style.Theme, width, height int) *confirmModel {
	return &confirmModel{
		title:     title,
		message:   message,
		theme:     theme,
		width:     width,
		height:    height,
		cursor:    0,
		confirmed: false,
	}
}

// Init 初始化
func (m *confirmModel) Init() tea.Cmd {
	return nil
}

// Update 更新状态
func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			m.cancelled = true
			return m, tea.Quit

		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}

		case "right", "l":
			if m.cursor < 1 {
				m.cursor++
			}

		case "enter", " ":
			m.quitting = true
			m.confirmed = m.cursor == 0
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View 渲染视图
func (m *confirmModel) View() string {
	if m.quitting {
		return ""
	}

	// 构建样式
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(m.theme.Colors.Primary).
		MarginTop(1).
		MarginBottom(1)

	messageStyle := lipgloss.NewStyle().
		MarginBottom(2)

	borderStyle := lipgloss.NewStyle().
		Border(m.theme.Styles.Border).
		BorderForeground(m.theme.Colors.Secondary).
		Padding(m.theme.Layout.Padding[0], m.theme.Layout.Padding[1]).
		Width(m.width - 4)

	// 按钮样式
	yesStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Success)
	noStyle := lipgloss.NewStyle().Foreground(m.theme.Colors.Error)

	if m.cursor == 0 {
		yesStyle = yesStyle.Bold(true).Reverse(true)
	} else {
		noStyle = noStyle.Bold(true).Reverse(true)
	}

	// 构建内容
	title := titleStyle.Render(m.title)
	message := messageStyle.Render(m.message)
	buttons := fmt.Sprintf("%s  %s",
		yesStyle.Render("[Yes]"),
		noStyle.Render("[No]"),
	)

	content := title + "\n\n" + message + "\n\n" + buttons

	return borderStyle.Render(content)
}
