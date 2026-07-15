# set_conn_config

## 命令功能

设置设备连接配置文件地址，用于在线分析场景。工具会读取该文件中配置的主机、BMC、交换机连接信息，并加密存储，设置成功后建议尽快删除包含明文密码的源文件。

## 命令格式

| 命令格式 | 描述        |
|---------|-----------|
| `set_conn_config <文件地址>` | 设置连接文件地址  |
| `set_conn_config ?` | 查看详情      |

## 参数说明

| 参数 | 类型 | 是否必填 | 说明 |
|------|-----|------|------|
| `<文件地址>` | string | 是 | 连接配置文件的路径。 |

## 配置文件结构

```conn.ini
[host]
# port指定端口,不写默认22, username指定用户名, password指定密码, private_key指定私钥文件
1.1.1.1 port="22" username="root" private_key="~/.ssh/your_private_key"
1.1.1.2 port="22" username="root" password="<your_password>"

[bmc]
1.1.1.3 username="Administrator" password="<your_password>"

[switch]
# 支持ip1-ip2 ip段方式填写(需保证账号密码相同), 通过step设置步长
1.1.1.4-1.1.1.10 step=2 username="root" password="<your_password>"

[config]
# 支持设置全局的私钥文件
private_key="~/.ssh/your_private_key"
```

集群设备连接信息配置说明：

1. 工具使用时，需要连接到各个设备上进行命令查询日志采集等操作。请参考以上 `conn.ini` 配置文件格式，配置集群连接信息。
2. 若支持免密连接，可以不填写密码。
3. 支持密钥方式登录。`private_key` 配置为私钥路径配置，可在每行单独配置，也可以在 `config` 选项框中为集群所有的环境配置。
4. 其中 ip 支持单 IP 配置，也支持 IP 范围配置（`step` 配置：step 默认为 1；确保用户名和密码相同）。
5. 其中 1620 前台等交换设备也视为交换机，请填写到 `switch` 选项框中。

> **安全提示**：`conn.ini` 中包含设备登录凭据，属于敏感信息。建议：
>
> - 配置完成后及时删除包含明文密码的配置文件（工具设置成功后会提示）。
> - 优先使用私钥免密方式登录，避免在配置文件中填写明文密码。
> - 限制配置文件的访问权限（如 `chmod 600 conn.ini`），仅限当前用户读取。

## 输出说明

- 设置成功时返回：`设置成功，请尽快删除包含明文密码的配置文件`。
- 设置失败时返回：`设置地址失败，异常：{err}` 或 `地址为空，请重新设置` 或 `地址{file_path}不存在，请重新设置` 或 `地址{file_path}不是文件，请重新设置`。

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect_diag
设置成功，请尽快删除包含明文密码的配置文件
# 其他日志输出...
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_conn_config /home/user/conn.ini
设置成功，请尽快删除包含明文密码的配置文件
```
