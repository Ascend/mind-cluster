# Intergenerational Upgrade Adaptation Guide

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-29T08:00:13.539Z pushedAt=2026-08-29T08:01:03.072Z -->

## Before You Start

This section provides adaptation references and guidance for using basic scheduling features after upgrading from a previous-generation product to a next-generation product.

## <term>Atlas A3 Inference Products</term>/<term>Atlas A3 Training Products</term> to <term>Ascend 950 Products</term>

### Installing and Deploying MindCluster

To use <term>Ascend 950 products</term>, you need to install MindCluster components of version 26.0.0 or later. For details about how to install and deploy MindCluster components, see [Installation and Deployment](../../03_installation_guide/menu_installation_guide.md).

### Creating a Service Image

It is recommended that you download the required training/inference base image from the [Ascend image repository](https://www.hiascend.com/developer/ascendhub) based on your system architecture (Arm or x86\_64), training/inference framework (PyTorch, MindSpore, or MindIE), and the hardware model of your device.

>[!NOTE]
>The base image does not contain files such as inference models or scripts. Therefore, you need to customize it according to your requirements (for example, adding inference script code and models) before using it.
>To upgrade to the <term>Ascend 950 products</term>, select an image with the keyword "950".

### Preparing the Job YAML

When upgrading from <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term> to <term>Ascend 950 products</term>, see [Table 1](#zh-cn_topic_0000001609074213_table5589101114528) for possible modifications to the job YAML. If the optional configuration items are not used, they can be ignored. [Table 2](#zh-cn_topic_0000001609074213_table5589101114529) provides a specific YAML configuration example using the Atlas 950 SuperPoD as an example.

**Table 1** Description of YAML file parameter changes

<a name="zh-cn_topic_0000001609074213_table5589101114528"></a>
<table>
<thead>
<tr>
<th>Parameter</th>
<th>Value</th>
<th>Description</th>
</tr>
</thead>
<tbody>
<tr>
<td>image</td>
<td>-</td>
<td>Image name. Modify it based on the actual situation (the image name created by the user in the <a href="../../07_references/02_common_operations.md#creating-an-image">Creating an Image</a> section).</td>
</tr>
<tr>
<td>replicas</td>
<td>Integer</td>
<td>Number of job replicas to run, determined by the actual number of nodes required.</td>
</tr>
<tr>
<td>ring-controller.atlas</td>
<td><ul><li>For <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>, set the value to ascend-910b.</li><li>For <term>Ascend 950 products</term>, change the value to ascend-npu.</li></ul></td>
<td>Used to distinguish the type of chip used by the job. Compared with <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>, the <term>Ascend 950 products</term> has changes in the chip resource name.</td>
</tr>
<tr>
<td>(Optional) huawei.com/schedule_policy</td>
<td>The value of this field must be determined based on the actual hardware model and chip layout: <ul><li>Atlas 350 accelerator card, single node with 8 chips, no UB interconnect between chips: chip1-node8</li><li>Atlas 350 accelerator card, single node with 8 chips, UB interconnect for every 4 chips: chip4-node8</li><li>Atlas 350 accelerator card, single node with 16 chips, no UB interconnect between chips: chip1-node16</li><li>Atlas 350 accelerator card, single node with 16 chip, UB interconnect for every 4 chips: chip4-node16</li><li>Atlas 650E server: chip8-node8</li><li>Atlas 850E SuperPoD: chip8-node8-sp</li><li>Atlas 950 SuperPoD: chip8-node8-ra64-sp</li></ul>For detailed description of this field, refer to the description of the "huawei.com/schedule_policy" field in Table 3 under <a href="../../06_api/01_volcano.md#podgroup">Parameter Description</a>.</td>
<td>Configures the AI chip layout form that the job needs to schedule. This field is optional when Volcano scheduling is used. Volcano selects an appropriate scheduling policy based on this field.</td>
</tr>
<tr>
<td>(Optional) sp-block</td>
<td>Specifies the number of chips in a logical SuperPoD.<p>For a single node, it must be consistent with the number of chips requested by the job.</p><p>For distributed deployment, it must be an integer multiple of the number of chips per node, and the total number of chips for the job must be an integer multiple of it.</p></td>
<td>This field needs to be retained or configured only when upgrading to the Atlas 850E SuperPoD and the Atlas 950 SuperPoD. When the sp-block field is specified, the cluster scheduling component divides the physical SuperPoD into logical SuperPoDs based on the splitting policy, which are used for logical SuperPoD affinity scheduling of training job. If you do not specify this field, Volcano scheduling sets the logical SuperPoD size of this job to the total number of NPUs configured for the job.<br> For details, see <a href="../../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md">UnifiedBus Network Description</a></td>
</tr>
<tr>
<td>(Optional) ra-block</td>
<td>Specifies the number of chips in a logical rack.<p>For a single node, it must be consistent with the number of chips requested by the job.</p><p>For distributed deployment, it must be an integer multiple of the number of chips per node, and the total number of chips for the job must be an integer multiple of it.</p></td>
<td>This field needs to be configured only when upgrading to the Atlas 950 SuperPoD. ra-block is used to specify the number of chips in a logical rack for logical rack affinity scheduling of training jobs. If you do not specify this field, Volcano scheduling sets the logical rack size of this job to 8, that is, logical rack affinity scheduling is disabled.<br> For details, see <a href="../../04_usage/03_basic_scheduling/01_affinity_scheduling/03_ascend_ai_processor_based_affinity.md">UnifiedBus Network Description</a></td>
</tr>
<tr>
<td>requests/limits</td>
<td><ul><li>For <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>, the resource name is uniformly set to huawei.com/Ascend910, and the requested resource value ranges from 1 to 16.</li><li>For <term>Ascend 950 products</term>, the resource name must be changed to huawei.com/npu. The requested resource value range: <ul><li>Atlas 850E SuperPoD, Atlas 650E server, and Atlas 950 SuperPoD: 1 to 8.</li><li>Atlas 350 accelerator card: determined by the actual number of NPUs on a single node.</li></ul></li></ul></td>
<td>Compared with <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>, <term>Ascend 950 products</term> have changes in the chip resource name, and the number of NPUs on a single node of <term>Ascend 950 products</term> varies depending on the actual hardware model.</td>
</tr>
<tr>
<td>ASCEND_VISIBLE_DEVICES</td>
<td><ul><li><term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>: The value is metadata.annotations['huawei.com/Ascend910']</li><li><term>Ascend 950 products</term>: The value must be changed to metadata.annotations['huawei.com/npu']</li></ul>
</td>
<td><p>This field configures the environment variables of the container. For example, for the complete configuration path: if ASCEND_VISIBLE_DEVICES corresponds to containers[0].env[0], the environment variable is configured in containers[0].env[0].valueFrom.fieldRef.fieldPath.</p><p>Ascend Docker Runtime obtains the value of this parameter to mount the corresponding type of NPU to the container. Compared with <term>Atlas A3 inference products</term>/<term>Atlas A3 training products</term>, <term>Ascend 950 products</term> have changes in the chip resource name.</p>
</td>
</tr>
</tbody>
</table>

**Table 2** Sample reference for the job YAML file of <term>Ascend 950 products</term>

<a name="zh-cn_topic_0000001609074213_table5589101114529"></a>

<table>
<thead align="left">
<tr>
<th>Usage Scenario</th>
<th>Job Type</th>
<th>Hardware Model</th>
<th>Framework</th>
<th>YAML Sample File Name</th>
<th>Download Link</th>
</tr>
</thead>
<tbody>
<tr>
<td rowspan="6">Training</td>
<td rowspan="2">Ascend Job</td>
<td rowspan="2">Atlas 950 SuperPoD</td>
<td>PyTorch</td>
<td>pytorch_multinodes_acjob_950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/without-ranktable/pytorch/pytorch_multinodes_acjob_950.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>mindspore_multinodes_acjob_950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/without-ranktable/mindspore/mindspore_multinodes_acjob_950.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td rowspan="2">Volcano Job</td>
<td rowspan="2">Atlas 950 SuperPoD <br> Atlas 850E SuperPoD <br> Atlas 350 accelerator card</td>
<td>PyTorch</td>
<td>a950_superpod_pytorch_vcjob.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_pytorch_vcjob.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>a950_superpod_mindspore_vcjob.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_mindspore_vcjob.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td rowspan="2">Deployment</td>
<td rowspan="2">Atlas 950 SuperPoD <br> Atlas 850E SuperPoD <br> Atlas 350 accelerator card</td>
<td>PyTorch</td>
<td>a950_superpod_pytorch_deployment.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_pytorch_deployment.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td>MindSpore</td>
<td>a950_superpod_mindspore_deployment.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/train/basic-training/ranktable/yaml/950/a950_superpod_mindspore_deployment.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td rowspan="3">Inference</td>
<td>Ascend Job</td>
<td>Atlas 950 SuperPoD <br> Atlas 850E SuperPoD <br> Atlas 350 accelerator card</td>
<td>-</td>
<td>pytorch_multinodes_acjob_infer_950_with_ranktable.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/inference/volcano/pytorch_multinodes_acjob_infer_950_with_ranktable.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td>Volcano Job</td>
<td>Atlas 950 SuperPoD <br> Atlas 850E SuperPoD <br> Atlas 350 accelerator card</td>
<td>-</td>
<td>infer-vcjob-950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/inference/volcano/infer-vcjob-950.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
<tr>
<td>Deployment</td>
<td>Atlas 950 SuperPoD <br> Atlas 850E SuperPoD <br> Atlas 350 accelerator card</td>
<td>-</td>
<td>infer-deploy-950.yaml</td>
<td><a href="https://gitcode.com/Ascend/mindcluster-deploy/blob/branch_v26.1.0/samples/inference/volcano/infer-deploy-950.yaml" target="_blank" rel="noopener noreferrer">Obtain YAML</a></td>
</tr>
</tbody>
</table>
