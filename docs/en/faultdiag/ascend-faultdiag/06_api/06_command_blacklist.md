# blacklist Command (Fault Log Masking)

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:29:08.164Z pushedAt=2026-08-24T02:51:43.280Z -->

## Function Description

Used to manage log masking rules and filter out log information that does not need attention.

Currently, only ERROR logs of CANN App logs are supported for masking operations.

## Command Format

```shell
ascend-fd blacklist [-h] (-a ADD | -f FILE | -s | -d DELETE [DELETE ...]) [--force]
```

## Parameter Description

| Parameter        | Type                  | Mandatory                          | Description                                                          |
|-------------|-----------------------|------------------------------------|----------------------------------------------------------------------|
| `-h`, `--help`  | -                     | No                                 | Displays help information.                                           |
| `-a`, `--add`   | String/List[String]   | Mandatory (mutually exclusive with `-f`, `-s`, `-d`) | Adds masking rules containing keywords. Multiple keywords are supported. |
| `-f`, `--file`  | String                | Mandatory (mutually exclusive with `-a`, `-s`, `-d`) | Path of the JSON file for importing masking rules.                 |
| `-s`, `--show`  | -                     | Mandatory (mutually exclusive with `-a`, `-f`, `-d`) | Displays the current masking rules.                                |
| `-d`, `--delete` | Integer/List[Integer] | Mandatory (mutually exclusive with `-a`, `-f`, `-s`) | Deletes masking rules. Multiple rule numbers are supported.        |
| `--force`     | -                     | Optional (only used with `-d`, `-f`)  | Skips the confirmation prompt when deleting or importing (overwriting existing rules). |

## Usage Examples

### Add a Masking Rule

Add a masking rule that contains a specified keyword:

```shell
ascend-fd blacklist -a "ERROR_KEYWORD"
```

Add a rule that contains multiple keywords:

```shell
ascend-fd blacklist -a "ERROR1 ERROR2 ERROR3"
```

A masking rule supports a maximum of 10 keywords, separated by spaces.

### Import Masking Rules

Import masking rules in batches through a JSON file (existing rules will be overwritten):

```shell
ascend-fd blacklist -f <file.json>
```

The JSON file format is as follows:

```json
    {
        "blacklist":[
            ["ERROR2","ERROR3","ERROR4"],
            ["ERR1","ERR2","ERR3","ERR4"]
        ]
    }
```

Skip the confirmation prompt:

```shell
ascend-fd blacklist -f <file.json> --force
```

### View Masking Rules

```shell
ascend-fd blacklist -s
```

Example output:

```text
[BLACKLIST]
0. ERROR1, ERROR2, ERROR3
1. ERR_A, ERR_B
```

### Delete Masking Rules

```shell
ascend-fd blacklist -d <rule_id>
```

Delete multiple rules:

```shell
ascend-fd blacklist -d 0 1
```

Skip the confirmation prompt:

```shell
ascend-fd blacklist -d 0 --force
```

## Notes

- A keyword can contain up to 200 characters, including uppercase and lowercase letters, digits, and special characters (such as `-`, `.`, `/`, etc.).
- A rule can contain up to 10 keywords.
- Up to 50 masking rules can be saved. When the limit is exceeded, the earliest rules are discarded.
- A keyword containing `\` must be enclosed in quotation marks, for example, `"ERR\OR"`.
- Masking rule data is stored in the `$HOME/.ascend_faultdiag/custom-blacklist.json` file.
- Users can specify the masking rule file path by modifying the `ASCEND_FD_HOME_PATH` environment variable. For details, see [Environment Variables](../07_references/01_common_operations.md#environment-variable-description).
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
