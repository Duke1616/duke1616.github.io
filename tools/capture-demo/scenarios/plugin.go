package scenarios

func init() {
	Register(PluginScenario)
}

// PluginScenario 插件体系全流程截图场景
// 对应文档：docs/cmdb/plugin.md
var PluginScenario = Scenario{
	Name:        "plugin",
	Description: "SSH 插件配置与 Web Shell / SFTP 运维工作区",
	// 场景专属脱敏规则（自动作用于本场景内全部步骤截屏，隔离互不污染）
	Masks: map[string]string{
		"82.156.165.98":  "10.0.12.88",
		"12222":          "22",
		"linuxserver.io": "root",
		"openssh-server": "prod-app-01",
	},
	Steps: []Step{
		// 01 插件中心：模型链路视图
		{
			Title:  "插件定义列表",
			Output: "images/plugin/01-plugin-center.png",
			Actions: []Action{
				Nav("/cmdb/model/plugin-define"),
				Wait(".el-card, .plugin-item, [class*='plugin'], .el-table__row"),
				Sleep(800),
			},
		},

		// 02 插件动作能力：SSH 动作契约
		{
			Title:  "动作能力配置",
			Output: "images/plugin/02-plugin-actions.png",
			Actions: []Action{
				Click("动作能力"),
				Sleep(600),
				Wait(".el-table__row, .action-item, [class*='action']"),
				Sleep(400),
			},
		},

		// 03 资产列表：操作栏插件动作注入
		{
			Title:  "主机资产列表",
			Output: "images/plugin/03-resource-actions.png",
			Actions: []Action{
				Nav("/cmdb/resource/list?uid=host&name=%E4%B8%BB%E6%9C%BA"),
				Wait(".el-table__row, .el-table .el-table__body tr"),
				Sleep(500),
			},
		},

		// 04 连接方式选择弹窗
		{
			Title:  "连接方式选择",
			Output: "images/plugin/04-runtime-select.png",
			Actions: []Action{
				ClickInPlace("Web Shell"),
				Sleep(1500),
			},
		},

		// 05 Web Shell 终端会话
		{
			Title:  "Web Shell 终端会话",
			Output: "images/plugin/05-terminal-session.png",
			Actions: []Action{
				Click("连接"),
				Sleep(3000),
			},
		},

		// 06 Web Sftp 文件管理器
		{
			Title:  "Web Sftp 文件管理器",
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
