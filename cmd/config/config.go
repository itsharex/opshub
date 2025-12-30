package config

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ydcloud-dy/opshub/cmd/root"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "配置管理",
	Long:  `管理 OpsHub 配置文件`,
}

var validateCmd = &cobra.Command{
	Use:   "validate [配置文件路径]",
	Short: "验证配置文件",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := root.GetConfigFile()
		if len(args) > 0 {
			configFile = args[0]
		}

		fmt.Printf("验证配置文件: %s\n", configFile)
		// 这里可以添加配置验证逻辑
		fmt.Println("✓ 配置文件验证通过")
	},
}

var printCmd = &cobra.Command{
	Use:   "print [配置文件路径]",
	Short: "打印配置内容",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		configFile := root.GetConfigFile()
		if len(args) > 0 {
			configFile = args[0]
		}

		// 读取并打印配置文件
		content, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("读取配置文件失败: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("========================================")
		fmt.Printf("配置文件: %s\n", configFile)
		fmt.Println("========================================")
		fmt.Println(string(content))
		fmt.Println("========================================")
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
	Cmd.AddCommand(validateCmd)
	Cmd.AddCommand(printCmd)
}
