# Upgrade with Helm<a name="ZH-CN_TOPIC_0000002479226453"></a>

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-08-27T02:28:39.377Z pushedAt=2026-08-27T06:46:31.974Z -->

## Upgrade Description<a name="section_helm_upgrade_desc"></a>

This document describes how to upgrade MindCluster components using Helm.

**Version compatibility**: Cross-major-version upgrades are supported, for example, from 7.x.x to 26.x.x.

**Usage Constraints**

- Only Helm 3.x is supported.
- Components that support upgrade using Helm include:
    - [Ascend Device Plugin](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226928)
    - [Ascend Operator](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426817)
    - [Volcano](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386902)
    - [ClusterD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511346859)
    - [NodeD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386924)
    - [NPU Exporter](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226948)
    - [Infer Operator](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426821)
    - [K8s RDMA Shared Dev Plugin](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002524312660)
- To upgrade [Ascend Docker Runtime](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426843), [Container Manager](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002524312655), [TaskD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386914), and [MindIO](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226942), refer to the [Manual Upgrade](../../05_developer_guide/00_installation_deployment/01_upgrade.md#ZH-CN_TOPIC_0000002479226452) section.

## Pre-upgrade Preparation<a name="section_helm_upgrade_prepare"></a>

1. Install the Helm command on the management node<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>. If Helm 3.x already exists in the environment, skip this step.
   1. Before installing Helm, refer to [Helm Version Support Policy](https://v3.helm.sh/docs/v3/topics/version_skew/) to check the version compatibility between Helm and K8s, and select a Helm version based on the actual situation.
   2. Refer to the [Helm Installation Documentation](https://helm.sh/docs/v3/intro/install) to install the Helm command on the management node.
   3. After successful installation, run the following command to check the Helm version:

        ```bash
        helm version
        ```

        Output example:

        ```ColdFusion
        version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean",    GoVersion:"go1.19.5"}
        ```

2. Confirm whether components are managed by Helm<a name="section_check_helm_upgrade"></a>. Before performing the upgrade, confirm whether the components to be upgraded is already managed by Helm so that the appropriate upgrade method can be selected.
   1. Log in to the K8s management node and run the following command to view the list of `Releases` managed by Helm in the current cluster.

       ```bash
       helm list -A
       ```

       Output example:

       ```ColdFusion
       NAME               NAMESPACE   REVISION  UPDATED                                  STATUS       CHART                                        APP VERSION
       mindcluster        default    1         2026-03-24 15:30:00.000000000 +0800 CST  deployed  mindcluster-deploy-tool-26.1.0                26.1.0
       mindcluster-crds   default    1         2026-03-24 15:25:00.000000000 +0800 CST  deployed     mindcluster-crds-deploy-tool-26.1.0           26.1.0
       ```

   2. Determine the component upgrade method based on the command output.
       - If the output contains `Releases` named **mindcluster** and **mindcluster-crds** with STATUS **deployed**, the components are already managed by Helm. See [Upgrading a Release Instance with Helm](#section_helm_upgrade) to upgrade them.
       - If the output does not contain the preceding `Release` information, the components are not managed by Helm. See [Helm Deploy Release Instance](#section_kubectl_to_helm) to upgrade them.

       >[!NOTE]
       >If the components are reinstalled using `kubectl delete` and `kubectl apply` after being installed through Helm, the recreated resources will lack Helm metadata, but the preceding `Release` information can still be queried. In this scenario, perform [Step 1](#li1471945063444_helm_download) and [Step 2](#li1471945063445_add_meta) to add Helm metadata back to the resources, and then see [Upgrading a Release Instance with Helm](#section_helm_upgrade) to upgrade them.

## Deploying a Release Instance with Helm<a name="section_kubectl_to_helm"></a>

If the components are installed manually through kubectl and have not yet been brought under Helm management, follow the steps below to upgrade the components to a new version.

1. <a name="li1471945063444_helm_download"></a>Download and decompress the deployment tool.

    ```bash
    # Replace {version} in the command with the corresponding version number, such as 26.1.0.
    wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
    unzip Ascend-helm-deploy-tool_{version}_linux.zip
    ```

    For the purpose of each decompressed file, refer to [Table 4](../02_installation/00_helm_installation.md#table15274931175244). The output example is as follows:

    ```ColdFusion
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
    -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
    ```

    Here, `{version}` indicates the MindCluster component version, `{chart_version}` indicates the Helm Chart version.

2. <a name="li1471945063445_add_meta"></a>Run the following command to add Helm metadata to existing resources.

    ```bash
    sed -i 's/\r$//' helm_tool.sh && chmod +x helm_tool.sh

    #(Optional)
    bash helm_tool.sh --help # View the script command parameters.

    bash helm_tool.sh --add-helm-meta-all # Add Helm metadata to resources.

    # (Optional) If an Ascend Device Plugin DaemonSet of a version earlier than v26.1.0 exists in the current cluster, run this command to clean up the old resources. If it does not exist, skip this step.
    bash helm_tool.sh --delete-old-demonset
    ```

    The following is an output example, indicating that the Helm metadata has been added successfully:

    ```ColdFusion
    ...
    ============ Done ==============
    ```

3. Install the Release instance of MindCluster CRD.
    > [!NOTE]
    >- The following three components contain CRDs: Ascend Operator, Volcano, and Infer Operator. If you do not need to upgrade these three components, skip this step.
    >- If CRDs differ between the two versions before and after component upgrade:
    >   1. Upgrade CRD first, and then upgrade the component.
    >   2. This may interrupt workloads. Confirm this before the upgrade.
    >- Select either **default configuration installation** or **custom configuration installation** as needed.
   - **Default configuration installation**: If the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) meets requirements, run the following commands.

       ```bash
       #(Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
       # Perform the installation.
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
       ```

   - **Custom configuration installation**: If the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) does not meet requirements, create a `crds-values.yaml` file, copy the content of the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) into the `crds-values.yaml` file, modify the relevant configuration, and then run the following commands.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify template syntax and check whether the generated configuration meets expectations.
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
       # Perform the installation.
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
       ```

       The following output example indicates a successful installation:

       ```ColdFusion
       Release "mindcluster-crds" does not exist. Installing it now.
       NAME: mindcluster-crds
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 1
       TEST SUITE: None
       ```

4. Install the `Release` instance of MindCluster components.
    > [!NOTE]
    >- The **default configuration installation** method downloads the images of components from the Ascend image repository. If your nodes cannot connect to the internet and the images are not cached locally, the upgrade may fail.
    >- When upgrading components to version 26.1.0, note: In the Ascend image repository, the image tags for version 26.1.0 components are `v26.1.0-ubuntu22.04`, `v26.1.0-alpinelatest`, or `v26.1.0-openeuler24.03`. However, in the default configuration of the 26.1.0 Helm tool, the tag is specified only as `v26.1.0`. If the **default configuration installation** method is used, an error occurs during installation because the image cannot be found. Therefore, use the **custom configuration installation** method to configure the correct tag.
    >- Choose either the **default configuration installation** or the **custom configuration installation** method as needed.
   - **Default configuration installation**: If the [default configuration of components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) meets requirements, run the following command.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
       # Formally perform the installation.
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz
       ```

       > [!NOTE]
       > If an image failure occurs after running the command, see [Image Pull Failure During Component Installation or Upgrade with Helm](https://gitcode.com/Ascend/mind-cluster/issues/1013) for troubleshooting.

   - **Custom configuration installation**: If the [default configuration of components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) does not meet your requirements, create a `values.yaml` file, copy the YAML content of the [default configuration of application components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) into the `values.yaml` file, modify the relevant configurations, and then run the following command.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
       # Perform the installation.
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
       ```

       The output example is as follows, indicating a successful installation:

       ```ColdFusion
       Release "mindcluster" does not exist. Installing it now.
       NAME: mindcluster
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 1
       TEST SUITE: None
       ```

5. Confirm the component upgrade status. For details, see [Component Status Confirmation](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390).
6. If the component status is abnormal after the upgrade, troubleshoot the cause and handle it as follows:
    - After modifying the configuration, refer to [Upgrading a Release Instance with Helm](#section_helm_upgrade) to upgrade again.
    - After [uninstalling the component using Helm](../05_uninstallation/02_helm_uninstallation.md#ZH-CN_TOPIC_0000002511426390), [install the component using Helm](../02_installation/00_helm_installation.md#ZH-CN_centerIC_0000002479226452) again. This method may interrupt workloads. Confirm this before the upgrade.

## Upgrading a Release Instance with Helm<a name="section_helm_upgrade"></a>

If a component has been installed through Helm and is managed by Helm, follow the steps below to upgrade it to a new version.

1. Download and decompress the deployment tool.

    ```bash
    # Replace {version} in the command with the corresponding version number, for example, 26.1.0.
    wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
    unzip Ascend-helm-deploy-tool_{version}_linux.zip
    ```

    For the purpose of each decompressed file, see [Table 4](../02_installation/00_helm_installation.md#table15274931175244). The following shows an output example of the file list:

    ```ColdFusion
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
    -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
    ```

2. Upgrade the `Release` instance of MindCluster CRDs.
    >[!NOTE]
    >- The following three components contain CRDs: Ascend Operator, Volcano, and Infer Operator. If you do not need to upgrade these three components, skip this step.
    >- If CRDs differ between the two versions before and after component upgrade:
    >   - Upgrade CRDs first, and then upgrade components.
    >   - This may cause workload interruption. Confirm this before the upgrade.
    >- Select either **default configuration upgrade** or **custom configuration upgrade** as needed.
   - **Default configuration upgrade**: If the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) meets your requirements, run the following command.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify template syntax and check whether the generated configuration meets expectations.
       helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
       # Formally perform the upgrade.
       helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
       ```

   - **Custom configuration upgrade**: If the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) does not meet your requirements, create a `crds-values.yaml` file, copy the YAML content of the [CRD default configuration](../02_installation/00_helm_installation.md#default_crds_yaml_install_config) into the `crds-values.yaml` file, modify the relevant configuration, and then run the following command.
      >[!NOTICE]
      >If only a single component is upgraded, keep the configurations of other installed components in `crds-values.yaml` consistent with those during installation. Do not set the Enabled parameter of other installed components to false; otherwise, the resources of the corresponding components will be deleted.

      ```bash
      # (Optional) --dry-run does not actually create any resources. It can be used to verify template syntax and check whether the generated configuration meets expectations.
      helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
      # Formally perform the upgrade.
      helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
      ```

       The output example is as follows:

       ```ColdFusion
       Release "mindcluster-crds" has been upgraded. Happy Helming!
       NAME: mindcluster-crds
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 2
       TEST SUITE: None
       ```

3. Upgrade the `Release` instance of the MindCluster components.
   >[!NOTE]
   >- The **default configuration upgrade method** downloads the component images from the Ascend image repository. If your node cannot connect to the internet and the images are not cached locally, the upgrade may fail.
   >- When upgrading components of version 26.1.0, note: In the Ascend image repository, the image tags of the 26.1.0 components are `v26.1.0-ubuntu22.04`, `v26.1.0-alpinelatest`, or `v26.1.0-openeuler24.03`. However, in the default configuration of the 26.1.0 Helm tool, the tag is specified only as `v26.1.0`. If **default configuration upgrade** is used, an error is reported during the upgrade because the image cannot be found. Therefore, use **custom configuration upgrade** to configure the correct tag.
   >- Select either **default configuration upgrade** or **custom configuration upgrade** as required.
   - **Default configuration upgrade**: If the [default configuration of the components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) meets user requirements, run the following command.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify template syntax and check whether the generated configuration meets expectations.
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
       # Formally perform the upgrade.
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz
       ```

       > [!NOTE]
       > After the command is executed, if image pull fails, see [Image Pull Failure During Helm Tool Installation or Component Upgrade](https://gitcode.com/Ascend/mind-cluster/issues/1013) for troubleshooting.

   - **Custom configuration upgrade**: If the [default configuration of components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) does not meet user requirements, create a `values.yaml` file, copy the YAML content of the [default configuration of components](../02_installation/00_helm_installation.md#default_app_yaml_install_config) into the `values.yaml` file, modify the relevant configuration, and then run the following command.
       >[!NOTICE]
       >If only a single component is upgraded, the configurations of other installed components in `values.yaml` must remain consistent with those during installation. Do not set `Enabled` of other installed components to `false`, otherwise the resources of the corresponding components will be deleted.

       ```bash
       # (Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
       # Perform the upgrade.
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
       ```

       The output example is as follows:

       ```ColdFusion
       Release "mindcluster" has been upgraded. Happy Helming!
       NAME: mindcluster
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 2
       TEST SUITE: None
       ```

4. Confirm the component upgrade status. For details, see [Component Status Confirmation](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390).
5. If the component status is abnormal after the upgrade, troubleshoot the cause and handle it as follows:
   - Roll back to the version before the upgrade by referring to [Version Rollback](#section_helm_rollback).
   - Modify the configuration and use [Helm for upgrading a Release instance](#section_helm_upgrade) again.
   - After [uninstalling the component with Helm](../05_uninstallation/02_helm_uninstallation.md#ZH-CN_TOPIC_0000002511426390), [install the component with Helm](../02_installation/00_helm_installation.md#ZH-CN_centerIC_0000002479226452) again. This method may interrupt workloads. Confirm this before the upgrade.

## Version Rollback<a name="section_helm_rollback"></a>

If components run abnormally after an upgrade, use the Helm rollback feature to restore the version before the upgrade. Version rollback applies only to `Release` instances that have been upgraded through Helm upgrade. Helm records the `Revision` history of each upgrade.

1. Run the following command to view the upgrade history of the `Release` instance.
    - View the upgrade history of the component `Release` instance:

      ```bash
      helm history mindcluster
      ```

      The output example is as follows:

      ```ColdFusion
      REVISION  UPDATED                   STATUS      CHART                                APP VERSION  DESCRIPTION
      1         2026-03-24 15:30:00.000   superseded  mindcluster-deploy-tool-26.0.0        26.0.0       Install complete
      2         2026-03-25 10:00:00.000   deployed    mindcluster-deploy-tool-26.1.0        26.1.0       Upgrade complete
      ```

    - View the upgrade history of the `Release` instance for component CRDs:

       ```bash
       helm history mindcluster-crds
       ```

       The output example is as follows:

      ```ColdFusion
      REVISION  UPDATED                   STATUS      CHART                                APP VERSION  DESCRIPTION
      1         2026-03-24 15:30:00.000   superseded  mindcluster-crds-deploy-tool-26.0.0        26.0.0       Install complete
      2         2026-03-25 10:00:00.000   deployed    mindcluster-crds-deploy-tool-26.1.0        26.1.0       Upgrade complete
      ```

    >[!NOTE]
    >Helm's `REVISION` is a simple incrementing integer whose core purpose is to record and roll back: each change generates a new `REVISION`, and when needed, you can quickly restore to any previous stable version by the `REVISION` number, thereby achieving traceability of component releases and rapid fault recovery.

2. Run the following command to roll back the CRDs to the specified `Revision` version. The following uses REVISION 1 as an example:

    ```bash
    helm rollback mindcluster-crds 1
    ```

    The rollback example is as follows:

    ```ColdFusion
    Rollback was a success! Happy Helming!
    ```

3. Run the following command to roll back components to the specified `Revision`. The following uses REVISION 1 as an example:

    ```bash
    helm rollback mindcluster 1
    ```

    The rollback example is as follows:

    ```ColdFusion
    Rollback was a success! Happy Helming!
    ```

4. Confirm the component status. For details, see [Component Status Confirmation](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390).
