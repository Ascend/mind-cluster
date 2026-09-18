# Fault Diagnosis

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:16:00.085Z pushedAt=2026-08-24T02:22:41.533Z -->

After completing [log collection](02_log_collection.md) and [log parsing](03_log_parse.md) with the ascend-fd-tk tool, use the diagnosis function to perform fault analysis on the parsed structured data.

## Diagnosis Commands

The tool provides two diagnosis modes:

| Command                                                    | Description                                                     |
|-------------------------------------------------------|--------------------------------------------------------|
| [auto_collect_diag](../06_api/04_parse_diagnosis/03_auto_collect_diag.md) | One-click diagnosis. Integrates the collection, parsing, and diagnosis steps, equivalent to `auto_collect` + `auto_diag`. Suitable for scenarios where end-to-end diagnosis is completed at once. |
| [auto_diag](../06_api/04_parse_diagnosis/02_auto_diag.md)            | Diagnosis only (collection must be completed first).                                            |

## Using the `auto_collect_diag` Command

### Online Scenario

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect_diag
Diagnosis complete
```

Interactive mode (display command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> auto_collect_diag
Diagnosis complete
```

### Offline Scenario

Non-interactive mode (display command and output):

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect_diag
Diagnosis complete
```

Interactive mode (display command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs
Setting successful
>>> set_bmc_dump_log /path/to/bmc_logs
Setting successful
>>> set_switch_dump_log /path/to/switch_logs
Setting successful
>>> auto_collect_diag
Diagnosis complete
```

## Using the `auto_diag` Command

### Online Scenario

Non-interactive mode (display command and output):

```bash
# Step 1: Collection and parsing
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

# Step 2: Perform diagnosis
ascend-fd-tk auto_diag
Diagnosis complete
```

Interactive mode (display command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>> auto_diag
Diagnosis complete
```

### Offline Scenario

Non-interactive mode (display command and output):

```bash
# Step 1: Collection and parsing
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

# Step 2: Perform diagnosis
ascend-fd-tk auto_diag
Diagnosis complete
```

Interactive mode (display command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs
Setting successful
>>> set_bmc_dump_log /path/to/bmc_logs
Setting successful
>>> set_switch_dump_log /path/to/switch_logs
Setting successful
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>> auto_diag
Diagnosis complete
```

## Viewing the Diagnosis Report

After diagnosis is complete, the report is automatically generated in the report subdirectory under the tool's home directory. For details, see [Diagnostic/Inspection Report Description](06_fault_analysis_report.md).
