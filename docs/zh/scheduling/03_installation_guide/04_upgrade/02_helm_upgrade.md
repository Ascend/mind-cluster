# 使用helm升级<a name="ZH-CN_TOPIC_0000002479226453"></a>

## 升级说明<a name="section_upgrade_desc"></a>

本文档介绍如何通过helm升级mindcluster组件。支持使用helm升级的mindcluster组件包括：

- ascend-device-plugin
- ascend-operator
- ascend-for-volcano
- clusterd
- noded
- npu-exporter
- infer-operator

>[!NOTE]
>
>- 仅支持使用helm 3.x版本进行升级。
>- docker-runtime、taskd和container-manager等组件不支持通过helm管理，请参考[手动升级](01_manual_upgrade.md)对应组件章节进行升级。
>- 升级前请确认集群中无正在使用mindcluster组件管理的工作负载，避免业务中断。

## 确认组件是否通过helm管理<a name="section_check_helm_upgrade"></a>

在执行升级前，请先确认待升级的组件是否已通过helm管理，以选择对应的升级方式。

1. 以root用户登录K8s管理节点。

2. 执行以下命令，查看当前集群中通过helm管理的Release列表。

    ```bash
    helm list -A
    ```

3. 根据回显结果判断组件的安装方式。

    - 若回显中存在名称为**mindcluster**和**mindcluster-crds**的Release，且STATUS为**deployed**，表示组件已通过helm管理，请参见[helm upgrade升级](#section_helm_upgrade)进行升级。
    - 若回显中不存在上述Release，表示组件未通过helm管理，请参见[接管资源后升级](#section_kubectl_to_helm)进行升级。

## 接管资源后升级<a name="section_kubectl_to_helm"></a>

若组件是通过kubectl手动安装的，尚未纳入helm管理，需要先接管资源，再使用helm install升级到新版本。请先从[MindCluster 发行版](https://gitcode.com/Ascend/mind-cluster/releases)页面下载对应版本的部署工具压缩包Ascend-helm-deploy-tool_{version}_linux.zip并解压，获取add_helm_meta.sh脚本和tgz安装包。
1. 执行以下命令，接管资源并安装新版本组件。
>[!NOTE]
>
>- 若需要自定义参数配置，可参考[yaml默认配置](../02_installation/helm_installation.md#默认配置)分别创建crds-values.yaml和values.yaml文件，在升级时使用`-f crds-values.yaml`和`-f values.yaml`指定。参数说明请参考[参数说明](../02_installation/helm_installation.md#参数说明)章节。
   - **helm 3.17以下版本**：使用add_helm_meta.sh脚本为已有资源添加helm元数据后，再执行helm install。

       ```bash
       dos2unix add_helm_meta.sh && chmod +x add_helm_meta.sh
       bash add_helm_meta.sh all
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz # 可增加-f crds-values.yaml指定自定义参数
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz # 可增加-f values.yaml指定自定义参数
       ```

   - **helm 3.17及以上版本**：除上述方式外，还可在helm install时通过--takeover-ship参数自动接管已有资源。

       ```bash
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --takeover-ship # 可增加-f crds-values.yaml指定自定义参数
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz --takeover-ship # 可增加-f values.yaml指定自定义参数
       ```
 2. 确认组件升级状态，请参考[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。

## helm upgrade升级<a name="section_helm_upgrade"></a>

若组件已通过helm安装并纳入helm管理，可直接使用helm upgrade升级到新版本。请先从[MindCluster 发行版](https://gitcode.com/Ascend/mind-cluster/releases)页面下载对应版本的部署工具压缩包Ascend-helm-deploy-tool_{version}_linux.zip并解压，获取tgz安装包。

>[!NOTE]
>
>- 若需要自定义参数配置，可参考[yaml默认配置](../02_installation/helm_installation.md#默认配置)分别创建crds-values.yaml和values.yaml文件，在升级时使用`-f crds-values.yaml`和`-f values.yaml`指定。参数说明请参考[参数说明](../02_installation/helm_installation.md#参数说明)章节。

1. 升级mindcluster crd资源和应用组件。

    ```bash
    helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz # 可增加-f crds-values.yaml指定自定义参数
    helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz # 可增加-f values.yaml指定自定义参数
    ```
2. 确认组件升级状态，请参考[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。

## 版本回退<a name="section_rollback"></a>

若升级后组件运行异常，可通过helm的回退功能恢复到升级前的版本。版本回退仅适用于通过helm upgrade升级的组件，helm会记录每次升级的Revision历史。首次使用helm install安装的组件无历史Revision，无法回退，可通过卸载后重新安装旧版本进行恢复。

1. 执行以下命令，查看Release的升级历史。

    ```bash
    helm history mindcluster
    ```

    回显示例如下：

    ```bash
    REVISION  UPDATED                   STATUS      CHART                                APP VERSION  DESCRIPTION
    1         2026-03-24 15:30:00.000   superseded  mindcluster-deploy-tool-1.0.0        26.0.0       Install complete
    2         2026-03-25 10:00:00.000   deployed    mindcluster-deploy-tool-1.1.0        26.1.0       Upgrade complete
    ```

2. 执行以下命令，回退到指定的Revision版本。以回退到REVISION 1为例：

    ```bash
    helm rollback mindcluster 1
    ```

    若crd资源也需要回退：

    ```bash
    helm rollback mindcluster-crds 1
    ```

3. 确认组件运行状态，请参考[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。
