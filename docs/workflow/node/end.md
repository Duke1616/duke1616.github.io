# 🏁 结束节点 (End Node)

结束节点标志着工单完整生命周期的终结。

## 💾 结算操作

当流程到达结束节点后，系统会执行以下核心操作：

1. **状态转变**：工单从「流转中」变更为「已结束」。
2. **事件触发**：触发 `EventNotify` 事件，通知发起人最终结果。
3. **数据固化**：锁定所有 `Order Data`，不再允许任何修改。
4. **历史存档**：生成完成的审批链条快照，提供给审计页面。

## ⚙️ 技术定义

在后端代码中，结束节点被标记为 `NodeType: 3`。

```go
n := model.Node{
    NodeID: node.ID, 
    NodeName: NodeName,
    NodeType: 3, 
    PrevNodeIDs: l.FindPrevNodeIDs(node.ID),
    NodeStartEvents: []string{"EventNotify"},
}
```

---

> [!NOTE]
> 一个流程可以配置多个结束节点，代表不同的业务结局（如：已撤销、已执行、执行失败）。
