# clear

<!-- md-trans-meta sourceCommit=e50286ce9b97df080579d7c724af5b63dfeb835e translatedAt=2026-08-24T02:17:50.219Z pushedAt=2026-08-24T02:22:41.571Z -->

## Command Function

Clears the content displayed on the current terminal screen to facilitate viewing subsequent output. This command affects only the screen display and does not clear the internal state or cached data of the tool.

## Command Format

| Command Format | Description |
|---------|------|
| `clear` | Clears the screen |
| `clear ?` | Views details |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view the command usage.

## Output Description

The current terminal screen display is cleared after command execution (without affecting the internal state of the tool).

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk clear
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> clear
```
