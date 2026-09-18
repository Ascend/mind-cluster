# set_switch_dump_log

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:18:42.051Z pushedAt=2026-08-24T02:22:41.592Z -->

## Command Function

Sets the switch command output/log export directory for offline analysis scenarios.

## Command Format

| Command Format | Description |
|---------|------------------|
| `set_switch_dump_log <directory>` | Sets the switch command output/log export directory |
| `set_switch_dump_log ?` | Views details |

## Parameters

| Name | Type | Mandatory | Description |
|------|-----|------|------|
| `<directory>` | String | Yes | Directory for switch command output/log export. |

## Supported Log Types

For details, see [Offline Log Collection (Switch)](../../05_usage/02_log_collection.md#switch-offline-log).

## Output Description

- On success, the command returns `Setting succeeded`.

- On failure, the command returns `The address is empty. Set it again`, `The address {dir_path} does not exist. Set it again`, or `The address {dir_path} is not a folder. Set it again`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_switch_dump_log /data/switch_logs auto_collect_diag
Setting successful
# Other log output...
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_switch_dump_log /data/switch_logs
Setting successful
```
