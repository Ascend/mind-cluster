# Obtaining Software Packages<a name="ZH-CN_TOPIC_0000002479386476"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-26T11:47:01.572Z pushedAt=2026-06-27T00:32:25.552Z -->

To obtain the corresponding software, see [Downloading Software Packages](#section10979172103311); to obtain the source-code of the corresponding software packages, see [Open-Source Component Source Code](#section149534517468).

## Downloading Software Packages<a name="section10979172103311"></a>

>[!NOTE]
><i>\{version\}</i> indicates the software version number, and <i>\{arch\}</i> indicates the CPU architecture.

**Table 1** Software packages of each component

<a name="table13465342493"></a>

| Component | Software Package  | Description | Download Link |
|--|--|--|--|
| Ascend Docker Runtime | Ascend-docker-runtime\_<i>{version}</i>\_linux-<i>{arch}</i>.run | Ascend Docker Runtime package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| NPU Exporter | Ascend-mindxdl-npu-exporter\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | NPU Exporter package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Ascend Device Plugin | Ascend-mindxdl-device-plugin\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Ascend Device Plugin package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Volcano | Ascend-mindxdl-volcano\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Volcano package.<p>Select an appropriate version for installation based on the compatibility between K8s and open-source Volcano. For specific versions, see [the official Volcano official website](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility).</p><ul><li>The K8s version range compatible with Volcano v1.7.0 is 1.19.x to 1.28.x.</li><li>The K8s version range compatible with Volcano v1.9.0 is 1.21.x to 1.28.x.</li></ul>| [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Infer Operator | Ascend-mindxdl-infer-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Infer Operator package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Ascend Operator | Ascend-mindxdl-ascend-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Ascend Operator package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| NodeD | Ascend-mindxdl-noded\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | NodeD package| [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| ClusterD | Ascend-mindxdl-clusterd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | ClusterD package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| TaskD | Ascend-mindxdl-taskd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | TaskD package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Container Manager | Ascend-mindxdl-container-manager\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Container Manager package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| MindIO | Ascend-mindxdl-mindio\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | MindIO package | [Download Link](https://gitcode.com/Ascend/mind-cluster/releases/v26.0.0) |
| Resilience Controller | Ascend-mindxdl-resilience-controller\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Resilience Controller package | [Download Link](https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann) |
| Elastic Agent | Ascend-mindxdl-elastic\_<i>{version}</i>\_linux-<i>{arch}</i>.zip | Elastic Agent package| [Download Link](https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann) |

>[!NOTE]
>Resilience Controller and Elastic Agent have reached their end of life in version 7.3.0. Please obtain packages from versions prior to 7.3.0.

## Open-Source Component Source Code<a name="section149534517468"></a>

The cluster scheduling system provides open-source components such as Ascend Docker Runtime, NPU Exporter, Ascend Device Plugin, Volcano, Ascend Operator, NodeD, and ClusterD. If you need to understand the source code or customize component development, obtain the corresponding component source code according to [Table 2](#table978944123012).

**Table 2**  Component source code

<a name="table978944123012"></a>

|Component|Source Code Address|
|--|--|
|Ascend Docker Runtime|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-docker-runtime>|
|NPU Exporter|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/npu-exporter>|
|Ascend Device Plugin|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-device-plugin>|
|Volcano|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-for-volcano>|
|Ascend Operator|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-operator>|
|NodeD|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/noded>|
|ClusterD|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/clusterd>|
|TaskD|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/taskd>|
|Container Manager|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/container-manager>|
|Infer Operator|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/infer-operator>|
|MindIO|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/mindio>|
