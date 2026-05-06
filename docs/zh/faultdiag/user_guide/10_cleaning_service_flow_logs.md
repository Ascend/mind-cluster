# 清洗业务流日志<a name="ZH-CN_TOPIC_0000002256402980"></a>

**操作步骤<a name="section13122026153211"></a>**

1. 从MindCluster Ascend FaultDiag组件中，导入业务流清洗接口。

    ```shell
    from ascend_fd import parse_fault_type
    ```

2. 清洗业务流日志。

    ```shell
    results, err_msg_list = parse_fault_type(input_log_list)
    ```

input_log_list输入格式如下所示，该示例不可直接使用，用户需根据实际情况修改业务流输入的相关信息。

```text
[
    {
        "log_domain": {
            "server": "10.1.1.1",
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
    },
 ...
]
```

**表 1** input_log_list参数说明

|字段|参数类型|是否必选|描述|
|--|--|--|--|
|log_domain|Dictionary|是|日志域。|
|server|String|是|服务器地址。|
|device|List|是|发生过故障的全量卡信息。|
|log_items|List|是|日志项。|
|item_type|String|是|日志类型。|
|log_lines|List|是|待解析的日志行。|

**表 2**  err_msg_list参数说明

|字段|参数类型|描述|
|--|--|--|
|错误信息|List|接口执行过程中产生的错误信息。|

results输出格式示例如下。

```text
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

**表 3**  results参数说明

|字段|参数类型|描述|
|--|--|--|
|error_type|String|故障码。|
|fault_domain|String|故障领域。|
|attribute|Dictionary|故障属性。|
|key_info|String|关键日志。|
|component|String|故障组件。|
|cause|String|故障原因。|
|description|String|故障描述。|
|suggestion|String|建议方案。|
|device_list|List|发生该故障的设备列表。|
|server|String|服务器地址。|
|device|List|device卡信息。|
