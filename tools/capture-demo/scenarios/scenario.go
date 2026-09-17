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
	// 支持原生 input 和 el-input 组件
	Fill ActionType = "fill"

	// FillBySelector 按 CSS 选择器精确定位输入框并填入值
	// 当 Fill 的模糊匹配找不到时使用（如多个同名字段的复杂表单）
	FillBySelector ActionType = "fill_selector"

	// SelectOption 在 El-Select 下拉组件中选择指定选项（按选项文字匹配）
	// Field: 下拉框的 placeholder 或 label 文字（用于定位哪个下拉框）
	// Value: 要选择的选项文字
	SelectOption ActionType = "select_option"

	// Submit 点击提交/确认按钮，并等待接口响应完成
	// Text: 按钮文字（如 "确认"、"提交"、"保存"）
	// WaitSelector: 提交成功后等待出现的元素选择器（可选，如 ".el-table__row"）
	Submit ActionType = "submit"

	// WaitToast 等待成功提示消息出现后自动消失（用于确认操作已完成）
	// 不填 Text 则等待任意 .el-message--success 出现即可
	WaitToast ActionType = "wait_toast"

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

	// Text 点击目标文本，用于 ClickText / Submit
	Text string

	// Selector CSS 选择器，用于 ClickSelector / WaitVisible / Scroll / FillBySelector
	// ClickSelector：选择器不存在时会静默跳过
	Selector string

	// Field 表单字段标识（placeholder / name / type / label 文字），用于 Fill / SelectOption
	Field string

	// Value 要填入或选择的值，用于 Fill / FillBySelector / SelectOption
	Value string

	// WaitSelector 操作完成后额外等待的 CSS 选择器（可选），用于 Submit
	// 例如提交后等待列表行出现：".el-table__row"
	WaitSelector string

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
