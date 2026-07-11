# Detailed Description of Data Sources

## Overview

This document provides a detailed description of the various data sources supported by Ascend-faultdiag-toolkit, including:

- Data types and specific data items
- Data collection methods (SSH online collection/offline log parsing)
- Data source locations (specific commands/file paths)

The Ascend-faultdiag-toolkit supports data source collection from the following three device types:

- Host
- Baseboard Management Controller (BMC)
- Switch

Each device supports the following two collection methods:

- **SSH online collection**: Connects to the device via SSH and executes commands to obtain real-time data.
- **Offline log parsing**: Parses exported log files to extract historical data.

## Host Data Source

### SSH Online Collection

| Data Item | Collection Command | Description |
|--------|------------------------------------------------------|------|
| Hostname | `hostname` | Obtains the host name |
| NPU mapping information | `npu-smi info -m` | Obtains the NPU mapping relationship (NPU ID, chip ID, physical ID) |
| Optical module information | `hccn_tool -i {chip_phy_id} -optical -g` | Obtains the optical module information of the specified physical chip |
| Link status | `hccn_tool -i {chip_phy_id} -link_stat -g` | Obtains the link status of the specified physical chip |
| Statistics information | `hccn_tool -i {chip_phy_id} -stat -g` | Obtains the statistics information of the specified physical chip |
| LLDP information | `hccn_tool -i {chip_phy_id} -lldp -g` | Obtains the LLDP neighbor information of the specified physical chip |
| NPU type | `lspci \| grep 'Device d80'` | Obtains the NPU device type |
| System serial number | `dmidecode -s system-serial-number` | Obtains the system serial number |
| HCCS information | `npu-smi info -t hccs -i {npu_id} -c {chip_id}` | Obtains the HCCS information of the specified NPU and chip |
| SPOD information | `npu-smi info -t spod-info -i {npu_id} -c {chip_id}` | Obtains the SPOD information of the specified NPU and chip |
| MSNPUREPORT log | `msnpureport` | Generates and parses the MSNPUREPORT log |
| RoCE speed | `hccn_tool -i {chip_phy_id} -speed -g` | Obtains the RoCE speed of the specified physical chip |
| RoCE duplex mode | `hccn_tool -i {chip_phy_id} -duplex -g` | Obtains the RoCE duplex mode of the specified physical chip |
| Network health status | `hccn_tool -i {chip_phy_id} -net_health -g` | Obtains the network health status of the specified physical chip |
| Link status | `hccn_tool -i {chip_phy_id} -link -g` | Obtains the link status of the specified physical chip |
| CDR information | `hccn_tool -i {chip_phy_id} -scdr -t 5` | Obtains the CDR information of the specified physical chip |
| DFX configuration | `hccn_tool -i {chip_phy_id} -optical -g dfx_cfg` | Obtains the DFX configuration of the specified physical chip |
| Optical module loopback test | `hccn_tool -i {npu_id} -optical -t {model}` | Performs an optical module loopback test |

### Offline Log Parsing

Host offline logs support three different configuration versions, each with its own set of file paths and parsing keywords.

#### Version 1 Configuration (ParseConfigCollectionV1)

**File Structure**

```text
log_directory/
├── hccn_tool.log        # Network Configuration Tool Log
├── npu_card_info.log    # NPU Card Information Log
├── pcie_info.log        # PCIe Information Log
└── version_info.log     # Version Information Log
```

**Data Items and Parsing Configuration**

| Data Type | Parsing Keyword | Log File Path | Description |
|----------|------------|--------------|------|
| LLDP | `"lldp"` | `hccn_tool.log` | Link Layer Discovery Protocol information |
| SPEED | `"speed -g"` | `hccn_tool.log` | Port speed information |
| OPTICAL | `"optical"` | `hccn_tool.log` | Optical module information |
| LINK_STAT | `"link stat"` | `hccn_tool.log` | Link statistics |
| STAT | `"stat"` | `hccn_tool.log` | Performance statistics |
| HCCN_LINK_STATUS | `"link"`| `hccn_tool.log` | HCCN link status |
| CDR_SNR | `"cdr5 snr 1 times"`| `hccn_tool.log` | CDR SNR information |
| SPOD_INFO | `"Collect spod-info info for all NPUs"` | `npu_card_info.log` | SPOD information |
| HCCS | `"Collect hccs info for all NPUs"` | `npu_card_info.log` | HCCS protocol information |
| NPU_TYPE | `"lspci"` | `pcie_info.log` | NPU type information |`
| SN | `"timeout 30s dmidecode -t1"` | `version_info.log` | Serial number information |

#### Version 2 Configuration (ParseConfigCollectionV2)

**File Structure**

```text
log_directory/
├── hccn_log/
│   ├── net_conf.log     # Network Configuration Log
│   ├── optical.log      # Optical Module Log
│   └── stat.log         # Statistics Log
├── npu_smi_log/
│   └── npu_smi.log      # NPU SMI Log
└── pcie_log/
    └── pcie.log         # PCIe Log
```

**Data Items and Parsing Configuration**

| Data Type | Parsing Key | Log File Path | Description |
|----------|------------|--------------|------|
| LLDP | `"lldp"` | `hccn_log/net_conf.log` | Link Layer Discovery Protocol information |
| SPEED | `"speed"` | `hccn_log/net_conf.log` | Port speed information |
| OPTICAL | `"optical"` | `hccn_log/optical.log` | Optical module information |
| LINK_STAT | `"link stat"` | `hccn_log/optical.log` | Link statistics |
| NET_HEALTH | `"health info"` | `hccn_log/optical.log` | Network health information |
| HCCN_LINK_STATUS | `"link info"` | `hccn_log/optical.log` | HCCN link status |
| STAT | `"stat"` | `hccn_log/stat.log` | Performance statistics |
| SPOD_INFO | `"spod_info"` | `npu_smi_log/npu_smi.log` | SPOD information |
| HCCS | `"hccs"` | `npu_smi_log/npu_smi.log` | HCCS protocol information |
| NPU_TYPE | `"pcie"` | `pcie_log/pcie.log` | NPU type information |

#### Version 3 Configuration (ParseConfigCollectionV3)

**File Structure**

```text
log_directory/
├── lldp.log        # LLDP Log
└── optical.log     # Optical Module Log
```

**Data Items and Parsing Configuration**

| Data Type | Parsing Key | Log File Path | Description |
|----------|-------------|---------------|------|
| LLDP | `"lldp"` | `lldp.log` | Link Layer Discovery Protocol information |
| SPEED | `"speed info"` | `optical.log` | Port speed information |
| OPTICAL | `"optical"` | `optical.log` | Optical module information |
| LINK_STAT | `"link stat"` | `optical.log` | Link statistics |
| HCCN_LINK_STATUS | `"link info"` | `optical.log` | HCCN link status |

### MSNPUREPORT Logs

In addition to the above configuration, parsing of logs generated by the MSNPUREPORT tool is also supported:

- Path: Timestamp-named subdirectory under the log directory (e.g., `2023-10-01_14-30-00`)
- Content: Detailed NPU information

## BMC Data Source

### SSH Online Collection

**Data Items and Sources**

| Data Item | Collection Command | Description |
|--------|----------|------|
| BMC serial number | `ipmcget -d serialnumber` | Obtains the BMC serial number |
| BMC date and time | `ipmcget -d time` | Obtains the current BMC time |
| SEL log | `ipmcget -d sel -v list` | Obtains the system event log |
| Sensor information | `ipmcget -t sensor -d list` | Obtains the sensor list and status |
| Health event | `ipmcget -d healthevents` | Obtains the health event log |
| Diagnostic information | `ipmcget -d diaginfo` | Obtains BMC diagnostic information and downloads it as a compressed package |
| Optical module historical log | - | Obtains optical module historical records|

**Diagnostic Information Compressed Package**

- Remote path: `/tmp/dump_info.tar.gz`
- Local storage path: `CommonPath.TOOL_HOME_BMC_DUMP_CACHE_DIR`
- Naming format: `{host}_{sn_num}_{subfix_date_time}.tar.gz`

### Offline Log Parsing

**Log Source**

BMC offline logs are typically obtained through the diagnostic information archive (`dump_info.tar.gz`) generated by the `ipmcget -d diaginfo` command. After decompression, the directory structure is as follows:

```text
dump_info/
└── AppDump/
    ├── bmc_network/
    │   └── network_info.txt      # Network information file
    ├── frudata/
    │   └── fruinfo.txt           # FRU information file
    ├── event/
    │   ├── sel.txt               # System event log
    │   └── current_event.txt     # Current health event
    ├── sensor/
    │   └── sensor_info.txt       # Sensor information
    ├── network_adapter/
    │   └── optical_module/
    │       └── optical_module_history_info_log.csv  # Optical module history log 1
    └── CpuMem/
        └── NpuIO/
            └── optical_module_history_info_log.csv  # Optical module history log 2
```

**Data Items and Sources**

| Data Item | Log File Path | Parsing Method |
|--------|--------------|----------|
| BMC IP Address | `AppDump/bmc_network/network_info.txt` | Extracts the `"IP Address"` field from the file |
| Serial Number | `AppDump/frudata/fruinfo.txt` | Extracts `"System Serial Number"` from FRU information |
| SEL Information | `AppDump/event/sel.txt` | Directly reads SEL log content |
| Sensor Information | `AppDump/sensor/sensor_info.txt` | Directly reads sensor information content |
| Health Event | `AppDump/event/current_event.txt` | Directly reads health event content |
| Optical Module Historical Log | <ul><li>`AppDump/network_adapter/optical_module/optical_module_history_info_log.csv`</li><li>`AppDump/CpuMem/NpuIO/optical_module_history_info_log.csv`</li></ul> | Parses CSV-formatted optical module history records to extract link interruption records |

## Switch Data Source

### SSH Online Collection

| Data Item | Collection Command | Description |
|--------|----------|------|
| Switch Serial Number | `display license esn` | Obtains the switch ESN serial number |
| Interface Summary | `dis int b \| no-more` | Obtains the basic status of all interfaces |
| Switch Name | `dis cu \| in sysname` | Obtains the switch system name |
| Optical Module Information | `dis optical-module interface {interface}` | Obtains the optical module information of the specified interface |
| Bit Error Rate | `display interface troubleshooting {interface}` | Obtains the troubleshooting information of the specified interface |
| LLDP Neighbor | `dis lldp nei b \| n` | Obtains the LLDP neighbor summary |
| Active Alarm | `display alarm active \| no-more` | Obtains the current active alarms |
| Historical Alarm | `display alarm history \| no-more` | Obtains the historical alarm records |
| Interface Information | `display interface \| no-more` | Obtains the detailed information of all interfaces |
| Current Time | `display clock \| include -` | Obtains the current time of the switch |
| HCCS Proxy Response Statistics | `display hccs proxy response statistics \| no-more` | Obtains the HCCS proxy response statistics |
| HCCS Proxy Response Details | `display hccs proxy response detail interface {interface}` | Obtains the HCCS proxy response details of the specified interface |
| HCCS Route Miss | `display hccs route miss statistics \| no-more` | Obtains the HCCS route miss statistics |
| Port Link Status | `display for info enp s 1 c {chip_id} "get port link start 0 end 47" \| no-more` | Obtains the chip port link status |
| Port Statistics | `display for info enp s 1 c {chip_id} "get port statistic count port {port_id} module {module} type 0 path 2" \| no-more` | Obtains the port statistics |
| HCCS Port-Invalid Drop | `display hccs port-invalid drop statistics \| no-more` | Obtains the HCCS port-invalid drop statistics |
| Port Credit Back-Pressure Statistics | `display qos port-credit back-pressure statistics \| no-more` | Obtains the port credit back-pressure statistics |
| HCCS Port SNR | `display interface hilink snr \| n` | Obtains the HCCS port signal-to-noise ratio |
| Transceiver Information | `display interface transceiver verbose \| no-more` | Obtains the detailed interface transceiver information |
| Interface Channel Information | `display interface information \| no-more` | Obtains the interface channel information |
| Serdes Dump Information | `display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}" \| no-more` | Obtains the Serdes dump information |

### Offline Log Parsing

Switch offline logs primarily come in the following two forms:

- **CLI command output log**: Contains the execution results of various switch commands.
- **Diagnostic information output**:Structured logs generated by switch diagnostic tools.

**Log File Locations**

- CLI output log: Usually a single file (such as `switch_cli_output.txt`) or multiple files categorized by command.
- Diagnostic information output: Usually a file with a diagnostic marker (such as `diag_info.txt`).

**Data Types and Sources**

| Data Type | Log Source Characteristics | Description |
|----------|--------------|------|
| Active Alarm Details | Content block containing `AlarmId, AlarmName, AlarmType, State : active` fields | Extracts detailed information of active alarms from CLI output logs |
| Historical Alarm Details | Content block containing `AlarmId, AlarmName, AlarmType, State : cleared` fields | Extracts detailed information of cleared alarms from CLI output logs |
| Active Alarms | Table containing `Sequence, AlarmId, Severity, Date Time, Description` fields | Extracts active alarm summaries from CLI output logs |
| Historical Alarms | Table containing `Sequence, AlarmId, Severity, Date Time, Description` fields with historical markers | Extracts historical alarm records from CLI output logs |
| LLDP Neighbors | Table containing `Local Interface, Exptime(s), Neighbor Interface, Neighbor Device` fields | Extracts LLDP neighbor information from CLI output logs |
| Optical Module Information | Table containing `Items, Value, HighAlarm, HighWarn, LowAlarm, Status` fields | Extracts optical module monitoring data from CLI output logs |
| Interface Summary | Table containing `Interface, PHY, Protocol, InUti, OutUti, inErrors, outErrors` fields | Extracts basic interface status from CLI output logs |
| Bit Error Rate | Content block containing `Current state, Speed` fields | Extracts interface bit error rate information from CLI output logs |
| Interface Information | Content block containing `current state, Description, Port Mode` fields | Extracts detailed interface configuration from CLI output logs |
| License ESN | Content block containing `MainBoard, ESN` fields | Extracts switch serial number from CLI output logs |
| System Clock | Content block containing `clock, Time Zone` fields | Extracts the current switch time from CLI output logs |
| Transceiver Information | Content block containing the `transceiver information:` marker | Extracts detailed transceiver information from CLI output logs |
| HCCS Related Information | Various tables and content blocks containing the `HCCS` keyword | Extracts various statistical information related to the HCCS protocol from CLI output logs |
