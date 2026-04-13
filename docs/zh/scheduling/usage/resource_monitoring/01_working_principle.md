# 实现原理<a name="ZH-CN_TOPIC_0000002511346971"></a>

资源监测特性的实现原理如[图1](#fig167794421598)所示。

**图 1**  特性原理<a name="fig167794421598"></a>  
![](../../../figures/scheduling/特性原理.png "特性原理")

NPU Exporter组件通过gRPC服务调用K8s中的标准化接口CRI，获取容器相关信息；通过exec调用hccn\_tool工具，获取芯片的网络信息；通过dlopen/dlsym调用DCMI接口，获取芯片信息，并上报给Prometheus。

>[!NOTE]
>使用Telegraf的用户，直接调用NPU Exporter组件，获取相关信息。
