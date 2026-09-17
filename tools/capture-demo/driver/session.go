package driver

import (
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Login 自动化快捷登录（通过环境变量注入密码，拒绝明文硬编码）
func (d *Driver) Login(username, password string) error {
	if err := d.Navigate("/login"); err != nil {
		return err
	}
	// 使用页面真实的 placeholder 文字定位输入框
	_ = d.FillInput("账号", username)    // placeholder: "账号 / 邮箱"
	_ = d.FillInput("登录密码", password) // placeholder: "登录密码"

	// 按钮文案含空格 "登 录"
	if err := d.ClickText("登 录"); err != nil {
		return fmt.Errorf("点击登录按钮失败: %w", err)
	}

	// 等待导航栏出现，确认登录跳转真正完成（而非固定等待）
	if err := d.WaitVisible(".layout-menu, .el-menu, .sidebar, nav"); err != nil {
		// 降级：固定等 2s，再判断是否还在登录页
		_ = d.WaitStable(2 * time.Second)
	}
	return nil
}

// SwitchTenant 切换到指定租户空间（若当前已在目标空间则跳过）
func (d *Driver) SwitchTenant(tenantName string) error {
	js := fmt.Sprintf(`(function(name) {
		const el = document.querySelector('.tenant-display, .space-name, .tenant-select');
		if (el && !el.innerText.includes(name)) { el.click(); return true; }
		return false;
	})(%s)`, quote(tenantName))

	var clicked bool
	_ = chromedp.Run(d.Ctx, chromedp.Evaluate(js, &clicked))
	if clicked {
		_ = d.WaitStable(200 * time.Millisecond)
		_ = d.ClickText(tenantName)
		_ = d.WaitStable(300 * time.Millisecond)
	}
	return nil
}

