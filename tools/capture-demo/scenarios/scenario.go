package scenarios

// ActionType 定义支持的原子操作类型
type ActionType string

const (
	ActionNavigate       ActionType = "navigate"
	ActionClickText      ActionType = "click_text"
	ActionClickSelector  ActionType = "click_selector"
	ActionClickInPlace   ActionType = "click_inplace"
	ActionWaitVisible    ActionType = "wait_visible"
	ActionFillInput      ActionType = "fill"
	ActionFillBySelector ActionType = "fill_selector"
	ActionSelectOpt      ActionType = "select_option"
	ActionSubmitBtn      ActionType = "submit"
	ActionWaitToast      ActionType = "wait_toast"
	ActionMaskText       ActionType = "mask"
	ActionSleepWait      ActionType = "sleep"
	ActionScrollTo       ActionType = "scroll"
	ActionHoverText      ActionType = "hover"
)

// Action 描述一个原子操作
type Action struct {
	Type         ActionType
	Path         string
	Text         string
	Selector     string
	Field        string
	Value        string
	WaitSelector string
	Masks        map[string]string
	Ms           int
	InPopup      bool
}

// Step 描述"执行一组动作后截一张图"的流程单元
type Step struct {
	Title   string   // 步骤标题（用于 CLI 进度显示）
	Output  string   // 截图保存路径（相对 docs/public，如 images/plugin/01.png）
	Actions []Action // 依序执行的原子动作列表
}

// Scenario 描述一篇文档页面所有截图的完整场景
type Scenario struct {
	Name        string            // 场景标识符（CLI 参数通过此名称匹配）
	Description string            // 场景业务描述（用于 CLI 帮助与列表展示）
	Masks       map[string]string // 场景专属数据脱敏映射（仅在该场景生命周期内生效，隔离互不污染）
	Steps       []Step            // 有序步骤
}

// Action 构造辅助函数

// Nav 页面跳转
func Nav(path string) Action {
	return Action{Type: ActionNavigate, Path: path}
}

// Click 点击指定文本内容的按钮或链接
func Click(text string) Action {
	return Action{Type: ActionClickText, Text: text}
}

// ClickPopup 在点击后弹出的新标签页中接管后续操作
func ClickPopup(text string) Action {
	return Action{Type: ActionClickText, Text: text, InPopup: true}
}

// ClickCSS 按 CSS 选择器点击元素
func ClickCSS(selector string) Action {
	return Action{Type: ActionClickSelector, Selector: selector}
}

// ClickInPlace 点击按钮同时拦截 window.open 在当前页面导航
func ClickInPlace(text string) Action {
	return Action{Type: ActionClickInPlace, Text: text}
}

// Wait 等待 CSS 选择器对应的元素出现在视口
func Wait(selector string) Action {
	return Action{Type: ActionWaitVisible, Selector: selector}
}

// Fill 按输入框 label/placeholder 填入内容（自动触发响应式同步）
func Fill(field, value string) Action {
	return Action{Type: ActionFillInput, Field: field, Value: value}
}

// FillCSS 按 CSS 选择器精确定位填入内容
func FillCSS(selector, value string) Action {
	return Action{Type: ActionFillBySelector, Selector: selector, Value: value}
}

// Select 在下拉菜单中选中指定文本选项
func Select(field, value string) Action {
	return Action{Type: ActionSelectOpt, Field: field, Value: value}
}

// Submit 提交表单并等待响应完成
func Submit(btnText string, waitSelector ...string) Action {
	var ws string
	if len(waitSelector) > 0 {
		ws = waitSelector[0]
	}
	return Action{Type: ActionSubmitBtn, Text: btnText, WaitSelector: ws}
}

// Hover 鼠标悬停在文本元素上
func Hover(text string) Action {
	return Action{Type: ActionHoverText, Text: text}
}

// Sleep 等待指定毫秒
func Sleep(ms int) Action {
	return Action{Type: ActionSleepWait, Ms: ms}
}

// Scroll 滚动到选择器对应的元素
func Scroll(selector string) Action {
	return Action{Type: ActionScrollTo, Selector: selector}
}

// Mask 局部补充脱敏
func Mask(masks map[string]string) Action {
	return Action{Type: ActionMaskText, Masks: masks}
}
