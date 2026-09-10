# Node Collector<a name="ZH-CN_TOPIC_0000002524428762"></a>

- 使用[集群运维Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690)特性的用户，必须安装Node Collector。
- Node Collector以DaemonSet方式部署在每个计算节点，负责从宿主机采集任务日志并在本地完成日志清洗。
- Node Collector的部署通过节点标签选择计算节点，仅会调度到带有 `workerselector=dls-worker-node` 标签的节点上。节点标签的创建请参见[创建节点标签](./01_preparing_for_installation.md#创建节点标签)。
- Node Collector和Agent Core共用同一个镜像ascend-clusterops-agent，镜像的获取（制作或拉取）请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)。
- 部署Node Collector前，需先完成Agent Core的部署，详细说明请参见 [Agent Core](./16_agent_core.md)。

## 操作步骤<a name="section15023132772914"></a>

1. 以root用户登录K8s管理节点，并执行以下命令，查看Node Collector镜像和版本号是否正确。

    ```shell
    docker images | grep ascend-clusterops-agent
    ```

    回显示例如下：

    ```ColdFusion
    ascend-clusterops-agent       v26.2.0              c532e9d0889c        About an hour ago         112MB
    ```

    - 是，执行[步骤2](#li4121412112712)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和拉取。

2. <a name="li4121412112712"></a>将Node Collector软件包解压目录下的YAML文件（node-collector.yaml），拷贝到K8s管理节点上任意目录。

3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Node Collector的启动参数。
   - 节点标签：Node Collector通过 `nodeSelector` 字段的 `workerselector=dls-worker-node` 标签选择计算节点，请根据实际节点标签修改。
   - 更新采集契约：采集契约（collect_manifest.yaml）默认内置在镜像内，如需现场调整，可通过Kubectl Plugin执行以下命令将本地采集契约写入集群ConfigMap。

     ```shell
     kubectl ascend_diag --collect-manifest {collect_manifest.yaml}
     ```

4. 在管理节点的YAML所在路径，执行以下命令，启动Node Collector。

    ```shell
    kubectl apply -f node-collector.yaml
    ```

    启动示例如下：

    ```ColdFusion
    serviceaccount/node-collector created
    ...
    daemonset.apps/node-collector created
    service/node-collector created
    ```

5. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，每个计算节点均出现 **Running** 状态的Pod表示组件启动成功。

    ```ColdFusion
    NAME                    READY   STATUS    RESTARTS   AGE
    node-collector-8j4kq    1/1     Running   0          45s
    node-collector-m4j4r    1/1     Running   0          45s
    ```
