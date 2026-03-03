---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "ECMDB"
  text: "企业级 CMDB + 工单一体化平台"
  tagline: "现代化企业级配置管理数据库与智能工单系统"
  image:
    src: /logo-icon.png
    alt: ECMDB Logo
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: 在线环境（demo/123456）
      link: http://82.156.165.98:8888/

features:
  - title: 📊 CMDB 资产管理
    details: 自定义模型属性、加密属性、关联关系，基于 Percona MongoDB 的全文搜索，支持复杂关联查询与可视化资产关系图谱。
  - title: 📋 工单流程中心
    details: 可视化拖拽式流程设计器，基于 form-create 的动态表单生成，深度集成飞书审批与丰富的自动化任务处理能力。
  - title: 👥 权限管理系统
    details: 支持 LDAP 多认证方式，基于 Casbin 的灵活策略配置，实现细粒度的菜单、按钮、API 接口权限控制。
  - title: 📅 排班管理系统
    details: 支持日、小时灵活时间单位的排班，基于 RRULE 算法智能调度计算，并支持临时调班与直观的可视化排班表展示。
---

