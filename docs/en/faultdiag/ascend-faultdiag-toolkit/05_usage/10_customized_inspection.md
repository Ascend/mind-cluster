# Customized Inspection

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:47.204Z pushedAt=2026-08-24T02:22:41.547Z -->

This feature is suitable for customizing inspection dimensions on demand to meet link inspection requirements in customer-specific scenarios. The inspection feature is a beta feature and is not recommended for use in production environments.

## Non-interactive Mode (Commands and Outputs)

The non-interactive mode chains multiple commands in a single line for execution, which is suitable for automated O&M scenarios.

>[!NOTE]
> The command `set_config_dir` for setting the configuration file path is optional.

### 1. Clear Cache

Before executing a task, it is recommended to clear the cache to prevent previous diagnosis results from affecting the current diagnosis.

```bash
ascend-fd-tk clear_cache
Cache cleared
```

### 2. Configure Data Source and Run Inspection

- Online scenario

    ```bash
    # Step 1: Collect information.
    ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

    # Step 2: Run inspection for a specific customer type.
    ascend-fd-tk auto_inspection <customer_type>
    Diagnosis complete
    ```

- Offline scenario

    ```bash
    # Step 1: Collect information.
    ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect
    Collection complete. If all collections are completed, please proceed with diagnosis/inspection.

    # Step 2: Run the inspection for a specific customer type.
    ascend-fd-tk auto_inspection <customer_type>
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

### 3. Configure Data Source

Refer to the data source configuration steps in [Online Diagnosis](07_online_diagnosis.md) or [Offline Diagnosis](08_offline_diagnosis.md).

### 4. Inspection

Use the [auto_inspection](../06_api/05_inspection/auto_inspection.md) command to perform inspection for a specific customer type.

```bash
>>> auto_inspection <customer_type>
Diagnosis complete
```

## Viewing the Inspection Report

After the inspection is complete, the report is automatically generated in the report subdirectory under the tool's home directory. For details, see [Diagnostics/Inspection Report Description](06_fault_analysis_report.md).
