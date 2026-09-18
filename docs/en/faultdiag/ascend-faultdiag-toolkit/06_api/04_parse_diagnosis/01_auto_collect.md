# auto_collect

<!-- md-trans-meta sourceCommit=af1baaf30f7bd8ba5afaadd02ab4719614fcf8ee translatedAt=2026-08-24T02:19:01.792Z pushedAt=2026-08-24T02:22:41.600Z -->

## Command Function

Starts automatic information collection, supports offline and online collection, and is applicable to batch collection across different network planes.

| Feature | Description                                             |
|------|------------------------------------------------|
| Collection Mode | Supports offline collection and online collection                                    |
| Data Source | Automatically completes log collection and data parsing based on the configured data source (online connection configuration or offline log directory)         |
| Applicable Scenario | Applicable to batch collection across different network planes (`auto_collect` can be executed multiple times to collect device information from different batches) |
| Follow-up Step | After collection is complete, diagnosis or inspection must be performed on the parsed structured data              |

## Command Format

| Command Format | Description |
|---------|------|
| `auto_collect` | Starts automatic information collection, supports offline and online collection, applicable to batch collection across different network planes |
| `auto_collect ?` | Views details |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view the command usage.

## Data Source Requirements

Before running `auto_collect`, ensure that at least one of the following conditions is met:

- Online device connection information has been configured through `set_conn_config`.
- The offline log directory has been set through `set_host_dump_log`/`set_bmc_dump_log`/`set_switch_dump_log`.

## Output Description

When the command is executed successfully, it returns `Collection complete. If all collections are completed, please proceed with diagnosis/inspection.`

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini auto_collect
# Other log output...
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> auto_collect
Collection complete. If all collections are completed, please proceed with diagnosis/inspection.
```
