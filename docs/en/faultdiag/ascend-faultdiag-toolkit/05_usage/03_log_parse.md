# Log Parsing

Use the [auto_collect](../06_api/04_parse_diagnosis/01_auto_collect.md) command for data parsing, converting raw logs into structured data for subsequent diagnosis and inspection.

| Collection Method   | Parsing Behavior                                   |
|--------|----------------------------------------|
| Online data collection | Automatically collect logs via SSH + parsing → Write parsed results to `{home_directory}/cache` |
| Offline log collection   | Collect offline log files in advance → Parsing → Write parsed results to `{home_directory}/cache`  |

## Online Data Collection and Parsing

The configuration file `conn.ini` contains information such as device IP, account, and password/key. For detailed configuration, see [set_conn_config](../06_api/02_config/02_set_conn_config.md).

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_conn_config /path/to/conn.ini auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
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
```

## Offline Log Parsing

In a cluster scenario, simply place logs from multiple devices into the corresponding directories. The directory structure is as follows:

```text
# Simply place the collected log compressed packages directly into the directory (directory names can be customized). The tool supports automatic extraction of log compressed packages — no manual extraction is required. At least one log directory is required for diagnosis.
host log collection directory/
    ├── {host_01_file_name}.tar.gz
    ├── {host_02_file_name}.tar.gz
    └── ...
bmc log collection directory/
    ├── {bmc_01_file_name}.tar.gz
    ├── {bmc_02_file_name}.tar.gz
    └── ...
switch log collection directory/
    ├── {switch_01_file_name}.zip
    ├── {switch_02_file_name}.zip
    └── ...
```

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_config_dir /path/to/your_config_path set_host_dump_log host_log_collection_directory/ set_bmc_dump_log bmc_log_collection_directory/ set_switch_dump_log switch_log_collection_directory/ auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_config_dir /path/to/your_config_path
Setting successful. Configuration directory: /path/to/your_config_path
>>> set_host_dump_log host_log_collection_directory/
Setting successful.
>>> set_bmc_dump_log bmc_log_collection_directory/
Setting successful.
>>> set_switch_dump_log switch_log_collection_directory/
Setting successful.
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
```

## Default Path Reading

When the connection configuration file `conn.ini` or the offline log directory is not manually set, the tool automatically reads the following default files or directories under the current execution path. You need to create the relevant files or directories in advance, configure device connection information, and place offline logs into the corresponding directories.

- Connection configuration: `conn.ini` (When using this file in the current execution path, before running the parsing/diagnosis task multiple times, you need to execute the `clear_cache` command to clear the cache.)
- BMC log directory: `bmc_dump_log/`
- Server log directory: `host_dump_log/`
- Switch log directory: `switch_dump_log/`
