# cobra

> Enhanced Cobra CLI with interactive TUI support using Decorator Pattern

`github.com/ZHLX2005/cobra` is an enhanced version of [spf13/cobra](https://github.com/spf13/cobra) that maintains 100% API compatibility while adding interactive TUI (Terminal User Interface) support **using the decorator pattern**.

## Features

- **Zero Code Changes** - Enhance existing cobra commands without modifying imports
- **Decorator Pattern** - Non-invasive enhancement using `cobra.Enhance()`
- **100% API Compatible** - Keep using `github.com/spf13/cobra` as before
- **Interactive TUI** - Automatic TUI generation based on command tree
- **Customizable** - Custom renderers and panel builders
- **Theme System** - Built-in themes (default, dark, light, minimal, dracula, nord, monokai)

## Installation

```bash
go get github.com/ZHLX2005/cobra
```

## Quick Start

### Decorator Pattern (Recommended)

**No need to change your existing imports!** Just wrap your command with `cobra.Enhance()`:

```go
package main

import (
    "github.com/spf13/cobra"  // Keep your original import!
)

var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    Run: func(cmd *cobra.Command, args []string) {
        // Your existing code
    },
}

func main() {
    // Just wrap your existing command!
    enhancedCmd := cobra.Enhance(rootCmd,
        cobra.WithTUIEnabled(true),
        cobra.WithTheme("dracula"),
    )

    enhancedCmd.Execute()
}
```

### For Existing Projects

Got an existing cobra project? Just add the enhancement wrapper:

```go
import (
    "github.com/spf13/cobra"

    // Import the enhancement package
    cobrax "github.com/ZHLX2005/cobra/cobra"
)

// Your existing command - NO CHANGES NEEDED!
var rootCmd = &cobra.Command{
    Use:   "glm4v",
    Short: "GLM-4.5V Tool",
    Run: func(cmd *cobra.Command, args []string) {
        // Your existing logic
    },
}

func init() {
    // Your existing flags
    rootCmd.Flags().StringVar(&apiKey, "api-key", "", "API Key")
    // ...
}

func main() {
    // Just wrap it before Execute()!
    enhanced := cobrax.Enhance(rootCmd,
        cobrax.WithTUIEnabled(true),
        cobrax.WithTheme("nord"),
    )

    enhanced.Execute()
}
```

---

### Traditional Usage (Full Replacement)

If you prefer to fully replace cobra:

```go
package main

import (
    "github.com/ZHLX2005/cobra/cobra"
    "github.com/ZHLX2005/cobra/tui/style"
)

func main() {
    rootCmd := cobra.NewCommand("myapp",
        cobra.WithShort("My application"),
        cobra.WithTUIEnabled(true),
    )

    // Set TUI theme
    rootCmd.SetTUIConfig(&cobra.TUIConfig{
        Enabled:       true,
        Theme:         style.NewTheme(style.ThemeDefault),
        ShowFlags:     true,
    })

    serverCmd := cobra.NewCommand("server",
        cobra.WithShort("Start the server"),
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            port, _ := cmd.Flags().GetString("port")
            fmt.Printf("Starting server on port %s\n", port)
        }),
    )

    serverCmd.Command.Flags().StringP("port", "p", "8080", "Server port")
    rootCmd.AddCommand(serverCmd)

    rootCmd.Execute()
}
```

### Launching TUI

There are three ways to launch the TUI interface:

1. **Flag**: `./myapp --tui`
2. **Environment**: `COBRA_TUI=true ./myapp`
3. **Programmatic**: Enable TUI in code (see example above)

### CLI Mode (Traditional)

The traditional CLI mode still works as expected:

```bash
./myapp server --port=9090
```

## Configuration

### TUI Config Options

```go
type TUIConfig struct {
    Enabled              bool              // Enable TUI
    Renderer             Renderer          // Custom renderer
    Theme                *Theme            // Theme
    ShowDescription      bool              // Show command descriptions
    ShowFlags            bool              // Show flag configuration panel
    InteractiveMode      InteractiveMode   // Auto/TUI/CLI
    AutoDetect           bool              // Auto-detect terminal capability
    ConfirmBeforeExecute bool              // Confirm before execution
}
```

### Functional Options

```go
rootCmd.SetTUIConfig(cobra.NewTUIConfig(
    cobra.WithTUIEnabled(true),
    cobra.WithTUITheme(style.NewTheme(style.ThemeDracula)),
    cobra.WithTUIShowDescription(true),
    cobra.WithTUIConfirmBeforeExecute(true),
))
```

### Available Themes

- `ThemeDefault` - Default blue theme
- `ThemeDark` - Dark theme
- `ThemeLight` - Light theme
- `ThemeMinimal` - Minimal theme (no borders)
- `ThemeDracula` - Dracula theme
- `ThemeNord` - Nord theme
- `ThemeMonokai` - Monokai theme

## API Reference

### Command Creation

```go
// Create a new command with options
cmd := cobra.NewCommand("name",
    cobra.WithShort("Short description"),
    cobra.WithLong("Long description"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        // Command logic
    }),
    cobra.WithTUIEnabled(true),
)
```

### Command Options

- `WithShort(short string)` - Set short description
- `WithLong(long string)` - Set long description
- `WithRun(fn func(*Command, []string))` - Set run function
- `WithTUIConfig(config *TUIConfig)` - Set TUI configuration
- `WithTUIEnabled(enabled bool)` - Enable/disable TUI

### TUI Methods

```go
// Enable TUI
cmd.EnableTUI()

// Disable TUI
cmd.DisableTUI()

// Set TUI configuration
cmd.SetTUIConfig(config)

// Set custom renderer
cmd.SetTUIRenderer(renderer)

// Set panel builder
cmd.SetPanelBuilder(builder)
```

## Customization

### Custom Renderer

Implement the `Renderer` interface to create custom TUI experiences:

```go
type MyRenderer struct {
    // Your fields
}

func (r *MyRenderer) RenderCommandMenu(cmd *cobra.Command, options []tui.MenuItem) (*cobra.Command, error) {
    // Custom menu rendering logic
    return selected, nil
}

func (r *MyRenderer) RenderFlagForm(cmd *cobra.Command, flags []tui.FlagItem) (map[string]string, error) {
    // Custom form rendering logic
    return values, nil
}

func (r *MyRenderer) RenderConfirmation(title, message string) (bool, error) {
    // Custom confirmation logic
    return confirmed, nil
}

func (r *MyRenderer) RenderHelp(cmd *cobra.Command) error {
    // Custom help rendering
    return nil
}

func (r *MyRenderer) Cleanup() error {
    // Cleanup resources
    return nil
}

// Use custom renderer
cmd.SetTUIRenderer(&MyRenderer{})
```

### Custom Panel Builder

```go
import "github.com/ZHLX2005/cobra/ext"

type MyPanelBuilder struct {
    ext.DefaultPanelBuilder
}

func (b *MyPanelBuilder) BuildMenu(cmd *cobra.Command) *ext.MenuPanel {
    panel := b.DefaultPanelBuilder.BuildMenu(cmd)
    // Customize panel
    return panel
}

cmd.SetPanelBuilder(&MyPanelBuilder{})
```

## Examples

See the [examples](examples/) directory for complete examples:

- [examples/basic](examples/basic/) - Basic TUI application

## Compatibility

This library is designed to be a drop-in replacement for `spf13/cobra`. All existing code using `spf13/cobra` will continue to work without modifications.

### Migration from spf13/cobra

Simply change the import:

```go
// Before
import "github.com/spf13/cobra"

// After
import "github.com/ZHLX2005/cobra/cobra"
```

Then optionally enable TUI:

```go
cmd.EnableTUI()
```

## TUI Controls

### Menu Panel

- `↑/↓` or `j/k` - Navigate
- `Enter` or `Space` - Select
- `Esc` or `q` - Quit

### Flag Form Panel

- `↑/↓` or `Tab/Shift+Tab` - Navigate between flags
- `E` - Enter edit mode
- `Enter` - Save and quit
- `Esc` - Cancel

### Confirmation Panel

- `←/→` or `h/l` - Navigate options
- `Enter` - Confirm selection

## Keyboard Shortcuts

Global shortcuts:

- `Ctrl+C` - Quit
- `Esc` - Go back / Cancel
- `Q` - Quit

## Project Status

This is an active development project. The API is subject to change.

## License

This project maintains the same license as the original cobra project (Apache 2.0).

## Acknowledgments

- [spf13/cobra](https://github.com/spf13/cobra) - The original CLI framework
- [charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Styling framework

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.
