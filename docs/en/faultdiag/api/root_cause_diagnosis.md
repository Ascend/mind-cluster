# Root Cause Node Diagnosis Interface<a name="ZH-CN_TOPIC_0000002288027410"></a>

## Prototype<a name="section1652101232010"></a>

```shell
diag_root_cluster(input_log_list: list)
```

## Function<a name="section15129533192"></a>

Uses MindCluster Ascend FaultDiag to diagnose root cause node information for errors occurring in a cluster.

## Parameters<a name="section12416184719242"></a>

| Parameter | Required | Type | Description |
|--|--|--|--|
| `input_log_list` | Yes | List |`results` data obtained using the [Root Cause Node cleaning interface](./root_cause_cleaning.md). |

## Returns<a name="section14225151742812"></a>

| Parameter | Type | Description |
|--|--|--|
| `results` | Dictionary | Information about the root cause node where an error occurred |
| `err_msg_list` | List | Error information generated during interface execution |
