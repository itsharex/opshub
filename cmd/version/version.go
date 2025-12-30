package version

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/ydcloud-dy/opshub/cmd/root"
)

var (
	// Version 版本号
	Version = "1.0.0"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GoVersion Go版本
	GoVersion = "unknown"
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Long:  `显示 OpsHub 的版本信息`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("========================================")
		fmt.Println("           OpsHub 运维管理平台")
		fmt.Println("========================================")
		fmt.Printf("版本号:     %s\n", Version)
		fmt.Printf("Git提交:    %s\n", GitCommit)
		fmt.Printf("构建时间:   %s\n", BuildTime)
		fmt.Printf("Go版本:     %s\n", GoVersion)
		fmt.Println("========================================")
	},
}

func init() {
	root.Cmd.AddCommand(Cmd)
}
