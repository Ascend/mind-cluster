# set_config_dir

## 命令功能

设置配置文件目录路径，工具会自动扫描该目录下的配置文件并加载。当前支持加载机房位置配置文件 `LLD.xlsx`（包含灵衢 L1/L2 网络对应关系），用于在诊断报告中关联机柜、机房等位置维度信息；同时支持加载阈值配置文件 `threshold_config.json` 覆盖默认诊断阈值。

## 命令格式

| 命令格式 | 描述 |
|---------|------|
| `set_config_dir <目录路径>` | 设置配置文件目录路径 |
| `set_config_dir ?` | 查看详情 |

## 参数说明

| 参数 | 类型 | 是否必填 | 说明                                                                  |
|------|-----|------|---------------------------------------------------------------------|
| `<目录路径>` | string | 是 | 配置文件所在目录路径。目录内需包含 `LLD.xlsx`、`threshold_config.json`。 |

## LLD.xlsx 文件结构

需要提供 `LLD.xlsx` 文件，样例文件可参考：[LLD.xlsx](../../../../resource/LLD.xlsx)。该文件需包含两个 Sheet：

| Sheet 名 | 必填项 | 用途 |
|----------|--------|------|
| 灵衢L1网络对应关系 | 服务器、机房名称、机柜编号、主机 SN、L1 名称、L1_IP、L1_SN | 描述主机与 L1 交换机的对应关系 |
| 灵衢L2网络对应关系 | 设备名、机房名称、机柜编号、管理 IP 配置、SN | 描述 L2 交换机的机房位置信息 |

## 阈值配置文件说明

阈值配置文件 `threshold_config.json` 用于覆盖各诊断项的默认阈值，便于根据实际环境调整判障标准。该文件需放置在 `set_config_dir` 设置的配置目录下，工具加载配置时自动读取；未放置该文件时使用代码内置默认阈值。完整示例文件可参考[threshold_config.json](../../../../resource/threshold_config.json)。

配置文件为 JSON 对象，每个配置项对应一个阈值名称，阈值字段统一嵌套在 `threshold` 子对象内，例：

```json
{
    "TX_POWER_MW": {"threshold": {"low_value_alarm": "0.2", "high_value_alarm": "2.5", "desc": "tx power", "unit": "mW"}},
    "HOST_SNR_DB": {"threshold": {"low_value_warn": "20", "low_value_alarm": "18", "desc": "host snr", "unit": "dB"}}
}
```

阈值字段说明：

| 字段 | 类型 | 说明                                  |
|------|------|-------------------------------------|
| `low_value_alarm` / `high_value_alarm` | string | 低值/高值告警阈值（数值类字符串）                   |
| `low_value_warn` / `high_value_warn` | string | 低值/高值预警阈值（数值字符串）                       |
| `normal_value_alarm` | string | 正常值判定字符串（如 `"UP"`、`"present"`）      |
| `desc` | string | 阈值描述，最长 128 字符                      |
| `unit` | string | 单位（如 `"dB"`、`"mW"`、`"mA"`），最长 16 字符 |

阈值配置项含义：

| 配置项 | 含义                               | 单位 |
|--------|----------------------------------|------|
| `CDR_HOST_SNR_DB` | CDR（时钟数据恢复）主机侧信噪比                | dB |
| `CDR_MEDIA_SNR_DB` | CDR（时钟数据恢复）介质侧信噪比                | dB |
| `CHIP_CPU_PORT_SNR_LINE` | CPU 与 L1 交换机间端口 SNR（线性值，约 56 dB） | - |
| `CHIP_NPU_PORT_SNR_LINE` | NPU 与 L1 交换机间端口 SNR（线性值）         | - |
| `SWITCH_PORT_SNR_LINE` | L1 与 L2 交换机间端口 SNR（线性值）          | - |
| `DUPLEX_THRESHOLD` | 端口双工模式，正常值 `Full`                | - |
| `NET_HEALTH_THRESHOLD` | 网络健康检查结果，正常值 `Success`           | - |
| `LINK_STATUS_THRESHOLD` | 链路状态，正常值 `UP`                    | - |
| `OPTICAL_PRESENT_THRESHOLD` | 光模块在位状态，正常值 `present`            | - |
| `SNR_LANE_DIFF_DB` | 光模块各 lane 间 SNR 的最大差值            | dB |
| `POWER_LANE_DIFF_DB` | 光模块各 lane 间光功率的最大差值              | dBm |
| `HCCN_LINK_DOWN_CNT` | 24 小时内 HCCN 链路 down 次数           | 次 |
| `TX_POWER_MW` | host侧光模块发射光功率（mW 口径）             | mW |
| `RX_POWER_MW` | host侧光模块接收光功率（mW 口径）             | mW |
| `TX_POWER_DBM` | switch侧光模块发射光功率（dBm 口径）          | dBm |
| `RX_POWER_DBM` | switch侧光模块接收光功率（dBm 口径）            | dBm |
| `TX_BIAS_MA` | 光模块发射偏置电流                        | mA |
| `HOST_SNR_DB` | 光模块主机侧信噪比（电口/芯片侧）                | dB |
| `MEDIA_SNR_DB` | 光模块介质侧信噪比（光口/线缆侧）                | dB |
| `NIC_TX_BIAS_MA` | 网卡（NIC SFP）光模块发射偏置电流             | mA |
| `NIC_TX_POWER_DBM` | 网卡（NIC SFP）光模块发射光功率              | dBm |
| `NIC_RX_POWER_DBM` | 网卡（NIC SFP）光模块接收光功率              | dBm |
| `NIC_HOST_SNR_DB` | 网卡（NIC SFP）光模块主机侧信噪比             | dB |
| `NIC_MEDIA_SNR_DB` | 网卡（NIC SFP）光模块介质侧信噪比             | dB |
| `NIC_LANE_FLAG` | 网卡 lane 异常标志，正常值 `0`             | - |

说明：

- `DUPLEX_THRESHOLD`、`NET_HEALTH_THRESHOLD`、`LINK_STATUS_THRESHOLD`、`OPTICAL_PRESENT_THRESHOLD`、`NIC_LANE_FLAG` 为字符串正常值判定：仅当实际值等于配置的正常值时判定为正常，其余均判定为异常。
- `HCCN_LINK_DOWN_CNT` 超过预警阈值判定亚健康，超过告警（故障）阈值判定异常。
- 网卡（NIC）相关阈值仅 Ascend 950系列产品使用。

配置校验规则：

- 未配置的字段保持代码内置默认值；值为空（`null` 或空白字符串）时同样视为未配置，使用默认值。
- 数值字段（`low_value_alarm` 等）必须为合法数值类字符串，否则该字段被忽略。
- `desc`/`unit`/`normal_value_alarm` 必须为字符串，非字符串时该字段被忽略。
- `desc` 超过 128 字符、`unit` 超过 16 字符时该字段被忽略。不建议修改`desc`和`unit`，因为会影响判断和结果展示。
- 未知字段名、非对象配置项等非法内容会被忽略并记录告警日志，不影响其他配置生效。

>[!NOTE]
> `set_config_dir` 为可选配置命令，未设置时工具仍可工作，但不会进行机房位置维度的关联分析、故障阈值按默认值配置。

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
