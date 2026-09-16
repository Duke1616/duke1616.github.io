# 生产部署与高可用指南

ECMDB 作为一个涵盖资产中枢、工单流程、监控告警与分布式任务调度的生产级平台，建议在生产环境中根据各子系统和中间件的特征进行合理规划与高可用部署。


## 1. 架构拓扑与硬件规格建议

```mermaid
graph TD
    LB[生产负载均衡器 Nginx / SLB] --> Web[ECMDB-Web (Vue 3 静态资源 / 容器)]
    LB --> Core[ECMDB 核心服务集群 (Go 实例 x2)]
    Core --> Alert[EAlert 告警引擎]
    Core --> Kafka[Kafka 消息中间件]
    Kafka --> ETask[ETask 分布式 Worker 集群]

    Core --> MySQL[(MySQL 8.0+ 关系数据)]
    Core --> Mongo[(Percona MongoDB 7.0+ 资产元数据)]
    Core --> Redis[(Redis 7.x 缓存与分布式锁)]
    ETask --> Etcd[(Etcd 3.5+ 服务注册与选主)]
```

### 推荐生产配置标准

| 节点组件 | 最低配置 (测试/小规模) | 推荐配置 (生产基线: 1000+ 节点) | 存储与高可用建议 |
| :--- | :--- | :--- | :--- |
| **ECMDB 核心服务** | 2C 4G | 4C 8G (双节点主备/负载) | 无状态应用，支持水平扩容 |
| **MySQL** | 2C 4G | 4C 16G (SSD) | 一主一从，启用 binlog 与自动备份 |
| **Percona MongoDB** | 4C 8G | 8C 16G (NVMe SSD) | **必须使用支持 ngram 插件版本**，建议 Replica Set 副本集 |
| **Redis** | 2C 4G | 2C 8G | Redis Sentinel 哨兵模式或高可用云 Redis |
| **Kafka & Zookeeper** | 2C 4G | 4C 8G (3 节点集群) | 保障日志削峰与异步任务吞吐 |
| **Etcd** | 2C 2G | 2C 4G (3 节点 Raft 集群) | 负责 ETask Worker 注册与租约维护 |


## 2. 生产 Docker Compose 编排模板

在物理服务器或中小型生产场景下，可通过 Docker Compose 实现标准部署：

```yaml
version: '3.8'

networks:
  ecmdb-net:
    driver: bridge

volumes:
  mysql_data:
  mongo_data:
  redis_data:

services:
  # 1. 关系型数据库 MySQL
  mysql:
    image: mysql:8.0.36
    container_name: ecmdb-mysql
    restart: always
    environment:
      MYSQL_ROOT_PASSWORD: YourSecureRootPassword
      MYSQL_DATABASE: ecmdb
      MYSQL_USER: ecmdb
      MYSQL_PASSWORD: YourSecureDbPassword
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci --default-authentication-plugin=mysql_native_password
    volumes:
      - mysql_data:/var/lib/mysql
    networks:
      - ecmdb-net

  # 2. 资产元数据 Percona MongoDB (带 ngram 索引支持)
  mongodb:
    image: percona/percona-server-mongodb:7.0
    container_name: ecmdb-mongodb
    restart: always
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: YourSecureMongoPassword
    volumes:
      - mongo_data:/data/db
    networks:
      - ecmdb-net

  # 3. 缓存与锁 Redis
  redis:
    image: redis:7.2-alpine
    container_name: ecmdb-redis
    restart: always
    command: redis-server --requirepass YourSecureRedisPassword
    volumes:
      - redis_data:/data
    networks:
      - ecmdb-net

  # 4. ECMDB 核心后台 API
  ecmdb-backend:
    image: duke1616/ecmdb:latest
    container_name: ecmdb-backend
    restart: always
    depends_on:
      - mysql
      - mongodb
      - redis
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - MYSQL_DSN=ecmdb:YourSecureDbPassword@tcp(mysql:3306)/ecmdb?charset=utf8mb4&parseTime=True&loc=Local
      - MONGO_URI=mongodb://admin:YourSecureMongoPassword@mongodb:27017/?authSource=admin
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=YourSecureRedisPassword
    networks:
      - ecmdb-net

  # 5. 前端 Web Nginx
  ecmdb-web:
    image: duke1616/ecmdb-web:latest
    container_name: ecmdb-web
    restart: always
    ports:
      - "80:80"
    depends_on:
      - ecmdb-backend
    networks:
      - ecmdb-net
```


## 3. 核心环境变量字典

生产运行中可以通过配置文件 `config.yaml` 或直接注入环境变量对服务进行精细化调整：

| 变量名 | 默认值 / 示例 | 说明 |
| :--- | :--- | :--- |
| `SERVER_PORT` | `8080` | 后端服务监听端口 |
| `JWT_SECRET` | 随机强字符串 (>= 32位) | 用于 API 鉴权 Token 的对称签名密钥，生产务必替换 |
| `AES_SECRET_KEY` | 32位密文秘钥 | 用于 CMDB 敏感密码字段加密的 AES-256 密钥，**切勿丢失** |
| `LOG_LEVEL` | `info` / `warn` / `error` | 日志输出级别，生产建议 `info` |
| `FEISHU_APP_ID` | `cli_xxx` | 飞书开放平台应用 ID（用于工单推送） |
| `FEISHU_APP_SECRET` | `sec_xxx` | 飞书应用凭证 Secret |


## 4. 数据库初始化与备份策略

### 首次初始化
在容器或二进制首次启动后，运行内置初始化指令以创建系统内置角色、菜单及 Casbin 基础策略：
```bash
# 进入容器执行数据初始化
docker exec -it ecmdb-backend ./ecmdb init
```

### 数据定期冷备脚本建议
建议通过 Linux Cron 定期将 MySQL 与 MongoDB 备份归档至对象存储或独立备份机：

```bash
#!/usr/bin/env bash
# 定时备份脚本: backup_ecmdb.sh
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_DIR="/data/backup/ecmdb_${DATE}"
mkdir -p "${BACKUP_DIR}"

# 1. 备份 MySQL 数据
docker exec ecmdb-mysql mysqldump -uecmdb -pYourSecureDbPassword ecmdb | gzip > "${BACKUP_DIR}/mysql_ecmdb.sql.gz"

# 2. 备份 MongoDB 资产库
docker exec ecmdb-mongodb mongodump -uadmin -pYourSecureMongoPassword --authenticationDatabase admin -d ecmdb --gzip --archive="${BACKUP_DIR}/mongo_ecmdb.archive.gz"

# 3. 清理 15 天前的过期备份
find /data/backup -mindepth 1 -maxdepth 1 -type d -mtime +15 -exec rm -rf {} +
```


## 5. 版本平滑升级步骤

当发布新版本镜像或二进制时，请遵循以下升级规范：

1. **提前备份**：升级前对当前 MySQL 与 MongoDB 执行全量快照。
2. **拉取新镜像**：`docker compose pull ecmdb-backend ecmdb-web`。
3. **滚动重启**：`docker compose up -d` 重新创建容器。
4. **运行数据库迁移**：若新版本涉及表结构迭代，系统会在启动时自动执行迁移逻辑，或通过 `./ecmdb migrate` 确认完成。
5. **验证功能**：登录系统，验证 CMDB 检索与工单流转正常。
