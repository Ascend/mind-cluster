# Diagnostic Items

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:18:11.615Z pushedAt=2026-08-24T02:22:41.579Z -->

Diagnostic items are categorized by device type, and each diagnostic item specifies its online command source and offline log path. The same device may involve multiple diagnostic items, and each diagnostic item corresponds to one or more online commands and offline logs. The toolkit includes a total of **40** diagnostic items, covering the host, BMC, switch, HCCS, and general dimensions.

## Diagnostic Item Index

| Device Type | Number of Diagnostic Items | Detailed Description |
|----------|------------|----------|
| Host-side Diagnostics | 18 | [Host-side Diagnostics](#host-side-diagnostics) |
| BMC Related Diagnostics | 2 | [BMC Related Diagnostics](#bmc-related-diagnostics) |
| Switch-side Diagnostics | 11 | [Switch-side Diagnostics](#switch-side-diagnostics) |
| HCCS Related Diagnostics | 8 | [HCCS Related Diagnostics](#hccs-related-diagnostics) |
| General Diagnostics | 1 | [General Diagnostics](#general-diagnostics) |

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
<td>Optical Module Communication Link Fault Detection</td>
<td rowspan="7"><code>hccn_tool -i {chip_phy_id} -optical -g</code></td>
<td rowspan="7">V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code></td>
<td>Check whether the optical module control link is reachable</td>
<td>Optical module: Control link unreachable</td>
</tr>
<tr>
<td>Optical Module Presence Detection</td>
<td>Check whether the optical module is present</td>
<td>Optical module not present, status: NA</td>
</tr>
<tr>
<td>High-power Enable Register Status Detection</td>
<td>Check the high power enable register status</td>
<td>The optical module is in low power mode, high power enable reg:0x00</td>
</tr>
<tr>
<td>Single-ended Optical Module Optical Power Detection</td>
<td>Check whether the TX/RX power values exceed the threshold range</td>
<td>Optical module optical power abnormality, RX power -18.5 dBm; below the threshold of -15 dBm</td>
</tr>
<tr>
<td>Single-ended Optical Module SNR Detection</td>
<td>Check whether the Host SNR/Media SNR values are below the threshold</td>
<td>Optical module SNR abnormal: lane0: Host SNR value 7.2 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical Module SNR Inter-lane Difference</td>
<td>Check whether the SNR difference between different lanes exceeds the threshold</td>
<td>Optical module SNR inter-lane difference abnormal: the difference between LANE0 and LANE3 is 4.2 dB</td>
</tr>
<tr>
<td>Optical Module Los/LoL Detection</td>
<td>Check whether the Rx Los, Tx Los, Rx LoL, and Tx LoL status values are greater than 0</td>
<td>Optical module Rx Los indicator abnormal, status: 1</td>
</tr>
<tr>
<td>Optical Module uncorr_cw_cnt Detection</td>
<td rowspan="2">Automatic collection by <code>msnpureport</code></td>
<td rowspan="2">V1-V3: timestamp subdirectory under the <code>msnpureport</code> report directory</td>
<td>Check whether uncorr_cw_cnt is greater than 10 for 3 consecutive times</td>
<td>uncorr_cw_cnt > 10 occurred for 3 consecutive times, occurrence time: 2025-10-01-14:30:00.123456, 2025-10-01-14:30:01.234567, 2025-10-01-14:30:02.345678</td>
</tr>
<tr>
<td>Optical Module IIC Communication Fault Detection</td>
<td>Detect IIC communication abnormal events</td>
<td>IIC abnormality detected: trans status[0x40], error status[0x10], the NPU onboard optical module adapter may be faulty</td>
</tr>
<tr>
<td>Optical module initialization light-on status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -dfx_cfg -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>Check whether the TX Disable status is disabled</td>
<td>The optical module is in laser-off state, tx disable status: 1</td>
</tr>
<tr>
<td>Optical Module CDR SNR Detection</td>
<td><code>hccn_tool -i {chip_phy_id} -cdr_snr -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>Check whether the CDR Host SNR / Media SNR values are below the threshold</td>
<td>CDR SNR abnormal, Host SNR value 6.8 dB, below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Optical module port status, network health status, and connection status detection</td>
<td><code>hccn_tool -i {chip_phy_id} -link -g</code>, <code>hccn_tool -i {chip_phy_id} -net_health -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code> (net_health not supported)</td>
<td>Check whether the network health status and connection status of the NPU port deviate from normal thresholds, and include the peer switch and port information.</td>
<td>Port optical module status is abnormal, network health status: abnormal, connection status: down. Peer switch: SWITCH-01, peer port: 10GE1/0/1.</td>
</tr>
<tr>
<td>RoCE Port Configuration Detection</td>
<td><code>hccn_tool -i {chip_phy_id} -speed -g</code>, <code>hccn_tool -i {chip_phy_id} -duplex -g</code>, <code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/net_conf.log</code><br>V3: <code>optical.log</code>/<code>lldp.log</code> (duplex not supported)</td>
<td>Locate the peer switch port through LLDP information, and compare whether the speed and duplex mode on both ends are consistent (no alarm is raised when either end is auto)</td>
<td>The connection information of the NPU port and the peer switch: SWITCH-01, ip: 0.0.0.1, port 10GE1/0/1 is inconsistent. Local Speed: 100G, Duplex: full. Peer Speed: 50G, Duplex: full</td>
</tr>
<tr>
<td>NPU Peer LLDP Information Missing Detection</td>
<td><code>hccn_tool -i {chip_phy_id} -lldp -g</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>lldp.log</code></td>
<td>Check whether the lldp information of the peer of the NPU optical module is collected</td>
<td>The lldp information of the peer of the NPU optical module is not collected</td>
</tr>
<tr>
<td>Loopback Detection</td>
<td><code>hccn_tool -i {npu_id} -optical -t {model}</code></td>
<td>V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: Not Supported</td>
<td>Determine the fault location based on the loopback test status code: if the port is down after loopback type 1 is enabled, it is determined as a local fault; if the port is up after loopback type 1 is enabled but down after loopback type 2 is enabled, it is determined as an optical module fault/contamination on the local port.</td>
<td>The local port is down after loopback type 1 is enabled, diagnosed as a local fault.</td>
</tr>
<tr>
<td>Dual-ended Optical Module Optical Power Detection</td>
<td rowspan="3"><code>hccn_tool -i {chip_phy_id} -optical -g</code> (host), <code>dis optical-module interface {interface}</code> (switch), <code>hccn_tool -i {chip_phy_id} -lldp -g</code> (host)</td>
<td rowspan="3">Host: V1: <code>hccn_tool.log</code><br>V2: <code>hccn_log/optical.log</code><br>V3: <code>optical.log</code>; Switch: <code>switch_cli_output.txt</code></td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the TX/RX power of the dual-ended optical module.</td>
<td>Optical module optical power abnormality: local RX power -18.5 dBm is below the threshold -15dBm, while the peer switch TX power -12.0 dBm is normal.</td>
</tr>
<tr>
<td>Dual-ended Optical Module SNR Detection</td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the Host SNR / Media SNR of the dual-ended optical module.</td>
<td>Local SNR value 7.2 dB is below the threshold of 8.0 dB, while the peer switch SNR value 9.5 dB is normal.</td>
</tr>
<tr>
<td>Dual-ended Optical Module Current Detection</td>
<td>Obtain the peer port through LLDP information, and perform comparative analysis on the bias current of the dual-ended optical module</td>
<td>The local bias current is 85 mA, below the threshold of 90mA, while the peer switch bias current is 105 mA, which is normal</td>
</tr>
</tbody>
</table>

>[!NOTE]
> For details about host log V1/V2/V3 versions, see [Offline Log Collection (Host)](../05_usage/02_log_collection.md#host-offline-log).

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
<td>BMC Alarm Fault Code Detection</td>
<td><code>ipmcget -d sel -v list</code></td>
<td><code>dump_info/AppDump/event/sel.txt</code></td>
<td>Parse the event codes in the BMC SEL log, match abnormal descriptions by error code and keyword, and output handling suggestions.</td>
<td>See the alarm code list below for details</td>
</tr>
<tr>
<td>BMC Optical Module History Information Detection</td>
<td><code>ipmcget -t sensor -d list</code></td>
<td><code>dump_info/AppDump/sensor/sensor_info.txt</code>, <code>dump_info/AppDump/network_adapter/optical_module/optical_module_history_info_log.csv</code></td>
<td>When optical module history information is detected, the linkdown time is recorded (which may indicate a transient interruption or hardware fault), and the optical power, bias current, Host SNR/Media SNR are checked per lane to determine whether they exceed thresholds, and whether the Tx Los/Rx Los status values are greater than 0</td>
<td>NPU linkdown detected, recorded at 2025-10-01 14:30:00, possibly caused by intermittent disconnection or hardware fault; lane0: RX power -18.5 dBm is below the threshold -15dBm; Tx los value 0x1 is greater than 0</td>
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
<td>A multi-bit ECC fault has occurred</td>
<td>Please run a stress test on the NPU on-chip memory</td>
</tr>
<tr>
<td>HBM ECC</td>
<td><code>0x80e18402</code></td>
<td>Multi-bit ECC fault, isolated rows have reached 64</td>
<td>Please replace the NPU spare part immediately</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb800a</code></td>
<td>AIV operator timeout, NPU warm reset</td>
<td>It is recommended to perform AICode stress testing on the hardware</td>
</tr>
<tr>
<td>AIV</td>
<td><code>0x80cb8009</code></td>
<td>AIV bus access error</td>
<td>It is recommended to run AICode stress testing on the hardware.</td>
</tr>
<tr>
<td>NPU health</td>
<td><code>0x56000003</code></td>
<td>NPU health status critical alarm</td>
<td>1. Check whether the chip temperature is too high (possibly due to abnormal heat dissipation, excessive ambient temperature, or blocked air inlet/outlet); 2. Check whether it is a software issue such as an address exception or memory leak; 3. If the issue cannot be resolved, contact technical support.</td>
</tr>
<tr>
<td>NPU Card Loss</td>
<td><code>0x56000005</code></td>
<td>NPU connection has been lost alarm</td>
<td>A card loss fault has occurred. Please contact O&M personnel for handling.</td>
</tr>
<tr>
<td>NPU Overheating</td>
<td><code>0x56000009</code></td>
<td>NPU Overheating Shutdown</td>
<td>NPU overheating shutdown. Please contact O&M personnel.</td>
</tr>
<tr>
<td>Abnormal Power-off</td>
<td><code>0x2C000007</code></td>
<td>AI module abnormal power-off alarm/system abnormal power-off alarm/NPU abnormal power-off alarm (subdivided by voltage link identifier: <code>V_1V2_DVDD_HBM02_FIX</code>, <code>V_0V9_AIC_DVFS_DA</code>, <code>V_AVDD12_HVCC</code>, <code>V_AVDD08_LVCC</code>, <code>V_0V8_DVDD_SIOE</code>)</td>
<td>AC restart; if not recovered, replace the corresponding module; or 20A PSIP fault/12V capacitor failure, contact O&M for handling</td>
</tr>
<tr>
<td>Power-on Timeout</td>
<td><code>0x2C00002B</code></td>
<td>Power-on timeout alarm, voltage drop on the mainboard (<code>V_AVDD12_HVCC</code>)</td>
<td>1. Check whether the external power supply meets the overall power consumption requirements of the server; 2. Completely power off and then power on by unplugging the power cable or removing the board, and check whether the alarm is cleared; 3. If the issue persists, contact technical support to replace the possibly involved components</td>
</tr>
<tr>
<td>Power-on Timeout</td>
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
<td>PSU Over-temperature</td>
<td><code>0x5D000005</code></td>
<td>PSU over-temperature alarm</td>
<td>The PSU temperature is too high. Please contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid Cooling</td>
<td><code>0x12000023</code></td>
<td>Liquid cooling device leakage</td>
<td>Liquid cooling device leakage occurred. Please contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Liquid cooling (LAAC)</td>
<td><code>0x120000C3</code></td>
<td>Liquid cooling device (LAAC) abnormal, liquid cooling pump not present</td>
<td>Liquid cooling pump not present. Contact O&M personnel for handling.</td>
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
<td>Fan redundancy failure. Please contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x04000007</code></td>
<td>Air cooling module fault, large fan speed deviation</td>
<td>The fan speed deviation is large. Please contact O&M personnel for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x18000003</code></td>
<td>Air cooling module fault, fan backplane power supply fault</td>
<td>Fan backplane power supply fault. Please contact O&M for handling.</td>
</tr>
<tr>
<td>Fan</td>
<td><code>0x1800000D</code></td>
<td>Air cooling module fault, fan backplane MCU self-check abnormal</td>
<td>Fan backplane MCU self-check abnormal, contact O&M for handling</td>
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
<td>Switch Port BER Detection</td>
<td><code>display interface troubleshooting | no-more</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the port BER exceeds the threshold <code>5.0e-06</code></td>
<td>BER bit error rate 1.2e-05 exceeds the threshold 5e-06.</td>
</tr>
<tr>
<td>Switch CRC Error Alarm Detection</td>
<td rowspan="3"><code>display alarm active</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Parse the alarm information of error code <code>0x081300BC</code>, and identify the port with rapidly increasing CRC errors and its peer device/port.</td>
<td>Port 10GE1/0/1 CRC rapid growth alarm count 1500, threshold 1000, peer device SWITCH-02, peer port 10GE1/0/2</td>
</tr>
<tr>
<td>Switch Port Lane Reduction Detection</td>
<td>Parse the alarm information of error code <code>0xF10509</code> to identify the port where lane reduction occurred and its peer port information</td>
<td>Port 10GE1/0/1 experienced lane reduction, peer port information: 10GE1/0/2</td>
</tr>
<tr>
<td>Switch Port Los Alarm Detection</td>
<td>Parse the alarm information of error code <code>0x8130059</code>, and identify the port and cause of the optical module link Los alarm.</td>
<td>Optical module link Los alarm, cause: received signal loss.</td>
</tr>
<tr>
<td>Optical module status detection (State-flag/Datapath State/Module State)</td>
<td><code>display interface transceiver verbose</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check the optical module Flag (transmit/receive optical indicator, expected Normal), Datapath State (channel state, expected Active), and Module State (power mode, expected Ready) field values, and output abnormal items by lane</td>
<td>Port 10GE1/0/1 flag information abnormal: lane2 transmit/receive optical indicator Flag value abnormal: 0x03, expected: Normal</td>
</tr>
<tr>
<td>Single-ended Optical Module Optical Power Detection</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code></td>
<td rowspan="3"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the optical module TX/RX power values exceed the threshold range</td>
<td>Optical module optical power abnormality: RX power -19.0 dBm is below the threshold of -15 dBm</td>
</tr>
<tr>
<td>Single-ended optical module SNR detection</td>
<td>Check whether the optical module Host SNR / Media SNR values are below the threshold</td>
<td>Local SNR abnormality, SNR value 6.5 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>Single-ended Optical Module Current Detection</td>
<td>Check whether the optical module bias current exceeds the threshold range</td>
<td>Local current abnormality: bias current 78 mA is below the threshold of 90 mA</td>
</tr>
<tr>
<td>Dual-ended Optical Module Optical Power Detection</td>
<td rowspan="3"><code>dis optical-module interface {interface}</code> (switch), <code>hccn_tool -i {chip_phy_id} -optical -g</code> (host)</td>
<td rowspan="3">Switch <code>switch_cli_output.txt</code>, host <code>hccn_log/optical.log</code></td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the TX/RX power of the dual-ended optical module</td>
<td>Optical module optical power abnormality: local switch RX power -18.5 dBm, peer host TX power -12.0 dBm</td>
</tr>
<tr>
<td>Dual-ended Optical Module SNR Detection</td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the SNR of the dual-ended optical module</td>
<td>The local switch SNR value is 7.2 dB, and the peer host SNR value is 9.5 dB</td>
</tr>
<tr>
<td>Dual-ended Optical Module Current Detection</td>
<td>Obtain the peer through the port mapping relationship, and perform comparative analysis on the bias current of the dual-ended optical modules</td>
<td>The local switch bias current is 8 mA, and the peer host bias current is 105 mA</td>
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
<td>HCCS Serdes Detection</td>
<td><code>display for info enp s 1 c {chip_id} "get port serdes dump-info macro-id {port_id} lane-id {lane_id} hilink {type}"</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>1. hilink_type is 4. Check whether the CDR is out of lock (<code>cdr_los == "1"</code>)<br>2. hilink_type is 1. Check whether the power supply fault code <code>csr119_data</code> starts with <code>0x380</code></td>
<td>Switch chip: 0, port: 24, CDR loss of lock detected, power supply fault detected, fault code: 0x3802</td>
</tr>
<tr>
<td>HCCS Port SNR Detection (Source End)</td>
<td><code>display interface hilink snr</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the SNR value of the abnormal lane in the switch port hilink SNR is below the threshold</td>
<td>lane0 SNR value 6.8 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>HCCS Port SNR Detection (Destination)</td>
<td><code>display interface hilink snr</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Check whether the SNR of the HCCS switch chip port is below the threshold, and include the peer XPU (CPU/NPU) port information</td>
<td>Peer NPU0 port, lane 0 SNR value 6.8 dB is below the threshold of 8.0 dB</td>
</tr>
<tr>
<td>HCCS RP TX Timeout Detection (Local)</td>
<td rowspan="2"><code>display hccs proxy response statistics</code>, <code>display hccs proxy response detail interface {interface}</code></td>
<td rowspan="2"><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>When an RP TX timeout occurs, determine the local port fault in combination with conditions such as the port being down for a long time, intermittent port disconnection, and packet congestion (RP/VOQ)</td>
<td>Switch port down for a long time; or switch port intermittent disconnection; or RP packet congestion</td>
</tr>
<tr>
<td>HCCS RP TX Timeout Detection (Peer)</td>
<td>When an RP TX timeout occurs, determine the peer port fault based on conditions such as route miss in the LP direction, LP packet congestion, and VOQ packet congestion on the peer port</td>
<td>LP direction route miss</td>
</tr>
<tr>
<td>HCCS RX Timeout Detection</td>
<td><code>display hccs proxy response statistics</code>, <code>display interface</code>, <code>display for info enp s 1 c {chip_id} "get port link start 0 end 47"</code></td>
<td><code>switch_cli_output.txt</code>, <code>diag_info.txt</code></td>
<td>Filter the interfaces where RX timeout occurs (<code>rp_rx</code> or <code>lp_tx</code> timeout), and determine the fault cause in combination with long-term port down, port flapping, link lane degradation, XPU device abnormality, and other conditions</td>
<td>Switch port down for a long time; or switch port flapping; or switch link lane degradation; or XPU device abnormality</td>
</tr>
<tr>
<td>HCCS Link Degradation Detection (Port Level)</td>
<td rowspan="2"><code>ipmcget -d healthevents</code></td>
<td rowspan="2"><code>dump_info/AppDump/event/current_event.txt</code></td>
<td>Parse the event description of error code <code>0x28000049</code> to locate the fault between the CPU board UBC/macro and the L1 port</td>
<td>A fault occurred between Cpu0 UBC0 macro0 CPU board 0 and the L1 port</td>
</tr>
<tr>
<td>HCCS Link Degradation Detection (Board Level)</td>
<td>When all CPU ports of an L1 switch chip are abnormal, it is determined as a board-level fault.</td>
<td>All ports of the L1 switch chip are abnormal.</td>
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
<td>Inter-lane Power Difference Detection</td>
<td><code>hccn_tool -i {chip_phy_id} -optical -g</code> (host), <code>dis optical-module interface {interface}</code> (switch), <code>ipmcget -t sensor -d list</code> (BMC)</td>
<td>Host <code>hccn_log/optical.log</code>, switch <code>switch_cli_output.txt</code>, BMC <code>dump_info/AppDump/sensor/sensor_info.txt</code></td>
<td>Calculate the difference between the maximum and minimum TX/RX power values across different lanes of the same port, and determine whether it exceeds the threshold (3dB). Supports detection from three dimensions: host, switch, and BMC.</td>
<td>The difference between the maximum and minimum TX port lane values exceeds 3 dB. Actual maximum lane0: -12.0 dBm, minimum lane3: -16.0 dBm</td>
</tr>
</tbody>
</table>
