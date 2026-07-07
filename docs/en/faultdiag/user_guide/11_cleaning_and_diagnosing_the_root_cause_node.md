# Cleaning and Diagnosing the Root Cause Node<a name="ZH-CN_TOPIC_0000002322636653"></a>

**Procedure<a name="section13964159810"></a>**

1. Import the root cause node cleaning and diagnosis interfaces from the MindCluster Ascend FaultDiag component.

    ```Python
    from ascend_fd import parse_root_cluster
    from ascend_fd import diag_root_cluster
    ```

2. Clean the root cause node.

    ```Python
    # Root cause node cleaning result and errors that occur during the cleaning
    rc_parse_results, rc_parse_err_msg = parse_root_cluster(input_log_list)
    ```

3. Diagnose the cleaned root cause node.

    ```Python
    # Root cause node diagnosis result and errors that occur during the diagnosis
    results, err_msg_list = diag_root_cluster(rc_parse_results)
    ```

The following is an example of the `input_log_list` format, which is for reference only. You need to modify the input information of the root cause node cleaning as required.

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

**Table 1** `input_log_list` description

|Field|Type|Mandatory (Yes/No)|Description|
|--|--|--|--|
|log_domain|Dictionary|Yes|Log domain|
|server|String|Yes|Server address|
|instance_id|String|Yes|Instance name|
|log_items|List|Yes|Log item|
|item_type|String|Yes|Log type|
|pid|Int|Yes|Process ID|
|device_id|Int|No|Device ID|
|rank_id|Int|No|Communicator's rank ID|
|log_lines|List|Yes|Log line to be parsed|

**Table 2** `err_msg_list` description

|Field|Type|Description|
|--|--|--|
|Error message|List|Error message generated during interface execution|

Example of the `results` output format:

```text
{
    'analyze_success': True,
    'fault_description': {
        'code': 102,
        'string': 'No error log is found in the Plog of all valid nodes. The root cause node cannot be located. Check whether the task is normal.'
    },
    'root_cause_device': ['ALL Device'],
    'device_link': [],
    'remote_link': '',
    'first_error_device': '',
    'last_error_device': ''
}
```

**Table 3** `results` description

|Field|Type|Description|
|--|--|--|
|analyze_success|Bool|Specifies whether the diagnosis is successful. <ul><li>`True`: success</li><li>`False`: failure</li></ul>|
|fault_description|Dictionary|Fault description|
|code|Int|Fault code|
|string|String|Fault code description|
|root_cause_device|List|Root cause device information|
|device_link|List|Root cause node chain|
|remote_link|String|Inter-device waiting chain|
|first_error_device|String|Device where the first fault occurs|
|last_error_device|String|Device where the latest fault occurs|
