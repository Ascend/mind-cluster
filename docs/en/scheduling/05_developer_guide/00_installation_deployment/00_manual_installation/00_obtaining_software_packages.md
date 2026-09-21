# Obtaining Software Packages<a name="ZH-CN_TOPIC_0000002479386476"></a>

To obtain the corresponding software, see [Downloading Software Packages](#section10979172103311); to obtain the source-code of the corresponding software packages, see [Open-Source Component Source Code](#section149534517468).

## Downloading Software Packages<a name="section10979172103311"></a>

>[!NOTE]
><i>\{version\}</i> indicates the software version number, for example, v26.1.0. <i>\{arch\}</i> indicates the CPU architecture, which can be x86_64 or aarch64.

**Table 1** Software packages of each component

<a name="table13465342493"></a>

| Component | Software Package  | Description | Download Link |
|--|--|--|--|
|Ascend Docker Runtime|Ascend-docker-runtime\_<i>{version}</i>\_linux-<i>{arch}</i>.run|Ascend Docker Runtime package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|NPU Exporter|Ascend-mindxdl-npu-exporter\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|NPU Exporter package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Ascend Device Plugin|Ascend-mindxdl-device-plugin\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Ascend Device Plugin package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Volcano|Ascend-mindxdl-volcano\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Volcano package. Volcano includes volcano-controller and volcano-scheduler.<p>Refer to the target version based on the Kubernetes and open-source Volcano compatibility. For details, see [Kubernetes versions in official Volcano website](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility).</p><ul><li>Volcano v1.7.0 is compatible with Kubernetes 1.19.x to 1.28.x.</li><li>Volcano v1.9.0 is compatible with Kubernetes 1.21.x to 1.29.x.</li><li>Volcano v1.12.0 is compatible with Kubernetes 1.21.x to 1.34.x.</li></ul>|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Infer Operator|Ascend-mindxdl-infer-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Infer Operator package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Ascend Operator|Ascend-mindxdl-ascend-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Ascend Operator package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|NodeD|Ascend-mindxdl-noded\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|NodeD package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|ClusterD|Ascend-mindxdl-clusterd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|ClusterD package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|TaskD|Ascend-mindxdl-taskd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|TaskD package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Container Manager|Ascend-mindxdl-container-manager\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Container Manager package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|MindIO|Ascend-mindxdl-mindio\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|MindIO package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|K8s RDMA Shared Dev Plugin|Ascend-mindxdl-k8s-rdma-shared-dev-plugin\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|K8s RDMA Shared Dev Plugin package|[Click here](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Resilience Controller|Ascend-mindxdl-resilience-controller\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Resilience Controller package|[Click here](https://www.hiascend.com/en/developer/software/mindcluster/download?versionId=460&ids=53%2C154%2C%2C58%2C60%2C63%2C)|
|Elastic Agent|Ascend-mindxdl-elastic\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Elastic Agent package|[Click here](https://www.hiascend.com/en/developer/software/mindcluster/download?versionId=460&ids=53%2C154%2C%2C58%2C60%2C63%2C)|

>[!NOTE]
>Resilience Controller and Elastic Agent have reached their end of life in version 7.3.0. Please obtain packages from versions prior to 7.3.0.

## Verifying Software Package SUM Value<a name="section51703441649"></a>

To prevent software packages from being maliciously tampered with during transmission or storage, perform integrity verification when downloading the software package.
>[!NOTE]
><i>\{version\}</i> indicates the software version number, for example, v26.1.0. <i>\{arch\}</i> indicates the CPU architecture, which can be x86_64 or aarch64.

1. Download the [checksum file](https://gitcode.com/Ascend/mind-cluster/releases) corresponding to the software package. The file name ends with `.sha256sum`.
2. Place the software package and the checksum file in the same directory, and run the following command to perform the verification.
   Taking the Ascend Device Plugin component as an example, the directory structure is as follows:

   ```CodeFusion
   .
   ├── Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip  // Software package
   └── Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip.sha256sum  // Digital signature file
   ```

   Run the following command to verify the integrity of the software package:

   ```shell
   sha256sum -c Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip.sha256sum # Please replace {version} and {arch} with actual values
   ```

   If the output is as follows, the verification has passed:

   ```CodeFusion
   Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip: OK
   ```

## Open-Source Component Source Code<a name="section149534517468"></a>

The source code of open-source components like Ascend Docker Runtime, NPU Exporter, Ascend Device Plugin, K8s RDMA Shared Dev Plugin, Volcano, Ascend Operator, NodeD, and ClusterD is provided. If you need to access the source code or customize component development, obtain the source code for the corresponding components according to [Table 2](#table978944123012).

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
|K8s RDMA Shared Dev Plugin|<https://gitcode.com/Ascend/mind-cluster/tree/master/component/k8s-rdma-shared-dev-plugin>|
