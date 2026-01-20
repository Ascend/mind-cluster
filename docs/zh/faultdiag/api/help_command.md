# 帮助命令接口<a name="ZH-CN_TOPIC_0000001644316721"></a>

**接口原型<a name="zh-cn_topic_0000001511259053_section15931137194112"></a>**

```
ascend-fd -h
```

或

```
ascend-fd --help
```

**功能说明<a name="zh-cn_topic_0000001511259053_section11116160164115"></a>**

查询命令与参数含义以及使用说明。

**参数说明<a name="zh-cn_topic_0000001511538701_section122149111390"></a>**

**表 1**  参数说明

|**参数**|**缩写**|**是否必选**|**值类型**|**说明**|
|--|--|--|--|--|
|--help|-h|否|-|查询二级命令与参数含义以及使用说明。|


**返回说明<a name="zh-cn_topic_0000001511259053_section1072365114014"></a>**

返回参数与使用说明。

```
usage: ascend-fd [-h] {version,parse,diag,blacklist,entity,single-diag} ...
Ascend Fault Diag
positional arguments:
  {version,parse,diag,blacklist,entity,single-diag}
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


