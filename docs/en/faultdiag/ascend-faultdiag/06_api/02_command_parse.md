# parse Command (Log Parsing)

<!-- md-trans-meta sourceCommit=0bdfcd26262ca7cc1165bb55c8f28c6a29b377ef translatedAt=2026-08-24T02:28:38.421Z pushedAt=2026-08-24T02:51:43.276Z -->

## Function Description

Parses raw logs and extracts valid information.

## Command Format

```shell
ascend-fd parse [-h] [-i INPUT_PATH] -o OUTPUT_PATH \
    [--host_log HOST_LOG] [--device_log DEVICE_LOG] \
    [--train_log TRAIN_LOG [TRAIN_LOG ...]] [--process_log PROCESS_LOG] \
    [--env_check ENV_CHECK] [--dl_log DL_LOG] [--mindie_log MINDIE_LOG] \
    [--amct_log AMCT_LOG] [--bmc_log BMC_LOG] [--lcne_log LCNE_LOG] \
    [--bus_log BUS_LOG] [--pymotor_vllm_log PYMOTOR_VLLM_LOG] \
    [--custom_log CUSTOM_LOG] [-p]
```

## Parameter Description

| Parameter          | Type   | Mandatory (Yes/No)             | Description                                                       |
|--------------------|--------|------------------------|-------------------------------------------------------------------|
| `-h`, `--help`         | -      | No                     | Displays help information.                                        |
| `-i`, `--input_path`   | String | No                     | Input path of preprocessed data.                                  |
| `-o`, `--output_path`  | String | Yes                    | Output path of the cleaning result.                               |
| `--host_log`         | String | No                     | Host OS log directory.                                            |
| `--device_log`       | String | No                     | Device-side log directory.                                        |
| `--train_log`        | String | No                     | User training and inference log directory, up to 20 entries.      |
| `--process_log`      | String | No                     | CANN App log directory.                                           |
| `--env_check`        | String | No                     | NPU network port, status information, and resource information directory. |
| `--dl_log`           | String | No                     | MindCluster component log directory.                              |
| `--mindie_log`       | String | No                     | MindIE component log directory.                                   |
| `--amct_log`         | String | No                     | AMCT component log directory.                                     |
| `--bmc_log`          | String | No                     | BMC-side log directory.                                           |
| `--lcne_log`         | String | No, mutually exclusive with `--bus_log` | LCNE log directory, with the same effect as `--bus_log`. |
| `--bus_log`          | String | No, mutually exclusive with `--lcne_log` | LCNE log directory, with the same effect as `--lcne_log`. |
| `--pymotor_vllm_log` | String | No                     | PyMotor/vLLM log directory.                                       |
| `--custom_log`       | String | No                     | Custom parsing file directory.                                    |
| `-p`, `--performance`  | -      | No                     | Parses data of the two performance degradation detection modules: device resources and network congestion. |

## Usage Examples

### Basic Parsing

```shell
ascend-fd parse -i /tmp/log_dir -o /tmp/parse_out
```

### Parsing with Performance Degradation Data

```shell
ascend-fd parse -i /tmp/log_dir -o /tmp/parse_out -p
```

### Specific-Component Parsing Logs

```shell
ascend-fd parse --process_log /tmp/cann_log --train_log /tmp/train_log -o /tmp/parse_out
```

## Parsing Output Result

Parsing output directory structure:

```text
└── parsing output directory
    ├── ascend-kg-parser.json
    ├── ascend-kg-analyzer.json
    ├── ascend-rc-parser.json
    ├── device_ip_info.json
    ├── mindie-cluster-info.json
    ├── server-info.json
    ├── nad_clean.csv
    ├── nic_clean.csv
    ├── process_{core_num}.csv
    ...
    └── plog-parser-{pid}-{0/1}.log
```

### Parsing Result Description

| File                          | Description                                                           |
|-------------------------------|-----------------------------------------------------------------------|
| `ascend-kg-parser.json`       | Fault event analysis parsing result. (Legacy file, compatible with ascend-fd 6.0.0 and earlier versions) |
| `ascend-kg-analyzer.json`     | Fault event analysis parsing result. (New file, for ascend-fd versions after 6.0.0) |
| `ascend-rc-parser.json`       | Root cause node analysis parsing result                              |
| `device_ip_info.json`         | Device IP information                                                 |
| `plog-parser-{pid}-{0/1}.log` | Parsed logs after root cause node analysis, saved by PID            |
| `mindie-cluster-info.json`    | MindIE Pod log parsing result                                        |
| `server-info.json`            | MindIE server information                                   |
| `nad_clean.csv`               | Computing frequency reduction parsing result (requires the `-p` parameter) |
| `nic_clean.csv`               | Network congestion parsing result (requires the `-p` parameter)      |
| `process_{core_num}.csv`      | CPU resource preemption parsing result (requires the `-p` parameter) |

## Notes

- Before parsing, ensure that the output directory has at least 5 GB of available drive space.
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
