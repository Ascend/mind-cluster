# Root Cause Cleaning Interface<a name="ZH-CN_TOPIC_0000002322563625"></a>

## Prototype<a name="section1652101232010"></a>

```shell
parse_root_cluster(input_log_list: list)
```

## Function<a name="section15129533192"></a>

Uses MindCluster Ascend FaultDiag to clean node information in a cluster.

## Parameters<a name="section12416184719242"></a>

|Parameter|Required|Type|Description|
|--|--|--|--|
|`input_log_list`|Yes|List|Node information input by the user.|

## Returns<a name="section14225151742812"></a>

|Parameter|Type|Description|
|--|--|--|
|`results`|List|Integrated cleaning results|
|`err_msg_list`|List|Error messages generated during interface execution|
