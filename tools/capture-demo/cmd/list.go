package cmd

import (
	"fmt"
	"strings"

	"capture-demo/scenarios"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "列出所有已注册的文档截图场景",
	Run: func(cmd *cobra.Command, args []string) {
		all := scenarios.All()
		fmt.Println("已注册的截图场景列表:")
		fmt.Println(strings.Repeat("-", 74))
		for _, s := range all {
			desc := lo.Ternary(s.Description != "", s.Description, "-")
			fmt.Printf("  %-14s  (%2d 步)  %s\n", s.Name, len(s.Steps), desc)
		}
		fmt.Println(strings.Repeat("-", 74))
		fmt.Println("运行命令: capture-demo run <场景名> 或 capture-demo run --all")
	},
}
