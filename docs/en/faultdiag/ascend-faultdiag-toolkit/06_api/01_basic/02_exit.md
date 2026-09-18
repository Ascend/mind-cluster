# exit

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:17:42.107Z pushedAt=2026-08-24T02:22:41.565Z -->

## Command Function

Exits the interactive program of the diagnostic tool. After execution, the tool process terminates. This command is meaningful only in interactive mode.

## Command Format

| Command Format | Description |
|---------|------|
| `exit` | Exits the program |
| `exit ?` | Views details |

## Parameter Description

No parameters. `?` is a built-in help indicator used to view command usage.

## Output Description

Before exiting, the console returns `Goodbye!` and then the tool process terminates.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk exit
Goodbye!
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> exit
Goodbye!
```
