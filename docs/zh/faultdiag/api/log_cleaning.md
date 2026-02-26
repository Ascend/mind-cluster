# 日志清洗接口<a name="ZH-CN_TOPIC_0000001541788954"></a>

**接口原型<a name="zh-cn_topic_0000001461778658_section876116162918"></a>**

-   集成所有日志进行清洗。

    ```
    ascend-fd parse -i 采集目录 -o 清洗输出目录 
    ```

-   分类输入日志目录进行清洗。

    ```
    ascend-fd parse --host_log 主机侧操作系统日志采集目录 --device_log Device侧日志采集目录 --train_log 用户训练及推理日志采集目录 --process_log CANN应用类日志采集目录 --env_check NPU网口、状态信息、资源信息采集目录 --dl_log MindCluster组件日志采集目录 --mindie_log MindIE组件日志采集目录 --amct_log AMCT组件日志采集目录 --custom_log 自定义解析文件目录名 -o 清洗输出目录 
    ```

-   （可选）若有BMC侧日志，执行如下。

    ```
    ascend-fd parse --bmc_log BMC侧日志目录 -o 清洗结果保存目录
    ```

    如：

    ```
    ascend-fd parse --bmc_log  "bmc/worker-00" -o "auto_diag_combine/bmc/worker-00"
    ```

-   （可选）若有LCNE侧日志，执行如下。

    ```
    ascend-fd parse --lcne_log LCNE侧日志目录 -o 清洗结果保存目录
    ```

    如：

    ```
    ascend-fd parse --lcne_log  "lcne/worker-111" -o "auto_diag_combine/lcne/worker-111"
    ```

>[!NOTE] 说明 
>-   同时共用-i与详细日志采集目录参数时，会优先读取详细日志采集目录参数的输入值，再根据-i参数读取剩余日志采集目录。
>-   若-i参数与8个详细日志采集目录参数同时配置时，-i参数不生效。
>-   至少需要指定--input\_path、--host\_log、--device\_log、--train\_log、--process\_log、--env\_check、--dl\_log、--mindie\_log、--amct\_log、--custom\_log其中一个参数，否则清洗命令会执行失败。
>-   清洗命令指定的输出目录磁盘空间需大于5G，空间不足可能导致部分清洗结果丢失，进而导致诊断结果异常或不准确。

**功能说明<a name="zh-cn_topic_0000001461778658_section10145143713297"></a>**

启动日志清洗任务。训练及推理失败后，对运行日志、NPU环境检查文件等原始日志进行清洗工作。

**参数说明<a name="zh-cn_topic_0000001461778658_section1094205815292"></a>**

**表 1**  参数说明

|**参数**|**缩写**|**是否必选**|**值类型**|**说明**|
|--|--|--|--|--|
|--host_log|无|否|String|主机侧操作系统日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--device_log|无|否|String|Device侧日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--train_log|无|否|String|用户训练及推理日志采集目录。<ul><li>--train_log支持多个路径输入，路径可以是单个采集日志的文件名也可以是转储日志的采集目录。但最多只会读取20个路径，多余的部分将被废弃。</li><li>在使用--train_log指定文件名时，用户训练及推理日志将不再有命名约束限制；而在使用--train_log指定路径时，其路径下以.txt或.log结尾的文件将被视为训练及推理日志。</li></ul>|
|--process_log|无|否|String|CANN应用类日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--env_check|无|否|String|NPU网口、状态信息、资源信息采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--dl_log|无|否|String|MindCluster的Ascend Device Plugin、NodeD、Ascend Docker Runtime、NPU Exporter、Volcano组件日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--mindie_log|无|否|String|MindIE的组件MindIE Server、MindIE LLM、MindIE SD、MindIE RT、MindIE Torch、MindIE MS、MindIE Benchmark、MindIE Client产生的日志目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--amct_log|无|否|String|AMCT组件日志目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--bmc_log|无|否|String|BMC组件日志目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--lcne_log|无|否|String|LCNE组件日志目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--bus_log|无|否|String|A5 LCNE组件日志目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--custom_log|无|否|String|自定义解析文件目录。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--input_path|-i|否|String|预处理数据输入路径，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--output_path|-o|是|String|清洗完毕数据输出路径，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--performance|-p|否|Bool|指定该参数时将执行所有清洗模块。不指定则只执行根因节点与故障事件两个模块的日志清洗功能。|
|--help|-h|否|-|查询二级命令与参数含义以及使用说明。|


**返回说明<a name="zh-cn_topic_0000001461778658_section2134184616351"></a>**

日志清洗任务执行状态。

```
The parse job starts. Please wait. Job id: [****], run log file is [****].
These job ['模块1', '模块2'...] succeeded.
The parse job is complete.
```



