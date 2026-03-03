# 🚀 快速开始

## 🌐 在线演示

**立即体验**：无需安装，直接访问在线演示环境

- **演示地址**：[http://82.156.165.98:8888](http://82.156.165.98:8888)
- **登录方式**：标准登录
- **演示账户**：demo
- **演示密码**：123456

> 💡 **提示**：演示环境拥有平台的读取权限，如需可写权限请主动联系，下方有联系方式

## 💻 系统要求
- Go 1.21+
- Docker & Docker Compose
- MongoDB 4.4+
- MySQL 8.0+
- Redis 6.0+

## 🚀 一键部署

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

4. （可选）初始化工单飞书通知模版
```bash
go run main.go init ticket-notify-template
```

## 🔑 默认账户
- **用户名**：admin
- **密码**：123456

## 📚 本地开发指南

1. 安装依赖
```bash
go mod tidy
```

2. 启动服务
```bash
go run main.go start
```

3. 初始化数据
```bash
go run main.go init
```
