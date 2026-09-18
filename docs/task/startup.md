# 运行模式与配置实战

ETask 采用单二进制多模式架构。所有角色整合于 `etask` 中，统一通过参数指定运行模式与配置文件：

```bash
./etask server --mode <mode> --config <path-to-config.yaml>
```

---

## 运行模式选型矩阵

各模式职责、基础设施依赖与网络端口对比：

| 运行角色 | 启动参数 | 核心依赖 | 开放端口 | 适用场景与特性 |
| :--- | :--- | :--- | :--- | :--- |
| **调度中心** | `--mode scheduler` | MySQL, Redis, S3, Etcd | HTTP `8765`, gRPC `9002` | 核心管控区，负责调度抢占、状态机与制品归档 |
| **隔离区代理** | `--mode agent` | Kafka, Etcd, gRPC | **零入站端口暴露** | DMZ / 边缘受限网络，全单向出站连接（拉取任务、注册心跳与下载制品） |
| **直连执行端** | `--mode executor` | Etcd, gRPC | gRPC `9004` (默认直推) | 核心内网环境，支持 PULL 长轮询反压或 PUSH 直推 |
| **单机全功能** | `--mode all` | 单机聚合全组件 | HTTP `8765`, gRPC `9002` | 本地开发调试或轻量边缘一体化部署 |

---

## 快速启动命令

::: code-group

```bash [Scheduler]
# 启动调度中心控制面
./etask server --mode scheduler --config config/scheduler.yaml
```

```bash [Agent]
# 启动隔离区 Agent（单向长连接 Kafka，零端口暴露）
./etask server --mode agent --config config/agent.yaml
```

```bash [Executor]
# 启动直连 Executor（默认 PUSH 直推，可选 PULL 边缘长轮询）
./etask server --mode executor --config config/execute.yaml
```

```bash [All]
# 单进程一体化拉起调度中心与本地执行引擎
./etask server --mode all --config config/all.yaml
```

:::

---

## 配置文件模板

::: code-group

```yaml [scheduler.yaml]
log:
  debug: true

# 数据库存储与行级排他锁 (FOR UPDATE SKIP LOCKED)
mysql:
  dsn: "root:123456@tcp(10.0.1.10:3306)/etask?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s&multiStatements=true&interpolateParams=true"

# 分布式协调与就绪唤醒广播 (Pub/Sub)
redis:
  addr: "10.0.1.11:6379"
  password: "YourRedisPassword"

# 隔离区消息总线 (用于向 Agent 模式下发任务)
kafka:
  network: tcp
  addresses:
    - 10.0.1.20:9092

# 服务注册与发现
etcd:
  endpoints:
    - 10.0.1.12:2379

# 代码制品对象存储
artifact:
  temp_dir: "/tmp/etask/artifacts"
  storage:
    driver: "s3"
    s3:
      endpoint: "http://minio.internal:9000"
      secure: false
      bucket: "etask-artifacts"
      access_key: "minio_access_key"
      secret_key: "minio_secret_key"

# 控制台与 RESTful API
web:
  host: "0.0.0.0"
  port: 8765

# gRPC 通信端口与凭据
grpc:
  server:
    scheduler:
      name: scheduler
      listen_addr: "0.0.0.0:9002"
      advertise_addr: "" # 广播地址，留空自动检测
      auth_token: "1234567890"
  client:
    # 投递告警通知
    ealert:
      name: "ealert"
      auth_token: "1234567890"
    eiam:
      name: "eiam"
      auth_token: "1234567890"

# 核心调度循环参数
scheduler:
  schedule_interval: 10s
  batch_size: 100
  batch_timeout: 3s
  preempted_timeout: 10m
  max_concurrent_tasks: 1000
```

```yaml [agent.yaml]
log:
  debug: true

# Agent 订阅标识与工作协程
agent:
  name: "dmz-agent-01"
  topic: "agent_execute"
  desc: "DMZ 隔离区专用异步脚本执行代理"
  worker_count: 4
  isolation_level: "SHARED" # 资源池隔离级别：SHARED (共享) 或 DEDICATED (独占)

# 本地制品缓存 (与调度端下载的大型包同步)
artifact_cache:
  dir: "/var/lib/etask/artifact-cache"
  max_download_size: "512MB"
  max_unpacked_size: "2GB"
  max_file_count: 10000
  max_cache_size: "10GB"

# 本地多语言执行引擎运行时
runtime:
  workspace_dir: "/tmp/etask/runs"
  workspace_max_age: 24h
  shell:
    enabled: true
    binary: "/bin/bash"
  python:
    enabled: true
    binary: "/usr/bin/python3"
  ansible:
    enabled: true
    binary: "ansible-playbook"
    # 密码凭据依赖该命令；默认值为 sshpass
    # sshpass_binary: "sshpass"
    # credential_root: "/run/credentials/etask-ansible"
    # known_hosts_file: "/etc/etask/ssh/known_hosts"
    # credentials:
    #   production-linux:
    #     type: "private_key"
    #     username: "etask"
    #     private_key_file: "production-linux-key"
    #   legacy-linux:
    #     type: "password"
    #     username: "root"
    #     password_file: "legacy-linux-password"
  sandbox:
    mode: auto # auto (root 自动降权为普通用户), off (关闭降权)
    uid: 65534
    gid: 65534
  max_code_size: "4MB"
  max_args_size: "1MB"
  max_variables_size: "1MB"
  max_log_line_size: "1MB"
  max_result_size: "4MB"
  archive:
    enabled: true
    failed_only: true
    dir: "/var/lib/etask/script-archive"
    max_age: 168h
    max_size: "10GB"

# 任务分发总线 (单向出站长连接，零入站端口暴露)
kafka:
  network: tcp
  addresses:
    - 192.168.10.50:9092

# 节点租约登记与服务发现 (单向出站外联注册到 /etask/kafka)
etcd:
  endpoints:
    - 192.168.10.50:2379

# 调度中心 gRPC 客户端 (单向出站拉取大文件制品与实时状态同步)
grpc:
  client:
    scheduler:
      name: "scheduler"
      auth_token: "1234567890"
```

```yaml [execute.yaml]
log:
  debug: true

# 执行节点集群配置
executor:
  id: "node-01"
  # 任务执行模式：
  # - PUSH: (默认首选) 调度中心主动推给节点，微秒级即刻直达
  # - PULL: 边缘长轮询模式。节点主动向中心拉取任务，适用于受限网络环境
  mode: "PUSH"
  desc: "核心局域网通用算力执行节点"
  isolation_level: "SHARED"

# 服务注册与发现
etcd:
  endpoints:
    - 10.0.1.12:2379

# 本地制品缓存
artifact_cache:
  dir: "/var/lib/etask/artifact-cache"
  max_download_size: "512MB"
  max_unpacked_size: "2GB"
  max_cache_size: "10GB"

# 本地脚本执行环境
runtime:
  workspace_dir: "/tmp/etask/runs"
  workspace_max_age: 24h
  shell:
    enabled: true
    binary: "/bin/bash"
  python:
    enabled: true
    binary: "python3"
  ansible:
    enabled: true
    binary: "ansible-playbook"
    # 密码凭据依赖该命令；默认值为 sshpass
    # sshpass_binary: "sshpass"
    # credential_root: "/run/credentials/etask-ansible"
    # known_hosts_file: "/etc/etask/ssh/known_hosts"
    # credentials:
    #   production-linux:
    #     type: "private_key"
    #     username: "etask"
    #     private_key_file: "production-linux-key"
    #   legacy-linux:
    #     type: "password"
    #     username: "root"
    #     password_file: "legacy-linux-password"
  sandbox:
    mode: auto
    uid: 65534
    gid: 65534

# gRPC 服务监听与客户端连接
grpc:
  server:
    executor:
      name: execute
      listen_addr: "0.0.0.0:9004" # PUSH 模式接收调度中心推送
      advertise_addr: "" # 广播地址，留空自动检测
      # 身份认证 Token。留空且启用 Redis 时自动开启集群 RSA 动态自愈认证；显式填写则使用对称 HMAC 验证
      auth_token: ""
  client:
    scheduler:
      name: "scheduler"
      auth_token: "1234567890"
```

```yaml [all.yaml]
log:
  debug: true

# 制品仓库配置 (支持 local 本地存储或 s3/MinIO 对象存储)
artifact:
  temp_dir: "/tmp/etask/artifacts"
  storage:
    driver: "local" # 可选 local 或 s3
    local:
      root: "./data/artifact-store"
    s3:
      endpoint: "http://127.0.0.1:9000"
      secure: false
      region: "us-east-1"
      bucket: "etask-artifacts"
      prefix: "codebook"
      access_key: "minio_access_key"
      secret_key: "minio_secret_key"
      session_token: ""

# Agent 模式配置 (消费 Kafka 队列的执行代理)
agent:
  name: "local-agent"
  topic: "agent_execute"
  worker_count: 4
  desc: "单机全功能内置 Agent 代理"
  isolation_level: "SHARED" # 资源池隔离级别：SHARED (共享) 或 DEDICATED (独占)

# 本地制品缓存限制 (Agent 与 Executor 共享)
artifact_cache:
  dir: "./data/artifact-cache"
  max_download_size: "512MB"
  max_unpacked_size: "2GB"
  max_file_count: 10000
  max_cache_size: "10GB"

# 本地多语言执行引擎运行时
runtime:
  workspace_dir: "/tmp/etask/runs"
  workspace_max_age: 24h
  shell:
    enabled: true
    binary: "/bin/bash"
  python:
    enabled: true
    binary: "python3"
  ansible:
    enabled: true
    binary: "ansible-playbook"
    # 密码凭据依赖该命令；默认值为 sshpass
    # sshpass_binary: "sshpass"
    # credential_root: "/run/credentials/etask-ansible"
    # known_hosts_file: "/etc/etask/ssh/known_hosts"
    # credentials:
    #   production-linux:
    #     type: "private_key"
    #     username: "etask"
    #     private_key_file: "production-linux-key"
    #   legacy-linux:
    #     type: "password"
    #     username: "root"
    #     password_file: "legacy-linux-password"
  sandbox:
    mode: auto # root 自动降权，非 root 保持现状
    uid: 65534
    gid: 65534
  max_code_size: "4MB"
  max_args_size: "1MB"
  max_variables_size: "1MB"
  max_log_line_size: "1MB"
  max_result_size: "4MB"
  archive:
    enabled: true
    failed_only: true
    dir: "./data/script-archive"
    max_age: 168h
    max_size: "10GB"

# 执行器集群配置
executor:
  id: "local-executor-01"
  # 任务执行模式：
  # - PUSH: (默认首选) 调度中心主动推给节点，微秒级即刻直达
  # - PULL: 边缘长轮询模式。节点主动向中心拉取任务，适用于受限网络环境
  mode: "PUSH"
  desc: "单机内置通用执行器节点"
  isolation_level: "SHARED"

# Codebook AI 助手 (仅调度/控制面使用)
ai:
  provider: "rawchat" # 可选 openai、rawchat、qwen
  endpoint: "https://rawchat.cn/codex"
  model: "gpt-5.6-sol"
  timeout: "180s"
  max_output_tokens: 8192
  max_concurrency: 4
  reasoning_effort: "low"

# 关系型数据库存储
mysql:
  dsn: "root:123456@tcp(127.0.0.1:3306)/etask?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local&timeout=1s&readTimeout=3s&writeTimeout=3s&multiStatements=true&interpolateParams=true"

# 分布式缓存与广播总线
redis:
  addr: "127.0.0.1:6379"
  password: "YourRedisPassword"

# 敏感字段加解密密钥 (AES-256)
encryption:
  version: "V1"
  key: "1234567890CryptoAesKey00"

# 异步消息队列 (隔离区任务下发与状态流转)
kafka:
  network: tcp
  addresses:
    - 127.0.0.1:9092

# 服务注册、发现与分布式锁
etcd:
  endpoints:
    - 127.0.0.1:2379

# 控制台与 RESTful API 服务
web:
  host: "0.0.0.0"
  port: 8765

# 统一鉴权与权限策略中枢对接 (EIAM)
policy:
  auth_url: "http://127.0.0.1:9000"
  discovery_url: "http://127.0.0.1:9000"
  discovery_token: "eiam_sct_your_token_here"

# gRPC 服务端与客户端通信
grpc:
  server:
    scheduler:
      name: scheduler
      listen_addr: "0.0.0.0:9002"
      advertise_addr: "" # 广播地址，留空自动检测
      auth_token: "1234567890"
    executor:
      name: execute
      listen_addr: "0.0.0.0:9004"
      advertise_addr: ""
      # 身份认证 Token。留空且启用 Redis 时自动开启集群 RSA 动态自愈认证；显式填写则使用对称 HMAC 验证
      auth_token: ""
  client:
    ealert:
      name: "ealert"
      auth_token: "1234567890"
    eiam:
      name: "eiam"
      auth_token: "1234567890"
    scheduler:
      name: "scheduler"
      auth_token: "1234567890"

# 调度器抢占与批处理参数
scheduler:
  schedule_interval: 10s
  renew_interval: 5s
  batch_size: 100
  batch_timeout: 3s
  preempted_timeout: 10m
  max_concurrent_tasks: 1000
  token_acquire_timeout: 3s

# 异常任务补偿机制
compensator:
  retry:
    batch_size: 100
    min_duration: 1s
  reschedule:
    batch_size: 100
    min_duration: 1s
  interrupt:
    batch_size: 100
    min_duration: 1s
  termination:
    batch_size: 100
    min_duration: 1s
```

:::

::: tip Ansible 剧本编排与免密互信独立指南
关于 Ansible 驱动的凭据安全隔离、受控端 SSH 免密授权与 Playbook 资产清单引用实战，已剥离为专项文档，详见：[Ansible 剧本编排](/task/runner/ansible)。
:::

---

## 运维与诊断技巧

### 日志级别动态热切
无需重启进程，直接修改配置文件中的 `debug` 字段即刻热生效：

```yaml
log:
  debug: true # 保存后由 Viper (fsnotify) 自动捕获生效
```

### 集群动态自愈认证
调度中心直推任务到 Executor 时采用 **RS256 非对称签名机制**，零静态配置负担：
- **私钥持有端 (Scheduler)**：首次启动时在 Redis 原子生成 2048 位私钥（`etask:auth:scheduler:signing_key:default`），调度直推时签发 JWT；
- **公钥验证端 (Executor)**：通过 Redis 获取公钥（`etask:auth:scheduler:public_key:default`）进行本地验签（本地 5 分钟缓存 + singleflight 防击穿）；
- **向后兼容**：若未配置 Redis 且在 `auth_token` 显式指定密钥，系统自适应退回对称 HMAC 校验模式。
