package main

import (
	"fmt"
	"os"

	"github.com/ZHLX2005/cobra/cobra"
	"github.com/ZHLX2005/cobra/tui/style"
)

func main() {
	// 创建根命令
	rootCmd := cobra.NewCommand("myapp",
		cobra.WithShort("My application"),
		cobra.WithLong("My application is a demo application for cobra-x."),
	)

	// 设置 TUI 配置
	rootCmd.SetTUIConfig(&cobra.TUIConfig{
		Enabled:              true,
		Theme:                style.NewTheme(style.ThemeDefault),
		ShowDescription:      true,
		ShowFlags:            true,
		ConfirmBeforeExecute: true,
	})

	// 添加 server 命令
	serverCmd := cobra.NewCommand("server",
		cobra.WithShort("Start the server"),
		cobra.WithLong("Start the server with the specified configuration."),
		cobra.WithRun(func(cmd *cobra.Command, args []string) {
			port, _ := cmd.Flags().GetString("port")
			host, _ := cmd.Flags().GetString("host")
			tls, _ := cmd.Flags().GetBool("tls")

			fmt.Printf("Starting server on %s:%s\n", host, port)
			if tls {
				fmt.Println("TLS enabled")
			}
		}),
	)

	// 添加 server 命令的 flags
	serverCmd.Command.Flags().StringP("port", "p", "8080", "Server port")
	serverCmd.Command.Flags().StringP("host", "H", "0.0.0.0", "Server host")
	serverCmd.Command.Flags().BoolP("tls", "t", false, "Enable TLS")
	serverCmd.Command.Flags().IntP("workers", "w", 4, "Number of worker threads")

	// 添加 client 命令
	clientCmd := cobra.NewCommand("client",
		cobra.WithShort("Start the client"),
		cobra.WithLong("Start the client with the specified configuration."),
		cobra.WithRun(func(cmd *cobra.Command, args []string) {
			server, _ := cmd.Flags().GetString("server")
			timeout, _ := cmd.Flags().GetInt("timeout")

			fmt.Printf("Connecting to server: %s\n", server)
			fmt.Printf("Timeout: %d seconds\n", timeout)
		}),
	)

	// 添加 client 命令的 flags
	clientCmd.Command.Flags().StringP("server", "s", "localhost:8080", "Server address")
	clientCmd.Command.Flags().IntP("timeout", "t", 30, "Connection timeout in seconds")

	// 添加 config 命令
	configCmd := cobra.NewCommand("config",
		cobra.WithShort("Manage configuration"),
		cobra.WithLong("Manage application configuration."),
	)

	// config init 子命令
	configInitCmd := cobra.NewCommand("init",
		cobra.WithShort("Initialize configuration"),
		cobra.WithRun(func(cmd *cobra.Command, args []string) {
			force, _ := cmd.Flags().GetBool("force")
			fmt.Println("Initializing configuration...")
			if force {
				fmt.Println("Force mode enabled")
			}
		}),
	)
	configInitCmd.Command.Flags().BoolP("force", "f", false, "Force overwrite existing config")

	// config show 子命令
	configShowCmd := cobra.NewCommand("show",
		cobra.WithShort("Show configuration"),
		cobra.WithRun(func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			fmt.Printf("Showing configuration in %s format\n", format)
		}),
	)
	configShowCmd.Command.Flags().StringP("format", "f", "yaml", "Output format (yaml, json)")

	configCmd.AddCommand(configInitCmd, configShowCmd)

	// 添加所有命令到根命令
	rootCmd.AddCommand(serverCmd, clientCmd, configCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
