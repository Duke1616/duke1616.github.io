# 版本发布

通过工单系统将代码发布流程标准化，实现从构建到生产部署的全链路管控。

### 业务概览
- **发起方**：通常由产品经理或 Release Manager 发起发布工单。
- **构建机制**：代码通过 GitLab CI 进行流水线构建并产出镜像。
- **协作模式**：前端、后端、算法等研发人员分别在工单表单中填写其对应的构建 ID。
- **动态执行**：系统支持自动化分支判定，若本次版本不涉及某个端的更新，对应的自动化节点将依据逻辑自动“跳过”执行。

## 1、流程设计

整体流程分为：**提交发布申请 → 研发人员补充构建 ID / 用户审批 → 自动化发布执行 → 群通知执行结果**。

<img src="/images/ticket/cases/deploy/workflow.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="发布流程设计" />

---

## 2、模版与执行项配置

在发起发布工单前，需预设好**工单申请表单**与**自动化执行模版**，两者共同决定了数据的输入与处理逻辑。

### 2.1 工单模版

<img src="/images/ticket/cases/deploy/template.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="工单模版设计" />

**核心字段说明：**
- **环境**：多选框（支持后端、前端、算法），用于指明本次发布的影响范围，后续网关节点将据此进行流程分支跳转。
- **版本**：单行输入框，定义本次发布的版本标识。
- **版本详情**：多行输入框，用于记录详细的变更日志或发布备注。

### 2.2 执行单元

自动化节点并不直接配置执行逻辑，而是引用预设的“代码模版”。执行单元的相关环境与处理器已在模版中提前定义。

<img src="/images/ticket/cases/deploy/runner.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="Runner配置" />


## 3、流程节点配置

详细说明流程图中各节点的作用及其配置详情。

### 3.1 用户节点

配置审批人并收集研发填写的关键数据（如构建 ID）。审批人收到待办后进入详情页进行处理。

<img src="/images/ticket/cases/deploy/user.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="用户节点配置" />

**配置示例：**
- **节点名称**：前端（示例）
- **审批人员**：指定人员（如：栾凯朝）
- **审批模式**：单人处理，若涉及多方协作可配置各个端的专用处理节点。

**表单字段配置详情：**

| 配置项 | 字段 1：前端 ID | 字段 2：提示信息 |
| :--- | :--- | :--- |
| **显示名称** | 前端 ID | 提示信息 |
| **字段标识** | `frontend` | `tips` |
| **字段类型** | 单行文本 | 提示信息 |
| **必填项** | ✅ 开启 | ❌ 关闭 |
| **只读模式** | ❌ 关闭 | ✅ 开启 |
| **占位提示 / 内容** | 请输入 gitlab ci 构建ID | 可检查是否有配置需要调整；如需变更，及时与运维沟通并同步处理。 |
| **正则校验** | `^[a-z0-9]{8}$` | - |

> [!NOTE]
> 以上配置示例以**前端发布**为例。若涉及**算法或后端**发布，其审批表单的配置逻辑与之相同，仅需根据业务需求调整“显示名称”与“字段标识”。

---

### 3.2 自动化节点

当流程流转至此节点时，系统会自动调用上文 2.2 章节中预设的“环境更新”代码模版。

<img src="/images/ticket/cases/deploy/automation.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="自动化节点配置" />

**节点关联说明：**
- **节点名称**：自动化-部署
- **执行标签**：`deploy_test`（需与模版及 Agent 标签保持一致）
- **参数传递**：系统会自动将表单中填写的构建 ID（如 `frontend`）转化为 JSON 格式，作为第一个参数传递给对应的 Shell 脚本（详见第 4 章节）。



---

### 3.3 条件连线

在流程图中，网关节点后的**连接线**承担了逻辑判定的职责。通过在连线上配置表达式，系统可以动态决定流程的流转方向。

<img src="/images/ticket/cases/deploy/express.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="连线条件配置" />

**配置说明：**
- **连线属性**：逻辑判定并非在网关节点本身，而是由其引出的各条“连线”定义的。
- **按需流转**：通过在连线上编写 SQL 表达式（如 `$environment in ('backend')`），控制该分支的激活条件。若条件满足，流程进入该分支；若不满足，系统将自动跳过该分支的所有后续任务。
- **并行解耦**：这种设计允许一个网关引出多条具有独立条件的连线，从而实现前端、后端、算法等任务的并发、按需执行。

---

### 3.4 群通知节点

发布任务完成后，系统将通过飞书机器人推送详细的执行结果卡片，实现发布进度的“即时反馈”与信息“透明化”。

<img src="/images/ticket/cases/deploy/chat_01.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="群通知配置" />

<img src="/images/ticket/cases/deploy/chat_02.png" style="width: 100%; border-radius: 8px; margin: 16px 0;" alt="群通知消息预览" />

**配置亮点：**
- **智能标题**：采用统一的标题规则（如 `发版执行结果`），方便用户快速识别消息意图。
- **精准推送**：配置关联特定的飞书群组（如：`【ECMDB】- 环境发版`），确保广播信息发送给相关干系人，避免信息炸弹。

## 4、自动化任务脚本

以下是自动化节点所关联的完整发布脚本逻辑。脚本通过解析工单上下文 `args` 中的 JSON 数据，实现对指定服务（前端、后端、算法）的精准镜像拉取与重启，从而达成镜像发布的自动化闭环。

```shell
#!/bin/bash
# 脚本描述信息：部署 docker-compose 服务，支持多模块部分更新逻辑

args=$1
vars=$2
source "$vars"

# 引入内部结果输出工具
source ./third_party/utils/want_result.sh

COMPOSE_FILE="/app/agent/docker-compose.yml"
LOG_FILE="/app/agent/deploy.log"

# -----------------------------
# 1. 变量解析与数据预处理
# -----------------------------
frontend=$(echo "$args" | jq -r '.frontend')
backend=$(echo "$args" | jq -r '.backend')
agent=$(echo "$args" | jq -r '.agent')

get_current_tag() {
    local service=$1
    grep "$service" -A2 "$COMPOSE_FILE" | grep image | awk -F: '{print $NF}'
}

# -----------------------------
# 2. 准备更新环境（处理跳过逻辑）
# -----------------------------
prepare_updates() {
    CURRENT_FRONTEND=$(get_current_tag frontend)
    CURRENT_BACKEND=$(get_current_tag backend)
    CURRENT_AGENT=$(get_current_tag agent)

    # 处理来自工单的空值输入
    [ "$frontend" = "null" ] && frontend=""
    [ "$backend"  = "null" ] && backend=""
    [ "$agent"    = "null" ] && agent=""

    # 确定最终使用的镜像 Tag（优先使用补充的 ID，否则保留当前值）
    FRONTEND_TAG=${frontend:-$CURRENT_FRONTEND}
    BACKEND_TAG=${backend:-$CURRENT_BACKEND}
    AGENT_TAG=${agent:-$CURRENT_AGENT}

    # 构建需要执行的具体服务列表
    UPDATED_SERVICES=()
    [ -n "$frontend" ] && UPDATED_SERVICES+=("frontend")
    [ -n "$backend"  ] && UPDATED_SERVICES+=("backend")
    [ -n "$agent"    ] && UPDATED_SERVICES+=("agent")

    export FRONTEND_TAG BACKEND_TAG AGENT_TAG
}

# -----------------------------
# 3. 核心执行流程：拉取与重启
# -----------------------------
pull_images() {
    if [ ${#UPDATED_SERVICES[@]} -eq 0 ]; then
        echo "$(date '+%F %T') - 无目标服务更新数据，跳过发布流程" | tee -a "$LOG_FILE"
        return
    fi
    echo "$(date '+%F %T') - 执行拉取操作: ${UPDATED_SERVICES[*]}" | tee -a "$LOG_FILE"
    docker-compose -f "$COMPOSE_FILE" pull "${UPDATED_SERVICES[@]}"
}

restart_services() {
    if [ ${#UPDATED_SERVICES[@]} -eq 0 ]; then
        return
    fi
    echo "$(date '+%F %T') - 执行滚动重启: ${UPDATED_SERVICES[*]}" | tee -a "$LOG_FILE"
    docker-compose -f "$COMPOSE_FILE" up -d --no-deps "${UPDATED_SERVICES[@]}"
}

# -----------------------------
# 4. 统计结果并回传系统
# -----------------------------
print_result() {
    want_result "发布状态" "success"
    want_result "完成时间" "$(date '+%F %T')"
    [ ${#UPDATED_SERVICES[@]} -gt 0 ] && want_result "更新集合" "${UPDATED_SERVICES[*]}"
}

# -----------------------------
# 入口函数
# -----------------------------
main() {
    prepare_updates
    pull_images
    restart_services
    print_result
}

main "$@"
```

---

> [!TIP]
> **关于按需执行逻辑（Skip）**：
> 自动化脚本通过检测流程变量是否为空来实现 **“定向精准更新”**。
> - **逻辑实现**：若提交工单阶段未填写某个模块（如：算法端）的构建 ID，脚本将自动将其从更新队列中剔除。
> - **业务价值**：通过锁定非目标模块的现有运行状态（既不拉取新镜像也不触发重启），确保了在混合部署环境下的 **零干扰发布**。这允许在一个复杂的流程模板中，根据实际业务需求只发布某一个特定端，而保持其他服务的绝对稳定。
