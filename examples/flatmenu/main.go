// This example demonstrates the flat menu view feature
// All executable commands are displayed in a single list with their full paths
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	cobrax "github.com/ZHLX2005/cobrax/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flatmenu",
	Short: "Demo application showing flat menu view",
	Long: `This example demonstrates the flat menu view feature of cobrax.

When you run with --tui flag, all executable commands are displayed
in a single list with their full paths, making it easier to navigate
complex command hierarchies.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Root command executed (should not see this in TUI mode)")
	},
}

func init() {
	// Add some example flags
	rootCmd.Flags().StringP("output", "o", "text", "Output format")
	rootCmd.Flags().BoolP("verbose", "v", false, "Verbose output")
}

func main() {
	// Add various subcommands to demonstrate flat menu
	setupCommands()

	// Enhance with TUI
	enhancedCmd := cobrax.Enhance(rootCmd,
		cobrax.WithEnhanceTUIEnabled(true),
		cobrax.WithEnhanceTheme("dracula"),
		cobrax.WithEnhanceTUIConfirm(true),
	)

	if err := enhancedCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func setupCommands() {
	// Database commands
	dbCmd := &cobra.Command{
		Use:   "db",
		Short: "Database operations",
	}

	dbMigrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running database migrations...")
		},
	}
	dbMigrateCmd.Flags().String("direction", "up", "Migration direction (up/down)")

	dbSeedCmd := &cobra.Command{
		Use:   "seed",
		Short: "Seed database with initial data",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Seeding database...")
		},
	}
	dbSeedCmd.Flags().String("source", "data.json", "Seed data source")

	dbBackupCmd := &cobra.Command{
		Use:   "backup",
		Short: "Backup database",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Backing up database...")
		},
	}
	dbBackupCmd.Flags().String("output", "backup.sql", "Backup output file")

	dbCmd.AddCommand(dbMigrateCmd, dbSeedCmd, dbBackupCmd)

	// Server commands
	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "Server operations",
	}

	serverStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the server",
		Run: func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetString("port")
			fmt.Printf("Starting server on port %s...\n", port)
		},
	}
	serverStartCmd.Flags().StringP("port", "p", "8080", "Server port")

	serverStopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Stopping server...")
		},
	}

	serverRestartCmd := &cobra.Command{
		Use:   "restart",
		Short: "Restart the server",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Restarting server...")
		},
	}

	serverCmd.AddCommand(serverStartCmd, serverStopCmd, serverRestartCmd)

	// User commands
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "User management",
	}

	userListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Listing users...")
		},
	}
	userListCmd.Flags().StringP("format", "f", "table", "Output format (table/json)")

	userCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Run: func(cmd *cobra.Command, args []string) {
			username, _ := cmd.Flags().GetString("username")
			email, _ := cmd.Flags().GetString("email")
			fmt.Printf("Creating user: %s (%s)\n", username, email)
		},
	}
	userCreateCmd.Flags().StringP("username", "u", "", "Username (required)")
	userCreateCmd.Flags().StringP("email", "e", "", "Email address")
	_ = userCreateCmd.MarkFlagRequired("username")

	userDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a user",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Deleting user...")
		},
	}
	userDeleteCmd.Flags().BoolP("force", "f", false, "Force deletion without confirmation")

	userCmd.AddCommand(userListCmd, userCreateCmd, userDeleteCmd)

	// Config commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
	}

	configGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get configuration value",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Getting configuration...")
		},
	}

	configSetCmd := &cobra.Command{
		Use:   "set",
		Short: "Set configuration value",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Setting configuration...")
		},
	}

	configInitCmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize configuration",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Initializing configuration...")
		},
	}
	configInitCmd.Flags().BoolP("force", "f", false, "Overwrite existing config")

	configCmd.AddCommand(configGetCmd, configSetCmd, configInitCmd)

	// Add all commands to root
	rootCmd.AddCommand(dbCmd, serverCmd, userCmd, configCmd)
}
