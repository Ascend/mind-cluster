# 日志收集与数据源

本文档从用户视角介绍日志采集流程，说明 ascend-fd-tk 工具如何获取服务器、BMC、交换机链路故障诊断所需的日志数据。

## 采集模式概览

工具支持两种日志采集模式：

| 模式 | 适用场景 | 前置条件 |
|------|----------|----------|
| **在线 SSH 采集** | 设备可网络访问（IP / 账号密码 / 密钥 / 免密），实时采集 | 工具所在节点到目标设备 22 端口可达 |
| **离线日志收集** | 已获取到日志文件，仅需归档分析 | 提前收集日志到工具所在节点的目录 |

> **快速导航**：[服务器（Host）日志](#服务器host日志) | [BMC 日志](#BMC-log) | [交换机日志](#交换机日志)

<a id="服务器host日志"></a>

## 服务器（Host）日志

### 1. 在线采集

工具会自动通过 SSH 在服务器上执行以下命令进行采集：

| 类别 | 命令 | 用途 |
|------|------|------|
| 系统信息 | `hostname` | 主机名 |
| 系统信息 | `dmidecode -s system-serial-number` | 主机 SN |
| NPU 类型 | `lspci \| grep 'Device d80' --color=never` | NPU 型号识别 |
| NPU 映射 | `npu-smi info -m` | NPU / 芯片 ID 映射 |
| msnpureport 日志 | `msnpureport` | device 侧日志导出 |
| 光模块 | `hccn_tool -i {chip_phy_id} -optical -g` | 光模块功率 / SNR / CDR |
| 光模块 DFX 配置 | `hccn_tool -i {chip_phy_id} -optical -g dfx_cfg` | 光模块 DFX 配置 |
| 链路统计 | `hccn_tool -i {chip_phy_id} -link_stat -g` | 链路层统计（错包、丢包） |
| 链路状态 | `hccn_tool -i {chip_phy_id} -link -g` | 链路 up / down / 健康状态 |
| 网络健康 | `hccn_tool -i {chip_phy_id} -net_health -g` | 网络健康状态 |
| CDR SNR | `hccn_tool -i {chip_phy_id} -scdr -t 5` | CDR SNR 信息 |
| 性能 | `hccn_tool -i {chip_phy_id} -stat -g` | 性能计数器 |
| LLDP | `hccn_tool -i {chip_phy_id} -lldp -g` | LLDP 邻居 |
| HCCS | `npu-smi info -t hccs -i {npu_id} -c {chip_id}` | HCCS 协议信息 |
| SPOD | `npu-smi info -t spod-info -i {npu_id} -c {chip_id}` | SPOD 故障定位 |
| RoCE 速率 | `hccn_tool -i {chip_phy_id} -speed -g` | RoCE 速率 |
| RoCE 双工 | `hccn_tool -i {chip_phy_id} -duplex -g` | RoCE 双工模式 |

### 2. 离线日志采集

工具支持 3 个版本的离线日志结构，版本识别由工具自动完成，无需手动指定。通过以下任意一种脚本收集日志，收集后获得 `{file_name}.tar.gz`，直接将压缩包放入日志采集目录即可。

- 版本 1：通过 `tool_log_collection_out_version_all_<version>.sh` 收集日志。
- 版本 2：通过 `A3device日志一键采集脚本<version>.sh` 收集日志。
- 版本 3：通过 `link_down_collect_<version>.sh` 收集日志。

#### 版本 1

```text
host日志采集目录/
└── {file_name}.tar.gz/
    ├── 时间戳目录（如 2023-10-01_14-30-00）   # 使用 msnpureport 导出的 device 侧日志
    ├── hccn_tool.log                       # 网络配置工具日志
    ├── npu_card_info.log                   # NPU 卡信息
    ├── pcie_info.log                       # PCIe 信息
    └── version_info.log                    # 版本信息

# 注：工具支持压缩包自动解析，使用时无需手动解压。上述目录结构展示了压缩包内部层级，用于说明包内的重要文件。
```

#### 版本 2

```text
host日志采集目录/
└── {file_name}.tar.gz/
    ├── 时间戳目录（如 2023-10-01_14-30-00）    # 使用 msnpureport 导出的 device 侧日志
    ├── hccn_log/
    │   ├── net_conf.log
    │   ├── optical.log
    │   └── stat.log
    ├── npu_smi_log/
    │   └── npu_smi.log
    └── pcie_log/
        └── pcie.log

# 注：工具支持压缩包自动解析，使用时无需手动解压。上述目录结构展示了压缩包内部层级，用于说明包内的重要文件。
```

#### 版本 3

```text
host日志采集目录/
└── {file_name}.tar.gz/
    ├── 时间戳目录（如 2023-10-01_14-30-00）    # 使用 msnpureport 导出的 device 侧日志
    ├── lldp.log
    └── optical.log

# 注：工具支持压缩包自动解析，使用时无需手动解压。上述目录结构展示了压缩包内部层级，用于说明包内的重要文件。
```

---

<a id="BMC-log"></a>

## BMC 日志

### 1. 在线采集

工具通过 BMC IPMI 协议自动采集以下数据项：

| 数据项 | 采集命令 | 说明 |
|--------|----------|------|
| BMC 序列号 | `ipmcget -d serialnumber` | 设备唯一标识 |
| BMC 日期时间 | `ipmcget -d time` | 事件时间戳基准 |
| SEL 日志 | `ipmcget -d sel -v list` | 系统事件日志（核心诊断依据） |
| 传感器信息 | `ipmcget -t sensor -d list` | 温度 / 电压 / 风扇等传感器 |
| 健康事件 | `ipmcget -d healthevents` | 当前健康告警 |

此外，工具也提供内置命令 `collect_bmc_dump_info` 用于在线收集 BMC dump info 日志。该命令通过 `ipmcget -d diaginfo` 触发 BMC 一键收集，将 `dump_info.tar.gz` 下载到 `家目录/cache/bmc_dump_cache/`。下载的压缩包解压后包含上表所示数据项，可直接放入 BMC 日志采集目录用于离线诊断。

```bash
# 配置 BMC 信息（IP / 账号密码 / 密钥 / 免密）与 BMC 日志采集
ascend-fd-tk set_conn_config /home/user/conn.ini collect_bmc_dump_info
收集完成，请查看日志路径{...}
```

### 2. 离线日志采集

支持通过以下方式收集 BMC 日志：

- **方式 1**：通过 BMC 网页，使用"一键收集"按钮下载日志。
- **方式 2**：登录 BMC 平台，使用 `ipmcget -d diaginfo` 命令收集日志。

将以上方式采集的日志压缩包统一放到一个目录中，在清洗日志时会自动解压分析日志信息。

BMC 日志采集目录结构如下：

```text
bmc日志采集目录/
└── {file_name}.tar.gz/
    └── dump_info/
        └── AppDump/
            ├── bmc_network/network_info.txt           # 网络信息
            ├── frudata/fruinfo.txt                    # FRU 信息
            ├── event/sel.txt                          # 系统事件日志
            ├── event/current_event.txt                # 当前健康事件
            ├── sensor/sensor_info.txt                 # 传感器信息
            ├── network_adapter/optical_module/optical_module_history_info_log.csv  # 光模块历史1
            └── CpuMem/NpuIO/optical_module_history_info_log.csv                    # 光模块历史2

# 注：工具支持压缩包自动解析，使用时无需手动解压。上述目录结构展示了压缩包内部层级，用于说明包内的重要文件。
```

---
<a id="交换机日志"></a>

## 交换机日志

### 1. 在线采集

通过 SSH 登录交换机后，工具自动执行以下命令：

| 数据 | 命令 |
|------|------|
| 序列号 | `display license esn` |
| 接口摘要 | `dis int b \| no-more` |
| 交换机名称 | `dis cu \| in sysname` |
| 光模块 | `dis optical-module interface {interface} \| no-more` |
| 误码率 | `display interface troubleshooting \| no-more` |
| LLDP 邻居 | `dis lldp nei b \| n` |
| 活动告警 | `display alarm active \| no-more` |
| 历史告警 | `display alarm history \| no-more` |
| 接口信息 | `display interface \| no-more` |
| 当前时间 | `display clock \| include -` |
| HCCS 能力检测 | `display hccs eid ub-instance 0 \| no-more` |
| HCCS 代理响应统计 | `display hccs proxy response statistics \| no-more` |
| HCCS 代理响应详情 | `display hccs proxy response detail \| no-more` |
| HCCS 路由缺失 | `display hccs route miss statistics \| no-more` |
| HCCS 映射表 | `display hccs decode and map table \| in MAP_TABLE \| no-more` |
| 端口链路状态 | `display for info enp s 1 c {chip_id} "get port link start 0 end 47" \| no-more` |
| 端口统计 | `display for info enp s 1 c {chip_id} "get port statistic count port {port_id} module {module} type 0 path 2" \| no-more` |
| HCCS 端口无效丢弃 | `display hccs port-invalid drop statistics \| no-more` |
| 端口信用背压统计 | `display qos port-credit back-pressure statistics \| no-more` |
| 端口 SNR | `display for info enp s 1 c {chip_id} "get port snr port-id {port_id}"` |
| 接口 hilink SNR | `dis int hilink snr \| n` |
| 收发器信息 | `display interface transceiver verbose \| no-more` |
| 接口通道信息 | `display interface information \| no-more` |
| Serdes 转储信息 | `display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}" \| no-more` |

### 2. 离线日志采集

交换机离线日志主要需要以下两种日志：

- **CLI 命令输出日志（diag 文本日志）**：包含各种交换机命令的执行结果。收集方式如下：
  - 方式 1：登录交换机后执行 `display diagnostic-information {filename}.txt`。
  - 方式 2：登录交换机后手动执行关键命令（必须包含 `display current-configuration`），将回显的文本保存到 `.txt` 文件并导出。
- **诊断日志**：由交换机诊断工具生成的结构化日志（`diagnostic_information.zip`）。登录交换机后执行 `collect diagnostic information`，并导出 zip 包。

Switch 日志目录结构如下：

```text
switch日志采集目录/
└── {file_name}.zip/
    ├── *.txt                                   # CLI 命令输出日志（任意 .txt 文件名）
    └── diagnostic_information.zip/
        ├── slot_1.zip/
        │   └── tempdir/
        │       └── port_down_status.log        # 端口 down 状态日志
        └── logfile_slot_1.zip/
            └── tempdir/
                └── diag.log.zip/
                    └── diag.log                # 诊断日志（含交换机名称、SNR 等信息）

# 注：工具支持压缩包自动解析（含嵌套压缩包），使用时无需手动解压。上述目录结构展示了压缩包内部层级，用于说明包内的重要文件。
```
