package driver

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)


// Options 截屏驱动核心配置
type Options struct {
	BaseURL      string
	Headless     bool
	WindowWidth  int64
	WindowHeight int64
	DPR          float64 // 设备像素比（默认 2.0 高清视网膜）
	Timeout      time.Duration
	ChromePath   string
}

// Driver 自动化文档截屏驱动器
type Driver struct {
	Ctx         context.Context
	Cancel      context.CancelFunc
	AllocCtx    context.Context
	AllocCancel context.CancelFunc
	Opts        Options
	newTargetCh chan target.ID
}

// New 创建并初始化驱动器
func New(opts Options) (*Driver, error) {
	if opts.WindowWidth == 0 {
		opts.WindowWidth = 1600
	}
	if opts.WindowHeight == 0 {
		opts.WindowHeight = 1000
	}
	if opts.DPR == 0 {
		opts.DPR = 2.0
	}
	if opts.Timeout == 0 {
		opts.Timeout = 120 * time.Second
	}
	if opts.ChromePath == "" {
		// 优先检测 Mac 本地 Chrome
		if _, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"); err == nil {
			opts.ChromePath = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		}
	}

	execOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", opts.Headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.WindowSize(int(opts.WindowWidth), int(opts.WindowHeight)),
	)
	if opts.ChromePath != "" {
		execOpts = append(execOpts, chromedp.ExecPath(opts.ChromePath))
	}

	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(), execOpts...)
	ctx, cancel := chromedp.NewContext(allocCtx)

	d := &Driver{
		Ctx:         ctx,
		Cancel:      cancel,
		AllocCtx:    allocCtx,
		AllocCancel: allocCancel,
		Opts:        opts,
		newTargetCh: make(chan target.ID, 10),
	}

	// 监听新窗口/Tab 弹出事件
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*target.EventTargetCreated); ok {
			if e.TargetInfo.Type == "page" && e.TargetInfo.OpenerID != "" {
				d.newTargetCh <- e.TargetInfo.TargetID
			}
		}
	})

	// 初始化设备分辨率
	err := chromedp.Run(ctx,
		emulation.SetDeviceMetricsOverride(opts.WindowWidth, opts.WindowHeight, opts.DPR, false),
	)
	if err != nil {
		d.Close()
		return nil, fmt.Errorf("初始化设备分辨率失败: %w", err)
	}

	return d, nil
}

// Close 释放并关闭浏览器进程
func (d *Driver) Close() {
	if d.Cancel != nil {
		d.Cancel()
	}
	if d.AllocCancel != nil {
		d.AllocCancel()
	}
}

// Navigate 跳转到指定路径或完整 URL
func (d *Driver) Navigate(pathOrURL string) error {
	fullURL := pathOrURL
	if !strings.HasPrefix(pathOrURL, "http://") && !strings.HasPrefix(pathOrURL, "https://") {
		baseURL := strings.TrimRight(d.Opts.BaseURL, "/")
		path := strings.TrimLeft(pathOrURL, "/")
		fullURL = fmt.Sprintf("%s/%s", baseURL, path)
	}

	err := chromedp.Run(d.Ctx, chromedp.Navigate(fullURL))
	if err != nil {
		return fmt.Errorf("导航至 %s 失败: %w", fullURL, err)
	}

	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// InjectCleanStyles 注入无动画、无过渡与隐藏临时提示的全局净化样式（零侵入）
func (d *Driver) InjectCleanStyles() error {
	const cleanCSS = `
		(function() {
			if (document.getElementById('__doc_capture_clean_style__')) return;
			const style = document.createElement('style');
			style.id = '__doc_capture_clean_style__';
			style.textContent = ` + "`" + `
				*, *::before, *::after {
					transition: none !important;
					animation: none !important;
				}
				.el-message, .el-notification, .el-overlay-message-box, .v-popper__popper {
					display: none !important;
				}
			` + "`" + `;
			(document.head || document.documentElement).appendChild(style);
		})()
	`
	return chromedp.Run(d.Ctx, chromedp.Evaluate(cleanCSS, nil))
}

// WaitStable 使用 MutationObserver 等待 DOM 树静止（默认 300ms 无变动即视为稳定）
func (d *Driver) WaitStable(quietPeriod ...time.Duration) error {
	quietMs := 300
	if len(quietPeriod) > 0 {
		quietMs = int(quietPeriod[0].Milliseconds())
	}

	jsWait := fmt.Sprintf(`
		new Promise(resolve => {
			let timer = null;
			const observer = new MutationObserver(() => {
				clearTimeout(timer);
				timer = setTimeout(() => {
					observer.disconnect();
					resolve(true);
				}, %d);
			});
			observer.observe(document.body || document.documentElement, {
				childList: true,
				subtree: true,
				attributes: true,
				characterData: true
			});
			// 保底触发
			timer = setTimeout(() => {
				observer.disconnect();
				resolve(true);
			}, %d);
		})
	`, quietMs, quietMs)

	ctxTimeout, cancel := context.WithTimeout(d.Ctx, 15*time.Second)
	defer cancel()

	return chromedp.Run(ctxTimeout, chromedp.Evaluate(jsWait, nil))
}

// ClickText 按可见文本智能查找并点击元素（自动支持 button、span、a、menu-item 等）
func (d *Driver) ClickText(text string) error {
	jsClick := fmt.Sprintf(`
		(function() {
			const targetText = %q.trim();
			// 优先找按钮、链接、菜单项等常见可点击控件
			const candidates = Array.from(document.querySelectorAll('button, a, .el-button, .el-menu-item, .el-dropdown-menu__item, .el-sub-menu__title, span, div'));
			
			// 1. 优先精确匹配
			let match = candidates.find(el => {
				const t = (el.innerText || el.textContent || '').trim();
				return t === targetText && el.offsetParent !== null;
			});

			// 2. 其次包含匹配
			if (!match) {
				match = candidates.find(el => {
					const t = (el.innerText || el.textContent || '').trim();
					return t.includes(targetText) && el.offsetParent !== null;
				});
			}

			if (match) {
				match.scrollIntoView({ behavior: 'instant', block: 'center' });
				match.click();
				return true;
			}
			return false;
		})()
	`, text)

	var found bool
	err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsClick, &found))
	if err != nil {
		return fmt.Errorf("执行查找并点击 %q 失败: %w", text, err)
	}
	if !found {
		return fmt.Errorf("未在当前页面找到包含文本 %q 的可见可点击元素", text)
	}

	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// FillInput 智能查找输入框并填入值（自动触发 Vue 响应式 input / change 事件）
func (d *Driver) FillInput(queryOrTypeOrPlaceholder, value string) error {
	jsFill := fmt.Sprintf(`
		(function() {
			const target = %q.toLowerCase();
			const val = %q;
			const inputs = Array.from(document.querySelectorAll('input, textarea'));
			
			let input = inputs.find(i => {
				const p = (i.getAttribute('placeholder') || '').toLowerCase();
				const n = (i.getAttribute('name') || '').toLowerCase();
				const t = (i.getAttribute('type') || '').toLowerCase();
				return p.includes(target) || n.includes(target) || t === target;
			});

			if (!input && (target === 'username' || target === 'user' || target === 'account')) {
				input = inputs.find(i => i.type === 'text' || !i.type);
			}
			if (!input && (target === 'password' || target === 'pass')) {
				input = inputs.find(i => i.type === 'password');
			}

			if (input) {
				input.focus();
				input.value = val;
				input.dispatchEvent(new Event('input', { bubbles: true }));
				input.dispatchEvent(new Event('change', { bubbles: true }));
				return true;
			}
			return false;
		})()
	`, queryOrTypeOrPlaceholder, value)

	var filled bool
	err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsFill, &filled))
	if err != nil {
		return fmt.Errorf("填入表单输入框 %q 失败: %w", queryOrTypeOrPlaceholder, err)
	}
	if !filled {
		return fmt.Errorf("未找到匹配 %q 的输入框", queryOrTypeOrPlaceholder)
	}
	return nil
}

// Login 自动化快捷登录
func (d *Driver) Login(username, password string) error {
	if err := d.Navigate("/login"); err != nil {
		return err
	}

	_ = d.FillInput("username", username)
	_ = d.FillInput("password", password)

	if err := d.ClickText("登录"); err != nil {
		// 备用：尝试包含“登 录”
		if err2 := d.ClickText("登 录"); err2 != nil {
			return err
		}
	}

	return d.WaitStable(500 * time.Millisecond)
}

// SwitchTenant 智能切换指定租户空间
func (d *Driver) SwitchTenant(tenantName string) error {
	jsTenant := fmt.Sprintf(`
		(function() {
			const tenantEl = document.querySelector('.tenant-display, .space-name, .tenant-select');
			if (tenantEl && !tenantEl.innerText.includes(%q)) {
				tenantEl.click();
				return true;
			}
			return false;
		})()
	`, tenantName)

	var clicked bool
	_ = chromedp.Run(d.Ctx, chromedp.Evaluate(jsTenant, &clicked))
	if clicked {
		_ = d.WaitStable(200 * time.Millisecond)
		_ = d.ClickText(tenantName)
		_ = d.WaitStable(300 * time.Millisecond)
	}
	return nil
}

// WaitForPopup 触发操作并自动捕获新弹出的浏览器标签页或窗口，返回接管该新窗口的独立 Driver
func (d *Driver) WaitForPopup(triggerAction func() error) (*Driver, error) {
	if err := triggerAction(); err != nil {
		return nil, fmt.Errorf("触发弹窗动作失败: %w", err)
	}

	select {
	case targetID := <-d.newTargetCh:
		popupCtx, cancel := chromedp.NewContext(d.AllocCtx, chromedp.WithTargetID(targetID))
		popupDriver := &Driver{
			Ctx:         popupCtx,
			Cancel:      cancel,
			AllocCtx:    d.AllocCtx,
			AllocCancel: func() {}, // 共享 AllocContext，不单独 Cancel
			Opts:        d.Opts,
			newTargetCh: make(chan target.ID, 5),
		}

		// 注入高清配置与样式净化
		_ = chromedp.Run(popupCtx,
			emulation.SetDeviceMetricsOverride(d.Opts.WindowWidth, d.Opts.WindowHeight, d.Opts.DPR, false),
		)
		_ = popupDriver.InjectCleanStyles()
		_ = popupDriver.WaitStable(500 * time.Millisecond)
		return popupDriver, nil

	case <-time.After(15 * time.Second):
		return nil, fmt.Errorf("等待新窗口/标签页弹出超时")
	}
}

// WaitVisible 快捷等待元素可见
func (d *Driver) WaitVisible(selector string) error {
	ctxTimeout, cancel := context.WithTimeout(d.Ctx, 15*time.Second)
	defer cancel()
	return chromedp.Run(ctxTimeout, chromedp.WaitVisible(selector, chromedp.ByQuery))
}

// Capture 高清截取当前视口画面并自动保存到指定文件
func (d *Driver) Capture(outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	_ = d.InjectCleanStyles()
	_ = d.WaitStable(200 * time.Millisecond)

	var buf []byte
	err := chromedp.Run(d.Ctx, chromedp.CaptureScreenshot(&buf))
	if err != nil {
		return fmt.Errorf("执行截屏失败: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("保存截图文件失败: %w", err)
	}

	fmt.Printf("📸 [Capture] 已高清导出: %s (%d KB)\n", outputPath, len(buf)/1024)
	return nil
}

// ClickSelector 按 CSS 选择器点击元素；元素不存在时返回 error（调用方可选择忽略）
func (d *Driver) ClickSelector(selector string) error {
	// 先检测元素是否存在（避免 chromedp 默认 30s 超时等待）
	jsCheck := fmt.Sprintf(`!!document.querySelector(%q)`, selector)
	var exists bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsCheck, &exists)); err != nil || !exists {
		return fmt.Errorf("选择器 %q 对应的元素不存在", selector)
	}

	ctxTimeout, cancel := context.WithTimeout(d.Ctx, 5*time.Second)
	defer cancel()
	if err := chromedp.Run(ctxTimeout, chromedp.Click(selector, chromedp.ByQuery)); err != nil {
		return fmt.Errorf("点击选择器 %q 失败: %w", selector, err)
	}
	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// MaskText 截图前零侵入脱敏：把页面所有可见文本节点中匹配的子串替换为演示数据
// masks: key=真实内容, value=替换成的展示内容
func (d *Driver) MaskText(masks map[string]string) error {
	if len(masks) == 0 {
		return nil
	}

	// 构建 JS 替换逻辑：遍历所有文本节点，逐条替换
	var replacePairs strings.Builder
	replacePairs.WriteString("const masks = {\n")
	for k, v := range masks {
		replacePairs.WriteString(fmt.Sprintf("  %q: %q,\n", k, v))
	}
	replacePairs.WriteString("};\n")

	jsInject := fmt.Sprintf(`
		(function() {
			%s
			function walkTextNodes(node) {
				if (node.nodeType === Node.TEXT_NODE) {
					let val = node.nodeValue;
					for (const [real, fake] of Object.entries(masks)) {
						val = val.split(real).join(fake);
					}
					node.nodeValue = val;
				} else {
					for (const child of node.childNodes) {
						walkTextNodes(child);
					}
				}
			}
			walkTextNodes(document.body);
		})()
	`, replacePairs.String())

	return chromedp.Run(d.Ctx, chromedp.Evaluate(jsInject, nil))
}

// ScrollIntoView 将指定 CSS 选择器的元素滚动到视口中央
func (d *Driver) ScrollIntoView(selector string) error {
	jsScroll := fmt.Sprintf(`
		(function() {
			const el = document.querySelector(%q);
			if (el) {
				el.scrollIntoView({ behavior: 'instant', block: 'center' });
				return true;
			}
			return false;
		})()
	`, selector)

	var found bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsScroll, &found)); err != nil {
		return fmt.Errorf("scrollIntoView 失败: %w", err)
	}
	if !found {
		return fmt.Errorf("未找到选择器 %q 对应的元素", selector)
	}
	return d.WaitStable(100 * time.Millisecond)
}

// SelectOption 操作 El-Select 下拉组件，按选项文字选择目标项
// field: 用于定位下拉框的 placeholder 文字（如 "请选择认证类型"）
// value: 要选择的选项文字（如 "publickey"）
func (d *Driver) SelectOption(field, value string) error {
	// 第一步：找到下拉框并点击展开
	jsOpen := fmt.Sprintf(`
		(function() {
			const targets = Array.from(document.querySelectorAll(
				'.el-select, .el-select-v2, [class*="select"]'
			));
			// 按 placeholder 文字定位到正确的下拉框
			const el = targets.find(el => {
				const input = el.querySelector('input');
				const placeholder = (input?.getAttribute('placeholder') || '').trim();
				const label = (el.previousElementSibling?.innerText || '').trim();
				return placeholder.includes(%q) || label.includes(%q);
			});
			if (el) {
				(el.querySelector('input') || el).click();
				return true;
			}
			return false;
		})()
	`, field, field)

	var opened bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsOpen, &opened)); err != nil || !opened {
		return fmt.Errorf("未找到 placeholder 为 %q 的下拉框", field)
	}

	// 等待下拉菜单弹出
	_ = d.WaitStable(300 * time.Millisecond)

	// 第二步：在弹出的选项列表中点击目标选项
	// El-Select 的选项渲染在 body 底部的 .el-select-dropdown 中（脱离父级 DOM 树）
	jsSelect := fmt.Sprintf(`
		(function() {
			const dropdowns = document.querySelectorAll(
				'.el-select-dropdown .el-select-dropdown__item, ' +
				'.el-select-dropdown__list li, ' +
				'.el-popper .el-select-dropdown__item'
			);
			const target = Array.from(dropdowns).find(item => {
				const t = (item.innerText || item.textContent || '').trim();
				return t === %q || t.includes(%q);
			});
			if (target) {
				target.click();
				return true;
			}
			return false;
		})()
	`, value, value)

	var selected bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsSelect, &selected)); err != nil || !selected {
		return fmt.Errorf("未找到选项 %q", value)
	}
	return d.WaitStable(200 * time.Millisecond)
}

// FillBySelector 按 CSS 选择器精确定位输入框并填入值（兼容 Vue 响应式）
// 适用于多个同名 placeholder 字段共存的复杂表单
func (d *Driver) FillBySelector(selector, value string) error {
	jsFill := fmt.Sprintf(`
		(function() {
			const el = document.querySelector(%q);
			if (!el) return false;
			// 如果是 el-input，找到内部的 input 元素
			const input = el.tagName === 'INPUT' ? el : el.querySelector('input, textarea');
			if (!input) return false;
			input.focus();
			// 清空原有内容
			const nativeInputValueSetter = Object.getOwnPropertyDescriptor(
				window.HTMLInputElement.prototype, 'value'
			)?.set;
			nativeInputValueSetter?.call(input, %q);
			input.dispatchEvent(new Event('input', { bubbles: true }));
			input.dispatchEvent(new Event('change', { bubbles: true }));
			return true;
		})()
	`, selector, value)

	var filled bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(jsFill, &filled)); err != nil {
		return fmt.Errorf("FillBySelector %q 失败: %w", selector, err)
	}
	if !filled {
		return fmt.Errorf("选择器 %q 未找到可填写的输入框", selector)
	}
	return nil
}

// WaitToast 等待成功提示（el-message--success）出现并消失
// 用于确认提交/创建操作已被后端接受
// timeout: 最长等待时间，超时后不报错（降级为静默跳过）
func (d *Driver) WaitToast(timeout time.Duration) error {
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	jsWait := `
		new Promise(resolve => {
			const check = () => {
				const toast = document.querySelector('.el-message--success, .el-notification--success');
				if (toast) {
					resolve(true);
				} else {
					setTimeout(check, 100);
				}
			};
			check();
			// 保底超时
			setTimeout(() => resolve(false), 5000);
		})
	`

	ctxTimeout, cancel := context.WithTimeout(d.Ctx, timeout)
	defer cancel()

	var appeared bool
	_ = chromedp.Run(ctxTimeout, chromedp.Evaluate(jsWait, &appeared))

	if appeared {
		// Toast 出现了，等它自动消失（el-message 默认 3s 后消失）
		_ = d.WaitStable(400 * time.Millisecond)
	}
	// Toast 不出现也不报错，调用方已通过 WaitSelector 验证结果
	return nil
}

