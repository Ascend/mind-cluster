# API接口说明<a name="ZH-CN_TOPIC_0000001683212357"></a>

## 概述<a name="ZH-CN_TOPIC_0000001574561160"></a>

MindCluster Ascend FaultDiag组件提供的功能接口包含命令接口和SDK接口。用户可以通过调用接口实现相关功能。

命令接口：直接使用命令方式。提供了日志清洗、故障诊断、单机故障诊断、自定义配置文件、自定义故障实体、屏蔽故障日志、查询版本号和帮助命令接口。

SDK接口：代码级接口，直接调用函数和方法等方式。提供了业务流清洗、根因节点清洗、根因节点诊断、故障事件清洗及故障事件诊断接口。

**表 1**  命令接口

|**命令**|**功能说明**|
|--|--|
|ascend-fd parse|日志清洗命令。启动日志清洗任务，清洗训练/推理过程中采集的中间结果数据。|
|ascend-fd diag|故障诊断命令。启动故障分析任务，分析故障根因，并输出分析报告。|
|ascend-fd single-diag|单机故障诊断命令。启动单机故障分析任务，输出分析报告。|
|ascend-fd entity|自定义故障实体命令。用户可自定义故障实体，MindCluster Ascend FaultDiag支持对自定义的故障进行日志清洗、故障诊断、屏蔽故障日志等功能。|
|ascend-fd blacklist|屏蔽故障日志命令。故障关键词的日志信息，将不被记录到日志清洗后的文件中。|
|ascend-fd config|自定义配置文件命令。用户可以自定义配置是否支持清洗ModelArts关键日志、配置读取控制台日志大小、配置解析自定义的文件。|
|ascend-fd version|版本信息查看命令。查询组件版本信息。|
|ascend-fd -h|查询帮助信息。|


**表 2**  SDK接口

|命令|功能说明|
|--|--|
|parse_fault_type|业务流清洗接口。|
|parse_root_cluster|根因节点清洗接口。|
|diag_root_cluster|根因节点诊断接口。|
|parse_knowledge_graph|故障事件清洗接口。|
|diag_knowledge_graph|故障事件诊断接口。|


使用本组件时，会在$\{HOME\}/.ascend\_faultdiag目录下生成操作日志和运行日志，日志目录结构如下。

```
${HOME}/.ascend_faultdiag
└── ascend_faultdiag_operation.log    # 操作日志
└── RUN_LOG                           # 运行日志
  └─ 20241104142355468743_6797877f-7143-443f-a9c6-361e33032c5c
```

日志保存机制如下：日志文件大小不超过10M，超过限制大小后将自动转储另一个日志文件，同PID的日志文件数量不超过10个，超过限制个数时将自动覆盖最早创建的日志。

