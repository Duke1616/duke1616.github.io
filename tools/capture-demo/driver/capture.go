package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

// Capture 高清截取当前视口并保存到文件（自动创建目录，自动执行全局数据脱敏）
func (d *Driver) Capture(outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}
	_ = d.InjectCleanStyles()

	// 截图前自动应用全局脱敏规则（零侵入无感替换生产敏感信息）
	if len(d.GlobalMasks) > 0 {
		_ = d.MaskText(d.GlobalMasks)
	}

	_ = d.WaitStable(200 * time.Millisecond)

	var buf []byte
	if err := chromedp.Run(d.Ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return fmt.Errorf("截屏失败: %w", err)
	}
	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("保存截图失败: %w", err)
	}
	fmt.Printf("   -> 已保存: %s (%d KB)\n", strings.TrimPrefix(outputPath, "../../docs/public/"), len(buf)/1024)
	return nil
}

// PageTitle 返回当前页面 <title>
func (d *Driver) PageTitle() (string, error) {
	var title string
	err := chromedp.Run(d.Ctx, chromedp.Title(&title))
	return title, err
}

// BodyText 返回 body 可见文字的前 n 个字符（用于路由探测与断言）
func (d *Driver) BodyText(n int) (string, error) {
	var text string
	err := chromedp.Run(d.Ctx, chromedp.Evaluate(
		fmt.Sprintf(`document.body.innerText.replace(/\s+/g,' ').trim().slice(0,%d)`, n),
		&text,
	))
	return text, err
}
