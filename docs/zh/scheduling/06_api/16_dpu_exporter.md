# Prometheus Metrics接口<a name="ZH-CN_TOPIC_0000002511346366"></a>

## 功能说明<a name="section_dpu_metrics_desc"></a>

提供Metrics接口，供Prometheus调用和集成。

## URL<a name="section_dpu_metrics_url"></a>

`GET http://ip:port/metrics`

>[!NOTE]
>DPU Exporter默认侦听端口8080，请求IP为部署节点IP。

## 请求参数<a name="section_dpu_metrics_params"></a>

无

## 响应说明<a name="section_dpu_metrics_response"></a>

按照Prometheus的专用格式返回数据，仅供参考，以实际回显为准。DPU Exporter采集两类指标：

- [全局指标（DPU卡级）](#section_global_metrics)
- [Interface级指标](#section_interface_metrics)

## 全局指标（DPU卡级）<a name="section_global_metrics"></a>

通过`hinicadm5 counter -t 1 -i hinic0 -f feature:ROCE`命令查询DPU卡的RoCE计数器（详见[hinicadm5 counter命令说明](https://support.huawei.com/enterprise/zh/doc/EDOC1100587368/3885f9bc?idPath=23710424%7C251364417%7C261608284%7C262279635%7C266404674#ZH-CN_TOPIC_0000002578666279)），经白名单过滤并格式化后输出。指标名格式为`dpu_<metric_name>`（metric_name为小写），label包含`card`和`card_type`。

### 默认白名单

当未配置自定义白名单时，使用以下默认白名单过滤指标：

`roce_err_ctr_*`、`roce_warn_ctr_*`、`roce_cmdq_ctr_roce_cmd_2err_qp`、`roce_cmdq_ctr_roce_cmd_sqerr2rts_qp`、`roce_cmdq_ctr_roce_data_cqe_ro_enable`、`roce_cmdq_ctr_roce_rq_cqe_128_enable`、`roce_cmdq_ctr_roce_sq_cqe_128_enable`、`roce_cmdq_ctr_shadow_function_invalid`、`roce_dp_ctr_ccp_token_not_enough`、`roce_dp_ctr_db_mtu_error_cnt`、`roce_dp_ctr_sq_datalen_over_limit_cnt`、`roce_dp_ctr_rr_ecn_rx`、`roce_dp_ctr_sw_ecn_rx`、`roce_dp_ctr_cnp_rx_entry`、`roce_dp_ctr_cnp_tx_entry`、`roce_dp_ctr_fast_cnp_event_entry`、`roce_dp_ctr_port_cnp_rx_entry`、`roce_dp_ctr_port_cnp_tx_entry`。

白名单的配置方法和匹配规则详见[白名单配置说明](../05_developer_guide/00_installation_deployment/00_manual_installation/13_dpu_exporter.md#白名单配置说明)。

### 响应示例

```text
# HELP dpu_roce_cmdq_ctr_ext_cmd roce counter: cmdq_ctr_ext_cmd
# TYPE dpu_roce_cmdq_ctr_ext_cmd gauge
dpu_roce_cmdq_ctr_ext_cmd{card="hinic0",card_type="CAL_2X400G_UB_EXP"} 2
dpu_roce_cmdq_ctr_roce_cc_resource_drain{card="hinic0",card_type="CAL_2X400G_UB_EXP"} 1
...
```

## Interface级指标<a name="section_interface_metrics"></a>

通过读取`/sys/class/net/<interface_name>/`目录下的文件采集网络接口的链路状态和流量统计指标。指标名格式为`dpu_interface_<metric_name>`，label包含`card`和`interface`。

**表 1**  Interface级指标

|指标名|说明|Labels|类型|
|------|----|------|----|
|dpu_interface_carrier|物理链路载波状态（1=up, 0=down）|card, interface|gauge|
|dpu_interface_carrier_changes|物理链路载波状态发生改变的累计次数|card, interface|counter|
|dpu_interface_operstate|网络接口的当前操作状态（1=up, 0=down, -1=other）|card, interface|gauge|
|dpu_interface_collisions|发送时检测到的数据包冲突数|card, interface|counter|
|dpu_interface_multicast|网络接口成功接收的多播（Multicast）数据包总数|card, interface|counter|
|dpu_interface_rx_bytes|网络接口成功接收的总数据量|card, interface|counter|
|dpu_interface_rx_packets|网络接口成功接收的总数据包数量|card, interface|counter|
|dpu_interface_rx_errors|接收过程中发生的总错误数|card, interface|counter|
|dpu_interface_rx_dropped|接收时丢弃的数据包数|card, interface|counter|
|dpu_interface_rx_crc_errors|接收时CRC失败的数据包数|card, interface|counter|
|dpu_interface_rx_frame_errors|接收时帧对齐错误的数据包数|card, interface|counter|
|dpu_interface_rx_fifo_errors|接收时由于硬件FIFO缓冲区溢出而丢失的数据包数|card, interface|counter|
|dpu_interface_rx_length_errors|接收时长度不符合规范的数据包数|card, interface|counter|
|dpu_interface_rx_missed_errors|接收时因硬件原因而错过/丢失的数据包数|card, interface|counter|
|dpu_interface_rx_over_errors|接收端过载导致的错误数|card, interface|counter|
|dpu_interface_rx_compressed|接收到的压缩数据包数量|card, interface|counter|
|dpu_interface_rx_nohandler|接收到的数据包因没有对应的协议处理器而被内核丢弃的数量|card, interface|counter|
|dpu_interface_tx_bytes|网络接口成功发送的总数据量|card, interface|counter|
|dpu_interface_tx_packets|网络接口成功发送的总数据包数量|card, interface|counter|
|dpu_interface_tx_errors|发送过程中发生的总错误数|card, interface|counter|
|dpu_interface_tx_dropped|发送时丢弃的数据包数|card, interface|counter|
|dpu_interface_tx_aborted_errors|发送过程中被异常中止的错误次数|card, interface|counter|
|dpu_interface_tx_carrier_errors|发送时载波错误数|card, interface|counter|
|dpu_interface_tx_fifo_errors|发送时由于硬件FIFO缓冲区下溢（欠载）而未能成功发送的数据包数|card, interface|counter|
|dpu_interface_tx_heartbeat_errors|发送时心跳信号丢失的错误数|card, interface|counter|
|dpu_interface_tx_window_errors|发送时的窗口错误数|card, interface|counter|
|dpu_interface_tx_compressed|发送的压缩数据包数量|card, interface|counter|

### 响应示例

```text
# HELP dpu_interface_carrier link carrier status (1=up, 0=down)
# TYPE dpu_interface_carrier gauge
dpu_interface_carrier{card="hinic0",interface="ens3f0"} 1
dpu_interface_rx_bytes{card="hinic0",interface="ens3f0"} 1234567890
dpu_interface_tx_bytes{card="hinic0",interface="ens3f0"} 987654321
...
```
