# clear_cache

## Command Function

Clears the cache.

- Clears cache files generated during tool execution.
- It is recommended to run this command before executing a new diagnostic task to prevent the previous diagnostic results from interfering with the current diagnosis.
- If the clearing does not take effect, open the tool in administrator mode.

## Command Format

| Command Format | Description                                              |
|---------|-------------------------------------------------|
| `clear_cache` | Clears the cache. Be sure to run this command before executing a new diagnostic task to avoid interference with diagnostic results. (If the clearing does not take effect, open the tool in administrator mode.) |
| `clear_cache ?` | Views details.                                            |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view the command usage.

## Output Description

- Success: `Cache cleared`
- Failure: `Failed to clean up {collect_cache}: {err}`

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk clear_cache
Cache cleared
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> clear_cache
Cache cleared
```
