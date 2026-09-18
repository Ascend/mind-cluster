# Log Collection

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:17.581Z pushedAt=2026-08-24T02:22:41.538Z -->

This section introduces the log collection process from the user's perspective, explaining how ascend-fd-tk obtains the log data required for diagnosing link faults on servers (Host), BMC, and switches.

## Collection Mode Overview

The tool supports two data collection modes:

| Mode | Applicable Scenario | Prerequisites |
|------|----------|----------|
| **Online SSH collection** | Devices are network-accessible (IP/account and password/key/password-free), and data is collected in real time. | The port from the node where the tool resides to the target device is reachable. |
| **Offline log collection** | Log files have already been obtained and only need to be archived and analyzed. | Logs have been collected in advance to a directory on the node where the tool resides. |

>[!NOTE]
>
> - For online collection commands, select the corresponding reference documentation based on the product model, for example: [Atlas 800T A3 SuperPoD iBMC Commands](https://support.huawei.com/enterprise/en/doc/EDOC1100517755/f3754fb1/about-this-document), [Switch Command Reference](https://support.huawei.com/enterprise/en/switches/s3700-s5700-s6700-pid-259602657?category=reference-guides&subcategory=command-reference).
> - Offline log directory structure description: The tool supports automatic parsing of compressed packages, so no manual decompression is required during use. The offline log directory structure below shows the internal hierarchy of the compressed package to illustrate the important files within it.

**Quick Navigation**

- [Server (Host) Logs](#server-host-logs)
- [BMC Logs](#bmc-logs)
- [Switch Logs](#switch-logs)

<a id="server-host-log"></a>

## Server (Host) Logs

### Online Data Collection

The tool automatically runs the following commands on the server over SSH for collection.

| Category | Command | Description |
|------|------|------|
| System information | `hostname` | Host name |
| System information | `dmidecode -s system-serial-number` | Host SN |
| NPU type | `lspci \| grep 'Device d80' --color=never` | NPU model identifier |
| NPU mapping | `npu-smi info -m` | NPU/chip ID mapping |
| msnpureport log | `msnpureport` | Device-side log export |
| Optical module | `hccn_tool -i {chip_phy_id} -optical -g` | Optical module power/SNR/CDR |
| Optical module DFX configuration | `hccn_tool -i {chip_phy_id} -optical -g dfx_cfg` | Optical module DFX configuration |
| Link statistics | `hccn_tool -i {chip_phy_id} -link_stat -g` | Link-layer statistics (error packets, packet loss) |
| Link status | `hccn_tool -i {chip_phy_id} -link -g` | Link up/down/health status |
| Network health | `hccn_tool -i {chip_phy_id} -net_health -g` | Network health status |
| CDR SNR | `hccn_tool -i {chip_phy_id} -scdr -t 5` | CDR SNR information |
| Performance | `hccn_tool -i {chip_phy_id} -stat -g` | Performance counters |
| LLDP | `hccn_tool -i {chip_phy_id} -lldp -g` | LLDP neighbors |
| HCCS | `npu-smi info -t hccs -i {npu_id} -c {chip_id}` | HCCS protocol information |
| SPOD | `npu-smi info -t spod-info -i {npu_id} -c {chip_id}` | SPOD fault location |
| RoCE speed | `hccn_tool -i {chip_phy_id} -speed -g` | RoCE speed |
| RoCE duplex | `hccn_tool -i {chip_phy_id} -duplex -g` | RoCE duplex mode |

<a id="host-offline-log"></a>

### Offline Log Collection

The tool supports three versions of offline log structures. Version identification is performed automatically by the tool and does not require manual specification. Collect logs by using any of the following [host log collection scripts](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/ascend-fd-tk/host_collector). After collection, a `{file_name}.tar.gz` file is obtained. Place the compressed package directly into the log collection directory.

- Version 1: Collect logs by executing the `tool_log_collection_out_version_all_<version>.sh` script.
- Version 2: Collect logs by executing the `device_log_collect_<version>.sh` script.
- Version 3: Collect logs by executing the `link_down_collect_<version>.sh` script.

#### Version 1

```text
host log collection directory/
└── {file_name}.tar.gz/
    ├── timestamp directory (e.g., 2023-10-01_14-30-00)   # Device-side logs exported using msnpureport
    ├── hccn_tool.log                       # Network configuration tool logs
    ├── npu_card_info.log                   # NPU card information
    ├── pcie_info.log                       # PCIe information
    └── version_info.log                    # Version information
```

#### Version 2

```text
host log collection directory/
└── {file_name}.tar.gz/
    ├── timestamp directory (e.g., 2023-10-01_14-30-00)    # Device-side logs exported using msnpureport
    ├── hccn_log/
    │   ├── net_conf.log
    │   ├── optical.log
    │   └── stat.log
    ├── npu_smi_log/
    │   └── npu_smi.log
    └── pcie_log/
        └── pcie.log
```

#### Version 3

```text
host log collection directory/
└── {file_name}.tar.gz/
    ├── timestamp directory (e.g., 2023-10-01_14-30-00)    # Device-side logs exported using msnpureport
    ├── lldp.log
    └── optical.log
```

<a id="BMC-log"></a>

## BMC Logs

### Online Data Collection

The tool automatically collects the following data items through the BMC IPMI protocol.

| Data Item | Collection Command | Description |
|--------|----------|------|
| BMC serial number | `ipmcget -d serialnumber` | Unique device identifier |
| BMC date and time | `ipmcget -d time` | Event timestamp baseline |
| SEL log | `ipmcget -d sel -v list` | System event log (core diagnostic basis) |
| Sensor information | `ipmcget -t sensor -d list` | Temperature/voltage/fan and other sensors |
| Health events | `ipmcget -d healthevents` | Current health alarms |

In addition, the tool provides a built-in command [collect_bmc_dump_info](../06_api/03_collect/collect_bmc_dump_info.md) for online collection of BMC dump info logs. This command triggers BMC one-click collection through `ipmcget -d diaginfo` and downloads `dump_info.tar.gz` to `{home_directory}/cache/bmc_dump_cache/`. After decompression, the downloaded archive contains the data items shown in the table above and can be directly placed into the BMC log collection directory for offline diagnosis.

Example of BMC log collection in non-interactive mode (showing commands and output):

```bash
# Configure BMC information (IP/account and password/key/password-free) and BMC log collection
ascend-fd-tk set_conn_config /home/user/conn.ini collect_bmc_dump_info
Collection complete. Please check the log path: {path}
```

<a id="BMC-offline-log"></a>

### Offline Log Collection

BMC logs can be collected using either of the following methods:

- **Method 1**: On the [BMC web page](https://support.huawei.com/enterprise/en/doc/EDOC1100517755/897d7845/login?idPath=23710424|251366513|22892968|252309113|261716443), use the "one-click collection" button to download logs.
- **Method 2**: Log in to the [BMC platform](https://support.huawei.com/enterprise/en/doc/EDOC1100517755/897d7845/login?idPath=23710424|251366513|22892968|252309113|261716443) and use the `ipmcget -d diaginfo` command to collect logs.

The BMC log collection directory structure is as follows:

```text
BMC log collection directory/
└── {file_name}.tar.gz/
    └── dump_info/
        └── AppDump/
            ├── bmc_network/network_info.txt           # Network information
            ├── frudata/fruinfo.txt                    # FRU information
            ├── event/sel.txt                          # System event logs
            ├── event/current_event.txt                # Current health events
            ├── sensor/sensor_info.txt                 # Sensor information
            ├── network_adapter/optical_module/optical_module_history_info_log.csv  # Optical module history 1
            └── CpuMem/NpuIO/optical_module_history_info_log.csv                    # Optical module history 2
```

<a id="switch-log"></a>

## Switch Logs

### Online Data Collection

After logging in to the switch via SSH, the tool automatically executes the following commands.

| Data | Command |
|------|------|
| Serial number | `display license esn` |
| Interface summary | `dis int b \| no-more` |
| Switch name | `dis cu \| in sysname` |
| Optical module | `dis optical-module interface {interface} \| no-more` |
| Bit error rate | `display interface troubleshooting \| no-more` |
| LLDP neighbors | `dis lldp nei b \| n` |
| Active alarms | `display alarm active \| no-more` |
| Historical alarms | `display alarm history \| no-more` |
| Interface information | `display interface \| no-more` |
| Current time | `display clock \| include -` |
| HCCS capability detection | `display hccs eid ub-instance 0 \| no-more` |
| HCCS proxy response statistics | `display hccs proxy response statistics \| no-more` |
| HCCS proxy response details | `display hccs proxy response detail \| no-more` |
| HCCS route missing | `display hccs route miss statistics \| no-more` |
| HCCS mapping table | `display hccs decode and map table \| in MAP_TABLE \| no-more` |
| Port link status | `display for info enp s 1 c {chip_id} "get port link start 0 end 47" \| no-more` |
| Port statistics | `display for info enp s 1 c {chip_id} "get port statistic count port {port_id} module {module} type 0 path 2" \| no-more` |
| HCCS port-invalid drop statistics | `display hccs port-invalid drop statistics \| no-more` |
| Port credit back-pressure statistics | `display qos port-credit back-pressure statistics \| no-more` |
| Port SNR | `display for info enp s 1 c {chip_id} "get port snr port-id {port_id}"` |
| Interface hilink SNR | `dis int hilink snr \| n` |
| Transceiver information | `display interface transceiver verbose \| no-more` |
| Interface channel information | `display interface information \| no-more` |
| Serdes dump information | `display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}" \| no-more` |

<a id="switch-offline-log"></a>

### Offline Log Collection

The following two types of logs are mainly required for switch offline logs:

- **CLI command output logs (diag text logs)**: Contain the execution results of various switch commands. Use either of the following methods to collect them:
  - Method 1: After logging in to the switch, run `display diagnostic-information {filename}.txt`.
  - Method 2: After logging in to the switch, manually run the key commands (which must include `display current-configuration`), save the output text to a `.txt` file, and export it. For the commands to run, see [Switch Commands](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/ascend-fd-tk/switch_collector).
- **Diagnostic logs**: Structured logs generated by the switch diagnostic tool (`diagnostic_information.zip`). Log in to the switch, run `collect diagnostic information`, and export the zip package.

Compress all logs collected via the above methods into a single compressed package and place it directly into the switch log collection directory. The logs will be automatically extracted and analyzed during parsing.

The switch log directory structure is as follows:

```text
Switch log collection directory/
└── {file_name}.zip/
    ├── *.txt                                   # CLI command output logs (any .txt filename)
    └── diagnostic_information.zip/
        ├── slot_1.zip/
        │   └── tempdir/
        │       └── port_down_status.log        # Port down status logs
        └── logfile_slot_1.zip/
            └── tempdir/
                └── diag.log.zip/
                    └── diag.log                # Diagnostic logs (includes switch name, SNR, etc.)
```
