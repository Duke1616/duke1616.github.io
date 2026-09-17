package driver

// ── 内嵌 JavaScript 脚本集合 ────────────────────────────────────────────────
// 将注入到页面的 JS 片段集中隔离于此，与 Go 交互方法解耦，保持业务代码清爽。

// jsCleanStyles 禁用所有过渡动画，隐藏临时提示层（幂等，可安全重复注入）
const jsCleanStyles = `(function() {
	if (document.getElementById('__cap_clean__')) return;
	const s = document.createElement('style');
	s.id = '__cap_clean__';
	s.textContent = '*, *::before, *::after { transition: none !important; animation: none !important; }' +
		'.el-message, .el-notification, .el-overlay-message-box, .v-popper__popper { display: none !important; }';
	(document.head || document.documentElement).appendChild(s);
})()`

// jsWaitStableTemplate MutationObserver 等待 DOM 静止（%d = 静默期毫秒数）
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
	if (!el && ['username','user','account','账号','邮箱'].some(k => q.includes(k)))
		el = inputs.find(i => (i.type === 'text' || !i.type) && i.offsetParent);
	if (!el && ['password','pass','密码'].some(k => q.includes(k)))
		el = inputs.find(i => i.type === 'password');
	if (!el) return false;
	el.focus();
	const setter = Object.getOwnPropertyDescriptor(window.HTMLInputElement.prototype, 'value')?.set;
	setter?.call(el, '');
	el.dispatchEvent(new Event('input', { bubbles: true }));
	setter?.call(el, val);
	el.dispatchEvent(new Event('input', { bubbles: true }));
	el.dispatchEvent(new Event('change', { bubbles: true }));
	return true;
})(%s, %s)`

// jsFillBySelector 按 CSS 选择器精确填值
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

// jsOpenSelect 定位并展开 El-Select 下拉
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

// jsSelectDropdownItem 选中下拉项
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

// jsMaskText 深度遍历所有文本节点替换敏感数据
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

// jsScrollIntoView 滚动元素到视口中心
const jsScrollIntoView = `(function(sel) {
	const el = document.querySelector(sel);
	if (!el) return false;
	el.scrollIntoView({ behavior: 'instant', block: 'center' });
	return true;
})(%s)`

// jsExistsSelector 判断元素是否存在
const jsExistsSelector = `!!document.querySelector(%s)`

// jsWaitToast 轮询等待成功消息
const jsWaitToast = `new Promise(resolve => {
	const check = () => {
		if (document.querySelector('.el-message--success, .el-notification--success')) return resolve(true);
		setTimeout(check, 100);
	};
	check();
	setTimeout(() => resolve(false), 5000);
})`

// jsHover 模拟鼠标悬停
const jsHover = `(function(target) {
	const all = Array.from(document.querySelectorAll('button, a, .el-button, .el-tooltip, [data-tooltip], span, div'));
	const text = t => (t.innerText || t.textContent || '').trim();
	const el = all.find(e => text(e) === target && e.offsetParent)
		|| all.find(e => text(e).includes(target) && e.offsetParent);
	if (!el) return false;
	el.dispatchEvent(new MouseEvent('mouseenter', { bubbles: true }));
	el.dispatchEvent(new MouseEvent('mouseover', { bubbles: true }));
	return true;
})(%s)`

// jsInterceptWindowOpen 拦截当前页面的 window.open 转为就地导航
const jsInterceptWindowOpen = `(function() {
	if (!window.__origOpen) {
		window.__origOpen = window.open;
		window.open = function(url) {
			if (url) { window.location.href = url; }
			return null;
		};
	}
})()`
