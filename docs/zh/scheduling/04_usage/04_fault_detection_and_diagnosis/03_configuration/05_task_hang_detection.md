# 配置业务故障与任务卡死检测<a name="ZH-CN_TOPIC_0000002479387566"></a>

## 配置文件说明<a name="ZH-CN_TOPIC_0000002479387566_section_specifications "></a>

任务卡死故障由Ascend Device Plugin组件负责检测。此功能涉及以下两个配置文件：

- faultCode.json：配置任务卡死故障的故障级别。
- hangDetectionConfig.json：配置任务卡死检测功能开关和检测指标阈值。

<a name="zh-cn_topic_0000002479387566_table_hangdetectionconfig"></a>

**表 1**  hangDetectionConfig.json字段说明

|参数|说明|默认值|
|--|--|--|
|Enabled|是否开启卡死故障检测。设为false，表示不开启；设为true，表示开启。|true|
|AICoreUtilization|AICore利用率阈值，单位为%。|5|
|HbmMemoryDelta|HBM显存使用率增量阈值，单位为%。|0|
|TrafficDelta|网络通信流量增量阈值，单位为包/分钟。|100|
|CPUTimeDelta|进程CPU时间增量阈值，单位为秒。|5|
|DetectDuration|连续检测到卡死状态的次数阈值，达到此次数后上报故障。若此阈值太小，可能会导致误检测，请用户谨慎修改。|5|

>[!NOTE]
>
>- 本轮AICore利用率、HBM显存使用率增量、网络通信流量增量和进程CPU时间增量均小于阈值，则卡死状态次数加1。
>- hangDetectionConfig.json配置文件中任一阈值参数小于0时，将使用默认值替代。

## （可选）配置任务卡死检测参数<a name="ZH-CN_TOPIC_0000002479387566_custom"></a>

在制作Ascend Device Plugin镜像时，会将任务卡死检测配置文件hangDetectionConfig.json内置在镜像中，启动Ascend Device Plugin时会读取这个文件的配置，作为当前任务卡死处理依据。

如果用户想要自定义任务卡死检测参数，可以在制作Ascend Device Plugin镜像时，修改对应的hangDetectionConfig.json配置文件。

**操作步骤<a name="zh-cn_topic_0000002479387566_section_custom_hangDetectionConfig"></a>**

1. 登录环境，进入Ascend Device Plugin解压后的目录。
2. 执行**vi hangDetectionConfig.json**命令编辑配置文件。

    ```shell
    vi hangDetectionConfig.json
    ```

    修改对应配置参数，参数说明可参考[表1](#zh-cn_topic_0000002479387566_table_hangdetectionconfig)。修改完成后，按“Esc”键，输入:wq!保存并退出。

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

## （可选）修改任务卡死故障处理级别<a name="zh-cn_topic_0000002479387566_section_custom_hangfaultlevel"></a>

可参照[表5](./08_public_faults.md#zh-cn_topic_0000002181110120_section4960201383813)查看任务卡死故障200001002的相关信息。若用户需要对任务卡死后的NPU进行隔离或其他操作，可参照[使用faultCode.json配置故障级别](./03_chip_faults.md#zh-cn_topic_0000001951258609_section112139052513)小节修改此故障码的故障级别，修改后的mindx-dl-fault-config示例如下：

```json
   ...
  "SeparateNPUCodes":[
    ...,"200001002"
  ],
    ...
```
