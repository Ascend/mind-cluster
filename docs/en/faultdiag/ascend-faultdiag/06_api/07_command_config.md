# config Command (Custom Configuration)

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:29:19.032Z pushedAt=2026-08-24T02:51:43.282Z -->

## Function Description

Used to manage custom configuration files. You can configure whether to support cleaning ModelArts key logs, configure the size of console logs to read, configure parsing of custom files, and so on.

## Syntax

```shell
ascend-fd config [-h] (-u UPDATE | -s | -c)
```

## Parameter Description

| Parameter    | Type   | Mandatory                          | Description                                                       |
|--------------|--------|------------------------------------|-------------------------------------------------------------------|
| `-h`, `--help`   | -      | No                                 | Displays help information.                                        |
| `-u`, `--update` | String | Mandatory (mutually exclusive with `-s`, `-c`) | Path of the JSON file for adding or modifying custom configuration. |
| `-s`, `--show`   | -      | Mandatory (mutually exclusive with `-u`, `-c`) | Views the current custom configuration information.               |
| `-c`, `--check`  | -      | Mandatory (mutually exclusive with `-u`, `-s`) | Validates the `custom-fd-config.json` file.                         |

## Usage Examples

### Add or Modify Configuration

Add or modify custom configuration through a JSON file:

```shell
ascend-fd config -u <custom-config.json>
```

- `custom-config.json` is the user-defined input file.
- For the JSON file, refer to [Configuration File Description](#configuration-file-description).

Example output:

```text
The custom config file was updated successfully.
```

### View Configuration

```shell
ascend-fd config -s
```

### Verify the Configuration File

If you directly modify the `$HOME/.ascend_faultdiag/custom-fd-config.json` file, run the following command to verify it:

```shell
ascend-fd config -c
```

## Configuration File Description

The custom configuration file is in JSON format.

- JSON file example

    ```json
    {
        "enable_model_asrt": false,
        "train_log_size": 1048576,
        "custom_parse_file": [
            {
                "file_path_glob": "test_custom/*.log",
                "log_time_format": "%Y-%m-%d-%H:%M:%S.%f",
                "source_file": ["CustomLog"]
            }
        ],
        "timezone_config" : {
            "lcne" : true
        }
    }
    ```

- JSON file field description

    | Field                                 | Type         | Default Value | Description                                                          |
    |---------------------------------------|--------------|---------------|---------------------------------------------------------------|
    | `enable_model_asrt`                   | Boolean      | `false`         | Whether to parse ModelArts key logs                               |
    | `train_log_size`                      | Integer      | `1048576` (1MB) | Configures the size of console logs to read, in bytes                             |
    | `custom_parse_file`                   | List[Object] | []            | Configures custom files to parse, with a maximum of 10 supported                          |
    | `custom_parse_file[].file_path_glob`  | String       | —             | Matches files by Unix-style wildcard pattern under the large directory specified by `--custom_log` |
    | `custom_parse_file[].log_time_format` | String       | —             | Time format of the log file, following the standard date-time format string                |
    | `custom_parse_file[].source_file`     | List[String] | —             | Log file type, with a maximum of 10 supported                                  |
    | `timezone_config`                     | Object       | —             | Time zone configuration                                                      |
    | `timezone_config.lcne`                | Boolean      | false         | Whether to convert LCNE log time zone                                   |

> [!NOTE]
>
> - If you have configured custom file parsing rules, that is, the `custom_parse_file` field in the JSON file, the command (`ascend-fd parse --custom_log worker0/ -o <output_dir>`) can be executed to parse the custom files. Files matched by the wildcard pattern (`worker0/test_custom/*.log` according to the example) will be parsed.
> - For custom log file parsing, only the `--custom_log command` is supported; the `-i` command is not supported.

## Notes

- Custom configuration data is stored in the `$HOME/.ascend_faultdiag/custom-fd-config.json` file.
- You can specify the configuration file path by modifying the `ASCEND_FD_HOME_PATH` environment variable. For details, see [Environment Variables](../07_references/01_common_operations.md#environment-variable-description).
- For ascend-fd runtime error codes, see [Component Error Codes](../07_references/04_appendix.md#component-error-codes).
