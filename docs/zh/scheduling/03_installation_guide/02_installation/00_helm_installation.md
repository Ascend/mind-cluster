# 使用helm安装<a name="ZH-CN_centerIC_0000002479226452"></a>

## 安装说明<a name="ZH-CN_centerIC_0000002511346381_install_desc"></a>

helm是一个用于管理Kubernetes应用程序的工具，它可以帮助用户快速部署、升级和管理Kubernetes应用程序。MindCluster helm部署工具可以快速部署和管理MindCluster组件。

**使用约束**

- 仅支持使用helm 3.x版本。
- 支持使用helm安装的组件包括：
    - Ascend Device Plugin
    - Ascend Operator
    - Volcano
    - ClusterD
    - NodeD
    - NPU Exporter
    - Infer Operator
    - K8s RDMA Shared Dev Plugin
- 安装Container Manager组件请参考[手动安装Container Manager](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_container-manager.md#ZH-CN_TOPIC_0000002524428759)章节。
- TaskD和MindIO安装在业务容器中，不在本章节涉及的组件范围内。

## 安装前准备<a name="ZH-CN_centerIC_0000002511346381_install_prepare"></a>

1. 安装Ascend Docker Runtime<a name="zh-cn_centerIC_0000002511346381_install_prepare_docker_runtime"></a>。
   - 若未安装过Ascend Docker Runtime，请参考[手动安装Ascend Docker Runtime](../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md#ZH-CN_TOPIC_0000002479226434)章节，在所有节点上安装此组件。
   - 请参照[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002511426307)章节，在所有安装了该组件的节点上确认Ascend Docker Runtime的状态。

2. 在管理节点安装helm命令<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>。若环境中已经存在helm 3.x版本，可以跳过此步骤。
   - 安装helm前请参考[Helm版本支持策略](https://v3.helm.sh/zh/docs/v3/topics/version_skew/)查询helm与K8s间的版本兼容性，根据实际情况选择helm版本。
   - 请参考[Helm安装文档](https://helm.sh/zh/docs/v3/intro/install)，在管理节点安装helm命令。

   安装成功后，执行如下命令检查helm版本：

   ```bash
   helm version
   ```

   回显示例如下：

   ```ColdFusion
   version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean",    GoVersion:"go1.19.5"}
   ```

## 执行安装<a name="ZH-CN_centerIC_0000002511346381_install_exec"></a>

安装步骤如下：

1. 参考[创建节点标签](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#ZH-CN_TOPIC_0000002511426279)小节，给节点打标签。

   >[!NOTE]
   >- 默认日志路径无需用户手动创建，组件yaml文件initContainer命令会自动创建，默认日志路径可参考[集群调度组件日志路径列表](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#table957112617314)。
   >- 宿主机上可不新创建用户，只需要保证没有其他用户占用UID为9000的情况即可，用户信息可参考[创建用户](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#ZH-CN_TOPIC_0000002511346353)。

2. 获取MindCluster helm部署工具。
   1. 下载部署工具。

      ```bash
      # 请用户自行将命令中的{version}替换为对应版本号，如26.1.0
      wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
      ```

   2. 解压部署工具压缩包。

      ```bash
      unzip Ascend-helm-deploy-tool_{version}_linux.zip
      ```

   3. 查看解压后的文件。

      ```bash
      ls -l
      ```

      回显如下：

      ```bash
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
      -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
      ```

      > [!NOTE]
      > {version}表示MindCluster组件版本，如26.1.0。
      > {chart_version}表示helm chart版本，与MindCluster组件版本保持一致。
      > 解压后的文件用途请参考[表4](#table15274931175244)。

3. 使用helm安装mindcluster组件所需的Custom Resource Definitions（CRDs，自定义资源定义）的Release实例。
    > [!NOTE]
    >- 以下三个组件包含crd：Ascend Operator、Volcano和Infer Operator。若用户不需要安装这三个组件，可跳过此步骤。
    >- 请用户按需选择**默认配置安装**或**自定义配置安装**其中一种方式进行操作即可。
    - **默认配置安装**：若[crd默认配置](#default_crds_yaml_install_config)符合用户需求，可执行如下命令安装crd资源。

      ```bash
      #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
      # 正式执行安装
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
      ```

    - **自定义配置安装<a name="update_crd_values_before_install_crds"></a>**：若[crd默认配置](#default_crds_yaml_install_config)不符合用户需求，请创建crds-values.yaml文件，将[crd默认配置](#default_crds_yaml_install_config)的YAML文件内容复制到crds-values.yaml文件中，修改相关配置后执行如下命令：

      ```bash
      #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
      # 正式执行安装
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
      ```

      回显示例如下，表示安装成功：

      ```ColdFusion
      Release "mindcluster-crds" does not exist. Installing it now.
      NAME: mindcluster-crds
      LAST DEPLOYED: ...
      NAMESPACE: default
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```

4. 使用helm安装mindcluster应用组件的Release实例。
    > [!NOTE]
    >- **默认配置安装方式**会从昇腾镜像仓库下载应用组件的镜像。若用户节点无法连接互联网且本地未缓存镜像，可能会导致安装失败。
    >- 请用户按需选择**默认配置安装**或**自定义配置安装**其中一种方式进行操作即可。
    - **默认配置安装**：若[应用组件默认配置](#default_app_yaml_install_config)符合用户需求，可执行如下命令安装应用组件。

      ```bash
      #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
      # 正式执行安装
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz
      ```

    - **自定义配置安装<a name="update_app_values_before_install_app"></a>**：若[应用组件默认配置](#default_app_yaml_install_config)不符合用户需求，请创建values.yaml文件，将[应用组件默认配置](#default_app_yaml_install_config)的YAML文件内容复制到values.yaml文件中，修改相关配置。例如，修改Ascend Device Plugin组件的镜像名、日志路径和日志级别的配置示例如下：

        ```yaml
        ...
        ascend-device-plugin:
          enabled: true
          is310P1usoc: false
          volcanoType: true
          image:
            repository: "ascend-k8sdeviceplugin" # 修改Ascend Device Plugin镜像名
            tag: "v26.1.0"
            pullPolicy: "IfNotPresent"
          args: [ "device-plugin -volcanoType=true -presetVirtualDevice=true -logFile=/tmp/devicePlugin.log -logLevel=-1 --enable-healthz=true --healthz-address=11251" ] #日志路径改为/tmp/devicePlugin.log，日志级别改为Debug级别。
        ...
        ```

      然后执行如下命令：

      ```bash
      #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
      # 正式执行安装
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
      ```

      回显示例如下，表示安装成功：

      ```ColdFusion
      Release "mindcluster" does not exist. Installing it now.
      NAME: mindcluster
      LAST DEPLOYED: ...
      NAMESPACE: default
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```

5. 参考[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节确认组件安装状态。
6. 若组件状态异常，请确认检查安装配置是否正确，排查异常原因后重新安装。
   - 重新安装前执行如下命令卸载相关资源。

     ```bash
     helm uninstall mindcluster-crds # 命令卸载crd
     helm uninstall mindcluster # 命令卸载应用组件
     ```

## 默认配置

- crd默认配置<a name="default_crds_yaml_install_config"></a>。
    > [!NOTE]
    >- 默认安装crd的组件包括：Infer Operator、Volcano、Ascend Operator，Volcano版本为v1.9.0。
    >- 参数说明可参见[表1](#table15274931175241)。

   ```yaml
   ascend-operator-crds:
     enabled: true              # 安装ascend-operator组件的crd
   ascend-for-volcano-crds:
     enabled: true              # 安装ascend-for-volcano组件的crd
     volcanoVersion: "v1.9.0"   # volcano crd的版本
   infer-operator-crds:
     enabled: true              # 安装infer-operator组件的crd
   ```

- 应用组件默认配置<a name="default_app_yaml_install_config"></a>。
    > [!NOTE]
    >- 默认安装的组件包括：Ascend Device Plugin、Ascend Operator、Volcano、ClusterD、NodeD、NPU Exporter和Infer Operator，Volcano版本为v1.9.0。
    >- 默认不安装的组件包括：K8s RDMA Shared Dev Plugin。
    >- 参数说明可参见[表2](#table15274931175242)和[表3](#table15274931175243)，其中[表3](#table15274931175243)中的参数未在下方YAML配置中展示，用户可根据实际情况新增或修改。

   ```yaml
   # 安装应用组件时的默认yaml配置如下
   clusterd:
     enabled: true                                                         # 安装ClusterD组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd"   # ClusterD组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # ClusterD组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # ClusterD组件镜像拉取策略，请根据实际情况修改

   noded:
     enabled: true                                                         # 安装NodeD组件
     enabledStorageCheck: ""                                                # 开启的共享存储故障检测类型，为空表示不开启
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/noded"      # NodeD组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # NodeD组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # NodeD组件镜像拉取策略，请根据实际情况修改

   npu-exporter:
     enabled: true                                                         # 安装NPU Exporter组件
     is310P1usoc: false                                                    # false表示产品不是Atlas 200I SoC A1 核心板
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter" # NPU Exporter组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # NPU Exporter组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # NPU Exporter组件镜像拉取策略，请根据实际情况修改

   ascend-operator:
     enabled: true                                                         # 安装Ascend Operator组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator" # Ascend Operator组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # Ascend Operator组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # Ascend Operator组件镜像拉取策略，请根据实际情况修改

   ascend-for-volcano:
     enabled: true                                                         # 安装Volcano组件
     volcanoVersion: "v1.9.0"                                              # 设置要安装的Volcano版本
     scheduler:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler"      # Volcano Scheduler组件镜像名，请根据实际情况修改
         tag: "v1.9.0-v26.1.0"                                                      # Volcano Scheduler组件镜像标签
         pullPolicy: "IfNotPresent"                                                 # Volcano Scheduler组件镜像拉取策略
     controller:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager" # Volcano Controller组件镜像名，请根据实际情况修改
         tag: "v1.9.0-v26.1.0"                                                          # Volcano Controller组件镜像标签，请根据实际情况修改
         pullPolicy: "IfNotPresent"                                                     # Volcano Controller组件镜像拉取策略

   infer-operator:
     enabled: true                                                         # 安装Infer Operator组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator" # Infer Operator组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # Infer Operator组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # Infer Operator组件镜像拉取策略，请根据实际情况修改

   ascend-device-plugin:
     enabled: true                                                         # 安装Ascend Device Plugin组件
     is310P1usoc: false                                                    # false表示产品不是Atlas 200I SoC A1 核心板
     volcanoType: true                                                     # true表示使用volcano进行调度，请根据实际情况修改
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin" # Ascend Device Plugin组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                                  # Ascend Device Plugin组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                                      # Ascend Device Plugin组件镜像拉取策略，请根据实际情况修改

   k8s-rdma-shared-dev-plugin:
     enabled: false                                                           # false表示不安装K8s RDMA Shared Dev Plugin组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dp" # K8s RDMA Shared Dev Plugin组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                              # K8s RDMA Shared Dev Plugin组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                                  # K8s RDMA Shared Dev Plugin组件镜像拉取策略，请根据实际情况修改
   ```

## 参数说明

**表 1**  crd资源的可配置参数说明
<a name="table15274931175241"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>crd所属组件</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>参数名称</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>取值类型</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.3"><p>说明</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Ascend Operator组件的crd。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>Volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Volcano组件的crd。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p><p>string</p><ul><li>v1.7.0</li><li>v1.9.0</li><li>v1.12.0</li></ul></p><p>默认值为v1.9.0</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>选择Volcano版本。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Infer Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Infer Operator组件的crd。</p></td>
  </tr>
</tbody>
</table>

**表 2**  组件的可配置参数说明
<a name="table15274931175242"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>所属组件</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>参数名称</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>取值类型</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.3"><p>说明</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>ClusterD</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>clusterd.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用ClusterD组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>NodeD</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用NodeD组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enabledStorageCheck</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为空</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>开启的共享存储故障检测类型，值域包括""、"dpc"、"dtfs"、"dpc,dtfs"、"container-snapshot"。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>NPU Exporter</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用NPU Exporter组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.is310P1usoc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示产品为Atlas 200I SoC A1 核心板。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Ascend Operator组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>Volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Volcano组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>取值包括：<ul><li>v1.7.0</li><li>v1.9.0</li><li>v1.12.0</li></ul></p><p>默认值为v1.9.0</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>选择启用的volcano版本。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Infer Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Infer Operator组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="3" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Device Plugin</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用Ascend Device Plugin组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.is310P1usoc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示产品为Atlas 200I SoC A1 核心板。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.volcanoType</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示使用volcano进行调度。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>K8s RDMA Shared Dev Plugin</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>k8s-rdma-shared-dev-plugin.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>设置为true表示启用K8s RDMA Shared Dev Plugin组件。</p></td>
  </tr>
</tbody>
</table>

**表 3**  组件的其他可配置参数
<a name="table15274931175243"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.51.1"><p>参数名称</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>取值范围</p></th>
    <th class="cellrowborder" valign="center" width="50%" id="mcps1.2.5.1.3"><p>说明</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.args</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string array</p><p>不设置或设置为""则使用组件默认启动命令。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的启动命令。</p><p>若用户为开启组件某些功能需要修改启动命令（如ascend-device-plugin开启热复位等），请自行设置此参数。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.repository</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><ul><li>设置为""则使用组件默认镜像名。</li><li>不设置则默认配置为昇腾镜像仓库地址。</li></ul></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像仓库地址或镜像名。</p><p>若使用昇腾镜像仓库地址，需保证节点能够正常访问互联网，否则会因缺乏镜像导致部署后组件状态异常。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.tag</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认镜像tag。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像版本标签。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.pullPolicy</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>取值为：<ul><li>IfNotPresent</li><li>Always</li><li>Never</li></ul></p><p>设置为""则使用组件默认镜像拉取策略。</p><p>不填则默认为IfNotPresent。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像拉取策略。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认内存请求大小。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的请求内存大小，如"512Mi"。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认CPU请求大小。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的请求CPU大小，如"500m"。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认内存限制大小。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的内存限制大小，如"1Gi"。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认CPU限制大小。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的CPU限制大小，如"1000m"。</p></td>
  </tr>
</tbody>
</table>

>[!NOTE]
> 表3中`<component>`取值为：clusterd、noded、npu-exporter、ascend-operator、**ascend-for-volcano.scheduler**、**ascend-for-volcano.controller**、infer-operator、ascend-device-plugin、k8s-rdma-shared-dev-plugin。

**表 4**  Helm部署工具压缩包文件列表说明
<a name="table15274931175244"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>文件名</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>文件用途</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>说明</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>mindcluster-crds-deploy-tool-{chart_version}.tgz</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>Helm Chart打包文件，用于在K8s集群中部署和管理MindCluster各组件所需的Custom Resource Definitions（CRDs，自定义资源定义）的部署工具。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>用户可配置的安装参数说明可参见<a href="#table15274931175241">表1</a>。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>mindcluster-deploy-tool-{chart_version}.tgz</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>Helm Chart打包文件，用于在K8s集群中部署和管理MindCluster项目各组件的部署工具。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>用户可配置的安装参数说明可参见<a href="#table15274931175242">表2</a>和<a href="#table15274931175243">表3</a>。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>helm_tool.sh</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>作用包括：<ul><li><p>给各组件资源添加helm chart元数据的脚本。</p></li><li>删除Ascend Device Plugin组件26.1.0版本前的DaemonSet资源</li></ul></p><p>仅在升级时使用。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 ">脚本会为以下资源打上helm 元数据，包括：<ul><li>Ascend Operator组件相关资源</li><li>Ascend Device Plugin组件相关资源</li><li>Volcano组件相关资源</li><li>ClusterD组件相关资源</li><li>NodeD组件相关资源</li><li>NPU Exporter组件相关资源</li><li>Infer Operator组件相关资源</li><li>K8s RDMA Shared Dev Plugin组件相关资源</li><li>命令空间，包括"mindx-dl"和"cluster-system"</li></ul></td>
  </tr>
</tbody>
</table>

 > [!NOTE]
 > {chart_version}表示helm chart版本，与MindCluster组件版本保持一致。
