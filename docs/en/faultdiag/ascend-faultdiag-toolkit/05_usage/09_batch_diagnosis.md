# Batch Diagnosis

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:37.974Z pushedAt=2026-08-24T02:22:41.544Z -->

This feature applies to scenarios with **multiple network planes** or **a large number of devices**. By executing `auto_collect` multiple times to collect device information in batches, and finally executing `auto_diag` once to complete the full diagnosis.

## Non-interactive Mode (Commands and Outputs)

In non-interactive mode, multiple commands are chained and executed in a single line, which is suitable for automated O&M scenarios.

>[!NOTE]
> The `set_config_dir` command for setting the configuration file path is optional.

### 1. Clear Cache

Before executing a task, it is recommended to clear the cache to prevent previous diagnosis results from affecting the current diagnosis.

```bash
ascend-fd-tk clear_cache
Cache cleared
```

### 2. Configure Data Sources and Batch Diagnose

- Online scenario

    ```bash
    # Batch 1: Collect devices on network plane A
    ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_a.ini auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

    # Batch 2: Collect devices on network plane B
    ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn_plane_b.ini auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

    ...

    # Unified diagnosis
    ascend-fd-tk auto_diag
    Diagnosis complete
    ```

- Offline scenario

    ```bash
    # Batch 1: Collect offline logs from plane A
    ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs_a set_bmc_dump_log /path/to/bmc_logs_a set_switch_dump_log /path/to/switch_logs_a auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection

    # Batch 2: Collect offline logs from plane B
    ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs_b set_bmc_dump_log /path/to/bmc_logs_b set_switch_dump_log /path/to/switch_logs_b auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection

    ...

    # Unified diagnosis
    ascend-fd-tk auto_diag
    Diagnosis complete
    ```

## Interactive Mode (Commands and Outputs)

### 1. Start the Tool

```bash
# Start the interactive command line.
ascend-fd-tk
```

After entering the `>>>` prompt, enter commands one by one. The tool automatically displays help information upon startup.

### 2. Clear Cache

Before executing a task, it is recommended to clear the cache to prevent previous diagnosis results from affecting the current diagnosis.

```bash
>>> clear_cache
Cache cleared
```

### 3. Set Configuration File Path (Optional)

```bash
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
```

### 4. Configure Data Sources and Batch Diagnose

- Online scenario

```bash
>>> set_conn_config /path/to/conn_plane_a.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>>
>>> # Switch to another network plane, reconfigure conn.ini, and collect again.
>>> set_conn_config /path/to/conn_plane_b.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

...

>>> auto_diag
Diagnosis complete
```

- Offline scenario

```bash
>>> set_host_dump_log /path/to/host_logs_a
Setting successful
>>> set_bmc_dump_log /path/to/bmc_logs_a
Setting successful
>>> set_switch_dump_log /path/to/switch_logs_a
Setting successful
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
>>>
>>> # Append the offline log directories of other batches and collect again.
>>> set_host_dump_log /path/to/host_logs_b
Setting successful
>>> set_bmc_dump_log /path/to/bmc_logs_b
Setting successful
>>> set_switch_dump_log /path/to/switch_logs_b
Setting successful
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

...

>>> auto_diag
Diagnosis complete
```

## Viewing the Diagnostic Report

After the diagnosis is complete, the report is automatically generated in the report subdirectory under the tool home directory. For details, see [Diagnostic/Inspection Report Description](06_fault_analysis_report.md).
