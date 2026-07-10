# 代际升级适配指南

## 使用前必读

本节内容旨在为用户提供从老代际设备升级到新代际设备后，使用基础调度特性时的适配参考与指导。

## Atlas A3设备 升级至 Ascend 950代际设备

### 安装部署MindCluster

Ascend 950代际设备主要包括Atlas 350 标卡、Atlas 850 系列硬件产品以及Atlas 950 SuperPoD。

使用Ascend 950代际设备 需安装26.0.0以上版本的MindCluster组件。MindCluster各组件安装部署方式可参考[安装部署](../../03_installation_guide/menu_installation_guide.md)章节。

### 制作业务镜像

推荐从[昇腾镜像仓库](https://www.hiascend.com/developer/ascendhub)根据用户的系统架构（ARM或者x86\_64）、训练/推理框架（PyTorch、MindSpore、MindIE）以及设备的芯片型号下载所需的训练/推理基础镜像。

>[!NOTE]
>基础镜像中不包含推理模型、脚本等文件，因此，用户需要根据自己的需求进行定制化修改（如加入推理脚本代码、模型等）后才能使用。
>升级到Ascend 950代际设备需选择带有“950”代际关键词的镜像。

### 准备任务YAML

从Atlas A3设备升级至Ascend 950代际设备场景下，任务YAML可能的修改点见[表2](#zh-cn_topic_0000001609074213_table5589101114528)，若其中可选配置项未使用，可忽略。[表3](#zh-cn_topic_0000001609074213_table5589101114529)以Atlas 950 PoD产品任务为例，提供了YAML文件的具体配置样例。

**表 2**  YAML文件变更参数说明

<a name="zh-cn_topic_0000001609074213_table5589101114528"></a>
<table>
<thead>
<tr>
<th>参数</th>
<th>取值</th>
<th>说明</th>
</tr>
</thead>
<tbody>
<tr>
<td>image</td>
<td>-</td>
<td>镜像名称，请根据实际修改（用户在<a href="../../07_references/02_common_operations.md#制作镜像">制作镜像</a>章节制作的镜像名称）。</td>
</tr>
<tr>
<td>replicas</td>
<td>整数</td>
<td>运行的任务副本数量，视实际所需节点数而定。</td>
</tr>
<tr>
<td>ring-controller.atlas</td>
<td><ul><li>Atlas A3设备取值为ascend-910b</li><li>Ascend 950代际设备需修改为ascend-npu</li></ul></td>
<td>用于区分任务使用的芯片的类型。相比于Atlas A3设备，Ascend 950代际设备在芯片资源名称上存在变更。</td>
</tr>
<tr>
<td>（可选）huawei.com/schedule_policy</td>
<td>该字段取值需要参考实际的硬件型号与芯片布局：<ul><li>Atlas 350 标卡，单节点8卡，卡间无UB互联: chip1-node8</li><li>Atlas 350 标卡，单节点8卡，每4卡UB互联: chip4-node8</li><li>Atlas 350 标卡，单节点16卡，卡间无UB互联: chip1-node16</li><li>Atlas 350 标卡，单节点16卡，每4卡UB互联: chip4-node16</li><li>Atlas 850系列硬件产品（非超节点）: chip8-node8</li><li>Atlas 850系列硬件产品（超节点）: chip8-node8-sp</li><li>Atlas 950 SuperPoD: chip8-node8-ra64-sp</li></ul>关于该字段的详细说明可参考：<a href="../../06_api/01_volcano.md#podgroup">参数说明</a>中表3对huawei.com/schedule_policy字段的说明</td>
<td>配置任务需要调度的AI芯片布局形态，使用Volcano调度时可选配置该字段。Volcano会根据该字段选择合适的调度策略。</td>
</tr>
<tr>
<td>（可选）sp-block</td>
<td>指定逻辑超节点芯片数量。<p>单机时需要和任务请求的芯片数量一致。</p><p>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</p></td>
<td>仅在升级到Atlas 850系列硬件产品（超节点）与Atlas 950 PoD设备时需要保留或配置该字段。指定sp-block字段，集群调度组件会在物理超节点上根据切分策略划分出逻辑超节点，用于训练任务的逻辑超节点亲和性调度。若用户未指定该字段，Volcano调度时会将此任务的逻辑超节点大小指定为任务配置的NPU总数。<br/> 了解详细说明请参见<a href="../../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a></td>
</tr>
<tr>
<td>（可选）ra-block</td>
<td>指定逻辑框芯片数量。<p>单机时需要和任务请求的芯片数量一致。</p><p>分布式时需要是节点芯片数量的整数倍，且任务总芯片数量是其整数倍。</p></td>
<td>仅在升级到Atlas 950 PoD设备时需要配置该字段。ra-block用于指定逻辑框芯片数量，用于训练任务的逻辑框亲和性调度，若用户未指定该字段，Volcano调度时会将此任务的逻辑框大小指定为8，即不开启逻辑框亲和性调度。<br/> 了解详细说明请参见<a href="../../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md#atlas-900-a3-superpod-超节点">灵衢总线设备节点网络说明</a></td>
</tr>
<tr>
<td>requests/limits</td>
<td><ul><li>Atlas A3设备资源名统一取值为huawei.com/Ascend910，请求资源的取值范围为1-16</li><li>Ascend 950代际设备资源名需修改为huawei.com/npu，请求资源的取值范围，Atlas 850系列硬件产品与Atlas 950 SuperPoD产品为1-8，Atlas 350 标卡设备的视实际单机的NPU数量而定</li></ul></td>
<td>相比于Atlas A3设备，Ascend 950代际设备在芯片资源名称上存在变更，且Ascend 950代际设备单机NPU数视实际硬件型号存在差异。</td>
</tr>
<tr>
<td>ASCEND_VISIBLE_DEVICES</td>
<td><ul><li>Atlas A3设备：取值为metadata.annotations['huawei.com/Ascend910']</li><li>Ascend 950代际设备：取值需修改为metadata.annotations['huawei.com/npu']</li></ul>
</td>
<td><p>该字段为容器的环境变量配置。完整配置路径例如：若ASCEND_VISIBLE_DEVICES对应环境变量键值containers[0].env[0]，则环境变量值配置于containers[0].env[0].valueFrom.fieldRef.fieldPath。</p><p>Ascend Docker Runtime会获取该参数值,用于给容器挂载相应类型的NPU。相比于Atlas A3设备，Ascend 950代际设备在芯片资源名称上存在变更。</p>
</td>
</tr>
</tbody>
</table>

**表 3**  Ascend 950代际设备任务YAML文件样例参考

<a name="zh-cn_topic_0000001609074213_table5589101114529"></a>

<table>
<thead align="left">
<tr>
<th>使用场景</th>
<th>任务类型</th>
<th>硬件型号</th>
<th>使用框架</th>
<th>YAML示例文件名</th>
<th>链接</th>
</tr>
</thead>
<tbody>
<tr>
<td rowspan="6">训练</td>
<td rowspan="2">Ascend Job</td>
<td rowspan="2">Atlas 950 SuperPoD</td>
<td>PyTorch</td>
<td>pytorch_multinodes_acjob_950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/without-ranktable/pytorch/pytorch_multinodes_acjob_950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>mindspore_multinodes_acjob_950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/without-ranktable/mindspore/mindspore_multinodes_acjob_950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td rowspan="2">Volcano Job</td>
<td rowspan="2">Atlas 950 SuperPoD <br/> Atlas 850系列硬件产品（超节点） <br/> Atlas 350 标卡</td>
<td>PyTorch</td>
<td>a950_superpod_pytorch_vcjob.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_pytorch_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>a950_superpod_mindspore_vcjob.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_mindspore_vcjob.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td rowspan="2">Deployment</td>
<td rowspan="2">Atlas 950 SuperPoD <br/> Atlas 850系列硬件产品（超节点） <br/> Atlas 350 标卡</td>
<td>PyTorch</td>
<td>a950_superpod_pytorch_deployment.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_pytorch_deployment.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>a950_superpod_mindspore_deployment.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_mindspore_deployment.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td rowspan="3">推理</td>
<td>Ascend Job</td>
<td>Atlas 950 SuperPoD <br/> Atlas 850系列硬件产品（超节点） <br/> Atlas 350 标卡</td>
<td>-</td>
<td>pytorch_multinodes_acjob_infer_950_with_ranktable.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/inference/volcano/pytorch_multinodes_acjob_infer_950_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td>Volcano Job</td>
<td>Atlas 950 SuperPoD <br/> Atlas 850系列硬件产品（超节点） <br/> Atlas 350 标卡</td>
<td>-</td>
<td>infer-vcjob-950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-vcjob-950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
<tr>
<td>Deployment</td>
<td>Atlas 950 SuperPoD <br/> Atlas 850系列硬件产品（超节点） <br/> Atlas 350 标卡</td>
<td>-</td>
<td>infer-deploy-950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.0.0/samples/inference/volcano/infer-deploy-950.yaml" target="_blank" rel="noopener noreferrer">获取YAML</a></td>
</tr>
</tbody>
</table>
