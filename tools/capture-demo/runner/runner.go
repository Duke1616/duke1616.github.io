package runner

import (
	"capture-demo/driver"
	"capture-demo/scenarios"
	"fmt"
	"strings"
	"time"
)

// Run 执行一个完整的 Scenario，处理所有步骤与原子动作
// outputBase 是截图输出的基准目录（绝对路径或相对路径），
// 例如 "../../docs/public"，最终路径 = outputBase + "/" + step.Output
func Run(bot *driver.Driver, scenario scenarios.Scenario, outputBase string) {
	fmt.Printf("\n🎬 开始执行场景: [%s]（共 %d 步）\n", scenario.Name, len(scenario.Steps))
	fmt.Println(strings.Repeat("─", 60))

	// currentBot 指向当前操作的 Driver 实例
	// 当某步骤触发弹窗并通过 InPopup 切换上下文后，后续步骤自动在新 Tab 中执行
	currentBot := bot
	var popupBot *driver.Driver

	for i, step := range scenario.Steps {
		fmt.Printf("\n📋 步骤 %d/%d: %s\n", i+1, len(scenario.Steps), step.Title)

		for _, action := range step.Actions {
			if err := executeAction(currentBot, bot, action, &currentBot, &popupBot); err != nil {
				// 单步失败不中断整个场景，打印警告后继续
				fmt.Printf("   ⚠️  动作 [%s] 执行失败: %v\n", action.Type, err)
			}
		}

		// 执行完所有动作后截图
		outputPath := strings.TrimRight(outputBase, "/") + "/" + strings.TrimLeft(step.Output, "/")
		if err := currentBot.Capture(outputPath); err != nil {
			fmt.Printf("   ❌ 截图失败: %v\n", err)
		} else {
			fmt.Printf("   ✅ 截图完成 → %s\n", outputPath)
		}
	}

	// 场景结束后关闭弹窗 Driver（若存在），归还控制权给主 bot
	if popupBot != nil {
		popupBot.Close()
		popupBot = nil
		currentBot = bot
	}

	fmt.Printf("\n🎉 场景 [%s] 全部完成\n", scenario.Name)
	fmt.Println(strings.Repeat("─", 60))
}

// executeAction 解释执行单个原子动作
func executeAction(
	currentBot *driver.Driver,
	mainBot *driver.Driver,
	action scenarios.Action,
	currentBotPtr **driver.Driver,
	popupBotPtr **driver.Driver,
) error {
	switch action.Type {

	case scenarios.Navigate:
		fmt.Printf("   → 导航至: %s\n", action.Path)
		return currentBot.Navigate(action.Path)

	case scenarios.ClickText:
		fmt.Printf("   → 点击文本: %q\n", action.Text)
		if action.InPopup {
			// InPopup=true：点击后等待新 Tab 弹出，后续步骤切换到新 Tab 中执行
			fmt.Printf("   → 等待新 Tab 弹出...\n")
			popupBot, err := mainBot.WaitForPopup(func() error {
				return currentBot.ClickText(action.Text)
			})
			if err != nil {
				return fmt.Errorf("等待弹窗失败: %w", err)
			}
			*popupBotPtr = popupBot
			*currentBotPtr = popupBot
			fmt.Printf("   → 已切换到新 Tab 上下文\n")
			return nil
		}
		return currentBot.ClickText(action.Text)

	case scenarios.ClickSelector:
		// 选择器不存在时静默跳过（拓扑居中按钮不一定存在）
		fmt.Printf("   → 点击选择器（静默）: %s\n", action.Selector)
		if action.InPopup {
			popupBot, err := mainBot.WaitForPopup(func() error {
				return currentBot.ClickSelector(action.Selector)
			})
			if err != nil {
				fmt.Printf("   → 新 Tab 未弹出，继续原窗口\n")
				return nil
			}
			*popupBotPtr = popupBot
			*currentBotPtr = popupBot
			return nil
		}
		// 静默：忽略错误
		_ = currentBot.ClickSelector(action.Selector)
		return nil

	case scenarios.WaitVisible:
		fmt.Printf("   → 等待元素可见: %s\n", action.Selector)
		return currentBot.WaitVisible(action.Selector)

	case scenarios.Fill:
		fmt.Printf("   → 填写表单字段 %q = %q\n", action.Field, action.Value)
		return currentBot.FillInput(action.Field, action.Value)

	case scenarios.Mask:
		fmt.Printf("   → 执行脱敏替换（%d 条）\n", len(action.Masks))
		return currentBot.MaskText(action.Masks)

	case scenarios.Sleep:
		fmt.Printf("   → 等待 %d ms\n", action.Ms)
		time.Sleep(time.Duration(action.Ms) * time.Millisecond)
		return nil

	case scenarios.Scroll:
		if action.Selector != "" {
			fmt.Printf("   → 滚动到元素: %s\n", action.Selector)
			return currentBot.ScrollIntoView(action.Selector)
		}
		return nil

	default:
		return fmt.Errorf("未知动作类型: %s", action.Type)
	}
}
