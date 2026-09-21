# Uninstallation with Helm<a name="ZH-CN_TOPIC_0000002511426390"></a>

<!-- md-trans-meta sourceCommit=1add3fb0e10984a5f5f80d259f8bb278e17a1a27 translatedAt=2026-08-27T02:28:11.677Z pushedAt=2026-08-27T06:46:31.966Z -->

## Uninstallation Description<a name="section_uninstall_desc"></a>

This document describes how to uninstall MindCluster components using Helm.

**Usage Constraints**

- Only Helm 3.x is supported.
- The components that can be uninstalled using Helm include:
  - Ascend Device Plugin
  - Ascend Operator
  - Volcano
  - ClusterD
  - NodeD
  - NPU Exporter
  - Infer Operator
  - K8s RDMA Shared Dev Plugin
- To uninstall Ascend Docker Runtime and Container Manager, perform the operations described in [Manual Uninstallation](../../05_developer_guide/00_installation_deployment/02_uninstallation.md).
- TaskD and MindIO are installed in service containers and are not covered by the components described in this chapter.

## Preparing for Uninstallation<a name="section_helm_upgrade_prepare"></a>

1. Install the Helm command on the management node<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>. If Helm 3.x already exists in the environment, you can skip this step.
   1. Before installing Helm, refer to [Helm Version Support Policy](https://v3.helm.sh/docs/v3/topics/version_skew/) to check the version compatibility between Helm and K8s, and select a Helm version based on the actual situation.
   2. Refer to [Helm Installation Documentation](https://helm.sh/docs/v3/intro/install) to install the Helm command on the management node.
   3. After the installation is successful, run the following command to check the Helm version:

        ```bash
        helm version
        ```

        Example Output:

        ```ColdFusion
        version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean",    GoVersion:"go1.19.5"}
        ```

2. Check whether the components are managed by Helm<a name="section_check_helm"></a>.
   1. Log in to the K8s management node and run the following command to view the list of `Releases` managed by Helm in the current cluster.

       ```bash
       helm list -A
       ```

       Example Output:

       ```ColdFusion
       NAME               NAMESPACE   REVISION  UPDATED                                  STATUS       CHART                                        APP VERSION
       mindcluster        default    1         2026-03-24 15:30:00.000000000 +0800 CST  deployed     mindcluster-deploy-tool-26.1.0                26.1.0
       mindcluster-crds   default    1         2026-03-24 15:25:00.000000000 +0800 CST  deployed     mindcluster-crds-deploy-tool-26.1.0           26.1.0
       ```

   2. Determine whether the components are managed by Helm based on the command output.
       - If the output contains Releases named **mindcluster** and **mindcluster-crds** with STATUS **deployed**, it indicates that the components are managed by Helm, and you can proceed with the Helm uninstallation.
       - If the output does not contain the preceding Releases, it indicates that the components are not managed by Helm. In this case, uninstall them by referring to [Manual Uninstallation](../../05_developer_guide/00_installation_deployment/02_uninstallation.md).

## Performing Uninstallation<a name="section_exec_uninstall"></a>

>[!NOTE]
>
>- The uninstallation must be performed on the K8s management node.
>- Before uninstallation, ensure that no workloads managed by MindCluster components are running in the cluster to avoid service interruption.

1. (Optional) Disable pingmesh UnifiedBus network detection that provides NPU network fault detection for the HCCS network within a SuperPoD (including intra-node and inter-node) and monitors network connectivity between SuperPoDs. Disabling pingmesh before uninstallation prevents residual network detection configurations from interfering with the cluster network after uninstallation.
    1. Run the following command to edit the `pingmesh-config` ConfigMap.

        ```bash
        kubectl edit cm -n cluster-system pingmesh-config
        ```

    2. Modify the value of the `activate` field.
        - If the SuperPoD ID exists in `pingmesh-config`, set `activate` under the field of this SuperPoD ID to `off`.
        - If the SuperPoD ID does not exist in `pingmesh-config`, configure it in either of the following two ways.
            - Add the information about this SuperPoD to `pingmesh-config` and set `activate` to `off`.
            - Delete the information about all SuperPoDs from `pingmesh-config` and set `activate` in the global configuration to `off`.

2. Uninstall the MindCluster components.

    ```bash
    helm uninstall mindcluster
    ```

    The following output indicates that a successful uninstallation:

    ```ColdFusion
    release "mindcluster" uninstalled
    ```

3. Uninstall the MindCluster CRD resources.

    Run the following command to uninstall the MindCluster CRD resources:

    ```bash
    helm uninstall mindcluster-crds
    ```

    The following output indicates that a successful uninstallation:

    ```ColdFusion
    release "mindcluster-crds" uninstalled
    ```

4. (Optional) Delete the namespaces. If no other resources exist in the `mindx-dl` and `cluster-system` namespaces, run the following command to delete the namespaces. Deleting a namespace deletes all resources in the namespace. Confirm before you perform this operation.

    ```bash
    kubectl delete ns mindx-dl cluster-system
    ```

    Example Output:

    ```ColdFusion
    namespace "mindx-dl" deleted
    namespace "cluster-system" deleted
    ```

5. (Optional) Delete the log files. Refer to [(Optional) Creating a Log Directory](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#optional-creating-log-directories) and delete the log directory of the cluster scheduling component on the corresponding node. Confirm before deletion. Take ClusterD as an example.

    ```bash
    rm -rf /var/log/mindx-dl/clusterd
    ```

6. Verify the uninstallation result.

    1. Run the following command to verify that Release has been deleted.

       ```bash
       helm list -A
       ```

       If no Release related to `mindcluster` and `mindcluster-crds` exists in the output, the uninstallation is successful.

    2. Run the following command to confirm that the related Pods have been deleted.

       ```bash
       kubectl get pods -n mindx-dl
       ```

       If the output indicates that the namespace does not exist or no related Pods exist, the components have been uninstalled.
