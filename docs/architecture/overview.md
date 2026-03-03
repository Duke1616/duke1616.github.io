# 🏗️ 系统架构

ECMDB 项目基于微服务和模块化的思想构建，分层清晰，便于扩展和维护。

## 🛠️ 技术栈

### 后端技术栈
- **语言框架**：Go 1.21+ + Gin + GORM
- **数据库**：Percona MongoDB + MySQL + Redis Stack
- **消息队列**：Kafka + Etcd
- **权限控制**：Casbin + JWT
- **流程引擎**：Easy-Workflow
- **依赖注入**：Google Wire

### 前端技术栈
- **框架**：Vue 3 + TypeScript
- **UI 组件**：Element Plus
- **状态管理**：Pinia
- **构建工具**：Vite

### 部署运维
- **容器化**：Docker + Docker Compose
- **CI/CD**：Github Action 自动构建镜像

## 📂 模块化开发大纲

代码组织结构基于领域驱动设计（DDD）的思想，将不同业务功能拆分为独立模块：

```text
ecmdb/
├── internal/
│   ├── model      # CMDB - 模型管理
│   ├── attribute  # CMDB - 字段属性
│   ├── resource   # CMDB - 资产数据
│   ├── relation   # CMDB - 关联关系
│   ├── order      # 工单 - 工单信息
│   ├── workflow   # 工单 - 流程图
│   ├── template   # 工单 - 工单模版
│   ├── task       # 工单 - 自动化任务
│   ├── runner     # 工单 - 执行单元
│   ├── worker     # 工单 - 工作节点
│   ├── codebook   # 工单 - 代码库
│   ├── discovery  # 工单 - 自动发现
│   ├── rota       # 排班 - 排班模块
│   ├── user       # 权限 - 用户模块
│   ├── role       # 权限 - 角色管理
│   ├── department # 权限 - 部门管理
│   ├── menu       # 权限 - 菜单管理
│   ├── permission # 权限 - 权限控制
│   └── terminal   # 终端 - Web终端
```

## 🔗 关联项目

本项目包含以下核心组件：

| 项目 | 描述 | 仓库地址 |
|------|------|----------|
| **前端界面** | Vue3 + TypeScript 现代化前端 | [ecmdb-web](https://github.com/Duke1616/ecmdb-web) |
| **任务执行器** | 分布式任务执行引擎 | [etask](https://github.com/Duke1616/etask) |
| **消息通知** | 统一消息通知服务 | [enotify](https://github.com/Duke1616/enotify) |

> 💡 **提示**：如果没有工单自动化任务需求，可以不部署任务执行器
