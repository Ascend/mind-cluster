# exit

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
