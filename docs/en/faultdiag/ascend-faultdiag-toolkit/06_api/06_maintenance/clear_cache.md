# clear_cache

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:19:26.686Z pushedAt=2026-08-24T02:22:41.624Z -->

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
