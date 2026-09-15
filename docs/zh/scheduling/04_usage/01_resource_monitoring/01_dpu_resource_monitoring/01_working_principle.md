# 实现原理<a name="ZH-CN_TOPIC_0000002524312667"></a>

DPU资源监测特性的实现原理（以接入Prometheus场景为例）如下所示。

1. DPU Exporter组件周期性调用网卡管理工具hinicadm5，查询DPU卡的RoCE计数器，获取DPU全局指标，放入缓存。
2. DPU Exporter组件周期性读取sysfs文件系统"/sys/class/net/&lt;interface_name&gt;/"目录下的文件，获取interface级指标，放入缓存。
3. DPU Exporter组件实现Prometheus的指标接口（/metrics），供Prometheus周期性获取缓存中的数据信息。

采集周期、指标白名单等配置项可通过config.json配置文件修改，配置修改后自动生效，无需重启组件，详见[DPU Exporter安装部署](../../../05_developer_guide/00_installation_deployment/00_manual_installation/13_dpu_exporter.md)章节。

>[!NOTE]
>
> - DPU Exporter默认仅提供HTTP服务，如需使用更为安全的HTTPS服务，请自行修改源码进行适配。
