# 故障诊断接口<a name="ZH-CN_TOPIC_0000001541948914"></a>

**接口原型<a name="zh-cn_topic_0000001511538701_section124882040143613"></a>**

```
ascend-fd diag -i 诊断输入目录 -o 诊断结果输出目录 
```

**功能说明<a name="zh-cn_topic_0000001511538701_section12230185113815"></a>**

启动故障诊断任务。日志清洗结束后，依据日志清洗结果诊断故障事件。

**参数说明<a name="zh-cn_topic_0000001511538701_section122149111390"></a>**

**表 1**  参数说明

|**参数**|**缩写**|**是否必选**|**值类型**|**说明**|
|--|--|--|--|--|
|--input_path|-i|是|String|清洗完毕数据输入路径。|
|--output_path|-o|是|String|诊断结果输出路径。|
|--performance|-p|否|Bool|指定该参数时将执行所有诊断模块。不指定则只执行根因节点与故障事件两个模块的故障诊断功能。|
|--help|-h|否|-|查询二级命令与参数含义以及使用说明。|
|--scene|-s|否|String|诊断场景，默认host。可选host或super_pod。<ul><li>host：单独诊断主机Host日志场景。</li><li>super_pod：诊断超节点日志，包括Host、BMC、LCNE目录结构的日志场景。</li></ul>|


**返回说明<a name="zh-cn_topic_0000001511538701_section1714345618323"></a>**

故障诊断任务执行状态。

```
The diag job starts. Please wait. Job id: [****], run log file is [****].
诊断内容
The diag job is complete.
```



