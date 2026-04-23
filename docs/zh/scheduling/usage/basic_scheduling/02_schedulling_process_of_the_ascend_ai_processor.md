# 昇腾AI处理器的调度流程<a name="ZH-CN_TOPIC_0000002511427051"></a>

整体调度逻辑如下所示，其中Ascend Device Plugin的功能是发现昇腾AI处理器资源并上报；Volcano组件为华为在开源Volcano框架上进行适配修改的调度器。

## 调度流程<a name="section7296392473"></a>

- **调度流程1**

    **图 1**  调度流程1<a name="fig63404531065"></a>  
    ![](../../../figures/scheduling/调度流程1.png "调度流程1")

    默认情况下，Volcano启动YAML中self-maintain-available-card参数的值为true。昇腾AI处理器的调度流程如下所示：

    1. Ascend Device Plugin组件上报昇腾AI处理器健康状态。
    2. 用户调用kube-apiserver创建使用NPU的业务容器，如vcjob。
    3. Volcano组件通过节点信息和ConfigMap信息计算当前可用的昇腾AI处理器。
    4. Volcano组件根据亲和性调度原则，将昇腾AI处理器分配的情况写入Pod的Annotations字段中，同时写入分配时的时间戳。Volcano组件写入资源信息后向Kubernetes提交绑定Pod申请。
    5. 在每个信息上报周期，Ascend Device Plugin从Pod的Annotations中读取芯片的挂载信息。如需修正，则通过kube-apiserver更新回Pod的Annotation中。修正的Annotation包括：huawei.com/资源名、huawei.com/AscendReal、ascend.kubectl.kubernetes.io/ascend-910-configuration。
    6. kubelet监测到有Pod调度到自己所在节点，调用Ascend Device Plugin的Allocate函数挂载NPU设备。同时也支持使用Ascend Docker Runtime挂载NPU设备。
    7. Ascend Device Plugin查询当前所在的Node中处于Pending状态的Pod列表，得到亲和性调度后时间戳最小的Pod，获取挂载的device ID，反馈给kubelet进行设备挂载。

- **调度流程2**

    **图 2**  调度流程2<a name="fig39301952134114"></a>  
    ![](../../../figures/scheduling/调度流程2.png "调度流程2")

    如果Volcano启动YAML中的self-maintain-available-card参数的值配置为false，昇腾AI处理器的调度流程如下所示：

    1. Ascend Device Plugin组件上报昇腾AI处理器健康状态。
    2. Ascend Device Plugin通过kube-apiserver将当前空闲的昇腾AI处理器（健康昇腾AI处理器  - 已使用的昇腾AI处理器）信息写到ConfigMap“mindx-dl-deviceinfo-\{_nodeName_\}”的“DeviceInfo”字段中。
    3. 用户调用kube-apiserver创建使用NPU的业务容器，如vcjob。
    4. Volcano组件通过“DeviceInfo”获取当前可用的昇腾AI处理器。
    5. Volcano组件根据亲和性调度原则，将昇腾AI处理器分配的情况写入Pod的“Annotations”字段中，同时写入分配时的时间戳。Volcano组件写入资源信息后向Kubernetes提交绑定Pod申请。
    6. kubelet监测到有Pod调度到自己所在节点，调用Ascend Device Plugin的Allocate函数挂载NPU设备。同时也支持使用Ascend Docker Runtime挂载NPU设备。
    7. Ascend Device Plugin查询当前所在的Node中处于Pending状态的Pod列表，得到亲和性调度后时间戳最小的Pod，获取挂载的device ID，反馈给kubelet进行设备挂载。
    8. Ascend Device Plugin更新“DeviceInfo”字段中的可分配昇腾AI处理器。

## 具体交互字段说明<a name="section154080418522"></a>

1. Ascend Device Plugin（开源代码版本）以ConfigMap形式上报节点资源，上报资源的形式为“huawei.com/资源名：资源名+物理ID”。格式如[图3](#fig83207421331)所示。图中标出部分表示可用昇腾AI处理器列表，是全部的健康昇腾AI处理器减去被Volcano分配的昇腾AI处理器。全部的健康昇腾AI处理器信息通过调用NPU驱动接口获取，而被Volcano分配的芯片是通过遍历当前Node上所有满足条件的Pod，即Pod的状态为非Failed或者Succeeded，且Pod的“Annotations”字段上有Volcano分配的昇腾AI处理器信息。

    >[!NOTE] 
    >- 用户可通过登录后台环境，执行**kubectl describe cm mindx-dl-deviceinfo-_\{__nodeName__\}_  -n kube-system**命令获取上报的资源信息。
    >- 该字段“huawei.com/资源名”正在日落，后续版本该字段不再呈现。默认情况下，节点的可用芯片由Volcano维护，该字段不生效。如果需要生效，可以修改Volcano的配置参数“self-maintain-available-card”值为false。

    **图 3**  节点NPU资源信息<a name="fig83207421331"></a>  
    ![](../../../figures/scheduling/节点NPU资源信息.png "节点NPU资源信息")

2. Volcano组件通过节点信息和ConfigMap信息计算当前可用的昇腾AI处理器。（如果Volcano配置开关self-maintain-available-card关闭，Volcano会以“huawei.com/资源名”为key，读取“DeviceInfo”字段信息作为可用昇腾AI处理器的依据。）根据亲和性调度策略，判断出任务需要的符合亲和性规则的昇腾AI处理器后（即分配给任务的昇腾AI处理器）。Volcano会将分配芯片信息写入任务Pod的“Annotations”，如[图4](#fig29119551778)标出的第一个部分所示；第二个需要写入的字段为“predicate-time”，表示为任务分配资源的当前时间，不需要向可读时间格式做转换，可比较大小即可。kubelet监测到有Pod调度到自己所在节点，调用Device-plugin的Allocate函数挂载NPU设备。

    **图 4**  分配给Pod的NPU信息<a name="fig29119551778"></a>  
    ![](../../../figures/scheduling/分配给Pod的NPU信息.png "分配给Pod的NPU信息")

3. Ascend Device Plugin在收到Allocate请求时（以2卡任务为例），因为Allocate输入的参数是kubelet随机分配的，如[图4](#fig29119551778)中的“huawei.com/kltDev”字段所示，可能是不符合亲和性规则的昇腾AI处理器ID，例如Ascend910-7和Ascend910-0。

    此时Ascend Device Plugin会找到当前Node上所有的满足条件的Pod（Pod的状态为非Failed或者Succeeded），且Pod的“Annotations”字段中存在Volcano写入的分配的昇腾AI处理器ID，昇腾AI处理器数量和kubelet分配昇腾AI处理器数量要一致。

    再从满足条件的Pod中，选择“predicate-time”最小的Pod，并把这个Pod的“predicate-time”改为最大的Uint值（避免下次再选到）。解析Pod的“Annotations”字段，得到Volcano分配的昇腾AI处理器信息，例如Ascend910-0和Ascend910-1，把它们对应的挂载路径等信息返回，并且将真正分配的昇腾AI处理器信息写入到Pod的“Annotations”中的“huawei.com/AscendReal”字段中。
