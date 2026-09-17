package runner

import (
	"capture-demo/driver"
	"capture-demo/scenarios"
	"fmt"
	"strings"
	"time"

	"github.com/samber/lo"
)

// Run 执行一个完整的 Scenario
// outputBase: 截图输出基准目录（如 "../../docs/public"）
func Run(bot *driver.Driver, scenario scenarios.Scenario, outputBase string) {
	fmt.Printf("\n🎬 场景 [%s]（共 %d 步）\n%s\n", scenario.Name, len(scenario.Steps), divider())

	ctx := newRunContext(bot, outputBase)

	for i, step := range scenario.Steps {
		fmt.Printf("\n📋 %d/%d  %s\n", i+1, len(scenario.Steps), step.Title)
		ctx.runStep(step)
	}

	ctx.cleanup()
	fmt.Printf("\n✅ 场景 [%s] 完成\n%s\n", scenario.Name, divider())
}

// ── runContext 封装单次场景执行的可变状态 ──────────────────────────────────

// runContext 持有场景执行期间的上下文：当前激活窗口 + 弹窗窗口
type runContext struct {
	main       *driver.Driver // 主窗口（始终保留）
	active     *driver.Driver // 当前操作的窗口（可能切换到弹窗 Tab）
	popup      *driver.Driver // 弹出的子窗口（若存在）
	outputBase string
}

func newRunContext(bot *driver.Driver, outputBase string) *runContext {
	return &runContext{
		main:       bot,
		active:     bot,
		outputBase: strings.TrimRight(outputBase, "/"),
	}
}

// runStep 执行单个步骤：依序执行动作 → 截图
func (c *runContext) runStep(step scenarios.Step) {
	for _, action := range step.Actions {
		if err := c.dispatch(action); err != nil {
			// 单步失败不中断，打印警告后继续
			fmt.Printf("   ⚠️  [%s] %v\n", action.Type, err)
		}
	}
	if step.Output == "" {
		return // 允许纯操作步骤（无截图）
	}
	path := c.outputBase + "/" + strings.TrimLeft(step.Output, "/")
	if err := c.active.Capture(path); err != nil {
		fmt.Printf("   ❌ 截图失败: %v\n", err)
	}
}

// cleanup 场景结束后关闭弹窗、归还控制权到主窗口
func (c *runContext) cleanup() {
	if c.popup != nil {
		c.popup.Close()
		c.popup = nil
		c.active = c.main
	}
}

// switchToPopup 切换当前活动窗口到弹出的新 Tab
func (c *runContext) switchToPopup(popup *driver.Driver) {
	c.popup = popup
	c.active = popup
	fmt.Printf("   → 已切换到新 Tab\n")
}

// ── dispatch：将 Action 分发到对应处理逻辑 ─────────────────────────────────

// dispatch 根据 Action.Type 调用对应的驱动方法
func (c *runContext) dispatch(a scenarios.Action) error {
	switch a.Type {

	case scenarios.Navigate:
		fmt.Printf("   → 导航: %s\n", a.Path)
		return c.active.Navigate(a.Path)

	case scenarios.ClickText:
		return c.handleClick(a, func() error { return c.active.ClickText(a.Text) })

	case scenarios.ClickSelector:
		// 选择器不存在时静默跳过（如拓扑居中按钮）
		return c.handleClick(a, func() error { return c.active.ClickSelector(a.Selector) })

	case scenarios.WaitVisible:
		fmt.Printf("   → 等待: %s\n", a.Selector)
		return c.active.WaitVisible(a.Selector)

	case scenarios.Fill:
		fmt.Printf("   → 填写 [%s] = %q\n", a.Field, a.Value)
		return c.active.FillInput(a.Field, a.Value)

	case scenarios.FillBySelector:
		fmt.Printf("   → 填写 [%s] = %q\n", a.Selector, a.Value)
		return c.active.FillBySelector(a.Selector, a.Value)

	case scenarios.SelectOption:
		fmt.Printf("   → 下拉 [%s] → %q\n", a.Field, a.Value)
		return c.active.SelectOption(a.Field, a.Value)

	case scenarios.Submit:
		return c.handleSubmit(a)

	case scenarios.WaitToast:
		fmt.Printf("   → 等待 Toast\n")
		return c.active.WaitToast(5 * time.Second)

	case scenarios.Mask:
		fmt.Printf("   → 脱敏（%d 条）\n", len(a.Masks))
		return c.active.MaskText(a.Masks)

	case scenarios.Sleep:
		fmt.Printf("   → 等待 %d ms\n", a.Ms)
		time.Sleep(time.Duration(a.Ms) * time.Millisecond)
		return nil

	case scenarios.Scroll:
		fmt.Printf("   → 滚动到: %s\n", a.Selector)
		return c.active.ScrollIntoView(a.Selector)

	default:
		return fmt.Errorf("未知动作类型: %s", a.Type)
	}
}

// handleClick 统一处理点击动作，含 InPopup 切换逻辑
func (c *runContext) handleClick(a scenarios.Action, clickFn func() error) error {
	// lo.Ternary: InPopup=false 时用 Text，否则用 Selector 作为日志标识
	label := lo.Ternary(a.Text != "", a.Text, a.Selector)
	fmt.Printf("   → 点击: %s\n", label)

	if !a.InPopup {
		err := clickFn()
		if err != nil && a.Type == scenarios.ClickSelector {
			return nil // ClickSelector：元素不存在时静默跳过
		}
		return err
	}

	// InPopup：等待新 Tab 弹出后切换上下文
	popup, err := c.main.WaitForPopup(clickFn)
	if err != nil {
		if a.Type == scenarios.ClickSelector {
			fmt.Printf("   → 新 Tab 未弹出，继续当前窗口\n")
			return nil
		}
		return fmt.Errorf("等待弹窗失败: %w", err)
	}
	c.switchToPopup(popup)
	return nil
}

// handleSubmit 点击提交按钮 → 等待 Toast → 可选等待后置选择器
func (c *runContext) handleSubmit(a scenarios.Action) error {
	fmt.Printf("   → 提交: 点击 %q\n", a.Text)
	if err := c.active.ClickText(a.Text); err != nil {
		return err
	}
	_ = c.active.WaitToast(5 * time.Second)

	if a.WaitSelector != "" {
		fmt.Printf("   → 等待结果: %s\n", a.WaitSelector)
		if err := c.active.WaitVisible(a.WaitSelector); err != nil {
			fmt.Printf("   ⚠️  后置选择器 %q 未出现（继续）\n", a.WaitSelector)
		}
	}
	return nil
}

// ── 工具函数 ───────────────────────────────────────────────────────────────

func divider() string { return strings.Repeat("─", 56) }
