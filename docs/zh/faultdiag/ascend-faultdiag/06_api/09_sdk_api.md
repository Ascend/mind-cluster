# SDK 接口参考

## 调用说明

使用 SDK 时，会在 `$HOME/.ascend_faultdiag` 目录下生成操作日志和运行日志，目录结构如下：

```text
$HOME/.ascend_faultdiag
└── ascend_faultdiag_operation.log    # 操作日志
└── RUN_LOG                           # 运行日志
  └─ 20241104142355468743_6797877f-7143-443f-a9c6-361e33032c5c
```

> [!NOTE]
>
> - 日志文件大小不超过 10MB，超过限制大小后将自动转储到另一个日志文件。
> - 同 PID 的日志文件数量最大不超过 10 个，超过限制个数后将自动覆盖最早创建的日志。

## SDK 接口定义

### parse_fault_type

业务日志清洗接口。

#### 接口导入

```python
from ascend_fd import parse_fault_type
```

#### 接口定义

```python
parse_fault_type(input_log_list: list) -> Tuple[List, List]
```

##### 请求参数表

| 参数           | 类型 | 说明                   |
|----------------|------|------------------------|
| input_log_list | List | 用户输入的业务日志列表 |

- **input_log_list 示例**

```json
[
    {
        "log_domain": {
            "server": "10.1.1.1",
            "port": 8080,
            "device": ["0", "1"]
        },
        "log_items": [
            {
                "item_type": "MindIE",
                "log_lines": [
                    "[ERROR] xxx",
                    "[ERROR] yyy"
                ]
            }
        ]
    }
]
```

- **input_log_list 字段说明**

| 字段                    | 类型         | 必填 | 说明                   |
|-------------------------|--------------|------|------------------------|
| `log_domain`            | Object       | 是   | 日志域信息             |
| `log_domain.server`     | String       | 是   | 服务器 IP              |
| `log_domain.port`       | Integer      | 是   | 服务器端口             |
| `log_domain.device`     | List[String] | 是   | 发生过故障的全量卡信息 |
| `log_items`             | List[Object] | 是   | 日志项列表             |
| `log_items[].item_type` | String       | 是   | 日志项类型             |
| `log_items[].log_lines` | List[String] | 是   | 待解析的日志行         |

##### 返回值表

| 返回值       | 类型         | 说明                             |
|--------------|--------------|----------------------------------|
| results      | List         | 清洗整合的结果                   |
| err_msg_list | List[String] | 接口执行过程中产生的错误信息列表 |

- **results 示例**

```json
[
    {
        "error_type": "AISW_MindIE_MS_HttpServer_01",
        "fault_domain": "Software",
        "attribute": {
            "key_info": "",
            "component": "MindIE",
            "module": "MS",
            "cause": "Httpserver通信超时",
            "description": "等待时间超过设定的时延。",
            "suggestion": [
                "1. 请联系华为工程师处理；"
            ]
        },
        "device_list": [
            {
                "server": "172.0.0.1",
                "device": [
                    "0", "1", "2"
                ]
            }
        ]
    }
]
```

- **results 字段说明**

| 字段                    | 类型         | 必返回 | 说明          |
|-------------------------|--------------|--------|---------------|
| `error_type`            | String       | 是     | 故障类型      |
| `fault_domain`          | String       | 是     | 故障域        |
| `attribute`             | Object       | 是     | 故障属性      |
| `attribute.key_info`    | String       | 是     | 故障关键信息  |
| `attribute.component`   | String       | 是     | 故障组件      |
| `attribute.module`      | String       | 是     | 故障模块      |
| `attribute.cause`       | String       | 是     | 故障原因      |
| `attribute.description` | String       | 是     | 故障描述      |
| `attribute.suggestion`  | List[String] | 是     | 建议方案      |
| `device_list`           | List[Object] | 是     | 故障设备列表  |
| `device_list[].server`  | String       | 是     | 故障服务器 IP |
| `device_list[].device`  | List[String] | 是     | device 卡信息 |

- **err_msg_list 示例**

```json
["Input validation failed, the reason is: [Invalid parameter type for 'input_log_list', it should be 'list'.]"]
```

### parse_root_cluster

根因节点清洗接口。

#### 接口导入

```python
from ascend_fd import parse_root_cluster
```

#### 接口定义

```python
parse_root_cluster(input_log_list: list) -> Tuple[List, List]
```

##### 请求参数表

| 参数           | 类型 | 说明                   |
|----------------|------|------------------------|
| input_log_list | List | 用户输入的节点信息列表 |

- **input_log_list 示例**

```json
[
    {
        "log_domain": {
        "server": "10.1.1.1",
        "instance_id": "instance_name"
        },
        "log_items": [
        {
            "item_type": "plog",
            "pid": 3199,
            "device_id": 0,
            "rank_id": 0,
            "log_lines": [
                "[ERROR] xxx."
            ]
        }
        ]
    }
]
```

- **input_log_list 字段说明**

| 字段                     | 类型         | 必填 | 说明           |
|--------------------------|--------------|------|----------------|
| `log_domain`             | Object       | 是   | 日志域信息     |
| `log_domain.server`      | String       | 是   | 服务器 IP      |
| `log_domain.instance_id` | String       | 是   | 实例名称       |
| `log_items`              | List[Object] | 是   | 日志项列表     |
| `log_items[].item_type`  | String       | 是   | 日志项类型     |
| `log_items[].pid`        | Integer      | 是   | 进程 ID        |
| `log_items[].device_id`  | Integer      | 否   | 设备 ID        |
| `log_items[].rank_id`    | Integer      | 否   | 通信域 ID      |
| `log_items[].log_lines`  | List[String] | 是   | 待解析的日志行 |

##### 返回值表

| 返回值       | 类型 | 说明                             |
|--------------|------|----------------------------------|
| results      | List | 清洗整合后的日志信息             |
| err_msg_list | List | 接口执行过程中产生的错误信息列表 |

### diag_root_cluster

根因节点诊断接口。

#### 接口导入

```python
from ascend_fd import diag_root_cluster
```

#### 接口定义

```python
diag_root_cluster(input_log_list: list) -> Tuple[Dict, List]
```

##### 请求参数表

| 参数           | 类型 | 说明                                   |
|----------------|------|----------------------------------------|
| input_log_list | List | parse_root_cluster 返回的 results 数据 |

##### 返回值表

| 返回值       | 类型         | 说明                             |
|--------------|--------------|----------------------------------|
| results      | Dictionary   | 发生错误的根因节点信息           |
| err_msg_list | List[String] | 接口执行过程中产生的错误信息列表 |

- **results 示例**

```json
{
    "analyze_success": true,
    "fault_description": {
        "code": 102,
        "string": "所有有效节点的Plog都没有错误日志信息，无法定位根因节点。同时请确认是否为正常的任务？"
    },
    "root_cause_device": ["ALL Device"],
    "device_link": [],
    "remote_link": "",
    "first_error_device": "",
    "last_error_device": ""
}
```

- **results 字段说明**

| 字段                       | 类型         | 必返回 | 说明                                |
|----------------------------|--------------|--------|-------------------------------------|
| `analyze_success`          | Boolean      | 是     | 是否诊断成功，true 成功，false 失败 |
| `fault_description`        | Object       | 是     | 故障描述                            |
| `fault_description.code`   | Integer      | 是     | 故障码                              |
| `fault_description.string` | String       | 是     | 故障码描述                          |
| `root_cause_device`        | List[String] | 是     | 根因设备信息                        |
| `device_link`              | List         | 是     | 根因节点链                          |
| `remote_link`              | String       | 是     | 卡间等待链                          |
| `first_error_device`       | String       | 是     | 任务中最早发生错误的 Device         |
| `last_error_device`        | String       | 是     | 任务中最晚发生错误的 Device         |

- **err_msg_list 示例**

```json
["The list of workers to be checked is empty. Please check the root cluster diag result."]
```

### parse_knowledge_graph

故障事件清洗接口。

#### 接口导入

```python
from ascend_fd import parse_knowledge_graph
```

#### 接口定义

```python
parse_knowledge_graph(input_log_list: list, custom_entity: dict = None) -> Tuple[List, List]
```

##### 请求参数表

| 参数           | 类型         | 说明                                     |
|----------------|--------------|------------------------------------------|
| input_log_list | List         | 用户输入的故障日志列表                   |
| custom_entity  | Dict（可选） | 自定义故障实体，仅本次调用有效，不会落盘 |

- **input_log_list 示例**

```json
[
    {
        "log_domain": {
            "server": "10.1.1.1"
        },
        "log_items": [
            {
                "item_type": "MindIE",
                "path": "/log/debug/mindie-ms_11_202411061400.log",
                "device_id": 0,
                "modification_time": "2025-08-21 23:50:59.999999",
                "component": "Controller",
                "log_lines": [
                    "[ERROR] xxx."
                ]
            }
        ]
    }
]
```

- **input_log_list 字段说明**

| 字段                            | 类型         | 必填 | 说明                                                                                  |
|---------------------------------|--------------|------|---------------------------------------------------------------------------------------|
| `log_domain`                    | Object       | 是   | 日志域信息                                                                            |
| `log_domain.server`             | String       | 是   | 服务器 IP                                                                             |
| `log_items`                     | List[Object] | 是   | 日志项列表                                                                            |
| `log_items[].item_type`         | String       | 是   | 日志项类型                                                                            |
| `log_items[].path`              | String       | 否   | 日志文件路径。清洗 NPU 环境检查文件（npu_info_before.txt / npu_info_after.txt）时必填 |
| `log_items[].device_id`         | Integer      | 否   | 设备 ID                                                                               |
| `log_items[].modification_time` | String       | 否   | 日志修改时间                                                                          |
| `log_items[].component`         | String       | 否   | 故障组件                                                                              |
| `log_items[].log_lines`         | List[String] | 是   | 待解析的日志行                                                                        |

- **custom_entity 示例**

```json
{
    "41001": {
    "attribute.class": "Software",
    "attribute.component": "AI Framework",
    "attribute.module": "Compiler",
    "attribute.cause_zh": "抽象类型合并失败",
    "attribute.description_zh": "对函数输出求梯度时，抽象类型不匹配，导致抽象类型合并失败。",
    "attribute.suggestion_zh": [
        "1. 检查求梯度的函数的输出类型与sens_param的类型是否相同，如果不相同，修改为相同类型；",
        "2. 自动求导报错Type Join Failed"
    ],
    "attribute.error_case": [
        "grad = ops.GradOperation(sens_param=True)",
        "# test_net输出类型为tuple(Tensor, Tensor)",
        "def test_net(a, b):",
        "    return a, b"
        ],
    "attribute.fixed_case": [
        "grad = ops.GradOperation(sens_param=True)",
        "# test_net输出类型为tuple(Tensor, Tensor)",
        "def test_net(a, b):",
        "    return a, b"
        ],
    "rule": [
        {
            "dst_code": "20106"
        }
    ],
    "source_file": "TrainLog",
    "regex.in": [
        "Abstract type", "cannot join with"
        ]
    }
}
```

- **custom_entity 字段说明**

> 故障码（顶层 key）为用户自定义的故障码，不能与 MindCluster Ascend FaultDiag 已支持的故障码相同。
> 详细字段定义请参考 [自定义故障实体](../06_api/05_command_entity.md#JSON-参数说明)。

| 字段                       | 类型         | 说明               |
|----------------------------|--------------|--------------------|
| `attribute.class`          | String       | 故障类别           |
| `attribute.component`      | String       | 故障组件           |
| `attribute.module`         | String       | 故障模块           |
| `attribute.cause_zh`       | String       | 故障原因（中文）   |
| `attribute.description_zh` | String       | 故障描述（中文）   |
| `attribute.suggestion_zh`  | List[String] | 建议方案（中文）   |
| `attribute.error_case`     | List[String] | 错误示例           |
| `attribute.fixed_case`     | List[String] | 修复示例           |
| `rule`                     | List[Object] | 诊断规则列表       |
| `source_file`              | String       | 来源文件           |
| `regex.in`                 | List[String] | 匹配的正则模式列表 |

##### 返回值表

| 返回值       | 类型         | 说明                             |
|--------------|--------------|----------------------------------|
| results      | List         | 清洗整合后相关性较高的故障事件   |
| err_msg_list | List[String] | 接口执行过程中产生的错误信息列表 |

### diag_knowledge_graph

故障事件诊断接口。

#### 接口导入

```python
from ascend_fd import diag_knowledge_graph
```

#### 接口定义

```python
diag_knowledge_graph(input_log_list: list) -> Tuple[List, List]
```

##### 请求参数表

| 参数           | 类型 | 说明                                      |
|----------------|------|-------------------------------------------|
| input_log_list | List | parse_knowledge_graph 返回的 results 数据 |

##### 返回值表

| 返回值       | 类型         | 说明                             |
|--------------|--------------|----------------------------------|
| results      | List         | 分析后的故障事件诊断报告         |
| err_msg_list | List[String] | 接口执行过程中产生的错误信息列表 |

- **results 示例**

```json
[
    {
        "analyze_success": true,
        "version_info": {},
        "note": "",
        "fault": [
            {
                "code": "NORMAL_OR_UNSUPPORTED",
                "component": "",
                "module": "",
                "cause_zh": "故障事件分析模块无结果",
                "description_zh": "故障事件分析模块无结果，可能为正常训练作业，无故障发生。如果训练任务异常中断，存在问题无法解决，请联系华为工程师处理。",
                "suggestion_zh": "1. 若存在问题无法解决，请联系华为工程师定位排查",
                "class": "",
                "fault_source": ["1.1.1.1 device-Unknown"],
                "fault_chains": []
            }
        ]
    }
]
```

- **results 字段说明**

| 字段                     | 类型         | 必返回 | 说明                                |
|--------------------------|--------------|--------|-------------------------------------|
| `analyze_success`        | Boolean      | 是     | 分析是否成功，true 成功，false 失败 |
| `version_info`           | Object       | 是     | 版本信息                            |
| `note`                   | String       | 是     | 备注                                |
| `fault`                  | List[Object] | 是     | 故障事件列表                        |
| `fault[].code`           | String       | 是     | 故障码                              |
| `fault[].component`      | String       | 是     | 故障组件                            |
| `fault[].module`         | String       | 是     | 故障模块                            |
| `fault[].cause_zh`       | String       | 是     | 故障原因（中文）                    |
| `fault[].description_zh` | String       | 是     | 故障描述（中文）                    |
| `fault[].suggestion_zh`  | String       | 是     | 故障建议（中文）                    |
| `fault[].class`          | String       | 是     | 故障类别                            |
| `fault[].fault_source`   | List[String] | 是     | 故障来源                            |
| `fault[].fault_chains`   | List         | 是     | 故障传播链                          |

- **err_msg_list 示例**

```json
["Validation for the input list[0] failed, the reason is: ParamError: input_log_list[0].server is missing",
 "Validation for the input list[2] failed, the reason is: ParamError: input_log_list[2].fault is missing"]
```
