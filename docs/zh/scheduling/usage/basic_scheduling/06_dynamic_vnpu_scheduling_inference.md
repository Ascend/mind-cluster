# 动态vNPU调度（推理）<a name="ZH-CN_TOPIC_0000002511427045"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002511347087"></a>

**前提条件<a name="section121807404519"></a>**

在命令行场景下使用动态vNPU调度特性，需要确保已经安装如下组件；若没有安装，可以参考[安装部署](../../installation_guide/03_installation.md)章节进行操作。动态vNPU调度特性只支持使用Volcano作为调度器，不支持使用其他调度器。

- Volcano
- Ascend Device Plugin
- Ascend Docker Runtime
- ClusterD
- NodeD

**使用方式<a name="zh-cn_topic_0000001559979444_section91871616135119"></a>**

动态vNPU调度特性的使用方式如下：

- 通过命令行使用：安装集群调度组件，通过命令行使用动态vNPU调度特性。
- 集成后使用：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section10769161412815"></a>**

- 资源监测可以和推理场景下的所有特性一起使用。
- 集群中同时运行多个推理任务，每个任务使用的特性可以不同，但不能同时存在使用静态vNPU的任务和使用动态vNPU的任务。
- 动态vNPU调度特性需要搭配算力虚拟化特性一起使用，关于动态虚拟化的相关说明和操作请参见[动态虚拟化](../virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#动态虚拟化)章节。
- 动态vNPU调度仅支持下发单副本数或者多副本数的单机任务，每个副本独立工作，不支持分布式任务。

**支持的产品形态<a name="section169961844182917"></a>**

Atlas 推理系列产品

**使用流程<a name="zh-cn_topic_0000001559979444_section246711128536"></a>**

通过命令行使用动态vNPU调度特性流程可以参见[图1](#zh-cn_topic_0000001559979444_fig242524985412)。

**图 1**  使用流程<a name="zh-cn_topic_0000001559979444_fig242524985412"></a>  
![](../../../figures/scheduling/使用流程-3.png "使用流程-3")

算力动态虚拟化实例涉及到相关集群调度组件的参数配置，请参见[动态虚拟化](../virtual_instance/virtual_instance_with_hdk/06_mounting_vnpu.md#动态虚拟化)章节完成修改。

## 实现原理<a name="ZH-CN_TOPIC_0000002511427057"></a>

根据推理任务类型的不同，特性的原理图略有差异。

**vcjob任务<a name="section11346231114"></a>**

vcjob任务原理图如[图1](#fig1918122131712)所示。

**图 1**  vcjob任务调度原理图<a name="fig1918122131712"></a>  
![](../../../figures/scheduling/vcjob任务调度原理图-4.png "vcjob任务调度原理图-4")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - kubelet上报节点芯片数量到节点对象（node）中。
    - Ascend Device Plugin定期上报AI Core数量到Node中。
    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。

2. ClusterD读取device-info-cm和node-info-cm中信息后，将信息分别写入cluster-info-device-cm和cluster-info-node-cm中。
3. 用户通过kubectl或者其他深度学习平台下发vcjob任务。
4. volcano-controller为任务创建相应PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
5. 当集群资源满足任务要求时，volcano-controller创建任务Pod。
6. volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入动态虚拟化的模板信息。
7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin根据模板信息动态虚拟化NPU。Ascend Docker Runtime协助挂载相应资源。

**deploy任务<a name="section41019364253"></a>**

deploy任务原理图如[图2](#fig349112913199)所示。

**图 2**  deploy任务调度原理图<a name="fig349112913199"></a>  
![](../../../figures/scheduling/deploy任务调度原理图-5.png "deploy任务调度原理图-5")

各步骤说明如下：

1. 集群调度组件定期上报节点和芯片信息。
    - Ascend Device Plugin定期上报AI Core数量到Node中。
    - 当节点上存在故障时，NodeD定期上报节点健康状态、节点硬件故障信息、节点DPC共享存储故障信息到node-info-cm中。
2. ClusterD读取device-info-cm和node-info-cm中信息后，将信息分别写入cluster-info-device-cm和cluster-info-node-cm中。
3. 用户通过kubectl或者其他深度学习平台下发deploy任务。
4. kube-controller为任务创建相应Pod。
5. volcano-controller创建任务PodGroup。关于PodGroup的详细说明，可以参考[开源Volcano官方文档](https://volcano.sh/zh/docs/v1-9-0/podgroup/)。
6. volcano-scheduler根据节点和芯片拓扑信息为任务选择合适节点，并在Pod的annotation上写入动态虚拟化的模板信息。
7. kubelet创建容器时，调用Ascend Device Plugin挂载芯片，Ascend Device Plugin根据Pod的annotation模板信息动态虚拟化NPU。Ascend Docker Runtime协助挂载相应资源。

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002479227144"></a>

### 制作镜像<a name="ZH-CN_TOPIC_0000002511427049"></a>

**获取推理镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section971616541059"></a>**

可选择以下方式中的一种来获取推理镜像。

- 推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据系统架构（ARM或者x86\_64）下载**推理基础镜像（**如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)**）**。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE]  
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

- （可选）可基于推理基础镜像定制用户自己的推理镜像，制作过程请参见[使用Dockerfile构建推理镜像](../../common_operations.md#使用dockerfile构建推理镜像)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section1294572963118"></a>**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../../security_hardening.md#容器安全加固)章节进行操作。

### 脚本适配<a name="ZH-CN_TOPIC_0000002511347067"></a>

本章节以昇腾镜像仓库中推理镜像为例为用户介绍操作流程，该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，确保该环境可以正常访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1. 确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2. 在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE]  
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。

### 准备任务YAML<a name="ZH-CN_TOPIC_0000002479387122"></a>

>[!NOTE]  
>如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev”目录下的设备。其他目录（如“/usr”）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001558853680_zh-cn_topic_0000001609074213_section14665181617334"></a>**

1. 获取相应的YAML文件。

    **表 1**  YAML说明

    <a name="table0265132716351"></a>
    <table><thead align="left"><tr id="row132651727163516"><th class="cellrowborder" valign="top" width="15.36%" id="mcps1.2.5.1.1"><p id="p1447515933616"><a name="p1447515933616"></a><a name="p1447515933616"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="18.2%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="37.769999999999996%" id="mcps1.2.5.1.3"><p id="p626512711358"><a name="p626512711358"></a><a name="p626512711358"></a>YAML名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="28.67%" id="mcps1.2.5.1.4"><p id="p3265172773514"><a name="p3265172773514"></a><a name="p3265172773514"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row826513275355"><td class="cellrowborder" valign="top" width="15.36%" headers="mcps1.2.5.1.1 "><p id="p278965223717"><a name="p278965223717"></a><a name="p278965223717"></a>Deployment</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" width="18.2%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p8853185832112"><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><span id="ph165178910439"><a name="ph165178910439"></a><a name="ph165178910439"></a>Atlas 推理系列产品</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="37.769999999999996%" headers="mcps1.2.5.1.3 "><p id="p142651427103519"><a name="p142651427103519"></a><a name="p142651427103519"></a>infer-deploy-dynamic.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="28.67%" headers="mcps1.2.5.1.4 "><p id="p1826522718352"><a name="p1826522718352"></a><a name="p1826522718352"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-deploy-dynamic.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row9265727173515"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p191941452171418"><a name="p191941452171418"></a><a name="p191941452171418"></a>Volcano Job</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p15629131423715"><a name="p15629131423715"></a><a name="p15629131423715"></a>infer-vcjob-dynamic.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p1626592713355"><a name="p1626592713355"></a><a name="p1626592713355"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-dynamic.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    </tbody>
    </table>

2. 将YAML文件上传至管理节点任意目录，并根据实际情况修改文件内容。

    在Atlas 推理系列产品上，以infer-deploy-dynamic.yaml为例，申请1个AI Core的参数配置示例如下。

    ```Yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: resnetinfer1-1-deploy
      labels:
        app: infers
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: infers
      template:
        metadata:
          labels:
            app: infers
            fault-scheduling: "grace"           # 重调度所使用的label
             # 以下参数说明请参见表infer-deploy-dynamic.yaml参数说明
            ring-controller.atlas: ascend-310P 
            vnpu-dvpp: "null"         
            vnpu-level: "low"           
        spec:
          schedulerName: volcano              # 需要使用MindCluster的调度器Volcano
          nodeSelector:
            host-arch: huawei-arm
          containers:
            - image: ubuntu-infer:v1   # 示例镜像
    ...
    
              resources:
                requests:
                  huawei.com/npu-core: 1        # 使用静态虚拟化的vir01模板动态虚拟化NPU
                limits:
                  huawei.com/npu-core: 1        # 数值与requests保持一致
    ```

    **表 2**  infer-deploy-dynamic.yaml参数说明

    <a name="table116201128162111"></a>
    <table><thead align="left"><tr id="row362062812113"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="p11620628192119"><a name="p11620628192119"></a><a name="p11620628192119"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.2"><p id="p13620192817213"><a name="p13620192817213"></a><a name="p13620192817213"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.3"><p id="p862022892120"><a name="p862022892120"></a><a name="p862022892120"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row136201528182116"><td class="cellrowborder" rowspan="2" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p56210289215"><a name="p56210289215"></a><a name="p56210289215"></a>vnpu-level</p>
    <p id="p262172815213"><a name="p262172815213"></a><a name="p262172815213"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p562182842111"><a name="p562182842111"></a><a name="p562182842111"></a>low</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662112892120"><a name="p662112892120"></a><a name="p662112892120"></a>低配，默认值，选择最低配置的虚拟化实例模板。</p>
    </td>
    </tr>
    <tr id="row196219286214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p146219285218"><a name="p146219285218"></a><a name="p146219285218"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p19621528112118"><a name="p19621528112118"></a><a name="p19621528112118"></a>性能优先。</p>
    <p id="p6621152812214"><a name="p6621152812214"></a><a name="p6621152812214"></a>在集群资源充足的情况下，将选择尽量高配的虚拟化实例模板；在整个集群资源已使用过多的情况下，如大部分物理NPU都已使用，每个物理NPU只剩下小部分AI Core，不足以满足高配虚拟化实例模板时，将使用相同AI Core数量下较低配置的其他模板。具体选择请参考<a href="../virtual_instance/virtual_instance_with_hdk/03_virtualization_templates.md">虚拟化模板</a>章节。</p>
    </td>
    </tr>
    <tr id="row1762192862114"><td class="cellrowborder" rowspan="3" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p462112842110"><a name="p462112842110"></a><a name="p462112842110"></a>vnpu-dvpp</p>
    <p id="p362120286216"><a name="p362120286216"></a><a name="p362120286216"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621122816219"><a name="p8621122816219"></a><a name="p8621122816219"></a>yes</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p662162819213"><a name="p662162819213"></a><a name="p662162819213"></a>该<span id="ph1762113285210"><a name="ph1762113285210"></a><a name="ph1762113285210"></a>Pod</span>使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762172862117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p46214285213"><a name="p46214285213"></a><a name="p46214285213"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p5621162812213"><a name="p5621162812213"></a><a name="p5621162812213"></a>该<span id="ph1362102815215"><a name="ph1362102815215"></a><a name="ph1362102815215"></a>Pod</span>不使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1262122852117"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="p462192852111"><a name="p462192852111"></a><a name="p462192852111"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="p11621102818211"><a name="p11621102818211"></a><a name="p11621102818211"></a>默认值，不关注是否使用DVPP。</p>
    </td>
    </tr>
    <tr id="row1762110285219"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="p2062182822111"><a name="p2062182822111"></a><a name="p2062182822111"></a>ring-controller.atlas</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.2 "><p id="p8621102882111"><a name="p8621102882111"></a><a name="p8621102882111"></a>ascend-310P</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.3 "><p id="p1762182892113"><a name="p1762182892113"></a><a name="p1762182892113"></a>任务使用<span id="ph1623844892113"><a name="ph1623844892113"></a><a name="ph1623844892113"></a>Atlas 推理系列产品</span>的标识。</p>
    </td>
    </tr>
    </tbody>
    </table>

    vnpu-level和vnpu-dvpp作用后，选择的vNPU模板可参考[表3](#zh-cn_topic_0000001557486210_table83781115185619)。

    **表 3**  dvpp和level作用结果

    <a name="zh-cn_topic_0000001557486210_table83781115185619"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001557486210_row1837817157565"><th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001557486210_p1024717408463"><a name="zh-cn_topic_0000001557486210_p1024717408463"></a><a name="zh-cn_topic_0000001557486210_p1024717408463"></a>AI Core请求数量</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001557486210_p192479402463"><a name="zh-cn_topic_0000001557486210_p192479402463"></a><a name="zh-cn_topic_0000001557486210_p192479402463"></a>vnpu-dvpp</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001557486210_p1024716402460"><a name="zh-cn_topic_0000001557486210_p1024716402460"></a><a name="zh-cn_topic_0000001557486210_p1024716402460"></a>vnpu-level</p>
    </th>
    <th class="cellrowborder" valign="top" width="9.221844368873775%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001557486210_p8247440174613"><a name="zh-cn_topic_0000001557486210_p8247440174613"></a><a name="zh-cn_topic_0000001557486210_p8247440174613"></a>是否降级</p>
    </th>
    <th class="cellrowborder" valign="top" width="22.69453890778156%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001557486210_p0247164034611"><a name="zh-cn_topic_0000001557486210_p0247164034611"></a><a name="zh-cn_topic_0000001557486210_p0247164034611"></a>选择模板</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001557486210_row1937814158561"><td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p685714392538"><a name="zh-cn_topic_0000001557486210_p685714392538"></a><a name="zh-cn_topic_0000001557486210_p685714392538"></a>1</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p14248174010469"><a name="zh-cn_topic_0000001557486210_p14248174010469"></a><a name="zh-cn_topic_0000001557486210_p14248174010469"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p1385717396538"><a name="zh-cn_topic_0000001557486210_p1385717396538"></a><a name="zh-cn_topic_0000001557486210_p1385717396538"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p38575391531"><a name="zh-cn_topic_0000001557486210_p38575391531"></a><a name="zh-cn_topic_0000001557486210_p38575391531"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p385603935319"><a name="zh-cn_topic_0000001557486210_p385603935319"></a><a name="zh-cn_topic_0000001557486210_p385603935319"></a>vir01</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row14379191520562"><td class="cellrowborder" rowspan="3" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p146191732155316"><a name="zh-cn_topic_0000001557486210_p146191732155316"></a><a name="zh-cn_topic_0000001557486210_p146191732155316"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p1248174014614"><a name="zh-cn_topic_0000001557486210_p1248174014614"></a><a name="zh-cn_topic_0000001557486210_p1248174014614"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p14619832145315"><a name="zh-cn_topic_0000001557486210_p14619832145315"></a><a name="zh-cn_topic_0000001557486210_p14619832145315"></a>low/其他值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p126198326538"><a name="zh-cn_topic_0000001557486210_p126198326538"></a><a name="zh-cn_topic_0000001557486210_p126198326538"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p3248164094613"><a name="zh-cn_topic_0000001557486210_p3248164094613"></a><a name="zh-cn_topic_0000001557486210_p3248164094613"></a>vir02_1c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row0379131512562"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p122481840184611"><a name="zh-cn_topic_0000001557486210_p122481840184611"></a><a name="zh-cn_topic_0000001557486210_p122481840184611"></a>null</p>
    <p id="zh-cn_topic_0000001557486210_p13302164084616"><a name="zh-cn_topic_0000001557486210_p13302164084616"></a><a name="zh-cn_topic_0000001557486210_p13302164084616"></a></p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p162489402463"><a name="zh-cn_topic_0000001557486210_p162489402463"></a><a name="zh-cn_topic_0000001557486210_p162489402463"></a>high</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p22482040124615"><a name="zh-cn_topic_0000001557486210_p22482040124615"></a><a name="zh-cn_topic_0000001557486210_p22482040124615"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p182481740174611"><a name="zh-cn_topic_0000001557486210_p182481740174611"></a><a name="zh-cn_topic_0000001557486210_p182481740174611"></a>vir02</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row53795153568"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p1324834017468"><a name="zh-cn_topic_0000001557486210_p1324834017468"></a><a name="zh-cn_topic_0000001557486210_p1324834017468"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p16248840154619"><a name="zh-cn_topic_0000001557486210_p16248840154619"></a><a name="zh-cn_topic_0000001557486210_p16248840154619"></a>vir02_1c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1038061520567"><td class="cellrowborder" rowspan="7" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p0248240134616"><a name="zh-cn_topic_0000001557486210_p0248240134616"></a><a name="zh-cn_topic_0000001557486210_p0248240134616"></a>4</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p10248164012460"><a name="zh-cn_topic_0000001557486210_p10248164012460"></a><a name="zh-cn_topic_0000001557486210_p10248164012460"></a>yes</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p3248184024610"><a name="zh-cn_topic_0000001557486210_p3248184024610"></a><a name="zh-cn_topic_0000001557486210_p3248184024610"></a>low/其他值</p>
    </td>
    <td class="cellrowborder" rowspan="3" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p4249114074618"><a name="zh-cn_topic_0000001557486210_p4249114074618"></a><a name="zh-cn_topic_0000001557486210_p4249114074618"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p8249540164619"><a name="zh-cn_topic_0000001557486210_p8249540164619"></a><a name="zh-cn_topic_0000001557486210_p8249540164619"></a>vir04_4c_dvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row338051565613"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p192491540164619"><a name="zh-cn_topic_0000001557486210_p192491540164619"></a><a name="zh-cn_topic_0000001557486210_p192491540164619"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p5249124011467"><a name="zh-cn_topic_0000001557486210_p5249124011467"></a><a name="zh-cn_topic_0000001557486210_p5249124011467"></a>vir04_3c_ndvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row738071535612"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p424914004612"><a name="zh-cn_topic_0000001557486210_p424914004612"></a><a name="zh-cn_topic_0000001557486210_p424914004612"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p192493409466"><a name="zh-cn_topic_0000001557486210_p192493409466"></a><a name="zh-cn_topic_0000001557486210_p192493409466"></a>vir04_3c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1538131575615"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p924924018462"><a name="zh-cn_topic_0000001557486210_p924924018462"></a><a name="zh-cn_topic_0000001557486210_p924924018462"></a>yes</p>
    </td>
    <td class="cellrowborder" rowspan="4" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p2249440184619"><a name="zh-cn_topic_0000001557486210_p2249440184619"></a><a name="zh-cn_topic_0000001557486210_p2249440184619"></a>high</p>
    </td>
    <td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p14272035114811"><a name="zh-cn_topic_0000001557486210_p14272035114811"></a><a name="zh-cn_topic_0000001557486210_p14272035114811"></a>-</p>
    <p id="zh-cn_topic_0000001557486210_p14249184019467"><a name="zh-cn_topic_0000001557486210_p14249184019467"></a><a name="zh-cn_topic_0000001557486210_p14249184019467"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p1324984017461"><a name="zh-cn_topic_0000001557486210_p1324984017461"></a><a name="zh-cn_topic_0000001557486210_p1324984017461"></a>vir04_4c_dvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row1438113158568"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p824916403462"><a name="zh-cn_topic_0000001557486210_p824916403462"></a><a name="zh-cn_topic_0000001557486210_p824916403462"></a>no</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p15249440164616"><a name="zh-cn_topic_0000001557486210_p15249440164616"></a><a name="zh-cn_topic_0000001557486210_p15249440164616"></a>vir04_3c_ndvpp</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row238110156568"><td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p1824974014620"><a name="zh-cn_topic_0000001557486210_p1824974014620"></a><a name="zh-cn_topic_0000001557486210_p1824974014620"></a>null</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p10249124011467"><a name="zh-cn_topic_0000001557486210_p10249124011467"></a><a name="zh-cn_topic_0000001557486210_p10249124011467"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p324964074618"><a name="zh-cn_topic_0000001557486210_p324964074618"></a><a name="zh-cn_topic_0000001557486210_p324964074618"></a>vir04</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row17381415115611"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p2249340144615"><a name="zh-cn_topic_0000001557486210_p2249340144615"></a><a name="zh-cn_topic_0000001557486210_p2249340144615"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p924924064613"><a name="zh-cn_topic_0000001557486210_p924924064613"></a><a name="zh-cn_topic_0000001557486210_p924924064613"></a>vir04_3c</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001557486210_row8381181518563"><td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001557486210_p102497405465"><a name="zh-cn_topic_0000001557486210_p102497405465"></a><a name="zh-cn_topic_0000001557486210_p102497405465"></a>8或8的倍数</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001557486210_p42491440174615"><a name="zh-cn_topic_0000001557486210_p42491440174615"></a><a name="zh-cn_topic_0000001557486210_p42491440174615"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001557486210_p5249114074614"><a name="zh-cn_topic_0000001557486210_p5249114074614"></a><a name="zh-cn_topic_0000001557486210_p5249114074614"></a>任意值</p>
    </td>
    <td class="cellrowborder" valign="top" width="9.221844368873775%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001557486210_p1224920403467"><a name="zh-cn_topic_0000001557486210_p1224920403467"></a><a name="zh-cn_topic_0000001557486210_p1224920403467"></a>-</p>
    </td>
    <td class="cellrowborder" valign="top" width="22.69453890778156%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001557486210_p102491940124613"><a name="zh-cn_topic_0000001557486210_p102491940124613"></a><a name="zh-cn_topic_0000001557486210_p102491940124613"></a>-</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 
    >AI Core的申请数量为8或8的倍数，表示使用整张NPU卡。

3. 挂载权重文件。

    ```Yaml
    ...
                  ports:     # 分布式训练集合通信端口
                    - containerPort: 2222      
                      name: ascendjob-port      
                  resources:
                    limits:
                      huawei.com/Ascend310P: 1   # 申请的芯片数
                    requests:
                      huawei.com/Ascend310P: 1   #与limits取值一致
                  volumeMounts:
    ...
                      # 权重文件挂载路径
                    - name: weights                  
                      mountPath: /path-to-weights
    ...
              volumes:
    ...
                # 权重文件挂载路径
                - name: weights
                  hostPath:
                    path: /path-to-weights  # 共享存储或者本地存储路径，请根据实际情况修改
    ...
    ```

    >[!NOTE] 
    >- /path-to-weights为模型权重，需要用户自行准备。mindie镜像可以参考镜像中$ATB\_SPEED\_HOME\_PATH/examples/models/llama3/README.md文件中的说明进行下载。
    >- ATB_SPEED_HOME_PATH默认路径为“/usr/local/Ascend/atb-models”，在source模型仓中set_env.sh脚本时已配置，用户无需自行配置。

4. 修改所选YAML中的容器启动命令，即“command”字段内容，如果没有则需添加。

    ```Yaml
    ...
          containers:
          - image: ubuntu-infer:v1
    ...
            command: ["/bin/bash", "-c", "cd $ATB_SPEED_HOME_PATH; python examples/run_pa.py --model_path /path-to-weights"]
            resources:
              requests:
    ...
    ```

### 下发任务<a name="ZH-CN_TOPIC_0000002479227134"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```shell
kubectl apply -f XXX.yaml
```

例如：

```shell
kubectl apply -f infer-deploy-dynamic.yaml
```

回显示例如下：

```ColdFusion
job.batch/resnetinfer1-2 created
```

>[!NOTE]  
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002511347071"></a>

**操作步骤**

1. <a name="zh-cn_topic_0000001609093161_zh-cn_topic_0000001609474293_section96791230183711011"></a>执行以下命令，查看Pod运行状况。

    ```shell
    kubectl get pod --all-namespaces
    ```

    回显示例如下：

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default          resnetinfer1-2-scpr5                      1/1     Running   0          8s
    ...
    ```

2. 查看运行推理任务的节点详情。
    1. 执行以下命令查看节点的名称。

        ```shell
        kubectl get node -A
        ```

    2. 根据上一步骤中查询到的节点名称，执行以下命令查看节点详情。

        ```shell
        kubectl describe node <nodename>
        ```

        回显示例如下：

        ```ColdFusion
        ...
        Allocated resources:
          (Total limits may be over 100 percent, i.e., overcommitted.)
          Resource              Requests     Limits
          --------              --------     ------
          cpu                   4 (2%)       3500m (1%)
          memory                2140Mi (0%)  4040Mi (0%)
          ephemeral-storage     0 (0%)       0 (0%)
          huawei.com/npu-core  4            4
        Events:
          Type    Reason    Age   From                Message
          ----    ------    ----  ----                -------
          Normal  Starting  36m   kube-proxy, ubuntu  Starting kube-proxy.
        ...
        ```

        在显示的信息中，找到“Allocated resources”下的**huawei.com/npu-core**，该参数取值在执行推理任务之后会增加，增加数量为推理任务使用的NPU芯片个数。

### 查看动态vNPU调度结果<a name="ZH-CN_TOPIC_0000002479387120"></a>

**操作步骤<a name="zh-cn_topic_0000001559013282_zh-cn_topic_0000001558675486_section96791230183711"></a>**

在管理节点执行以下命令查看推理结果。

```shell
kubectl logs -f resnetinfer1-2-scpr5
```

回显示例如下，以实际回显为准。

```ColdFusion
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Answer[0]:  Deep learning is a subset of machine learning that uses neural networks with multiple layers to model complex relationships between
[2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Generate[0] token num: (0, 20)
```

>[!NOTE]  
>_resnetinfer1-2-scpr5_：查看任务进程章节[步骤1](#zh-cn_topic_0000001609093161_zh-cn_topic_0000001609474293_section96791230183711011)中运行的任务名称。

### 删除任务<a name="ZH-CN_TOPIC_0000002511347065"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```shell
kubectl delete -f XXX.yaml
```

例如：

```shell
kubectl delete -f infer-deploy-dynamic.yaml
```

回显示例如下：

```ColdFusion
root@ubuntu:/home/test/yaml# kubectl delete -f infer-310p-1usoc.yaml 
job "resnetinfer1-1" deleted
```

## 集成后使用<a name="ZH-CN_TOPIC_0000002511347073"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1. 根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2. 根据K8s的官方API库，对任务进行创建、查询、删除等操作。
3. 创建、查询或删除任务时，用户需要将[示例YAML](#准备任务yaml)的内容转换成K8s官方API中定义的对象，通过官方API发送给K8s的API Server或者将YAML内容转换成JSON格式直接发送给K8s的API Server。
