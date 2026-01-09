# Flat Menu View Example

This example demonstrates the **flat menu view** feature of cobrax.

## What is Flat Menu View?

When you have a complex command hierarchy with many subcommands, navigating through nested menus can be tedious. The flat menu view shows **all executable commands in a single list** with their full paths, making it easier to quickly find and execute any command.

## Features Demonstrated

- **Flat command listing**: All commands visible at once
- **Full path display**: Shows command hierarchy (e.g., `db migrate`)
- **Direct execution**: Select any command directly without navigating through menus
- **Flag configuration**: Configure flags before execution
- **Confirmation dialog**: Preview command before executing

## Command Hierarchy

This example includes:

```
flatmenu
├── db (Database operations)
│   ├── migrate    - Run database migrations
│   ├── seed       - Seed database with initial data
│   └── backup     - Backup database
├── server (Server operations)
│   ├── start      - Start the server
│   ├── stop       - Stop the server
│   └── restart    - Restart the server
├── user (User management)
│   ├── list       - List all users
│   ├── create     - Create a new user
│   └── delete     - Delete a user
└── config (Configuration)
    ├── get        - Get configuration value
    ├── set        - Set configuration value
    └── init       - Initialize configuration
```

## Running the Example

### Traditional CLI Mode

```bash
# Run with subcommand
go run main.go server start --port=9090

# Show help
go run main.go --help
```

### TUI Mode (Flat Menu)

```bash
# Launch TUI with flat menu view
go run main.go --tui

# Or set environment variable
COBRA_TUI=true go run main.go
```

## TUI Experience

When you run with `--tui`, you'll see:

```
┌──────────────────────────────────────────────────────────┐
│             flatmenu Commands                            │
├──────────────────────────────────────────────────────────┤
│  ▶ db migrate    Run database migrations                 │
│    db seed       Seed database with initial data         │
│    db backup     Backup database                        │
│    server start  Start the server                       │
│    server stop   Stop the server                        │
│    server restart Restart the server                    │
│    user list     List all users                         │
│    user create   Create a new user                      │
│    user delete   Delete a user                          │
│    config get    Get configuration value                │
│    config set    Set configuration value                │
│    config init   Initialize configuration                │
│                                                          │
│  [↑↓ Navigate] [Enter Select] [Esc Quit]                │
└──────────────────────────────────────────────────────────┘
```

## Benefits

| Traditional Navigation | Flat Menu View |
|-----------------------|----------------|
| Navigate: server → start | Direct: server start |
| Navigate: db → migrate → select | Direct: db migrate |
| Multiple menu levels | Single list |
| Can lose context | Full path always visible |

## Next Steps

Try selecting different commands and:
1. Configure their flags interactively
2. See the command preview before execution
3. Confirm or cancel execution
