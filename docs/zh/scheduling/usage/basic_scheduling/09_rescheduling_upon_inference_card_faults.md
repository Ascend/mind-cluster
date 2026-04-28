# 推理卡故障重调度<a name="ZH-CN_TOPIC_0000002479387124"></a>

## 使用前必读<a name="ZH-CN_TOPIC_0000002479387116"></a>

集群调度组件管理的推理芯片资源出现故障后，集群调度组件可以对故障资源（对应芯片）进行隔离并自动进行重调度。

**前提条件<a name="section166381652174516"></a>**

- 使用推理卡故障重调度特性，需要确保已经安装如下组件。
    - Volcano（本特性只支持使用Volcano作为调度器，不支持使用其他调度器。）
    - Ascend Device Plugin
    - Ascend Docker Runtime
    - Ascend Operator
    - ClusterD
    - NodeD

- 若没有安装，可以参考[安装部署](../../installation_guide/03_installation/manual_installation/00_obtaining_software_packages.md)章节进行操作。

**使用方式<a name="zh-cn_topic_0000001559979444_section91871616135119"></a>**

推理卡故障重调度的使用方式如下：

- [通过命令行使用](#ZH-CN_TOPIC_0000002511427039)：安装集群调度组件，通过命令行使用推理卡故障重调度特性。
- [集成后使用](#ZH-CN_TOPIC_0000002479387118)：将集群调度组件集成到已有的第三方AI平台或者基于集群调度组件开发的AI平台。

**使用说明<a name="section10769161412815"></a>**

- 资源监测可以和推理场景下的所有特性一起使用。
- 集群中同时跑多个推理任务，每个任务使用的特性可以不同，但不能同时存在使用静态vNPU的任务和使用动态vNPU的任务。
- 推理卡故障重调度特性默认使用整卡调度；不支持静态vNPU调度；支持Atlas 推理系列产品使用动态vNPU调度。
- 推理卡故障重调度支持下发单副本数或者多副本数的单机任务，每个副本独立工作；只支持推理服务器（插Atlas 300I Duo 推理卡）和Atlas 800I A2 推理服务器、A200I A2 Box 异构组件部署acjob类型的分布式任务。

- 推理卡故障重调度支持vcjob或Deployment类型任务，且需在该类任务中增加故障重调度的开关的标签“fault-scheduling”，并将其设置为“grace”或者“force”。

**支持的产品形态<a name="section169961844182917"></a>**

支持以下产品使用推理卡故障重调度。

- 推理服务器（插Atlas 300I 推理卡）
- Atlas 推理系列产品
- Atlas 800I A2 推理服务器
- A200I A2 Box 异构组件
- Atlas 800I A3 超节点服务器
- Atlas 350 标卡

**使用流程<a name="zh-cn_topic_0000001559979444_section246711128536"></a>**

通过命令行使用推理卡故障重调度特性流程可以参见[图1](#zh-cn_topic_0000001559979444_fig242524985412)。

**图 1**  使用流程<a name="zh-cn_topic_0000001559979444_fig242524985412"></a>  
![](../../../figures/scheduling/使用流程-7.png "使用流程-7")

## 通过命令行使用（Volcano）<a name="ZH-CN_TOPIC_0000002511427039"></a>

### 制作镜像<a name="ZH-CN_TOPIC_0000002511427053"></a>

**获取推理镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section971616541059"></a>**

可选择以下方式中的一种来获取推理镜像。

- 推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据用户的系统架构（ARM或者x86\_64）下载推理基础镜像（如：[ascend-infer](https://www.hiascend.com/developer/ascendhub/detail/e02f286eef0847c2be426f370e0c2596)、[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)）。

    请注意，21.0.4版本之后推理基础镜像默认用户为非root用户，需要在下载基础镜像后对其进行修改，将默认用户修改为root。

    >[!NOTE] 
    >基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。

- （可选）如果用户需要更个性化的推理环境，可基于已下载的推理基础镜像，再[使用Dockerfile对其进行修改](../../common_operations.md#使用dockerfile构建容器镜像tensorflow)。

    完成定制化修改后，用户可以给推理镜像重命名，以便管理和使用。

**加固镜像<a name="zh-cn_topic_0000001609173557_zh-cn_topic_0000001558675566_section1294572963118"></a>**

下载或者制作的推理基础镜像可以进行安全加固，提升镜像安全性，可参见[容器安全加固](../../security_hardening.md#容器安全加固)章节进行操作。

### 脚本适配<a name="ZH-CN_TOPIC_0000002479227172"></a>

本章节以昇腾镜像仓库中推理镜像为例为用户介绍操作流程，该镜像已经包含了推理示例脚本，实际推理场景需要用户自行准备推理脚本。在拉取镜像前，需要确保当前环境的网络代理已经配置完成，且能成功访问昇腾镜像仓库。

**从昇腾镜像仓库获取示例脚本<a name="section8181015175911"></a>**

1. 确保服务器能访问互联网后，访问[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)。
2. 在左侧导航栏选择推理镜像，然后选择[mindie](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)镜像，获取推理示例脚本。

    >[!NOTE]  
    >若无下载权限，请根据页面提示申请权限。提交申请后等待管理员审核，审核通过后即可下载镜像。

### 准备任务YAML<a name="ZH-CN_TOPIC_0000002511427029"></a>

>[!NOTE]  
>
>- 如果用户不使用Ascend Docker Runtime组件，Ascend Device Plugin只会帮助用户挂载“/dev”目录下的设备。其他目录（如“/usr”）用户需要自行修改YAML文件，挂载对应的驱动目录和文件。容器内挂载路径和宿主机路径保持一致。
>- 因为Atlas 200I SoC A1 核心板场景不支持Ascend Docker Runtime，用户也无需修改YAML文件。

**操作步骤<a name="zh-cn_topic_0000001558853680_zh-cn_topic_0000001609074213_section14665181617334"></a>**

1. 下载YAML文件。

    **表 1**  任务类型与硬件型号对应YAML文件

    <a name="zh-cn_topic_0000001609074213_table15169151021912"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001609074213_row16169201019192"><th class="cellrowborder" valign="top" width="19.97%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001609074213_p4169191017192"><a name="zh-cn_topic_0000001609074213_p4169191017192"></a><a name="zh-cn_topic_0000001609074213_p4169191017192"></a>任务类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="20.03%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001609074213_p20181111517147"><a name="zh-cn_topic_0000001609074213_p20181111517147"></a><a name="zh-cn_topic_0000001609074213_p20181111517147"></a>硬件型号</p>
    </th>
    <th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001609074213_p181811156149"><a name="zh-cn_topic_0000001609074213_p181811156149"></a><a name="zh-cn_topic_0000001609074213_p181811156149"></a>YAML文件路径</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.5.1.4"><p id="p1693015221828"><a name="p1693015221828"></a><a name="p1693015221828"></a>获取链接</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001609074213_row2169191091919"><td class="cellrowborder" rowspan="3" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p6169510191913"><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><a name="zh-cn_topic_0000001609074213_p6169510191913"></a><span id="zh-cn_topic_0000001609074213_ph183921109162"><a name="zh-cn_topic_0000001609074213_ph183921109162"></a><a name="zh-cn_topic_0000001609074213_ph183921109162"></a>Volcano</span>调度的Deployment任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p8853185832112"><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><a name="zh-cn_topic_0000001609074213_p8853185832112"></a><span id="zh-cn_topic_0000001609074213_ph238151934915"><a name="zh-cn_topic_0000001609074213_ph238151934915"></a><a name="zh-cn_topic_0000001609074213_ph238151934915"></a>Atlas 200I SoC A1 核心板</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001609074213_p1116971091915"><a name="zh-cn_topic_0000001609074213_p1116971091915"></a><a name="zh-cn_topic_0000001609074213_p1116971091915"></a>infer-deploy-310p-1usoc.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p784716567219"><a name="p784716567219"></a><a name="p784716567219"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/tree/branch_v26.0.0/samples/inference/volcano/infer-deploy-310p-1usoc.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>Atlas 950 SuperPoD</p><p>Atlas 850 系列硬件产品（超节点）</p><p>Atlas 350 标卡</p></td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>infer-deploy-950.yaml</p></td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-deploy-950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001609074213_row17169201091917"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001609074213_p14853125832110"><a name="zh-cn_topic_0000001609074213_p14853125832110"></a><a name="zh-cn_topic_0000001609074213_p14853125832110"></a>其他类型推理节点</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001609074213_p51692100191"><a name="zh-cn_topic_0000001609074213_p51692100191"></a><a name="zh-cn_topic_0000001609074213_p51692100191"></a>infer-deploy.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-deploy.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row1137784216212"><td class="cellrowborder" rowspan="2" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="p9442102131620"><a name="p9442102131620"></a><a name="p9442102131620"></a>Volcano Job任务</p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="p367438101714"><a name="p367438101714"></a><a name="p367438101714"></a><span id="ph56332010913"><a name="ph56332010913"></a><a name="ph56332010913"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p168721535300"><a name="p168721535300"></a><a name="p168721535300"></a><span id="ph56342369338"><a name="ph56342369338"></a><a name="ph56342369338"></a>A200I A2 Box 异构组件</span></p>
    <p id="p17604333153213"><a name="p17604333153213"></a><a name="p17604333153213"></a><span id="ph12174764117"><a name="ph12174764117"></a><a name="ph12174764117"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="p8442112171619"><a name="p8442112171619"></a><a name="p8442112171619"></a>infer-vcjob-910.yaml</p>
    </td>
    <td class="cellrowborder" rowspan="1" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p15442424164"><a name="p15442424164"></a><a name="p15442424164"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-910.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>Atlas 950 SuperPoD</p><p>Atlas 850 系列硬件产品（超节点）</p><p>Atlas 350 标卡</p></td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>infer-vcjob-950.yaml</p></td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row3552077269"><td class="cellrowborder" rowspan="3" valign="top" width="19.97%" headers="mcps1.2.5.1.1 "><p id="p6861171325411"><a name="p6861171325411"></a><a name="p6861171325411"></a>Ascend Job任务</p>
    <p id="p12446175211817"><a name="p12446175211817"></a><a name="p12446175211817"></a></p>
    <p id="p5735201117263"><a name="p5735201117263"></a><a name="p5735201117263"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="20.03%" headers="mcps1.2.5.1.2 "><p id="p1328416110919"><a name="p1328416110919"></a><a name="p1328416110919"></a>推理服务器（插<span id="ph93658382564"><a name="ph93658382564"></a><a name="ph93658382564"></a>Atlas 300I Duo 推理卡</span>）</p>
    </td>
    <td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.3 "><p id="p10861813135419"><a name="p10861813135419"></a><a name="p10861813135419"></a>pytorch_acjob_infer_310p_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.5.1.4 "><p id="p1986116136544"><a name="p1986116136544"></a><a name="p1986116136544"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/pytorch_acjob_infer_310p_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr id="row512231072611"><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p id="p1611216221297"><a name="p1611216221297"></a><a name="p1611216221297"></a><span id="ph10342125017508"><a name="ph10342125017508"></a><a name="ph10342125017508"></a>Atlas 800I A2 推理服务器</span></p>
    <p id="p981315183317"><a name="p981315183317"></a><a name="p981315183317"></a><span id="ph176921116163312"><a name="ph176921116163312"></a><a name="ph176921116163312"></a>A200I A2 Box 异构组件</span></p>
    <p id="p4470103717329"><a name="p4470103717329"></a><a name="p4470103717329"></a><span id="ph1695943783214"><a name="ph1695943783214"></a><a name="ph1695943783214"></a>Atlas 800I A3 超节点服务器</span></p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p id="p4446185212815"><a name="p4446185212815"></a><a name="p4446185212815"></a>pytorch_multinodes_acjob_infer_{xxx}b_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p id="p962512301913"><a name="p962512301913"></a><a name="p962512301913"></a><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/pytorch_multinodes_acjob_infer_910b_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" headers="mcps1.2.5.1.1 "><p>Atlas 950 SuperPoD</p><p>Atlas 850 系列硬件产品（超节点）</p><p>Atlas 350 标卡</p></td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.2 "><p>pytorch_multinodes_acjob_infer_950_with_ranktable.yaml</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.5.1.3 "><p><a href="https://gitcode.com/Ascend/mindxdl-deploy/blob/branch_v26.0.0/samples/inference/volcano/pytorch_multinodes_acjob_infer_950_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 
    >Volcano支持Job类型任务，但是Job类型任务的YAML需要用户自行根据示例YAML修改适配。

2. 在[整卡调度](./04_full_npu_scheduling_and_static_vnpu_scheduling_inference.md#准备任务yaml)或者[动态vNPU调度](./06_dynamic_vnpu_scheduling_inference.md#准备任务yaml)的YAML配置基础上，增加如下字段启用重调度功能，以整卡调度的infer-deploy.yaml为例。

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
    ...
             fault-scheduling: grace               # 添加该字段
             ring-controller.atlas: ascend-310   # 添加该字段
        spec:
          schedulerName: volcano
          nodeSelector:
            host-arch: huawei-arm           # Select the os arch. If the os arch is x86, change it to huawei-x86.
    ...
    ```

    **表 2**  fault-scheduling配置项值列表

    <a name="table0396162644916"></a>
    <table><thead align="left"><tr id="row7397112634917"><th class="cellrowborder" valign="top" width="16.48%" id="mcps1.2.4.1.1"><p id="p1339762674911"><a name="p1339762674911"></a><a name="p1339762674911"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="29.42%" id="mcps1.2.4.1.2"><p id="p1139718264499"><a name="p1139718264499"></a><a name="p1139718264499"></a>取值</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.1%" id="mcps1.2.4.1.3"><p id="p123971426144911"><a name="p123971426144911"></a><a name="p123971426144911"></a>含义</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row9397182614491"><td class="cellrowborder" rowspan="2" valign="top" width="16.48%" headers="mcps1.2.4.1.1 "><p id="p113974261490"><a name="p113974261490"></a><a name="p113974261490"></a>fault-scheduling</p>
    <p id="p878610718561"><a name="p878610718561"></a><a name="p878610718561"></a></p>
    </td>
    <td class="cellrowborder" valign="top" width="29.42%" headers="mcps1.2.4.1.2 "><p id="p18397192617495"><a name="p18397192617495"></a><a name="p18397192617495"></a>grace</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.1%" headers="mcps1.2.4.1.3 "><p id="p4397112614920"><a name="p4397112614920"></a><a name="p4397112614920"></a>任务使用重调度开关，并在过程中先优雅删除原Pod。</p>
    </td>
    </tr>
    <tr id="row1378627165613"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001570873348_p2032819617590"><a name="zh-cn_topic_0000001570873348_p2032819617590"></a><a name="zh-cn_topic_0000001570873348_p2032819617590"></a>force</p>
    </td>
    <td class="cellrowborder" valign="top" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001570873348_p113286645910"><a name="zh-cn_topic_0000001570873348_p113286645910"></a><a name="zh-cn_topic_0000001570873348_p113286645910"></a>配置任务采用强制删除模式，在过程中强制删除原<span id="zh-cn_topic_0000001570873348_ph38454178285"><a name="zh-cn_topic_0000001570873348_ph38454178285"></a><a name="zh-cn_topic_0000001570873348_ph38454178285"></a>Pod</span>。</p>
    <p id="zh-cn_topic_0000001570873348_p3206181674916"><a name="zh-cn_topic_0000001570873348_p3206181674916"></a><a name="zh-cn_topic_0000001570873348_p3206181674916"></a></p>
    </td>
    </tr>
    <tr id="row11397142634918"><td class="cellrowborder" valign="top" width="16.48%" headers="mcps1.2.4.1.1 "><p id="p7397026134913"><a name="p7397026134913"></a><a name="p7397026134913"></a>ring-controller.atlas</p>
    </td>
    <td class="cellrowborder" valign="top" width="29.42%" headers="mcps1.2.4.1.2 "><a name="ul16397426184918"></a><a name="ul16397426184918"></a><ul id="ul16397426184918"><li>推理服务器（插<span id="ph3690191194813"><a name="ph3690191194813"></a><a name="ph3690191194813"></a>Atlas 300I 推理卡</span>）：ascend-310</li><li><span id="ph56912120486"><a name="ph56912120486"></a><a name="ph56912120486"></a>Atlas 推理系列产品</span>：ascend-310P</li><li><span id="ph16267162611508"><a name="ph16267162611508"></a><a name="ph16267162611508"></a>Atlas 800I A2 推理服务器</span>、<span id="ph344045773615"><a name="ph344045773615"></a><a name="ph344045773615"></a>A200I A2 Box 异构组件</span>、<span id="ph1175141233710"><a name="ph1175141233710"></a><a name="ph1175141233710"></a>Atlas 800I A3 超节点服务器</span>：ascend-<span id="ph4487202241512"><a name="ph4487202241512"></a><a name="ph4487202241512"></a><em id="zh-cn_topic_0000001519959665_i1489729141619"><a name="zh-cn_topic_0000001519959665_i1489729141619"></a><a name="zh-cn_topic_0000001519959665_i1489729141619"></a>{xxx}</em></span>b</li></ul>
    </td>
    <td class="cellrowborder" valign="top" width="54.1%" headers="mcps1.2.4.1.3 "><p id="p1397826104915"><a name="p1397826104915"></a><a name="p1397826104915"></a>用于校验任务使用的芯片类型。</p>
    </td>
    </tr>
    </tbody>
    </table>

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

### 下发任务<a name="ZH-CN_TOPIC_0000002511427027"></a>

在管理节点示例YAML所在路径，执行以下命令，使用YAML下发推理任务。

```shell
kubectl apply -f XXX.yaml
```

例如：

```shell
kubectl apply -f infer-310p-1usoc.yaml
```

回显示例如下：

```ColdFusion
job.batch/resnetinfer1-2 created
```

>[!NOTE]  
>如果下发任务成功后，又修改了任务YAML，需要先执行kubectl delete -f _XXX_.yaml命令删除原任务，再重新下发任务。

### 查看任务进程<a name="ZH-CN_TOPIC_0000002511427025"></a>

**操作步骤<a name="zh-cn_topic_0000001609093161_zh-cn_topic_0000001609474293_section96791230183711"></a>**

执行以下命令，查看Pod运行状况。

```shell
kubectl get pod --all-namespaces
```

回显示例如下：

```ColdFusion
NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
...
default          resnetinfer1-2-scpr5                      1/1     Running   0          20m
...
```

### 查看推理卡故障重调度结果<a name="ZH-CN_TOPIC_0000002511347069"></a>

当推理任务运行中出现故障时，Volcano会将该任务调度到其他NPU上。

**操作步骤<a name="section18664151111415"></a>**

1. 执行以下命令，查看任务运行状况。

    ```shell
    kubectl get pod --all-namespaces
    ```

    回显示例如下，任务名称由**resnetinfer1-2-scpr5**变为**resnetinfer1-2-xsdsf**，表示故障重调度特性运行成功。该任务名称由随机字符串生成，以实际名称为准。

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE
    ...
    default      resnetinfer1-2-xsdsf                    1/1    Running   0       10s
    ...
    ```

2. 执行如下命令，查看该任务的日志。

    ```shell
    kubectl logs -f resnetinfer1-2-xsdsf
    ```

    回显示例如下。

    ```ColdFusion
    [2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Answer[0]:  Deep learning is a subset of machine learning that uses neural networks with multiple layers to model complex relationships between
    [2025-02-24 19:13:09,331] [2269] [281472887965984] [llm] [INFO] [logging.py-331] : Generate[0] token num: (0, 20)
    ```

### 删除任务<a name="ZH-CN_TOPIC_0000002479387108"></a>

在示例YAML所在路径下，执行以下命令，删除对应的推理任务。

```shell
kubectl delete -f XXX.yaml
```

例如：

```shell
kubectl delete -f infer-310p-1usoc.yaml
```

回显示例如下：

```ColdFusion
root@ubuntu:/home/test/yaml# kubectl delete -f infer-310p-1usoc.yaml 
job "resnetinfer1-2" deleted
```

## 集成后使用<a name="ZH-CN_TOPIC_0000002479387118"></a>

本章节需要用户熟悉编程开发，以及对K8s有一定了解。如果用户已有AI平台或者想基于集群调度组件开发AI平台，需要完成以下内容：

1. 根据编程语言找到对应的K8s的[官方API库](https://github.com/kubernetes-client)。
2. 根据K8s官方提供的API库，对任务进行创建、查询、删除等操作。
3. 创建、查询或删除操作任务时，用户需要将[示例YAML](#准备任务yaml)的内容转换成K8s官方API中定义的对象，通过官方库里面提供的API发送给K8s的API Server或者将YAML内容转换为JSON格式直接发送给K8s的API Server。
