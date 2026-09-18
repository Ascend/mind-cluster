# set_bmc_dump_log

<!-- md-trans-meta sourceCommit=9d708be25ec5fef354b48beecd57d9b30752b536 translatedAt=2026-08-24T02:18:36.087Z pushedAt=2026-08-24T02:22:41.588Z -->

## Command Function

Sets the BMC export log directory for offline analysis scenarios.

## Command Format

| Command Format | Description            |
|---------|---------------|
| `set_bmc_dump_log <directory>` | Sets the BMC export log directory |
| `set_bmc_dump_log ?` | Views details          |

## Parameter Description

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<directory>` | string | Yes | BMC log directory |

## Supported Log Types

For details, see [Offline Log Collection (BMC)](../../05_usage/02_log_collection.md#BMC-offline-log).

## Output Description

- On success, the command returns `Setting successful`.
- On failure, the command returns `The address is empty. Set it again`, `The address {dir_path} does not exist. Set it again`, or `The address {dir_path} is not a folder. Set it again`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_bmc_dump_log /data/bmc_logs auto_collect_diag
Setting successful
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_bmc_dump_log /data/bmc_logs
Setting successful
```
