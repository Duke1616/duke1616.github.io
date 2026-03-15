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
        text: '工单系统',
        items: [
          { text: '概念总览', link: '/workflow/concept' },
          {
            text: '节点库',
            collapsed: true,
            items: [
              { text: '节点总览', link: '/workflow/node/overview' },
              { text: '开始节点', link: '/workflow/node/start' },
              { text: '审批节点', link: '/workflow/node/approval' },
              { text: '执行节点', link: '/workflow/node/task' },
              { text: '条件节点', link: '/workflow/node/condition' },
              { text: '抄送节点', link: '/workflow/node/notify' },
              { text: '结束节点', link: '/workflow/node/end' }
            ]
          },
          { text: '使用案例', link: '/workflow/cases' }
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
