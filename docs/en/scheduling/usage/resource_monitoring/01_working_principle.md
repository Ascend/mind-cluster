# Implementation Principles<a name="ZH-CN_TOPIC_0000002511346971"></a>

The implementation principles of the resource monitoring feature are shown in [Figure 1](#fig167794421598).

**Figure 1**  Feature principles<a name="fig167794421598"></a>
![](../../../figures/scheduling/feature-principles.png)

NPU Exporter calls the standardized CRI interface in K8s through the gRPC service to obtain container-related information; calls the hccn_tool through exec to obtain the network information of the chip; calls the DCMI through dlopen/dlsym to obtain chip information, and reports it to Prometheus.

>[!NOTE]
>Users who use Telegraf can directly call NPU Exporter to obtain relevant information.
