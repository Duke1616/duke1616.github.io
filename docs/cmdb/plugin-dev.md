# 插件契约与模型注入

ECMDB 采用声明式契约驱动机制：插件开发者只需在 Go 结构体中声明所需字段与拓扑关系。在管理台保存插件绑定时，系统自动完成模型创建、字段映射、拓扑注入与敏感凭据加密。

---

## 1. 核心推导能力

### 1.1 自动建模型
插件使用未在 CMDB 中定义的全新模型时（如跳板机网关模型 `AuthGateway`）：
- 结构体实现自描述方法（`DescribeModel` 返回模型名称与分组）；
- 保存绑定时，系统自动创建模型、分组与属性字段。

### 1.2 字段映射（`field`）
存量模型的物理字段名与插件逻辑字段不一致时：
- 结构体标签使用 `field` 参数指定物理字段名（如 `field=ip`）；
- 系统在数据交互时自动完成映射，插件代码无需适配底层表结构。

### 1.3 拓扑关系注入（`in` / `out`）
插件依赖关联模型（如主机依赖跳板机网关）时：
- 在结构体中声明关联模型的切片字段；
- 标签指定关联模型与方向：
  - `out=default`：当前模型指向目标模型；
  - `in=default`：目标模型指向当前模型；
- 保存绑定时，系统自动建立两端模型的拓扑准入规则。

### 1.4 字段类型与枚举推导
内省引擎根据 Go 原生类型与接口自动推导 CMDB 字段规格：

| Go 类型定义 | 推导类型 | 说明 |
| :--- | :--- | :--- |
| 实现 `EnumOptions() []string` | `list` | 自动提取枚举列表作为下拉选项 |
| `[]byte` | `multiline` | 多行文本，结合敏感字段检测自动启用密文存储 |
| `time.Time` | `datetime` | 日期时间 |
| `bool` | `boolean` | 布尔开关 |
| `int` / `float` 等数值类型 | `number` | 数值校验 |
| 其他类型 | `string` | 单行文本 |

> 也可在标签中通过 `type=multiline` 或 `options=passwd,publickey` 手动指定。

### 1.5 敏感凭据加密
- 字段名包含 `password`、`private_key`、`secret`、`token` 时，系统自动标记为加密字段；
- 底层使用 AES-GCM 密文存储，查询默认脱敏；
- 仅在发起连接时临时在内存中解密下发，连接建立后即刻销毁。

---

## 2. 结构体声明示例

以 SSH 插件为例，展示枚举、私钥与网关拓扑声明：

```go
package define

import "github.com/Duke1616/ecmdb-plugins/pkg/model"

// 认证方式枚举
type AuthType string

const (
    AuthTypePasswd     AuthType = "passwd"
    AuthTypePublicKey  AuthType = "publickey"
    AuthTypePassphrase AuthType = "passphrase"
)

// 实现枚举接口，推导为下拉单选 list
func (AuthType) EnumOptions() []string {
    return []string{
        string(AuthTypePasswd),
        string(AuthTypePublicKey),
        string(AuthTypePassphrase),
    }
}

// 主机连接端点
type Endpoint struct {
    model.BaseResource
    Host       string   `plugin:"host,field=ip,label=主机地址,required"`
    Port       int      `plugin:"port,label=SSH端口,default=22"`
    Username   string   `plugin:"username,label=登录账号,required"`
    Password   string   `plugin:"password,label=登录密码"`
    PrivateKey []byte   `plugin:"private_key,label=私钥凭证"`
    AuthType   AuthType `plugin:"auth_type,label=认证方式,default=passwd"`
    Sort       int      `plugin:"sort,label=排序权重"`
}

// 跳板机网关
type Gateway struct {
    model.BaseResource
    Host       string   `plugin:"host,label=网关地址,required"`
    Port       int      `plugin:"port,label=网关端口,default=22"`
    Username   string   `plugin:"username,label=网关账号,required"`
    Password   string   `plugin:"password,label=网关密码"`
    PrivateKey []byte   `plugin:"private_key,label=网关私钥"`
    AuthType   AuthType `plugin:"auth_type,label=认证方式,default=passwd"`
    Sort       int      `plugin:"sort,label=排序权重"`
}

// 目标资产：组合端点与级联网关
type ConnectionTarget struct {
    Endpoint
    Gateways []Gateway `plugin:"gateways,model=AuthGateway,name=跳板机网关,group=安全凭据,in=default"`
}

// 模型不存在时自动创建的中文名与分组
func (ConnectionTarget) DescribeModel() (string, string) {
    return "主机资产", "计算资源"
}
```

---

## 3. 插件契约导出

在插件初始化时通过 `Target[T]` 注册：

```go
package define

import "github.com/Duke1616/ecmdb/pkg/plugin"

func (p Provider) Definition() (plugin.Definition, error) {
    reg := plugin.NewRegistry(
        PluginUID,
        "SSH",
        plugin.Type("builtin"),
        plugin.Version("1.0.1"),
        plugin.Description("基于 CMDB 主机和登录网关关系提供 SSH 终端与 SFTP 文件管理能力。"),
        plugin.ExternalServiceRuntime(p.cfg.Upstream, plugin.RuntimeHealthPath("/healthz")),
    )

    return plugin.Target[ConnectionTarget](reg, ModelHost).
        Workspace(
            ActionTerminal,
            "Web Shell",
            plugin.Icon("terminal"),
            plugin.Permission(PermissionConnect),
            plugin.CardFields("name", "ip"),
            plugin.Prop("connectionType", "Web Shell"),
        ).
        Workspace(
            ActionSFTP,
            "Web Sftp",
            plugin.Icon("folder"),
            plugin.Permission(PermissionConnect),
            plugin.CardFields("name", "ip"),
            plugin.Prop("connectionType", "Web Sftp"),
        ).
        Definition()
}
```

---

## 4. 获取解密后的强类型数据

连接请求触发时，插件通过 `InputRootOne` 直接读取主站解析好的强类型对象：

```go
package define

import (
    "github.com/Duke1616/ecmdb/pkg/plugin/codec"
    "github.com/Duke1616/ecmdb/pkg/plugin/types"
)

// DecodeTarget 从上下文反序列化已解密的连接目标
func DecodeTarget(actionCtx types.ActionContext) (ConnectionTarget, error) {
    return codec.InputRootOne[ConnectionTarget](actionCtx)
}
```

---

## 5. 标签参数说明

Go 结构体字段在内省推导时按类型分为两类：**普通属性字段**与**关联模型字段**。两者的标签参数作用域完全不同：

### 5.1 普通属性字段（基础数据类型）

用于修饰资产自身的属性（如 `string`、`int`、`[]byte`、枚举等）：

```go
Host       string   `plugin:"host,field=ip,label=主机地址,required"`
Port       int      `plugin:"port,label=SSH端口,default=22"`
AuthType   AuthType `plugin:"auth_type,label=认证方式,default=passwd"`
```

| 参数 | 作用说明 | 示例 | 默认行为 |
| :--- | :--- | :--- | :--- |
| **首参数** | 字段逻辑标识（Go 侧取数键，推荐显式声明） | `host`、`port` | 缺省时优先取 `json` 标签，其次取首字母小写的字段名（如 `Host` -> `host`） |
| `field` | 映射到的 CMDB 模型物理属性 UID | `field=ip` | 缺省与首参数一致 |
| `label` | 属性在前端界面展示的中文名称 | `label=主机地址` | 缺省与首参数一致 |
| `type` | 显式指定 CMDB 字段类型（覆盖默认推导） | `type=multiline` | 依据 Go 类型或枚举自动推导 |
| `options` | 下拉列表候选项（逗号分隔） | `options=passwd,publickey` | 优先调用 `EnumOptions()` |
| `default` | 默认缺省值 | `default=22`、`default=passwd` | 无默认值 |
| `required` | 标记该属性为必填项 | `required` | 默认为选填 |
| `secure` | 标记为敏感加密属性（落库 AES 信封加密、列表默认脱敏不公开展示，运行时自动解密注入） | `secure`、`secure=false` | 默认依据字段名是否包含 `password`、`private_key`、`secret`、`token` 智能推导；显式声明 `secure=false` 可覆盖关闭（支持别名 `encrypt`） |

---

### 5.2 关联模型字段（结构体 / 切片）

用于修饰关联的另一张资产模型及其拓扑关系（如 `[]Gateway`、`Database` 等）：

```go
Gateways []Gateway `plugin:"gateways,model=AuthGateway,name=跳板机网关,group=安全凭据,in=default"`
```

| 参数 | 作用说明 | 示例 | 默认行为 |
| :--- | :--- | :--- | :--- |
| **首参数** | 拓扑子节点逻辑标识（推荐显式声明） | `gateways` | 缺省时优先取 `json` 标签，其次取首字母小写的字段名（如 `Gateways` -> `gateways`） |
| `model` | 关联的目标 CMDB 模型 UID | `model=AuthGateway` | 必填（声明关联子模型） |
| `name` | 目标模型中文名（自动建模型时使用） | `name=跳板机网关` | 缺省与 `model` 相同 |
| `group` | 目标模型所属分组（自动建模型时使用） | `group=安全凭据` | 缺省归入默认基础分组 |
| `in` | 反向关联（目标模型 -> 当前模型） | `in=default`、`in=run` | 与 `out` 二选一 |
| `out` | 正向关联（当前模型 -> 目标模型） | `out=default` | 与 `in` 二选一 |
| `cardinality` | 关联数量基数（`one` / `many`） | `cardinality=many` | 切片默认 `many`，单结构体默认 `one` |
| `required` | 标记拓扑链路为强依赖（不可为空） | `required` | 默认为选填 |

