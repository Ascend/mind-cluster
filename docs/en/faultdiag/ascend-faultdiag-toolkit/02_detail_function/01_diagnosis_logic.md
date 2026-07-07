# Diagnosis Logic

## Diagnosis Logic Overview

Ascend-faultdiag-toolkit implements fault diagnosis for cluster devices based on multi-source data collection and analysis. It supports two modes: online SSH collection and offline log parsing.

## Host-Related Diagnosis

### Host Optical Module Comprehensive Diagnosis

**Inputs**

- Host SSH online collection:
  - Optical module information: `hccn_tool -i {chip_phy_id} -optical -g`
  - Link status: `hccn_tool -i {chip_phy_id} -link_stat -g`
- Host offline log (version 1):
  - `hccn_tool.log` (Optical information)
  - `npu_card_info.log` (NPU information)
- Host offline log (Version 2):
  - `hccn_log/optical.log` (Optical module information)
  - `npu_smi_log/npu_smi.log` (NPU information)
- Host offline log (version 3):
  - `optical.log` (optical module information)

**Diagnosis Logic**

Checks multiple dimensions including optical module status, power, SNR, CDR parameters, `uncorr_cw_cnt`, and IIC faults, comparing against predefined thresholds to identify anomalies.

**Abnormal Outputs**

- Optical module power anomaly: When the TX/RX power value exceeds the threshold range

  Example: "Optical module power anomaly, TX power: -5dBm (threshold range: -3~0dBm)"
- SNR anomaly: When the SNR value falls below the threshold (e.g., 12dB)

  Example: "Optical module SNR anomaly, current value: 10.5dB (threshold: 12dB)"
- CDR loss of lock: When the CDR status is "Unlock"

  Example: "Optical module CDR unlock, status: Unlock"
- `uncorr_cw_cnt` exceeding threshold: When the `uncorr_cw_cnt` value exceeds the predefined threshold

  Example: "Optical module uncorr_cw_cnt over-threshold, current value: 1,000 (threshold: 500)"
- IIC communication fault: When the IIC communication status is "Failed"

  Example: "Optical module IIC communication fault, status: Failed"

### Host Loopback Diagnosis

**Inputs**

- Host SSH online collection: `hccn_tool -i {npu_id} -optical -t {model}` (loopback test result)
- Host offline log (version 1): `hccn_tool.log` (Optical information)
- Host offline log (version 2): `hccn_log/optical.log` (optical module information)
- Host offline log (version 3): `optical.log` (optical module information)

**Diagnosis Logic**

Checks the status code of the loopback test to identify loopback failures.

**Abnormal Outputs**

- Loopback test failed: when the loopback test status code is a non-zero value

  Example: "Loopback test failed, status code: 1, indicating a host internal optical link or module fault"

### Host Optical Module Los/LoL Diagnosis

**Inputs**

- Host SSH online collection: `hccn_tool -i {chip_phy_id} -optical -g` (`tx_los`, `rx_los`, `rx_lol` status)
- Host offline log (Version 1): `hccn_tool.log` (Optical information)
- Host offline log (version 2): `hccn_log/optical.log` (optical module information)
- Host offline log (version 3): `optical.log` (optical module information)

**Diagnosis Logic**

Parses optical module status fields to determine if there is a loss of optical signal or laser shutdown.

**Abnormal Outputs**

- TX Los alarm: when the `tx_los` field value is "1"

  Example: "Optical module TX Los alarm, tx_los status: 1"
- RX Los alarm: when the `rx_los` field value is "1"

  Example: "Optical module RX Los alarm, rx_los status: 1"
- RX LoL alarm: when the `rx_lol` field value is "1"

  Example: "Optical module RX LoL alarm, rx_lol status: 1"

### Host NPU Port Status Diagnosis

**Inputs**

- Host SSH online collection:
  - Link status: `hccn_tool -i {chip_phy_id} -link -g`
  - Network health status: `hccn_tool -i {chip_phy_id} -net_health -g`
- Host offline log (version 1): `hccn_tool.log` (link status information)
- Host offline log (version 2): `hccn_log/optical.log` (network health information)

**Diagnosis Logic**

Checks the NPU port's optical module presence status, network health status, and connection status.

**Abnormal Outputs**

- Port fault: When the port status field value is "Fault"

  Example: "NPU port fault, status: Fault"
- Network anomaly: When the network health status field value is "Abnormal"

  Example: "NPU port network anomaly, health status: Abnormal"
- Connection disconnected: When the connection status field value is "Disconnected"

  Example: "NPU port connection disconnected, Status: Disconnected"

### Inter-Host Optical Link Diagnosis

**Inputs**

- Host SSH online collection: `hccn_tool -i {chip_phy_id} -optical -g` (optical module parameters)
- Host offline log (version 1): `hccn_tool.log` (Optical information)
- Host offline log (version 2): `hccn_log/optical.log` (optical module information)
- Host offline log (version 3): `optical.log` (optical module information)

**Diagnosis Logic**

Performs multi-dimensional checks on parameters such as optical module power, SNR, and current for inter-host connections.

**Abnormal Outputs**

- Inter-host optical link power anomaly: When the TX/RX power value of the optical module connected between hosts exceeds the threshold range

  Example: "Inter-host optical link power anomaly, TX power: -6dBm (threshold range: -3~0dBm)"
- SNR anomaly: When the SNR value of the optical module connected between hosts falls below the threshold

  Example: "Inter-host optical link SNR anomaly, current value: 11.2dB (threshold: 12dB)"
- Current anomaly: When the current value for the inter-host connection exceeds the threshold range.

  Example: "Inter-host optical link current anomaly, current value: 15.5mA (threshold range: 5~15mA)"

### RoCE Port Configuration Diagnosis

**Inputs**

- Host SSH online collection:
  - Port speed: `hccn_tool -i {chip_phy_id} -speed -g`
  - Duplex mode: `hccn_tool -i {chip_phy_id} -duplex -g`
- Host offline log (Version 1): `hccn_tool.log` (speed information)
- Host offline log (Version 2): `hccn_log/net_conf.log` (port speed information)
- Host offline log (Version 3): `optical.log` (speed info)

**Diagnosis Logic**

Compares the speed and duplex mode configurations between the NPU port and the peer switch port.

**Abnormal Outputs**

- Speed mismatch: When the NPU port speed is inconsistent with the peer switch end port speed

  Example: "RoCE port speed mismatch, NPU port: 100Gbps, switch port: 50Gbps"
- Duplex mode mismatch: When the NPU port duplex mode is inconsistent with that of the peer switch

  Example: "RoCE port duplex mode mismatch, NPU port: Full, switch port: Half"

## BMC-Related Diagnosis

### BMC Error Code Analysis

**Inputs**

- BMC SSH online collection: `ipmcget -d sel -v list` (SEL log)
- BMC offline logs:
  - `AppDump/event/sel.txt` (system EVENT logs)
  - `AppDump/event/current_event.txt` (current health events)

**Diagnosis Logic**

Parses event codes in BMC logs to identify hardware anomalies.

**Abnormal Outputs**

- Multi-Bit ECC: When the event code contains "0x80e01801"

  Example: "BMC hardware anomaly, event code: 0x80e01801, multi-bit ECC occurred"
- Multi-Bit ECC, with 6a isolated lines: When the event code contains "0x80e18402"

  Example: "BMC hardware anomaly, event code: 0x80e18402, multi-bit ECC fault, isolated lines reaching 64"
- AIV operator timeout, NPU hot reset: When the event code contains "0x80cb800a"

  Example: "BMC hardware anomaly, event code: 0x80cb800a, AIV operator timeout, NPU hot reset"
- AIV bus access error: When the event code contains "0x80cb8009"

  Example: "BMC hardware anomaly, event code: 0x80cb8009, AIV bus access error"

### BMC Optical Module Diagnosis

**Inputs**

- BMC SSH online collection:
  - Sensor information: `ipmcget -t sensor -d list`
  - Optical module history logs
- BMC offline logs:
  - `AppDump/network_adapter/optical_module/optical_module_history_info_log.csv` (optical module history log 1)
  - `AppDump/CpuMem/NpuIO/optical_module_history_info_log.csv` (optical module history log 2)

**Diagnosis Logic**

Checks whether parameters such as optical power, bias current, and SNR exceed thresholds, and whether the Los status is abnormal.

**Abnormal Outputs**

- Optical module power anomaly: When the optical power value exceeds the threshold range

  Example: "BMC optical module power anomaly, RX power: -25dBm (threshold range: -20~-10dBm)"
- Bias current anomaly: When the bias current value exceeds the threshold range

  Example: "BMC optical module bias current anomaly, current value: 16.2mA (threshold range: 5~15mA)"
- SNR anomaly: When the SNR value falls below the threshold

  Example: "BMC optical module SNR anomaly, current value: 10.8dB (threshold: 12dB)"
- Link Los alarm: When the Los status field value is "1"

  Example: "BMC optical module link Los alarm, Los status: 1"

### HCCS Link Degradation Diagnosis

**Inputs**

- BMC SSH online collection: `ipmcget -d healthevents` (health EVENT logs)
- BMC offline log: `AppDump/event/current_event.txt` (current health events)

**Diagnosis Logic**

Parses the event description for a specific error code (0x28000049) to locate the faulty port.

**Abnormal Outputs**

HCCS link degradation: When the event code contains "0x28000049"

Example: "HCCS link degradation, event code: 0x28000049, indicating a faulty L1 switch port or CPU tray"

## Switch-Related Diagnosis

### Switch Optical Module Diagnosis

**Inputs**

- Switch SSH online collection: `dis optical-module interface {interface}` (optical module information)
- Switch offline log: Optical module information table in command output (containing `Items, Value, HighAlarm, HighWarn, LowAlarm, Status` fields)

**Diagnosis Logic**

Analyzes parameters such as optical module power, SNR, and current between switch ports, supporting single-end and dual-end diagnosis.

**Abnormal Outputs**

- Optical module power anomaly: When the TX/RX power value exceeds the threshold range

  Example: "Switch optical module power anomaly, TX power: -4dBm (threshold range: -3~0dBm)"
- SNR anomaly: When the SNR value falls below the threshold

  Example: "Switch optical module SNR anomaly, current value: 11.5dB (threshold: 12dB)"
- Current anomaly: When the current value exceeds the threshold range

  Example: "Switch optical module current anomaly, current value: 15.8mA (threshold range: 5~15mA)"

### Switch Port Error Rate Diagnosis

**Inputs**

- Switch SSH online collection: `display interface troubleshooting {interface}` (error rate information)
- Switch offline log: bit error rate information in command output (containing `Current state` and `Speed` fields)

**Diagnosis Logic**

Checks whether the port's bit error rate (BER) exceeds the threshold.

**Abnormal Outputs**

BER exceeding the threshold: when the BER exceeds the predefined threshold (e.g., 1e-12)

Example: "Switch port bit error rate exceeds threshold, current value: 5e-12 (threshold: 1e-12), indicating a link quality issue"

### Switch CRC Error Diagnosis

**Inputs**

- Switch SSH online collection: `display alarm active` (active alarm)
- Switch offline log: active alarm information table (including `Sequence`, `AlarmId`, `Severity`, `Date Time`, and `Description` fields)

**Diagnosis Logic**

Parses alarm information for a specific error code (0x081300bc) to identify ports with rapidly growing CRC errors.

**Abnormal Outputs**

Rapid increase in port CRC errors: When the alarm information contains the error code "0x081300bc".

Example: "Rapid increase in CRC errors on the switch port, Alarm ID: 0x081300bc, Port: GigabitEthernet0/0/1, indicating a link quality issue or hardware fault."

### Switch Port Lane Reduction Diagnosis

**Inputs**

- Switch SSH online collection: `display alarm active` (active alarm)
- Switch offline log: active alarm information table (including `Sequence`, `AlarmId`, `Severity`, `Date Time`, and `Description` fields)

**Diagnosis Logic**

Parses alarm information for a specific error code (0xf10509) to identify ports with lane reduction.

**Abnormal Outputs**

Port lane reduction alarm: When the alarm information contains the error code "0xf10509"

Example: "Switch port lane reduction alarm, alarm ID: 0xf10509, port: GigabitEthernet0/0/2, indicating a link or hardware fault"

### Switch Optical Module Los Alarm Diagnosis

**Inputs**

- Switch SSH online collection: `display alarm active` (active alarm)
- Switch offline log: Active alarm information table (including `Sequence`, `AlarmId`, `Severity`, `Date Time`, and `Description` fields)

**Diagnosis Logic**

Parses alarm information for a specific error code (0x8130059) to identify ports with optical module link Los alarms.

**Abnormal Outputs**

Optical module link Los alarm: When the alarm information contains error code "0x8130059"

Example: "Switch optical module link Los alarm, alarm ID: 0x8130059, port: GigabitEthernet0/0/3, indicating optical signal loss"

### Switch Optical Module Status Diagnosis

**Inputs**

- Switch SSH online collection: `display interface transceiver verbose` (transceiver details)
- Switch offline log: transceiver information in command output (content blocks containing the `transceiver information:` tag)

**Diagnosis Logic**:

Checks the status values of fields such as `State Flag`, `Datapath State`, and `Module State`.

**Abnormal Output**

Optical module status anomaly: When the values of the `State Flag`, `Datapath State`, or `Module State` fields are abnormal

Example: "Switch optical module status anomaly, State Flag: 0x00000001, Datapath State: Fault, indicating issues with optical transceiver metrics, channel status, or power mode"

## General Diagnosis

### Port Lane Power Difference Diagnosis

**Inputs**

- Host SSH online collection: `hccn_tool -i {chip_phy_id} -optical -g` (optical module lane power information)
- Switch SSH online collection: `dis optical-module interface {interface}` (optical module lane power information)
- BMC SSH online collection: sensor information (`ipmcget -t sensor -d list`)
- Offline log: optical module lane power data (log files from host, switch, or BMC)

**Diagnosis Logic**

Calculates the difference between the maximum and minimum power values across different lanes of the same port and checks if it exceeds the threshold (3 dB).

**Abnormal Outputs**

Excessive lane power difference: When the difference between the maximum and minimum power values across different lanes on the same port exceeds the threshold (3 dB).

Example: "Excessive port lane power difference, Port: eth0, Maximum difference: 4.2 dB (Threshold: 3 dB), indicating an internal port lane fault."

## HCCS-Related Diagnosis

### HCCS RP TX Timeout Diagnosis

**Inputs**

- Switch SSH online collection:
  - `display hccs proxy response statistics` (HCCS proxy response statistics)
  - `display hccs proxy response detail interface {interface}` (HCCS proxy response details)
- Switch offline log: HCCS-related statistics in command output (tables containing HCCS keywords)

**Diagnosis Logic**

Analyzes the interface and address mapping relationships for RP TX timeouts, checking port status and peer interface status.

**Abnormal Outputs**

RP TX timeout caused by long-term port down, intermittent disconnection, or packet trapping: When the RP TX timeout count in HCCS agent response statistics is greater than 0

Example: "HCCS RP TX timeout, interface: eth0, timeout count: 10, possible causes: long-term port down, intermittent disconnection, or packet trapping"

### HCCS RX Timeout Diagnosis

**Input**

- Switch SSH online collection:
  - `display hccs proxy response statistics` (HCCS agent response statistics)
  - `display interface` (interface status)
  - `display for info enp s 1 c {chip_id} "get port link start 0 end 47"` (port link status)
- Switch offline logs:
  - HCCS-related statistics in command output (tables containing the HCCS keyword)
  - Interface status information in command output (containing the `current state`, `Description`, and `Port Mode` fields)

**Diagnosis Logic**

Filters interfaces with RX timeouts and determines the cause of the fault by checking link status, lane degradation, etc.

**Abnormal Outputs**

RX timeout caused by long-term port down, intermittent disconnection, link lane reduction, or XPU device anomaly: When the RX timeout count in HCCS agent response statistics is greater than 0

Example: "HCCS RX timeout, interface: eth1, timeout count: 5, possible causes: port intermittent disconnection or link lane reduction"

### HCCS Serdes Diagnosis

**Inputs**

- Switch SSH online collection: `display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}"` (Serdes dump information)
- Switch offline log: Serdes dump information in command output

**Diagnosis Logic**

Checks CDR lock status and power fault codes to identify Serdes anomalies.

**Abnormal Outputs**

- Unlock CDR: When the CDR status is "Unlock"

  Example: "HCCS Serdes CDR unlocked, port: eth0, lane: 1"
- Power fault: When the power status is "Fault"

  Example: "HCCS Serdes power supply fault, port: eth0, lane: 2"

### HCCS Port SNR Diagnosis

**Inputs**

- Switch SSH online collection: `display interface hilink snr` (HCCS port signal-to-noise ratio (SNR))
- Switch offline log: HCCS port SNR information in command output (tables containing output of `display interface hilink snr`)

**Diagnosis Logic**

Compares port SNR values against a threshold to identify ports with SNR below the threshold.

**Abnormal Outputs**

SNR below threshold: When the SNR value is below the threshold (e.g., 12 dB)

Example: "HCCS port SNR abnormal, interface: eth2, current SNR: 10.5 dB (threshold: 12 dB), indicating a link quality issue"
