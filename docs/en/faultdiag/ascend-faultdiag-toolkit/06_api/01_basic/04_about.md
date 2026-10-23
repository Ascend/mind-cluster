# about

## Command Function

Views the version information of the diagnostic tool, which is commonly used to provide version information during installation verification or problem locating.

## Command Format

| Command Format | Description |
|---------|------|
| `about` | Views information about the diagnostic tool |
| `about ?` | Views details |

## Parameter Description

No parameters. `?` is a built-in help identifier used to view command usage.

## Output Description

The console returns the version information: `MindCluster ascend-faultdiag-toolkit: <version>`.

## Examples

Non-interactive mode (command and output):

```bash
ascend-fd-tk about
        MindCluster ascend-faultdiag-toolkit: v0.10
```

Interactive mode (command and output):

```bash
ascend-fd-tk
>>> about
        MindCluster ascend-faultdiag-toolkit: v0.10
```
