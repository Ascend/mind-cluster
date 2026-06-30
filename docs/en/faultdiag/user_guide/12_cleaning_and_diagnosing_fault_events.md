# Cleaning and Diagnosing Fault Events<a name="ZH-CN_TOPIC_0000002287924152"></a>

**Procedure<a name="section13964159810"></a>**

1. Import the fault event cleaning API and fault event diagnosis API from the MindCluster Ascend FaultDiag component.

    ```Python
    from ascend_fd import parse_knowledge_graph
    from ascend_fd import diag_knowledge_graph
    ```

2. Clean fault events.

    ```Python
    # Fault event cleaning results and errors that occur during the cleaning process
    kg_parse_results, kg_parse_err_msg = parse_knowledge_graph(input_log_list, custom_entity)
    ```

3. Diagnose the cleaned fault events.

    ```Python
    # Fault event diagnosis results and errors that occurred during the diagnosis process
    results, err_msg_list = diag_knowledge_graph(kg_parse_results)
    ```

The input format of input_log_list is as follows. This example cannot be used directly. Users need to modify the relevant information for root cause node cleaning input based on the actual situation.

```text
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
           '[ERROR] xxx.'
        ]
      }
    ]
  }
]
```

**Table 1** Description of input_log_list parameters

|Field|Type|Mandatory|Description|
|--|--|--|--|
|log_domain|Dictionary|Yes|Log domain.|
|server|String|Yes|Server address.|
|log_items|List|Yes|Log items.|
|item_type|String|Yes|Log type.|
|path|String|No|Log file path. This parameter is mandatory when cleaning the NPU environment check files npu_info_before.txt or npu_info_after.txt.|
|device_id|Int|No|Device card ID.|
|modification_time|String|No|Log modification time. When cleaning training and inference console logs and MindIE component logs, this time is used as the fault occurrence time.|
|component|String|No|Component name. Currently, only the Coordinator and Controller components of MindIE are supported.|
|log_lines|List|Yes|Log lines to be parsed.|

The input format of custom_entity is as follows. This example cannot be used directly. Users need to modify the relevant information of the custom fault entity based on the actual situation.

```text
{
    "41001": {      #Fault code. Users need to customize the fault code based on the actual situation. It cannot be the same as the fault codes already supported by MindCluster Ascend FaultDiag.
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
            "# test_net output type is tuple(Tensor, Tensor)",
            "def test_net(a, b):",
            "    return a, b"
              ],
        "attribute.fixed_case": [
            "grad = ops.GradOperation(sens_param=True)",
            "# test_net output type is tuple(Tensor, Tensor)",
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

> **NOTE**
>For details about the parameters in the custom_entity custom fault entity, see [Table 1 Description](./04_customizing_fault_entities.md).

**Table 2** err_msg_list Description

|Field|Type|Description|
|--|--|--|
|Error message|List|Error information generated during API execution.|

The following is an example of the results output format.

```text
[
    {
        'analyze_success': True,
        'version_info': {},
        'note': '',
        'fault': [{
                'code': 'NORMAL_OR_UNSUPPORTED',
                'component': '',
                'module': '',
                'cause_zh': '故障事件分析模块无结果',
                'description_zh': '故障事件分析模块无结果，可能为正常训练作业，无故障发生。如果训练任务异常中断，存在问题无法解决，请联系华为工程师处理。',
                'suggestion_zh': '1. 若存在问题无法解决，请联系华为工程师定位排查',
                'class': '',
                'fault_source': ['1.1.1.1 device-Unknown'],
                'fault_chains': []
            }
        ]
    }
]
```

**Table 3** Description of results parameters

|Field|Type|Description|
|--|--|--|
|analyze_success|Bool|Whether the diagnosis is successful.<ul><li>Diagnosis successful: True</li><li>Diagnosis failed: False</li></ul>|
|version_info|Dictionary|Version information.|
|note|String|Remarks.|
|fault|List|List of fault events.|
|code|String|Fault code.|
|component|String|Faulty component.|
|module|String|Faulty module.|
|cause_zh|String|Fault cause.|
|description_zh|String|Fault description.|
|suggestion_zh|String|Suggested solution.|
|class|String|Fault category.|
|fault_source|List|Fault source.|
|fault_chains|List|Fault propagation chain.|
