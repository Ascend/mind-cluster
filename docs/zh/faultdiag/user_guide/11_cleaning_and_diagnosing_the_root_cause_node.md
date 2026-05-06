# 根因节点清洗及诊断<a name="ZH-CN_TOPIC_0000002322636653"></a>

**操作步骤<a name="section13964159810"></a>**

1. 从MindCluster Ascend FaultDiag组件中，导入根因节点清洗接口、根因节点诊断接口。

    ```shell
    from ascend_fd import parse_root_cluster
    from ascend_fd import diag_root_cluster
    ```

2. 清洗根因节点。

    ```shell
    # 根因节点清洗结果与清洗过程中发生的错误
    rc_parse_results, rc_parse_err_msg = parse_root_cluster(input_log_list)
    ```

3. 诊断清洗后的根因节点。

    ```shell
    # 根因节点诊断结果与诊断过程中发生的错误
    results, err_msg_list = diag_root_cluster(rc_parse_results)
    ```

input_log_list输入格式如下所示，该示例不可直接使用，用户需根据实际情况修改根因节点清洗输入的相关信息。

```text
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
|instance_id|String|是|实例名称。|
|log_items|List|是|日志项。|
|item_type|String|是|日志类型。|
|pid|Int|是|进程号。|
|device_id|Int|否|设备卡号。|
|rank_id|Int|否|通信域卡号。|
|log_lines|List|是|待解析的日志行。|

**表 2**  err_msg_list参数说明

|字段|参数类型|描述|
|--|--|--|
|错误信息|List|接口执行过程中产生的错误信息。|

results输出格式示例如下。

```text
{
    'analyze_success': True,
    'fault_description': {
        'code': 102,
        'string': '所有有效节点的Plog都没有错误日志信息，无法定位根因节点。同时请确认是否为正常的任务？'
    },
    'root_cause_device': ['ALL Device'],
    'device_link': [],
    'remote_link': '',
    'first_error_device': '',
    'last_error_device': ''
}
```

**表 3**  results参数说明

|字段|参数类型|描述|
|--|--|--|
|analyze_success|Bool|诊断是否成功。<ul><li>诊断成功：True</li><li>诊断失败：False</li></ul>|
|fault_description|Dictionary|故障描述。|
|code|Int|故障码。|
|string|String|故障码描述。|
|root_cause_device|List|根因设备信息。|
|device_link|List|根因节点链。|
|remote_link|String|卡间等待链。|
|first_error_device|String|任务中最早发生错误的Device。|
|last_error_device|String|任务中最晚发生错误的Device。|
