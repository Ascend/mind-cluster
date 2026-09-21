# Installing with Helm<a name="ZH-CN_centerIC_0000002479226452"></a>

<!-- md-trans-meta sourceCommit=6b15c7c280d50f1cd26f995458d18bf687b8e5bb translatedAt=2026-08-27T02:29:08.120Z pushedAt=2026-08-27T06:46:31.977Z -->

## Installation Description<a name="ZH-CN_centerIC_0000002511346381_install_desc"></a>

Helm is a tool for managing Kubernetes applications. It helps users quickly deploy, upgrade, and manage Kubernetes applications. The MindCluster Helm can quickly deploy and manage MindCluster components.

**Usage Constraints**

- Only Helm 3.x is supported.
- The components that can be installed using Helm include:
    - [Ascend Device Plugin](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226928)
    - [Ascend Operator](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426817)
    - [Volcano](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386902)
    - [ClusterD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511346859)
    - [NodeD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386924)
    - [NPU Exporter](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226948)
    - [Infer Operator](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426821)
    - [K8s RDMA Shared Dev Plugin](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002524312660)
- To instal [Container Manager](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002524312655), refer to the [Manually Installing Container Manager](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_container_manager.md) chapter.
- [TaskD](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479386914) and [MindIO](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002479226942) are installed in service containers and are not within the scope of the components covered in this chapter.

## Preparations Before Installation<a name="ZH-CN_centerIC_0000002511346381_install_prepare"></a>

1. Install [Ascend Docker Runtime](../../01_introduction/01_component_description.md#ZH-CN_TOPIC_0000002511426843)<a name="zh-cn_centerIC_0000002511346381_install_prepare_docker_runtime"></a>.
   1. If Ascend Docker Runtime has not been installed, refer to the [Manually Installing Ascend Docker Runtime](../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md#ZH-CN_TOPIC_0000002479226434) chapter to install this component on all nodes.
   2. Refer to [Component Status Confirmation](../03_confirming_status.md#ZH-CN_TOPIC_0000002511426307) to confirm the status of Ascend Docker Runtime on all nodes where this component is installed.

2. Install the Helm command on the management node<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>. If Helm 3.x already exists in the environment, you can skip this step.
   1. Before installing Helm, refer to [Helm Version Support Policy](https://v3.helm.sh/docs/v3/topics/version_skew/) to check the version compatibility between Helm and Kubernetes, and select a Helm version based on the actual situation.
   2. Refer to the [Helm installation documentation](https://helm.sh/docs/v3/intro/install) to install the Helm command on the management node.
   3. After the installation is successful, run the following command to check the Helm version:

      ```bash
      helm version
      ```

      The output is as follows:

      ```ColdFusion
      version.BuildInfo{Version:"v3.17.0", GitCommit:"065003584b62a79f329070a946936374936021d6", GitTreeState:"clean",    GoVersion:"go1.19.5"}
      ```

## Performing Installation<a name="ZH-CN_centerIC_0000002511346381_install_exec"></a>

The installation steps are as follows:

1. Refer to [Creating Node Labels](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#ZH-CN_TOPIC_0000002511426279) to label the nodes.

   >[!NOTE]
   >- The default log path does not need to be manually created. It is automatically created by the `initContainer` command in the component YAML file. For the default log path, refer to [Cluster Scheduling Component Log Path List](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#table957112617314).
   >- You do not need to create a new user on the host. You only need to ensure that no other user occupies UID 9000. For user information, refer to [Creating a User](../../05_developer_guide/00_installation_deployment/00_manual_installation/01_preparing_for_installation.md#ZH-CN_TOPIC_0000002511346353).

2. Obtain MindCluster Helm.
   1. Download the deployment tool.

      ```bash
      # Replace {version} in the command with the corresponding version number, for example, 26.1.0.
      wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
      ```

   2. Decompress the package.

      ```bash
      unzip Ascend-helm-deploy-tool_{version}_linux.zip
      ```

   3. View the decompressed files.

      ```bash
      ls -l
      ```

      The output is as follows:

      ```ColdFusion
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
      -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
      -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
      ```

      > [!NOTE]
      >- `{version}` indicates the MindCluster component version, such as 26.1.0.
      >- `{chart_version}` indicates the Helm Chart version, which is consistent with the MindCluster component version.
      >- For the purposes of the extracted files, refer to [Table 4](#table15274931175244).

3. Use Helm to install the `Release` instance (Helm deployment unit) of Custom Resource Definitions (CRDs) required by MindCluster components.
    > [!NOTE]
    >- The following three components contain CRDs: Ascend Operator, Volcano, and Infer Operator. If you do not need to install them, skip this step.
    >- Select either **default configuration installation** or **custom configuration installation** as needed.
    - **Default configuration installation**: If the [CRD default configuration](#default_crds_yaml_install_config) meets your requirements, run the following command to install the CRD resources.

      ```bash
      #(Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
      # Perform the installation.
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
      ```

    - **Custom configuration installation<a name="update_crd_values_before_install_crds"></a>**: If the [CRD default configuration](#default_crds_yaml_install_config) does not meet user requirements, create a crds-values.yaml file, copy the YAML content of the [CRD default configuration](#default_crds_yaml_install_config) into the `crds-values.yaml` file, modify the relevant configuration, and then run the following command:

      ```bash
      #(Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
      # Perform the installation.
      helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
      ```

      The output is as follows, indicating a successful installation:

      ```ColdFusion
      Release "mindcluster-crds" does not exist. Installing it now.
      NAME: mindcluster-crds
      LAST DEPLOYED: ...
      NAMESPACE: default
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```

4. Use Helm to install the `Release` instance of MindCluster components.
    > [!NOTE]
    >- The **default configuration installation method** downloads the images of the components from the Ascend image repository. If your nodes cannot connect to the internet and the images are not cached locally, the installation may fail.
    >- When installing the 26.1.0 version components, note that in the Ascend image repository, the image tags of the 26.1.0 version components are `v26.1.0-ubuntu22.04`, `v26.1.0-alpinelatest`, or `v26.1.0-openeuler24.03`, but in the default configuration of the 26.1.0 version Helm tool, the tag only specifies `v26.1.0`. If the **default configuration installation** is used, an error will be reported during installation because the image cannot be found. Therefore, use the **custom configuration installation** method to configure the correct tag.
    >- Choose either the **default configuration installation** or the **custom configuration installation** method as needed.
    - **Default configuration installation**: If the [component default configuration](#default_app_yaml_install_config) meets requirements, run the following command to install them.

      ```bash
      #(Optional) --dry-run does not actually create any resources. It can be used to verify template syntax and check whether the generated configuration meets expectations.
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
      # Formally execute the installation.
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz
      ```

      > [!NOTE]
      > After running the command, if image pulling fails, refer to [Image Pull Failure When Installing or Upgrading Components with Helm](https://gitcode.com/Ascend/mind-cluster/issues/1013) for handling.

    - **Custom configuration installation<a name="update_app_values_before_install_app"></a>**: If the [component default configuration](#default_app_yaml_install_config) does not meet requirements, create a `values.yaml` file, copy the content of the [component default configuration](#default_app_yaml_install_config) into the `values.yaml` file, and modify the relevant configuration. For example, the configuration example for modifying the image name, log level, and offline hot reset of Ascend Device Plugin is as follows:

        ```yaml
        ...
        ascend-device-plugin:
          enabled: true
          is310P1usoc: false
          volcanoType: true
          image:
            repository: "ascend-k8sdeviceplugin" # Modify the Ascend Device Plugin image name.
            tag: "v26.1.0"
            pullPolicy: "IfNotPresent"
          args: [ "device-plugin -volcanoType=true -presetVirtualDevice=true -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=-1 --enable-healthz=true --healthz-address=11251 -hotReset=2" ] # Change the log level to Debug level and enable the offline hot reset function.
        ...
        ```

Then run the following command:

      ```bash
      # (Optional) --dry-run does not actually create any resources. It can be used to verify the template syntax and check whether the generated configuration meets expectations.
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
      # Perform installation.
      helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
      ```

      The output is as follows, indicating a successful installation:

      ```ColdFusion
      Release "mindcluster" does not exist. Installing it now.
      NAME: mindcluster
      LAST DEPLOYED: ...
      NAMESPACE: default
      STATUS: deployed
      REVISION: 1
      TEST SUITE: None
      ```

1. Refer to [Component Status Confirmation](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390) to confirm the component installation status.
2. If the component status is abnormal, check whether the installation configuration is correct, troubleshoot the cause of the abnormality, and then reinstall.
   - Before reinstalling, run the following commands to uninstall the related resources.

     ```bash
     helm uninstall mindcluster-crds # Uninstall CRDs.
     helm uninstall mindcluster      # Uninstall components.
     ```

## Default Configuration

- CRD default configuration<a name="default_crds_yaml_install_config"></a>
    > [!NOTE]
    >- The components whose CRDs are installed by default include Infer Operator, Volcano, and Ascend Operator. The Volcano version is v1.9.0.
    >- For parameter descriptions, see [Table 1](#table15274931175241).

   ```yaml
   ascend-operator-crds:
     enabled: true              # Install the CRD of Ascend Operator.
   ascend-for-volcano-crds:
     enabled: true              # Install the CRD of the ascend-for-volcano component.
     volcanoVersion: "v1.9.0"   # Version of the Volcano CRD.
   infer-operator-crds:
     enabled: true              # Install the CRD of Infer Operator.
   ```

- Component default configuration<a name="default_app_yaml_install_config"></a>
    > [!NOTE]
    >- Components installed by default include Ascend Device Plugin, Ascend Operator, Volcano, ClusterD, NodeD, NPU Exporter, and Infer Operator. The Volcano version is v1.9.0.
    >- Components not installed by default include K8s RDMA Shared Dev Plugin.
    >- For parameter descriptions, see [Table 2](#table15274931175242) and [Table 3](#table15274931175243). The parameters in [Table 3](#table15274931175243) are not shown in the YAML configuration below. You can add or modify them as needed.
    >- In version 26.1.0, the default value of the `tag` field is `v26.1.0`, without the `-openeuler24.03`, `-ubuntu22.04`, or `-alpinelatest` suffix. This is inconsistent with the image tags in the Ascend image repository. Using the default value directly may cause image pull failures. You can set it to a tag with a suffix, such as `v26.1.0-openeuler24.03`, `v26.1.0-ubuntu22.04`, or `v1.9.0-v26.1.0-alpinelatest`.

   ```yaml
   # The default YAML configuration for installing components is as follows.
   clusterd:
     enabled: true                                                         # Install ClusterD.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd"   # ClusterD image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # ClusterD image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # ClusterD image pull policy. Modify it based on the actual situation.

   noded:
     enabled: true                                                         # Install NodeD.
     enabledStorageCheck: ""                                                # Type of shared storage fault detection to enable; empty means disabled.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/noded"      # NodeD image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # NodeD image tag. Modify it based on the actual situation. Future versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # NodeD image pull policy. Modify it based on the actual situation.

   npu-exporter:
     enabled: true                                                         # Install NPU Exporter.
     is310P1usoc: false                                                    # false indicates that the product is not an Atlas 200I SoC A1 core board.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter" # NPU Exporter image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # NPU Exporter image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # NPU Exporter image pull policy. Modify it based on the actual situation.

   ascend-operator:
     enabled: true                                                         # Install Ascend Operator.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator" # Ascend Operator image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # Ascend Operator image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # Ascend Operator image pull policy. Modify it based on the actual situation.

   ascend-for-volcano:
     enabled: true                                                         # Install Volcano.
     volcanoVersion: "v1.9.0"                                              # Set the Volcano version to be installed.
     scheduler:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler"      # Volcano Scheduler image name. Modify it based on the actual situation.
         # The image tag in the Ascend image repository is "v1.9.0-v26.1.0-openeuler24.03" or "v1.9.0-v26.1.0-alpinelatest"
         tag: "v1.9.0-v26.1.0"                                                      # Volcano Scheduler image tag. Future versions (including patch versions) will add the suffixes "-openeuler24.03" and "-alpinelatest"
         pullPolicy: "IfNotPresent"                                                 # Volcano Scheduler image pull policy
     controller:
       image:
         repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager" # Volcano Controller image name. Modify it based on the actual situation
         # The image tag in the Ascend image repository is "v1.9.0-v26.1.0-openeuler24.03" or "v1.9.0-v26.1.0-alpinelatest"
         tag: "v1.9.0-v26.1.0"                                                          # Volcano Controller image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-alpinelatest".
         pullPolicy: "IfNotPresent"                                                     # Volcano Controller image pull policy.

   infer-operator:
     enabled: true                                                         # Install Infer Operator.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator" # Infer Operator image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # Infer Operator image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # Infer Operator image pull policy. Modify it based on the actual situation.

   ascend-device-plugin:
     enabled: true                                                         # Install Ascend Device Plugin.
     is310P1usoc: false                                                    # false indicates that the product is not the Atlas 200I SoC A1 core board.
     volcanoType: true                                                     # true indicates that Volcano is used for scheduling. Modify it based on the actual situation.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin" # Ascend Device Plugin image name. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                      # Ascend Device Plugin image tag. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                          # Ascend Device Plugin image pull policy. Modify it based on the actual situation.

   k8s-rdma-shared-dev-plugin:
     enabled: false                                                           # false indicates that K8s RDMA Shared Dev Plugin is not installed.
     image:
       repository: "swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dp" # Image name of K8s RDMA Shared Dev Plugin. Modify it based on the actual situation.
       # The image tag in the Ascend image repository is "v26.1.0-openeuler24.03" or "v26.1.0-ubuntu22.04".
       tag: "v26.1.0"                                                              # Image tag of K8s RDMA Shared Dev Plugin. Modify it based on the actual situation. Later versions (including patch versions) will add the suffixes "-openeuler24.03" and "-ubuntu22.04".
       pullPolicy: "IfNotPresent"                                                  # Image pull policy of K8s RDMA Shared Dev Plugin. Modify it based on the actual situation.
   ```

## Parameter Description

**Table 1**  Configurable parameters of CRD resources
<a name="table15274931175241"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>Component to Which CRD Belongs</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>Parameter Name</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>Value Type</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.3"><p>Description</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that CRD of Ascend Operator is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>Volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that CRD of Volcano is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano-crds.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p></p><p>string</p><ul><li>v1.7.0</li><li>v1.9.0</li><li>v1.12.0</li></ul><p>The default value is v1.9.0</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Select the Volcano version.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Infer Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator-crds.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that CRD of Infer Operator is enabled.</p></td>
  </tr>
</tbody>
</table>

**Table 2**  Configurable component parameters
<a name="table15274931175242"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>Component</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>Parameter Name</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>Value Type</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.3"><p>Description</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>ClusterD</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>clusterd.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that ClusterD is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>NodeD</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that NodeD is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>noded.enabledStorageCheck</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>default value is empty</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Enabled shared storage fault detection type. The value range includes "", "dpc", "dtfs", "dpc,dtfs", and "container-snapshot".</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>NPU Exporter</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>default value is true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that NPU Exporter is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>npu-exporter.is310P1usoc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is false.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that the product is the Atlas 200I SoC A1 core board.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>Default value is true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that Ascend Operator is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="2" valign="center" headers="mcps1.2.5.1.1 "><p>Volcano</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that Volcano is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-for-volcano.volcanoVersion</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>Values include:<ul><li>v1.7.0</li><li>v1.9.0</li><li>v1.12.0</li></ul></p><p>The default value is v1.9.0.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Select the volcano version to enable.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>Infer Operator</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>infer-operator.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that Infer Operator is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" rowspan="3" valign="center" headers="mcps1.2.5.1.1 "><p>Ascend Device Plugin</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that Ascend Device Plugin is enabled.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.is310P1usoc</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is false.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicates that the product is an Atlas 200I SoC A1 core board.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>ascend-device-plugin.volcanoType</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>The default value is true.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that Volcano is used for scheduling.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>K8s RDMA Shared Dev Plugin</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>k8s-rdma-shared-dev-plugin.enabled</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>bool</p><p>default value is false</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Set to true indicating that K8s RDMA Shared Dev Plugin is enabled.</p></td>
  </tr>
</tbody>
</table>

**Table 3**  Other configurable component parameters
<a name="table15274931175243"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.51.1"><p>Parameter Name</p></th>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.5.1.2"><p>Value Range</p></th>
    <th class="cellrowborder" valign="center" width="50%" id="mcps1.2.5.1.3"><p>Description</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.args</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string array</p><p>If not set or set to "", then the component default startup command is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the startup command of the component.</p><p>If you need to modify the startup command to enable certain component features (for example, enabling hot reset for ascend-device-plugin), set this parameter as needed.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.repository</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><ul><li>If set to "", then the component default image name is used.</li><li>If not set, the default configuration is the Ascend image repository address.</li></ul></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the image repository address or image name of the component.</p><p>If the Ascend image repository address is used, ensure that nodes can access the internet normally; otherwise, the component status may be abnormal after deployment due to missing images.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.tag</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>If not set or set to "", then the component default image tag is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the image version tag of the component.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.image.pullPolicy</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>The value can be:<ul><li>IfNotPresent</li><li>Always</li><li>Never</li></ul></p><p>If set to "", the component default image pull policy is used.</p><p>If not set, the default is IfNotPresent.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the image pull policy of the component.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>If not set or set to "", then the component default memory request size is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the component's requested memory size, such as "512Mi".</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.requests.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>If not set or set to "", then the component default CPU request size is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the component's requested CPU size, such as "500m".</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.memory</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>If not set or set to "", then the component default memory limit is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the component memory limit, such as "1Gi".</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>&lt;component&gt;.resources.limits.cpu</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>string</p><p>If not set or set to "", then the component default CPU limit is used.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.3 "><p>Indicates the CPU limit of the component, such as "1000m".</p></td>
  </tr>
</tbody>
</table>

>[!NOTE]
> In Table 3, the value of `<component>` can be `clusterd`, `noded`, `npu-exporter`, `ascend-operator`, `ascend-for-volcano.scheduler`, `ascend-for-volcano.controller`, `infer-operator`, `ascend-device-plugin`, `k8s-rdma-shared-dev-plugin`. Taking `clusterd` as an example, `<component>.image.repository` is replaced with `clusterd.image.repository`.

**Table 4**  Files in the Helm package
<a name="table15274931175244"></a>
<table>
<thead align="left">
  <tr>
    <th class="cellrowborder" valign="center" width="20%" id="mcps1.2.51.1"><p>File Name</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>Usage</p></th>
    <th class="cellrowborder" valign="center" width="30%" id="mcps1.2.5.1.2"><p>Description</p></th>
  </tr>
</thead>
<tbody>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>mindcluster-crds-deploy-tool-{chart_version}.tgz</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>A Helm Chart package file, which is a deployment tool used to deploy and manage the CRDs required by each MindCluster component in a K8s cluster.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>For the description of user-configurable installation parameters, see <a href="#table15274931175241">Table 1</a>.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>mindcluster-deploy-tool-{chart_version}.tgz</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>Helm Chart package file, a deployment tool used to deploy and manage the MindCluster components in a K8s cluster.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>For the user-configurable installation parameter description, see <a href="#table15274931175242">Table 2</a> and <a href="#table15274931175243">Table 3</a>.</p></td>
  </tr>
  <tr>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.1 "><p>helm_tool.sh</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 "><p>Its functions include: <ul><li><p>A script that adds Helm metadata to the resources of each component.</p></li><li>Deletes the DaemonSet resources of Ascend Device Plugin before version 26.1.0</li></ul></p><p>Used only during upgrade.</p></td>
    <td class="cellrowborder" valign="center" headers="mcps1.2.5.1.2 ">The script adds Helm metadata to the following resources, including: <ul><li>Ascend Operator-related resources</li><li>Ascend Device Plugin-related resources</li><li>Volcano-related resources</li><li>ClusterD-related resources</li><li>NodeD -related resources</li><li>NPU Exporter-related resources</li><li>Infer Operator-related resources</li><li>K8s RDMA Shared Dev Plugin-related resources</li><li>Namespaces, including "mindx-dl" and "cluster-system"</li></ul></td>
  </tr>
</tbody>
</table>

 > [!NOTE]
 > `{chart_version}` indicates the Helm Chart version, which is consistent with the MindCluster component version.
