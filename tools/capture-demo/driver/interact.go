package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/samber/lo"
)

// ── 核心页面导航与状态监听 ──────────────────────────────────────────────────

// Navigate 跳转到指定路径（相对路径自动补全 BaseURL）并等待 DOM 稳定
func (d *Driver) Navigate(pathOrURL string) error {
	url := pathOrURL
	if !strings.HasPrefix(pathOrURL, "http://") && !strings.HasPrefix(pathOrURL, "https://") {
		url = strings.TrimRight(d.Opts.BaseURL, "/") + "/" + strings.TrimLeft(pathOrURL, "/")
	}
	if err := chromedp.Run(d.Ctx, chromedp.Navigate(url)); err != nil {
		return fmt.Errorf("导航至 %s 失败: %w", url, err)
	}
	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// InjectCleanStyles 注入无动画、隐藏 Toast 的纯净样式（幂等）
func (d *Driver) InjectCleanStyles() error {
	return chromedp.Run(d.Ctx, chromedp.Evaluate(jsCleanStyles, nil))
}

// WaitStable 使用 MutationObserver 监听 DOM 静止（默认 300ms 无变动判定为稳定）
func (d *Driver) WaitStable(quietPeriod ...time.Duration) error {
	ms := 300
	if len(quietPeriod) > 0 {
		ms = int(quietPeriod[0].Milliseconds())
	}
	js := fmt.Sprintf(jsWaitStableTemplate, ms, ms)
	ctx, cancel := context.WithTimeout(d.Ctx, 15*time.Second)
	defer cancel()
	return chromedp.Run(ctx, chromedp.Evaluate(js, nil))
}

// WaitVisible 等待指定 CSS 选择器的元素出现（最长 15s）
func (d *Driver) WaitVisible(selector string) error {
	ctx, cancel := context.WithTimeout(d.Ctx, 15*time.Second)
	defer cancel()
	return chromedp.Run(ctx, chromedp.WaitVisible(selector, chromedp.ByQuery))
}

// ── 语义化点击与交互 ────────────────────────────────────────────────────────

// ClickText 按可见文本查找并点击元素（精确匹配优先，自动降级为包含匹配）
func (d *Driver) ClickText(text string) error {
	js := fmt.Sprintf(jsClickText, quote(text))
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("点击文本 %q 失败: %w", text, err)
	}
	if !ok {
		return fmt.Errorf("未找到包含文本 %q 的可见元素", text)
	}
	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// ClickSelector 按 CSS 选择器点击元素（元素不存在时快速返回 error，不阻塞等待超时）
func (d *Driver) ClickSelector(selector string) error {
	var exists bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsExistsSelector, quote(selector)), &exists)); err != nil || !exists {
		return fmt.Errorf("选择器 %q 对应的元素不存在", selector)
	}
	ctx, cancel := context.WithTimeout(d.Ctx, 5*time.Second)
	defer cancel()
	if err := chromedp.Run(ctx, chromedp.Click(selector, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("点击选择器 %q 失败: %w", selector, err)
	}
	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// ClickTextInPlace 点击按钮同时拦截 window.open 使其在当前 Tab 导航（适用于 Web Shell / Web Sftp）
func (d *Driver) ClickTextInPlace(text string) error {
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsInterceptWindowOpen, nil)); err != nil {
		return fmt.Errorf("注入 window.open 拦截器失败: %w", err)
	}
	return d.ClickText(text)
}

// RestoreWindowOpen 恢复被拦截的 window.open
func (d *Driver) RestoreWindowOpen() {
	_ = chromedp.Run(d.Ctx, chromedp.Evaluate(`
		if (window.__origOpen) { window.open = window.__origOpen; delete window.__origOpen; }
	`, nil))
}

// Hover 鼠标悬停在包含指定文本的元素上（触发气泡 Tooltip 或二级菜单）
func (d *Driver) Hover(text string) error {
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsHover, quote(text)), &ok)); err != nil {
		return fmt.Errorf("hover 目标 %q 失败: %w", text, err)
	}
	return d.WaitStable(200 * time.Millisecond)
}

// ── 表单录入与组件操控 ──────────────────────────────────────────────────────

// FillInput 按 placeholder/name/type 模糊定位输入框并填入值（触发 Vue 响应式数据同步）
func (d *Driver) FillInput(field, value string) error {
	js := fmt.Sprintf(jsFillInput, quote(field), quote(value))
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("填写字段 %q 失败: %w", field, err)
	}
	if !ok {
		return fmt.Errorf("未找到匹配 %q 的输入框", field)
	}
	return nil
}

// FillBySelector 按 CSS 选择器精确输入
func (d *Driver) FillBySelector(selector, value string) error {
	js := fmt.Sprintf(jsFillBySelector, quote(selector), quote(value))
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("FillBySelector %q 失败: %w", selector, err)
	}
	if !ok {
		return fmt.Errorf("选择器 %q 未找到输入框", selector)
	}
	return nil
}

// SelectOption 操作 El-Select 下拉组件，按 field 定位并选中 value 选项
func (d *Driver) SelectOption(field, value string) error {
	var opened bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsOpenSelect, quote(field)), &opened)); err != nil || !opened {
		return fmt.Errorf("未找到 placeholder/label 含 %q 的下拉框", field)
	}
	_ = d.WaitStable(300 * time.Millisecond)

	var selected bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsSelectDropdownItem, quote(value)), &selected)); err != nil || !selected {
		return fmt.Errorf("未在下拉列表中找到选项 %q", value)
	}
	return d.WaitStable(200 * time.Millisecond)
}

// ── 数据脱敏、滚动与反馈等待 ────────────────────────────────────────────────

// MaskText 将指定文本节点中的敏感内容零侵入替换为演示数据
func (d *Driver) MaskText(masks map[string]string) error {
	if len(masks) == 0 {
		return nil
	}
	return chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsMaskText, marshalMasks(masks)), nil))
}

// ScrollIntoView 将指定 CSS 选择器的元素平滑滚动到视口中央
func (d *Driver) ScrollIntoView(selector string) error {
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsScrollIntoView, quote(selector)), &ok)); err != nil {
		return fmt.Errorf("scrollIntoView 失败: %w", err)
	}
	if !ok {
		return fmt.Errorf("未找到选择器 %q 对应的元素", selector)
	}
	return d.WaitStable(100 * time.Millisecond)
}

// WaitToast 等待操作成功 Toast 提示出现（超时静默跳过）
func (d *Driver) WaitToast(timeout time.Duration) error {
	if timeout == 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(d.Ctx, timeout)
	defer cancel()
	var appeared bool
	_ = chromedp.Run(ctx, chromedp.Evaluate(jsWaitToast, &appeared))
	if appeared {
		_ = d.WaitStable(400 * time.Millisecond)
	}
	return nil
}

// ── 辅助工具函数 ────────────────────────────────────────────────────────────

func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

func marshalMasks(masks map[string]string) string {
	pairs := lo.Map(lo.Entries(masks), func(e lo.Entry[string, string], _ int) string {
		return fmt.Sprintf("%q:%q", e.Key, e.Value)
	})
	return "{" + strings.Join(pairs, ",") + "}"
}
