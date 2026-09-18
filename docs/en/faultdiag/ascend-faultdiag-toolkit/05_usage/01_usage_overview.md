# Feature Overview

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:15:58.934Z pushedAt=2026-08-24T02:22:41.531Z -->

The ascend-fd-tk tool provides complete end-to-end fault diagnosis capabilities, covering the entire process from data collection to fault localization. This section introduces the usage and typical scenarios of the tool's features.

## Usage

| Usage | Applicable Scenario | Feature |
|------|----------|------|
| Interactive | Temporary debugging, troubleshooting, step-by-step operations | Enter the `>>>` prompt and input commands one by one |
| Non-interactive | Automated O&M, scheduled tasks, script integration | Complete the entire process with a single command |

## Home Directory

Data generated during tool execution is stored in the home directory, and the path varies by platform:

- On Linux, the path is based on the user home directory (`~`), and the home directory is `~/.ascend-faultdiag-toolkit/`.
- On Windows, the path is based on the **current working directory** (the directory where the tool is started), and the home directory is `{current_working_directory}/.ascend-faultdiag-toolkit/`.

Directory structure:

| Path | Purpose |
|------|------|
| `{home_directory}/cache/` | Parsing result cache information (host/BMC/ switch categories, JSON files stored by device IP or directory name) |
| `{home_directory}/logs/ascend-fd-tk.log` | Tool execution log |
| `{home_directory}/report/` | Output directory for diagnosis/inspection reports (`diag_report_{YYYYMMDD_HHMMSS}.xlsx` / `inspection_errors.csv`) |
| `{home_directory}/encrypted_conn_config` | Encrypted file of the online connection configuration file `conn.ini` |

>[!NOTE]
> The maximum size of a single tool execution log file is 10 MB. When the file size reaches the threshold, log rotation and archiving are triggered automatically. Archived files are named `ascend-fd-tk.log.1`, `ascend-fd-tk.log.2`, and so on, where a smaller number indicates a more recent log generation time. The system retains up to 5 archived log files.

## Features

| Feature | Description |
|------|------|
| [Log Collection](02_log_collection.md) and [Log Parsing](03_log_parse.md) | In online mode, logs are collected and parsed automatically; in offline mode, logs are collected first and then parsed. |
| [Fault Diagnosis](04_fault_diagnosis.md) | Performs fault detection and root cause analysis on the parsed data, and generates an Excel diagnostic report. |
| [Fault Inspection](05_fault_inspection.md) | Performs batch health checks based on predefined rules for different customer types, and generates a CSV inspection report. |

## Scenario-Specific Command Execution

| Process              | Applicable Scenario | Core Steps |
|-----------------|----------|----------|
| [Online Diagnosis Flow](07_online_diagnosis.md) | Device accessible (IP/credentials available) | `clear_cache` → `set_config_dir` (optional) → `set_conn_config` → `auto_collect_diag` or `auto_collect` + `auto_diag` |
| [Offline Diagnosis Flow](08_offline_diagnosis.md)      | Only logs obtainable | `clear_cache` → `set_config_dir` (optional) → `set_host_dump_log`/`set_bmc_dump_log`/`set_switch_dump_log` (optional) → `auto_collect_diag` or `auto_collect` + `auto_diag` |
| [Batch Diagnosis Flow](09_batch_diagnosis.md)      | Multiple network planes, large number of devices | `clear_cache` → `set_config_dir` (optional) → repeat N times [`set_conn_config` + `auto_collect` or `set_host_dump_log`/`set_bmc_dump_log`/`set_switch_dump_log` (optional) + `auto_collect`] → `auto_diag` |
| [Customized Inspection](10_customized_inspection.md)     | Batch health check with predefined rules for different customer types | `clear_cache` → `set_config_dir` (optional) → `set_conn_config` or `set_host_dump_log` / `set_bmc_dump_log`/`set_switch_dump_log` (optional) → `auto_collect` + `auto_inspection` |
