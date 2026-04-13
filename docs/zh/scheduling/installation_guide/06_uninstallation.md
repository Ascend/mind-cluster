# 卸载<a name="ZH-CN_TOPIC_0000002511426389"></a>

- 卸载Ascend Docker Runtime组件，请参见[卸载Ascend Docker Runtime](#section6134163311244)进行操作。
- 卸载Container Manager组件，请参见[卸载Container Manager组件](#section1461059103619)进行操作。
- 卸载NPU Exporter、Ascend Device Plugin、Volcano、ClusterD、Ascend Operator、Infer Operator、NodeD和Resilience Controller，请参见[卸载其他组件](#section6361146202520)。

**卸载Ascend Docker Runtime<a name="section6134163311244"></a>**

- 情况一：使用不同安装路径。

    用户在卸载Ascend Docker Runtime时需要针对不同容器引擎，根据[步骤2](#li345320287225)进行两次卸载操作，每次卸载需要指定相应的安装路径，即--install-path参数。

- 情况二：使用相同安装路径。

    用户在卸载Ascend Docker Runtime时，只需根据[步骤2](#li345320287225)进行一次卸载操作。卸载完成之后需要手动将另一引擎的daemon.json文件还原为Ascend Docker Runtime安装之前的内容。

若用户需要保留其中一个容器引擎，需要在Ascend Docker Runtime卸载之后，针对相应场景进行重新安装。

1. （可选）关闭pingmesh灵衢网络检测。
    1. 登录环境，进入NodeD解压目录。
    2. 执行以下命令编辑pingmesh-config文件。

        ```shell
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3. 修改activate字段的取值。
        - 如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        - 如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            - 在配置文件中新增该超节点信息，并将activate为off。
            - 删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

2. <a name="li345320287225"></a>可以选择以下方式中的一种卸载Ascend Docker Runtime软件。
    - 方式一：（推荐）使用软件包卸载
        1. 首先进入安装包（run包）所在路径。

            ```shell
            cd <path to run package>
            ```

        2. 执行以下卸载命令，在**默认安装路径**下卸载Ascend Docker Runtime。

            - Docker场景（或K8s集成Docker场景）

                ```shell
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall
                ```

            - Containerd场景（或K8s集成Containerd场景）

                ```shell
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall --install-scene=containerd
                ```

            >[!NOTE] 
            >- Docker配置文件路径不是默认的“/etc/docker/daemon.json”时，需要新增“--config-file-path”参数，用于指定该配置文件路径。
            >- Containerd的配置文件路径不是默认的“/etc/containerd/config.toml”时，需要新增“--config-file-path”参数，用于指定该配置文件路径。
            >- 如需要卸载指定安装路径下的Ascend Docker Runtime，需要在卸载命令中新增“--install-path=<path\>”参数。

            回显示例如下，表示卸载成功。

            ```ColdFusion
            Uncompressing ascend-docker-runtime  100%
            ...
            [INFO] ascend-docker-runtime uninstall success
            ```

    - 方式二：使用脚本卸载

        1. 首先进入Ascend Docker Runtime的安装路径下的“script”目录（默认安装路径为：“/usr/local/Ascend/Ascend-Docker-Runtime”）：

            ```shell
            cd /usr/local/Ascend/Ascend-Docker-Runtime/script
            ```

        2. 运行卸载的脚本进行卸载。

            - Docker场景（或K8s集成Docker场景）

                ```shell
                uninstall.sh docker docker <daemon.json文件路径>
                ```

            - Containerd场景（或K8s集成Containerd场景）

                ```shell
                uninstall.sh containerd containerd <config.toml文件路径>
                ```

            >[!NOTE]
            >- 可以不指定Docker的配置文件daemon.json路径，不指定时默认使用“/etc/docker/daemon.json”。
            >- 可以不指定Containerd的配置文件config.toml路径，不指定时默认使用“/etc/containerd/config.toml”。

        回显示例如下，表示卸载成功。

        ```ColdFusion
        [INFO]: You will recover Docker's daemon
        ...
        [INFO] uninstall.sh exec success
        ```

3. （可选）在K8s集成Containerd的场景下，如果需要还原修改的kubeadm-flags.env，请参见[K8s官方文档](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/)，还原配置文件kubeadm-flags.env。其他场景可跳过该步骤。
4. 重启服务。
    - Docker场景（或K8s集成Docker场景）

        ```shell
        systemctl daemon-reload && systemctl restart docker
        ```

    - Containerd场景（或K8s集成Containerd场景）

        ```shell
        systemctl daemon-reload && systemctl restart containerd
        ```

**卸载Container Manager组件<a name="section1461059103619"></a>**

1. 以root用户登录Container Manager组件部署的节点。
2. 依次执行以下命令，卸载Container Manager组件系统服务。

    ```shell
    # 停止Container Manager系统服务
    systemctl stop container-manager.timer
    systemctl disable container-manager.timer
    systemctl stop container-manager.service
    systemctl disable container-manager.service
    
    # 删除Container Manager系统服务
    rm -f /etc/systemd/system/container-manager.service
    rm -f /etc/systemd/system/container-manager.timer
    systemctl daemon-reload
    systemctl reset-failed
    
    # 删除对应Container Manager二进制文件
    chattr -i /usr/local/bin/container-manager
    rm -f /usr/local/bin/container-manager
    ```

3. 删除日志文件，请确认实际路径后再删除。

    ```shell
    rm -rf /var/log/mindx-dl/container-manager
    ```

**卸载其他组件<a name="section6361146202520"></a>**

支持卸载集群调度组件，用户可以卸载组件后重新安装最新版本组件。通过逐一卸载各组件，并删除对应的命名空间、日志目录、配置文件等，请根据安装方式选择对应的卸载方式。

1. （可选）关闭pingmesh灵衢网络检测。
    1. 登录环境，进入NodeD解压目录。
    2. 执行以下命令编辑pingmesh-config文件。

        ```shell
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3. 修改activate字段的取值。
        - 如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        - 如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            - 在配置文件中新增该超节点信息，并将activate为off。
            - 删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

2. 卸载组件。根据组件的安装方式，选择以下对应的卸载方式。
    - 通过容器方式卸载。各组件卸载方法类似，均为进入该组件配置文件YAML所在目录，并执行删除操作实现，此操作需要在K8s的管理节点操作。以卸载Ascend Device Plugin为例说明，请用户自行完成其余组件卸载。

        1. 以root用户登录管理节点。
        2. 进入Ascend Device PluginYAML配置文件所在目录（如：“/home/ascend-device-plugin”）。

            ```shell
            cd /home/ascend-device-plugin
            ```

        3. 在Ascend Device Plugin组件安装环境下，执行以下命令，卸载Ascend Device Plugin。

            ```shell
            kubectl delete -f device-plugin-volcano-v{version}.yaml
            ```

            回显示例如下：

            ```ColdFusion
            serviceaccount "ascend-device-plugin-sa-910" deleted
            clusterrole.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-role-910" deleted
            clusterrolebinding.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-rolebinding-910" deleted
            deployment.apps "ascend-device-plugin-daemonset-910" deleted
            ```

        >[!NOTE] 
        >Ascend Device Plugin配合Volcano使用时，会创建ConfigMap，执行如下命令进行删除。
        >
        >```shell
        >kubectl delete cm mindx-dl-deviceinfo-<node-name> -n kube-system
        >```

    - 通过二进制方式卸载。以卸载NPU Exporter为例说明，请用户自行完成其余组件卸载。
        1. 以root用户登录组件部署的节点。
        2. 在NPU Exporter组件安装环境下，依次执行如下命令卸载NPU Exporter组件。

            ```shell
            systemctl stop npu-exporter.service
            systemctl disable npu-exporter.service
            chattr -i /etc/systemd/system/npu-exporter.service
            rm -f /etc/systemd/system/npu-exporter.service
            systemctl daemon-reload
            systemctl reset-failed
            chattr -i /usr/local/bin/npu-exporter
            rm -f /usr/local/bin/npu-exporter
            ```

3. 删除命名空间。NPU Exporter的命名空间npu-exporter和Volcano的命名空间volcano-system在卸载组件时就已经同步删除，用户可以跳过本步骤。

    执行如下命令，卸载安装集群调度组件时创建的namespace。删除namespace会删除该namespace下的所有资源，请确认后再执行。

    ```shell
    kubectl delete ns mindx-dl
    ```

    回显示例如下：

    ```ColdFusion
    namespace "mindx-dl" deleted
    ```

4. 删除日志文件。参考[创建日志目录](./03_installation.md#创建日志目录)章节，在对应节点上删除集群调度组件的日志目录。以ClusterD为例，请确认后再删除。

    ```shell
    rm -rf /var/log/mindx-dl/clusterd
    ```

5. （可选）卸载Resilience Controller时，若导入了证书和KubeConfig文件，则需要删除证书和KubeConfig文件，请确认后再删除。

    ```shell
    rm -rf /etc/mindx-dl/resilience-controller
    ```
