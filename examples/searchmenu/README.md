# Search Menu Example

This example demonstrates the **search menu** feature for quickly finding commands in large CLI applications.

## What is Search Menu?

When your application has many commands, navigating through them can be time-consuming. The search menu allows you to:
- **Filter commands in real-time** as you type
- **Search by command name, description, or ID**
- **Navigate results** with arrow keys
- **Quickly access** any command regardless of hierarchy

## Features

### Search Mode

Press **`/`** or **`Ctrl+S`** to enter search mode:

```
┌──────────────────────────────────────────────────────────┐
│              Search commands                             │
│                                                          │
│  /test_                                                   │
│                                                          │
│  ▶ test unit        Run unit tests                      │
│    test integration Run integration tests               │
│    test e2e         Run end-to-end tests                │
│                                                          │
│  [Type to search] [Enter Select] [Esc Exit search]      │
└──────────────────────────────────────────────────────────┘
```

### Filtering

When you filter commands, the results update instantly:

```
┌──────────────────────────────────────────────────────────┐
│              Select a command                           │
│                                                          │
│  Filter: deploy                                         │
│                                                          │
│  ▶ deploy staging   Deploy to staging environment      │
│    deploy production Deploy to production environment   │
│    deploy rollback  Rollback deployment                │
│                                                          │
│  [↑↓ Navigate] [Enter Select] [/ Search] [Ctrl+R Clear]│
└──────────────────────────────────────────────────────────┘
```

## Keyboard Shortcuts

| Key | Mode | Action |
|-----|------|--------|
| `/` or `Ctrl+S` | Navigation | Enter search mode |
| `Esc` | Search | Exit search mode |
| `Ctrl+U` | Search | Clear search input |
| `Ctrl+R` | Navigation | Clear filter |
| `↑/↓` or `j/k` | Both | Navigate items |
| `Enter` | Both | Select item |
| `Ctrl+C` or `q` | Both | Quit |

## Command Structure

This example includes 20+ commands across multiple categories:

```
searchmenu
├── dev (Development)
│   ├── build      - Build the application
│   ├── test       - Run tests
│   ├── lint       - Run linter
│   └── fmt        - Format code
├── deploy (Deployment)
│   ├── staging    - Deploy to staging
│   ├── production - Deploy to production
│   └── rollback   - Rollback deployment
├── test (Testing)
│   ├── unit       - Run unit tests
│   ├── integration - Run integration tests
│   └── e2e        - Run E2E tests
├── config (Configuration)
│   ├── get        - Get configuration
│   ├── set        - Set configuration
│   └── list       - List all configuration
├── db (Database)
│   ├── migrate    - Run migrations
│   ├── seed       - Seed database
│   └── reset      - Reset database
└── server (Server)
    ├── start      - Start server
    ├── stop       - Stop server
    ├── restart    - Restart server
    └── status     - Check status
```

## Running the Example

```bash
# Run the demo
go run main.go

# Try searching for:
# - "test"  -> Shows all test commands
# - "deploy" -> Shows deployment commands
# - "server" -> Shows server commands
# - "config" -> Shows configuration commands
```

## Use Cases

The search menu is especially useful for:

1. **Large CLI Applications**: When you have 20+ commands
2. **Deep Hierarchies**: Commands nested 3+ levels deep
3. **Infrequent Commands**: When you don't remember the exact command path
4. **New Users**: When learning available commands

## Benefits

| Without Search | With Search |
|----------------|-------------|
| Navigate: dev → test | Type: "test" |
| Navigate: deploy → production | Type: "prod" |
| Navigate: server → restart | Type: "restart" |
| Multiple menu levels | Instant access |

## Integration with cobrax

This example shows the search menu directly. In production use with cobrax:

```go
import (
    "github.com/spf13/cobra"
    cobrax "github.com/ZHLX2005/cobrax/cobra"
)

var rootCmd = &cobra.Command{...}

func main() {
    // The decorator automatically provides search functionality
    enhanced := cobrax.Enhance(rootCmd,
        cobrax.WithEnhanceTUIEnabled(true),
    )
    enhanced.Execute()
}
```

When you run with `--tui`, the search menu is available automatically!
