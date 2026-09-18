# Online Diagnosis

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:16:27.086Z pushedAt=2026-08-24T02:22:41.540Z -->

This feature is applicable to accessible cluster devices. Configure device connection information to automatically perform data collection and fault diagnosis.

## Non-interactive Mode (Commands and Outputs)

Non-interactive mode chains multiple commands into a single line, suitable for automated O&M scenarios.

### 1. Clear Cache

Before executing a task, it is recommended to clear the cache to prevent previous diagnosis results from affecting the current diagnosis.

```bash
ascend-fd-tk clear_cache
Cache cleared
```

### 2. Configure Online Data Source and Perform One-Click Diagnosis

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect_diag
Diagnosis complete
```

The command `set_config_dir` for setting the configuration file path is optional.

## Interactive Mode (Commands and Outputs)

### 1. Start the Tool

```bash
# Start the interactive command line
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

In online collection mode, you need to configure the IP address, account, password/key, and other information in the configuration file `conn.ini`. For detailed configuration, see [set_conn_config](../06_api/02_config/02_set_conn_config.md).

```bash
>>> set_conn_config /path/to/conn.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
```

### 5. One-Click Diagnosis

```bash
# Automatically complete collection and diagnosis
>>> auto_collect_diag
Diagnosis complete
```

## View Diagnostic Report

After diagnosis is complete, the report is automatically generated in the report subdirectory under the tool home directory. For details, see [Diagnostic/Inspection Report Description](06_fault_analysis_report.md).
