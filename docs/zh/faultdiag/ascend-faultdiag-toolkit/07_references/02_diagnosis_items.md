# 诊断项参考

诊断项按设备类型分类，每个诊断项均明确了其在线命令来源和离线日志路径。同一设备可能涉及多个诊断项，每个诊断项对应一个或多个在线命令与离线日志。工具共内置 **40 项** 诊断项，覆盖主机、BMC、交换机、HCCS 及通用维度。

## 诊断项索引

| 设备类型 | 诊断项数量 | 详细说明 |
|----------|------------|----------|
| 主机侧相关诊断 | 18 | [查看](#主机侧相关诊断) |
| BMC 相关诊断 | 2 | [查看](#bmc-相关诊断) |
| 交换机相关诊断 | 11 | [查看](#交换机相关诊断) |
| HCCS 相关诊断 | 8 | [查看](#hccs-相关诊断) |
| 通用诊断 | 1 | [查看](#通用诊断) |

## 主机侧相关诊断

<table>
<thead>
<tr>
<th>诊断项</th>
<th>在线命令</th>
<th>离线日志路径</th>
<th>诊断逻辑</th>
<th>异常输出示例</th>
</tr>
</thead>
<tbody>
<tr>
<td>光模块通信链路故障检测</td>
<td rowspan="7"><code>hccn_tool -i {chip_phy_id} -optical -g</code></td>
<td rowspan="7">V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code></td>
<td>检查光模块控制链路是否可达</td>
<td>光模块Control link unreachable</td>
</tr>
<tr>
<td>光模块在位检测</td>
<td>检查光模块是否在位</td>
<td>光模块未在位，状态：NA</td>
</tr>
<tr>
<td>高功耗使能寄存器状态检测</td>
<td>检查高功率使能寄存器状态</td>
<td>光模块处于低功率模式，high power enable reg:0x00</td>
</tr>
<tr>
<td>单端光模块光功率检测</td>
<td>检查 TX/RX 功率值是否超出阈值范围</td>
<td>光模块光功率异常，RX功率-18.5dBm低于阈值-15dBm</td>
</tr>
<tr>
<td>单端光模块SNR检测</td>
<td>检查 Host SNR / Media SNR 值是否低于阈值</td>
<td>光模块SNR异常：lane0: Host SNR值7.2dB低于阈值8.0dB</td>
</tr>
<tr>
<td>光模块SNR LANE间差值</td>
<td>检查不同 lane 间的 SNR 差值是否超过阈值</td>
<td>光模块SNR LANE间差值异常：LANE0与LANE3差值为4.2dB</td>
</tr>
<tr>
<td>光模块Los/LoL检测</td>
<td>检查 Rx Los、Tx Los、Rx LoL、Tx LoL 状态值是否大于 0</td>
<td>光模块Rx Los指标异常，状态：1</td>
</tr>
<tr>
<td>光模块uncorr_cw_cnt检测</td>
<td rowspan="2"><code>msnpureport</code> 自动采集</td>
<td rowspan="2">V1-V3: <code>msnpureport</code> 报告目录下的时间戳子目录</td>
<td>检查连续 3 次 uncorr_cw_cnt 是否大于 10</td>
<td>持续连续3次出现uncorr_cw_cnt > 10，发生时间：2025-10-01-14:30:00.123456，2025-10-01-14:30:01.234567，2025-10-01-14:30:02.345678</td>
</tr>
<tr>
<td>光模块IIC 通信故障检测</td>
<td>检测 IIC 通信异常事件</td>
<td>检测到IIC异常：trans status[0x40]，error status[0x10]，NPU板载光模块转接器可能存在故障</td>
</tr>
<tr>
<td>光模块初始化开光状态检测</td>
<td><code>hccn_tool -i {chip_phy_id} -dfx_cfg -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: 暂不支持</td>
<td>检查 TX Disable 状态是否为禁用</td>
<td>光模块处于关光状态，tx disable status：1</td>
</tr>
<tr>
<td>光模块CDR SNR检测</td>
<td><code>hccn_tool -i {chip_phy_id} -cdr_snr -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: 暂不支持</td>
<td>检查 CDR 的 Host SNR / Media SNR 值是否低于阈值</td>
<td>CDR SNR异常，Host SNR值为6.8dB低于阈值8.0dB</td>
</tr>
<tr>
<td>光模块端口状态、网络健康状态、连接状态检测</td>
<td><code>hccn_tool -i {chip_phy_id} -link -g</code>、<code>hccn_tool -i {chip_phy_id} -net_health -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code>（net_health不支持）</td>
<td>检查 NPU 端口的网络健康状态与连接状态是否偏离正常阈值，并附带对端交换机与端口信息</td>
<td>端口光模块状态异常，网络健康状态：abnormal，连接状态：down。 对端交换机：SWITCH-01，对端端口：10GE1/0/1。</td>
</tr>
<tr>
<td>RoCE端口配置检测</td>
<td><code>hccn_tool -i {chip_phy_id} -speed -g</code>、<code>hccn_tool -i {chip_phy_id} -duplex -g</code>、<code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/net_conf.log</code><br>V3: <code>optical.log</code>/<code>lldp.log</code>（duplex不支持）</td>
<td>通过 LLDP 信息定位对端交换机端口，对比两端速率与双工模式是否一致（任一端为 auto 时不告警）</td>
<td>NPU端口与对端交换机：SWITCH-01，ip：0.0.0.1，端口10GE1/0/1连接信息不相同，本端Speed：100G，Duplex：full。对端Speed：50G，Duplex：full</td>
</tr>
<tr>
<td>NPU对端lldp信息缺失检测</td>
<td><code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>lldp.log</code></td>
<td>检查 NPU 光模块对端 lldp 信息是否采集到</td>
<td>未采集到NPU光模块对端lldp信息</td>
</tr>
<tr>
<td>环回检测</td>
<td><code>hccn_tool -i {npu_id} -optical -t {model}</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: 暂不支持</td>
<td>根据环回测试状态码判定故障位置：环回类型 1 后端口 down 判定为本端故障；环回类型 1 后端口 up 但环回类型 2 后端口 down 判定为本端端口光模块故障/脏污</td>
<td>本端环回类型1后端口down，诊断为本端故障</td>
</tr>
<tr>
<td>双端光模块光功率检测</td>
<td rowspan="3"><code>hccn_tool -i {chip_phy_id} -optical -g</code>（主机）、<code>dis optical-module interface {interface}</code>（交换机）、<code>hccn_tool -i {chip_phy_id} -lldp -g</code>（主机）</td>
<td rowspan="3">主机：V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code>；交换机：<code>switch_cli_output.txt</code></td>
<td>通过 LLDP 信息获取对端端口，对双端光模块 TX/RX 功率进行对比分析</td>
<td>光模块光功率异常：本端RX功率-18.5dBm低于阈值-15dBm，对端交换机TX功率-12.0dBm正常</td>
</tr>
<tr>
<td>双端光模块SNR检测</td>
<td>通过 LLDP 信息获取对端端口，对双端光模块 Host SNR / Media SNR 进行对比分析</td>
<td>本端SNR值为7.2dB低于阈值8.0dB，对端交换机SNR值为9.5dB正常</td>
</tr>
<tr>
<td>双端光模块电流检测</td>
<td>通过 LLDP 信息获取对端端口，对双端光模块偏置电流进行对比分析</td>
<td>本端偏置电流85mA低于阈值90mA，对端交换机偏置电流105mA正常</td>
</tr>
</tbody>
</table>

> Host 日志 V1/V2/V3 版本详情请参考 [host 离线日志采集](../05_usage/02_log_collection.md#host-offline-log)。

## BMC 相关诊断

<table>
<thead>
<tr>
<th>诊断项</th>
<th>在线命令</th>
<th>离线日志路径</th>
<th>诊断逻辑</th>
<th>异常输出示例</th>
</tr>
</thead>
<tbody>
<tr>
<td>BMC告警故障码检测</td>
<td><code>ipmcget -d sel -v list</code></td>
<td><code>dump_info/AppDump/event/sel.txt</code></td>
<td>解析 BMC SEL 日志中的事件代码，按错误码与关键字匹配异常描述并输出处置建议</td>
<td>详见下方告警码列表</td>
</tr>
<tr>
<td>BMC光模块历史信息检测</td>
<td><code>ipmcget -t sensor -d list</code></td>
<td><code>dump_info/AppDump/sensor/sensor_info.txt</code>、<code>dump_info/AppDump/network_adapter/optical_module/optical_module_history_info_log.csv</code></td>
<td>检测到光模块历史信息即记录 linkdown 时间（可能为闪断或硬件故障），并按 lane 检查光功率、偏置电流、Host SNR / Media SNR 是否超出阈值，Tx Los / Rx Los 状态值是否大于 0</td>
<td>NPU存在linkdown，记录时间2025-10-01 14:30:00，可能为闪断或硬件故障；lane0: RX功率-18.5dBm低于阈值-15dBm；Tx los值0x1大于0</td>
</tr>
</tbody>
</table>

### 告警码详情

下表列出工具识别的主要 BMC 告警码及其处置建议。同一错误码可能根据事件描述关键字（如电压链路标识）进一步细分为多种故障。

<table>
<thead>
<tr>
<th>事件类别</th>
<th>事件码</th>
<th>异常描述</th>
<th>处置建议</th>
</tr>
</thead>
<tbody>
<tr>
<td>HBM ECC</td>
<td><code>0x80e01801</code></td>
<td>发生多 Bit ECC 故障</td>
<td>请对该 NPU HBM 进行压测</td>
</tr>
<tr>
<td>HBM ECC</td>
<td><code>0x80e18402</code></td>
<td>多 Bit ECC 故障，隔离行已满 64</td>
<td>请立即更换 NPU 备件</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb800a</code></td>
<td>AIV 算子超时，NPU 热复位</td>
<td>建议对硬件做 AICode 压测</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb8009</code></td>
<td>AIV 总线访问错误</td>
<td>建议对硬件做 AICode 压测</td>
</tr>
<tr>
<td>NPU 健康</td>
<td><code>0x56000003</code></td>
<td>NPU 健康状态紧急告警</td>
<td>1、检查芯片温度是否过高（可能是散热异常、环境温度过高或进风口/出风口堵塞）；2、检查是否是地址异常或者内存泄漏等软件问题；3、若无法解决，请联系技术支持</td>
</tr>
<tr>
<td>NPU 掉卡</td>
<td><code>0x56000005</code></td>
<td>NPU connection has been lost 告警</td>
<td>发生掉卡故障，请联系运维处理</td>
</tr>
<tr>
<td>NPU 过热</td>
<td><code>0x56000009</code></td>
<td>NPU 过热关机</td>
<td>NPU 过热关机，请联系运维处理</td>
</tr>
<tr>
<td>异常下电</td>
<td><code>0x2C000007</code></td>
<td>AI 模组异常下电告警/系统异常下电告警/NPU 异常下电告警（按电压链路标识细分：<code>V_1V2_DVDD_HBM02_FIX</code>、<code>V_0V9_AIC_DVFS_DA</code>、<code>V_AVDD12_HVCC</code>、<code>V_AVDD08_LVCC</code>、<code>V_0V8_DVDD_SIOE</code>）</td>
<td>AC 重启，不恢复则更换对应模组；或 20A PSIP 故障/12V 电容失效，请联系运维处理</td>
</tr>
<tr>
<td>上电超时</td>
<td><code>0x2C00002B</code></td>
<td>上电超时告警，主板有电压跌落（<code>V_AVDD12_HVCC</code>）</td>
<td>1、检查外部供电是否满足服务器整机功耗要求；2、通过拔插电源线或拔插单板彻底下电再上电，检查告警是否清除；3、若无法解决，请联系技术支持更换可能涉及的部件</td>
</tr>
<tr>
<td>上电超时</td>
<td><code>0x5D00001D</code></td>
<td>54V 上电超时/NPU 异常掉电（按电压链路标识细分：<code>54V0_HAM</code>、<code>V_DVDD25_2V5_HBM_FIX</code>、<code>V_DVDD075_HBMPHY_FIX</code>、<code>V_AVDD08_LVCC</code>）</td>
<td>NPU 的 54V 链路上的某器件失效，或 6A PSIP 故障/12V 电容失效，或 20A PSIP 故障/12V 电容失效，请联系运维处理</td>
</tr>
<tr>
<td>NPU 异常掉电</td>
<td><code>0x5D00001F</code></td>
<td>NPU 异常掉电告警（按电压链路标识细分：<code>V_DVDD09_BUS_DVFS</code>、<code>PG_12V0_</code>、<code>V_DVDD25_2V5_HBM_FIX</code>、<code>V_DVDD075_HBMPHY_FIX</code>、<code>PG_54V0_HAM</code>、<code>V_DRMOS</code>）</td>
<td>AC 重启不恢复则更换对应模组；或 6A/20A PSIP 故障；或 54V 链路器件失效；或电池砖高温触发保护掉电，请联系运维处理</td>
</tr>
<tr>
<td>PSU 过温</td>
<td><code>0x5D000005</code></td>
<td>PSU 过温告警</td>
<td>PSU 温度过高，请联系运维处理</td>
</tr>
<tr>
<td>液冷</td>
<td><code>0x12000023</code></td>
<td>液冷装置发生漏液</td>
<td>液冷装置发生漏液，请联系运维处理</td>
</tr>
<tr>
<td>液冷（LAAC）</td>
<td><code>0x120000C3</code></td>
<td>液冷装置(LAAC)异常，液冷泵不在位</td>
<td>液冷泵不在位，请联系运维处理</td>
</tr>
<tr>
<td>液冷（LAAC）</td>
<td><code>0x120000C7</code></td>
<td>液冷装置(LAAC)异常，液冷泵转速异常</td>
<td>液冷泵转速异常，请联系运维处理</td>
</tr>
<tr>
<td>液冷（LAAC）</td>
<td><code>0x120000C9</code></td>
<td>液冷装置(LAAC)异常，液冷泵故障</td>
<td>液冷泵故障，请联系运维处理</td>
</tr>
<tr>
<td>风扇</td>
<td><code>0x04000005</code></td>
<td>风冷散热模块故障，风扇冗余失效</td>
<td>风扇冗余失效，请联系运维处理</td>
</tr>
<tr>
<td>风扇</td>
<td><code>0x04000007</code></td>
<td>风冷散热模块故障，风扇转速偏差大</td>
<td>风扇转速偏差大，请联系运维处理</td>
</tr>
<tr>
<td>风扇</td>
<td><code>0x18000003</code></td>
<td>风冷散热模块故障，风扇背板电源故障</td>
<td>风扇背板电源故障，请联系运维处理</td>
</tr>
<tr>
<td>风扇</td>
<td><code>0x1800000D</code></td>
<td>风冷散热模块故障，风扇背板 MCU 自检异常</td>
<td>风扇背板 MCU 自检异常，请联系运维处理</td>
</tr>
</tbody>
</table>

## 交换机相关诊断

<table>
<thead>
<tr>
<th>诊断项</th>
<th>在线命令</th>
<th>离线日志路径</th>
<th>诊断逻辑</th>
<th>异常输出示例</th>
</tr>
</thead>
<tbody>
<tr>
<td>交换机端口误码率检测</td>
<td><code>display interface troubleshooting | no-more</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>检查端口的 BER 误码率是否超过阈值 <code>5.0e-06</code></td>
<td>BER误码率1.2e-05大于阈值5e-06。</td>
</tr>
<tr>
<td>交换机CRC错误告警检测</td>
<td rowspan="3"><code>display alarm active</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>解析错误码 <code>0x081300BC</code> 的告警信息，识别 CRC 错误快速增长的端口及其对端设备/端口</td>
<td>端口10GE1/0/1 CRC快速增长告警统计次数1500，阈值1000，对端设备SWITCH-02，对端端口10GE1/0/2</td>
</tr>
<tr>
<td>交换机端口降Lane检测</td>
<td>解析错误码 <code>0xF10509</code> 的告警信息，识别发生降 lane 的端口及其对端端口信息</td>
<td>端口10GE1/0/1发生降lane，对端端口信息：10GE1/0/2</td>
</tr>
<tr>
<td>交换机端口Los告警检测</td>
<td>解析错误码 <code>0x8130059</code> 的告警信息，识别光模块链路 Los 告警的端口及原因</td>
<td>光模块链路Los告警，原因：接收信号丢失</td>
</tr>
<tr>
<td>光模块状态检测（State-flag/Datapath State/Module State）</td>
<td><code>display interface transceiver verbose</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>检查光模块 Flag（收发光指标，预期 Normal）、Datapath State（通道状态，预期 Active）、Module State（功率模式，预期 Ready）字段值，按 lane 输出异常项</td>
<td>端口10GE1/0/1 flag信息异常：lane2 收发光指标 Flag值异常：0x03，预期为：Normal</td>
</tr>
<tr>
<td>单端光模块光功率检测</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>检查光模块 TX/RX 功率值是否超出阈值范围</td>
<td>光模块光功率异常：RX功率-19.0dBm低于阈值-15dBm</td>
</tr>
<tr>
<td>单端光模块SNR检测</td>
<td>检查光模块 Host SNR / Media SNR 值是否低于阈值</td>
<td>本端信噪比SNR异常，SNR值为6.5dB低于阈值8.0dB</td>
</tr>
<tr>
<td>单端光模块电流检测</td>
<td>检查光模块偏置电流是否超出阈值范围</td>
<td>本端电流异常，偏置电流78mA低于阈值90mA</td>
</tr>
<tr>
<td>双端光模块光功率检测</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code>（交换机）、<code>hccn_tool -i {chip_phy_id} -optical -g</code>（主机）</td>
<td rowspan="3">交换机 <code>switch_cli_output.txt</code>、主机 <code>hccn_log/optical.log</code></td>
<td>通过端口映射关系获取对端，对双端光模块 TX/RX 功率进行对比分析</td>
<td>光模块光功率异常：本端交换机RX功率-18.5dBm，对端主机TX功率-12.0dBm</td>
</tr>
<tr>
<td>双端光模块SNR检测</td>
<td>通过端口映射关系获取对端，对双端光模块 SNR 进行对比分析</td>
<td>本端交换机SNR值为7.2dB，对端主机SNR值为9.5dB</td>
</tr>
<tr>
<td>双端光模块电流检测</td>
<td>通过端口映射关系获取对端，对双端光模块偏置电流进行对比分析</td>
<td>本端交换机偏置电流85mA，对端主机偏置电流105mA</td>
</tr>
</tbody>
</table>

## HCCS 相关诊断

<table>
<thead>
<tr>
<th>诊断项</th>
<th>在线命令</th>
<th>离线日志路径</th>
<th>诊断逻辑</th>
<th>异常输出示例</th>
</tr>
</thead>
<tbody>
<tr>
<td>HCCS Serdes检测</td>
<td><code>display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}"</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>1. hilink_type 取值为 4。检查 CDR 是否失锁（<code>cdr_los == "1"</code>）<br>2. hilink_type 取值为 1。电源故障码 <code>csr119_data</code> 是否以 <code>0x380</code> 开头</td>
<td>交换芯片：0，端口：24，存在CDR失锁，存在电源故障，故障码：0x3802</td>
</tr>
<tr>
<td>HCCS 端口SNR检测（源端）</td>
<td><code>display interface hilink snr</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>检查交换机端口 hilink SNR 中异常 lane 的 SNR 值是否低于阈值</td>
<td>lane0 SNR值6.8dB低于阈值8.0dB</td>
</tr>
<tr>
<td>HCCS 端口SNR检测（目的端）</td>
<td><code>display interface hilink snr</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>检查 HCCS 交换芯片端口 SNR 是否低于阈值，并附带对端 XPU（CPU/NPU）端口信息</td>
<td>对端NPU0端口，lane 0 SNR值6.8dB低于阈值8.0dB</td>
</tr>
<tr>
<td>HCCS RP TX 超时检测（本端）</td>
<td rowspan="2"><code>display hccs proxy response statistics</code>、<code>display hccs proxy response detail interface {interface}</code></td>
<td rowspan="2"><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>当发生 RP TX 超时时，结合端口长期 down、端口闪断、窝包（RP/VOQ）等情况判定本端端口故障</td>
<td>交换机端口长期down；或 交换机端口闪断；或 rp窝包</td>
</tr>
<tr>
<td>HCCS RP TX 超时检测（对端）</td>
<td>当发生 RP TX 超时时，结合对端端口 LP 方向路由 miss、LP 窝包、VOQ 窝包等情况判定对端端口故障</td>
<td>lp方向路由miss</td>
</tr>
<tr>
<td>HCCS RX超时检测</td>
<td><code>display hccs proxy response statistics</code>、<code>display interface</code>、<code>display for info enp s 1 c {chip_id} "get port link start 0 end 47"</code></td>
<td><code>switch_cli_output.txt</code>、<code>diag_info.txt</code></td>
<td>筛选发生 RX 超时的接口（<code>rp_rx</code> 或 <code>lp_tx</code> 超时），结合端口长期 down、端口闪断、链路降 lane、XPU 设备异常等判断故障原因</td>
<td>交换机端口长期down；或 交换机端口闪断；或 交换机链路降lane；或 xpu设备异常</td>
</tr>
<tr>
<td>HCCS链路降级检测（端口级别）</td>
<td rowspan="2"><code>ipmcget -d healthevents</code></td>
<td rowspan="2"><code>dump_info/AppDump/event/current_event.txt</code></td>
<td>解析错误码 <code>0x28000049</code> 的事件描述，定位 CPU 板 UBC/macro 与 L1 端口之间的故障</td>
<td>Cpu0 UBC0 macro0 CPU board 0与L1端口之间发生故障</td>
</tr>
<tr>
<td>HCCS链路降级检测（板级别）</td>
<td>当某 L1 交换芯片所有 CPU 端口均异常时，判定为板级别故障</td>
<td>L1交换芯片所有端口异常</td>
</tr>
</tbody>
</table>

## 通用诊断

<table>
<thead>
<tr>
<th>诊断项</th>
<th>在线命令</th>
<th>离线日志路径</th>
<th>诊断逻辑</th>
<th>异常输出示例</th>
</tr>
</thead>
<tbody>
<tr>
<td>端口Lane间功率差检测</td>
<td><code>hccn_tool -i {chip_phy_id} -optical -g</code>（主机）、<code>dis optical-module interface {interface}</code>（交换机）、<code>ipmcget -t sensor -d list</code>（BMC）</td>
<td>主机 <code>hccn_log/optical.log</code>、交换机 <code>switch_cli_output.txt</code>、BMC <code>dump_info/AppDump/sensor/sensor_info.txt</code></td>
<td>计算同一端口不同 lane 间 TX/RX 功率的最大值与最小值差值，判断是否超过阈值（3dB）。支持主机、交换机和 BMC 三个维度的检测</td>
<td>TX端口Lane最大值和最小值差值大于3dB，实际最大值lane0：-12.0dBm，最小值lane3：-16.0dBm</td>
</tr>
</tbody>
</table>
