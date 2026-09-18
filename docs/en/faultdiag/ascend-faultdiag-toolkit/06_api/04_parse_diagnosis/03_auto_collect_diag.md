# auto_collect_diag

<!-- md-trans-meta sourceCommit=9d708be25ec5fef354b48beecd57d9b30752b536 translatedAt=2026-08-24T02:19:09.932Z pushedAt=2026-08-24T02:22:41.606Z -->

## Command Function

Starts one-click automatic collection and diagnosis, and automatically executes the collection (online device collection or offline log collection) and diagnosis process, which is equivalent to executing `auto_collect` + `auto_diag` in sequence.

**Applicable scenarios**: The device is directly accessible or logs are ready, and you want to complete end-to-end diagnosis in one step. It is suitable for one-time diagnosis on a single network plane. If you need to collect data from multiple network planes in batches, use `auto_collect` + `auto_diag` to execute step by step.

## Command Format

| Command Format | Description                         |
|---------|----------------------------|
| `auto_collect_diag` | Starts one-click automatic collection (online device collection or offline log collection) diagnosis |
| `auto_collect_diag ?` | Views details                       |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view the command usage.

## Execution Process

After `auto_collect_diag` is executed, the tool automatically completes diagnosis as follows:

1. **Automatic collection**: Collects device logs and status information based on the configured data sources (online connection/offline log directory).
2. **Data parsing**: Parses raw logs, extracts structured data, and stores it in the cache.
3. **Fault diagnosis**: Checks all diagnostic items on the parsed data to locate faults.
4. **Report generation**: Output the diagnostic report in Excel format to the `report/` subdirectory under the home directory.

## Output Description

- Success: `Diagnosis completed`
- Report generation failed: `Report generation failed. Please release the file lock and run 'auto_diag' again to regenerate the report.`

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect_diag
# Other log output...
Diagnosis completed
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> auto_collect_diag
Diagnosis completed
```
