# Fault Event Cleaning Interface<a name="ZH-CN_TOPIC_0000002322636661"></a>

## Prototype<a name="section1652101232010"></a>

```shell
parse_knowledge_graph(input_log_list: list, custom_entity: dict = None)
```

## Function<a name="section15129533192"></a>

Cleans fault logs.

## Parameters<a name="section12416184719242"></a>

|Parameter |Required|Type|Description|
|--|--|--|--|
|`input_log_list`|Yes|List|Fault logs input by the user.|
|`custom_entity`|No|Dictionary|Custom fault entity input by the user. This parameter is for temporary use and will not be persisted to the JSON file.|

## Returns<a name="section14225151742812"></a>

|Parameter|Type|Description|
|--|--|--|
|`results`|List|Results of log cleaning and integration.|
|`err_msg_list`|List|Error messages generated during the interface execution.|
