# Diagnostic Items

Diagnostic items are categorized by device type, and each diagnostic item specifies its online command source and offline log path. The same device may involve multiple diagnostic items, and each diagnostic item corresponds to one or more online commands and offline logs. The toolkit includes various diagnostic items, covering the host, BMC, switch, HCCS, and general dimensions.

## Diagnostic Item Index

| Device Type | Detailed Description |
|----------|----------|
| Host-side Diagnostics | [Host-side Diagnostics](#host-side-diagnostics) |
| BMC Related Diagnostics | [BMC Related Diagnostics](#bmc-related-diagnostics) |
| Switch-side Diagnostics | [Switch-side Diagnostics](#switch-side-diagnostics) |
| HCCS Related Diagnostics | [HCCS Related Diagnostics](#hccs-related-diagnostics) |
| General Diagnostics | [General Diagnostics](#general-diagnostics) |

## Host-side Diagnostics

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Offline Log</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>Optical Module communication link fault detection</td>
<td rowspan="8"><code>hccn_tool -i {chip_phy_id} -optical -g</code></td>
<td rowspan="8">V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code></td>
<td>Check whether the optical module control link is reachable.</td>
<td>Optical module: Control link unreachable</td>
</tr>
<tr>
<td>Optical module presence detection</td>
<td>Check whether the optical module is present.</td>
<td>Optical module not present, status: NA</td>
</tr>
<tr>
<td>High-power enable register status detection</td>
<td>Check the high power enable register status.</td>
<td>The optical module is in low power mode, high power enable reg:0x00</td>
</tr>
<tr>
<td>Single-ended optical module optical power detection</td>
<td>Check whether the TX/RX power values exceed the threshold range.</td>
<td>Optical module optical power abnormality, RX power -18.5 dBm; below the threshold of -15 dBm</td>
</tr>
<tr>
<td>Single-ended optical module SNR detection</td>
<td>Check whether the Host SNR/Media SNR values are below the threshold.</td>
<td>Optical module SNR abnormal: lane0: Host SNR value 7.2 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical module SNR inter-lane difference</td>
<td>Check whether the SNR difference between different lanes exceeds the threshold.</td>
<td>Optical module SNR inter-lane difference abnormal: the difference between LANE0 and LANE3 is 4.2 dB</td>
</tr>
<tr>
<td>Optical module Los detection</td>
<td>Check whether the Rx Los and Tx Los status values are greater than 0.</td>
<td>Optical module Rx Los indicator abnormal, status: 1</td>
</tr>
<tr>
<td>Optical module LoL detection</td>
<td>Check whether the Rx LoL and Tx LoL status values are greater than 0.</td>
<td>Optical module Rx LoL indicator abnormal, status: 1</td>
</tr>
<tr>
<td>Optical module uncorr_cw_cnt detection</td>
<td rowspan="2">Automatic collection by <code>msnpureport</code></td>
<td rowspan="2">V1-V3: timestamp subdirectory under the <code>msnpureport</code> report directory</td>
<td>Check whether uncorr_cw_cnt is greater than 10 for 3 consecutive times.</td>
<td>uncorr_cw_cnt > 10 occurred for 3 consecutive times, occurrence time: 2025-10-01-14:30:00.123456, 2025-10-01-14:30:01.234567, 2025-10-01-14:30:02.345678</td>
</tr>
<tr>
<td>Optical module IIC communication fault detection</td>
<td>Detect IIC communication abnormal events.</td>
<td>IIC abnormality detected: trans status[0x40], error status[0x10], the NPU onboard optical module adapter may be faulty</td>
</tr>
<tr>
<td>Optical module initialization light-on status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -optical -g dfx_cfg</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>Check whether the TX Disable status is disabled.</td>
<td>The optical module is in laser-off state, tx disable status: 1</td>
</tr>
<tr>
<td>Optical module CDR SNR detection</td>
<td rowspan="2"><code>hccn_tool -i {chip_phy_id} -scdr -t 5</code></td>
<td rowspan="2">V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>Check whether the CDR Host SNR/Media SNR values are below the threshold.</td>
<td>CDR SNR abnormal, Host SNR value 6.8 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical module CDR SNR inter-lane difference detection</td>
<td>Check whether the difference between the Host SNR/Media SNR of different CDR lanes exceeds the threshold.</td>
<td>CDR SNR inter-lane difference abnormal: the difference between LANE0 and LANE3 is 4.2 dB</td>
</tr>
<tr>
<td>Port network health status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -net_health -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code> (net_health not supported)</td>
<td>Check whether the network health status of the NPU port deviates from the normal threshold, and provides the peer switch and port information.</td>
<td>Port optical module status abnormal, network health status: abnormal. Peer switch: SWITCH-01, peer port: 10GE1/0/1.</td>
</tr>
<tr>
<td>Port connection status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -link -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code></td>
<td>Check whether the connection status of the NPU port deviates from the normal threshold, and provides the peer switch and port information.</td>
<td>Port optical module status abnormal, connection status: down. Peer switch: SWITCH-01, peer port: 10GE1/0/1.</td>
</tr>
<tr>
<td>Link status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -link_stat -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code></td>
<td>Count the number of port link downs within 24 hours: a count of 5 or more is determined as a link fault (critical), and 3 or more as a link sub-healthy state (warning).</td>
<td>Link fault: the number of link downs within 24 hours is 5. Down time: 2025-10-01 14:30:00, 2025-10-01 14:35:00</td>
</tr>
<tr>
<td>RoCE port configuration detection</td>
<td><code>hccn_tool -i {chip_phy_id} -speed -g</code>, <code>hccn_tool -i {chip_phy_id} -duplex -g</code>, <code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/net_conf.log</code><br>V3: <code>optical.log</code>/<code>lldp.log</code> (duplex not supported)</td>
<td>Locate the peer switch port through LLDP information, and compare whether the speed and duplex mode on both ends are consistent (no alarm is raised when either end is auto).</td>
<td>The connection information of the NPU port and the peer switch: SWITCH-01, ip: 0.0.0.1, port 10GE1/0/1 is inconsistent. Local Speed: 100G, Duplex: full. Peer Speed: 50G, Duplex: full</td>
</tr>
<tr>
<td>NPU peer LLDP information missing detection</td>
<td><code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>lldp.log</code></td>
<td>Check whether the LLDP information of the peer of the NPU optical module is collected.</td>
<td>The LLDP information of the peer of the NPU optical module is not collected</td>
</tr>
<tr>
<td>Loopback type 1 detection (local fault determination)</td>
<td rowspan="2"><code>hccn_tool -i {npu_id} -optical -t {model}</code></td>
<td rowspan="2">V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>If the port goes down after loopback type 1 is detected during the loopback test, it is determined as a local fault.</td>
<td>The local port is down after loopback type 1 is detected, it is diagnosed as a local fault.</td>
</tr>
<tr>
<td>Loopback type 2 detection (local port optical module fault/dirt determination)</td>
<td>If the port is up after loopback type 1 is detected but goes down after loopback type 2 is detected, it is determined as a local port optical module fault or dirt.</td>
<td>The port is up after local loopback type 1 is detected and down after loopback type 2 is detected, which is diagnosed as a local port optical module fault or dirt.</td>
</tr>
<tr>
<td>Dual-ended optical module optical power detection</td>
<td rowspan="3"><code>hccn_tool -i {chip_phy_id} -optical -g</code> (host), <code>dis optical-module interface {interface}</code> (switch), <code>hccn_tool -i {chip_phy_id} -lldp -g</code> (host)</td>
<td rowspan="3">Host: V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code>; Switch: <code>switch_cli_output.txt</code></td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the TX/RX power of the dual-ended optical module.</td>
<td>Optical module optical power abnormality: local RX power -18.5 dBm is below the threshold -15dBm, while the peer switch TX power -12.0 dBm is normal.</td>
</tr>
<tr>
<td>Dual-ended optical module SNR detection</td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the Host SNR/Media SNR of the dual-ended optical module.</td>
<td>Local SNR value 7.2 dB is below the threshold of 8.0 dB, while the peer switch SNR value 9.5 dB is normal.</td>
</tr>
<tr>
<td>Dual-ended optical module current detection</td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the bias current of the dual-ended optical module.</td>
<td>The local bias current is 85 mA, below the threshold of 90mA, while the peer switch bias current is 105 mA, which is normal.</td>
</tr>
</tbody>
</table>

>[!NOTE]
> For details about host log V1/V2/V3 versions, see [Offline Log Collection (Host)](../05_usage/02_log_collection.md#host-offline-log).

<!-- npu="950" id1 -->
### Additional Host-side Diagnostics for <term>Ascend 950PR&950DT Products</term>

The host-side diagnostic items of <term>Ascend 950PR&950DT products</term> have the same diagnostic logic as the corresponding items of <term>Atlas A3 products</term>, except that <term>Ascend 950PR&950DT products</term> additionally provide NIC SFP lane-level diagnostics.

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>Optical module presence detection</td>
<td rowspan="9"><code>hccn_tool -g -optical -i {npu_id} -u {udie_id} -p {port_id}</code></td>
<td>Check whether the optical module is present.</td>
<td>Optical module not present; status: NA.</td>
</tr>
<tr>
<td>High-power enable register status detection</td>
<td>Check the high power enable register status.</td>
<td>The optical module is in low power mode; high power enable reg:0x00</td>
</tr>
<tr>
<td>Optical module Los detection</td>
<td>Check whether the TxFault/TxLos/RxLos Flag status value of each lane is normal.</td>
<td>Optical module TxLos flag indicator abnormal: Lane1: 0x1</td>
</tr>
<tr>
<td>Optical module LoL detection</td>
<td>Check whether the TxLol/RxLol Flag status value of each lane is normal (LPO optical modules do not support the Lol flag, and the check is skipped automatically).</td>
<td>Optical module RxLol Flag indicator abnormal, status: Lane1: 0x1</td>
</tr>
<tr>
<td>Optical module bias current detection</td>
<td>Check whether the bias value of each lane exceeds the threshold range</td>
<td>Optical module bias abnormal: Lane1 bias current 78 mA is below the threshold of 90 mA</td>
</tr>
<tr>
<td>Optical module optical power detection</td>
<td>Check whether the TxPower/RxPower values of each lane exceed the threshold range.</td>
<td>Optical module optical power abnormal: Lane1 RX power -18.5 dBm is below the threshold of -15 dBm</td>
</tr>
<tr>
<td>Optical module Host SNR detection</td>
<td>Check whether the Host SNR value of each lane is lower than the threshold.</td>
<td>Optical module Host SNR abnormal: Lane1 Host SNR value 7.2 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical module Media SNR detection</td>
<td>Check whether the Media SNR value of each lane is lower than the threshold.</td>
<td>Optical module Media SNR abnormal: Lane1 Media SNR 7.2 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical module SNR inter-lane difference detection</td>
<td>Check whether the difference in Host SNR/Media SNR between lanes exceeds the threshold.</td>
<td>Optical module SNR inter-lane difference abnormal: Host SNR difference between Lane0 and Lane3 is 4.2 dB</td>
</tr>
<tr>
<td>Port virtual link priority credit insufficiency detection</td>
<td><code>hccn_tool -g -credit -i {npu_id} -u {udie_id} -p {port_id}</code></td>
<td>Check whether the allocated quantity (link_alloc_vl_pri_credit) and the used quantity (link_cur_used_pri_credit) of the credit of each priority are equal and non-zero (if both the allocated and used quantities are 0, the priority is not enabled, which is not considered an exception).</td>
<td>Port virtual link priority credit insufficiency: vl1 allocated quantity=8, used quantity=8</td>
</tr>
<tr>
<td>NIC optical module Los detection</td>
<td rowspan="5"><code>hinicadm5 sfp -i {card_name} -p {port_id}</code></td>
<td>Check whether the TxLos/RxLos status value of each lane on the NIC port is non-zero.</td>
<td>NIC port Lane1 TxLos abnormal: actual value 0x1 (normal value 0)</td>
</tr>
<tr>
<td>NIC optical module LoL detection</td>
<td>Check whether the TxCdrLol/RxCdrLol status value of each lane on the NIC port is non-zero.</td>
<td>NIC port Lane1 RxCdrLol abnormal: actual value 0x1 (normal value 0)</td>
</tr>
<tr>
<td>NIC optical module optical power detection</td>
<td>Check whether the TxPower/RxPower values of each lane of the NIC port exceed the threshold range.</td>
<td>NIC port tx_power abnormal: Lane1 RX power -18.5 dBm is below the threshold of -15 dBm</td>
</tr>
<tr>
<td>NIC optical module bias current detection</td>
<td>Check whether the bias value of each lane of the NIC port exceeds the threshold range.</td>
<td>NIC port bias abnormal: Lane1 bias current 78 mA is below the threshold of 90 mA</td>
</tr>
<tr>
<td>NIC optical nodule SNR detection</td>
<td>Check whether the Host SNR/Media SNR values of each lane of the NIC port are below the threshold.</td>
<td>NIC port host_snr abnormal: Lane1 Host SNR value 7.2 dB is below the threshold of 8.0 dB</td>
</tr>
</tbody>
</table>
<!-- end id1 -->

## BMC Related Diagnostics

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Offline Log</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>BMC alarm fault code detection</td>
<td><code>ipmcget -d sel -v list</code></td>
<td><code>dump_info/AppDump/event/sel.txt</code></td>
<td>Parse the event codes in the BMC SEL log, match abnormal descriptions by error code and keyword, and output handling suggestions.</td>
<td>See the alarm code details list below.</td>
</tr>
<tr>
<td>BMC optical module linkdown history detection</td>
<td rowspan="5"><code>ipmcget -t sensor -d list</code></td>
<td rowspan="5"><code>dump_info/AppDump/sensor/sensor_info.txt</code>, <code>dump_info/AppDump/network_adapter/optical_module/optical_module_history_info_log.csv</code></td>
<td>When optical module history information is detected, the linkdown time is recorded (it may be a flapping link or a hardware fault).</td>
<td>NPU linkdown exists, recorded at 2025-10-01 14:30:00. It may be a link flap or a hardware fault.</td>
</tr>
<tr>
<td>BMC optical module lane-level optical power detection</td>
<td>Check whether the TX/RX optical power of each lane exceeds the threshold range.</td>
<td>lane0: RX power -18.5 dBm, below the threshold of -15 dBm</td>
</tr>
<tr>
<td>BMC optical module lane-level bias current detection</td>
<td>Check whether the bias current of each lane exceeds the threshold range.</td>
<td>lane1: bias current 78 mA, below the threshold of 90 mA</td>
</tr>
<tr>
<td>BMC optical module lane-level SNR detection</td>
<td>Check whether the Host SNR/Media SNR of each lane is lower than the threshold.</td>
<td>lane0: Host SNR value 7.2 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>BMC optical module lane-level Los status detection</td>
<td>Check whether the Tx Los/Rx Los status value of each lane is greater than 0.</td>
<td>Tx los value 0x1 is greater than 0</td>
</tr>
<tr>
<td>BMC link degradation detection (port level)</td>
<td rowspan="2"><code>ipmcget -d healthevents</code></td>
<td rowspan="2"><code>dump_info/AppDump/event/current_event.txt</code></td>
<td>Parse the event description of error code <code>0x28000049</code> to locate the fault between the UBC/macro of the CPU board and the L1 port.</td>
<td>A fault occurs between Cpu0 UBC0 macro0 CPU board 0 and the L1 port.</td>
</tr>
<tr>
<td>BMC link degradation detection (board level)</td>
<td>When all CPU ports of an L1 switch chip are abnormal, it is determined as a board-level fault.</td>
<td>All ports of the L1 switch chip are abnormal.</td>
</tr>
</tbody>
</table>

### Alarm Code Details

The following table lists the main BMC alarm codes recognized by the tool and their handling suggestions. The same error code may be further subdivided into multiple faults based on event description keywords (such as voltage link identifiers).

<table>
<thead>
<tr>
<th>Event Category</th>
<th>Event Code</th>
<th>Abnormal Description</th>
<th>Recommended Action</th>
</tr>
</thead>
<tbody>
<tr>
<td>HBM ECC</td>
<td><code>0x80e01801</code></td>
<td>Multi-bit ECC fault occurred</td>
<td>Please run a stress test on the NPU on-chip memory.</td>
</tr>
<tr>
<td>HBM ECC</td>
<td><code>0x80e18402</code></td>
<td>Multi-bit ECC fault, isolated rows reached 64</td>
<td>Please replace the NPU spare part immediately.</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb800a</code></td>
<td>AIV operator timeout, NPU warm reset</td>
<td>It is recommended to perform AICore stress testing on the hardware.</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb8009</code></td>
<td>AIV bus access error</td>
<td>It is recommended to run AICore stress testing on the hardware.</td>
</tr>
<tr>
<td>HCCS</td>
<td><code>0x2800004B</code></td>
<td>HCCS link disconnection alarm</td>
<td>1. Power off the device and check for poor contact. 2. Try replacing the HCCS cable. 3. Replace the NPU board if necessary.</td>
</tr>
<tr>
<td>NPU health</td>
<td><code>0x56000003</code></td>
<td>NPU health status critical alarm</td>
<td>1. Check whether the chip temperature is too high (possibly due to abnormal heat dissipation, excessive ambient temperature, or blocked air inlet/outlet); 2. Check whether it is a software issue such as an address exception or memory leak; 3. If the issue cannot be resolved, contact technical support.</td>
</tr>
<tr>
<td>NPU Loss</td>
<td><code>0x56000005</code></td>
<td>Alarm: NPU connection has been lost</td>
<td>A card loss fault has occurred. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>NPU overheating</td>
<td><code>0x56000009</code></td>
<td>NPU shutdown due to overheating</td>
<td>The NPU is shutdown due to overheating. Contact O&M personnel.</td>
</tr>
<tr>
<td>Abnormal power-off</td>
<td><code>0x2C000007</code></td>
<td>AI module abnormal power-off alarm/system abnormal power-off alarm/NPU abnormal power-off alarm (subdivided by voltage link identifier: <code>V_1V2_DVDD_HBM02_FIX</code>, <code>V_0V9_AIC_DVFS_DA</code>, <code>V_AVDD12_HVCC</code>, <code>V_AVDD08_LVCC</code>, <code>V_0V8_DVDD_SIOE</code>)</td>
<td>AC restart; if not recovered, replace the corresponding module; or 20A PSIP fault/12V capacitor failure, contact O&M for handling</td>
</tr>
<tr>
<td>Power-on timeout</td>
<td><code>0x2C00002B</code></td>
<td>Power-on timeout alarm, voltage drop on the mainboard (<code>V_AVDD12_HVCC</code>)</td>
<td>1. Check whether the external power supply meets the overall power consumption requirements of the server; 2. Completely power off and then power on by unplugging the power cable or removing the board, and check whether the alarm is cleared; 3. If the issue persists, contact technical support to replace the possibly involved components</td>
</tr>
<tr>
<td>Power-on timeout</td>
<td><code>0x5D00001D</code></td>
<td>54V power-on timeout/NPU abnormal power loss (subdivided by voltage link identifier: <code>54V0_HAM</code>, <code>V_DVDD25_2V5_HBM_FIX</code>, <code>V_DVDD075_HBMPHY_FIX</code>, <code>V_AVDD08_LVCC</code>)</td>
<td>A component on the NPU 54V link has failed, or a 6A PSIP fault/12V capacitor failure, or a 20A PSIP fault/12V capacitor failure has occurred. Contact O&M for handling.</td>
</tr>
<tr>
<td>NPU abnormal power loss</td>
<td><code>0x5D00001F</code></td>
<td>NPU abnormal power loss alarm (subdivided by voltage link identifier: <code>V_DVDD09_BUS_DVFS</code>, <code>PG_12V0_</code>, <code>V_DVDD25_2V5_HBM_FIX</code>, <code>V_DVDD075_HBMPHY_FIX</code>, <code>PG_54V0_HAM</code>, <code>V_DRMOS</code>)</td>
<td>If the fault persists after an AC restart, replace the corresponding module; or a 6A/20A PSIP fault; or a 54V link component failure; or battery brick high-temperature protection-triggered power-off has occurred. Contact O&M for handling.</td>
</tr>
<tr>
<td>PSU over-temperature</td>
<td><code>0x5D000005</code></td>
<td>PSU over-temperature alarm</td>
<td>The PSU temperature is too high. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid cooling</td>
<td><code>0x12000023</code></td>
<td>Liquid cooling device leakage</td>
<td>Liquid cooling device leakage occurred. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid cooling (LAAC)</td>
<td><code>0x120000C3</code></td>
<td>Liquid cooling device (LAAC) abnormal, liquid cooling pump not present</td>
<td>Liquid cooling pump is not present. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid cooling (LAAC)</td>
<td><code>0x120000C7</code></td>
<td>Liquid cooling device (LAAC) abnormal, liquid cooling pump speed abnormal</td>
<td>Liquid cooling pump speed is abnormal. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid cooling (LAAC)</td>
<td><code>0x120000C9</code></td>
<td>Liquid cooling device (LAAC) is abnormal, and the liquid cooling pump is faulty.</td>
<td>The liquid cooling pump is faulty. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x04000005</code></td>
<td>Air cooling module fault, fan redundancy failure</td>
<td>The fan redundancy fails. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x04000007</code></td>
<td>Air cooling module fault, large fan speed deviation</td>
<td>The fan speed deviation is large. Contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x18000003</code></td>
<td>Air cooling module fault, fan backplane power supply fault</td>
<td>The fan backplane power supply fails. Contact O&M for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x1800000D</code></td>
<td>Air cooling module fault, fan backplane MCU self-check abnormal</td>
<td>The fan backplane MCU self-check is abnormal. Please contact O&M for handling</td>
</tr>
</tbody>
</table>

## Switch-side Diagnostics

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Offline Log</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>Switch port BER detection</td>
<td><code>display interface troubleshooting | no-more</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the port BER exceeds the threshold <code>5.0e-06</code></td>
<td>BER bit error rate 1.2e-05 exceeds the threshold 5e-06.</td>
</tr>
<tr>
<td>Switch CRC error alarm detection</td>
<td rowspan="3"><code>display alarm active</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Parse the alarm information of error code <code>0x081300BC</code>, and identify the port with rapidly increasing CRC errors and its peer device/port.</td>
<td>Port 10GE1/0/1 CRC rapid growth alarm count 1500, threshold 1000, peer device SWITCH-02, peer port 10GE1/0/2</td>
</tr>
<tr>
<td>Switch port lane reduction detection</td>
<td>Parse the alarm information of error code <code>0xF10509</code> to identify the port where lane reduction occurred and its peer port information.</td>
<td>Port 10GE1/0/1 experienced lane reduction, peer port information: 10GE1/0/2</td>
</tr>
<tr>
<td>Switch port Los alarm detection</td>
<td>Parse the alarm information of error code <code>0x8130059</code>, and identify the port and cause of the optical module link Los alarm.</td>
<td>Optical module link Los alarm, cause: received signal loss.</td>
</tr>
<tr>
<td>Optical module Flag (transmit/receive optical indicator) status detection</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check the Flag field value of the optical module transmit/receive optical indicators. The expected value is Normal. Abnormal items are output by lane.</td>
<td>Port 10GE1/0/1 flag information abnormal: lane2 transmit/receive optical indicator Flag value abnormal: 0x03, expected: Normal</td>
</tr>
<tr>
<td>Optical module Datapath State (channel status) detection</td>
<td>Check the Datapath State field value of the optical module channel status. The expected value is Active. Abnormal items are output by lane.</td>
<td>Port 10GE1/0/1 flag information abnormal: lane2 Datapath State abnormal: Inactive, expected: Active</td>
</tr>
<tr>
<td>Optical module Module State (power mode) detection</td>
<td>Check the Module State field value of the optical module power mode. The expected value is Ready.</td>
<td>Port 10GE1/0/1 flag information abnormal: power mode Module State abnormal: PowerDown, expected: Ready</td>
</tr>
<tr>
<td>Single-ended optical module optical power detection</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the optical module TX/RX power values exceed the threshold range.</td>
<td>Optical module optical power abnormality: RX power -19.0 dBm is below the threshold of -15 dBm</td>
</tr>
<tr>
<td>Single-ended optical module SNR detection</td>
<td>Check whether the optical module Host SNR/Media SNR values are below the threshold.</td>
<td>Local SNR abnormality, SNR value 6.5 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Single-ended optical module current detection</td>
<td>Check whether the optical module bias current exceeds the threshold range.</td>
<td>Local current abnormality: bias current 78 mA is below the threshold of 90 mA</td>
</tr>
<tr>
<td>Optical module SNR inter-lane difference detection</td>
<td><code>dis optical-module interface {interface}</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the Host SNR/Media SNR difference between different lanes of the same port exceeds the threshold. Lanes with abnormal SNR are troubleshot first</td>
<td>Optical module SNR inter-lane difference abnormal: the difference between LANE0 and LANE3 is 4.2 dB</td>
</tr>
<tr>
<td>Dual-ended optical module optical power detection</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code> (switch), <code>hccn_tool -i {chip_phy_id} -optical -g</code> (host)</td>
<td rowspan="3">Switch <code>switch_cli_output.txt</code>, host <code>hccn_log/optical.log</code></td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the TX/RX power of the dual-ended optical module.</td>
<td>Optical module optical power abnormality: local switch RX power -18.5 dBm, peer host TX power -12.0 dBm</td>
</tr>
<tr>
<td>Dual-ended optical module SNR detection</td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the SNR of the dual-ended optical module.</td>
<td>The local switch SNR value is 7.2 dB, and the peer Host SNR value is 9.5 dB</td>
</tr>
<tr>
<td>Dual-ended optical module current detection</td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the bias current of the dual-ended optical modules.</td>
<td>The local switch bias current is 8 mA, and the peer host bias current is 105 mA</td>
</tr>
<tr>
<td>Port QoS credit insufficiency detection</td>
<td><code>dis forward information enp slot {slot_id} chip {chip_id} "get oda table qos credit current" | no-more</code> (For <term>Ascend 950PR&950DT products</term>, this command is executed after redirecting to the NPU slot through PoDManager.)</td>
<td>Not supported</td>
<td>Check whether the current available credit (current credit) of each VL priority is 0 and the allocated credit (alloc credit) is non-zero (alloc credit = used credit + current credit; if the allocated credit is 0, the priority is not enabled, which is not considered an exception)</td>
<td>chip15, port0 QoS credit insufficient: vl1 allocated quantity=8, current available quantity=0</td>
</tr>
</tbody>
</table>

## HCCS Related Diagnostics

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Offline Log</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>HCCS Serdes CDR loss-of-lock detection</td>
<td rowspan="2"><code>display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}"</code></td>
<td rowspan="2"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>When hilink_type is 4, check whether the CDR is out of lock (<code>cdr_los == "1"</code>).</td>
<td>Switch chip: 0, port: 24, CDR loss of lock</td>
</tr>
<tr>
<td>HCCS Serdes power fault detection</td>
<td>When hilink_type is 1, check whether the power fault code <code>csr119_data</code> starts with <code>0x380</code>.</td>
<td>Switch chip: 0, port: 24, power fault, fault code: 0x3802</td>
</tr>
<tr>
<td>HCCS port SNR detection (source)</td>
<td><code>display interface hilink snr</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the SNR value of the abnormal lane in the switch port hilink SNR is below the threshold.</td>
<td>lane0 SNR value 6.8 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>HCCS port SNR detection (destination)</td>
<td><code>display for info enp s 1 c {chip_id} "get port snr port-id {port_id}"</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the SNR of the HCCS switch chip port is below the threshold, and include the peer XPU (CPU/NPU) port information.</td>
<td>Peer NPU0 port, lane 0 SNR value 6.8 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (timeout statistics)</td>
<td rowspan="8"><code>display hccs proxy response statistics</code>, <code>display hccs proxy response detail interface {interface}</code></td>
<td rowspan="8"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Parse the RP TX timeout statistics and outputs the number of timeouts.</td>
<td>HCCS RP TX timeout, number of timeouts: 3</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (local port down for a long time)</td>
<td>If an RP TX timeout occurs and the local port has been down for a long time, a local port fault is determined</td>
<td>Switch port down for a long time.</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (local port flapping)</td>
<td>If an RP TX timeout occurs and the local port flaps, a local port fault is determined.</td>
<td>Switch port flapping</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (local RP packet accumulation)</td>
<td>If an RP TX timeout occurs and packets accumulate in the RP direction of the local port, a local port fault is determined.</td>
<td>RP packet accumulation</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (local VOQ packet accumulation)</td>
<td>If an RP TX timeout occurs and packets accumulate in the VOQ of the local port, a local port fault is determined.</td>
<td>VOQ packet accumulation</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (route miss in the peer LP direction)</td>
<td>If an RP TX timeout occurs and a route miss occurs in the LP direction of the peer port, a peer port fault is determined.</td>
<td>LP direction route miss</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (peer LP packet accumulation)</td>
<td>If an RP TX timeout occurs and packets accumulate in the LP of the peer port, a peer port fault is determined.</td>
<td>PL packet accumulation</td>
</tr>
<tr>
<td>HCCS RP TX timeout detection (peer VOQ packet accumulation)</td>
<td>If an RP TX timeout occurs and packets accumulate in the VOQ of the peer port, a peer port fault is determined.</td>
<td>VOQ packet accumulation</td>
</tr>
<tr>
<td>HCCS RX timeout detection (timeout statistics)</td>
<td rowspan="5"><code>display hccs proxy response statistics</code>, <code>display interface</code>, <code>display for info enp s 1 c {chip_id} "get port link start 0 end 47"</code></td>
<td rowspan="5"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Filter the interfaces where RX timeouts occur (<code>rp_rx</code> or <code>lp_tx</code> timeout), and outputs the number of timeouts</td>
<td>HCCS RX timeout, number of timeouts: 3</td>
</tr>
<tr>
<td>HCCS RX timeout detection (port down for a long time)</td>
<td>If an RX timeout occurs and the port has been down for a long time, a port fault is determined.</td>
<td>Switch port down for a long time</td>
</tr>
<tr>
<td>HCCS RX timeout detection (port flapping)</td>
<td>If an RX timeout occurs and the port flaps, a port fault is determined.</td>
<td>Switch port flapping</td>
</tr>
<tr>
<td>HCCS RX timeout detection (link lane reduction)</td>
<td>If an RX timeout occurs and the link is degraded to fewer lanes, a link fault is determined.</td>
<td>Switch link degraded to fewer lanes</td>
</tr>
<tr>
<td>HCCS RX timeout detection (XPU device exception)</td>
<td>If an RX timeout occurs and no port or link exception is found, the peer XPU (CPU/NPU) device is determined to be abnormal.</td>
<td>xPU device exception</td>
</tr>
</tbody>
</table>

## General Diagnostics

<table>
<thead>
<tr>
<th>Diagnostic Item</th>
<th>Online Command</th>
<th>Offline Log</th>
<th>Diagnostic Logic</th>
<th>Abnormal Output</th>
</tr>
</thead>
<tbody>
<tr>
<td>Host-side inter-lane power difference detection</td>
<td><code>hccn_tool -i {chip_phy_id} -optical -g</code></td>
<td><code>hccn_log/optical.log</code></td>
<td>Calculate the difference between the maximum and minimum TX/RX power between different lanes of the same port, and determines whether it exceeds the threshold (3 dB). If it is abnormal, troubleshoot the optical module and check the port.</td>
<td>The difference between the maximum and minimum TX port lane values exceeds 3 dB. Actual maximum lane0: -12.0 dBm, minimum lane3: -16.0 dBm</td>
</tr>
<tr>
<td>Switch-side inter-lane power difference detection</td>
<td><code>dis optical-module interface {interface}</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Calculate the difference between the maximum and minimum TX/RX power between different lanes of the same port, and determines whether it exceeds the threshold (3 dB). The threshold on the switch side is selected based on the port rate: 800G ports use the 800G optical module threshold, and other ports use the default threshold.</td>
<td>Switch port 10GE1/0/1 TX inter-lane difference greater than 3 dB, actual maximum lane0: -12.0 dBm, minimum lane3: -16.0 dBm</td>
</tr>
<tr>
<td>BMC-side inter-lane power difference detection</td>
<td><code>ipmcget -t sensor -d list</code></td>
<td><code>dump_info/AppDump/sensor/sensor_info.txt</code></td>
<td>Calculate the difference between the maximum and minimum TX/RX power between different lanes of the same port, and determines whether it exceeds the threshold (3 dB). If it is abnormal, troubleshoot the optical module and check the port.</td>
<td>BMC NPU0 port TX inter-lane difference greater than 3 dB, actual maximum lane0: -12.0 dBm, minimum lane3: -16.0 dBm</td>
</tr>
</tbody>
</table>
