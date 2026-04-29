# 故障事件清洗及诊断<a name="ZH-CN_TOPIC_0000002287924152"></a>

**操作步骤<a name="section13964159810"></a>**

1. 从MindCluster Ascend FaultDiag组件中，导入故障事件清洗接口、故障事件诊断接口。

    ```shell
    from ascend_fd import parse_knowledge_graph
    from ascend_fd import diag_knowledge_graph
    ```

2. 清洗故障事件。

    ```shell
    # 故障事件清洗结果与清洗过程中发生的错误
    kg_parse_results, kg_parse_err_msg = parse_knowledge_graph(input_log_list, custom_entity)
    ```

3. 诊断清洗后的故障事件。

    ```shell
    # 故障事件诊断结果与诊断过程中发生的错误
    results, err_msg_list = diag_knowledge_graph(kg_parse_results)
    ```

input_log_list输入格式如下所示，该示例不可直接使用，用户需根据实际情况修改根因节点清洗输入的相关信息。

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

**表 1** input_log_list参数说明

|字段|参数类型|是否必选|描述|
|--|--|--|--|
|log_domain|Dictionary|是|日志域。|
|server|String|是|服务器地址。|
|log_items|List|是|日志项。|
|item_type|String|是|日志类型。|
|path|String|否|日志文件路径。在清洗NPU环境检查文件npu_info_before.txt或npu_info_after.txt时，此参数必填。|
|device_id|Int|否|设备卡号。|
|modification_time|String|否|日志修改时间。在清洗训练及推理控制台日志和MindIE组件日志时，将此时间作为故障的发生时间。|
|component|String|否|组件名称。当前仅支持MindIE的Coordinator和Controller组件。|
|log_lines|List|是|待解析的日志行。|

custom_entity输入格式如下所示，该示例不可直接使用，用户需根据实际情况修改自定义故障实体的相关信息。

```text
{
    "41001": {      #故障码，用户需根据实际情况自定义故障码，不能与MindCluster Ascend FaultDiag已支持的故障码相同
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

>[!NOTE] 
>custom_entity自定义故障实体中的相关参数说明请参见[表1 参数说明](./04_customizing_fault_entities.md)。

**表 2**  err\_msg\_list参数说明

|字段|参数类型|描述|
|--|--|--|
|错误信息|List|接口执行过程中产生的错误信息。|

results输出格式示例如下。

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

**表 3**  results参数说明

|字段|参数类型|描述|
|--|--|--|
|analyze_success|Bool|诊断是否成功。<ul><li>诊断成功：True</li><li>诊断失败：False</li></ul>|
|version_info|Dictionary|版本信息。|
|note|String|备注。|
|fault|List|故障事件列表。|
|code|String|故障码。|
|component|String|故障组件。|
|module|String|故障模块。|
|cause_zh|String|故障原因。|
|description_zh|String|故障描述。|
|suggestion_zh|String|建议方案。|
|class|String|故障类别。|
|fault_source|List|故障来源。|
|fault_chains|List|故障传播链。|
