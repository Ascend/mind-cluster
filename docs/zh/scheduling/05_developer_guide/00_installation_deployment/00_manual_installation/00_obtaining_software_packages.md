# 获取软件包<a name="ZH-CN_TOPIC_0000002479386476"></a>

获取相应的软件可参见[下载软件包](#section10979172103311)；获取相应软件包的源码可参见[开源组件源码](#section149534517468)进行操作。

## 下载软件包<a name="section10979172103311"></a>

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

>[!NOTE]
><i>\{version\}</i>表示软件版本号，例如 v26.1.0。<i>\{arch\}</i>表示CPU架构，取值包括 x86_64 或 aarch64。

**表 1**  各组件软件包

<a name="table13465342493"></a>

|组件名称|软件包名称|说明|获取链接|
|--|--|--|--|
|Ascend Docker Runtime|Ascend-docker-runtime\_<i>{version}</i>\_linux-<i>{arch}</i>.run|Ascend Docker Runtime软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|NPU Exporter|Ascend-mindxdl-npu-exporter\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|NPU Exporter软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Ascend Device Plugin|Ascend-mindxdl-device-plugin\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Ascend Device Plugin软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Volcano|Ascend-mindxdl-volcano\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Volcano软件包。Volcano包含volcano-controller和volcano-scheduler两个容器。<p>请根据K8s和开源Volcano的兼容性选择合适的版本进行安装，具体版本请参见[Volcano官网中对应的Kubernetes版本](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility)。</p><ul><li>Volcano v1.7.0兼容的K8s版本范围为1.19.x~1.28.x。</li><li>Volcano v1.9.0兼容的K8s版本范围为1.21.x~1.29.x。</li><li>Volcano v1.12.0兼容的K8s版本范围为1.21.x~1.34.x。</li></ul>|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Infer Operator|Ascend-mindxdl-infer-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Infer Operator软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Ascend Operator|Ascend-mindxdl-ascend-operator\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Ascend Operator软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|NodeD|Ascend-mindxdl-noded\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|NodeD软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|ClusterD|Ascend-mindxdl-clusterd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|ClusterD软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|TaskD|Ascend-mindxdl-taskd\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|TaskD软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Container Manager|Ascend-mindxdl-container-manager\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Container Manager软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|MindIO|Ascend-mindxdl-mindio\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|MindIO软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|K8s RDMA Shared Dev Plugin|Ascend-mindxdl-k8s-rdma-shared-dev-plugin\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|K8s RDMA Shared Dev Plugin软件包。|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases/v26.1.0)|
|Resilience Controller|Ascend-mindxdl-resilience-controller\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Resilience Controller软件包。|[获取链接](https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann)|
|Elastic Agent|Ascend-mindxdl-elastic\_<i>{version}</i>\_linux-<i>{arch}</i>.zip|Elastic Agent软件包。|[获取链接](https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann)|

>[!NOTE]
>Resilience Controller和Elastic Agent组件已经在7.3.0版本日落，请获取7.3.0之前版本的软件包。

## 软件包 SUM 值校验步骤<a name="section51703441649"></a>

为了防止软件包在传递过程中或存储期间被恶意篡改，下载软件包时需进行完整性验证。
>[!NOTE]
><i>\{version\}</i>表示软件版本号，例如 v26.1.0。<i>\{arch\}</i>表示CPU架构，取值包括 x86_64 或 aarch64。

1. 下载软件包对应的[校验和文件](https://gitcode.com/Ascend/mind-cluster/releases)，文件名以`.sha256sum`结尾。
2. 将软件包与校验和文件放在同一目录下，执行以下命令进行校验。
   以Ascend Device Plugin组件为例，目录结构如下：

   ```CodeFusion
   .
   ├── Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip  // 软件包
   └── Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip.sha256sum  // 数字签名文件
   ```

   执行如下命令校验软件包完整性：

   ```shell
   sha256sum -c Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip.sha256sum
   ```

   回显如下，则校验通过：

   ```CodeFusion
   Ascend-mindxdl-device-plugin_{version}_linux-{arch}.zip: OK
   ```

## 开源组件源码<a name="section149534517468"></a>

集群调度提供Ascend Docker Runtime、NPU Exporter、Ascend Device Plugin、K8s RDMA Shared Dev Plugin、Volcano、Ascend Operator、NodeD和ClusterD等开源组件。如果用户需要了解源码或定制开发组件，则可根据[表2](#table978944123012)获取相应组件源码。

**表 2**  获取组件源码

<a name="table978944123012"></a>

|组件名|源码地址|
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
