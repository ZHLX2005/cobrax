# cobra-x

> Enhanced Cobra CLI with command tree visualization

`github.com/ZHLX2005/cobrax` is an enhanced version of [spf13/cobra](https://github.com/spf13/cobra) that maintains 100% API compatibility while adding **command tree visualization** with colored output.

## Features

- **Zero Code Changes** - Enhance existing cobra commands without modifying imports
- **Decorator Pattern** - Non-invasive enhancement using `cobra.Enhance()`
- **100% API Compatible** - Keep using `github.com/spf13/cobra` as before
- **Command Tree Display** - Show all commands with `--tree` flag
- **Colored Output** - Beautiful terminal output with lipgloss
- **Multiple Themes** - Built-in themes (default, dracula, nord, monokai, light)

## Installation

```bash
go get github.com/ZHLX2005/cobrax
```

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/ZHLX2005/cobrax/cobra"
)

func main() {
    rootCmd := cobra.NewCommand("myapp",
        cobra.WithShort("My application"),
        cobra.WithTreeTheme(cobra.DraculaTreeTheme()),
    )

    serverCmd := cobra.NewCommand("server",
        cobra.WithShort("Start the server"),
        cobra.WithRun(func(cmd *cobra.Command, args []string) {
            fmt.Println("Server started!")
        }),
    )

    rootCmd.AddCommand(serverCmd)
    rootCmd.Execute()
}
```

### Display Command Tree

Simply add the `--tree` flag to see all commands:

```bash
./myapp --tree
```

Output:
```
Command Tree (2 commands)

 1. myapp server ✓
     Start the server
```

### Theme Options

```bash
./myapp --tree --tree-theme=dracula
./myapp --tree --tree-theme=nord
./myapp --tree --tree-theme=monokai
./myapp --tree --tree-theme=light
```

### Additional Tree Options

```bash
# Show flags
./myapp --tree --tree-flags

# Hide descriptions
./myapp --tree --tree-long=false
```

## Decorator Pattern

Enhance existing cobra commands without modifying your code:

```go
import (
    "github.com/spf13/cobra"
    cobrax "github.com/ZHLX2005/cobrax/cobra"
)

// Your existing command - NO CHANGES NEEDED!
var rootCmd = &cobra.Command{
    Use:   "myapp",
    Short: "My application",
    Run: func(cmd *cobra.Command, args []string) {
        // Your existing logic
    },
}

func main() {
    // Just wrap it before Execute()!
    enhanced := cobrax.Enhance(rootCmd,
        cobrax.WithThemeName("nord"),
    )

    enhanced.Execute()
}
```

## Configuration

### Tree Display Options

```go
type TreeConfig struct {
    Theme       *TreeTheme  // Color theme
    ShowFlags   bool        // Show flags in tree
    ShowLong    bool        // Show descriptions
    IndentWidth int         // Indentation width
}
```

### Available Themes

- `DefaultTreeTheme()` - Default blue theme
- `DraculaTreeTheme()` - Dracula theme
- `NordTreeTheme()` - Nord theme
- `MonokaiTreeTheme()` - Monokai theme
- `LightTreeTheme()` - Light theme

## Command Options

- `WithShort(short string)` - Set short description
- `WithLong(long string)` - Set long description
- `WithRun(fn func(*Command, []string))` - Set run function
- `WithRunE(fn func(*Command, []string) error)` - Set run function with error
- `WithTreeTheme(theme *TreeTheme)` - Set tree theme

## API Reference

### Creating Commands

```go
// Create a new command with options
cmd := cobra.NewCommand("name",
    cobra.WithShort("Short description"),
    cobra.WithLong("Long description"),
    cobra.WithRun(func(cmd *cobra.Command, args []string) {
        // Command logic
    }),
    cobra.WithTreeTheme(cobra.NordTreeTheme()),
)
```

### Display Functions

```go
// Display flat tree
DisplayFlatTree(cmd, config)

// Display hierarchical tree
DisplayTree(cmd, config)

// Get command tree string
GetCommandTreeString(root, "dracula", false, true)
```

## Examples

See the [examples/basic](examples/basic/) directory for a complete example.

## Compatibility

This library is designed to be a drop-in enhancement for `spf13/cobra`. All existing code using `spf13/cobra` will continue to work without modifications.

### Migration from spf13/cobra

Simply change how you create commands:

```go
// Before
import "github.com/spf13/cobra"
rootCmd := &cobra.Command{Use: "myapp"}

// After
import "github.com/ZHLX2005/cobrax/cobra"
rootCmd := cobra.NewCommand("myapp")
```

## License

This project maintains the same license as the original cobra project (Apache 2.0).

## Acknowledgments

- [spf13/cobra](https://github.com/spf13/cobra) - The original CLI framework
- [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss) - Styling framework
