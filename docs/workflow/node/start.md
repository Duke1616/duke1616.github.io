# 🏁 开始节点 (Start Node)

开始节点是流程的唯一入口。在 ECMDB 中，它负责初始化工单实例并设置初识状态。

## 📝 核心配置

### 1. 基础信息
- **节点名称**：用于在流程图中标识入口。
- **消息通知**：`is_notify` 选项控制是否在工单发起成功后立即向相关干系人发送初始通知。

### 2. 人员权限
- **$starter**：系统默认将发起人标记为 `$starter`，作为后续流程变量（如发起人自动审批）的基础。

### 3. 数据层级
- 开始节点绑定的表单数据会持久化到工单的 `Order Data` 中。
- 这些字段在整个生命周期内全局可见，格式为 `{{key}}`。

## 💡 技术细节

在后端转换逻辑中，开始节点被赋予 `NodeType: 0`，并自动触发 `EventStart` 事件：

```go
n := model.Node{
    NodeID: node.ID, 
    NodeName: NodeName,
    NodeType: 0, 
    UserIDs: []string{"$starter"},
    NodeEndEvents: []string{"EventStart"},
}
```

---

> [!TIP]
> 建议在开始节点配置清晰的名称（如：资产申领入口），便于在审计日志中回溯。
