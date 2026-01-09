# 用户体验优化方案

## 概述

用户体验（UX）是 CLI/TUI 工具成功的关键因素。Cobra-X 已经提供了良好的基础体验，但仍有改进空间。本文档从交互设计、视觉反馈、错误处理和可访问性等方面提出优化方案。

## 当前体验分析

### 现有优势

1. **清晰的导航**: 支持键盘快捷键和方向键
2. **视觉反馈**: 选中项高亮显示
3. **帮助信息**: 实时显示命令描述和帮助文本
4. **确认机制**: 执行前确认命令

### 可改进点

1. **搜索体验**: 搜索功能可以更智能
2. **错误处理**: 错误信息可以更友好
3. **进度反馈**: 长时间操作缺少进度提示
4. **键盘导航**: 某些快捷键不够直观
5. **主题支持**: 主题切换不够便捷

## 优化方案

### 方案 1: 增强搜索体验

#### 模糊搜索实现

```go
// 增强的搜索功能
type EnhancedSearchModel struct {
    items           []MenuItem
    filteredItems   []MenuItem
    cursor          int
    searchQuery     string
    searchMode      bool
    highlightRanges []HighlightRange
}

type HighlightRange struct {
    Start int
    End   int
}

func (m *EnhancedSearchModel) filterItems() {
    if m.searchQuery == "" {
        m.filteredItems = m.items
        m.highlightRanges = nil
        return
    }

    m.filteredItems = make([]MenuItem, 0)
    query := strings.ToLower(m.searchQuery)

    for _, item := range m.items {
        if m.fuzzyMatch(item.Label, query) || m.fuzzyMatch(item.Description, query) {
            m.filteredItems = append(m.filteredItems, item)
        }
    }

    // 计算高亮范围
    m.calculateHighlights()
}

// 模糊匹配算法
func (m *EnhancedSearchModel) fuzzyMatch(text, query string) bool {
    text = strings.ToLower(text)

    // 精确匹配
    if strings.Contains(text, query) {
        return true
    }

    // 首字母匹配
    if m.matchInitials(text, query) {
        return true
    }

    // 连续字符匹配
    if m.matchSubsequence(text, query) {
        return true
    }

    return false
}

// 首字母匹配 (如 "gs" 匹配 "git status")
func (m *EnhancedSearchModel) matchInitials(text, query string) bool {
    words := strings.Fields(text)
    initials := ""
    for _, word := range words {
        if len(word) > 0 {
            initials += string(word[0])
        }
    }

    return strings.Contains(initials, query)
}

// 子序列匹配 (如 "gts" 匹配 "git status")
func (m *EnhancedSearchModel) matchSubsequence(text, query string) bool {
    textRunes := []rune(text)
    queryRunes := []rune(query)

    textIdx := 0
    queryIdx := 0

    for textIdx < len(textRunes) && queryIdx < len(queryRunes) {
        if textRunes[textIdx] == queryRunes[queryIdx] {
            queryIdx++
        }
        textIdx++
    }

    return queryIdx == len(queryRunes)
}

// 计算高亮范围
func (m *EnhancedSearchModel) calculateHighlights() {
    for i, item := range m.filteredItems {
        m.highlightRanges = m.findHighlightRanges(item.Label, m.searchQuery)
        // 也可以在描述中高亮
    }
}

func (m *EnhancedSearchModel) findHighlightRanges(text, query string) []HighlightRange {
    var ranges []HighlightRange
    lowerText := strings.ToLower(text)
    lowerQuery := strings.ToLower(query)

    start := 0
    for {
        idx := strings.Index(lowerText[start:], lowerQuery)
        if idx == -1 {
            break
        }

        ranges = append(ranges, HighlightRange{
            Start: start + idx,
            End:   start + idx + len(query),
        })

        start += idx + len(query)
    }

    return ranges
}

// 渲染高亮文本
func (m *EnhancedSearchModel) renderWithHighlight(text string, theme *style.Theme) string {
    if len(m.highlightRanges) == 0 {
        return text
    }

    runes := []rune(text)
    var result strings.Builder
    lastEnd := 0

    for _, rng := range m.highlightRanges {
        // 添加高亮前的文本
        result.WriteString(string(runes[lastEnd:rng.Start]))

        // 添加高亮文本
        highlighted := string(runes[rng.Start:rng.End])
        result.WriteString(theme.Styles.SelectedStyle.Render(highlighted))

        lastEnd = rng.End
    }

    // 添加剩余文本
    result.WriteString(string(runes[lastEnd:]))

    return result.String()
}
```

#### 搜索历史

```go
// 带历史的搜索模型
type SearchWithHistoryModel struct {
    *EnhancedSearchModel
    history       []string
    historyIndex  int
    maxHistory    int
}

func NewSearchWithHistoryModel(items []MenuItem, maxHistory int) *SearchWithHistoryModel {
    return &SearchWithHistoryModel{
        EnhancedSearchModel: &EnhancedSearchModel{
            items:         items,
            filteredItems: items,
        },
        maxHistory: maxHistory,
    }
}

func (m *SearchWithHistoryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.searchMode {
            return m.handleSearchKey(msg)
        }

    // ... 其他消息处理
    }

    return m.EnhancedSearchModel.Update(msg)
}

func (m *SearchWithHistoryModel) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "up":
        // 导航搜索历史
        if m.historyIndex > 0 {
            m.historyIndex--
            m.searchQuery = m.history[m.historyIndex]
            m.filterItems()
        }
        return m, nil

    case "down":
        // 导航搜索历史
        if m.historyIndex < len(m.history)-1 {
            m.historyIndex++
            m.searchQuery = m.history[m.historyIndex]
        } else {
            m.historyIndex = len(m.history)
            m.searchQuery = ""
        }
        m.filterItems()
        return m, nil

    case "enter":
        // 保存到历史
        if m.searchQuery != "" {
            m.history = append(m.history, m.searchQuery)
            if len(m.history) > m.maxHistory {
                m.history = m.history[1:]
            }
        }
        m.searchMode = false
        return m, tea.Quit

    // ... 其他按键
    }

    return m.EnhancedSearchModel.Update(msg)
}
```

### 方案 2: 智能错误处理

#### 友好的错误信息

```go
// 增强的错误处理
type ErrorHandler struct {
    theme *style.Theme
}

func NewErrorHandler(theme *style.Theme) *ErrorHandler {
    return &ErrorHandler{theme: theme}
}

func (h *ErrorHandler) HandleError(err error) string {
    if err == nil {
        return ""
    }

    // 解析错误类型
    switch e := err.(type) {
    case *FlagError:
        return h.renderFlagError(e)

    case *ValidationError:
        return h.renderValidationError(e)

    case *CommandError:
        return h.renderCommandError(e)

    default:
        return h.renderGenericError(err)
    }
}

func (h *ErrorHandler) renderFlagError(err *FlagError) string {
    var builder strings.Builder

    // 错误标题
    titleStyle := h.theme.Styles.ErrorStyle
    builder.WriteString(titleStyle.Render("❌ Configuration Error\n\n"))

    // 错误详情
    detailStyle := h.theme.Styles.ItemStyle
    builder.WriteString(detailStyle.Render(fmt.Sprintf("Flag: --%s\n", err.FlagName)))

    // 错误原因
    reasonStyle := h.theme.Styles.HelpStyle
    builder.WriteString(reasonStyle.Render(fmt.Sprintf("\nReason: %s\n", err.Reason)))

    // 建议
    if err.Suggestion != "" {
        suggestStyle := h.theme.Styles.SelectedStyle
        builder.WriteString(suggestStyle.Render(fmt.Sprintf("\n💡 Suggestion: %s\n", err.Suggestion)))
    }

    return builder.String()
}

func (h *ErrorHandler) renderValidationError(err *ValidationError) string {
    var builder strings.Builder

    titleStyle := h.theme.Styles.ErrorStyle
    builder.WriteString(titleStyle.Render("⚠️  Validation Error\n\n"))

    detailStyle := h.theme.Styles.ItemStyle
    builder.WriteString(detailStyle.Render(fmt.Sprintf("Field: %s\n", err.Field)))
    builder.WriteString(detailStyle.Render(fmt.Sprintf("Value: %s\n", err.Value)))

    reasonStyle := h.theme.Styles.HelpStyle
    builder.WriteString(reasonStyle.Render(fmt.Sprintf("\n%s\n", err.Message)))

    // 如果有有效值，显示它们
    if len(err.ValidValues) > 0 {
        validStyle := h.theme.Styles.SelectedStyle
        builder.WriteString(validStyle.Render("\nValid values:\n"))
        for _, v := range err.ValidValues {
            builder.WriteString(validStyle.Render(fmt.Sprintf("  • %s\n", v)))
        }
    }

    return builder.String()
}

// 自定义错误类型
type FlagError struct {
    FlagName   string
    Reason     string
    Suggestion string
}

func (e *FlagError) Error() string {
    return fmt.Sprintf("flag %s: %s", e.FlagName, e.Reason)
}

type ValidationError struct {
    Field       string
    Value       string
    Message     string
    ValidValues []string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}
```

#### 错误恢复

```go
// 带恢复的表单模型
type RecoverableFormModel struct {
    *formModel
    errors        map[string]error
    recoveryMode  bool
    recoveryField string
}

func (m *RecoverableFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if m.recoveryMode {
            return m.handleRecoveryKey(msg)
        }
        return m.formModel.Update(msg)

    default:
        return m.formModel.Update(msg)
    }
}

func (m *RecoverableFormModel) handleRecoveryKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    case "esc":
        // 取消恢复，使用默认值
        m.values[m.recoveryField] = m.getDefaultValue(m.recoveryField)
        delete(m.errors, m.recoveryField)
        m.recoveryMode = false
        return m, nil

    case "enter":
        // 尝试恢复
        if err := m.validateValue(m.recoveryField, m.values[m.recoveryField]); err != nil {
            // 仍然无效，保持恢复模式
            return m, nil
        }

        delete(m.errors, m.recoveryField)
        m.recoveryMode = false
        return m, nil

    default:
        // 输入新值
        return m.formModel.Update(msg)
    }
}

func (m *RecoverableFormModel) validateValue(field, value string) error {
    // 验证逻辑
    return nil
}

func (m *RecoverableFormModel) getDefaultValue(field string) string {
    // 返回默认值
    for _, item := range m.items {
        if item.Name == field {
            return item.DefaultValue
        }
    }
    return ""
}
```

### 方案 3: 进度反馈

#### 长时间操作的进度显示

```go
// 带进度的操作
type ProgressModel struct {
    title       string
    current     int
    total       int
    status      string
    cancel      bool
    theme       *style.Theme
    width       int
    height      int
    spinner     []string
    spinnerIdx  int
    tickChan    chan time.Ticker
}

func NewProgressModel(title string, total int, theme *style.Theme) *ProgressModel {
    return &ProgressModel{
        title:   title,
        total:   total,
        current: 0,
        status:  "Initializing...",
        theme:   theme,
        spinner: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
    }
}

func (m *ProgressModel) Init() tea.Cmd {
    return m.tick()
}

func (m *ProgressModel) tick() tea.Cmd {
    return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
        return TickMsg(t)
    })
}

func (m *ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case TickMsg:
        m.spinnerIdx = (m.spinnerIdx + 1) % len(m.spinner)
        return m, m.tick()

    case ProgressUpdateMsg:
        m.current = msg.Current
        m.status = msg.Status
        if msg.Done {
            m.cancel = true
            return m, tea.Quit
        }
        return m, nil

    case tea.KeyMsg:
        if msg.String() == "ctrl+c" || msg.String() == "q" {
            m.cancel = true
            return m, tea.Quit
        }
    }

    return m, nil
}

func (m *ProgressModel) View() string {
    if m.cancel {
        return ""
    }

    // 计算进度百分比
    percent := float64(m.current) / float64(m.total) * 100

    // 构建进度条
    barWidth := m.width - 20
    filled := int(float64(barWidth) * percent / 100)

    var bar strings.Builder
    for i := 0; i < barWidth; i++ {
        if i < filled {
            bar.WriteString("█")
        } else {
            bar.WriteString("░")
        }
    }

    // 组合视图
    titleStyle := m.theme.Styles.TitleStyle
    progressStyle := m.theme.Styles.SelectedStyle
    normalStyle := m.theme.Styles.ItemStyle

    var result strings.Builder
    result.WriteString(titleStyle.Render(m.title))
    result.WriteString("\n\n")
    result.WriteString(fmt.Sprintf("%s %s %3.0f%%\n", m.spinner[m.spinnerIdx], bar.String(), percent))
    result.WriteString(fmt.Sprintf("%s / %s\n", humanize.Bytes(uint64(m.current)), humanize.Bytes(uint64(m.total))))
    result.WriteString(fmt.Sprintf("\n%s%s\n", normalStyle.Render("Status: "), progressStyle.Render(m.status)))
    result.WriteString(fmt.Sprintf("\n%s[Ctrl+C to cancel]%s", normalStyle.Render(""), normalStyle.Render("")))

    return result.String()
}

type ProgressUpdateMsg struct {
    Current int
    Status  string
    Done    bool
}

type TickMsg time.Time
```

### 方案 4: 改进的键盘导航

#### Vim 风格导航

```go
// Vim 风格的键盘导航
type VimNavigationModel struct {
    items     []MenuItem
    cursor    int
    offset    int  // 滚动偏移
    yank      []int // 复制的索引
    mode      NavigationMode
    theme     *style.Theme
}

type NavigationMode int

const (
    NormalMode NavigationMode = iota
    VisualMode
    CommandMode
)

func (m *VimNavigationModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch m.mode {
        case NormalMode:
            return m.handleNormalMode(msg)
        case VisualMode:
            return m.handleVisualMode(msg)
        case CommandMode:
            return m.handleCommandMode(msg)
        }

    case tea.WindowSizeMsg:
        // 处理窗口大小变化
    }

    return m, nil
}

func (m *VimNavigationModel) handleNormalMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
    switch msg.String() {
    // Vim 移动
    case "j", "ctrl+n":
        m.moveDown(1)

    case "k", "ctrl+p":
        m.moveUp(1)

    case "h", "ctrl+b":
        m.moveUp(10) // 上一页

    case "l", "ctrl+f":
        m.moveDown(10) // 下一页

    case "g", "g":
        m.moveToFirst()

    case "G":
        m.moveToLast()

    // Vim 操作
    case "v":
        m.mode = VisualMode
        m.yank = []int{m.cursor}

    case "y":
        if len(m.yank) > 0 {
            // 复制选中的项
            return m, tea.Batch(
                func() tea.Msg {
                    return YankMsg{Items: m.getSelectedItems(m.yank)}
                },
            )
        }

    case "d":
        // 删除（标记为禁用）
        if len(m.yank) > 0 {
            for _, idx := range m.yank {
                m.items[idx].Disabled = true
            }
            m.yank = nil
            m.mode = NormalMode
        }

    case "u":
        // 撤销
        return m, func() tea.Msg {
            return UndoMsg{}
        }

    case "ctrl+r":
        // 重做
        return m, func() tea.Msg {
            return RedoMsg{}
        }

    case ":":
        m.mode = CommandMode
        return m, func() tea.Msg {
            return EnterCommandModeMsg{}
        }

    case "/":
        // 进入搜索模式
        return m, func() tea.Msg {
            return EnterSearchMsg{}
        }

    case "n":
        // 下一个搜索结果
        return m, func() tea.Msg {
            return NextSearchResultMsg{}
        }

    case "N":
        // 上一个搜索结果
        return m, func() tea.Msg {
            return PrevSearchResultMsg{}
        }

    case "enter", " ":
        if !m.items[m.cursor].Disabled {
            return m, tea.Quit
        }

    case "q", "ZZ":
        return m, tea.Quit
    }

    return m, nil
}

func (m *VimNavigationModel) moveDown(delta int) {
    newCursor := m.cursor + delta
    if newCursor >= len(m.items) {
        newCursor = len(m.items) - 1
    }
    m.cursor = newCursor

    // 调整滚动偏移
    visibleHeight := 20 // 根据实际高度计算
    if m.cursor >= m.offset+visibleHeight {
        m.offset = m.cursor - visibleHeight + 1
    }
}

func (m *VimNavigationModel) moveUp(delta int) {
    newCursor := m.cursor - delta
    if newCursor < 0 {
        newCursor = 0
    }
    m.cursor = newCursor

    // 调整滚动偏移
    if m.cursor < m.offset {
        m.offset = m.cursor
    }
}

func (m *VimNavigationModel) moveToFirst() {
    m.cursor = 0
    m.offset = 0
}

func (m *VimNavigationModel) moveToLast() {
    m.cursor = len(m.items) - 1
    visibleHeight := 20
    if m.cursor >= visibleHeight {
        m.offset = m.cursor - visibleHeight + 1
    }
}

func (m *VimNavigationModel) getSelectedItems(indices []int) []MenuItem {
    selected := make([]MenuItem, 0, len(indices))
    for _, idx := range indices {
        if idx >= 0 && idx < len(m.items) {
            selected = append(selected, m.items[idx])
        }
    }
    return selected
}
```

### 方案 5: 可访问性改进

#### 键盘快捷键帮助

```go
// 键盘快捷键帮助系统
type KeyBindingHelp struct {
    bindings []KeyBinding
    theme    *style.Theme
}

type KeyBinding struct {
    Keys      []string
    Action    string
    Category  string
}

func NewKeyBindingHelp(theme *style.Theme) *KeyBindingHelp {
    return &KeyBindingHelp{
        bindings: getDefaultBindings(),
        theme:    theme,
    }
}

func getDefaultBindings() []KeyBinding {
    return []KeyBinding{
        // 导航
        {Keys: []string{"↑", "k"}, Action: "Move up", Category: "Navigation"},
        {Keys: []string{"↓", "j"}, Action: "Move down", Category: "Navigation"},
        {Keys: []string{"Page Up", "Ctrl+B"}, Action: "Page up", Category: "Navigation"},
        {Keys: []string{"Page Down", "Ctrl+F"}, Action: "Page down", Category: "Navigation"},
        {Keys: []string{"Home", "g"}, Action: "Go to first", Category: "Navigation"},
        {Keys: []string{"End", "G"}, Action: "Go to last", Category: "Navigation"},

        // 选择
        {Keys: []string{"Enter", "Space"}, Action: "Select item", Category: "Selection"},
        {Keys: []string{"Esc", "q"}, Action: "Cancel/Quit", Category: "Selection"},

        // 搜索
        {Keys: []string{"/"}, Action: "Search", Category: "Search"},
        {Keys: []string{"n"}, Action: "Next search result", Category: "Search"},
        {Keys: []string{"N"}, Action: "Previous search result", Category: "Search"},

        // 编辑
        {Keys: []string{"e"}, Action: "Edit field", Category: "Editing"},
        {Keys: []string{"Ctrl+U"}, Action: "Clear field", Category: "Editing"},
    }
}

func (h *KeyBindingHelp) Render() string {
    // 按类别分组
    categories := make(map[string][]KeyBinding)
    for _, binding := range h.bindings {
        categories[binding.Category] = append(categories[binding.Category], binding)
    }

    var builder strings.Builder

    // 标题
    titleStyle := h.theme.Styles.TitleStyle
    builder.WriteString(titleStyle.Render("Keyboard Shortcuts\n\n"))

    // 按类别渲染
    categoryOrder := []string{"Navigation", "Selection", "Search", "Editing"}
    for _, category := range categoryOrder {
        bindings, ok := categories[category]
        if !ok || len(bindings) == 0 {
            continue
        }

        // 类别标题
        categoryStyle := h.theme.Styles.HeaderStyle
        builder.WriteString(categoryStyle.Render(fmt.Sprintf("%s\n", category)))

        // 绑定列表
        for _, binding := range bindings {
            keyStyle := h.theme.Styles.SelectedStyle
            descStyle := h.theme.Styles.ItemStyle

            keys := strings.Join(binding.Keys, ", ")
            builder.WriteString(fmt.Sprintf("  %-20s %s\n", keyStyle.Render(keys), descStyle.Render(binding.Action)))
        }

        builder.WriteString("\n")
    }

    return builder.String()
}

// 帮助面板模型
type HelpModel struct {
    content string
    theme   *style.Theme
    width   int
    height  int
}

func (m *HelpModel) Init() tea.Cmd {
    return nil
}

func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        case "q", "esc", "?":
            return m, tea.Quit
        }

    case tea.WindowSizeMsg:
        m.width = msg.Width
        m.height = msg.Height
    }

    return m, nil
}

func (m *HelpModel) View() string {
    borderStyle := lipgloss.NewStyle().
        Border(m.theme.Styles.Border).
        BorderForeground(m.theme.Colors.Secondary).
        Padding(1, 2).
        Width(m.width - 4)

    content := borderStyle.Render(m.content)
    helpText := m.theme.Styles.HelpStyle.Render("\n[?] Close")
    return content + helpText
}
```

#### 高对比度模式

```go
// 高对比度主题
func NewHighContrastTheme() *style.Theme {
    return &style.Theme{
        Name: "high-contrast",
        Colors: style.ColorScheme{
            Primary:    lipgloss.Color("15"),  // 白色
            Secondary:  lipgloss.Color("8"),   // 灰色
            Success:    lipgloss.Color("10"),  // 绿色（最亮）
            Warning:    lipgloss.Color("11"),  // 黄色（最亮）
            Error:      lipgloss.Color("12"),  // 红色（最亮）
            Muted:      lipgloss.Color("7"),   // 正常亮度
            Background: lipgloss.Color("0"),   // 黑色
            Foreground: lipgloss.Color("15"),  // 白色
        },
        Layout: style.LayoutConfig{
            Padding:      [4]int{1, 2, 1, 2},
            BorderWidth:  2,  // 更粗的边框
            MinWidth:     80,
        },
        Styles: style.StyleConfig{
            Border:        lipgloss.ThickBorder(),  // 粗边框
            TitleStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")),
            SelectedStyle: lipgloss.NewStyle().Bold(true).Reverse(true),  // 反转显示
            DisabledStyle: lipgloss.NewStyle().Faint(true),
            HelpStyle:     lipgloss.NewStyle(),
            ErrorStyle:    lipgloss.NewStyle().Bold(true).Blink(true),  // 闪烁
        },
    }
}
```

## 用户体验最佳实践

### 1. 一致的交互模式

```go
// 统一的交互模式配置
type UXConfig struct {
    NavigationStyle string // "vim" or "standard"
    ThemeName       string
    ShowHelp        bool
    ConfirmActions  bool
}

var DefaultUXConfig = &UXConfig{
    NavigationStyle: "standard",
    ThemeName:       "default",
    ShowHelp:        true,
    ConfirmActions:  true,
}
```

### 2. 渐进式增强

```go
// 渐进式功能启用
func ApplyUXProgressively(config *UXConfig) {
    // 基础功能
    enableBasicNavigation()

    // 中级功能
    if config.ShowHelp {
        enableHelpSystem()
    }

    // 高级功能
    if config.NavigationStyle == "vim" {
        enableVimNavigation()
    }

    // 实验性功能
    if os.Getenv("ENABLE_EXPERIMENTAL") == "true" {
        enableExperimentalFeatures()
    }
}
```

### 3. 用户反馈收集

```go
// 匿名使用统计
type Telemetry struct {
    enabled bool
    client  *http.Client
    endpoint string
}

func (t *Telemetry) RecordEvent(event string, data map[string]interface{}) {
    if !t.enabled {
        return
    }

    // 异步发送，不影响用户体验
    go func() {
        payload := map[string]interface{}{
            "event":     event,
            "timestamp": time.Now().Unix(),
            "data":      data,
        }

        json_data, _ := json.Marshal(payload)
        req, _ := http.NewRequest("POST", t.endpoint, bytes.NewBuffer(json_data))
        req.Header.Set("Content-Type", "application/json")

        resp, err := t.client.Do(req)
        if err != nil {
            return
        }
        defer resp.Body.Close()
    }()
}
```

## 性能与体验平衡

### 性能监控

```go
// UX 性能监控
type UXPerfMonitor struct {
    renderTimes []time.Duration
    inputLatency []time.Duration
    mutex       sync.RWMutex
}

func (m *UXPerfMonitor) RecordRender(duration time.Duration) {
    m.mutex.Lock()
    defer m.mutex.Unlock()

    m.renderTimes = append(m.renderTimes, duration)
    if len(m.renderTimes) > 100 {
        m.renderTimes = m.renderTimes[1:]
    }
}

func (m *UXPerfMonitor) GetAvgRenderTime() time.Duration {
    m.mutex.RLock()
    defer m.mutex.RUnlock()

    if len(m.renderTimes) == 0 {
        return 0
    }

    var sum time.Duration
    for _, d := range m.renderTimes {
        sum += d
    }
    return sum / time.Duration(len(m.renderTimes))
}

// 自适应质量设置
type AdaptiveQuality struct {
    monitor *UXPerfMonitor
}

func (a *AdaptiveQuality) AdjustSettings() {
    avgRender := a.monitor.GetAvgRenderTime()

    if avgRender > time.Millisecond*33 {  // 低于 30fps
        // 降低视觉质量
        reduceAnimationQuality()
        disableEffects()
    } else if avgRender < time.Millisecond*16 {  // 高于 60fps
        // 提升视觉质量
        enableAnimations()
        enableEffects()
    }
}
```

## 推荐实现

### 完整的 UX 优化配置

```go
// 完整的 UX 配置
type UXOptimization struct {
    Search          *SearchConfig
    ErrorHandling   *ErrorHandlingConfig
    Progress        *ProgressConfig
    Navigation      *NavigationConfig
    Accessibility   *AccessibilityConfig
}

type SearchConfig struct {
    EnableFuzzy      bool
    ShowSuggestions  bool
    HistorySize      int
}

type ErrorHandlingConfig struct {
    ShowSuggestions bool
    EnableRecovery   bool
    FriendlyMessages bool
}

type ProgressConfig struct {
    ShowSpinner      bool
    ShowPercentage   bool
    EnableCancel     bool
}

type NavigationConfig struct {
    Style           string  // "standard" or "vim"
    EnableScrolling bool
    PageScrollSize  int
}

type AccessibilityConfig struct {
    HighContrast    bool
    LargeText       bool
    KeyboardHelp    bool
}

// 应用 UX 优化
func ApplyUXOptimizations(config *UXOptimization) {
    if config.Search.EnableFuzzy {
        EnableFuzzySearch()
    }

    if config.ErrorHandling.FriendlyMessages {
        EnableFriendlyErrors()
    }

    if config.Accessibility.HighContrast {
        EnableHighContrastMode()
    }

    // ... 应用其他配置
}
```

## 总结

用户体验优化要点：

1. **增强搜索**: 模糊匹配、搜索历史、高亮显示
2. **智能错误**: 友好信息、恢复建议、错误分类
3. **进度反馈**: 进度条、状态更新、可取消
4. **键盘导航**: Vim 模式、快捷键帮助、一致性
5. **可访问性**: 高对比度、帮助文档、键盘支持

优秀的用户体验是工具成功的关键。建议从搜索和错误处理优化开始，逐步引入其他改进。始终收集用户反馈，持续迭代优化。
