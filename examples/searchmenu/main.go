// This example demonstrates the search menu feature
// Press '/' to enter search mode and filter commands
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	spf13cobra "github.com/spf13/cobra"
	"github.com/ZHLX2005/cobrax/tui"
	"github.com/ZHLX2005/cobrax/tui/style"
)

var rootCmd = &spf13cobra.Command{
	Use:   "searchmenu",
	Short: "Demo showing search functionality in TUI",
	Long: `This example demonstrates the search menu feature.

When you run with --tui, you can:
- Press '/' or Ctrl+S to enter search mode
- Type to filter commands in real-time
- Press Esc to exit search mode
- Press Ctrl+R to clear the filter

This makes it easy to find commands in large applications.`,
	Run: func(cmd *spf13cobra.Command, args []string) {
		fmt.Println("Root command executed")
	},
}

func main() {
	// Set up command structure
	setupCommands()

	// For this demo, we'll use a custom search menu
	// In production, the cobrax decorator will handle this
	runDemoSearchMenu()
}

func runDemoSearchMenu() {
	// Collect all executable commands
	items := collectCommandItems(rootCmd)

	// Create search menu model
	theme := style.NewTheme(style.ThemeDracula)
	model := tui.NewSearchMenuModel(items, theme, 80, 24)

	// Run the program
	p := tea.NewProgram(model, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	// Check result
	if searchModel, ok := result.(*tui.SearchMenuModel); ok {
		if searchModel.IsCancelled() {
			fmt.Println("Operation cancelled")
			return
		}

		filteredItems := searchModel.GetFilteredItems()
		cursor := searchModel.GetCursor()

		if cursor >= 0 && cursor < len(filteredItems) {
			selected := filteredItems[cursor]
			fmt.Printf("Selected command: %s\n", selected.Label)
			fmt.Printf("Description: %s\n", selected.Description)

			// Find and execute the command
			cmdPath := strings.Fields(selected.Label)
			targetCmd := findCommand(rootCmd, cmdPath)

			if targetCmd != nil && (targetCmd.Run != nil || targetCmd.RunE != nil) {
				fmt.Printf("\nExecuting: %s\n", targetCmd.Use)
				if targetCmd.Run != nil {
					targetCmd.Run(targetCmd, []string{})
				}
			}
		}
	}
}

func collectCommandItems(cmd *spf13cobra.Command) []tui.MenuItem {
	items := make([]tui.MenuItem, 0)

	// Add current command if runnable
	if cmd.Run != nil || cmd.RunE != nil {
		items = append(items, tui.MenuItem{
			ID:          cmd.Name(),
			Label:       cmd.Name(),
			Description: cmd.Short,
		})
	}

	// Recursively collect subcommands
	for _, subCmd := range cmd.Commands() {
		if subCmd.Hidden {
			continue
		}

		// If subcommand has its own subcommands, recursively collect
		if len(subCmd.Commands()) > 0 {
			subItems := collectCommandItems(subCmd)
			items = append(items, subItems...)
		} else if subCmd.Run != nil || subCmd.RunE != nil {
			// Leaf command, show with parent path
			items = append(items, tui.MenuItem{
				ID:          subCmd.Name(),
				Label:       getCommandPath(subCmd),
				Description: subCmd.Short,
			})
		}
	}

	return items
}

func getCommandPath(cmd *spf13cobra.Command) string {
	if cmd.Parent() == nil || cmd.Parent().Parent() == nil {
		return cmd.Name()
	}

	pathParts := []string{cmd.Name()}
	current := cmd

	for current.Parent() != nil && current.Parent().Parent() != nil {
		current = current.Parent()
		pathParts = append([]string{current.Name()}, pathParts...)
	}

	return strings.Join(pathParts, " ")
}

func findCommand(root *spf13cobra.Command, pathParts []string) *spf13cobra.Command {
	if len(pathParts) == 0 {
		return root
	}

	current := root
	for _, part := range pathParts {
		found := false
		for _, cmd := range current.Commands() {
			if cmd.Name() == part {
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

func setupCommands() {
	// Add a large set of commands for demonstration
	addDevCommands()
	addDeployCommands()
	addTestCommands()
	addConfigCommands()
	addDatabaseCommands()
	addServerCommands()
}

func addDevCommands() {
	devCmd := &spf13cobra.Command{
		Use:   "dev",
		Short: "Development tools",
	}

	devBuild := &spf13cobra.Command{
		Use:   "build",
		Short: "Build the application",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Building application...")
		},
	}

	devTest := &spf13cobra.Command{
		Use:   "test",
		Short: "Run tests",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running tests...")
		},
	}

	devLint := &spf13cobra.Command{
		Use:   "lint",
		Short: "Run linter",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running linter...")
		},
	}

	devFmt := &spf13cobra.Command{
		Use:   "fmt",
		Short: "Format code",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Formatting code...")
		},
	}

	devCmd.AddCommand(devBuild, devTest, devLint, devFmt)
	rootCmd.AddCommand(devCmd)
}

func addDeployCommands() {
	deployCmd := &spf13cobra.Command{
		Use:   "deploy",
		Short: "Deployment operations",
	}

	deployStaging := &spf13cobra.Command{
		Use:   "staging",
		Short: "Deploy to staging environment",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Deploying to staging...")
		},
	}

	deployProd := &spf13cobra.Command{
		Use:   "production",
		Short: "Deploy to production environment",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Deploying to production...")
		},
	}

	deployRollback := &spf13cobra.Command{
		Use:   "rollback",
		Short: "Rollback deployment",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Rolling back deployment...")
		},
	}

	deployCmd.AddCommand(deployStaging, deployProd, deployRollback)
	rootCmd.AddCommand(deployCmd)
}

func addTestCommands() {
	testCmd := &spf13cobra.Command{
		Use:   "test",
		Short: "Testing commands",
	}

	testUnit := &spf13cobra.Command{
		Use:   "unit",
		Short: "Run unit tests",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running unit tests...")
		},
	}

	testIntegration := &spf13cobra.Command{
		Use:   "integration",
		Short: "Run integration tests",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running integration tests...")
		},
	}

	testE2E := &spf13cobra.Command{
		Use:   "e2e",
		Short: "Run end-to-end tests",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running E2E tests...")
		},
	}

	testCmd.AddCommand(testUnit, testIntegration, testE2E)
	rootCmd.AddCommand(testCmd)
}

func addConfigCommands() {
	configCmd := &spf13cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	configGet := &spf13cobra.Command{
		Use:   "get",
		Short: "Get configuration value",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Getting configuration...")
		},
	}

	configSet := &spf13cobra.Command{
		Use:   "set",
		Short: "Set configuration value",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Setting configuration...")
		},
	}

	configList := &spf13cobra.Command{
		Use:   "list",
		Short: "List all configuration",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Listing configuration...")
		},
	}

	configCmd.AddCommand(configGet, configSet, configList)
	rootCmd.AddCommand(configCmd)
}

func addDatabaseCommands() {
	dbCmd := &spf13cobra.Command{
		Use:   "db",
		Short: "Database operations",
	}

	dbMigrate := &spf13cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Running migrations...")
		},
	}

	dbSeed := &spf13cobra.Command{
		Use:   "seed",
		Short: "Seed database",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Seeding database...")
		},
	}

	dbReset := &spf13cobra.Command{
		Use:   "reset",
		Short: "Reset database",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Resetting database...")
		},
	}

	dbCmd.AddCommand(dbMigrate, dbSeed, dbReset)
	rootCmd.AddCommand(dbCmd)
}

func addServerCommands() {
	serverCmd := &spf13cobra.Command{
		Use:   "server",
		Short: "Server operations",
	}

	serverStart := &spf13cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Starting server...")
		},
	}

	serverStop := &spf13cobra.Command{
		Use:   "stop",
		Short: "Stop the server",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Stopping server...")
		},
	}

	serverRestart := &spf13cobra.Command{
		Use:   "restart",
		Short: "Restart the server",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Restarting server...")
		},
	}

	serverStatus := &spf13cobra.Command{
		Use:   "status",
		Short: "Check server status",
		Run: func(cmd *spf13cobra.Command, args []string) {
			fmt.Println("Checking server status...")
		},
	}

	serverCmd.AddCommand(serverStart, serverStop, serverRestart, serverStatus)
	rootCmd.AddCommand(serverCmd)
}
