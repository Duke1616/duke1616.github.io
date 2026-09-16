# 快速开始

## 在线演示

**立即体验**：无需安装，直接访问在线演示环境

- **演示地址**：[www.fleetops.top](http://www.fleetops.top/)
- **登录方式**：标准登录
- **演示账户**：demo
- **演示密码**：123456

> [!NOTE]
> 出于安全性考虑，在线演示环境拥有全平台的只读查看权限。


## 快速部署

1. 创建网络
```bash
docker network create ecmdb
```

2. 启动服务
```bash
docker compose -p ecmdb -f deploy/docker-compose.yaml up -d
```

3. 初始化系统
```bash
docker exec -it ecmdb ./ecmdb init
```



## 本地开发指南

### 系统要求

为了保障系统的完整功能运行，建议准备以下环境：

- **基础环境**：Go 1.25+ | Node.js 20+
- **容器编排**：Docker & Docker Compose
- **核心数据**：Percona MongoDB 7.0+ (需支持 ngram) | MySQL 8.0+
- **中间件**：Redis Stack 7.2+ | Kafka | Etcd (用于分布式调度)


### 启动服务

1. 安装依赖
```bash
go mod download
```

2. 启动服务
```bash
go run main.go start
```

3. 初始化数据
```bash
go run main.go init
```