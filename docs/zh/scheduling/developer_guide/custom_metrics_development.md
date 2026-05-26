# 自定义指标开发<a name="ZH-CN_TOPIC_0000002512192053"></a>

支持通过如下两种方式开发自定义指标。

- 通过文件方式开发自定义指标

    用户根据[自定义指标文件](../api/npu_exporter/03_custom_metrics_file.md)，创建符合要求的自定义指标文件。启动NPU Exporter时，配置"-textMetricsFilePath"参数，指定该自定义指标文件的路径。详情请参见[NPU Exporter启动参数](installation_deployment/manual_installation/03_npu_exporter.md#参数说明)。NPU Exporter会在每个数据采集周期读取自定义指标文件，并将文件内容上报给Prometheus或Telegraf。

    开发示例如下：

    使用NPU Exporter集成并采集Devkit工具生成的hccs\_bandwidth指标，详情请参见[NPU Exporter集成Devkit部署指南](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/utils/npu-exporter)。关于hccs\_bandwidth指标信息的说明请参见[HCCS带宽监控](https://www.hikunpeng.com/document/detail/zh/kunpengdevps/profiler/profiler/KunpengDevKitCli_0251.html)。

- 通过插件方式开发自定义指标

    用户可通过编写插件的方式自定义指标，使用该插件前，开发者需要自行学习了解cgo、go相关语言特性，并阅读[README](https://gitcode.com/Ascend/mind-cluster/blob/master/component/npu-exporter/plugins/README.md)了解使用方法。

>[!NOTICE]
>
>- 自定义的指标不能与已有的指标名重复。
>- 开发者需对自定义插件的稳定性负责，确保不引入运行时panic等问题。
>- 开发者需要对自定义指标文件格式的正确性负责。
