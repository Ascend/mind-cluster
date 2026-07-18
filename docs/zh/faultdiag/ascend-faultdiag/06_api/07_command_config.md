# config 命令（自定义配置）

## 功能说明

用于管理自定义配置文件，可以配置是否支持清洗 ModelArts 关键日志、配置读取控制台日志大小、配置解析自定义的文件等。

## 语法

```shell
ascend-fd config [-h] (-u UPDATE | -s | -c)
```

## 参数说明

| 参数         | 类型   | 必选                   | 说明                                 |
|--------------|--------|------------------------|--------------------------------------|
| -h, --help   | -      | 否                     | 显示帮助信息                         |
| -u, --update | String | 必选（与 -s, -c 互斥） | 新增或修改自定义配置的 JSON 文件路径 |
| -s, --show   | -      | 必选（与 -u, -c 互斥） | 查看当前自定义配置信息               |
| -c, --check  | -      | 必选（与 -u, -s 互斥） | 校验 custom-fd-config.json 文件      |

## 使用示例

### 新增或修改配置

通过 JSON 文件新增或修改自定义配置：

```shell
ascend-fd config -u <custom-config.json>
```

> - `custom-config.json` 为用户自定义输入文件
> - JSON 文件参考[配置文件说明](#配置文件说明)

回显示例：

```text
The custom config file was updated successfully.
```

### 查看配置

```shell
ascend-fd config -s
```

### 校验配置文件

若用户直接修改 `$HOME/.ascend_faultdiag/custom-fd-config.json` 文件，可执行以下命令进行校验：

```shell
ascend-fd config -c
```

## 配置文件说明

自定义配置文件为 JSON 格式。

1. JSON 文件示例

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

2. JSON 文件字段说明

    | 字段                                  | 类型         | 默认值        | 说明                                                          |
    |---------------------------------------|--------------|---------------|---------------------------------------------------------------|
    | `enable_model_asrt`                   | Boolean      | false         | 是否支持清洗 ModelArts 关键日志                               |
    | `train_log_size`                      | Integer      | 1048576 (1MB) | 配置读取控制台日志大小，单位 Byte                             |
    | `custom_parse_file`                   | List[Object] | []            | 配置解析自定义的文件，最大支持 10 个                          |
    | `custom_parse_file[].file_path_glob`  | String       | —             | `--custom_log` 指定的大目录下，按 Unix 风格通配符模式匹配文件 |
    | `custom_parse_file[].log_time_format` | String       | —             | 日志文件的时间格式，遵循标准日期时间格式字符串                |
    | `custom_parse_file[].source_file`     | List[String] | —             | 日志文件类型，最大支持 10 个                                  |
    | `timezone_config`                     | Object       | —             | 时区配置                                                      |
    | `timezone_config.lcne`                | Boolean      | false         | 是否支持 LCNE 日志时区转换                                    |

> [!NOTE]
>
> - 若用户配置了自定义文件解析规则，即 JSON 文件中的 custom_parse_file 字段，可执行命令（`ascend-fd parse --custom_log worker0/ -o <output_dir>`）对自定义的解析文件进行清洗，将会清洗通配符模式（按照示例则为 worker0/test_custom/*.log）匹配到的文件。
> - 清洗自定义日志文件时，只支持 --custom_log 命令，不支持 -i 命令。

## 注意事项

- 自定义配置数据存储在 `$HOME/.ascend_faultdiag/custom-fd-config.json` 文件中。
- 用户可通过修改 `ASCEND_FD_HOME_PATH` 环境变量来指定配置文件路径，请查阅[环境变量](../07_references/01_common_operations.md#环境变量)。
- ascend-fd 运行错误码请查阅[组件错误码](../07_references/04_appendix.md#组件错误码)。
