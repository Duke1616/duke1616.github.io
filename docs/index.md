---
# https://vitepress.dev/reference/default-theme-home-page
layout: home

hero:
  name: "ECMDB"
  text: "一体化运维管理平台"
  tagline: "覆盖资产拓扑、工单流程、自动化作业、智能告警与权限治理的一体化运维协同中枢"
  image:
    src: /logo-icon.png
    alt: ECMDB Logo
  actions:
    - theme: brand
      text: 快速开始
      link: /guide/quick-start
    - theme: alt
      text: 在线环境（demo/123456）
      link: http://www.fleetops.top/

features:
  - title: 流程工单 · EFlow
    details: 低代码可视化拖拽表单设计，多级条件分支与审批流转矩阵，支持会签或签、飞书协同与全过程合规审计。
    link: /workflow/concept
  - title: 配置资产 · ECMDB
    details: Schema-less 动态模型引擎与多维关系拓扑，集成 ecmdb-plugins 扩展生态，支持零泄露挂载 WebSSH 与 SFTP 工作台。
    link: /cmdb/concept
  - title: 身份治理 · EIAM
    details: 多租户强隔离与组织架构基座，基于 Casbin + OPA 双引擎执行 PBAC 策略裁决与 AccessScope 行级数据边界约束。
    link: /eiam/concept
  - title: 自动化作业 · ETask
    details: 分布式运维作业引擎，支持跨节点高并发批量下发、Shell/Python 剧本调用与 SSE 流式终端日志实时回显。
    link: /task/template
  - title: 告警通知 · EAlert
    details: 告警治理与统一消息通知引擎，支持多源事件接入与抑制降噪、飞书/邮件多通道消息触达与 On-Call 可视化轮值排班。
    link: /guide/introduction
---
