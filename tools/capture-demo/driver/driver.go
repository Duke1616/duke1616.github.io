package driver

import (
	"context"
	"fmt"
	"os"
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
	DPR          float64       // 设备像素比（默认 2.0 视网膜高清）
	Timeout      time.Duration // 全局超时（默认 120s）
	ChromePath   string        // 留空则自动检测本地 Chrome
}

// Driver 自动化文档截屏驱动器（持有一个 Chrome Tab 上下文）
type Driver struct {
	Ctx         context.Context
	Cancel      context.CancelFunc
	AllocCtx    context.Context
	AllocCancel context.CancelFunc
	Opts        Options
	newTargetCh chan target.ID // 监听新 Tab 弹出
}

// New 创建并初始化驱动器，启动 Headless Chrome
func New(opts Options) (*Driver, error) {
	opts = applyDefaults(opts)

	execOpts := buildExecOptions(opts)
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

	// 监听新窗口/Tab 弹出事件（用于 WaitForPopup）
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if e, ok := ev.(*target.EventTargetCreated); ok {
			if e.TargetInfo.Type == "page" && e.TargetInfo.OpenerID != "" {
				d.newTargetCh <- e.TargetInfo.TargetID
			}
		}
	})

	if err := chromedp.Run(ctx, emulation.SetDeviceMetricsOverride(
		opts.WindowWidth, opts.WindowHeight, opts.DPR, false,
	)); err != nil {
		d.Close()
		return nil, fmt.Errorf("初始化设备分辨率失败: %w", err)
	}

	return d, nil
}

// Close 释放浏览器进程及所有相关资源
func (d *Driver) Close() {
	if d.Cancel != nil {
		d.Cancel()
	}
	if d.AllocCancel != nil {
		d.AllocCancel()
	}
}

// WaitForPopup 触发操作并等待新 Tab 弹出，返回接管新 Tab 的独立 Driver
func (d *Driver) WaitForPopup(trigger func() error) (*Driver, error) {
	if err := trigger(); err != nil {
		return nil, fmt.Errorf("触发弹窗动作失败: %w", err)
	}

	select {
	case targetID := <-d.newTargetCh:
		popupCtx, cancel := chromedp.NewContext(d.AllocCtx, chromedp.WithTargetID(targetID))
		popup := &Driver{
			Ctx:         popupCtx,
			Cancel:      cancel,
			AllocCtx:    d.AllocCtx,
			AllocCancel: func() {}, // 共享 AllocContext，不单独 Cancel
			Opts:        d.Opts,
			newTargetCh: make(chan target.ID, 5),
		}
		_ = chromedp.Run(popupCtx,
			emulation.SetDeviceMetricsOverride(d.Opts.WindowWidth, d.Opts.WindowHeight, d.Opts.DPR, false),
		)
		_ = popup.InjectCleanStyles()
		_ = popup.WaitStable(500 * time.Millisecond)
		return popup, nil

	case <-time.After(15 * time.Second):
		return nil, fmt.Errorf("等待新窗口/标签页弹出超时")
	}
}

// applyDefaults 填充 Options 中未设置的默认值
func applyDefaults(opts Options) Options {
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
		const macChrome = "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
		if _, err := os.Stat(macChrome); err == nil {
			opts.ChromePath = macChrome
		}
	}
	return opts
}

// buildExecOptions 构建 chromedp ExecAllocator 选项列表
func buildExecOptions(opts Options) []chromedp.ExecAllocatorOption {
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
	return execOpts
}
