// This example shows how to use the decorator pattern
// to enhance existing cobra commands WITHOUT changing imports.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra" // 保持原始导入不变！

	// 导入增强包，使用别名避免冲突
	cobrax "github.com/ZHLX2005/cobra/cobra"
)

var (
	apiKey  string
	baseURL string
	model   string
	stream  bool

	// 文件服务器配置
	fileServerURL  string
	fileServerToken string
)

var rootCmd = &cobra.Command{
	Use:   "glm4v",
	Short: "GLM-4.5V 多模态大模型命令行工具",
	Long: `GLM-4.5V 是智谱基于 MOE 架构的视觉推理模型。

使用方式：
1. 在当前目录创建 aitask.md 文件
2. 使用 {{.@文件路径}} 语法引用文件
3. 直接运行 glm4v

支持文件类型：
- 文本文件：.txt, .md, .go, .js, .py, .java, .c, .cpp, .html, .css, .json, .yaml, .xml 等
- 图片文件：.png, .jpg, .jpeg, .gif, .webp, .bmp 等
- 视频文件：.mp4, .avi, .mov, .mkv, .webm 等
- 文档文件：.pdf, .doc, .docx, .xls, .xlsx, .ppt, .pptx 等`,
	Run: func(cmd *cobra.Command, args []string) {
		// 零参数执行 - 自动查找 aitask.md
		wd, _ := os.Getwd()
		fmt.Printf("Working in: %s\n", wd)
		fmt.Printf("API Key: %s\n", apiKey)
		fmt.Printf("Model: %s\n", model)
		fmt.Println("执行任务...")
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "d237351671da318126fb5bd2f1372a08.EdkVfX8wE0JtcZpP", "API Key (必填)")
	rootCmd.PersistentFlags().StringVar(&baseURL, "base-url", "https://open.bigmodel.cn/api/paas/v4", "API Base URL")
	rootCmd.PersistentFlags().StringVar(&model, "model", "glm-4.5v", "模型名称")
	rootCmd.PersistentFlags().BoolVar(&stream, "stream", true, "是否使用流式输出")

	// 文件服务器配置
	rootCmd.PersistentFlags().StringVar(&fileServerURL, "file-server-url", "http://139.9.42.203:8988", "文件服务器地址")
	rootCmd.PersistentFlags().StringVar(&fileServerToken, "file-server-token", "123456", "文件服务器认证Token")

	// 从环境变量读取 API Key
	if apiKey == "" {
		apiKey = os.Getenv("GLM_API_KEY")
	}
}

func main() {
	// ============================================================
	// 关键：使用装饰器增强 rootCmd
	// 无需修改任何现有代码！
	// ============================================================
	enhancedRootCmd := cobrax.Enhance(rootCmd,
		cobrax.WithEnhanceTUIEnabled(true), // 启用 TUI
		cobrax.WithEnhanceTheme("dracula"),  // 使用 dracula 主题
		cobrax.WithEnhanceTUIConfirm(true),  // 执行前确认
	)

	// 执行增强后的命令
	if err := enhancedRootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
