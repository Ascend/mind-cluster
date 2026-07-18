# API 概述

ascend-fd 提供两种接口方式：

- **命令行接口**：通过 `ascend-fd` 命令使用，适合运维人员和普通用户。
- **SDK 接口**：作为第三方库被 Python 程序导入使用，适合开发者集成到自己的系统中。

## 命令行接口

### 命令格式

```shell
ascend-fd <子命令> [参数]
```

### 子命令列表

| 子命令                                     | 功能           | 说明                         |
|--------------------------------------------|----------------|------------------------------|
| [parse](./02_command_parse.md)             | 日志清洗       | 从原始日志中提取关键信息     |
| [diag](./03_command_diag.md)               | 故障诊断       | 分析单节点或多节点的故障根因 |
| [single-diag](./04_command_single_diag.md) | 单机故障诊断   | 在单台设备上快速诊断         |
| [entity](./05_command_entity.md)           | 自定义故障实体 | 管理自定义故障检测规则       |
| [blacklist](./06_command_blacklist.md)     | 屏蔽故障日志   | 管理日志屏蔽规则             |
| [config](./07_command_config.md)           | 查看配置文件   | 管理当前使用的配置文件路径   |
| [version](./08_command_version.md)         | 查看版本       | 显示当前 ascend-fd 版本      |

> [!NOTE]
>
> - 参数请参考子命令具体章节说明。

## SDK 接口

ascend-fd 提供 Python SDK，方便开发者集成到自动化流程中。

### SDK 模块

| 接口                                                           | 功能说明         |
|----------------------------------------------------------------|------------------|
| [parse_fault_type](./09_sdk_api.md#parse_fault_type)           | 业务日志清洗接口 |
| [parse_root_cluster](./09_sdk_api.md#parse_root_cluster)       | 根因节点清洗接口 |
| [diag_root_cluster](./09_sdk_api.md#diag_root_cluster)         | 根因节点诊断接口 |
| [parse_knowledge_graph](./09_sdk_api.md#parse_knowledge_graph) | 故障事件清洗接口 |
| [diag_knowledge_graph](./09_sdk_api.md#diag_knowledge_graph)   | 故障事件诊断接口 |
