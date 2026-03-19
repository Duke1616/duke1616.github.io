# 飞书集成展示

ECMDB 将工单系统与飞书进行集成，实现了审批、执行和结果反馈的全链路消息同步。通过引入飞书可交互卡片，大部分运维操作都可以在飞书对话框内直接完成。

## 1. 提交工单与即时反馈

在 Web 平台发起发布或变更工单后，发起人及相关协作方会第一时间收到飞书通知。

### 提单确认与便捷撤回
用户在 Web 端点击提交，系统受理成功后，机器人会向发起人发一条消息卡片。
- **即时确认**：卡片展示工单摘要，确认申请已进入流程。
- **便捷撤回**：如果发现填错，**直接在卡片中填写撤回原因并点击“撤销”按钮**，无需返回 Web 后台即可终止流程。
<img src="/images/ticket/cases/feishu/start.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="提单通知" />

---

## 2. 卡片式审批中心

审批人可以直接在飞书卡片中查看详情、回填参数并进行操作，极大减少了由于切换系统带来的上下文中断。

### 2.1 交互式审批与数据补录
当流程流转到审批环节时，飞书卡片会集中展示工单的核心参数（如待发布模块、环境等）。除了信息透显，系统还支持在卡片内直接完成数据的**即时补录**：
- **参数回填**：审批卡片自带输入框，允许审批人直接补录 `GitLab CI 构建 ID` 等关键信息。
- **一键决策**：填完参数后，点击卡片下方“批准”或“拒绝”即可瞬间流转流程，无需跳转 Web 页面。

<img src="/images/ticket/cases/feishu/approval.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="审批操作" />

### 2.2 审批数据校验
系统对审批回填的数据有着严格的检查逻辑。如果审批人在飞书端填写的 `构建 ID` 格式不正确或漏填必填项，系统会实时反馈验证错误，确保流入后续自动化节点的数据 100% 准确。
<img src="/images/ticket/cases/feishu/validate.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="数据校验" />

### 2.3 抄送感知
流程流转过程中的关键变动会同步推送到相关协作人员，保持团队内部信息透明。
<img src="/images/ticket/cases/feishu/cc.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="抄送通知" />

---

## 3. 自动化任务反馈

工单中的自动化发布任务（脚本执行）运行结果，会由系统通过三段式渲染卡片同步推送至关联群组中。

### 三段式群通知
消息卡片自动按照：**1. 工单信息**、**2. 用户提交数据**（如之前回填的 ID）、**3. 执行结果**（Time/Status）进行分区展示，信息覆盖全面且层级清晰。
<img src="/images/ticket/cases/feishu/chat.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="群通知" />

### 自动化执行通知
自动化处理完成后，系统会将产生的 Stdout/Stderr 日志摘要及最终结论实时推送到干系人。发版是否成功、执行了哪些模块，卡片一目了然。
<img src="/images/ticket/cases/feishu/automation.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="自动化执行通知" />

---

## 4. 全生命周期流程追踪

工单开启到最终结单，每个状态变更都会精准同步到相关人员。

### 进度看板跳转
虽然大部分操作可在飞书完成，但仍提供详情链接，一键跳转 Web 端可视化看板，查看更复杂的流程路径图。
<img src="/images/ticket/cases/feishu/progress.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="进度查看" />

### 结单状态对齐
发起人执行撤回或流程最终圆满闭环，后台逻辑判定产出结果后，会通过机器人告知相关干系人，确保整个流程在飞书端实现“有申请、有执行、有回响”的闭环。
<img src="/images/ticket/cases/feishu/revoke.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="撤回通知" />
<img src="/images/ticket/cases/feishu/end.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="结单通知" />
