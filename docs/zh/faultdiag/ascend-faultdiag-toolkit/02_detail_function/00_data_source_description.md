# 数据源详细描述

## 概述

本文档详细描述了链路诊断工具支持的各类数据源，包括：

- 数据类型及具体数据项
- 数据采集方式（SSH在线采集/离线日志解析）
- 数据来源位置（具体命令/文件路径）

链路诊断工具支持以下三种设备类型的数据源采集：

- 主机（Host）
- BMC（Baseboard Management Controller）
- 交换机（Switch）

每种设备支持以下两种采集方式：

- **SSH在线采集**：通过SSH连接设备，执行命令获取实时数据。
- **离线日志解析**：解析已导出的日志文件，提取历史数据。

## 主机（Host）数据源

### SSH在线采集

| 数据项 | 采集命令                                                 | 说明 |
|--------|------------------------------------------------------|------|
| 主机名 | `hostname`                                           | 获取主机名称 |
| NPU映射信息 | `npu-smi info -m`                                    | 获取NPU芯片映射关系（NPU ID、芯片ID、物理ID） |
| 光模块信息 | `hccn_tool -i {chip_phy_id} -optical -g`             | 获取指定物理芯片的光模块信息 |
| 链路状态 | `hccn_tool -i {chip_phy_id} -link_stat -g`           | 获取指定物理芯片的链路状态 |
| 统计信息 | `hccn_tool -i {chip_phy_id} -stat -g`                | 获取指定物理芯片的统计信息 |
| LLDP信息 | `hccn_tool -i {chip_phy_id} -lldp -g`                | 获取指定物理芯片的LLDP邻居信息 |
| NPU类型 | `lspci \| grep 'Device d80'`                         | 获取NPU设备类型 |
| 系统序列号 | `dmidecode -s system-serial-number`                  | 获取系统序列号 |
| HCCS信息 | `npu-smi info -t hccs -i {npu_id} -c {chip_id}`      | 获取指定NPU和芯片的HCCS信息 |
| SPOD信息 | `npu-smi info -t spod-info -i {npu_id} -c {chip_id}` | 获取指定NPU和芯片的SPOD信息 |
| MSNPUREPORT日志 | `msnpureport`                                        | 生成并解析MSNPUREPORT日志 |
| RoCE速度 | `hccn_tool -i {chip_phy_id} -speed -g`               | 获取指定物理芯片的RoCE速度 |
| RoCE双工模式 | `hccn_tool -i {chip_phy_id} -duplex -g`              | 获取指定物理芯片的RoCE双工模式 |
| 网络健康状态 | `hccn_tool -i {chip_phy_id} -net_health -g`          | 获取指定物理芯片的网络健康状态 |
| 链路状态 | `hccn_tool -i {chip_phy_id} -link -g`         | 获取指定物理芯片的链路状态 |
| CDR信息 | `hccn_tool -i {chip_phy_id} -scdr -t 5`              | 获取指定物理芯片的CDR信息 |
| DFX配置 | `hccn_tool -i {chip_phy_id} -optical -g dfx_cfg`     | 获取指定物理芯片的DFX配置 |
| 光模块环回测试 | `hccn_tool -i {npu_id} -optical -t {model}`          | 执行光模块环回测试 |

### 离线日志解析

主机离线日志支持三种不同版本的配置集合，每种集合对应不同的文件路径和解析关键词。

#### 版本1配置（ParseConfigCollectionV1）

**文件结构**

```text
日志目录/
├── hccn_tool.log        # 网络配置工具日志
├── npu_card_info.log    # NPU卡信息日志
├── pcie_info.log        # PCIe信息日志
└── version_info.log     # 版本信息日志
```

**数据项与解析配置**

| 数据类型 | 解析关键词 | 日志文件路径 | 说明 |
|----------|------------|--------------|------|
| LLDP | "lldp" | hccn_tool.log | 链路层发现协议信息 |
| SPEED | "speed -g" | hccn_tool.log | 端口速度信息 |
| OPTICAL | "optical" | hccn_tool.log | 光模块信息 |
| LINK_STAT | "link stat" | hccn_tool.log | 链路统计信息 |
| STAT | "stat" | hccn_tool.log | 性能统计信息 |
| HCCN_LINK_STATUS | "link" | hccn_tool.log | HCCN链路状态 |
| CDR_SNR | "cdr5 snr 1 times" | hccn_tool.log | CDR信噪比信息 |
| SPOD_INFO | "Collect spod-info info for all NPUs" | npu_card_info.log | SPOD信息 |
| HCCS | "Collect hccs info for all NPUs" | npu_card_info.log | HCCS协议信息 |
| NPU_TYPE | "lspci" | pcie_info.log | NPU类型信息 |
| SN | "timeout 30s dmidecode -t1" | version_info.log | 序列号信息 |

#### 版本2配置（ParseConfigCollectionV2）

**文件结构**

```text
日志目录/
├── hccn_log/
│   ├── net_conf.log     # 网络配置日志
│   ├── optical.log      # 光模块日志
│   └── stat.log         # 统计信息日志
├── npu_smi_log/
│   └── npu_smi.log      # NPU SMI日志
└── pcie_log/
    └── pcie.log         # PCIe日志
```

**数据项与解析配置**

| 数据类型 | 解析关键词 | 日志文件路径 | 说明 |
|----------|------------|--------------|------|
| LLDP | "lldp" | hccn_log/net_conf.log | 链路层发现协议信息 |
| SPEED | "speed" | hccn_log/net_conf.log | 端口速度信息 |
| OPTICAL | "optical" | hccn_log/optical.log | 光模块信息 |
| LINK_STAT | "link stat" | hccn_log/optical.log | 链路统计信息 |
| NET_HEALTH | "health info" | hccn_log/optical.log | 网络健康信息 |
| HCCN_LINK_STATUS | "link info" | hccn_log/optical.log | HCCN链路状态 |
| STAT | "stat" | hccn_log/stat.log | 性能统计信息 |
| SPOD_INFO | "spod_info" | npu_smi_log/npu_smi.log | SPOD信息 |
| HCCS | "hccs" | npu_smi_log/npu_smi.log | HCCS协议信息 |
| NPU_TYPE | "pcie" | pcie_log/pcie.log | NPU类型信息 |

#### 版本3配置（ParseConfigCollectionV3）

**文件结构**

```text
日志目录/
├── lldp.log        # LLDP日志
└── optical.log     # 光模块日志
```

**数据项与解析配置**

| 数据类型 | 解析关键词 | 日志文件路径 | 说明 |
|----------|------------|--------------|------|
| LLDP | "lldp" | lldp.log | 链路层发现协议信息 |
| SPEED | "speed info" | optical.log | 端口速度信息 |
| OPTICAL | "optical" | optical.log | 光模块信息 |
| LINK_STAT | "link stat" | optical.log | 链路统计信息 |
| HCCN_LINK_STATUS | "link info" | optical.log | HCCN链路状态 |

### MSNPUREPORT日志

除了上述配置之外，还支持解析MSNPUREPORT工具生成的日志：

- 路径：日志目录下的时间戳子目录（如`2023-10-01_14-30-00`）
- 内容：包含详细的NPU报告信息

## BMC数据源

### SSH在线采集

**数据项与来源**

| 数据项 | 采集命令 | 说明 |
|--------|----------|------|
| BMC序列号 | `ipmcget -d serialnumber` | 获取BMC序列号 |
| BMC日期时间 | `ipmcget -d time` | 获取BMC当前时间 |
| SEL日志 | `ipmcget -d sel -v list` | 获取系统事件日志 |
| 传感器信息 | `ipmcget -t sensor -d list` | 获取传感器列表及状态 |
| 健康事件 | `ipmcget -d healthevents` | 获取健康事件日志 |
| 诊断信息 | `ipmcget -d diaginfo` | 获取BMC诊断信息并下载为压缩包 |
| 光模块历史日志 | - | 获取光模块历史记录|

**诊断信息压缩包**

- 远程路径：`/tmp/dump_info.tar.gz`
- 本地存储路径：`CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR`
- 命名格式：`{host}_{sn_num}_{subfix_date_time}.tar.gz`

### 离线日志解析

**日志来源**

BMC离线日志通常通过`ipmcget -d diaginfo`命令生成的诊断信息压缩包（`dump_info.tar.gz`）获取，解压后包含以下目录结构：

```text
dump_info/
└── AppDump/
    ├── bmc_network/
    │   └── network_info.txt      # 网络信息文件
    ├── frudata/
    │   └── fruinfo.txt           # FRU信息文件
    ├── event/
    │   ├── sel.txt               # 系统事件日志
    │   └── current_event.txt     # 当前健康事件
    ├── sensor/
    │   └── sensor_info.txt       # 传感器信息
    ├── network_adapter/
    │   └── optical_module/
    │       └── optical_module_history_info_log.csv  # 光模块历史日志1
    └── CpuMem/
        └── NpuIO/
            └── optical_module_history_info_log.csv  # 光模块历史日志2
```

**数据项与来源**

| 数据项 | 日志文件路径 | 解析方式 |
|--------|--------------|----------|
| BMC IP地址 | `AppDump/bmc_network/network_info.txt` | 从文件中提取"IP Address"字段 |
| 序列号 | `AppDump/frudata/fruinfo.txt` | 从FRU信息中提取"System Serial Number" |
| SEL信息 | `AppDump/event/sel.txt` | 直接读取SEL日志内容 |
| 传感器信息 | `AppDump/sensor/sensor_info.txt` | 直接读取传感器信息内容 |
| 健康事件 | `AppDump/event/current_event.txt` | 直接读取健康事件内容 |
| 光模块历史日志 | <ul><li>`AppDump/network_adapter/optical_module/optical_module_history_info_log.csv`</li><li>`AppDump/CpuMem/NpuIO/optical_module_history_info_log.csv`</li></ul> | 解析CSV格式的光模块历史记录，提取链路中断记录 |

## 交换机（Switch）数据源

### SSH在线采集

| 数据项 | 采集命令 | 说明 |
|--------|----------|------|
| 交换机序列号 | `display license esn` | 获取交换机ESN序列号 |
| 接口摘要 | `dis int b \| no-more` | 获取所有接口的基本状态 |
| 交换机名称 | `dis cu \| in sysname` | 获取交换机系统名称 |
| 光模块信息 | `dis optical-module interface {interface}` | 获取指定接口的光模块信息 |
| 误码率 | `display interface troubleshooting {interface}` | 获取指定接口的故障排除信息 |
| LLDP邻居 | `dis lldp nei b \| n` | 获取LLDP邻居摘要 |
| 活动告警 | `display alarm active \| no-more` | 获取当前活动告警 |
| 历史告警 | `display alarm history \| no-more` | 获取历史告警记录 |
| 接口信息 | `display interface \| no-more` | 获取所有接口的详细信息 |
| 当前时间 | `display clock \| include -` | 获取交换机当前时间 |
| HCCS代理响应统计 | `display hccs proxy response statistics \| no-more` | 获取HCCS代理响应统计 |
| HCCS代理响应详情 | `display hccs proxy response detail interface {interface}` | 获取指定接口的HCCS代理响应详情 |
| HCCS路由缺失 | `display hccs route miss statistics \| no-more` | 获取HCCS路由缺失统计 |
| 端口链路状态 | `display for info enp s 1 c {chip_id} "get port link start 0 end 47" \| no-more` | 获取芯片端口链路状态 |
| 端口统计 | `display for info enp s 1 c {chip_id} "get port statistic count port {port_id} module {module} type 0 path 2" \| no-more` | 获取端口统计信息 |
| HCCS端口无效丢弃 | `display hccs port-invalid drop statistics \| no-more` | 获取HCCS端口无效丢弃统计 |
| 端口信用背压统计 | `display qos port-credit back-pressure statistics \| no-more` | 获取端口信用背压统计 |
| HCCS端口SNR | `display interface hilink snr \| n` | 获取HCCS端口信噪比 |
| 收发器信息 | `display interface transceiver verbose \| no-more` | 获取接口收发器详细信息 |
| 接口通道信息 | `display interface information \| no-more` | 获取接口通道信息 |
| Serdes转储信息 | `display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}" \| no-more` | 获取Serdes转储信息 |

### 离线日志解析

交换机离线日志主要有以下两种形式：

- **CLI命令输出日志**：包含各种交换机命令的执行结果。
- **诊断信息输出**：由交换机诊断工具生成的结构化日志。

**日志文件位置**

- CLI输出日志：通常为单个文件（如`switch_cli_output.txt`）或按命令分类的多个文件。
- 诊断信息输出：通常为带有诊断标记的文件（如`diag_info.txt`）。

**数据类型与来源**

| 数据类型 | 日志来源特征 | 说明 |
|----------|--------------|------|
| 活动告警详情 | 包含`AlarmId, AlarmName, AlarmType, State : active`字段的内容块 | 从CLI输出日志中提取活动告警的详细信息 |
| 历史告警详情 | 包含`AlarmId, AlarmName, AlarmType, State : cleared`字段的内容块 | 从CLI输出日志中提取已清除告警的详细信息 |
| 活动告警 | 包含`Sequence, AlarmId, Severity, Date Time, Description`字段的表格 | 从CLI输出日志中提取活动告警摘要 |
| 历史告警 | 包含`Sequence, AlarmId, Severity, Date Time, Description`字段的表格且带有历史标记 | 从CLI输出日志中提取历史告警记录 |
| LLDP邻居 | 包含`Local Interface, Exptime(s), Neighbor Interface, Neighbor Device`字段的表格 | 从CLI输出日志中提取LLDP邻居信息 |
| 光模块信息 | 包含`Items, Value, HighAlarm, HighWarn, LowAlarm, Status`字段的表格 | 从CLI输出日志中提取光模块监控数据 |
| 接口摘要 | 包含`Interface, PHY, Protocol, InUti, OutUti, inErrors, outErrors`字段的表格 | 从CLI输出日志中提取接口基本状态 |
| 误码率 | 包含`Current state, Speed`字段的内容块 | 从CLI输出日志中提取接口误码率信息 |
| 接口信息 | 包含`current state, Description, Port Mode`字段的内容块 | 从CLI输出日志中提取接口详细配置 |
| 许可证ESN | 包含`MainBoard, ESN`字段的内容块 | 从CLI输出日志中提取交换机序列号 |
| 系统时钟 | 包含`clock, Time Zone`字段的内容块 | 从CLI输出日志中提取交换机当前时间 |
| 收发器信息 | 包含`transceiver information:`标记的内容块 | 从CLI输出日志中提取收发器详细信息 |
| HCCS相关信息 | 包含`HCCS`关键字的各种表格和内容块 | 从CLI输出日志中提取HCCS协议相关的各种统计信息 |
