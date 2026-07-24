# 使用Helm卸载<a name="ZH-CN_TOPIC_0000002511426390"></a>

## 卸载说明<a name="section_uninstall_desc"></a>

本文档介绍如何通过Helm卸载MindCluster组件。

**使用约束**

- 仅支持使用Helm 3.x版本。
- 支持使用Helm卸载的组件包括：
  - Ascend Device Plugin
  - Ascend Operator
  - Volcano
  - ClusterD
  - NodeD
  - NPU Exporter
  - Infer Operator
  - K8s RDMA Shared Dev Plugin
- 卸载Ascend Docker Runtime和Container Manager组件请参照[手动卸载](../../05_developer_guide/00_installation_deployment/02_uninstallation.md#ZH-CN_TOPIC_0000002511426389)章节操作。
- TaskD和MindIO安装在业务容器中，不在本章节涉及的组件范围内。

## 卸载前准备<a name="section_helm_upgrade_prepare"></a>

1. 在管理节点安装Helm命令<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>。若环境中已经存在Helm 3.x版本，可以跳过此步骤。
   - 安装Helm前请参考[Helm版本支持策略](https://v3.helm.sh/zh/docs/v3/topics/version_skew/)查询Helm与K8s间的版本兼容性，根据实际情况选择Helm版本。
   - 请参考[Helm安装文档](https://helm.sh/zh/docs/v3/intro/install)，在管理节点安装Helm命令。

   安装成功后，执行如下命令检查Helm版本：

   ```bash
   helm version
   ```

   回显示例如下：

   ```ColdFusion
   version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean",    GoVersion:"go1.19.5"}
   ```

2. 确认组件是否通过Helm管理<a name="section_check_helm"></a>。
   1. 登录K8s管理节点，执行以下命令，查看当前集群中通过Helm管理的Release列表。

       ```bash
       helm list -A
       ```

       回显示例如下：

       ```ColdFusion
       NAME               NAMESPACE   REVISION  UPDATED                                  STATUS       CHART                                        APP VERSION
       mindcluster        default    1         2026-03-24 15:30:00.000000000 +0800 CST  deployed     mindcluster-deploy-tool-26.1.0                26.1.0
       mindcluster-crds   default    1         2026-03-24 15:25:00.000000000 +0800 CST  deployed     mindcluster-crds-deploy-tool-26.1.0           26.1.0
       ```

   2. 根据回显结果判断组件是否通过Helm管理。
       - 若回显中存在名称为**mindcluster**和**mindcluster-crds**的Release，且STATUS为**deployed**，表示组件已通过Helm管理，可继续执行Helm卸载操作。
       - 若回显中不存在上述Release，表示组件未通过Helm管理，请参照[手动卸载](../../05_developer_guide/00_installation_deployment/02_uninstallation.md#ZH-CN_TOPIC_0000002511426389)进行卸载。

## 执行卸载<a name="section_exec_uninstall"></a>

>[!NOTE]
>
>- 卸载操作需要在K8s管理节点执行。
>- 卸载前请确认集群中无正在使用MindCluster组件管理的工作负载，避免业务中断。

1. （可选）关闭pingmesh灵衢网络检测。pingmesh灵衢网络检测是针对超节点内部（包括节点内和节点间）的HCCS网络提供的NPU网络故障检测，用于监控超节点间网络连通性。卸载前关闭pingmesh可避免卸载后残留的网络检测配置干扰集群网络。
    1. 执行以下命令编辑pingmesh-config ConfigMap。

        ```bash
        kubectl edit cm -n cluster-system pingmesh-config
        ```

    2. 修改activate字段的取值。
        - 如果超节点ID在pingmesh-config ConfigMap中，修改该超节点ID字段下的activate为off。
        - 如果超节点ID不在pingmesh-config ConfigMap中，可通过以下2种方式进行设置。
            - 在pingmesh-config ConfigMap中新增该超节点信息，并将activate设置为off。
            - 删除pingmesh-config ConfigMap中所有超节点的信息，并将global配置中activate字段的值设置为off。

2. 卸载MindCluster应用组件。

    ```bash
    helm uninstall mindcluster
    ```

    回显示例如下，表示卸载成功。

    ```ColdFusion
    release "mindcluster" uninstalled
    ```

3. 卸载MindCluster CRD资源。

    执行以下命令，卸载MindCluster CRD资源。

    ```bash
    helm uninstall mindcluster-crds
    ```

    回显示例如下，表示卸载成功。

    ```ColdFusion
    release "mindcluster-crds" uninstalled
    ```

4. （可选）删除命名空间。若mindx-dl和cluster-system命名空间下已无其他资源，可执行如下命令删除命名空间。删除命名空间会删除该namespace下的所有资源，请确认后再执行。

    ```bash
    kubectl delete ns mindx-dl cluster-system
    ```

    回显示例如下：

    ```ColdFusion
    namespace "mindx-dl" deleted
    namespace "cluster-system" deleted
    ```

5. （可选）删除日志文件。参考[（可选）创建日志目录](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#可选创建日志目录)章节，在对应节点上删除集群调度组件的日志目录。以ClusterD为例，请确认后再删除。

    ```bash
    rm -rf /var/log/mindx-dl/clusterd
    ```

6. 确认卸载结果。

    1. 执行以下命令，确认Release已被删除。

       ```bash
       helm list -A
       ```

       若回显中不存在mindcluster和mindcluster-crds相关的Release，表示卸载成功。

    2. 执行以下命令，确认相关Pod已被删除。

       ```bash
       kubectl get pods -n mindx-dl
       ```

       若回显中提示命名空间不存在或无相关Pod，表示组件已卸载完成。
