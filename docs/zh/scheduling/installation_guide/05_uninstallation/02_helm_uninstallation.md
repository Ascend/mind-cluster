# 使用helm卸载<a name="ZH-CN_TOPIC_0000002511426390"></a>

## 卸载说明<a name="section_uninstall_desc"></a>

本文档介绍如何通过helm卸载mindcluster组件。仅支持卸载通过helm安装或升级后纳入helm管理的mindcluster组件，包括：

- ascend-device-plugin
- ascend-operator
- ascend-for-volcano
- clusterd
- noded
- npu-exporter
- infer-operator

>[!NOTE]
>
>- 若组件未通过helm安装或升级后纳入helm管理，请参照[手动卸载](01_manual_uninstallation.md)进行卸载，否则可能导致卸载不完整或集群状态异常。
>- docker-runtime、container-manager等组件不支持通过helm管理，请参照[手动卸载](01_manual_uninstallation.md)对应组件章节进行卸载。
>- 仅支持使用helm 3.x版本进行卸载。

## 确认组件是否通过helm管理<a name="section_check_helm"></a>

在执行helm卸载前，请先确认待卸载的组件是否已通过helm管理。若未通过helm管理，请参照[手动卸载](01_manual_uninstallation.md)进行操作。

1. 以root用户登录K8s管理节点。

2. 执行以下命令，查看当前集群中通过helm管理的Release列表。

    ```bash
    helm list -A
    ```

    回显示例如下：

    ```bash
    NAME               NAMESPACE   REVISION  UPDATED                                  STATUS    CHART                                        APP VERSION
    mindcluster        default    1         2026-03-24 15:30:00.000000000 +0800 CST  deployed  mindcluster-deploy-tool-1.1.0                26.1.0
    mindcluster-crds   default    1         2026-03-24 15:25:00.000000000 +0800 CST  deployed  mindcluster-crds-deploy-tool-1.1.0           26.1.0
    ```

3. 根据回显结果判断组件是否通过helm管理。

    - 若回显中存在名称为**mindcluster**和**mindcluster-crds**的Release，且STATUS为**deployed**，表示组件已通过helm管理，可继续执行helm卸载操作。
    - 若回显中不存在上述Release，表示组件未通过helm管理，请参照[手动卸载](01_manual_uninstallation.md)进行卸载。

4. （可选）若需进一步确认某个Release管理的资源详情，可执行以下命令查看。

    ```bash
    helm status mindcluster
    ```

    回显示例如下：

    ```bash
    NAME: mindcluster
    LAST DEPLOYED: ...
    NAMESPACE: default
    STATUS: deployed
    REVISION: 1
    TEST SUITE: None
    ```

## 执行卸载<a name="section_exec_uninstall"></a>

>[!NOTE]
>
>- 卸载操作需要在K8s管理节点执行。
>- 卸载前请确认集群中无正在使用mindcluster组件管理的工作负载，避免业务中断。
>- 卸载顺序：先卸载应用组件（mindcluster），再卸载crd资源（mindcluster-crds），顺序不可颠倒。

1. （可选）关闭pingmesh灵衢网络检测。
    1. 登录环境，进入NodeD解压目录。
    2. 执行以下命令编辑pingmesh-config文件。

        ```bash
        kubectl edit cm -n cluster-system pingmesh-config
        ```

    3. 修改activate字段的取值。
        - 如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        - 如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            - 在配置文件中新增该超节点信息，并将activate为off。
            - 删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

2. 卸载mindcluster应用组件。

    执行以下命令，卸载mindcluster应用组件。

    ```bash
    helm uninstall mindcluster
    ```

    回显示例如下，表示卸载成功。

    ```bash
    release "mindcluster" uninstalled
    ```

3. 卸载mindcluster crd资源。

    执行以下命令，卸载mindcluster crd资源。

    ```bash
    helm uninstall mindcluster-crds
    ```

    回显示例如下，表示卸载成功。

    ```bash
    release "mindcluster-crds" uninstalled
    ```

4. 删除命名空间。若mindx-dl命名空间下已无其他资源，可执行如下命令删除命名空间。删除命名空间会删除该namespace下的所有资源，请确认后再执行。

    ```bash
    kubectl delete ns mindx-dl
    ```

    回显示例如下：

    ```bash
    namespace "mindx-dl" deleted
    ```

5. 删除日志文件。参考[创建日志目录](../../developer_guide/installation_deployment/manual_installation/01_preparing_for_installation.md#创建日志目录)章节，在对应节点上删除集群调度组件的日志目录。以ClusterD为例，请确认后再删除。

    ```bash
    rm -rf /var/log/mindx-dl/clusterd
    ```

6. 确认卸载结果。

    执行以下命令，确认Release已被删除。

    ```bash
    helm list -A
    ```

    若回显中不存在mindcluster和mindcluster-crds相关的Release，表示卸载成功。

    执行以下命令，确认相关Pod已被删除。

    ```bash
    kubectl get pods -n mindx-dl
    ```

    若回显提示命名空间不存在或无相关Pod，表示组件已卸载完成。
