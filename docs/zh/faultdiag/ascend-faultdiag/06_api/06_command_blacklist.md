# blacklist命令（屏蔽故障日志）

## 功能说明

用于管理日志屏蔽规则，过滤掉不需要关注的日志信息。

当前仅支持对CANN应用类日志的ERROR日志进行屏蔽操作。

## 命令格式

```shell
ascend-fd blacklist [-h] (-a ADD | -f FILE | -s | -d DELETE [DELETE ...]) [--force]
```

## 参数说明

| 参数         | 类型                  | 必选                       | 说明                                         |
|--------------|-----------------------|----------------------------|----------------------------------------------|
| -h, --help   | -                     | 否                         | 显示帮助信息                                 |
| -a, --add    | String/List[String]   | 是（与-f、-s、-d互斥） | 新增包含关键词的屏蔽规则，支持传入多个关键词 |
| -f, --file   | String                | 是（与-a、-s、-d互斥） | 导入屏蔽规则JSON文件路径                   |
| -s, --show   | -                     | 是（与-a、-f、-d 互斥） | 查看当前已有的屏蔽规则                       |
| -d, --delete | Integer/List[Integer] | 是（与-a、-f、-s互斥） | 删除屏蔽规则，支持传入多个规则序号           |
| --force      | -                     | 否（只与-d、-f共用）   | 删除或导入（覆盖已有规则）时跳过确认提示     |

## 使用示例

### 新增屏蔽规则

添加包含指定关键词的屏蔽规则：

```shell
ascend-fd blacklist -a "ERROR_KEYWORD"
```

添加包含多个关键词的规则：

```shell
ascend-fd blacklist -a "ERROR1 ERROR2 ERROR3"
```

一条屏蔽规则最多支持10个关键词，使用空格分隔。

### 导入屏蔽规则

通过JSON文件批量导入屏蔽规则（会覆盖已有规则）：

```shell
ascend-fd blacklist -f <file.json>
```

JSON文件格式如下：

```json
    {
        "blacklist":[
            ["ERROR2","ERROR3","ERROR4"],
            ["ERR1","ERR2","ERR3","ERR4"]
        ]
    }
```

跳过确认提示：

```shell
ascend-fd blacklist -f <file.json> --force
```

### 查看屏蔽规则

```shell
ascend-fd blacklist -s
```

回显示例：

```output
[BLACKLIST]
0. ERROR1, ERROR2, ERROR3
1. ERR_A, ERR_B
```

### 删除屏蔽规则

```shell
ascend-fd blacklist -d <规则序号>
```

删除多条规则：

```shell
ascend-fd blacklist -d 0 1
```

跳过确认提示：

```shell
ascend-fd blacklist -d 0 --force
```

## 注意事项

- 关键词最长200个字符，支持大小写字母、数字和特殊字符（如 `-`、`.`、`/` 等）。
- 一条规则最多10个关键词。
- 最多保存50条屏蔽规则，超出后丢弃最早的规则。
- 包含 `\` 的关键词需加引号，如 `"ERR\OR"`。
- 屏蔽规则数据存储在 `$HOME/.ascend_faultdiag/custom-blacklist.json` 文件中。
- 用户可通过修改 `ASCEND_FD_HOME_PATH` 环境变量来指定屏蔽规则文件路径，请查阅[环境变量](../07_references/01_common_operations.md#环境变量)。
- ascend-fd 运行错误码请查阅[组件错误码](../07_references/04_appendix.md#组件错误码)。
