package scenarios

func init() {
	Register(PluginScenario)
}

// PluginScenario 插件体系全流程截图场景
// 对应文档：docs/cmdb/plugin.md
var PluginScenario = Scenario{
	Name:        "plugin",
	Description: "SSH 插件动作能力契约与 Web Shell / SFTP 运维工作区全流程",
	Steps: []Step{
		// ── 01 插件中心：模型链路视图 ─────────────────────────────────────────
		{
			Title:  "插件中心（模型链路视图）",
			Output: "images/plugin/01-plugin-center.png",
			Actions: []Action{
				Nav("/cmdb/model/plugin-define"),
				Wait(".el-card, .plugin-item, [class*='plugin'], .el-table__row"),
				Sleep(800),
			},
		},

		// ── 02 插件动作能力：SSH 动作契约拓扑 ─────────────────────────────────
		{
			Title:  "插件动作能力（SSH 动作契约）",
			Output: "images/plugin/02-plugin-actions.png",
			Actions: []Action{
				Click("动作能力"),
				Sleep(600),
				Wait(".el-table__row, .action-item, [class*='action']"),
				Sleep(400),
			},
		},

		// ── 03 资产列表：操作栏插件动作注入 ───────────────────────────────────
		{
			Title:  "资产列表（插件动作注入）",
			Output: "images/plugin/03-resource-actions.png",
			Actions: []Action{
				Nav("/cmdb/resource/list?uid=host&name=%E4%B8%BB%E6%9C%BA"),
				Wait(".el-table__row, .el-table .el-table__body tr"),
				Sleep(500),
			},
		},

		// ── 04 独立运维视窗：连接方式选择弹窗 ─────────────────────────────────
		{
			Title:  "独立运维视窗（连接方式选择）",
			Output: "images/plugin/04-runtime-select.png",
			Actions: []Action{
				ClickInPlace("Web Shell"),
				Sleep(1500),
			},
		},

		// ── 05 极客命令行终端：Web Shell 会话 ─────────────────────────────────
		{
			Title:  "极客命令行终端（Web Shell 会话）",
			Output: "images/plugin/05-terminal-session.png",
			Actions: []Action{
				Click("连接"),
				Sleep(3000),
			},
		},

		// ── 06 SFTP 远程工作台：Web Sftp 文件管理器 ───────────────────────────
		{
			Title:  "SFTP 远程工作台（Web Sftp 文件管理器）",
			Output: "images/plugin/06-sftp-workspace.png",
			Actions: []Action{
				Nav("/cmdb/resource/list?uid=host&name=%E4%B8%BB%E6%9C%BA"),
				Wait(".el-table__row, .el-table .el-table__body tr"),
				Sleep(500),
				ClickInPlace("Web Sftp"),
				Sleep(1500),
				Click("连接"),
				Sleep(3000),
			},
		},
	},
}
