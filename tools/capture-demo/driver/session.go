package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Login 自动化快捷登录（通过环境变量注入密码，拒绝明文硬编码）
func (d *Driver) Login(username, password string) error {
	if err := d.Navigate("/login"); err != nil {
		return err
	}
	_ = d.FillInput("username", username)
	_ = d.FillInput("password", password)

	// 兼容"登录" / "登 录" 两种按钮文案
	if err := d.ClickText("登录"); err != nil {
		if err2 := d.ClickText("登 录"); err2 != nil {
			return err
		}
	}
	return d.WaitStable(500 * time.Millisecond)
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

// Capture 高清截取当前视口并保存到文件（自动创建目录）
func (d *Driver) Capture(outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	_ = d.InjectCleanStyles()
	_ = d.WaitStable(200 * time.Millisecond)

	var buf []byte
	if err := chromedp.Run(d.Ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return fmt.Errorf("截屏失败: %w", err)
	}
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("保存截图失败: %w", err)
	}
	fmt.Printf("   📸 已导出 → %s (%d KB)\n", strings.TrimPrefix(outputPath, "../../docs/public/"), len(buf)/1024)
	return nil
}
