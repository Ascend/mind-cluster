# SDK API Reference

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-24T02:30:28.842Z pushedAt=2026-08-24T02:51:43.307Z -->

## Invocation Description

When the SDK is used, operation logs and runtime logs are generated in the `$HOME/.ascend_faultdiag` directory. The directory structure is as follows:

```text
$HOME/.ascend_faultdiag
└── ascend_faultdiag_operation.log    # Operation log
└── RUN_LOG                           # Run log
  └─ 20241104142355468743_6797877f-7143-443f-a9c6-361e33032c5c
```

> [!NOTE]
>
> - The size of a log file does not exceed 10 MB. When the size limit is exceeded, the log is automatically dumped to another log file.
> - The number of log files with the same PID does not exceed 10. When the number limit is exceeded, the earliest created log is automatically overwritten.

## SDK Interface Definitions

### parse_fault_type

Service log parsing interface.

#### Interface Import

```python
from ascend_fd import parse_fault_type
```

#### Interface Definition

```python
parse_fault_type(input_log_list: list) -> Tuple[List, List]
```

##### Request Parameter

| Parameter      | Type | Description                          |
|----------------|------|--------------------------------------|
| `input_log_list` | List | List of service logs entered by the user |

**input_log_list Example**

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

**input_log_list Field Description**

| Field                    | Type         | Mandatory | Description                          |
|--------------------------|--------------|-----------|--------------------------------------|
| `log_domain`             | Object       | Yes       | Log domain information               |
| `log_domain.server`      | String       | Yes       | Server IP                            |
| `log_domain.port`        | Integer      | Yes       | Server port                          |
| `log_domain.device`      | List[String] | Yes       | Information about all devices on which faults occurred |
| `log_items`              | List[Object] | Yes       | Log item list                        |
| `log_items[].item_type`  | String       | Yes       | Log item type                        |
| `log_items[].log_lines`  | List[String] | Yes       | Log lines to be parsed               |

##### Return Values

| Return Value | Type         | Description                             |
|--------------|--------------|-----------------------------------------|
| `results`      | List         | Result of parsing and integration operations      |
| `err_msg_list` | List[String] | List of errors generated during interface execution |

**results Example**

```json
[
    {
        "error_type": "AISW_MindIE_MS_HttpServer_01",
        "fault_domain": "Software",
        "attribute": {
            "key_info": "",
            "component": "MindIE",
            "module": "MS",
            "cause": "Httpserver communication timeout",
            "description": "The waiting time exceeds the configured delay.",
            "suggestion": [
                "1. Contact Huawei engineers for handling."
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

**results Field Description**

| Field                    | Type         | Mandatory Return | Description          |
|-------------------------|--------------|--------|---------------|
| `error_type`            | String       | Yes     | Fault type      |
| `fault_domain`          | String       | Yes     | Fault domain        |
| `attribute`             | Object       | Yes     | Fault attribute      |
| `attribute.key_info`    | String       | Yes     | Fault key information  |
| `attribute.component`   | String       | Yes     | Fault component      |
| `attribute.module`      | String       | Yes     | Fault module      |
| `attribute.cause`       | String       | Yes     | Fault cause      |
| `attribute.description` | String       | Yes     | Fault description      |
| `attribute.suggestion`  | List[String] | Yes     | Suggested solution      |
| `device_list`           | List[Object] | Yes     | Fault device list  |
| `device_list[].server`  | String       | Yes     | Fault server IP |
| `device_list[].device`  | List[String] | Yes     | Device information |

**err_msg_list Example**

```json
["Input validation failed, the reason is: [Invalid parameter type for 'input_log_list', it should be 'list'.]"]
```

### parse_root_cluster

Root cause node parsing interface.

#### Interface Import

```python
from ascend_fd import parse_root_cluster
```

#### Interface Definition

```python
parse_root_cluster(input_log_list: list) -> Tuple[List, List]
```

##### Request Parameter

| Name           | Type | Description                        |
|----------------|------|------------------------------------|
| `input_log_list` | List | List of node information entered by the user |

**input_log_list Example**

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

**input_log_list Field Description**

| Field                    | Type         | Mandatory | Description              |
|--------------------------|--------------|-----------|--------------------------|
| `log_domain`             | Object       | Yes       | Log domain information   |
| `log_domain.server`      | String       | Yes       | Server IP                |
| `log_domain.instance_id` | String       | Yes       | Instance name            |
| `log_items`              | List[Object] | Yes       | Log item list            |
| `log_items[].item_type`  | String       | Yes       | Log item type            |
| `log_items[].pid`        | Integer      | Yes       | Process ID               |
| `log_items[].device_id`  | Integer      | No        | Device ID                |
| `log_items[].rank_id`    | Integer      | No        | Communication domain ID  |
| `log_items[].log_lines`  | List[String] | Yes       | Log lines to be parsed   |

##### Return Values

| Return Value | Type         | Description                             |
|--------------|--------------|-----------------------------------------|
| `results`      | List         | Log information after parsing and integration operations |
| `err_msg_list` | List[String] | List of errors generated during interface execution |

**results Example**

Each element in the list corresponds to the parsing result of one server, with the key being the PID string. If the input contains MindIE logs, a MindIE parsing result object is appended to the end of the list.

```json
[
    {
        "3199": {
            "pid": "3199",
            "base": {
                "device_ip": "",
                "vNic_ip": "",
                "logic_device_id": "0",
                "phy_device_id": "0",
                "server_id": "10.1.1.1",
                "root_list": ["instance_name"],
                "timeout_param": {},
                "rank_info_list": [],
                "rank_map": {
                    "instance_name": {
                        "rank_id": "0",
                        "rank_num": 1,
                        "identifier": "",
                        "eid_plane_list": []
                    }
                },
                "server_name": "",
                "generation_info": ""
            },
            "error": {
                "first_error_module": "Notify",
                "first_error_time": "2026-05-29 15:35:10.674067",
                "cqe_links": [],
                "timeout_error_events_list": [],
                "cluster_exception": {},
                "transport_error_remote": null,
                "transport_init_error_happened": false
            },
            "tls_status": "",
            "start_train_time": "2026-05-29 15:30:00.000000",
            "end_train_time": "2026-05-29 15:40:00.000000",
            "lagging_time": "2026-05-29 00:00:00.000000",
            "recovery_success_time": "",
            "start_resumable_training_time": "",
            "plog_parsed_name": "plog-parser-3199-1.log",
            "show_logs": {
                "error": ["[ERROR] xxx"],
                "normal": []
            },
            "aicpu_notify_wait_remote": ""
        }
    },
    {
        "mindie": true,
        "link_error_info_map": {
            "10.1.1.1": ["10.1.1.2"]
        },
        "pull_kv_error_map": {
            "10.1.1.1": ["10.1.1.2"]
        }
    }
]
```

**results Field Description**

| Field                                              | Type         | Mandatory Return | Description                                          |
|----------------------------------------------------|--------------|------------------|------------------------------------------------------|
| `<pid>`                                            | Object       | Yes              | Parsed result object keyed by pid                    |
| `<pid>.pid`                                        | String       | Yes              | Process ID                                           |
| `<pid>.base`                                       | Object       | Yes              | Basic device information parsed from plog            |
| `<pid>.base.device_ip`                             | String       | Yes              | Device IP                                            |
| `<pid>.base.vNic_ip`                               | String       | Yes              | Virtual NIC IP                                       |
| `<pid>.base.logic_device_id`                       | String       | Yes              | Logical device ID                                    |
| `<pid>.base.phy_device_id`                         | String       | Yes              | Physical device ID                                   |
| `<pid>.base.server_id`                             | String       | Yes              | Server IP                                            |
| `<pid>.base.root_list`                             | List[String] | Yes              | List of root node identifiers                        |
| `<pid>.base.timeout_param`                         | Object       | Yes              | Timeout                                  |
| `<pid>.base.rank_map`                              | Object       | Yes              | Communication domain information, keyed by `instance_id` |
| `<pid>.base.rank_map.<instance_id>.rank_id`        | String       | Yes              | Rank ID within the communication domain              |
| `<pid>.base.rank_map.<instance_id>.rank_num`       | Integer      | Yes              | Total number of ranks within the communication domain |
| `<pid>.base.rank_map.<instance_id>.identifier`     | String       | Yes              | Communication domain identifier                      |
| `<pid>.base.rank_map.<instance_id>.eid_plane_list` | List         | Yes              | EID plane information list                           |
| `<pid>.base.server_name`                           | String       | Yes              | Server name                                          |
| `<pid>.base.generation_info`                       | String       | Yes              | Product generation information                       |
| `<pid>.error`                                      | Object       | Yes              | Error information                                    |
| `<pid>.error.first_error_module`                   | String       | Yes              | First error module                                   |
| `<pid>.error.first_error_time`                     | String       | Yes              | Time when the first error occurred                   |
| `<pid>.error.cqe_links`                            | List[String] | Yes              | CQE link information                                 |
| `<pid>.error.timeout_error_events_list`            | List         | Yes              | List of timeout error events                         |
| `<pid>.error.cluster_exception`                    | Object       | Yes              | Cluster exception information                        |
| `<pid>.error.transport_error_remote`               | Object       | Yes              | Remote peer information of the transport error       |
| `<pid>.error.transport_init_error_happened`        | Boolean      | Yes              | Whether a transport initialization error has occurred |
| `<pid>.tls_status`                                 | String       | Yes              | TLS status                                           |
| `<pid>.start_train_time`                           | String       | Yes              | Training start time                                  |
| `<pid>.end_train_time`                             | String       | Yes              | Training end time                                    |
| `<pid>.lagging_time`                               | String       | Yes              | Lagging time                                         |
| `<pid>.recovery_success_time`                      | String       | Yes              | Recovery success time                                |
| `<pid>.start_resumable_training_time`              | String       | Yes              | Resumable training start time                        |
| `<pid>.plog_parsed_name`                           | String       | Yes              | Name of the generated plog parsing file              |
| `<pid>.show_logs`                                  | Object       | Yes              | Display logs                                         |
| `<pid>.show_logs.error`                            | List[String] | Yes              | Error logs (up to several lines displayed)           |
| `<pid>.show_logs.normal`                           | List[String] | Yes              | Normal logs (up to several lines displayed)          |
| `<pid>.aicpu_notify_wait_remote`                   | String       | Yes              | Remote peer information for AICPU notify waiting     |
| `mindie`                                           | Boolean      | No               | Whether it is a MindIE parsing result (present only when MindIE logs are included) |
| `link_error_info_map`                              | Object       | No               | Link error information, keyed by local IP with the value being the list of remote IPs |
| `pull_kv_error_map`                                | Object       | No               | KV pull error information, keyed by local IP with the value being the list of remote IPs |

**err_msg_list Example**

```json
["Input validation failed, the reason is: [Invalid parameter type for 'input_log_list', it should be 'list'.]"]
```

### diag_root_cluster

Root cause node diagnosis interface.

#### Interface Import

```python
from ascend_fd import diag_root_cluster
```

#### Interface Definition

```python
diag_root_cluster(input_log_list: list) -> Tuple[Dict, List]
```

##### Request Parameter

| Name           | Type | Description                                   |
|----------------|------|-----------------------------------------------|
| `input_log_list` | List | Results data returned by `parse_root_cluster` |

##### Return Values

| Return Value | Type         | Description                                    |
|--------------|--------------|------------------------------------------------|
| `results`      | Dict         | Root cause node information where errors occur |
| `err_msg_list` | List[String] | List of errors generated during interface execution |

**results Example**

```json
{
    "analyze_success": true,
    "fault_description": {
        "code": 102,
        "string": "No error log information is found in the Plog of any valid node, so the root cause node cannot be located. Confirm whether this is a normal task."
    },
    "root_cause_device": ["ALL Device"],
    "device_link": [],
    "remote_link": "",
    "first_error_device": "",
    "last_error_device": ""
}
```

**results Field Description**

| Field                       | Type         | Mandatory Return | Description                                |
|----------------------------|--------------|--------|-------------------------------------|
| `analyze_success`          | Boolean      | Yes     | Whether the diagnosis is successful.<ul><li>`true`: Success</li><li>`false`: Failure</li></ul>|
| `fault_description`        | Object       | Yes     | Fault description                            |
| `fault_description.code`   | Integer      | Yes     | Fault code                              |
| `fault_description.string` | String       | Yes     | Fault code description                          |
| `root_cause_device`        | List[String] | Yes     | Root cause device information                        |
| `device_link`              | List         | Yes     | Root cause node chain                          |
| `remote_link`              | String       | Yes     | Inter-device waiting chain                          |
| `first_error_device`       | String       | Yes     | Device where the earliest error occurs in the task         |
| `last_error_device`        | String       | Yes     | Device where the latest error occurs in the task         |

**err_msg_list Example**

```json
["The list of workers to be checked is empty. Please check the root cluster diag result."]
```

### parse_knowledge_graph

Fault event parsing interface.

#### Interface Import

```python
from ascend_fd import parse_knowledge_graph
```

#### Interface Definition

```python
parse_knowledge_graph(input_log_list: list, custom_entity: dict = None) -> Tuple[List, List]
```

##### Request Parameters

| Name           | Type | Mandatory | Description                                     |
|----------------|------|-----------|-------------------------------------------------|
| `input_log_list` | List | Yes       | List of fault logs entered by the user          |
| `custom_entity`  | Dict | No        | Custom fault entity, valid only for this call and not persisted to disk |

**input_log_list Example**

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

**input_log_list Field Description**

| Field                          | Type         | Mandatory | Description                                                                                  |
|---------------------------------|--------------|-----------|----------------------------------------------------------------------------------------------|
| `log_domain`                    | Object       | Yes       | Log domain information                                                                       |
| `log_domain.server`             | String       | Yes       | Server IP                                                                                    |
| `log_items`                     | List[Object] | Yes       | Log item list                                                                                |
| `log_items[].item_type`         | String       | Yes       | Log item type                                                                                |
| `log_items[].path`              | String       | No        | Log file path. Mandatory when parsing NPU environment check files (`npu_info_before.txt`/`npu_info_after.txt`) |
| `log_items[].device_id`         | Integer      | No        | Device ID                                                                                    |
| `log_items[].modification_time` | String       | No        | Log modification time                                                                        |
| `log_items[].component`         | String       | No        | Fault component                                                                              |
| `log_items[].log_lines`         | List[String] | Yes       | Log lines to be parsed                                                                       |

**custom_entity Example**

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
        "# The output type of test_net is tuple(Tensor, Tensor).",
        "def test_net(a, b):",
        "    return a, b"
        ],
    "attribute.fixed_case": [
        "grad = ops.GradOperation(sens_param=True)",
        "# The output type of test_net is tuple(Tensor, Tensor).",
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

**custom_entity Field Description**

| Field                      | Type               | Description                          |
|----------------------------|--------------------|--------------------------------------|
| `attribute.class`          | String             | Fault category                       |
| `attribute.component`      | String             | Fault component                      |
| `attribute.module`         | String             | Fault module                         |
| `attribute.cause_zh`       | String             | Fault cause (Chinese)                |
| `attribute.description_zh` | String/List[String] | Fault description (Chinese)          |
| `attribute.suggestion_zh`  | String/List[String] | Suggested solution (Chinese)         |
| `attribute.error_case`     | String/List[String] | Error example                        |
| `attribute.fixed_case`     | String/List[String] | Correction example                        |
| `rule`                     | List[Object]       | Diagnostic rule list                 |
| `source_file`              | String             | Source file                          |
| `regex.in`                 | List[String]       | List of matching regular expressions |

> [!NOTE]
>
> - The fault code (top-level key) is a user-defined fault code and cannot be the same as currently supported fault codes. For supported fault codes, see [Supported Faults](../07_references/04_appendix.md#known-faults).
> - For detailed field definitions, see [Custom Fault Entities](../06_api/05_command_entity.md#json-parameter-description).

##### Return Values

| Return Value | Type         | Description                                        |
|--------------|--------------|----------------------------------------------------|
| `results`      | List         | Fault events with high relevance after cleaning and integration |
| `err_msg_list` | List[String] | List of errors generated during interface execution |

**results Example**

> [!NOTE]
>
> Each element in the list corresponds to the cleaning result of one server, including the fault event analysis (`root_causes`) of each device on that server.

```json
[
    {
        "server": "10.1.1.1",
        "fault": [
            {
                "parse_version": "26.1.0",
                "response": {
                    "0": {
                        "analyze_success": true,
                        "error": "None",
                        "root_causes": {
                            "AISW_MindIE_MS_HttpServer_01": {
                                "code": "AISW_MindIE_MS_HttpServer_01",
                                "entities_attribute": {
                                    "component": "MindIE",
                                    "module": "MS",
                                    "cause_zh": "Httpserver通信超时",
                                    "description_zh": "等待时间超过设定的时延。",
                                    "suggestion_zh": [
                                        "1. 请联系华为工程师处理;"
                                    ],
                                    "class": "Software"
                                },
                                "events_attribute": [
                                    {
                                        "event_code": "AISW_MindIE_MS_HttpServer_01",
                                        "key_info": "[ERROR] [MIE03E400008] [HttpServer] Http server on timeout.",
                                        "type": "MindIE_Controller",
                                        "source_device": "0",
                                        "occur_time": "2024-11-05 12:00:00.123456",
                                        "event_id": "key1"
                                    }
                                ],
                                "chains": {}
                            }
                        }
                    }
                }
            }
        ]
    }
]
```

**results Field Description**

| Field                                                   | Type         | Mandatory Return | Description                                       |
|--------------------------------------------------------|--------------|--------|--------------------------------------------|
| `server`                                               | String       | Yes     | Server IP                                  |
| `fault`                                                | List[Object] | Yes     | Fault analysis result list (each element corresponds to one parsing)   |
| `fault[].parse_version`                                | String       | Yes     | Parser version number                               |
| `fault[].response`                                     | Object       | Yes     | Fault event analysis of each device, with the key being source_device |
| `fault[].response.<source_device>.analyze_success`     | Boolean      | Yes     | Whether the analysis succeeds.<ul><li>`true`: success</li><li>`false`: failure</li></ul>        |
| `fault[].response.<source_device>.error`               | String       | Yes     | Error message (`"None"` when there is no error)              |
| `fault[].response.<source_device>.root_causes`         | Object       | Yes     | Root cause event dictionary, with the key being the fault code            |
| `root_causes.<code>.code`                              | String       | Yes     | Fault code                                     |
| `root_causes.<code>.entities_attribute`                | Object       | Yes     | Fault entity attribute                               |
| `root_causes.<code>.entities_attribute.component`      | String       | Yes     | Fault component                                   |
| `root_causes.<code>.entities_attribute.module`         | String       | Yes     | Fault module                                   |
| `root_causes.<code>.entities_attribute.cause_zh`       | String       | Yes     | Fault cause (Chinese)                           |
| `root_causes.<code>.entities_attribute.description_zh` | String       | Yes     | Fault description (Chinese)                           |
| `root_causes.<code>.entities_attribute.suggestion_zh`  | List[String] | Yes     | Suggested solution (Chinese)                           |
| `root_causes.<code>.entities_attribute.class`          | String       | Yes     | Fault category                                   |
| `root_causes.<code>.events_attribute`                  | List[Object] | Yes     | List of events that trigger this root cause                       |
| `root_causes.<code>.events_attribute[].event_code`     | String       | Yes     | Event fault code                                 |
| `root_causes.<code>.events_attribute[].key_info`       | String       | Yes     | Key log information of the event                           |
| `root_causes.<code>.events_attribute[].type`           | String       | Yes     | Event type                                   |
| `root_causes.<code>.events_attribute[].source_device`  | String       | Yes     | Source device of the event                               |
| `root_causes.<code>.events_attribute[].occur_time`     | String       | Yes     | Event occurrence time                               |
| `root_causes.<code>.events_attribute[].event_id`       | String       | Yes     | Unique event identifier                               |
| `root_causes.<code>.chains`                            | Object       | Yes     | Fault propagation chain                                 |

**err_msg_list Example**

```json
["Validation for the input list[0] failed, the reason is: ParamError: input_log_list[0].server is missing"]
```

### diag_knowledge_graph

Fault event diagnosis interface.

#### Interface Import

```python
from ascend_fd import diag_knowledge_graph
```

#### Interface Definition

```python
diag_knowledge_graph(input_log_list: list) -> Tuple[List, List]
```

##### Request Parameter

| Name           | Type | Description                              |
|----------------|------|------------------------------------------|
| `input_log_list` | List | Results data returned by `parse_knowledge_graph` |

##### Return Values

| Return Value | Type         | Description                                    |
|--------------|--------------|------------------------------------------------|
| `results`      | List         | Fault event diagnostic report after analysis    |
| `err_msg_list` | List[String] | List of errors generated during the interface execution |

**results Example**

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

**results Field Description**

| Field                     | Type         | Mandatory Return | Description                                |
|--------------------------|--------------|--------|-------------------------------------|
| `analyze_success`        | Boolean      | Yes     | Whether the analysis succeeds.<ul><li>`true`: Success</li><li>`false`: Failure</li></ul> |
| `version_info`           | Object       | Yes     | Version information                            |
| `note`                   | String       | Yes     | Note                                |
| `fault`                  | List[Object] | Yes     | Fault event list                        |
| `fault[].code`           | String       | Yes     | Fault code                              |
| `fault[].component`      | String       | Yes     | Fault component                            |
| `fault[].module`         | String       | Yes     | Fault module                            |
| `fault[].cause_zh`       | String       | Yes     | Fault cause (Chinese)                    |
| `fault[].description_zh` | String       | Yes     | Fault description (Chinese)                    |
| `fault[].suggestion_zh`  | String       | Yes     | Fault suggestion (Chinese)                    |
| `fault[].class`          | String       | Yes     | Fault category                            |
| `fault[].fault_source`   | List[String] | Yes     | Fault source                            |
| `fault[].fault_chains`   | List         | Yes     | Fault propagation chain                          |

**err_msg_list Example**

```json
["Validation for the input list[0] failed, the reason is: ParamError: input_log_list[0].server is missing",
 "Validation for the input list[2] failed, the reason is: ParamError: input_log_list[2].fault is missing"]
```
