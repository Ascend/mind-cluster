# Help Command Interface<a name="ZH-CN_TOPIC_0000001644316721"></a>

## Prototype<a name="zh-cn_topic_0000001511259053_section15931137194112"></a>

```shell
ascend-fd -h
```

or

```shell
ascend-fd --help
```

## Function<a name="zh-cn_topic_0000001511259053_section11116160164115"></a>

Queries the meanings and usage instructions of commands and parameters.

## Parameters<a name="zh-cn_topic_0000001511538701_section122149111390"></a>

**Table 1**  Parameters

|Parameter|Abbreviation|Required|Value Type|Description|
|--|--|--|--|--|
|`--help`|`-h`|No|-|Queries the meanings and usage instructions of level-2 commands and parameters.|

## Returns<a name="zh-cn_topic_0000001511259053_section1072365114014"></a>

Example: Return parameters and usage instructions.

```ColdFusion
usage: ascend-fd [-h] {version,parse,diag,blacklist,config,entity,single-diag} ...
Ascend Fault Diag
positional arguments:
  {version,parse,diag,blacklist,config,entity,single-diag}
    version             show ascend-fd version
    parse               parse origin log files
    diag                diag parsed log files
    blacklist           filter invalid CANN logs by blacklist for parsing
    config              custom configuration parsing files
    entity              perform operations on the user-defined faulty entity.
    single-diag         single parse and diag log files
optional arguments:
  -h, --help            show this help message and exit
```
