# set_config_dir

## 命令功能

设置配置文件目录路径，工具会自动扫描该目录下的配置文件并加载。当前支持加载机房位置配置文件 `LLD.xlsx`（包含灵衢 L1/L2 网络对应关系），用于在诊断报告中关联机柜、机房等位置维度信息。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `set_config_dir <目录路径>` | 设置配置文件目录路径 |
| `set_config_dir ?` | 查看详情 |

## 参数说明

| 参数 | 类型 | 是否必填 | 说明 |
|------|-----|------|------|
| `<文件目录>` | string | 是 | 配置文件所在目录路径。目录内需包含 `LLD.xlsx`。 |

## LLD.xlsx 文件结构

需要提供 `LLD.xlsx` 文件，样例文件可参考：[LLD.xlsx](../../../../resource/LLD.xlsx)。该文件包含两个 Sheet：

| Sheet 名 | 必填项 | 用途 |
|----------|--------|------|
| 灵衢L1网络对应关系 | 服务器、机房名称、机柜编号、主机 SN、L1 名称、L1_IP、L1_SN | 描述主机与 L1 交换机的对应关系 |
| 灵衢L2网络对应关系 | 设备名、机房名称、机柜编号、管理 IP 配置、SN | 描述 L2 交换机的机房位置信息 |

> `LLD.xlsx` 为可选配置，未设置时工具仍可工作，但不会进行机房位置维度的关联分析。

## 输出说明

- 设置成功时返回：`设置成功，配置目录：{dir_path}`。
- 设置失败时返回：`目录路径为空，请重新设置` 或 `目录{dir_path}不存在，请重新设置` 或 `路径{dir_path}不是目录，请重新设置`。

## 示例

非交互式方式（展示命令与回显）：

```bash
ascend-fd-tk set_config_dir /home/user/config set_conn_config /home/user/conn.ini auto_collect_diag
设置成功，配置目录：/home/user/config
# 其他日志输出...
```

交互式方式（展示命令与回显）：

```bash
ascend-fd-tk
>>> set_config_dir /home/user/config
设置成功，配置目录：/home/user/config
```
