# Offline Diagnosis

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:16:26.701Z pushedAt=2026-08-24T02:22:41.542Z -->

This feature is suitable for scenarios where offline logs have been collected in advance, and fault diagnosis is performed based on the existing logs.

## Non-interactive Mode (Commands and Outputs)

The non-interactive mode chains multiple commands in a single line for execution, which is suitable for automated O&M scenarios.

### 1. Clear Cache

Before running a task, clear the cache to prevent previous diagnosis results from affecting the current diagnosis:

```bash
ascend-fd-tk clear_cache
Cache cleared
```

### 2. Configure Offline Data Source and Perform One-Click Diagnosis

```bash
# Set the Host server log directory, BMC log directory, and switch log directory + diagnosis
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log /path/to/host_logs set_bmc_dump_log /path/to/bmc_logs set_switch_dump_log /path/to/switch_logs auto_collect_diag
Diagnosis complete
```

> [!NOTE]
> The command `set_config_dir` for setting the configuration file path is optional.

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

### 3. Set the Configuration File Path (Optional)

```bash
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
```

### 4. Configure Data Source

In offline analysis mode, you need to set the log directory. For detailed configuration, see [Log Collection](02_log_collection.md).

```bash
>>> set_host_dump_log /path/to/host_logs
Setting successful
>>> set_bmc_dump_log /path/to/bmc_logs
Setting successful
>>> set_switch_dump_log /path/to/switch_logs
Setting successful
```

### 5. One-Click Diagnosis

```bash
# Automatically complete collection and diagnosis.
>>> auto_collect_diag
Diagnosis complete
```

## Viewing the Diagnostic Report

After the diagnosis is complete, the report is automatically generated in the report subdirectory under the tool home directory. For details, see [Diagnostic/Inspection Report Description](06_fault_analysis_report.md).
