# Ansible 凭据与受控端互信

ETask 原生支持 Ansible Playbook 编排，核心设计原则是**代码与认证材料物理隔离**。Playbook 仓库、任务参数、调度消息中绝不包含明文私钥或口令，仅携带非敏感凭据别名（如 `production-linux`），并在执行节点运行时动态注入。

---

## 1. 核心隔离机制与安全红线

* **动态临时注入**：私钥仅在任务执行瞬间以 `0600` 权限拷贝至独占沙箱工作区，密码仅写入临时的 Extra Vars 文件；任务完成工作区物理清除，认证材料绝不进入命令行参数或审计日志。
* **强制严格主机指纹校验**：系统强制开启 `StrictHostKeyChecking=yes`，严格基于节点本地的 `known_hosts_file` 验签，彻底杜绝内网中间人劫持风险。
* **本地宿主目录权限底线**：`credential_root` 必须是绝对路径，目录本身及内部所有认证文件严禁组或其他用户写入（目录必须为 `0700`、文件必须为 `0600`）。若权限过宽，节点启动时将直接阻断退出，杜绝隐性带病运行。

---

## 2. 控制节点配置与本地安全加固

在部署了 ETask 单机版或独立 Executor 节点的宿主机上，需规范准备凭据根目录与主机指纹文件。

### 2.1 节点配置声明

在节点配置文件（`config.yaml`）中开启 Ansible 驱动并声明凭据目录：

```yaml
runner:
  ansible:
    enabled: true
    binary: "ansible-playbook"
    sshpass_binary: "sshpass"                    # 密码型凭据依赖该工具
    credential_root: "/run/credentials/etask-ansible" # 宿主机安全凭据根目录
    known_hosts_file: "/etc/etask/ssh/known_hosts"    # 集中主机指纹库
    credentials:
      production-linux:
        type: "private_key"
        username: "etask"
        private_key_file: "production-linux-key"
      legacy-linux:
        type: "password"
        username: "root"
        password_file: "legacy-linux-password"
```

### 2.2 宿主机目录与权限初始化

```bash
# 1. 创建凭据宿主安全目录并限定独占权限 (0700)
mkdir -p /run/credentials/etask-ansible
chmod 700 /run/credentials/etask-ansible

# 2. 放置私钥与密码文件，确保仅属主可读写 (0600)
touch /run/credentials/etask-ansible/production-linux-key
touch /run/credentials/etask-ansible/legacy-linux-password
chmod 600 /run/credentials/etask-ansible/*

# 3. 采集受控主机 SSH 指纹至 known_hosts (防中间人劫持)
mkdir -p /etc/etask/ssh
ssh-keyscan -H 10.0.1.11 10.0.1.12 >> /etc/etask/ssh/known_hosts
chmod 644 /etc/etask/ssh/known_hosts

# 4. 密码型凭据 (type: password) 必须预装 sshpass 工具
apt-get install -y sshpass || yum install -y sshpass
```

---

## 3. 受控目标主机免密授权

Ansible 底层基于标准 SSH 协议与受控目标机通信。ETask 控制端持有私钥，**受控目标机必须持有配对的公钥才能完成非对称签名校验**。将主机纳入 ETask 管控范围前，需完成受控端初始化：

### 第一步：提取公钥
```bash
# 在保存私钥的控制端宿主机或安全运维机提取
ssh-keygen -y -f /run/credentials/etask-ansible/production-linux-key > production-linux-key.pub
```

### 第二步：注入受控主机认证列表
将提取的公钥追加到目标受控节点运行账号（如 `etask` 或 `root`）的认证文件中，并严格设置 SSH 权限规范：
```bash
# 登录目标受控主机执行 (或通过装机初始化脚本、带外管理批量推送)
mkdir -p ~/.ssh
chmod 700 ~/.ssh

# 追加公钥并设置专属访问权限 (SSH 强制拒绝权限过宽的认证文件)
cat production-linux-key.pub >> ~/.ssh/authorized_keys
chmod 600 ~/.ssh/authorized_keys
```

### 第三步：受控端 SSHD 服务合规检查
检查目标主机 `/etc/ssh/sshd_config`，确认已开启公钥认证：
```ini
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
StrictModes yes
```

::: tip 排错锦囊
若执行 Playbook 提示 `Permission denied (publickey)`，九成原因为受控端 `~/.ssh` 目录拥有组写权限（如 `775`）或 `authorized_keys` 权限超过 `600`，被 SSHD 的 `StrictModes` 机制主动拒绝。
:::

---

## 4. Playbook 资产清单引用实战

在 Playbook 项目代码仓库中，只需在静态 YAML 或 INI Inventory 中声明 `etask_credential_ref`，即可实现不同主机组绑定不同安全凭据：

```yaml
all:
  children:
    # 核心生产集群：采用私钥认证
    production_nodes:
      vars:
        etask_credential_ref: production-linux
      hosts:
        10.0.1.11:
        10.0.1.12:

    # 遗留系统节点：采用受控主机账号密码认证
    legacy_nodes:
      vars:
        etask_credential_ref: legacy-linux
      hosts:
        10.0.2.21:
```

运行时节点按照「主机变量 > 子组变量 > 父组变量 > 任务默认凭据」优先级精准匹配，无需在项目代码中保留任何真实密钥材料。

::: tip 相关参考与扩展
* **节点完整配置**：关于包含 Ansible 驱动的执行端节点完整配置，详见 [运行模式与配置实战](/task/startup)。
* **引擎运行时契约**：关于 Ansible 工程项目物化与独占沙箱工作区生命周期，详见 [异构执行引擎与 SDK](/task/engine)。
:::
