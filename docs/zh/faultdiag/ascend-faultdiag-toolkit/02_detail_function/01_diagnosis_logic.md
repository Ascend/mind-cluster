# 诊断逻辑

## 诊断逻辑概述

链路诊断工具基于多源数据采集和分析，实现对集群设备的故障诊断。支持在线SSH采集和离线日志解析两种模式。

## 主机相关诊断片段

### 主机光模块综合诊断

**输入**

- 主机SSH在线采集：
  - 光模块信息：`hccn_tool -i {chip_phy_id} -optical -g`
  - 链路状态：`hccn_tool -i {chip_phy_id} -link_stat -g`
- 主机离线日志（版本1）：
  - `hccn_tool.log` (Optical信息)
  - `npu_card_info.log` (NPU信息)
- 主机离线日志（版本2）：
  - `hccn_log/optical.log` (光模块信息)
  - `npu_smi_log/npu_smi.log` (NPU信息)
- 主机离线日志（版本3）：
  - `optical.log` (光模块信息)

**诊断逻辑**

检查光模块状态、功率、SNR、CDR参数、uncorr_cw_cnt和IIC故障等多个维度，结合预定义阈值判断异常。

**异常输出**

- 光模块功率异常：当TX/RX功率值超出阈值范围时

  示例："光模块功率异常，TX功率：-5dBm（阈值范围：-3~0dBm）"
- SNR异常：当SNR值低于阈值（如12dB）时

  示例："光模块SNR异常，当前值：10.5dB（阈值：12dB）"
- CDR失锁：当CDR状态为"Unlock"时

  示例："光模块CDR失锁，状态：Unlock"
- uncorr_cw_cnt超阈值：当uncorr_cw_cnt值大于预定义阈值时

  示例："光模块uncorr_cw_cnt超阈值，当前值：1000（阈值：500）"
- IIC通信故障：当IIC通信状态为"Failed"时

  示例："光模块IIC通信故障，状态：Failed"

### 主机环回诊断

**输入**

- 主机SSH在线采集：`hccn_tool -i {npu_id} -optical -t {model}` (环回测试结果)
- 主机离线日志（版本1）：`hccn_tool.log` (Optical信息)
- 主机离线日志（版本2）：`hccn_log/optical.log` (光模块信息)
- 主机离线日志（版本3）：`optical.log` (光模块信息)

**诊断逻辑**

检查loopback测试的状态码，识别环回失败情况。

**异常输出**

环回测试失败：当环回测试状态码为非0值时

示例："环回测试失败，状态码：1，指示主机内部光链路或模块故障"

### 主机光模块Los/LoL诊断

**输入**

- 主机SSH在线采集：`hccn_tool -i {chip_phy_id} -optical -g` (tx_los、rx_los、rx_lol状态)
- 主机离线日志（版本1）：`hccn_tool.log` (Optical信息)
- 主机离线日志（版本2）：`hccn_log/optical.log` (光模块信息)
- 主机离线日志（版本3）：`optical.log` (光模块信息)

**诊断逻辑**

解析光模块状态字段，判断是否存在光信号丢失或激光关断。

**异常输出**

- TX Los告警：当tx_los字段值为"1"时

  示例："光模块TX Los告警，tx_los状态：1"
- RX Los告警：当rx_los字段值为"1"时

  示例："光模块RX Los告警，rx_los状态：1"
- RX LoL告警：当rx_lol字段值为"1"时

  示例："光模块RX LoL告警，rx_lol状态：1"

### 主机NPU端口状态诊断

**输入**

- 主机SSH在线采集：
  - 链路状态：`hccn_tool -i {chip_phy_id} -link -g`
  - 网络健康状态：`hccn_tool -i {chip_phy_id} -net_health -g`
- 主机离线日志（版本1）：`hccn_tool.log` (link status信息)
- 主机离线日志（版本2）：`hccn_log/optical.log` (网络健康信息)

**诊断逻辑**

检查NPU端口的光模块存在状态、网络健康状态和连接状态。

**异常输出**

- 端口故障：当端口状态字段值为"Fault"时

  示例："NPU端口故障，状态：Fault"
- 网络异常：当网络健康状态字段值为"Abnormal"时

  示例："NPU端口网络异常，健康状态：Abnormal"
- 连接断开：当连接状态字段值为"Disconnected"时

  示例："NPU端口连接断开，状态：Disconnected"

### 主机间光链路诊断

**输入**

- 主机SSH在线采集：`hccn_tool -i {chip_phy_id} -optical -g` (光模块参数)
- 主机离线日志（版本1）：`hccn_tool.log` (Optical信息)
- 主机离线日志（版本2）：`hccn_log/optical.log` (光模块信息)
- 主机离线日志（版本3）：`optical.log` (光模块信息)

**诊断逻辑**

对主机间连接的光模块功率、SNR和电流等参数进行多维度检查。

**异常输出**

- 主机间光链路功率异常：当主机间连接的光模块TX/RX功率值超出阈值范围时

  示例："主机间光链路功率异常，TX功率：-6dBm（阈值范围：-3~0dBm）"
- SNR异常：当主机间连接的光模块SNR值低于阈值时

  示例："主机间光链路SNR异常，当前值：11.2dB（阈值：12dB）"
- 电流异常：当主机间连接的光模块电流值超出阈值范围时

  示例："主机间光链路电流异常，当前值：15.5mA（阈值范围：5~15mA）"

### RoCE端口配置诊断

**输入**

- 主机SSH在线采集：
  - 端口速度：`hccn_tool -i {chip_phy_id} -speed -g`
  - 双工模式：`hccn_tool -i {chip_phy_id} -duplex -g`
- 主机离线日志（版本1）：`hccn_tool.log` (speed信息)
- 主机离线日志（版本2）：`hccn_log/net_conf.log` (端口速度信息)
- 主机离线日志（版本3）：`optical.log` (speed info)

**诊断逻辑**

对比NPU端口与对端交换机端口的速率和双工模式配置。

**异常输出**

- 速率不匹配：当NPU端口速率与对端交换机端口速率不一致时

  示例："RoCE端口速率不匹配，NPU端口：100Gbps，交换机端口：50Gbps"
- 双工模式不匹配：当NPU端口双工模式与对端交换机端口双工模式不一致时

  示例："RoCE端口双工模式不匹配，NPU端口：Full，交换机端口：Half"

## BMC相关诊断片段

### BMC错误码分析

**输入**

- BMC SSH在线采集：`ipmcget -d sel -v list` (SEL日志)
- BMC离线日志：
  - `AppDump/event/sel.txt` (系统事件日志)
  - `AppDump/event/current_event.txt` (当前健康事件)

**诊断逻辑**

解析BMC日志中的事件代码，识别硬件异常。

**异常输出**

- 多Bit ECC故障：当事件代码包含"0x80e01801"时

  示例："BMC硬件异常，事件代码：0x80e01801，发生多Bit ECC故障"
- 多Bit ECC故障，隔离行已满64：当事件代码包含"0x80e18402"时

  示例："BMC硬件异常，事件代码：0x80e18402，多Bit ECC故障，隔离行已满64"
- AIV算子超时，NPU热复位：当事件代码包含"0x80cb800a"时

  示例："BMC硬件异常，事件代码：0x80cb800a，AIV算子超时，NPU热复位"
- AIV总线访问错误：当事件代码包含"0x80cb8009"时

  示例："BMC硬件异常，事件代码：0x80cb8009，AIV总线访问错误"

### BMC光模块诊断

**输入**

- BMC SSH在线采集：
  - 传感器信息：`ipmcget -t sensor -d list`
  - 光模块历史日志
- BMC离线日志：
  - `AppDump/network_adapter/optical_module/optical_module_history_info_log.csv` (光模块历史日志1)
  - `AppDump/CpuMem/NpuIO/optical_module_history_info_log.csv` (光模块历史日志2)

**诊断逻辑**

检查光功率、偏置电流、SNR等参数是否超出阈值，以及Los状态是否异常。

**异常输出**

- 光功率异常：当光功率值超出阈值范围时

  示例："BMC光模块功率异常，RX功率：-25dBm（阈值范围：-20~-10dBm）"
- 偏置电流异常：当偏置电流值超出阈值范围时

  示例："BMC光模块偏置电流异常，当前值：16.2mA（阈值范围：5~15mA）"
- SNR异常：当SNR值低于阈值时

  示例："BMC光模块SNR异常，当前值：10.8dB（阈值：12dB）"
- 链路Los告警：当Los状态字段值为"1"时

  示例："BMC光模块链路Los告警，Los状态：1"

### HCCS链路降级诊断

**输入**

- BMC SSH在线采集：`ipmcget -d healthevents` (健康事件日志)
- BMC离线日志：`AppDump/event/current_event.txt` (当前健康事件)

**诊断逻辑**

解析特定错误代码(0x28000049)的事件描述，定位故障端口。

**异常输出**

HCCS链路降级：当事件代码包含"0x28000049"时

示例："HCCS链路降级，事件代码：0x28000049，指示故障的L1交换机端口或CPU板抽屉"

## 交换机相关诊断片段

### 交换机光模块诊断

**输入**

- 交换机SSH在线采集：`dis optical-module interface {interface}` (光模块信息)
- 交换机离线日志：CLI命令输出中的光模块信息表格（包含`Items, Value, HighAlarm, HighWarn, LowAlarm, Status`字段）

**诊断逻辑**

分析交换机端口间的光模块功率、SNR和电流等参数，支持单端和双端诊断。

**异常输出**

- 光模块功率异常：当TX/RX功率值超出阈值范围时

  示例："交换机光模块功率异常，TX功率：-4dBm（阈值范围：-3~0dBm）"
- SNR异常：当SNR值低于阈值时

  示例："交换机光模块SNR异常，当前值：11.5dB（阈值：12dB）"
- 电流异常：当电流值超出阈值范围时

  示例："交换机光模块电流异常，当前值：15.8mA（阈值范围：5~15mA）"

### 交换机端口误码率诊断

**输入**

- 交换机SSH在线采集：`display interface troubleshooting {interface}` (误码率信息)
- 交换机离线日志：CLI命令输出中的误码率信息（包含`Current state, Speed`字段的内容块）

**诊断逻辑**

检查端口的误码率是否超过阈值。

**异常输出**

误码率超阈值：当误码率值大于预定义阈值（如1e-12）时

示例："交换机端口误码率超阈值，当前值：5e-12（阈值：1e-12），指示链路质量问题"

### 交换机CRC错误告警诊断

**输入**

- 交换机SSH在线采集：`display alarm active` (活动告警)
- 交换机离线日志：活动告警信息表格（包含`Sequence, AlarmId, Severity, Date Time, Description`字段）

**诊断逻辑**

解析特定错误代码(0x081300bc)的告警信息，识别CRC错误快速增长的端口。

**异常输出**

端口CRC错误快速增长：当告警信息包含错误代码"0x081300bc"时

示例："交换机端口CRC错误快速增长，告警ID：0x081300bc，端口：GigabitEthernet0/0/1，指示链路质量或硬件故障"

### 交换机端口降lane诊断

**输入**

- 交换机SSH在线采集：`display alarm active` (活动告警)
- 交换机离线日志：活动告警信息表格（包含`Sequence, AlarmId, Severity, Date Time, Description`字段）

**诊断逻辑**

解析特定错误代码(0xf10509)的告警信息，识别发生降lane的端口。

**异常输出**

端口降lane告警：当告警信息包含错误代码"0xf10509"时

示例："交换机端口降lane告警，告警ID：0xf10509，端口：GigabitEthernet0/0/2，指示链路或硬件故障"

### 交换机光模块Los告警诊断

**输入**

- 交换机SSH在线采集：`display alarm active` (活动告警)
- 交换机离线日志：活动告警信息表格（包含`Sequence, AlarmId, Severity, Date Time, Description`字段）

**诊断逻辑**

解析特定错误代码(0x8130059)的告警信息，识别光模块链路Los告警的端口。

**异常输出**

光模块链路Los告警：当告警信息包含错误代码"0x8130059"时

示例："交换机光模块链路Los告警，告警ID：0x8130059，端口：GigabitEthernet0/0/3，指示光信号丢失"

### 交换机光模块状态诊断

**输入**

- 交换机SSH在线采集：`display interface transceiver verbose` (收发器详细信息)
- 交换机离线日志：CLI命令输出中的收发器信息（包含`transceiver information:`标记的内容块）

**诊断逻辑**：

检查光模块State Flag、Datapath State和Module State等字段的状态值。

**异常输出**

光模块状态异常：当State Flag、Datapath State或Module State字段值为异常状态时

示例："交换机光模块状态异常，State Flag：0x00000001，Datapath State：Fault，指示收发光指标、通道状态或功率模式问题"

## 通用诊断片段

### 端口lane功率差异诊断

**输入**

- 主机SSH在线采集：`hccn_tool -i {chip_phy_id} -optical -g` (光模块lane功率信息)
- 交换机SSH在线采集：`dis optical-module interface {interface}` (光模块lane功率信息)
- BMC SSH在线采集：传感器信息（`ipmcget -t sensor -d list`）
- 离线日志：光模块lane功率数据（来自主机、交换机或BMC的日志文件）

**诊断逻辑**

计算同一端口不同lane间的功率最大值和最小值差值，判断是否超过阈值（3db）。

**异常输出**

lane间功率差异过大：当同一端口不同lane间的功率最大值和最小值差值超过阈值（3db）时

示例："端口lane功率差异过大，端口：eth0，最大差值：4.2db（阈值：3db），指示端口内部lane故障"

## HCCS相关诊断片段

### HCCS RP TX超时诊断

**输入**

- 交换机SSH在线采集：
  - `display hccs proxy response statistics` (HCCS代理响应统计)
  - `display hccs proxy response detail interface {interface}` (HCCS代理响应详情)
- 交换机离线日志：CLI命令输出中的HCCS相关统计信息（包含HCCS关键字的表格）

**诊断逻辑**

分析RP TX超时的接口和地址映射关系，检查端口状态和对端接口状态。

**异常输出**

端口长期down、闪断、窝包等导致的RP TX超时：当HCCS代理响应统计中RP TX超时次数大于0时

示例："HCCS RP TX超时，接口：eth0，超时次数：10，可能原因：端口长期down、闪断或窝包"

### HCCS RX超时诊断

**输入**

- 交换机SSH在线采集：
  - `display hccs proxy response statistics` (HCCS代理响应统计)
  - `display interface` (接口状态)
  - `display for info enp s 1 c {chip_id} "get port link start 0 end 47"` (端口链路状态)
- 交换机离线日志：
  - CLI命令输出中的HCCS相关统计信息（包含HCCS关键字的表格）
  - CLI命令输出中的接口状态信息（包含`current state, Description, Port Mode`字段）

**诊断逻辑**

筛选发生RX超时的接口，结合链路状态、降lane情况等判断故障原因。

**异常输出**

端口长期down、闪断、链路降lane或XPU设备异常导致的RX超时：当HCCS代理响应统计中RX超时次数大于0时

示例："HCCS RX超时，接口：eth1，超时次数：5，可能原因：端口闪断或链路降lane"

### HCCS Serdes诊断

**输入**

- 交换机SSH在线采集：`display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}"` (Serdes转储信息)
- 交换机离线日志：CLI命令输出中的Serdes转储信息

**诊断逻辑**

检查CDR失锁状态和电源故障代码，识别Serdes异常。

**异常输出**

- CDR失锁：当CDR状态为"Unlock"时

  示例："HCCS Serdes CDR失锁，端口：eth0，lane：1"
- 电源故障：当电源状态字段值为"Fault"时

  示例："HCCS Serdes电源故障，端口：eth0，lane：2"

### HCCS端口SNR诊断

**输入**

- 交换机SSH在线采集：`display interface hilink snr` (HCCS端口信噪比)
- 交换机离线日志：CLI命令输出中的HCCS端口SNR信息（包含`display interface hilink snr`输出的表格）

**诊断逻辑**

对比端口SNR值与阈值，识别SNR低于阈值的异常端口。

**异常输出**

SNR低于阈值：当SNR值低于阈值（如12dB）时

示例："HCCS端口SNR异常，接口：eth2，当前SNR：10.5dB（阈值：12dB），指示链路质量问题"
