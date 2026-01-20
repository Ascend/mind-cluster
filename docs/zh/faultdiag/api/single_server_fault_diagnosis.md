# 单机故障诊断接口<a name="ZH-CN_TOPIC_0000002107732029"></a>

**接口原型<a name="section416811814162"></a>**

-   单机进行所有日志清洗，处理日志清洗结果诊断故障事件，输出分析报告。

    ```
    ascend-fd single-diag -i 采集目录 -o 单机诊断结果输出目录 
    ```

-   分类输入日志目录进行单机诊断。

    ```
    ascend-fd single-diag --host_log 主机侧操作系统日志采集目录 --device_log Device侧日志采集目录 --train_log 用户训练及推理日志采集目录 --process_log CANN应用类日志采集目录 --env_check NPU网口、状态信息、资源信息采集目录 --dl_log MindCluster组件日志采集目录 --mindie_log MindIE组件日志采集目录 --amct_log AMCT组件日志采集目录 -o 清洗输出目录 
    ```

>[!NOTE] 说明 
>-   同时共用-i与详细日志采集目录参数时，会优先读取详细日志采集目录参数的输入值，再根据-i参数读取剩余日志采集目录。
>-   若-i参数与8个详细日志采集目录参数同时配置时，-i参数不生效。
>-   至少需要指定--input\_path、--host\_log、--device\_log、--train\_log、--process\_log、--env\_check、--dl\_log、--mindie\_log、--amct\_log其中一个参数，否则清洗命令会执行失败。
>-   清洗命令指定的输出目录磁盘空间需大于5G，空间不足可能导致部分清洗结果丢失，进而导致诊断结果异常或不准确。

**功能说明<a name="section67721623124010"></a>**

启动单机诊断任务。训练及推理失败后，对单机运行日志、NPU环境检查文件等原始日志进行诊断工作。

**参数说明<a name="section7746133874017"></a>**

**表 1**  参数说明

|**参数**|**缩写**|**是否必选**|**值类型**|**说明**|
|--|--|--|--|--|
|--host_log|无|否|String|主机侧操作系统日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--device_log|无|否|String|Device侧日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--train_log|无|否|String|用户训练及推理日志采集目录。<ul><li>--train_log支持多个路径输入，路径可以是单个采集日志的文件名也可以是转储日志的采集目录。但最多只会读取20个路径，多余的部分将被废弃。</li><li>在使用--train_log指定文件名时，用户训练及推理日志将不再有命名约束限制；而在使用--train_log指定路径时，其路径下以.txt或.log结尾的文件将被视为训练及推理日志。</li></ul>|
|--process_log|无|否|String|CANN应用类日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--env_check|无|否|String|NPU网口、状态信息、资源信息采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--dl_log|无|否|String|MindCluster的Ascend Device Plugin、NodeD、Ascend Docker Runtime、NPU Exporter、Volcano组件日志采集目录，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--mindie_log|无|否|String|MindIE的组件MindIE Server、MindIE LLM、MindIE SD、MindIE RT、MindIE Torch、MindIE MS、MindIE Benchmark、MindIE Client产生的日志。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--amct_log|无|否|String|AMCT组件日志。仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--input_path|-i|否|String|预处理数据输入路径，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--output_path|-o|是|String|清洗完毕数据输出路径，仅支持数字、大小写字母和字符“~”，“-”，“+”，“_”，“.”，“/”，“ ”。|
|--help|-h|否|-|查询二级命令与参数含义以及使用说明。|


**返回说明<a name="section115671821144111"></a>**

单机诊断任务执行状态。

```
The single-diag job starts. Please wait. Job id: [****], run log file is [****].
诊断内容
The single-diag job is complete.
```
