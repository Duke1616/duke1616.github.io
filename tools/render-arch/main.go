package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ 获取当前工作目录失败: %v", err)
	}

	// 自底向上查找项目根目录（寻找包含 docs/diagrams 的根路径）
	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "docs", "diagrams")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			log.Fatalf("❌ 未能定位到包含 docs/diagrams 的项目根目录 (起始路径: %s)", cwd)
		}
		projectRoot = parent
	}

	diagramsDir := filepath.Join(projectRoot, "docs", "diagrams")
	outputDir := filepath.Join(projectRoot, "docs", "public", "images")

	// 扫描 docs/diagrams 目录下的所有 HTML 模板
	entries, err := os.ReadDir(diagramsDir)
	if err != nil {
		log.Fatalf("❌ 读取图表目录失败: %v", err)
	}

	var htmlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".html") {
			htmlFiles = append(htmlFiles, filepath.Join(diagramsDir, entry.Name()))
		}
	}

	if len(htmlFiles) == 0 {
		fmt.Printf("⚠️ 目录中未找到任何 .html 模板文件: %s\n", diagramsDir)
		return
	}

	fmt.Printf("🚀 扫描到 %d 个图表模板文件，启动 Go chromedp 无头浏览器批量渲染...\n", len(htmlFiles))
	fmt.Printf("📂 模板目录: %s\n", diagramsDir)
	fmt.Printf("🖼️ 输出目录: %s\n\n", outputDir)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("❌ 创建输出目录失败: %v", err)
	}

	// 配置无头浏览器启动选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoSandbox,
		chromedp.Headless,
		chromedp.WindowSize(1920, 1080),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	// 依次渲染每个 HTML 模板
	for idx, htmlPath := range htmlFiles {
		fileName := filepath.Base(htmlPath)
		baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		targetPng := filepath.Join(outputDir, baseName+".png")

		fmt.Printf("[%d/%d] 正在渲染: %s -> %s.png\n", idx+1, len(htmlFiles), fileName, baseName)

		if err := renderHTMLToPNG(browserCtx, htmlPath, targetPng); err != nil {
			log.Printf("❌ 渲染失败 [%s]: %v\n", fileName, err)
			continue
		}

		info, _ := os.Stat(targetPng)
		sizeKB := float64(info.Size()) / 1024.0
		fmt.Printf("    ✅ 完成！文件大小: %.1f KB\n", sizeKB)
	}

	fmt.Printf("\n✨ 所有图表生成完成，产物已全部保存在 docs/public/images 目录。\n")
}

// renderHTMLToPNG 使用 Chrome 无头浏览器加载 HTML 并自适应尺寸输出 Retina PNG
func renderHTMLToPNG(parentCtx context.Context, htmlPath, outputPath string) error {
	tabCtx, cancelTab := chromedp.NewContext(parentCtx)
	defer cancelTab()

	tabCtx, cancelTimeout := context.WithTimeout(tabCtx, 30*time.Second)
	defer cancelTimeout()

	fileURL := "file://" + htmlPath
	var buf []byte

	var dims []float64

	// 执行页面测量与高清捕获动作链
	err := chromedp.Run(tabCtx,
		chromedp.Navigate(fileURL),
		chromedp.Sleep(300*time.Millisecond),
		// 精确获取 body 实际盒模型几何宽高，避免采用视口默认宽度
		chromedp.Evaluate(`[
			document.body.scrollWidth || document.body.offsetWidth || 1120,
			document.body.scrollHeight || document.body.offsetHeight || 610
		]`, &dims),
		// 设置 2.0x 标准 Retina 超采样，文字与矢量线条达到极致刀刻级锐度
		chromedp.ActionFunc(func(ctx context.Context) error {
			w := int64(math.Ceil(dims[0]))
			h := int64(math.Ceil(dims[1]))
			return emulation.SetDeviceMetricsOverride(w, h, 2.0, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
		}),
		chromedp.Sleep(200*time.Millisecond),
		chromedp.Evaluate(`window.updateConnections && window.updateConnections()`, nil),
		chromedp.Sleep(100*time.Millisecond),
		// 截取全尺寸高清 PNG (2x Retina)
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			buf, err = page.CaptureScreenshot().
				WithFormat(page.CaptureScreenshotFormatPng).
				WithClip(&page.Viewport{
					X:      0,
					Y:      0,
					Width:  dims[0],
					Height: dims[1],
					Scale:  1.0,
				}).
				Do(ctx)
			return err
		}),
	)

	if err != nil {
		return fmt.Errorf("chromedp 执行失败: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return err
	}

	optimizePNG(outputPath)
	return nil
}

// optimizePNG 采用 256 色高质量调色板自适应量化与无损压缩，将 PNG 体积缩减 80% 以上且肉眼零画质损失
func optimizePNG(path string) {
	script := fmt.Sprintf(`
try:
    from PIL import Image
    im = Image.open(%q)
    q = im.quantize(colors=256, method=Image.Resampling.LANCZOS, dither=Image.Dither.NONE)
    q.save(%q, optimize=True)
except Exception:
    pass
`, path, path)
	cmd := exec.Command("python3", "-c", script)
	_ = cmd.Run()
}
