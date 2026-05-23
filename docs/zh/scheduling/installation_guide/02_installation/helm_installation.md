# 使用helm安装<a name="ZH-CN_centerIC_0000002479226452"></a>
## 安装说明<a name="ZH-CN_centerIC_0000002511346381"></a>
helm是一个用于管理Kubernetes应用程序的工具，它可以帮助用户快速部署、升级和管理Kubernetes应用程序。mindcluster helm部署安装工具可以快速部署和管理mindcluster组件。

**使用约束**
- 仅支持使用helm 3.x版本。
- 支持使用helm安装的mindcluster组件包括：
    - ascend-device-plugin
    - ascend-operator
    - ascend-for-volcano
    - clusterd
    - noded
    - npu-exporter
    - infer-operator
- docker-runtime，taskd和container-manager等组件请参考[手动安装](./manual_installation/menu_manual_installation.md)对应组件章节安装使用。

## 安装前准备<a name="ZH-CN_centerIC_0000002511346381"></a>
若环境中已经存在helm 3.x版本，可以跳过此小节。
- 安装helm前请参考[Helm版本支持策略](https://v3.helm.sh/zh/docs/v3/topics/version_skew/)查询helm与k8s间的版本兼容性，根据实际情况选择helm版本。
- 请参考[helm安装文档](https://helm.sh/zh/docs/v3/intro/install)，在管理节点安装helm命令。

安装成功后，执行helm version命令检查helm版本，回显示例如下：
```bash
version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean", GoVersion:"go1.19.5"}
```

## 执行安装<a name="ZH-CN_centerIC_0000002511346381"></a>

安装步骤如下
1. 获取mindcluster helm部署工具
    - 从[MindCluster 发行版](https://gitcode.com/Ascend/mind-cluster/releases)页面下载对应版本的部署工具压缩包Ascend-helm-deploy-tool_{version}_linux.zip。
    - 解压部署工具压缩包：
      ```bash
      unzip Ascend-helm-deploy-tool_{version}_linux.zip
      ```
      执行ls -l命令查看解压后的文件：
      ```bash
      ls -l
      ```
      回显如下：
      ```bash
      total 24
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
      -rw-r--r-- 1 root root  2026 Mar 24 15:25 add_helm_meta.sh
      ```
      > [!NOTE]
      > {version}表示mindcluster版本，如26.1.0。{chart_version}表示helm chart版本，如1.1.0。请根据实际版本替换。
      > 解压后的文件用途：
      > - mindcluster-crds-deploy-tool-{chart_version}.tgz：用于安装mindcluster组件的crd资源。
      > - mindcluster-deploy-tool-{chart_version}.tgz：用于安装mindcluster的应用组件。
      > - add_helm_meta.sh：添加helm chart元数据的脚本，安装章节可以不用执行。

2. 使用helm安装mindcluster crd资源
    - 参考[crd默认配置](#default_crds_yaml_install_config)小节，查看安装crd资源时的默认参数配置。
    - 若crd资源默认参数配置符合实际需求，可直接执行如下命令安装crd资源。
      ```bash
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
      ```
    - （可选）若crd资源默认参数配置不符合实际需求，可创建crds-values.yaml文件，修改参数配置后执行如下命令。
      - 执行如下命令
        ```bash
        helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
        ```
    - 回显示例如下，表示安装成功
      ```bash
      Release "mindcluster-crds" does not exist. Installing it now.
      NAME: mindcluster-crds
      LAST DEPLOYED: ...
      NAMESPACE: mindx-dl
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```
3. 使用helm安装mindcluster应用组件：
    - 参考[应用组件默认配置](#default_app_yaml_install_config)小节，查看安装应用组件时的默认参数配置。
    - 若应用组件默认参数配置符合实际需求，可直接执行如下命令安装应用组件。
      ```bash
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz
      ```
    - （可选）若应用组件默认参数配置不符合实际需求，可创建values.yaml文件，修改参数配置后执行如下命令。
      - 执行如下命令
        ```bash
        helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
        ```
    - 回显如下，表示安装成功
      ```bash
      Release "mindcluster" does not exist. Installing it now.
      NAME: mindcluster
      LAST DEPLOYED: ...
      NAMESPACE: mindx-dl
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```
4. 参考[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节确认组件安装状态。
5. 若组件状态异常，请确认检查安装参数配置是否正确，排查异常原因后重新安装。
   - 重新安装前执行如下命令卸载相关资源。
     ```bash
     helm uninstall mindcluster-crds # 命令卸载crd资源
     helm uninstall mindcluster # 命令卸载应用组件
     ```

## 默认配置
1. crd默认配置<a name="default_crds_yaml_install_config"></a>。安装crd时的默认配置如下所示，参数说明可参考[表1](#table15274931175241)。用户若想要自定义参数配置，可新增crds-values.yaml文件，复制以上默认配置到文件中，根据实际情况修改参数配置，在安装时使用-f crds-values.yaml指定参数配置文件。
   ```yaml
   ascend-operator-crds:
     enabled: true              # 安装ascend-operator组件的crd
   ascend-for-volcano-crds:
     enabled: true              # 安装ascend-for-volcano组件的crd
     volcanoVersion: "v1.7.0"   # volcano版本
   infer-operator-crds:
     enabled: true              # 安装infer-operator组件的crd
   ```

2. 应用组件默认配置<a name="default_app_yaml_install_config"></a>。参数说明可参考[表2](#table15274931175242)和[表3](#table15274931175243)，其中[表3](#table15274931175243)中的参数未在yaml示例中展示。用户若想要自定义参数配置，可新增values.yaml文件，复制以上默认配置到文件中，根据实际情况修改参数配置，在安装时使用-f values.yaml指定参数配置文件。

   ```yaml
   # 安装应用组件时的默认yaml配置如下
   clusterd:
     enabled: true                                                         # 是否安装clusterD组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd"   # clusterD组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # clusterD组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # clusterD组件镜像拉取策略，请根据实际情况修改

   noded:
     enabled: true                                                         # 是否安装noded组件
     enableDpc: false                                                       # 是否启用DPC功能
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/noded"      # noded组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # noded组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # noded组件镜像拉取策略，请根据实际情况修改

   npu-exporter:
     enabled: true                                                         # 是否安装npu-exporter组件
     is310P1usoc: false                                                    # 是否为310P-1usoc
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter" # npu-exporter组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # npu-exporter组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # npu-exporter组件镜像拉取策略，请根据实际情况修改

   ascend-operator:
     enabled: true                                                         # 是否安装ascend-operator组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator" # ascend-operator组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # ascend-operator组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # ascend-operator组件镜像拉取策略，请根据实际情况修改

   ascend-for-volcano:
     enabled: true                                                         # 是否安装ascend-for-volcano组件
     volcanoVersion: "v1.7.0"                                              # 设置要安装的Volcano版本
     scheduler:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler"      # vc-scheduler组件镜像名，请根据实际情况修改
         tag: "v1.7.0-v26.1.0"                                                      # vc-scheduler组件镜像标签
         pullPolicy: "IfNotPresent"                                                 # vc-scheduler组件镜像拉取策略
     controller:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager" # vc-controller-manager组件镜像名，请根   据实际情况修改
         tag: "v1.7.0-v26.1.0"                                                          # vc-controller-manager组件镜像标签，请   根据实际情况修改
         pullPolicy: "IfNotPresent"                                                     # vc-controller-manager组件镜像拉取策略

   infer-operator:
     enabled: true                                                         # 是否安装infer-operator组件
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator" # infer-operator组件镜像名，请根据实际情况修改
       tag: "v26.1.0"                                                      # infer-operator组件镜像标签，请根据实际情况修改
       pullPolicy: "IfNotPresent"                                          # infer-operator组件镜像拉取策略，请根据实际情况修改

   ascend-device-plugin:
     enabled: true                                                         # 是否安装ascend-device-plugin组件
     npuType: 910                                                          # 可选值包括：npu, 910, 310, 310P
     volcanoType: true                                                     # 是否使用volcano进行调度，请根据实际情况修改
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin" # ascend-k8sdeviceplugin组件镜像名，请根   据实际情况修改
       tag: "v26.1.0"                                                                  # ascend-k8sdeviceplugin组件镜像标签，请   根据实际情况修改
       pullPolicy: "IfNotPresent"                                                      # ascend-k8sdeviceplugin组件镜像拉取策   略，请根据实际情况修改
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
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>ascend-operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否安装ascend-operator组件的crd。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>ascend-for-volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否安装volcano的crd。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p><p>string</p><ul><li>v1.7.0</li><li>v1.9.0</li></ul></p><p>默认值为v1.7.0</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示选择安装的volcano版本。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>infer-operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否安装infer-operator组件的crd。</p></td>
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
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>clusterd</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>clusterd.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用clusterd组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>noded</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用noded组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enableDpc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否开启dpc功能。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>npu-exporter</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用npu-exporter组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.is310P1usoc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否为310P-1usoc。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>ascend-operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用ascend-operator组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>ascend-for-volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用ascend-for-volcano组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>取值范围包括：<ul><li>v1.7.0</li><li>v1.9.0</li></ul></p><p>默认值为v1.7.0</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示volcano版本。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>infer-operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用infer-operator组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="3" valign="center" headers="mcps1.2.5.1.1 "><p>ascend-device-plugin</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否启用ascend-device-plugin组件。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.npuType</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>取值范围包括：<ul><li>npu</li><li>910</li><li>310</li><li>310P</li></ul></p><p>默认值为910</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>集群节点使用的NPU卡类型。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.volcanoType</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>默认值为true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示是否使用volcano调度。</p></td>
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
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string array</p><p>不设置或设置为""则使用组件默认启动命令</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的启动命令。</p><p>若用户为开启组件某些功能需要修改启动命令（如ascend-device-plugin开启热复位等），请自行设置此参数。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.repository</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>设置为""则使用组件默认镜像名</p><p>不设置则默认配置为昇腾镜像仓库地址</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像仓库地址或镜像名。</p><p>若使用昇腾镜像仓库地址，需保证节点能够正常访问互联网，否则会因缺乏镜像导致部署后组件状态异常。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.tag</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认镜像tag</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像版本标签。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.pullPolicy</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>取值为：<ul><li>IfNotPresent</li><li>Always</li><li>Never</li></ul></p><p>设置为""则使用组件默认镜像拉取策略。</p><p>不填则默认为IfNotPresent。</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的镜像拉取策略。</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认内存请求大小</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的请求内存大小，如"512Mi"</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认CPU请求大小</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的请求CPU大小，如"500m"</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认内存限制大小</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的内存限制大小，如"1Gi"</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>不设置或设置为""则使用组件默认CPU限制大小</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>表示组件的CPU限制大小，如"1000m"</p></td>
  </tr>
</tbody>
</table>

>![](public_sys-resources/icon-note.gif) **说明：**
>表3中`<component>`取值为：clusterd、noded、npu-exporter、ascend-operator、ascend-for-volcano.scheduler、ascend-for-volcano.controller、infer-operator、ascend-device-plugin。
