# guide

<!-- md-trans-meta sourceCommit=9d708be25ec5fef354b48beecd57d9b30752b536 translatedAt=2026-08-24T02:18:09.444Z pushedAt=2026-08-24T02:22:41.577Z -->

## Command Function

Obtains the usage guide information of the diagnostic tool. It is recommended to run this command first when using the tool for the first time to understand the overall process.

## Command Format

| Command Format | Description |
|---------|------|
| `guide` | Obtains the wizard information. |
| `guide ?` | Views details. |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view command usage.

## Output Description

The console outputs a three-stage wizard: ① Preparation of collection content (online/offline/default paths) → ② Start collection/analysis and diagnosis → ③ Clear cache.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk guide
        I. Preparation of Collection Content
        Please select the device information to collect or logs to import based on the faulty device. It is not necessary to import full device information or logs. Set any of the following online or offline collection and analysis addresses as needed.

        1. Online Collection Preparation
        If online device information collection is required, use the "set_conn_config" command to set device information. For specific configuration, use "set_conn_config ?" to view details.

        2. Offline Log Parsing Preparation
        2.1 Set the server log directory path
        Use the "set_host_dump_log" command to set the offline log directory. For specific configuration, use "set_host_dump_log ?" to view details.

        2.2 Set the BMC log directory path
        Use the "set_bmc_dump_log" command to set the offline log directory. For specific configuration, use "set_bmc_dump_log ?" to view details.

        2.3 Set the switch command output text directory path
        Use the "set_switch_dump_log" command to set the offline log directory. For specific configuration, use "set_switch_dump_log ?" to view details.

        3. Default Read Paths
        When the above files or directories are not manually set, the tool automatically reads the following default files or directories under the execution path. Users need to manually create these files or directories in advance:
        Connection configuration: conn.ini
        BMC log directory: bmc_dump_log
        Host log directory: host_dump_log
        Switch log directory: switch_dump_log

        II. Start Collection/Analysis & Diagnosis
        Execute "auto_collect_diag" to start online collection/offline analysis and diagnosis.

        III. Clear Cache
        This tool supports staged collection with unified diagnosis. Therefore, cache remains after a single diagnosis session. If the diagnosis task is complete, please use "clear_cache" to clear the cache (if clearing is ineffective, please open the tool in administrator mode) to avoid affecting the next diagnosis result.

        Summary:
        1. First, use "set_conn_config" to set the device IP configuration file to access, or use "set_bmc_dump_log", "set_host_dump_log", and "set_switch_dump_log" to set offline log directories, or simply place logs directly into the default directories.
        2. As long as at least one of the above settings exists, you can use "auto_collect_diag" to collect/analyze and diagnose, and output the report.
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> guide
# Output same as above
```
