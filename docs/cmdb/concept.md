# 设计理念与架构

ECMDB 是平台生态中负责**数字资产中枢**与**动态元模型驱动**的核心微服务。

- **源码仓库**：[GitHub: Duke1616/ecmdb](https://github.com/Duke1616/ecmdb)
- **服务契约**：HTTP RESTful (`:8000`) / 内部 gRPC (`:8078`)
- **核心底座**：Golang 1.25、Percona Server for MongoDB 7.0+、Redis 7.0+

---

## 1. 总体架构

ECMDB 采用分层解耦架构，核心能力由元模型驱动引擎与专用存储底座协同实现：

![ECMDB 核心分层架构全景](/images/cmdb-core-architecture.png)

- **外部接入层**：承接统一 Web 控制台交互、自动化脚本与第三方系统纳管，提供 HTTP RESTful (`:8000`) 与内部 gRPC (`:8078`) 双通道，统一由 **EIAM** 执行身份认证与 PBAC 策略拦截；
- **Core 服务层**：收敛**三级元模型管理**、**资产平铺内联读写**与**双向拓扑图谱**，横切贯穿 `IResourceProtector` 敏感凭据统管；
- **底层引擎层**：基于 Percona MongoDB 统一集合（平铺存储 + 原生 ngram 倒排索引）提供高吞吐，辅以 Redis 分布式锁与热点缓存。

---

## 2. 核心设计特色

### 统一平铺内联存储 (`bson:",inline"`)

异构资产属性统一收敛在单一集合 `c_resources` 中，依靠 `model_uid` 区分模型，通过平铺打散存储在根文档：

```json
{
  "_id": "66e7f8...",
  "resource_id": 10086,
  "tenant_id": "system",
  "model_uid": "host",
  "name": "prod-app-01",
  "ip": "10.0.8.21",
  "password": "ENC:V1:a8f93c...", // Secure 敏感属性自动带版本前缀落盘加密
  "ctime": 1726488000
}
```

- **零 DDL 锁表**：新增自定义字段只需插入元数据，资产数据免 DDL、随写随存；
- **原生原子局部读写**：所有字段直接位于根路径，天然支持 MongoDB 原生 `$set` / `$unset` 局部原子更新；
- **高效复合索引**：支持自由针对高频平铺字段建立复合索引（如 `tenant_id + model_uid + ip`）。

### 原生通配 3-gram 全文检索

免除外部 Elasticsearch 与 Canal/Kafka 同步链路，依托 Percona MongoDB 原生 `ngram` 插件实现全字段模糊搜索：

```go
// internal/repository/dao/init.go
// $** 通配捕获任意动态属性，ngram 滑窗实现中文/IP片段分词
indexes := []mongo.IndexModel{
    {
        Keys:    bson.D{{Key: "$**", Value: "text"}},
        Options: options.Index().SetDefaultLanguage("ngram"),
    },
}
```

- **全字段自动纳入**：任意模型下新增的属性自动被倒排索引捕获，零维护成本；
- **零外部中间件**：单套数据库即可支撑 IP 片段、序列号和名称的毫秒级全文检索。

### 关系拓扑与穿透寻路

资产并非孤立存在，系统通过 `c_relations` 与 `c_relation_types` 构建双向网络：

- **类型约束**：预定义源与目标约束（如 `Host 运行于 Pod`、`Host 关联网关`），约束关系基数；
- **双向图谱解析**：支持向下依赖追踪与向上影响面反查；
- **级联网络穿透**：运维直连或下发脚本时，自动沿拓扑寻路合成跳板机网关上下文。

### 凭据全生命周期保护

对密码、私钥等机密数据执行闭环防护：

- **落盘阶段**：`Secure` 属性统一采用 **AES-GCM-256** 密文存储；
- **展示阶段**：列表与详情查询在接口层自动置空脱敏（展示 `[已脱敏]`），防止前端审查泄露；
- **直连阶段**：微服务插件（WebShell/SFTP）发起握手时通过内网 gRPC 瞬态解密，握手完毕即刻在内存彻底销毁。

---

## 3. 控制面与数据面解耦

ECMDB 将运维操作入口彻底与资产主站解耦：

- **ECMDB Core（控制面）**：专注元模型、资产管理与凭据安全，保障高可用；
- **ecmdb-plugins（数据面）**：独立微服务承接 SSH WebShell 长连接与 SFTP 流传输，崩溃或高负载完全不影响主站。

> [!TIP]
> 详细反向代理流程、插件自动注册契约与 Go Tag 绑定机制，请参阅：  
> **👉 [微服务插件体系](/cmdb/plugin)**

---

## 4. 快速导航

- **[模型管理与字段定义](/cmdb/model)**：模型创建、属性组与字段校验规则
- **[资产数据生命周期](/cmdb/asset)**：资产增删改查、ngram 检索与批量导入导出
- **[关系拓扑与依赖网](/cmdb/relation)**：拓扑类型约束与图谱连通性
- **[微服务插件体系](/cmdb/plugin)**：独立插件微服务架构与零信任握手全景
