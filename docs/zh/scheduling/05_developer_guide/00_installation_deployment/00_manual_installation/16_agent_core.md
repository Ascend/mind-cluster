# Agent Core<a name="ZH-CN_TOPIC_0000002524428761"></a>

- 使用[集群运维Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690)特性的用户，必须安装Agent Core。
- Agent Core以Deployment方式部署在管理节点，作为故障诊断的集中控制中心，负责接收诊断请求、按任务维度调度各节点采集日志并执行集中诊断。
- Agent Core和Node Collector共用同一个镜像ascend-clusterops-agent，镜像的获取（制作或拉取）请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)。
- 部署Agent Core前，需先完成[安装前准备](./01_preparing_for_installation.md)中的创建用户、创建日志目录和创建命名空间步骤。

## 操作步骤<a name="section15023132772914"></a>

1. 以root用户登录K8s管理节点，并执行以下命令，查看Agent Core镜像和版本号是否正确。

    ```shell
    docker images | grep ascend-clusterops-agent
    ```

    回显示例如下：

    ```ColdFusion
    ascend-clusterops-agent       v26.2.0              c532e9d0889c        About an hour ago         698MB
    ```

    - 是，执行[步骤2](#li4121412112711)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和拉取。

2. <a name="li4121412112711"></a>将Agent Core软件包解压目录下的YAML文件（agent-core.yaml），拷贝到K8s管理节点上任意目录。

3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Agent Core的启动参数。

4. 在管理节点的YAML所在路径，执行以下命令，启动Agent Core。

    ```shell
    kubectl apply -f agent-core.yaml
    ```

    启动示例如下：

    ```ColdFusion
    configmap/agent-core-task-crds created
    serviceaccount/agent-core created
    ...
    deployment.apps/agent-core created
    service/agent-core created
    ```

5. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现 **Running** 表示组件启动成功。

    ```ColdFusion
    NAME                          READY   STATUS    RESTARTS   AGE
    agent-core-7844cb867d-fwcj7   1/1     Running   0          45s
    ```
