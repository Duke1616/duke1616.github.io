package scenarios

// ActionType 定义支持的原子操作类型
type ActionType string

const (
	// Navigate 跳转到指定路径（相对 BaseURL）
	Navigate ActionType = "navigate"

	// ClickText 按可见文本内容点击元素
	ClickText ActionType = "click_text"

	// ClickSelector 按 CSS 选择器点击元素（元素不存在时跳过，不报错）
	ClickSelector ActionType = "click_selector"

	// WaitVisible 等待指定 CSS 选择器的元素出现在视口中
	WaitVisible ActionType = "wait_visible"

	// Fill 向输入框填入内容（按 placeholder / name / type 查找）
	Fill ActionType = "fill"

	// Mask 在截图前用 JS 将页面上的敏感文本替换为演示数据（零侵入脱敏）
	Mask ActionType = "mask"

	// Sleep 等待固定时长（单位 ms）
	Sleep ActionType = "sleep"

	// Scroll 将指定 CSS 选择器的容器滚动到顶部，或让 Selector 元素居中展示
	Scroll ActionType = "scroll"
)

// Action 描述一个原子操作
type Action struct {
	// Type 操作类型，必填
	Type ActionType

	// Path 跳转路径，Type=Navigate 时使用（相对路径如 /cmdb/plugins）
	Path string

	// Text 点击目标文本，Type=ClickText 时使用
	Text string

	// Selector CSS 选择器，Type=ClickSelector / WaitVisible / Scroll 时使用
	// ClickSelector：选择器不存在时会静默跳过
	Selector string

	// Field 表单字段标识（placeholder / name / type），Type=Fill 时使用
	Field string

	// Value 要填入的值，Type=Fill 时使用
	Value string

	// Masks 脱敏替换映射 key=真实内容 value=展示内容，Type=Mask 时使用
	// 会在截图前注入 JS，把页面所有可见文本节点中匹配的部分替换掉
	Masks map[string]string

	// Ms 等待毫秒数，Type=Sleep 时使用
	Ms int

	// InPopup 若为 true，本步骤之后所有动作都在点击后弹出的新 Tab 中执行
	// 只在 Type=ClickText 或 Type=ClickSelector 上生效
	InPopup bool
}

// Step 描述"执行一组动作后截一张图"的完整流程单元
type Step struct {
	// Title 步骤说明（仅用于日志输出，方便调试）
	Title string

	// Output 截图保存路径，相对于 docs/public/ 目录
	// 例如 "images/plugin/01-plugin-center.png"
	Output string

	// Actions 按序执行的原子动作列表
	Actions []Action
}

// Scenario 描述一篇文档页面所有截图的完整场景
type Scenario struct {
	// Name 场景名称，用于日志和进度显示
	Name string

	// Steps 有序的步骤列表，每个 Step 产出一张截图
	Steps []Step
}
