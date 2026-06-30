# Cleaning Service Flow Logs<a name="ZH-CN_TOPIC_0000002256402980"></a>

**Procedure<a name="section13122026153211"></a>**

1. Import the service flow cleaning interface from the MindCluster Ascend FaultDiag component.

    ```Python
    from ascend_fd import parse_fault_type
    ```

2. Clean the service flow logs.

    ```Python
    results, err_msg_list = parse_fault_type(input_log_list)
    ```

The input format of input_log_list is as follows. This example cannot be used directly. Users need to modify the relevant information of the service flow input based on the actual situation.

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

**Table 1** input_log_list parameter description

|Parameter|Type|Mandatory|Description|
|--|--|--|--|
|log_domain|Dictionary|Yes|Log domain.|
|server|String|Yes|Server address.|
|device|List|Yes|Information about all faulty cards.|
|log_items|List|Yes|Log items.|
|item_type|String|Yes|Log type.|
|log_lines|List|Yes|Log lines to be parsed.|

**Table 2** err_msg_list parameter description

|Parameter|Type|Description|
|--|--|--|
|Error information|List|Error information generated during interface execution.|

The following is an example of the results output format.

```text
[
    {
        "error_type": "AISW_MindIE_MS_HttpServer_01",
        "fault_domain": "Software",
        "attribute": {
        "key_info": "",
            "component": "MindIE",
            "module": "MS",
            "cause": "Httpserver Communication Timeout",
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

**Table 3** Description of results parameters

|Field|Type|Description|
|--|--|--|
|error_type|String|Fault code.|
|fault_domain|String|Fault domain.|
|attribute|Dictionary|Fault attributes.|
|key_info|String|Key logs.|
|component|String|Faulty component.|
|cause|String|Fault cause.|
|description|String|Fault description.|
|suggestion|String|Suggested solution.|
|device_list|List|List of devices where this fault occurs.|
|server|String|Server address.|
|device|List|Device information.|
