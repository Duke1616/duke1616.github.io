package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// 全局配置实例
var globalConfig struct {
	BaseURL     string
	Headless    bool
	OutputDir   string
	Username    string
	Password    string
	Tenant      string
	GlobalMasks map[string]string
}

// 默认生产环境数据脱敏字典（全场景自动生效）
var defaultGlobalMasks = map[string]string{
	"82.156.165.98":  "10.0.12.88",
	"12222":          "22",
	"linuxserver.io": "root",
	"openssh-server": "prod-app-01",
}

var rootCmd = &cobra.Command{
	Use:   "capture-demo",
	Short: "VitePress 文档自动化高清截图工具",
	Long: `capture-demo 是一款面向 EDOCS / ECMDB 生态系统文档的自动化无头/高清截屏套件。
基于 Chrome DevTools Protocol (CDP)，提供声明式步骤驱动与全自动生产敏感数据脱敏。`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// Execute CLI 入口函数
func Execute() {
	// 智能语法糖：若首个入参不是子命令或 Flag，自动前置插入 "run"（如: capture-demo plugin → capture-demo run plugin）
	args := os.Args[1:]
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		knownSubCommands := map[string]bool{
			"run": true, "exec": true,
			"list": true, "ls": true,
			"help": true, "completion": true,
		}
		if !knownSubCommands[args[0]] {
			rootCmd.SetArgs(append([]string{"run"}, args...))
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n❌ 执行异常: %v\n\n", err)
		os.Exit(1)
	}
}

func init() {
	// 注册全局持久化标志（所有子命令共享）
	rootCmd.PersistentFlags().StringVar(&globalConfig.BaseURL, "base-url", getEnvOr("DEMO_BASE_URL", "http://www.fleetops.top"), "目标系统访问地址")
	rootCmd.PersistentFlags().BoolVar(&globalConfig.Headless, "headless", true, "是否启用无头模式（调试时传 false 可打开 Chrome 界面）")
	rootCmd.PersistentFlags().StringVarP(&globalConfig.OutputDir, "output", "o", getEnvOr("OUTPUT_BASE", "../../docs/public"), "截图保存基准目录")
	rootCmd.PersistentFlags().StringVarP(&globalConfig.Username, "user", "u", getEnvOr("DEMO_USER", "admin"), "认证用户名")
	rootCmd.PersistentFlags().StringVarP(&globalConfig.Tenant, "tenant", "t", getEnvOr("DEMO_TENANT", "默认租户空间"), "目标租户空间名称")

	// 注册子命令
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(runCmd)
}

// prepareConfig 准备并归一化运行配置，完成敏感规则融合与校验
func prepareConfig() error {
	// 密码统一从环境变量注入，避免命令行明文参数泄露
	globalConfig.Password = getEnvOr("DEMO_PASS", "")
	if globalConfig.Password == "" {
		return fmt.Errorf("缺少登录密码凭据，请通过环境变量 DEMO_PASS 提供（例：DEMO_PASS=xxx go run . run plugin）")
	}

	// 路径规范化
	absOutput, err := filepath.Abs(globalConfig.OutputDir)
	if err == nil {
		globalConfig.OutputDir = absOutput
	}

	// 融合内置与环境变量自定义的脱敏规则
	customMasks := parseCustomMasks(os.Getenv("DEMO_MASKS"))
	globalConfig.GlobalMasks = lo.Assign(defaultGlobalMasks, customMasks)

	return nil
}

// parseCustomMasks 解析外部注入的 JSON 格式脱敏字典
func parseCustomMasks(rawJSON string) map[string]string {
	if strings.TrimSpace(rawJSON) == "" {
		return nil
	}
	var res map[string]string
	if err := json.Unmarshal([]byte(rawJSON), &res); err != nil {
		fmt.Printf("⚠️  解析 DEMO_MASKS 环境变量失败: %v\n", err)
		return nil
	}
	return res
}

// getEnvOr 读取环境变量，未提供时返回默认值
func getEnvOr(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
