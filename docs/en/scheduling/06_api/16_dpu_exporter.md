# Prometheus Metrics Interface<a name="ZH-CN_TOPIC_0000002511346366"></a>

## Function Description<a name="section_dpu_metrics_desc"></a>

Provides the Metrics interface for Prometheus to call and integrate.

## URL<a name="section_dpu_metrics_url"></a>

`GET http://ip:port/metrics`

>[!NOTE]
>DPU Exporter listens on port 8083 by default, and the request IP is the IP address of the deployment node.

## Request Parameters<a name="section_dpu_metrics_params"></a>

None

## Response Description<a name="section_dpu_metrics_response"></a>

Data is returned in the Prometheus-specific format, which is for reference only and the actual output shall prevail. DPU Exporter collects the following two types of metrics:

- [Global Metrics (DPU Level)](#section_global_metrics)
- [Interface-Level Metrics](#section_interface_metrics)

## Global Metrics (DPU Level)<a name="section_global_metrics"></a>

The RoCE counters of a DPU are queried by running the `hinicadm5 counter -t 1 -i hinic0 -f feature:ROCE` command (for details, see [hinicadm5 counter Command Description](https://support.huawei.com/enterprise/en/doc/EDOC1100606234/426cffd9/about-this-document?idPath=23710424|251364417|261608284|262279635|266404674)), filtered by the trustlist, formatted, and then output. The metric name format is `dpu_<metric_name>` (metric_name is in lowercase), and labels include `card` and `card_type`.

### Default Trustlist

When no custom trustlist is configured, the following default trustlist is used to filter metrics:

`roce_err_ctr_*`, `roce_warn_ctr_*`, `roce_cmdq_ctr_roce_cmd_2err_qp`, `roce_cmdq_ctr_roce_cmd_sqerr2rts_qp`, `roce_cmdq_ctr_roce_data_cqe_ro_enable`, `roce_cmdq_ctr_roce_rq_cqe_128_enable`, `roce_cmdq_ctr_roce_sq_cqe_128_enable`, `roce_cmdq_ctr_shadow_function_invalid`, `roce_dp_ctr_ccp_token_not_enough`, `roce_dp_ctr_db_mtu_error_cnt`, `roce_dp_ctr_sq_datalen_over_limit_cnt`, `roce_dp_ctr_rr_ecn_rx`, `roce_dp_ctr_sw_ecn_rx`, `roce_dp_ctr_cnp_rx_entry`, `roce_dp_ctr_cnp_tx_entry`, `roce_dp_ctr_fast_cnp_event_entry`, `roce_dp_ctr_port_cnp_rx_entry`, `roce_dp_ctr_port_cnp_tx_entry`.

For the configuration method and matching rules of the trustlist, see [Trustlist Configuration Description](../05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md#trustlist-configuration-description).

### Response Example

```text
# HELP dpu_roce_cmdq_ctr_ext_cmd roce counter: cmdq_ctr_ext_cmd
# TYPE dpu_roce_cmdq_ctr_ext_cmd gauge
dpu_roce_cmdq_ctr_ext_cmd{card="hinic0",card_type="CAL_2X400G_UB_EXP"} 2
dpu_roce_cmdq_ctr_roce_cc_resource_drain{card="hinic0",card_type="CAL_2X400G_UB_EXP"} 1
...
```

## Interface-Level Metrics<a name="section_interface_metrics"></a>

The link status and traffic statistics metrics of the network interface are collected by reading files in the `/sys/class/net/<interface_name>/` directory. The metric name format is `dpu_interface_<metric_name>`, and labels include `card` and `interface`.

**Table 1** Interface-level metrics

|Metric Name|Description|Labels|Type|
|------|----|------|----|
|dpu_interface_carrier|Physical link carrier status (1=up, 0=down, -1=unknown)|card, interface|gauge|
|dpu_interface_carrier_changes|Cumulative number of times the physical link carrier status changes|card, interface|counter|
|dpu_interface_operstate|Current operational state of the network interface (1=up, 0=down, -1=other)|card, interface|gauge|
|dpu_interface_collisions|Number of packet collisions detected during transmission|card, interface|counter|
|dpu_interface_multicast|Total number of multicast packets successfully received by the network interface|card, interface|counter|
|dpu_interface_rx_bytes|Total amount of data successfully received by the network interface|card, interface|counter|
|dpu_interface_rx_packets|Total number of packets successfully received by the network interface|card, interface|counter|
|dpu_interface_rx_errors|Total number of errors that occur during reception|card, interface|counter|
|dpu_interface_rx_dropped|Number of packets dropped during reception|card, interface|counter|
|dpu_interface_rx_crc_errors|Number of packets with CRC failures during reception|card, interface|counter|
|dpu_interface_rx_frame_errors|Number of packets with frame alignment errors during reception|card, interface|counter|
|dpu_interface_rx_fifo_errors|Number of packets lost during reception due to hardware FIFO buffer overflow|card, interface|counter|
|dpu_interface_rx_length_errors|Number of packets with non-conforming lengths during reception|card, interface|counter|
|dpu_interface_rx_missed_errors|Number of packets missed/lost during reception due to hardware reasons|card, interface|counter|
|dpu_interface_rx_over_errors|Number of errors caused by receiver overload|card, interface|counter|
|dpu_interface_rx_compressed|Number of compressed packets received|card, interface|counter|
|dpu_interface_rx_nohandler|Number of received packets dropped by the kernel because no corresponding protocol handler exists|card, interface|counter|
|dpu_interface_tx_bytes|Total amount of data successfully sent by the network interface|card, interface|counter|
|dpu_interface_tx_packets|Total number of packets successfully sent by the network interface|card, interface|counter|
|dpu_interface_tx_errors|Total number of errors that occur during transmission|card, interface|counter|
|dpu_interface_tx_dropped|Number of packets dropped during transmission|card, interface|counter|
|dpu_interface_tx_aborted_errors|Number of errors aborted abnormally during transmission|card, interface|counter|
|dpu_interface_tx_carrier_errors|Number of carrier errors during transmission|card, interface|counter|
|dpu_interface_tx_fifo_errors|Number of packets that failed to be sent during transmission due to hardware FIFO buffer underrun|card, interface|counter|
|dpu_interface_tx_heartbeat_errors|Number of heartbeat signal loss errors during transmission|card, interface|counter|
|dpu_interface_tx_window_errors|Number of window errors during transmission|card, interface|counter|
|dpu_interface_tx_compressed|Number of compressed packets sent|card, interface|counter|

### Response Example

```text
# HELP dpu_interface_carrier link carrier status (1=up, 0=down)
# TYPE dpu_interface_carrier gauge
dpu_interface_carrier{card="hinic0",interface="ens3f0"} 1
dpu_interface_rx_bytes{card="hinic0",interface="ens3f0"} 1234567890
dpu_interface_tx_bytes{card="hinic0",interface="ens3f0"} 987654321
...
```

## Health Probe<a name="section_dpu_health_probe"></a>

DPU Exporter starts the HTTP health probe service inside the component, which is used by the K8s livenessProbe mechanism to detect the component liveness status. The probe service is completely decoupled from the service HTTP service of DPU Exporter and uses a dedicated port (independent of the Metrics port 8083).

**Table 2** Health probe interfaces

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default port | 11259 |
| Protocol | HTTP (HTTPS when the `--tls-cert-file` and `--tls-private-key-file` parameters are correctly configured) |

**Table 3** Response description

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | The component is running properly | The response body is `ok` |
| 404 Not Found | The request path is not `/` | The probe responds only to the root path |
| 405 Method Not Allowed | The request method is not GET | K8s livenessProbe uses GET by default, so this is not triggered normally |
| 503 Service Unavailable | A custom health check callback is registered and the check fails | The response body contains the specific error information |

**K8s livenessProbe configuration example:**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11259
    scheme: HTTP
  initialDelaySeconds: 10
  periodSeconds: 15
  timeoutSeconds: 5
  failureThreshold: 3
```

For details about the startup parameters of the health probe, see [DPU Exporter Startup Parameters](../05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md#parameter-description).
