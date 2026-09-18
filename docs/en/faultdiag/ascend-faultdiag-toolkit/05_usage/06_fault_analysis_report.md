# Diagnostic/Inspection Report Description

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:54.068Z pushedAt=2026-08-24T02:22:41.549Z -->

This section describes the output location, file format, and field meanings of the ascend-fd-tk diagnostic report and inspection report, for interpreting diagnostic/inspection results.

## Report Description

| Report Item | Description |
|-------|--------------------------------------------------------------------------------------------------------------------|
| Report directory | Linux: `~/.ascend-faultdiag-toolkit/report/`<br>Windows: `{current_working_directory}/.ascend-faultdiag-toolkit/report/` |
| Report file name | Diagnostic report: `diag_report_{YYYYMMDD_HHMMSS}.xlsx` (for example, `diag_report_20260629_143022.xlsx`)<br/>Inspection report: `inspection_errors.csv` |
| Report format | Diagnostic report: Microsoft Excel (`.xlsx`), containing multiple sheets <br/> Inspection report: CSV format |
| Report character set | UTF-8 (Chinese descriptions display normally) |

**Quick Navigation**

- [Diagnostic Report](#diagnostic-report)
- [Inspection Report](#inspection-report)

<a id="Diagnostic Report"></a>

## Diagnostic Report

### Report Structure

A diagnostic report contains at most **7 sheets**, which are divided into three categories: link analysis, fault diagnosis, and optical module information. Each category of sheets is distinguished by the sheet label color. The sheets are independent of each other and can be opened, filtered, and sorted separately. When there is no data, the corresponding sheet is not generated.

<table>
  <thead>
    <tr>
      <th>Category</th>
      <th>Sheet Name</th>
      <th>Sheet Label Color</th>
      <th>Content</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan="2">Link analysis</td>
      <td>NPU&CPU to L1 link analysis</td>
      <td rowspan="2" style="background-color:#F4B183;color:#000;">Light orange (#F4B183)</td>
      <td>Downlink mapping relationship from NPU/CPU to the L1 switching fabric and link status analysis</td>
    </tr>
    <tr>
      <td>L1-to-L2 link analysis</td>
      <td>Uplink mapping relationship from L1 switching fabric to L2 bus devices and link status analysis</td>
    </tr>
    <tr>
      <td rowspan="3">Fault diagnosis</td>
      <td>In-band fault analysis (host)</td>
      <td rowspan="3" style="background-color:#9DC3E6;color:#000;">Light blue (#9DC3E6)</td>
      <td>All diagnostic exceptions on the host side</td>
    </tr>
    <tr>
      <td>Out-of-band fault analysis (BMC)</td>
      <td>All diagnostic exceptions on the BMC side</td>
    </tr>
    <tr>
      <td>Switch fault analysis (L1&L2&RoCE)</td>
      <td>All diagnostic exceptions on the switch side</td>
    </tr>
    <tr>
      <td rowspan="2">Optical module information</td>
      <td>Optical module information for host NPU&lt;-&gt;RoCE switch ports</td>
      <td rowspan="2" style="background-color:#A9D18E;color:#000;">Light green (#A9D18E)</td>
      <td>Optical module status between host NPU and RoCE switch ports</td>
    </tr>
    <tr>
      <td>Optical module information for inter-switch port connections</td>
      <td>Status of optical modules connected to ports between switches</td>
    </tr>
  </tbody>
</table>

In addition to the sheet label colors, some cells within the report also use colors to indicate fault severity.

<table>
  <thead>
    <tr>
      <th>Color</th>
      <th>Meaning</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td style="background-color:#F8D7DA;color:#000;">Light red (#F8D7DA)</td>
      <td>Fault/abnormality exists and requires priority handling</td>
    </tr>
    <tr>
      <td style="background-color:#FFF3CD;color:#000;">Light yellow (#FFF3CD)</td>
      <td>Alarm/attention required, troubleshooting recommended</td>
    </tr>
    <tr>
      <td style="background-color:#E6F9E6;color:#000;">Light green (#E6F9E6)</td>
      <td>Normal/passed, no handling required</td>
    </tr>
  </tbody>
</table>

### Fault Diagnosis Sheets

#### In-band Fault Analysis (Host) Sheet

| Field | Description |
|------|------|
| Host ID | Unique identifier of the host within the tool |
| Host Name | Output of the host `hostname` command |
| SN | Host serial number |
| Equipment Room Name | Equipment room name |
| Cabinet Number | Cabinet location number |
| NPU ID | NPU number |
| Physical Chip ID | Physical chip ID (`chip_phy_id`) |
| Fault Code | Fault code defined internally by the tool (empty for some diagnostic items) |
| Fault Information | Detailed description of the fault (including data values and thresholds) |
| Handling Suggestion | Recommended handling method |
| Link Fault Root Cause | Whether it is the root cause of this diagnosis (`yes`/`no`/`unknown`) |

#### Out-of-band Fault Analysis (BMC) Sheet

| Field | Description |
|------|------|
| BMC ID | BMC IP address |
| SN | BMC serial number |
| NPU ID | NPU number |
| Physical Chip ID | Physical chip ID |
| Fault Code/Fault Information/Handling Suggestion/Link Fault Root Cause | Same as the Host sheet |

#### Switch Fault Analysis (L1&L2&RoCE) Sheet

| Field | Description |
|------|--------------|
| Switch Name | Switch `sysname` |
| Switch ID | Switch IP address |
| SN | Switch serial number |
| Equipment Room Name | Equipment room name |
| Cabinet Number | Cabinet location number |
| Port | Switch port name |
| Fault Code/Fault Information/Handling Suggestion/Link Fault Root Cause | Same as the Host sheet |

### Link Analysis Sheet

#### NPU&CPU to L1 Link Analysis Sheet

| Field | Meaning               |
|------|------------------|
| Host ID | Unique identifier of the host within the tool     |
| Host Name | Host hostname      |
| Equipment Room Name | Equipment room name           |
| Cabinet Number | Cabinet location number           |
| NPU/CPU Type | Device type (NPU or CPU)  |
| NPU/CPU ID | Number of the NPU or CPU    |
| Physical Chip ID | Physical chip ID          |
| L1 Switch IP | L1 switch IP address     |
| L1 Switch Name | L1 switch name         |
| L1 Switch Board ID | L1 switching chip ID       |
| L1 Switch Board Physical Port Number | Physical port number on the L1 switching chip   |
| L1 Port | L1 switch port name        |
| hilink SNR | hilink signal-to-noise ratio       |
| Link Status | Link normal/abnormal status      |
| Link Analysis | Link fault analysis result |

#### L1-to-L2 Link Analysis Sheet

| Field | Description              |
|------|-----------------|
| L1 Switch IP | IP address of the L1 switch    |
| L1 Switch Name | Name of the L1 switch        |
| L1 Equipment Room Name | Name of the L1 equipment room         |
| L1 Cabinet Number | Cabinet location number of L1       |
| L1 Switching Chip ID | ID of the L1 switching chip      |
| L1 Switching Chip Physical Port Number | Physical port number on the L1 switching chip  |
| L1 Port | Port name of the L1 switch       |
| L1 hilink SNR | hilink signal-to-noise ratio on the L1 side |
| L1 host SNR | Host SNR on the L1 side   |
| L1 media SNR | Media SNR on the L1 side  |
| L2 Switch IP | IP address of the L2 switch    |
| L2 Switch Name | Name of the L2 switch        |
| L2 Equipment Room Name | Name of the L2 equipment room         |
| L2 Cabinet Number | Cabinet location number of L2       |
| L2 Port | Port name of the L2 switch       |
| L2 hilink SNR | hilink signal-to-noise ratio on the L2 side |
| L2 host SNR | Host SNR on the L2 side   |
| L2 media SNR | Media SNR on the L2 side  |
| Link Status | Normal/abnormal status of the link     |
| Link Analysis | Link fault analysis result        |

### Optical Module Information Sheet

#### Host NPU\<-\>RoCE Switch Port

This Sheet uses a two-section layout with "host side" + "switch side", and each side contains detailed optical module metrics for 4 lanes (lane 0 to lane 3).

**Host-Side Fields**

| Field Category | Field |
|---------|------|
| Device Information | Host ID, Host Name, Host SN, Equipment Room Name, Cabinet Number |
| Chip Information | NPU ID, Chip ID, Physical Chip ID, NPU Type, Host Side Port |
| Optical Module Basic Information | Host Side Optical Module Status, Optical Module Vendor, Optical Module Model, Optical Module SN, Optical Module Temperature, Optical Module Supply Voltage |
| Link Status | Host Side Link Speed, Duplex Mode, Network Health, Link Status, Link UP Count, Link DOWN Count |
| Lane Metrics (Lane 0~3) | TX Power, RX Power, TX Bias, Host SNR, Media SNR |

**Peer Switch-Side Fields**

| Field Category | Field |
|---------|------|
| Device Information | Peer Switch Name, Peer Switch ID, Peer Switch SN, Equipment Room Name, Cabinet Number |
| Port Information | Peer Switch Port, Port Speed, Port Duplex Mode, Port Status |
| Optical Module Basic Information | Peer Switch Optical Module Vendor, Optical Module Model, Optical Module SN, Optical Module Temperature |
| Lane Metrics (Lane 0~3) | TX Power, RX Power, SNR |

#### Inter-Switch Port Connection

This sheet uses a two-section layout of "local" + "peer", with each side containing optical module metrics for 4 lanes.

**Local Switch Fields**

| Field Category | Field |
|---------|------|
| Device Information | Local Switch Name, Switch ID, Switch SN, Equipment Room Name, Cabinet Number |
| Port Information | Local Port |
| Optical Module Basic Information | Local Optical Module Vendor, Optical Module Model, Optical Module SN, Optical Module Temperature |
| Lane Metrics (Lane 0~3) | TX Power, RX Power, SNR |

**Peer Switch Fields**

| Field Category | Field |
|---------|------|
| Device Information | Peer Switch Name, Switch ID, Switch SN, Equipment Room Name, Cabinet Number |
| Port Information | Peer Port |
| Optical Module Basic Information | Peer Optical Module Vendor, Optical Module Model, Optical Module SN, Optical Module Temperature |
| Lane metrics (Lane 0~3) | TX Power, RX Power, SNR |

### Report Interpretation

#### Step 1: Locate the Key Sheet

Focus by fault type:

| Fault Symptom | Priority Sheet |
|----------|----------|
| Host NPU training/inference exception, packet loss, high latency | In-band fault analysis (host) |
| BMC alarm indicator on, unable to log in to BMC | Out-of-band fault analysis (BMC) |
| Inter-node network unreachable, severe packet loss, intermittent disconnection | Switch fault analysis (L1&L2&RoCE) |
| Equipment-room-level fault (multiple cabinets and nodes) | Switch sheet first, then host/BMC |
| Link mapping relationship verification | NPU&CPU-to-L1 link analysis/L1-to-L2 link analysis |
| Optical module status verification | Host NPU\<-\>RoCE switch port/Inter-switch port connection |

#### Step 2: Filter Root Causes

In the "Fault Diagnosis" Sheet, filter the "Link Fault Root Cause" column for records with the value "yes". These are the root cause faults of this diagnosis identified by the tool.

- Records marked as "No" indicate that the fault may be a derivative of another fault (for example, "port down" is caused by "peer switch power failure"), and can be used as a reference but are not the primary items to handle.
- Records marked as "unknown" indicate that the tool cannot determine whether it is a root cause, and manual judgment based on the business scenario is required.

#### Step 3: Follow the Handling Suggestion

Check the handling suggestion recommended by the tool in the "Handling Suggestion" column.

#### Step 4: Verify Link and Optical Module Information

View the detailed link mapping relationships and optical module metrics in the "Link Analysis" sheet and the "Optical Module Information" sheet, and compare the local/peer data of abnormal ports.

<a id="Inspection Report"></a>

## Inspection Report

After executing the `auto_inspection` command, the tool generates an inspection report `inspection_errors.csv`, which records the optical module anomaly information of the port connections between the local device and the peer device. Unlike the Excel format of the diagnostic report, the inspection report is in **CSV format** and can be opened and viewed with a text editor or spreadsheet software (such as Microsoft Excel or WPS Spreadsheets).

### Inspection Report Fields

| Field | Description |
|------|------|
| A-end Device Name | Name of the local device |
| A-end IP | IP address of the local device |
| A-end Interface | Physical port identifier of the local device |
| A-end SN | Serial number of the local device |
| A-end Optical Module SN | Serial number of the optical module on the local port |
| B-end Device Name | Name of the peer device |
| B-end IP | IP address of the peer device |
| B-end Interface | Physical port identifier of the peer device |
| B-end SN | Serial number of the peer device |
| B-end Optical Module SN | Serial number of the optical module on the peer port |
| Problem Symptom | Description of the anomaly found during inspection (for example, abnormal optical module power, low SNR, abnormal link status, etc.) |

Each row in the inspection report represents a pair of port connections with an anomaly, where the A-end is the local device and the B-end is the peer device, facilitating quick identification of the optical module issues on both devices.

### Report Interpretation

The inspection report is mainly used to quickly identify physical connection anomalies at the optical module level in a cluster. It is recommended to interpret it by following these steps:

1. **Filter high-frequency issues**: Sort or filter by the "Problem Symptom" column to quickly locate the fault types that occur in a concentrated manner.
2. **Locate specific devices**: Locate the abnormal local port based on "A-end Device Name", "A-end IP", and "A-end Interface", and then confirm the peer information by combining "B-end Device Name", "B-end IP", and "B-end Interface".
3. **Verify optical module information**: Use "A-end Optical Module SN" and "B-end Optical Module SN" to query the corresponding optical modules, and replace the optical modules if necessary.
