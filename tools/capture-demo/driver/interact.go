package driver

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// ── 内嵌 JavaScript 片段 ────────────────────────────────────────────────────
// 所有注入到页面的 JS 集中在此处管理，与 Go 逻辑解耦。

// jsCleanStyles 禁用所有动画/过渡，隐藏临时提示层（幂等，可重复注入）
const jsCleanStyles = `(function() {
	if (document.getElementById('__cap_clean__')) return;
	const s = document.createElement('style');
	s.id = '__cap_clean__';
	s.textContent = '*, *::before, *::after { transition: none !important; animation: none !important; }' +
		'.el-message, .el-notification, .el-overlay-message-box, .v-popper__popper { display: none !important; }';
	(document.head || document.documentElement).appendChild(s);
})()`

// jsWaitStableTemplate MutationObserver 等待 DOM 静止（%d = 静默期毫秒数，出现两次）
const jsWaitStableTemplate = `new Promise(resolve => {
	let t;
	const ob = new MutationObserver(() => { clearTimeout(t); t = setTimeout(() => { ob.disconnect(); resolve(true); }, %d); });
	ob.observe(document.body || document.documentElement, { childList: true, subtree: true, attributes: true, characterData: true });
	t = setTimeout(() => { ob.disconnect(); resolve(true); }, %d);
})`

// jsClickText 按文本内容查找并点击元素（精确匹配优先，降级为包含匹配）
const jsClickText = `(function(target) {
	const all = Array.from(document.querySelectorAll('button, a, .el-button, .el-menu-item, .el-dropdown-menu__item, .el-sub-menu__title, span, div'));
	const text = t => (t.innerText || t.textContent || '').trim();
	const el = all.find(e => text(e) === target && e.offsetParent)
		|| all.find(e => text(e).includes(target) && e.offsetParent);
	if (!el) return false;
	el.scrollIntoView({ behavior: 'instant', block: 'center' });
	el.click();
	return true;
})(%s)`

// jsFillInput 按 placeholder/name/type 模糊定位输入框并填入值（触发 Vue 响应式）
const jsFillInput = `(function(query, val) {
	const q = query.toLowerCase();
	const inputs = Array.from(document.querySelectorAll('input, textarea'));
	let el = inputs.find(i => {
		const p = (i.getAttribute('placeholder') || '').toLowerCase();
		const n = (i.getAttribute('name') || '').toLowerCase();
		const t = (i.getAttribute('type') || '').toLowerCase();
		return p.includes(q) || n.includes(q) || t === q;
	});
	if (!el && (q === 'username' || q === 'user' || q === 'account')) el = inputs.find(i => i.type === 'text' || !i.type);
	if (!el && (q === 'password' || q === 'pass')) el = inputs.find(i => i.type === 'password');
	if (!el) return false;
	el.focus(); el.value = val;
	el.dispatchEvent(new Event('input', { bubbles: true }));
	el.dispatchEvent(new Event('change', { bubbles: true }));
	return true;
})(%s, %s)`

// jsFillBySelector 按 CSS 选择器精确填值（使用 nativeInputValueSetter 触发 Vue 响应式）
const jsFillBySelector = `(function(sel, val) {
	const root = document.querySelector(sel);
	if (!root) return false;
	const el = root.tagName === 'INPUT' ? root : root.querySelector('input, textarea');
	if (!el) return false;
	el.focus();
	const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
	setter?.call(el, val);
	el.dispatchEvent(new Event('input', { bubbles: true }));
	el.dispatchEvent(new Event('change', { bubbles: true }));
	return true;
})(%s, %s)`

// jsOpenSelect 按 placeholder/label 文字定位 El-Select 并展开下拉菜单
const jsOpenSelect = `(function(field) {
	const el = Array.from(document.querySelectorAll('.el-select, .el-select-v2')).find(s => {
		const ph = (s.querySelector('input')?.getAttribute('placeholder') || '').trim();
		const lb = (s.previousElementSibling?.innerText || '').trim();
		return ph.includes(field) || lb.includes(field);
	});
	if (!el) return false;
	(el.querySelector('input') || el).click();
	return true;
})(%s)`

// jsSelectDropdownItem 在已展开的 El-Select 弹出层中点击指定选项文字
const jsSelectDropdownItem = `(function(value) {
	const items = document.querySelectorAll(
		'.el-select-dropdown .el-select-dropdown__item, .el-select-dropdown__list li, .el-popper .el-select-dropdown__item'
	);
	const el = Array.from(items).find(i => {
		const t = (i.innerText || i.textContent || '').trim();
		return t === value || t.includes(value);
	});
	if (!el) return false;
	el.click();
	return true;
})(%s)`

// jsMaskText 遍历所有文本节点，将敏感内容替换为演示数据
const jsMaskText = `(function(masks) {
	function walk(node) {
		if (node.nodeType === Node.TEXT_NODE) {
			let v = node.nodeValue;
			for (const [k, r] of Object.entries(masks)) v = v.split(k).join(r);
			node.nodeValue = v;
		} else { node.childNodes.forEach(walk); }
	}
	walk(document.body);
})(%s)`

// jsScrollIntoView 将指定选择器的元素滚动到视口中央
const jsScrollIntoView = `(function(sel) {
	const el = document.querySelector(sel);
	if (!el) return false;
	el.scrollIntoView({ behavior: 'instant', block: 'center' });
	return true;
})(%s)`

// jsExistsSelector 判断选择器对应的元素是否存在
const jsExistsSelector = `!!document.querySelector(%s)`

// jsWaitToast 轮询等待成功 Toast 出现（最多 5s）
const jsWaitToast = `new Promise(resolve => {
	const check = () => {
		if (document.querySelector('.el-message--success, .el-notification--success')) return resolve(true);
		setTimeout(check, 100);
	};
	check();
	setTimeout(() => resolve(false), 5000);
})`

// ── 核心交互方法 ────────────────────────────────────────────────────────────

// Navigate 跳转到指定路径（相对路径自动拼接 BaseURL）并等待页面稳定
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

// InjectCleanStyles 注入禁用动画/隐藏 Toast 的样式（幂等）
func (d *Driver) InjectCleanStyles() error {
	return chromedp.Run(d.Ctx, chromedp.Evaluate(jsCleanStyles, nil))
}

// WaitStable 使用 MutationObserver 等待 DOM 静止（默认 300ms 无变动视为稳定）
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

// WaitVisible 等待指定 CSS 选择器的元素出现在视口（最长 15s）
func (d *Driver) WaitVisible(selector string) error {
	ctx, cancel := context.WithTimeout(d.Ctx, 15*time.Second)
	defer cancel()
	return chromedp.Run(ctx, chromedp.WaitVisible(selector, chromedp.ByQuery))
}

// ClickText 按可见文本查找并点击元素（精确匹配 → 包含匹配）
func (d *Driver) ClickText(text string) error {
	js := fmt.Sprintf(jsClickText, quote(text))
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("执行点击 %q 失败: %w", text, err)
	}
	if !ok {
		return fmt.Errorf("未找到包含文本 %q 的可见元素", text)
	}
	_ = d.InjectCleanStyles()
	return d.WaitStable()
}

// ClickSelector 按 CSS 选择器点击元素；元素不存在时快速返回 error（不等待超时）
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

// FillInput 按 placeholder/name/type 模糊定位输入框并填入值
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

// FillBySelector 按 CSS 选择器精确定位输入框并填入值（Vue 响应式安全）
func (d *Driver) FillBySelector(selector, value string) error {
	js := fmt.Sprintf(jsFillBySelector, quote(selector), quote(value))
	var ok bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(js, &ok)); err != nil {
		return fmt.Errorf("FillBySelector %q 失败: %w", selector, err)
	}
	if !ok {
		return fmt.Errorf("选择器 %q 未找到可填写的输入框", selector)
	}
	return nil
}

// SelectOption 操作 El-Select 下拉，按 field 定位下拉框，按 value 选中选项
func (d *Driver) SelectOption(field, value string) error {
	// 展开下拉
	var opened bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsOpenSelect, quote(field)), &opened)); err != nil || !opened {
		return fmt.Errorf("未找到 placeholder/label 含 %q 的下拉框", field)
	}
	_ = d.WaitStable(300 * time.Millisecond)

	// 点击目标选项（El-Select 弹层挂在 body 底部）
	var selected bool
	if err := chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsSelectDropdownItem, quote(value)), &selected)); err != nil || !selected {
		return fmt.Errorf("未在下拉列表中找到选项 %q", value)
	}
	return d.WaitStable(200 * time.Millisecond)
}

// MaskText 截图前零侵入脱敏：把所有文本节点中的真实内容替换为演示内容
func (d *Driver) MaskText(masks map[string]string) error {
	if len(masks) == 0 {
		return nil
	}
	return chromedp.Run(d.Ctx, chromedp.Evaluate(fmt.Sprintf(jsMaskText, marshalMasks(masks)), nil))
}

// ScrollIntoView 将指定选择器的元素滚动到视口中央
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

// WaitToast 等待成功 Toast 出现并消失（超时静默跳过，不报错）
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

// quote 将 Go 字符串转为 JS 中安全可用的双引号字面量（使用 %q 格式化）
func quote(s string) string {
	return fmt.Sprintf("%q", s)
}

// marshalMasks 将 Go map 序列化为 JS 对象字面量
func marshalMasks(masks map[string]string) string {
	var b strings.Builder
	b.WriteString("{")
	for k, v := range masks {
		b.WriteString(fmt.Sprintf("%q:%q,", k, v))
	}
	b.WriteString("}")
	return b.String()
}
