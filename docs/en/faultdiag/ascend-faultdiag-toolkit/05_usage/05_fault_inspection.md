# Fault Inspection

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:08.192Z pushedAt=2026-08-24T02:22:41.536Z -->

After completing [log collection](02_log_collection.md) and [log parsing](03_log_parse.md) with the ascend-fd-tk tool, use the [auto_inspection](../06_api/05_inspection/auto_inspection.md) command to perform batch health checks based on predefined rules for different customer types, and periodically inspect the cluster to detect potential link anomalies in advance. The inspection feature is a beta feature and is not recommended for use in production environments.

The main differences between diagnosis and inspection are as follows:

- **Diagnosis**: Performs root cause analysis based on log data collected when a fault has occurred, and outputs a `.xlsx` diagnosis report.
- **Inspection**: For scenarios where no explicit fault has occurred, performs batch health checks based on predefined rules, and outputs a `.csv` inspection report.

## Online Inspection

Non-interactive mode (command and output):

```bash
# Step 1: Collection and parsing
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

# Step 2: Execute inspection
ascend-fd-tk auto_inspection
Inspection complete
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_conn_config /path/to/conn.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>> auto_inspection
Inspection complete
```

## Offline Inspection

Non-interactive mode (command and output):

```bash
# Step 1: Collection and parsing
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

# Step 2: Execute inspection
ascend-fd-tk auto_inspection
Inspection complete
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_host_dump_log /path/to/host_logs
Setting successful.
>>> set_bmc_dump_log /path/to/bmc_logs
Setting successful.
>>> set_switch_dump_log /path/to/switch_logs
Setting successful.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>> auto_inspection
Inspection complete
```

## Viewing Inspection Reports

After the inspection is complete, the report is automatically generated in the report subdirectory under the tool's home directory. For details, see [Diagnostic/Inspection Report Description](06_fault_analysis_report.md).
