# Fault Diagnosis Interface<a name="ZH-CN_TOPIC_0000001541948914"></a>

## Prototype<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>

```shell
ascend-fd diag -i <Diagnosis_input_directory> -o <Diagnosis_result_output_directory>
```

## Function<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>

Starts a fault diagnosis task and diagnoses fault events based on the log cleaning results after log cleaning is complete.

## Parameters<a name="zh-cn_topic_0000001511538701_section122149111390"></a>

**Table 1** Parameters

|Parameter|Abbreviation|Required|Value Type|Description|
|--|--|--|--|--|
|`--input_path`|`-i`|Yes|String|Input path for cleaned data.|
|`--output_path`|`-o`|Yes|String|Output path for diagnostic results.|
|`--performance`|`-p`|No|Boolean|When specified, all diagnostic modules will be executed. If not specified, only the fault diagnosis functions of the root cause node and fault event modules will be executed.|
|`--help`|`-h`|No|-|Queries the meaning of level-2 commands and parameters, as well as usage instructions.|
|`--scene`|`-s`|No|String|Diagnosis scenario, defaulted to `host`. Options:<ul><li>`host`: Scenario for diagnosing individual host logs.</li><li>`super_pod`: Scenario for diagnosing SuperPoD logs, including host, BMC, and LCNE logs.</li></ul>|

## Returns<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>

Execution status of the fault diagnosis task:

```ColdFusion
The diag job starts. Please wait. Job id: [****], run log file is [****].
Diagnostic information
The diag job is complete.
```
