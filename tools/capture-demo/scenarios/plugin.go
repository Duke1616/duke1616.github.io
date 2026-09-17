package scenarios

// PluginScenario 插件体系全流程截图场景
// 对应文档：docs/cmdb/plugin.md 第 5 节"实操演练"
var PluginScenario = Scenario{
	Name: "plugin",
	Steps: []Step{
		{
			Title:  "插件拓扑自省（插件中心 - 模型链路）",
			Output: "images/plugin/01-plugin-center.png",
			Actions: []Action{
				{Type: Navigate, Path: "/cmdb/plugins/builtin.ssh"},
				// 尝试点击拓扑画布的"自适应/居中"工具按钮，让节点居中展示
				// 按钮不存在时静默跳过（不中断流程）
				{Type: ClickSelector, Selector: ".g6-toolbar button[title*='适'], .x6-toolbar-item[data-name*='fit'], [class*='toolbar'] button:last-child"},
				{Type: Sleep, Ms: 600},
			},
		},
		{
			Title:  "动作能力描述（插件动作契约）",
			Output: "images/plugin/02-plugin-actions.png",
			Actions: []Action{
				{Type: ClickText, Text: "动作能力"},
				{Type: Sleep, Ms: 400},
			},
		},
		{
			Title:  "资产列表注入（主机模型操作栏）",
			Output: "images/plugin/03-resource-actions.png",
			Actions: []Action{
				{Type: Navigate, Path: "/cmdb/resource/list?uid=host&name=%E4%B8%BB%E6%9C%BA"},
				// 脱敏：截图前把真实 IP / 端口 / 用户名替换为演示数据
				{Type: Mask, Masks: map[string]string{
					"82.156.165.98":  "10.0.12.88",
					"12222":          "22",
					"linuxserver.io": "root",
					"openssh-server": "prod-app-01",
				}},
			},
		},
		{
			Title:  "独立运维视窗（运行时选择器）",
			Output: "images/plugin/04-runtime-select.png",
			Actions: []Action{
				// NOTE: 截图 03 的页面中有 Web Shell 按钮，点击后会弹出新 Tab
				// InPopup=true 表示本步骤之后的动作在新 Tab 中执行
				{Type: ClickText, Text: "Web Shell", InPopup: true},
				{Type: Sleep, Ms: 800},
			},
		},
		{
			Title:  "极客命令行终端（xterm 连接成功）",
			Output: "images/plugin/05-terminal-session.png",
			Actions: []Action{
				// 在弹出的新 Tab 内，等待 xterm 终端画布渲染完成
				{Type: WaitVisible, Selector: ".xterm"},
				{Type: Sleep, Ms: 500},
				{Type: Mask, Masks: map[string]string{
					"82.156.165.98": "10.0.12.88",
				}},
			},
		},
		{
			Title:  "SFTP 远程工作台（文件管理器）",
			Output: "images/plugin/06-sftp-workspace.png",
			Actions: []Action{
				// 返回选择器页面，改选 SFTP 并截图
				{Type: Navigate, Path: "/cmdb/plugin-runtime?action=select&id=3"},
				{Type: ClickText, Text: "文件管理器"},
				{Type: Sleep, Ms: 800},
				{Type: Mask, Masks: map[string]string{
					"82.156.165.98": "10.0.12.88",
				}},
			},
		},
	},
}
