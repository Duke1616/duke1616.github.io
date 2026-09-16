# 契约治理工具链

在多团队协同或复杂的微服务架构下，接口鉴权经常面临两大工程痛点：
1. **权限码硬编码与拼写漂移**：开发人员在代码中使用字面量字符串（如 `"ticket:create"`），极易因大小写拼错、重构未同步导致鉴权失效。
2. **接口文档维护滞后**：传统 Swagger 需要在代码中手写海量的 `@Router`、`@Param` 注释，极易与实际入参出参结构脱节。

EIAM 提供了开箱即用的跨项目命令行工具链（CLI Tooling），将权限契约与文档生成提升至**编译期强类型安全**高度。


## 1. 强类型权限契约生成器 (`permgen`)

`permgen` 是一个基于 Go 原生 AST（抽象语法树）分析的权限契约编译器。

```text
业务 Handler 代码 (AST 扫描)
    -> 提取路由动作与权限码声明
    -> 编译生成强类型常量契约包 (pkg/contract/permission/)
    -> 生成全平台权限矩阵大盘文档 (docs/permissions.md)
```

### 核心收益
- **编译期拼写校验**：下游微服务在编写代码时，强制引用强类型常量（例如 `permission.Ticket.History`），拼写错误将在 `go build` 编译期被直接拦截。
- **权限依赖可视化**：自动分析接口间的前置依赖关系，输出全系统统一的权限矩阵大盘。

### 使用方法
```bash
# 扫描 Handler 目录并刷新当前工程的权限契约代码
permgen -s ./internal/web
```


## 2. 零注释 OpenAPI 3.0 生成器 (`swaggergen`)

传统的接口文档生成工具要求开发者在 Handler 前编写数十行复杂的文档注解，既侵入了业务逻辑，又容易遗漏维护。

`swaggergen` 采用纯静态 AST 结构体推导引擎：
- **零注释侵入**：直接解析 Handler 的入参结构体、响应包装类及路由规则。
- **标准规范输出**：自动导出标准的 OpenAPI 3.0 `swagger.json`。
- **内嵌交互式预览**：可直接生成三栏式离线交互网页（无需搭建额外的 Swagger UI 服务器），支持在线调试与 Bearer Token 鉴权。

### 使用方法
```bash
# 扫描业务代码并导出 OpenAPI 规范与静态预览页面
swaggergen -s ./internal/web -o ./api/docs/swagger.json --html ./api/docs/index.html
```


## 3. 跨微服务集成接入规范

EIAM 的工具链已解耦为可独立分发的二进制 CLI。生态内的各个微服务仓库（如 `ecmdb`、`etask`、`eflow`）无需复制代码，可通过 Go 官方工具链直接全局安装：

```bash
# 1. 安装权限契约生成器
go install github.com/Duke1616/eiam/cmd/permgen@latest

# 2. 安装零注释 OpenAPI 生成器
go install github.com/Duke1616/eiam/cmd/swaggergen@latest
```

在各下游微服务项目的根目录中，推荐通过 `Taskfile.yaml` 将其固化为自动化任务：

```yaml
# Taskfile.yaml 示例
tasks:
  gen:perm:
    desc: 扫描路由并导出强类型权限契约
    cmds:
      - permgen -s ./internal/web

  gen:swagger:
    desc: 零注释导出 OpenAPI 3.0 规范
    cmds:
      - swaggergen -s ./internal/web -o ./api/docs/swagger.json --html ./api/docs/index.html
```

> [!NOTE]
> 部署 EIAM 基础服务与整体环境准备，请参阅 [生产部署与高可用指南](/system/deploy)。
