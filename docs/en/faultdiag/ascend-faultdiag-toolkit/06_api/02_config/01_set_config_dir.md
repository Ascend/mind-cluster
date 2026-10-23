# set_config_dir

## Command Function

Sets the configuration file directory path. The tool automatically scans and loads the configuration files in this directory. Currently, it supports loading the equipment room configuration file `LLD.xlsx` (which contains the UnifiedBus L1/L2 network mapping relationships), used to associate location information such as the cabinets and equipment room in diagnostic reports; it also supports loading the threshold configuration file `threshold_config.json` to override the default diagnosis thresholds.

## Command Format

| Command Format | Description |
|---------|------|
| `set_config_dir <directory_path>` | Sets the configuration file directory path |
| `set_config_dir ?` | Views details |

## Parameter Description

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<directory_path>` | String | Yes | Directory path where the configuration file resides. The directory must contain `LLD.xlsx` and `threshold_config.json`. |

## LLD.xlsx Description

You need to provide the `LLD.xlsx` file. For a sample file, see [LLD.xlsx](../../../../resource/LLD.xlsx). The file must contain two sheets:

| Sheet Name | Mandatory Fields | Purpose |
|----------|--------|------|
| UnifiedBus L1 Network Mapping | Server, Equipment Room Name, Cabinet Number, Host SN, L1 Name, L1_IP, L1_SN | Describes the mapping between hosts and L1 switches |
| UnifiedBus L2 Network Mapping | Device Name, Equipment Room Name, Cabinet Number, Management IP Configuration, SN | Describes the machine room location information of L2 switches |

## threshold_config.json Description

The threshold configuration file `threshold_config.json` is used to override the default thresholds of each diagnosis item, so that the fault diagnosis criteria can be adjusted based on the actual environment. The file must be placed in the configuration directory set by `set_config_dir` and is automatically read when the tool loads the configuration. When the file is not placed, the built-in default thresholds in the code are used. For a complete sample file, see [threshold_config.json](../../../../resource/threshold_config.json).

The configuration file is a JSON object. Each configuration item corresponds to a threshold name, and the threshold fields are uniformly nested in the `threshold` sub-object, for example:

```json
{
    "TX_POWER_MW": {"threshold": {"low_value_alarm": "0.2", "high_value_alarm": "2.5", "desc": "tx power", "unit": "mW"}},
    "HOST_SNR_DB": {"threshold": {"low_value_warn": "20", "low_value_alarm": "18", "desc": "host snr", "unit": "dB"}}
}
```

Threshold field description:

| Field | Type | Description |
|------|-----|------|
| `low_value_alarm` / `high_value_alarm` | String | Low/high value alarm threshold (numeric string) |
| `low_value_warn` / `high_value_warn` | String | Low/high value warning threshold (numeric string) |
| `normal_value_alarm` | String | Normal value determination string (for example, `"UP"` or `"present"`) |
| `desc` | String | Threshold description, 128 characters at most |
| `unit` | String | Unit (for example, `"dB"`, `"mW"`, or `"mA"`), 16 characters at most |

Meaning of threshold configuration items:

| Configuration Item | Meaning | Unit |
|--------|--------|------|
| `CDR_HOST_SNR_DB` | CDR (clock data recovery) host-side SNR | dB |
| `CDR_MEDIA_SNR_DB` | CDR (clock data recovery) media-side SNR | dB |
| `CHIP_CPU_PORT_SNR_LINE` | Port SNR between the CPU and the L1 switch (linear value, about 56 dB) | - |
| `CHIP_NPU_PORT_SNR_LINE` | Port SNR between the NPU and the L1 switch (linear value) | - |
| `SWITCH_PORT_SNR_LINE` | Port SNR between the L1 and L2 switches (linear value) | - |
| `DUPLEX_THRESHOLD` | Port duplex mode, with the normal value `Full` | - |
| `NET_HEALTH_THRESHOLD` | Network health check result, with the normal value `Success` | - |
| `LINK_STATUS_THRESHOLD` | Link status, with the normal value `UP` | - |
| `OPTICAL_PRESENT_THRESHOLD` | Presence status of the optical module, with the normal value `present` | - |
| `SNR_LANE_DIFF_DB` | Maximum SNR difference among lanes of the optical module | dB |
| `POWER_LANE_DIFF_DB` | Maximum optical power difference among lanes of the optical module | dBm |
| `HCCN_LINK_DOWN_CNT` | Number of HCCN link down events within 24 hours | times |
| `TX_POWER_MW` | Transmit optical power of the host-side optical module (in mW) | mW |
| `RX_POWER_MW` | Receive optical power of the host-side optical module (in mW) | mW |
| `TX_POWER_DBM` | Transmit optical power of the switch-side optical module (in dBm) | dBm |
| `RX_POWER_DBM` | Receive optical power of the switch-side optical module (in dBm) | dBm |
| `TX_BIAS_MA` | Transmit bias current of the optical module | mA |
| `HOST_SNR_DB` | Host-side SNR of the optical module (electrical port/chip side) | dB |
| `MEDIA_SNR_DB` | Media-side SNR of the optical module (optical port/cable side) | dB |
| `NIC_TX_BIAS_MA` | Transmit bias current of the NIC (NIC SFP) optical module | mA |
| `NIC_TX_POWER_DBM` | Transmit optical power of the NIC (NIC SFP) optical module | dBm |
| `NIC_RX_POWER_DBM` | Receive optical power of the NIC (NIC SFP) optical module | dBm |
| `NIC_HOST_SNR_DB` | Host-side SNR of the NIC (NIC SFP) optical module | dB |
| `NIC_MEDIA_SNR_DB` | Media-side SNR of the NIC (NIC SFP) optical module | dB |
| `NIC_LANE_FLAG` | NIC lane exception flag, with the normal value `0` | - |
| `LPO_TX_BIAS_MA` | Transmit bias current of the LPO optical module (only for <term>Ascend 950PR&950DT products</term>) | mA |
| `LPO_TX_POWER_DBM` | Transmit optical power of the LPO optical module (only for <term>Ascend 950PR&950DT products</term>) | dBm |
| `LPO_RX_POWER_DBM` | Receive optical power of the LPO optical module (only for <term>Ascend 950PR&950DT products</term>) | dBm |
| `LPO_POWER_LANE_DIFF_DB` | Maximum optical power difference among lanes of the LPO optical module (only for <term>Ascend 950PR&950DT products</term>) | dBm |

Notes:

- For `DUPLEX_THRESHOLD`, `NET_HEALTH_THRESHOLD`, `LINK_STATUS_THRESHOLD`, `OPTICAL_PRESENT_THRESHOLD`, and `NIC_LANE_FLAG`, the normal value is determined by string comparison: a value is considered normal only when it equals the configured normal value; all other values are considered abnormal.
- `HCCN_LINK_DOWN_CNT` is considered sub-health when it exceeds the warning threshold, and abnormal when it exceeds the alarm (fault) threshold.
- NIC-related thresholds are used only by <term>Ascend 950PR&950DT products</term>.
- Optical module thresholds vary by module type: ODSP and LPO optical modules are supported. When the optical module type is not identified, the ODSP type is used by default.

Configuration validation rules:

- Fields that are not configured retain their default values. If the value is empty (`null` or an empty string), the field is considered unconfigured and the default value is used.
- Numeric fields (such as `low_value_alarm`) must be valid numeric strings; otherwise, the field is ignored.
- `desc`, `unit`, and `normal_value_alarm` must be strings; otherwise, the field is ignored.
- If `desc` exceeds 128 characters or `unit` exceeds 16 characters, the field is ignored. It is not recommended to modify `desc` or `unit`; otherwise, the determination and result display are affected.
- Invalid content such as unknown field names or non-object configuration items is ignored and a warning log is recorded, which does not affect the other configurations taking effect.

>[!NOTE]
> `set_config_dir` is an optional configuration command. The tool still works when it is not set, but it will not perform correlation analysis on the equipment room location dimension, and the fault thresholds are configured with default values.

## Output Description

- On success, the command returns `Configuration succeeded. Configuration directory: {dir_path}`.
- On failure, the command returns `The directory path is empty. Set it again.`,`The directory {dir_path} does not exist. Set it again.`, or `The path {dir_path} is not a directory. Set it again.`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_config_dir /home/user/config set_conn_config /home/user/conn.ini auto_collect_diag
Setting successful. Configuration directory: /home/user/config
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_config_dir /home/user/config
Setting successful. Configuration directory: /home/user/config
```
