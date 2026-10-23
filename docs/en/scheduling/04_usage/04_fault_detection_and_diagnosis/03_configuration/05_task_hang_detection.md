# Configuring Workload Fault and Task Hang Detection<a name="ZH-CN_TOPIC_0000002479387566"></a>

## Configuration File Description<a name="ZH-CN_TOPIC_0000002479387566_section_specifications "></a>

Task hang faults are detected by Ascend Device Plugin. This feature involves the following two configuration files:

- `faultCode.json`: Configures the fault level of task hang faults.
- `hangDetectionConfig.json`: Configures the switch of the task hang detection function and the detection metric thresholds.

<a name="zh-cn_topic_0000002479387566_table_hangdetectionconfig"></a>

**Table 1** hangDetectionConfig.json field description

|Parameter|Description|Default Value|
|--|--|--|
|Enabled|Whether to enable hang fault detection. If this parameter is set to false, detection is disabled. If this parameter is set to true, detection is enabled.|true|
|AICoreUtilization|AICore utilization threshold, in %.|5|
|HbmMemoryDelta|On-chip memory usage increment threshold, in %.|0|
|TrafficDelta|Network traffic increment threshold, in packets/minute.|100|
|CPUTimeDelta|Process CPU time increment threshold, in seconds.|5|
|DetectDuration|Threshold of the number of consecutive times the hang state is detected. After this threshold is reached, the fault is reported. If this threshold is too small, misdetection may occur. Modify it with caution.|5|

>[!NOTE]
>
>- If `AICoreUtilization`, `HbmMemoryDelta`, `TrafficDelta`, and `CPUTimeDelta` in the current round are all less than the thresholds, the hang state count increases by 1.
>- If any threshold parameter in the `hangDetectionConfig.json` file is less than 0, the default value is used instead.
>- Task hang detection may conflict with process-level recovery: during process-level recovery, various running metrics are generally low and are easily misjudged as a hang state. When process-level recovery is used, it is recommended that you disable hang detection or adjust the thresholds properly.

## (Optional) Configuring Task Hang Detection Parameters<a name="ZH-CN_TOPIC_0000002479387566_custom"></a>

When the Ascend Device Plugin image is built, the task hang detection configuration file `hangDetectionConfig.json` is built into the image. When Ascend Device Plugin is started, it reads the configuration in this file as the basis for handling the current task hang.

If you want to customize task hang detection parameters, you can modify the corresponding `hangDetectionConfig.json` file when building the Ascend Device Plugin image.

**Procedure<a name="zh-cn_topic_0000002479387566_section_custom_hangDetectionConfig"></a>**

1. Log in to the environment and go to the directory where Ascend Device Plugin is decompressed.
2. Run the `vi hangDetectionConfig.json` command to edit the configuration file.

    ```shell
    vi hangDetectionConfig.json
    ```

    Modify the corresponding configuration parameters. For parameter descriptions, see [Table 1](#zh-cn_topic_0000002479387566_table_hangdetectionconfig). After the modification is complete, press `Esc`, type `:wq!`, and press `Enter` to save and exit.

    ```json
    {
        "HangDetection": {
            "Enabled": true,
            "Threshold": {
                "AICoreUtilization": 5,
                "HbmMemoryDelta": 0,
                "TrafficDelta": 100,
                "CPUTimeDelta": 5,
                "DetectDuration": 5
            }
        }
    }
    ```

## (Optional) Modifying the Task Hang Fault Handling Level<a name="zh-cn_topic_0000002479387566_section_custom_hangfaultlevel"></a>

For information about task hang fault `200001002`, see [Table 5](./08_public_faults.md#zh-cn_topic_0000002181110120_section4960201383813). If you need to isolate the NPU after a task hang or perform other operations, you can modify the fault level of this fault code by referring to the [Using faultCode.json to Configure Fault Levels](./03_chip_faults.md#zh-cn_topic_0000001951258609_section112139052513) section. The modified `mindx-dl-fault-config` is as follows:

```json
   ...
  "SeparateNPUCodes":[
    ...,"200001002"
  ],
    ...
```
