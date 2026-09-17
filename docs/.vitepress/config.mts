import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
  title: "ECMDB",
  description: "企业级 CMDB、智能工单、自动化告警与分布式任务一体化平台",
  markdown: {
    config(md) {
      md.renderer.rules.table_open = () => '<div class="vp-table-wrap"><table tabindex="0">\n'
      md.renderer.rules.table_close = () => '</table></div>\n'
    }
  },
  themeConfig: {
    logo: '/logo-icon.png',

    search: {
      provider: 'local'
    },

    outline: {
      level: [2, 3],
      label: '本页导航'
    },

    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: '首页', link: '/' },
      {
        text: '系统指南',
        items: [
          { text: '平台概览与架构', link: '/guide/introduction' },
          { text: '快速开始', link: '/guide/quick-start' }
        ]
      },
      { text: '身份治理', link: '/eiam/concept' },
      { text: '资产管理', link: '/cmdb/concept' },
      { text: '工单中心', link: '/workflow/concept' },
      { text: '任务中心', link: '/task/architecture' },
      { text: '演示环境', link: 'http://www.fleetops.top/' }
    ],

    sidebar: [
      {
        text: '系统介绍',
        items: [
          { text: '平台概览与架构', link: '/guide/introduction' },
          { text: '快速开始', link: '/guide/quick-start' },
        ]
      },
      {
        text: '身份治理',
        collapsed: false,
        items: [
          { text: '设计理念与架构', link: '/eiam/concept' },
          { text: '鉴权与数据范围', link: '/eiam/pbac' },
          { text: '多租户与空间治理', link: '/eiam/tenancy' },
          { text: '组织架构与人员治理', link: '/eiam/organization' },
          {
            text: '身份认证与单点登录',
            collapsed: false,
            items: [
              { text: '身份认证与多源凭据', link: '/eiam/credential' },
              { text: '统一 IDP 与单点登录', link: '/eiam/idp' }
            ]
          },
          { text: '契约治理工具链', link: '/eiam/tooling' }
        ]
      },
      {
        text: '资产管理',
        collapsed: false,
        items: [
          { text: '设计理念', link: '/cmdb/concept' },
          { text: '模型定义', link: '/cmdb/model' },
          { text: '资产与全文检索', link: '/cmdb/asset' },
          { text: '拓扑与关联图谱', link: '/cmdb/relation' },
          {
            text: '微服务插件',
            collapsed: false,
            items: [
              { text: '架构与运行时', link: '/cmdb/plugin' },
              { text: '契约与模型注入', link: '/cmdb/plugin-dev' }
            ]
          }
        ]
      },
      {
        text: '工单模块',
        collapsed: false,
        items: [
          { text: '设计思想', link: '/workflow/concept' },
          {
            text: '工单中心',
            collapsed: true,
            items: [
              { text: '提交工单', link: '/workflow/ticket-start' },
              { text: '工单列表', link: '/workflow/ticket' },
            ]
          },
          { text: '模版管理', link: '/workflow/management/template' },
          {
            text: '流程管理',
            collapsed: true,
            items: [
              { text: '编排说明', link: '/workflow/management/workflow' },
              {
                text: '节点库',
                collapsed: true,
                items: [
                  { text: '节点总览', link: '/workflow/node/overview' },
                  { text: '开始节点', link: '/workflow/node/start' },
                  { text: '用户节点', link: '/workflow/node/user' },
                  { text: '自动化节点', link: '/workflow/node/automation' },
                  { text: '网关节点', link: '/workflow/node/gateway' },
                  { text: '群通知节点', link: '/workflow/node/chat' },
                  { text: '结束节点', link: '/workflow/node/end' }
                ]
              }
            ]
          },
          {
            text: '工单示例',
            collapsed: true,
            items: [
              { text: '版本发布', link: '/workflow/cases/deploy' },
              { text: '飞书集成', link: '/workflow/cases/feishu' }
            ]
          }
        ]
      },
      {
        text: '任务中心',
        collapsed: false,
        items: [
          { text: '架构与调度拓扑', link: '/task/architecture' },
          { text: '代码制品与模版', link: '/task/template' },
          { text: '环境变量与凭据', link: '/task/variable' },
          { text: '执行单元与调度路由', link: '/task/execution' },
          { text: '异构执行引擎与 SDK', link: '/task/engine' },
          { text: 'Ansible 凭据与互信', link: '/task/ansible-credential' },
          { text: '运行模式与配置实战', link: '/task/startup' }
        ]
      }
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/Duke1616/ecmdb' }
    ],

    footer: {
      message: '基于 MIT 许可发布',
      copyright: 'Copyright © 2024-present Duke1616'
    }
  },
  mermaid: {
    // mermaidConfig options here
  },
  vite: {
    ssr: {
      noExternal: ['mermaid', 'dayjs', 'vitepress-plugin-mermaid']
    },
    optimizeDeps: {
      include: ['mermaid', 'dayjs']
    }
  }
}))
