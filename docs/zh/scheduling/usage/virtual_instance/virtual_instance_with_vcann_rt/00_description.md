# 特性说明<a name="ZH-CN_TOPIC_0000002511426281vcann"></a>

基于vCANN-RT的虚拟化功能是指通过向vCANN-RT提供软切分配置文件的方式将物理机配置的NPU（昇腾AI处理器）挂载到容器中使用，虚拟化管理方式能够实现统一不同规格资源的分配和回收处理，满足多用户反复申请/释放资源的操作请求。

昇腾基于vCANN-RT的虚拟化实例功能的优点是可实现多个用户共同使用一台服务器，用户可以按需申请NPU的资源，降低了用户使用NPU算力的门槛和成本。多个用户共同使用一台服务器的NPU，并借助容器进行资源隔离，资源隔离性好，保证运行环境的平稳和安全，且资源分配与回收过程统一，从而方便多租户管理。

**原理介绍<a name="section154002962818vcann"></a>**

昇腾NPU硬件资源主要包括AICore（用于AI模型的计算）、AICPU、内存等，基于vCANN-RT的虚拟化实例功能主要原理是将上述硬件资源根据用户指定的资源需求，以软切分配置文件的方式通过vCANN-RT实现按需分配。例如用户只需要使用50% AICore的算力和2048MB的高带宽内存，系统就会创建一个npu_info配置文件，通过vCANN-RT向NPU芯片获取上述资源提供给容器使用，基于vCANN-RT的虚拟化实例方案如[图1 基于vCANN-RT的虚拟化实例方案](#fig987114711574vcann)所示。

**图 1**  基于vCANN-RT的虚拟化实例方案<a name="fig987114711574vcann"></a>  
![](../../../../figures/scheduling/virtual_instance_vcann.PNG "virtual_instance_vcann")

**产品支持说明<a name="section17326115542216vcann"></a>**

**表 1**  产品支持情况说明

<a name="table32786155236vcann"></a>

|产品系列|支持的场景|虚拟化方式|是否支持|
|--|--|--|--|
|<term>Atlas A2 推理系列产品</term><ul><li>Atlas 800I A2 推理服务器</li></ul>|在物理机生成软切分配置文件，挂载NPU和位置文件到容器|软切分虚拟化|是|
|<term>Atlas A3 推理系列产品</term><ul><li>Atlas 800I A3 超节点服务器</li></ul>|在物理机生成软切分配置文件，挂载NPU和位置文件到容器|软切分虚拟化|是|

**使用说明<a name="section1296713336303vcann"></a>**

- 软切分虚拟化基于[vCANN-RT](https://gitcode.com/openeuler/ubs-virt/blob/master/ubs-virt-enpu/vcann-rt/README.md)实现，直接将NPU重复挂载到多个容器，容器内的CANN按照配置好的比例使用NPU资源。
- 如果使用软切分虚拟化功能，需要先参见[软切分虚拟化](./01_soft_allocation_virtualization.md)，再进行挂载到容器操作。

**使用约束<a name="section911013420264vcann"></a>**

- 在软切分虚拟化场景下，一个容器只能挂载一个NPU。
- 任务YAML中requests对应的数据表示请求的NPU的AI Core百分比，不是真实NPU卡数。
- Atlas A3 推理系列产品使用软切分虚拟化功能时，必须开启单die直通模式，即在Ascend Device Plugin的YAML中，增加启动参数-useSingleDieMode=true。
- 物理NPU软切分虚拟化后，仅支持将物理NPU挂载到容器，不支持将该物理NPU直通到虚拟机。
- 在软切分虚拟化场景下，如果所有容器都挂载了相同的物理NPU，则该物理NPU必须采用相同的软切分策略。
- 对于<term>Atlas A2 推理系列产品</term>/<term>Atlas A3 推理系列产品</term>，一个Device上最多只能支持63个用户进程，Host最多只能支持Device个数\*63个进程，详情请参见[使用约束](https://www.hiascend.com/document/detail/zh/canncommercial/850/appdevg/acldevg/aclcppdevg_000222.html)。
