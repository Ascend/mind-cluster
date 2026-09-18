# collect_bmc_dump_info

<!-- md-trans-meta sourceCommit=9d708be25ec5fef354b48beecd57d9b30752b536 translatedAt=2026-08-24T02:18:53.476Z pushedAt=2026-08-24T02:22:41.598Z -->

## Command Function

Collects BMC dump info logs online. Based on the BMC connection information configured by `set_conn_config`, it remotely logs in to the BMC to perform log collection and downloads the logs to the local cache directory. This command only triggers collection on the BMC side and does not automatically start diagnostics.

## Command Format

| Command Format | Description                    |
|---------|-----------------------|
| `collect_bmc_dump_info` | Collects BMC dump info logs online |
| `collect_bmc_dump_info ?` | Views details                  |

## Parameter Description

- No parameters. `?` is a built-in help identifier used to view the command usage.
- Before execution, configure the BMC device connection information using the `set_conn_config` command.

## Output Description

After the collection is complete, the console returns `Collection complete. Check the log path {path}`. Logs are stored in `{home_directory}/cache/bmc_dump_cache` by default.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk set_conn_config /home/user/conn.ini collect_bmc_dump_info
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
Collection complete. Please check the log path: /home/user/.ascend-faultdiag-toolkit/cache/bmc_dump_cache
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> set_conn_config /home/user/conn.ini
Setting successful. Please delete the configuration file containing plaintext passwords as soon as possible.
>>> collect_bmc_dump_info
Collection complete. Please check the log path: /home/user/.ascend-faultdiag-toolkit/cache/bmc_dump_cache
```
