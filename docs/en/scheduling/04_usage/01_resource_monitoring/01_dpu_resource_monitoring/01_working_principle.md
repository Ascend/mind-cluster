# Working Principle<a name="ZH-CN_TOPIC_0000002524312667"></a>

The implementation principle of the DPU resource monitoring feature (taking the Prometheus scenario as an example) is as follows:

1. DPU Exporter periodically calls the NIC management tool hinicadm5 to query the RoCE counters of the DPU, obtains DPU global metrics, and stores them in the cache.
2. DPU Exporter periodically reads the files in the "/sys/class/net/&lt;interface_name&gt;/" directory of the sysfs file system to obtain interface-level metrics, and stores them in the cache.
3. DPU Exporter implements the Prometheus metrics API (/metrics), so that Prometheus can periodically obtain the data in the cache.

Configuration items such as the collection period and the metric trustlist can be modified through the config.json file. The modification takes effect automatically without restarting the component. For details, see [DPU Exporter Installation and Deployment](../../../05_developer_guide/00_installation_deployment/00_manual_installation/12_dpu_exporter.md).

>[!NOTE]
>
> - By default, DPU Exporter provides only HTTP services. To use the more secure HTTPS service, modify the source code for adaptation.
