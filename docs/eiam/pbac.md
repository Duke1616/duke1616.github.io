# 鉴权与数据范围

EIAM 采用类 AWS IAM 风格的细粒度策略定义规范，将**接口功能准入**（能不能访问）与**行级数据范围**（能看哪几行）彻底解耦。

---

## 1. 策略模型五大要素

标准策略由五个核心要素构成：

| 策略要素 | 判定位置 | 核心作用与职责 |
| :--- | :--- | :--- |
| **Effect** | 鉴权服务端 | 裁决效果：`Allow`（显式允许）或 `Deny`（显式拒绝）。**Deny 拥有最高优先级，一票否决** |
| **Action** | 鉴权服务端 | 操作能力标识，支持通配符（如 `ticket:list`、`ticket:manager:*`） |
| **Resource** | 鉴权服务端 | 目标资源的统一资源标识（URN），支持通配符（如 `urn:ticket:*`） |
| **Condition** | 鉴权服务端（即时求值） | **接口级准入约束**。仅判定系统可信属性（IP、时间、MFA 状态），不满足直接返回 403 阻断 |
| **AccessScope** | 业务持久层（延迟编译） | **行级数据可见范围**。描述业务字段的过滤规则，由微服务持久层自动编译为参数化 SQL |

---

## 2. Condition 与 AccessScope 的本质区别

两者虽采用相同的谓词语法结构，但在执行时机与管控边界上有严格划分：

| 维度 | 接口准入条件 Condition | 行级数据范围 AccessScope |
| :--- | :--- | :--- |
| **管控目标** | 管“能不能调用该接口” | 管“能查看哪些业务数据行” |
| **执行位置** | EIAM 鉴权中心内部立即求值 | 业务微服务持久层（GORMx）延迟编译 |
| **判定属性** | 仅限系统可信上下文（用户名、时间、客户端 IP） | 业务实体字段（创建人、协同处理人、归属部门等） |
| **未命中结果** | 鉴权失败，直接返回 `403 Forbidden` | 允许调用接口，但 SQL 追加条件仅返回匹配行 |

> [!IMPORTANT]
> 业务字段严禁写入 Condition，系统属性禁止写入 AccessScope。职责解耦确保鉴权中心始终保持无状态，不与业务数据库表结构耦合。

---

## 3. 典型策略定义

### 示例 A：管理员全局读写（无 AccessScope 约束）
不声明 AccessScope 时，表示拥有全局数据可见度，微服务将执行无过滤查询：

```json
{
  "Effect": "Allow",
  "Action": ["ticket:manager:history"],
  "Resource": ["*"]
}
```

### 示例 B：员工仅查看自己创建或参与的工单
通过 `ref` 动态引用当前登录用户的用户名，EIAM 在鉴权时将占位符动态替换为真实值：

```json
{
  "Effect": "Allow",
  "Action": ["ticket:manager:history"],
  "Resource": ["*"],
  "AccessScope": {
    "any": [
      {
        "predicate": {
          "key": "ticket:create_by",
          "operator": "StringEquals",
          "values": [{ "type": "ref", "value": "principal:username" }]
        }
      },
      {
        "predicate": {
          "key": "ticket:related_users",
          "operator": "ForAnyValue:StringEquals",
          "values": [{ "type": "ref", "value": "principal:username" }]
        }
      }
    ]
  }
}
```

---

## 4. 谓词操作符速查

EIAM 内部的求值引擎（`pkg/pbac/condition.go`）原生支持以下比较操作符：

| 类型 | 操作符 | 语义说明与示例 |
| :--- | :--- | :--- |
| **字符串** | `StringEquals` / `StringNotEquals` | 精确匹配（如 `ticket:create_by == "zhangsan"`） |
| | `StringEqualsIgnoreCase` | 忽略大小写匹配（如工单状态匹配） |
| | `StringContains` | 子串包含匹配 |
| **集合多值** | `ForAnyValue:StringEquals` | 集合中只要有任一元素匹配即为真（如协同人列表） |
| | `ForAllValues:StringNotEquals` | 集合中所有元素均不匹配时为真 |
| **数值比较** | `NumericEquals` / `NumericLessThan` / `NumericGreaterThan` | 数值精确、小于、大于匹配（如配额限制） |
| **时间范围** | `DateLessThan` / `DateGreaterThan` | 早于或晚于指定 RFC3339 时间（如截止期） |
| **IP 网段** | `IpAddress` / `NotIpAddress` | 基于 `netip` 匹配单一 IP 或 CIDR 网段（如 `10.0.0.0/8`） |
| **布尔判断** | `Bool` | 布尔真假判定（如 `ticket:is_urgent == true`） |

---

## 5. 微服务全自动编译落地（以 EFlow 为例）

业务微服务无需手动拼接 SQL，通过 EIAM SDK 与 GORM 编译器只需两步即可闭环：

### 1. 路由层声明数据过滤契约
在路由定义时通过 `.AccessScope()` 声明支持的契约规范与预设模板，该契约会自动上报至管理控制台：

```go
// eflow/internal/web/ticket/handler.go
g.POST("/history", h.Define("历史工单", "history").
    Needs(permission.Template.ViewByIds, permission.Manager.Get).
    // 声明绑定 ticket_history.v1 契约及预设数据范围模板
    AccessScope(ticketpbac.HistoryProfile, ticketpbac.HistoryPresets...).
    Bind(ginx.B[HistoryReq](h.History)),
)
```

### 2. 持久层一行自动编译注入
在数据访问层（DAO），通过 `pbacgorm.Apply` 自动从上下文中提取 AccessScope 并安全编译为参数化 SQL：

```go
// eflow/internal/repository/dao/ticket.go
func (g *gormTicketDAO) ListHistory(ctx context.Context, userId string, status []int, offset, limit int64) ([]Ticket, error) {
    var result []Ticket

    // 自动提取 AccessScope 并结合白名单编译为参数化 SQL 注入 query
    query, err := pbacgorm.Apply(ctx, g.ticketQuery(ctx, userId, status), ticketpbac.History)
    if err != nil {
        return nil, err
    }

    err = query.Order("ctime desc").Limit(int(limit)).Offset(int(offset)).Find(&result).Error
    return result, err
}
```

- **参数化绑定**：所有动态值均通过 `?` 注入，杜绝 SQL 拼接与注入风险；
- **业务代码零侵入**：开发人员无需在业务逻辑中编写任何 `if/else` 拼接权限条件。

---

## 6. 默认防御原则 (Fail-Closed)

系统严格遵循失效安全防御策略，杜绝因疏漏导致的数据越权：

| 异常场景 | 防御行为 |
| :--- | :--- |
| API 声明了 AccessScope 契约但未绑定合法 Action | 直接拦截拒绝执行 |
| 策略生成了 AccessScope，但目标 API 未声明 `FilterProfile` | 直接拦截拒绝执行 |
| EIAM 返回的 Profile 契约与本地 SDK 声明不一致 | 抛出契约不匹配错误并终止执行 |
| 策略中包含业务 Profile 白名单不支持的字段 Key 或操作符 | 拒绝执行并终止数据库查询 |
