package main

import (
	"capture-demo/driver"
	"capture-demo/runner"
	"capture-demo/scenarios"
	"fmt"
	"log"
	"os"
)

func main() {
	fmt.Println("🚀 启动声明式文档截图流水线...")

	// 1. 初始化截屏驱动（Headless Chrome，2.0x 视网膜高清）
	bot, err := driver.New(driver.Options{
		BaseURL:      getEnvOr("DEMO_BASE_URL", "http://www.fleetops.top"),
		Headless:     true,
		WindowWidth:  1600,
		WindowHeight: 1000,
		DPR:          2.0,
	})
	if err != nil {
		log.Fatalf("❌ 驱动初始化失败: %v", err)
	}
	defer bot.Close()

	// 2. 自动登录（通过环境变量注入，杜绝明文密码硬编码）
	username := getEnvOr("DEMO_USER", "admin")
	password := getEnvOr("DEMO_PASS", "")
	if password == "" {
		log.Fatal("❌ 请通过环境变量 DEMO_PASS 提供登录密码")
	}

	fmt.Println("🔑 执行登录...")
	if err := bot.Login(username, password); err != nil {
		log.Fatalf("❌ 登录失败: %v", err)
	}
	fmt.Println("✅ 登录成功")

	// 3. 切换租户空间（可选）
	if tenant := os.Getenv("DEMO_TENANT"); tenant != "" {
		fmt.Printf("🏢 切换到租户空间: %s\n", tenant)
		_ = bot.SwitchTenant(tenant)
	} else {
		_ = bot.SwitchTenant("默认租户空间")
	}

	// 4. 确定截图基准输出目录（相对本工具，指向 docs/public）
	outputBase := getEnvOr("OUTPUT_BASE", "../../docs/public")

	// ─────────────────────────────────────────────────────────────
	// 5. 声明式执行场景列表
	//    新增一篇文档的截图只需：
	//      a. 在 scenarios/ 包下创建 xxx.go 文件
	//      b. 在下方列表中追加一行 runner.Run(bot, scenarios.XxxScenario, outputBase)
	// ─────────────────────────────────────────────────────────────
	runner.Run(bot, scenarios.PluginScenario, outputBase)

	fmt.Println("\n🎊 全部场景截图完成！")
}

// getEnvOr 读取环境变量，若未设置则返回默认值
func getEnvOr(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
