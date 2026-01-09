package style

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme TUI 主题配置
// 定义了 TUI 界面的颜色、布局和样式
type Theme struct {
	// Name 主题名称
	Name string

	// Colors 颜色配置
	Colors ColorScheme

	// Layout 布局配置
	Layout LayoutConfig

	// Styles 样式配置
	Styles StyleConfig
}

// ColorScheme 颜色配置
type ColorScheme struct {
	// Primary 主色调 - 用于标题、选中项等
	Primary lipgloss.Color

	// Secondary 次要色调 - 用于边框、分隔线等
	Secondary lipgloss.Color

	// Success 成功色 - 用于成功提示
	Success lipgloss.Color

	// Warning 警告色 - 用于警告提示
	Warning lipgloss.Color

	// Error 错误色 - 用于错误提示
	Error lipgloss.Color

	// Muted 弱化色 - 用于禁用项、次要信息
	Muted lipgloss.Color

	// Background 背景色
	Background lipgloss.Color

	// Foreground 前景色
	Foreground lipgloss.Color
}

// LayoutConfig 布局配置
type LayoutConfig struct {
	// Padding 内边距
	Padding [4]int // top, right, bottom, left

	// Margin 外边距
	Margin [4]int // top, right, bottom, left

	// BorderWidth 边框宽度
	BorderWidth int

	// MinWidth 最小宽度
	MinWidth int

	// MinHeight 最小高度
	MinHeight int

	// MaxWidth 最大宽度（0 表示无限制）
	MaxWidth int

	// MaxHeight 最大高度（0 表示无限制）
	MaxHeight int
}

// StyleConfig 样式配置
type StyleConfig struct {
	// Border 边框样式
	Border lipgloss.Border

	// TitleStyle 标题样式
	TitleStyle lipgloss.Style

	// HeaderStyle 表头样式
	HeaderStyle lipgloss.Style

	// ItemStyle 列表项样式
	ItemStyle lipgloss.Style

	// SelectedStyle 选中项样式
	SelectedStyle lipgloss.Style

	// DisabledStyle 禁用项样式
	DisabledStyle lipgloss.Style

	// HelpStyle 帮助文本样式
	HelpStyle lipgloss.Style

	// ErrorStyle 错误文本样式
	ErrorStyle lipgloss.Style
}

// 主题常量
const (
	ThemeDefault  = "default"
	ThemeDark     = "dark"
	ThemeLight    = "light"
	ThemeMinimal  = "minimal"
	ThemeDracula  = "dracula"
	ThemeNord     = "nord"
	ThemeMonokai  = "monokai"
)

// DefaultTheme 返回默认主题
func DefaultTheme() *Theme {
	return NewTheme(ThemeDefault)
}

// NewTheme 创建指定名称的主题
func NewTheme(name string) *Theme {
	switch name {
	case ThemeDark:
		return darkTheme()
	case ThemeLight:
		return lightTheme()
	case ThemeMinimal:
		return minimalTheme()
	case ThemeDracula:
		return draculaTheme()
	case ThemeNord:
		return nordTheme()
	case ThemeMonokai:
		return monokaiTheme()
	default:
		return defaultTheme()
	}
}

// defaultTheme 默认主题（蓝色调）
func defaultTheme() *Theme {
	return &Theme{
		Name: ThemeDefault,
		Colors: ColorScheme{
			Primary:   lipgloss.Color("86"),  // blue
			Secondary: lipgloss.Color("245"), // gray
			Success:   lipgloss.Color("82"),  // green
			Warning:   lipgloss.Color("228"), // yellow
			Error:     lipgloss.Color("196"), // red
			Muted:     lipgloss.Color("242"), // dim gray
			Background: lipgloss.Color("235"), // dark blue
			Foreground: lipgloss.Color("255"), // white
		},
		Layout: defaultLayout(),
		Styles: defaultStyles(),
	}
}

// darkTheme 暗色主题
func darkTheme() *Theme {
	theme := defaultTheme()
	theme.Name = ThemeDark
	theme.Colors.Background = lipgloss.Color("236")
	theme.Colors.Foreground = lipgloss.Color("252")
	return theme
}

// lightTheme 亮色主题
func lightTheme() *Theme {
	return &Theme{
		Name: ThemeLight,
		Colors: ColorScheme{
			Primary:    lipgloss.Color("26"),  // blue
			Secondary:  lipgloss.Color("245"), // gray
			Success:    lipgloss.Color("28"),  // green
			Warning:    lipgloss.Color("214"), // orange
			Error:      lipgloss.Color("160"), // red
			Muted:      lipgloss.Color("245"), // gray
			Background: lipgloss.Color("255"), // white
			Foreground: lipgloss.Color("16"),  // black
		},
		Layout: defaultLayout(),
		Styles: defaultStyles(),
	}
}

// minimalTheme 极简主题
func minimalTheme() *Theme {
	return &Theme{
		Name: ThemeMinimal,
		Colors: ColorScheme{
			Primary:    lipgloss.Color("7"),   // white
			Secondary:  lipgloss.Color("8"),   // dark gray
			Success:    lipgloss.Color("7"),   // white
			Warning:    lipgloss.Color("7"),   // white
			Error:      lipgloss.Color("7"),   // white
			Muted:      lipgloss.Color("8"),   // dark gray
			Background: lipgloss.Color("0"),   // black
			Foreground: lipgloss.Color("7"),   // white
		},
		Layout: minimalLayout(),
		Styles: minimalStyles(),
	}
}

// draculaTheme Dracula 主题
func draculaTheme() *Theme {
	return &Theme{
		Name: ThemeDracula,
		Colors: ColorScheme{
			Primary:    lipgloss.Color("#BD93F9"), // purple
			Secondary:  lipgloss.Color("#6272A4"), // comment
			Success:    lipgloss.Color("#50FA7B"), // green
			Warning:    lipgloss.Color("#F1FA8C"), // yellow
			Error:      lipgloss.Color("#FF5555"), // red
			Muted:      lipgloss.Color("#6272A4"), // comment
			Background: lipgloss.Color("#282A36"), // background
			Foreground: lipgloss.Color("#F8F8F2"), // foreground
		},
		Layout: defaultLayout(),
		Styles: defaultStyles(),
	}
}

// nordTheme Nord 主题
func nordTheme() *Theme {
	return &Theme{
		Name: ThemeNord,
		Colors: ColorScheme{
			Primary:    lipgloss.Color("#88C0D0"), // frost
			Secondary:  lipgloss.Color("#4C566A"), // dark
			Success:    lipgloss.Color("#A3BE8C"), // green
			Warning:    lipgloss.Color("#EBCB8B"), // yellow
			Error:      lipgloss.Color("#BF616A"), // red
			Muted:      lipgloss.Color("#4C566A"), // dark
			Background: lipgloss.Color("#2E3440"), // polar night
			Foreground: lipgloss.Color("#D8DEE9"), // snow storm
		},
		Layout: defaultLayout(),
		Styles: defaultStyles(),
	}
}

// monokaiTheme Monokai 主题
func monokaiTheme() *Theme {
	return &Theme{
		Name: ThemeMonokai,
		Colors: ColorScheme{
			Primary:    lipgloss.Color("#66D9EF"), // cyan
			Secondary:  lipgloss.Color("#75715E"), // comment
			Success:    lipgloss.Color("#A6E22E"), // green
			Warning:    lipgloss.Color("#E6DB74"), // yellow
			Error:      lipgloss.Color("#F92672"), // magenta
			Muted:      lipgloss.Color("#75715E"), // comment
			Background: lipgloss.Color("#272822"), // background
			Foreground: lipgloss.Color("#F8F8F2"), // foreground
		},
		Layout: defaultLayout(),
		Styles: defaultStyles(),
	}
}

// defaultLayout 默认布局配置
func defaultLayout() LayoutConfig {
	return LayoutConfig{
		Padding:      [4]int{1, 2, 1, 2},
		Margin:       [4]int{0, 0, 0, 0},
		BorderWidth:  1,
		MinWidth:     60,
		MinHeight:    20,
		MaxWidth:     120,
		MaxHeight:    0,
	}
}

// minimalLayout 极简布局配置
func minimalLayout() LayoutConfig {
	return LayoutConfig{
		Padding:      [4]int{1, 1, 1, 1},
		Margin:       [4]int{0, 0, 0, 0},
		BorderWidth:  0,
		MinWidth:     40,
		MinHeight:    10,
		MaxWidth:     0,
		MaxHeight:    0,
	}
}

// defaultStyles 默认样式配置
func defaultStyles() StyleConfig {
	normalBorder := lipgloss.NormalBorder()
	return StyleConfig{
		Border:        normalBorder,
		TitleStyle:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		HeaderStyle:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")),
		ItemStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("255")),
		SelectedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Background(lipgloss.Color("235")).Bold(true),
		DisabledStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("242")),
		HelpStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("242")).Faint(true),
		ErrorStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true),
	}
}

// minimalStyles 极简样式配置
func minimalStyles() StyleConfig {
	hiddenBorder := lipgloss.Border{
		Top:         " ",
		Bottom:      " ",
		Left:        " ",
		Right:       " ",
		TopLeft:     " ",
		TopRight:    " ",
		BottomLeft:  " ",
		BottomRight: " ",
	}
	return StyleConfig{
		Border:        hiddenBorder,
		TitleStyle:    lipgloss.NewStyle().Bold(true),
		HeaderStyle:   lipgloss.NewStyle().Bold(true),
		ItemStyle:     lipgloss.NewStyle(),
		SelectedStyle: lipgloss.NewStyle().Bold(true).Reverse(true),
		DisabledStyle: lipgloss.NewStyle().Faint(true),
		HelpStyle:     lipgloss.NewStyle().Faint(true),
		ErrorStyle:    lipgloss.NewStyle().Bold(true),
	}
}

// GetBorderWidth 获取边框宽度
func (t *Theme) GetBorderWidth() int {
	return t.Layout.BorderWidth
}

// GetPadding 获取内边距
func (t *Theme) GetPadding() [4]int {
	return t.Layout.Padding
}

// GetMargin 获取外边距
func (t *Theme) GetMargin() [4]int {
	return t.Layout.Margin
}

// GetSize 获取内容区域大小（减去边框和内边距）
func (t *Theme) GetContentSize(width, height int) (int, int) {
	hPadding := t.Layout.Padding[1] + t.Layout.Padding[3]
	vPadding := t.Layout.Padding[0] + t.Layout.Padding[2]
	border := t.Layout.BorderWidth * 2

	contentWidth := width - hPadding - border
	contentHeight := height - vPadding - border

	return contentWidth, contentHeight
}
