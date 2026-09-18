package cmd

import (
	"fmt"
	"strings"
	"time"

	"capture-demo/driver"
	"capture-demo/runner"
	"capture-demo/scenarios"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var runAllFlag bool

var runCmd = &cobra.Command{
	Use:     "run [scenario...]",
	Aliases: []string{"exec"},
	Short:   "执行指定的一个或多个截图场景",
	Long: `执行已注册的文档截图场景。可指定单个或多个场景名称，如：
  capture-demo run plugin
  capture-demo run --all`,
	RunE: runScenarios,
}

func init() {
	runCmd.Flags().BoolVarP(&runAllFlag, "all", "a", false, "执行所有已注册的场景")
}

// runScenarios 核心场景执行编排（供 runCmd 与 rootCmd 复用）
func runScenarios(cmd *cobra.Command, args []string) error {
	// 1. 配置加载与校验
	if err := prepareConfig(); err != nil {
		return err
	}

	// 2. 匹配目标场景列表
	targets, err := resolveTargets(args, runAllFlag)
	if err != nil {
		return err
	}

	// 3. 构造 CDP 浏览器驱动
	bot, err := driver.New(driver.Options{
		BaseURL:      globalConfig.BaseURL,
		Headless:     globalConfig.Headless,
		WindowWidth:  1600,
		WindowHeight: 1000,
		DPR:          2.0, // 2.0x Retina 视网膜高清采样
	})
	if err != nil {
		return fmt.Errorf("初始化截屏驱动失败: %w", err)
	}
	defer bot.Close()

	// 4. 统一登录认证
	fmt.Printf("[info] 正在登录: %s... ", globalConfig.Username)
	if err := bot.Login(globalConfig.Username, globalConfig.Password); err != nil {
		return fmt.Errorf("登录失败: %w", err)
	}
	fmt.Println("成功")

	// 5. 租户空间切换
	if globalConfig.Tenant != "" {
		fmt.Printf("[info] 切换租户: %s\n", globalConfig.Tenant)
		if err := bot.SwitchTenant(globalConfig.Tenant); err != nil {
			fmt.Printf("[warn] 切换租户失败: %v（继续执行）\n", err)
		}
	}

	// 6. 执行目标场景（各场景独立脱敏，互不干扰污染）
	totalSteps := lo.SumBy(targets, func(s scenarios.Scenario) int {
		return len(s.Steps)
	})
	fmt.Printf("[info] 开始执行截图任务: 场景数=%d, 总步骤数=%d\n", len(targets), totalSteps)
	start := time.Now()

	for _, sc := range targets {
		if len(globalConfig.ExtraMasks) > 0 {
			sc.Masks = lo.Assign(sc.Masks, globalConfig.ExtraMasks)
		}
		runner.Run(bot, sc, globalConfig.OutputDir)
	}

	duration := time.Since(start).Round(time.Millisecond)
	fmt.Printf("\n[done] 所有场景执行完成，总耗时: %v\n", duration)
	return nil
}

// resolveTargets 根据命令行入参匹配目标场景
func resolveTargets(args []string, runAll bool) ([]scenarios.Scenario, error) {
	all := scenarios.All()
	if runAll || len(args) == 0 || lo.Contains(args, "all") {
		return all, nil
	}

	var targets []scenarios.Scenario
	for _, name := range args {
		sc, ok := scenarios.Get(name)
		if !ok {
			return nil, fmt.Errorf("未找到场景 %q，当前可用场景: [%s]", name, strings.Join(scenarios.ListNames(), ", "))
		}
		targets = append(targets, sc)
	}
	return targets, nil
}
