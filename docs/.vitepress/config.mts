import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// https://vitepress.dev/reference/site-config
export default withMermaid(defineConfig({
  title: "ECMDB",
  description: "企业级 CMDB、智能工单、自动化告警与分布式任务一体化平台",
  themeConfig: {
    logo: '/logo-icon.png',

    search: {
      provider: 'local'
    },

    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/guide/quick-start' },
      { text: '系统概览', link: '/guide/introduction' }
    ],

    sidebar: [
      {
        text: '系统介绍',
        items: [
          { text: '核心概览', link: '/guide/introduction' },
          { text: '快速开始', link: '/guide/quick-start' },
        ]
      },
      {
        text: '工单模块',
        items: [
          { text: '设计思想', link: '/workflow/concept' },
          { text: '模版管理', link: '/workflow/management/template' },
          {
            text: '流程管理',
            collapsed: false,
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
            text: '使用示例',
            collapsed: true,
            items: [
              { text: '应用发布', link: '/workflow/cases/deploy' },
              { text: '权限申请', link: '/workflow/cases/permission' }
            ]
          }
        ]
      },
      {
        text: '任务中心',
        items: [
          { text: '任务模版', link: '/task/template' },
          { text: '执行节点', link: '/task/execution' }
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
