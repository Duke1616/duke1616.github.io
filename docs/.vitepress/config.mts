import { defineConfig } from 'vitepress'

// https://vitepress.dev/reference/site-config
export default defineConfig({
  base: '/edocs/',
  title: "ECMDB",
  description: "企业级 CMDB + 工单一体化平台",
  themeConfig: {
    logo: '/logo-icon.png',

    search: {
      provider: 'local'
    },

    // https://vitepress.dev/reference/default-theme-config
    nav: [
      { text: '首页', link: '/' },
      { text: '快速开始', link: '/guide/quick-start' }
    ],

    sidebar: [
      {
        text: '指南',
        items: [
          { text: '快速开始', link: '/guide/quick-start' },
          { text: '核心功能', link: '/guide/features' },
        ]
      },
      {
        text: '系统架构',
        items: [
          { text: '架构概述', link: '/architecture/overview' }
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
  }
})
