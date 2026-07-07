# Fault Event Diagnosis Interface<a name="ZH-CN_TOPIC_0000002287924160"></a>

## Prototype<a name="section1652101232010"></a>

```shell
diag_knowledge_graph(input_log_list: list)
```

## Function<a name="section15129533192"></a>

Diagnoses cleaned fault events and outputs a diagnosis report.

## Parameters<a name="section12416184719242"></a>

|Name|Required| Type|Description|
|--|--|--|--|
|`input_log_list`|Yes|List|`results` data from each node obtained using the [Fault Event Cleaning Interface](./fault_event_cleaning.md).<p>Note: If `"source": "ccae"` appears in the node parameters, the cleaning results for that node may be inaccurate.</p>|

## Returns<a name="section14225151742812"></a>

|Parameter|Type|Description|
|--|--|--|
|`results`|List|Integrated fault event diagnosis report|
|`err_msg_list`|List|Error messages generated during interface execution|
