# Service Stream Cleaning Interface<a name="ZH-CN_TOPIC_0000002256436654"></a>

## Prototype<a name="section1652101232010"></a>

```shell
parse_fault_type(input_log_list: list)
```

## Function <a name="section15129533192"></a>

Uses the fault mode defined by MindCluster Ascend FaultDiag to clean service stream logs.

## Parameters<a name="section12416184719242"></a>

|Parameter |Required|Type|Description|
|--|--|--|--|
|`input_log_list`|Yes|List|Service stream input by the user.|

## Returns<a name="section14225151742812"></a>

|Parameter |Type|Description|
|--|--|--|
|`results`|List|Integrated cleaning results|
|`err_msg_list`|List|Error messages generated during interface execution|
