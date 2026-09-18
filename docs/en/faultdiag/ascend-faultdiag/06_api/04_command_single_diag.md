# single-diag Command (Single-Server Fault Diagnostics)

<!-- md-trans-meta sourceCommit=0bdfcd26262ca7cc1165bb55c8f28c6a29b377ef translatedAt=2026-08-24T02:28:47.969Z pushedAt=2026-08-24T02:51:43.278Z -->

## Function Description

Used to directly complete log cleaning and fault diagnosis on a single server, without the need for multi-server log dumping.

## Command Format

```shell
ascend-fd single-diag [-h] [-i INPUT_PATH] -o OUTPUT_PATH \
    [--host_log HOST_LOG] [--device_log DEVICE_LOG] \
    [--train_log TRAIN_LOG [TRAIN_LOG ...]] [--process_log PROCESS_LOG] \
    [--env_check ENV_CHECK] [--dl_log DL_LOG] [--mindie_log MINDIE_LOG] \
    [--amct_log AMCT_LOG] [--bus_log BUS_LOG] \
    [--pymotor_vllm_log PYMOTOR_VLLM_LOG]
```

## Parameter Description

| Parameter               | Type                | Mandatory | Description                                     |
|--------------------|---------------------|-----------|-------------------------------------------------|
| `-h`, `--help`         | -                   | No        | Displays help information.                      |
| `-i`, `--input_path`   | String              | No        | Preprocessing data input path.                  |
| `-o`, `--output_path`  | String              | Yes       | Diagnostic result output path.                   |
| `--host_log`         | String              | No        | Host-side OS log directory.                     |
| `--device_log`       | String              | No        | Device-side log directory.                      |
| `--train_log`        | String/List[String] | No        | User training and inference log directory. Multiple directories can be passed. |
| `--process_log`      | String              | No        | CANN App log directory.                         |
| `--env_check`        | String              | No        | NPU network port, status information, and resource information directory. |
| `--dl_log`           | String              | No        | MindCluster component log directory.            |
| `--mindie_log`       | String              | No        | MindIE component log directory.                 |
| `--amct_log`         | String              | No        | AMCT log directory.                   |
| `--bus_log`          | String              | No        | LCNE log directory.                   |
| `--pymotor_vllm_log` | String              | No        | PyMotor/vLLM log directory.                     |

## Usage Examples

### Command Execution

```shell
ascend-fd single-diag -i /tmp/log_dir -o /tmp/diag_out
```

### Diagnosis by Input Log Directories

```shell
ascend-fd single-diag --process_log {collection_directory}/process_log -o /tmp/diag_out
```

## Notes

- Single-server diagnosis returns the fault event analysis result by default.
- If a fault is diagnosed, the status code is the specific fault code; if no fault is diagnosed, the status code is `NORMAL_OR_UNSUPPORTED`.
- Single-server diagnosis scans fault events in all valid logs on the node.
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
- For single-server diagnostic results, see [Basic Diagnosis](03_command_diag.md#basic-diagnosis).
