# help

<!-- md-trans-meta sourceCommit=9d708be25ec5fef354b48beecd57d9b30752b536 translatedAt=2026-08-24T02:17:43.045Z pushedAt=2026-08-24T02:22:41.569Z -->

## Command Function

Displays help information for all available commands, facilitating quick retrieval of tool capabilities.

## Command Format

| Command Format | Description |
|---------|------|
| `help` | Displays help information |
| `help ?` | Views details |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view command usage.

## Output Description

The console outputs all available commands and their brief help information in column-aligned format.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk help
help                       - Display help information
exit                       - Exit the program
clear                      - Clear screen
about                      - View about diagnostic tool
guide                      - Get guided information
set_config_dir             - Set configuration directory path. Supports " set_config_dir <directory_path> " setting, or " set_config_dir ? " to view details
set_conn_config            - Set connection file path. Supports " set_conn_config <file_path> " setting, or " set_conn_config ? " to view details
set_host_dump_log          - Set server exported log directory. Supports " set_host_dump_log <directory> " to set the directory, or " set_host_dump_log ? " to view details
set_bmc_dump_log           - Set BMC exported log directory. Supports " set_bmc_dump_log <directory> " to set the directory, or " set_bmc_dump_log ? " to view details
set_switch_dump_log        - Set switch command output export directory. Supports " set_switch_dump_log <directory> " to set the directory, or " set_switch_dump_log ? " to view details
collect_bmc_dump_info      - Collect BMC dump info logs online
auto_collect               - Start automatic information collection, supports offline and online collection, suitable for staged collection across different network planes
auto_inspection            - Start inspection result diagnosis, suitable for unified diagnosis after staged collection
auto_diag                  - Start automatic diagnosis, suitable for unified diagnosis after staged collection
auto_collect_diag          - Start one-click automatic collection (online device collection or offline log collection) and diagnosis
clear_cache                - Clear cache. Must be executed before starting a new diagnosis task! Prevents interference with diagnosis results (if clearing is ineffective, please open the tool in administrator mode)
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> help
# Output is the same as above.
```
