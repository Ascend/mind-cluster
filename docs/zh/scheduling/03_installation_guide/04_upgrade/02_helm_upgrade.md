# 使用helm升级<a name="ZH-CN_TOPIC_0000002479226453"></a>

## 升级说明<a name="section_helm_upgrade_desc"></a>

本文档介绍如何通过helm升级MindCluster组件。

**版本兼容性说明**：可跨大版本升级，如从 7.x.x 升级到 26.x.x。

**使用约束**

- 仅支持使用helm 3.x版本。
- 支持使用helm升级的组件包括：
    - Ascend Device Plugin
    - Ascend Operator
    - Volcano
    - ClusterD
    - NodeD
    - NPU Exporter
    - Infer Operator
    - K8s RDMA Shared Dev Plugin
- 升级Ascend Docker Runtime、Container Manager、TaskD和MindIO组件请参考[手动升级](../../05_developer_guide/00_installation_deployment/01_upgrade.md#ZH-CN_TOPIC_0000002479226452)章节操作。

## 升级前准备<a name="section_helm_upgrade_prepare"></a>

1. 在管理节点安装helm命令<a name="zh-cn_centerIC_0000002511346381_install_prepare_helm"></a>。若环境中已经存在helm 3.x版本，可以跳过此步骤。
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

2. 确认组件是否通过helm管理<a name="section_check_helm_upgrade"></a>。在执行升级前，需先确认待升级的组件是否已通过helm管理，以选择对应的升级方式。
   1. 登录K8s管理节点，执行以下命令，查看当前集群中通过helm管理的Release列表。

       ```bash
       helm list -A
       ```

       回显示例如下：

       ```ColdFusion
       NAME               NAMESPACE   REVISION  UPDATED                                  STATUS       CHART                                        APP VERSION
       mindcluster        default    1         2026-03-24 15:30:00.000000000 +0800 CST  deployed  mindcluster-deploy-tool-26.1.0                26.1.0
       mindcluster-crds   default    1         2026-03-24 15:25:00.000000000 +0800 CST  deployed     mindcluster-crds-deploy-tool-26.1.0           26.1.0
       ```

   2. 根据回显结果判断组件的升级方式。
       - 若回显中存在名称为**mindcluster**和**mindcluster-crds**的Release，且STATUS为**deployed**，表示组件已通过helm管理，请参见[helm upgrade升级组件](#section_helm_upgrade)进行升级。
       - 若回显中不存在上述Release信息，表示组件未通过helm管理，请参见[helm install升级组件](#section_kubectl_to_helm)进行升级。

## helm install升级组件<a name="section_kubectl_to_helm"></a>

若组件是通过kubectl手动安装的，尚未纳入helm管理，需要先为组件资源添加helm元数据，再使用helm install安装Release实例，从而将组件升级到新版本。

1. 下载并解压部署工具：

    ```bash
    # 请用户自行将命令中的{version}替换为对应版本号，如26.1.0
    wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
    unzip Ascend-helm-deploy-tool_{version}_linux.zip
    ```

    解压后的各文件用途请参考[表4](../02_installation/00_helm_installation.md#table15274931175244)，文件列表回显示例如下：

    ```ColdFusion
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
    -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
    ```

2. 执行以下命令，为已有资源添加helm元数据。

    ```bash
    sed -i 's/\r$//' helm_tool.sh && chmod +x helm_tool.sh
    #（可选）查看脚本命令参数
    bash helm_tool.sh --help
    bash helm_tool.sh --all # 给资源添加helm元数据，并且删除ascend-device-plugin组件v26.1.0版本前的daemonset。用户可使用--help查看脚本命令参数，然后根据需求设置参数。
    ```

    回显示例如下，表示添加helm元数据成功：

    ```ColdFusion
    ...
    ============ Done ==============
    ```

3. 安装MindCluster crd的Release实例。
      > [!NOTE]
    >- 以下三个组件包含crd：Ascend Operator、Volcano和Infer Operator。若用户不需要升级这三个组件，可跳过此步骤。
    >- 若组件升级前后两个版本的crd定义有变更：
    >   1. 需先升级crd，再升级应用组件;
    >   2. 可能会导致工作负载中断，请用户在升级前确认。
    >- 请用户按需选择**默认配置安装**或**自定义配置安装**其中一种方式进行操作即可。
   - **默认配置安装**：若[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)符合用户需求，可执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
       # 正式执行安装
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
       ```

   - **自定义配置安装**：若[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)不符合用户需求，请创建crds-values.yaml文件，将[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)的YAML文件内容复制到crds-values.yaml文件中，修改相关配置后执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
       # 正式执行安装
       helm install mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
       ```

       回显示例如下，表示安装成功。

       ```ColdFusion
       Release "mindcluster-crds" does not exist. Installing it now.
       NAME: mindcluster-crds
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 1
       TEST SUITE: None
       ```

4. 安装MindCluster应用组件的Release实例。
    > [!NOTE]
    >- **默认配置安装**方式会从昇腾镜像仓库下载应用组件的镜像。若用户节点无法连接互联网且本地未缓存镜像，可能会升级失败。
    >- 请用户按需选择**默认配置安装**或**自定义配置安装**其中一种方式进行操作即可。
   - **默认配置安装**：若[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)符合用户需求，可执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
       # 正式执行安装
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz
       ```

   - **自定义配置安装**：若[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)不符合用户需求，请创建values.yaml文件，将[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)的YAML文件内容复制到values.yaml文件中，修改相关配置后执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
       # 正式执行安装
       helm install mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
       ```

       回显示例如下，表示安装成功。

       ```ColdFusion
       Release "mindcluster" does not exist. Installing it now.
       NAME: mindcluster
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 1
       TEST SUITE: None
       ```

5. 确认组件升级状态，详细请参见[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。
6. 若升级后，组件状态异常，可排查异常原因，然后按照如下方法处理：
    - 修改配置后参考[helm upgrade升级组件](#section_helm_upgrade)重新升级。
    - [使用helm卸载](../05_uninstallation/02_helm_uninstallation.md#ZH-CN_TOPIC_0000002511426390)组件后，重新[使用helm安装](../02_installation/00_helm_installation.md#ZH-CN_centerIC_0000002479226452)组件。此方法可能会导致工作负载中断，请用户在升级前确认。

## helm upgrade升级组件<a name="section_helm_upgrade"></a>

若组件已通过helm安装并纳入helm管理，可直接使用helm upgrade升级到新版本。

1. 下载并解压部署工具。

    ```bash
    # 请用户自行将命令中的{version}替换为对应版本号，如26.1.0
    wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-helm-deploy-tool_{version}_linux.zip
    unzip Ascend-helm-deploy-tool_{version}_linux.zip
    ```

    解压后的各文件用途请参考[表4](../02_installation/00_helm_installation.md#table15274931175244)，文件列表回显示例如下：

    ```ColdFusion
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-crds-deploy-tool-{chart_version}.tgz
    -r-------- 1 root root  2026 Mar 24 15:25 mindcluster-deploy-tool-{chart_version}.tgz
    -rw-r--r-- 1 root root  2026 Mar 24 15:25 helm_tool.sh
    ```

2. 升级MindCluster crd的Release实例。
    >[!NOTE]
    >- 以下三个组件包含crd：Ascend Operator、Volcano和Infer Operator。若用户不需要升级这三个组件，可跳过此步骤。
    >- 若组件升级前后两个版本的crd定义有变更：
    >   - 需先升级crd，再升级应用组件。
    >   - 可能会导致工作负载中断，请用户在升级前确认。
    >- 请用户按需选择**默认配置升级**或**自定义配置升级**其中一种方式进行操作即可。
   - **默认配置升级**：若[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)符合用户需求，可执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz --dry-run
       # 正式执行升级
       helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz
       ```

   - **自定义配置升级**：若[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)不符合用户需求，请创建crds-values.yaml文件，将[crd默认配置](../02_installation/00_helm_installation.md#default_crds_yaml_install_config)的YAML文件内容复制到crds-values.yaml文件中，修改相关配置后执行如下命令。
      >[!WARNING]
      >若只升级单个组件，crds-values.yaml中其他已安装组件的配置请保持与安装时的配置一致，不能将其他已安装组件的Enabled参数设置为false，否则对应组件的资源会被删除。

      ```bash
      #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
      helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml --dry-run
      # 正式执行升级
      helm upgrade mindcluster-crds mindcluster-crds-deploy-tool-{chart_version}.tgz -f crds-values.yaml
      ```

       回显示例如下：

       ```ColdFusion
       Release "mindcluster-crds" has been upgraded. Happy Helming!
       NAME: mindcluster-crds
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 2
       TEST SUITE: None
       ```

3. 升级MindCluster应用组件的Release实例。
   >[!NOTE]
   >- **默认配置升级**会从昇腾镜像仓库下载应用组件的镜像。若用户节点无法连接互联网且本地未缓存镜像，可能会升级失败。
   >- 请用户按需选择**默认配置升级**或**自定义配置升级**其中一种方式进行操作即可。
   - **默认配置升级**：若[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)符合用户需求，可执行如下命令。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz --dry-run
       # 正式执行升级
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz
       ```

   - **自定义配置升级**：若[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)不符合用户需求，请创建values.yaml文件，将[应用组件默认配置](../02_installation/00_helm_installation.md#default_app_yaml_install_config)的YAML文件内容复制到values.yaml文件中，修改相关配置后执行如下命令。
       >[!WARNING]
       >若只升级单个组件，values.yaml中其他已安装组件的配置请保持与安装时的配置一致，不能将其他已安装组件的Enabled参数设置为false，否则对应组件的资源会被删除。

       ```bash
       #（可选）--dry-run不实际创建任何资源，可以用来验证模板语法、检查生成的配置是否符合预期
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml --dry-run
       # 正式执行升级
       helm upgrade mindcluster mindcluster-deploy-tool-{chart_version}.tgz -f values.yaml
       ```

       回显示例如下：

       ```ColdFusion
       Release "mindcluster" has been upgraded. Happy Helming!
       NAME: mindcluster
       LAST DEPLOYED: ...
       NAMESPACE: default
       STATUS: deployed
       REVISION: 2
       TEST SUITE: None
       ```

4. 确认组件升级状态，详细请参见[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。
5. 若升级后，组件状态异常，可排查异常原因，然后按照如下方法处理：
   - 参考[版本回退](#section_helm_rollback)回退到升级前的版本。
   - 修改配置后重新[使用helm upgrade升级](#section_helm_upgrade)。
   - [使用helm卸载](../05_uninstallation/02_helm_uninstallation.md#ZH-CN_TOPIC_0000002511426390)组件后，重新[使用helm安装](../02_installation/00_helm_installation.md#ZH-CN_centerIC_0000002479226452)组件。此方法可能会导致工作负载中断，请用户在升级前确认。

## 版本回退<a name="section_helm_rollback"></a>

若升级后组件运行异常，可通过helm的回退功能恢复到升级前的版本。版本回退仅适用于通过helm upgrade升级过的Release实例，helm会记录每次升级的Revision历史。

1. 执行以下命令，查看Release的升级历史。
    - 查看应用组件Release实例的升级历史：

      ```bash
      helm history mindcluster
      ```

      回显示例如下：

      ```ColdFusion
      REVISION  UPDATED                   STATUS      CHART                                APP VERSION  DESCRIPTION
      1         2026-03-24 15:30:00.000   superseded  mindcluster-deploy-tool-26.0.0        26.0.0       Install complete
      2         2026-03-25 10:00:00.000   deployed    mindcluster-deploy-tool-26.1.0        26.1.0       Upgrade complete
      ```

    - 查看组件CRD的Release实例的升级历史：

       ```bash
       helm history mindcluster-crds
       ```

       回显示例如下：

      ```ColdFusion
      REVISION  UPDATED                   STATUS      CHART                                APP VERSION  DESCRIPTION
      1         2026-03-24 15:30:00.000   superseded  mindcluster-crds-deploy-tool-26.0.0        26.0.0       Install complete
      2         2026-03-25 10:00:00.000   deployed    mindcluster-crds-deploy-tool-26.1.0        26.1.0       Upgrade complete
      ```

    >[!NOTE]
    >Helm的REVISION是一个简单的递增整数，它的核心作用是记录和回滚：每次应用变更都会生成一个新的REVISION，需要时可以通过REVISION号快速恢复到过去的任意稳定版本，从而实现应用发布的可追溯和故障快速恢复。

2. 执行以下命令，回退crd到指定Revision版本。以回退到REVISION 1为例：

    ```bash
    helm rollback mindcluster-crds 1
    ```

    回显示例如下：

    ```ColdFusion
    Rollback was a success! Happy Helming!
    ```

3. 执行以下命令，回退应用到指定的Revision版本。以回退到REVISION 1为例：

    ```bash
    helm rollback mindcluster 1
    ```

    回显示例如下：

    ```ColdFusion
    Rollback was a success! Happy Helming!
    ```

4. 确认组件运行状态，详细请参见[组件状态确认](../03_confirming_status.md#ZH-CN_TOPIC_0000002479386390)章节。
