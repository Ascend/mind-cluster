# set_host_dump_log

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:18:33.942Z pushedAt=2026-08-24T02:22:41.586Z -->

## Command Function

Sets the server export log directory for offline analysis scenarios.

## Command Format

| Command Format | Description |
|---------|------|
| `set_host_dump_log <directory>` | Sets the server export log directory |
| `set_host_dump_log ?` | Views details |

## Parameter Description

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<directory>` | String | Yes | Server log directory |

## Supported Log Types

For details, see [Offline Log Collection (Host)](../../05_usage/02_log_collection.md#host-offline-log).

## Output Description

- On success, the command returns `Setting successful`.
- On failure, the command returns `The address is empty. Please set it again`, `The address {dir_path} does not exist. Please set it again`, or `The address {dir_path} is not a folder. Please set it again`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_host_dump_log /data/host_logs auto_collect_diag
Setting successful
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_host_dump_log /data/host_logs
Setting successful
```
