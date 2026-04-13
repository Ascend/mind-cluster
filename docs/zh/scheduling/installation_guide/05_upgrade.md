# 升级<a name="ZH-CN_TOPIC_0000002479226452"></a>

## 升级说明<a name="ZH-CN_TOPIC_0000002511346381"></a>

本章节旨在指导用户将MindCluster集群调度组件升级到新版本。MindCluster集群调度组件的升级支持以下2种方式。

- 全量升级：此种升级方式不仅会升级各组件的二进制镜像文件，而且升级后可对组件的配置文件进行修改。此种升级方式支持跨版本升级，例如，用户可从5.0.x版本升级到7.0.x版本。
- 升级镜像：此种升级方式仅升级各组件的二进制文件，不支持修改权限、启动参数等，无需进行升级前环境检查。此种升级方式仅支持在同一个版本内进行升级。

    **表 1**  升级方式说明

    <a name="table1527494117524"></a>
    <table><thead align="left"><tr id="row327404115216"><th class="cellrowborder" valign="top" width="17.5%" id="mcps1.2.5.1.1"><p id="p627494165216"><a name="p627494165216"></a><a name="p627494165216"></a>升级方式</p>
    </th>
    <th class="cellrowborder" valign="top" width="25.990000000000002%" id="mcps1.2.5.1.2"><p id="p92749419529"><a name="p92749419529"></a><a name="p92749419529"></a>是否支持跨版本升级</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.240000000000002%" id="mcps1.2.5.1.3"><p id="p19274134120522"><a name="p19274134120522"></a><a name="p19274134120522"></a>是否需要停止训练/推理任务</p>
    </th>
    <th class="cellrowborder" valign="top" width="26.27%" id="mcps1.2.5.1.4"><p id="p15533184405419"><a name="p15533184405419"></a><a name="p15533184405419"></a>参考章节</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1727434112526"><td class="cellrowborder" valign="top" width="17.5%" headers="mcps1.2.5.1.1 "><p id="p1027414185220"><a name="p1027414185220"></a><a name="p1027414185220"></a>全量升级</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.990000000000002%" headers="mcps1.2.5.1.2 "><p id="p3274841105220"><a name="p3274841105220"></a><a name="p3274841105220"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.240000000000002%" headers="mcps1.2.5.1.3 "><p id="p927454111524"><a name="p927454111524"></a><a name="p927454111524"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.27%" headers="mcps1.2.5.1.4 "><p id="p6533944195419"><a name="p6533944195419"></a><a name="p6533944195419"></a><a href="#升级说明">升级说明</a>-<a href="#升级其他组件">升级其他组件</a>章节</p>
    </td>
    </tr>
    <tr id="row8274241115212"><td class="cellrowborder" valign="top" width="17.5%" headers="mcps1.2.5.1.1 "><p id="p202747416524"><a name="p202747416524"></a><a name="p202747416524"></a>升级镜像</p>
    </td>
    <td class="cellrowborder" valign="top" width="25.990000000000002%" headers="mcps1.2.5.1.2 "><p id="p1327413412527"><a name="p1327413412527"></a><a name="p1327413412527"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.240000000000002%" headers="mcps1.2.5.1.3 "><p id="p3274144175214"><a name="p3274144175214"></a><a name="p3274144175214"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="26.27%" headers="mcps1.2.5.1.4 "><p id="p25334441543"><a name="p25334441543"></a><a name="p25334441543"></a><a href="#升级镜像">升级镜像</a>章节</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE] 
    >本章节不适用的场景：用户对旧版本MindCluster集群调度组件的源代码（不含配置文件）进行了修改，请分析版本代码差异后再进行升级。

**升级环境检查<a name="section19242859587"></a>**

在进行各组件的升级步骤前，请根据实际安装场景，选择相应的组件进行检查。

1. 检查是否有正在运行的任务。若用户正在执行的任务，请等待任务执行完成或提前停止任务后，再升级MindCluster组件。
    1. 请执行以下命令检查是否有正在运行的任务。

        ```shell
        kubectl get pods -A
        ```

        回显示例如下。

        ```ColdFusion
        NAMESPACE        NAME                                       READY   STATUS    RESTARTS         AGE
        default          ubuntu-pod                                 1/1     Running   32 (118m ago)    3d18h ...  
        ```

    2. 进入任务YAML所在路径，执行以下命令停止任务。

        ```shell
        kubectl delete -f  xxx.yaml              # xxx表示任务YAML的名称，请根据实际情况填写    
        ```

2. （可选）检查pingmesh灵衢网络检测开关是否已关闭。
    1. 登录环境，进入NodeD解压目录。
    2. 执行以下命令编辑pingmesh-config文件。

        ```shell
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

        如果回显如下所示，表示pingmesh灵衢网络检测开关已关闭。无需执行[步骤3](#li1427143773119)。

        ```ColdFusion
        Error from server (NotFound): configmaps "pingmesh-config" not found
        ```

    3. <a name="li1427143773119"></a>（可选）修改activate字段的取值。
        - 如果超节点ID在pingmesh-config文件中，修改该超节点ID字段下的activate为off。
        - 如果超节点ID不在pingmesh-config文件中，可通过以下2种方式进行设置。
            - 在配置文件中新增该超节点信息，并将activate为off。
            - 删除pingmesh-config文件中所有超节点的信息，并将global配置中activate字段的值设置为off。

3. 检查已安装的MindCluster组件。
    - （可选）**检查TaskD组件**。执行以下命令进入容器内部，查看TaskD组件安装状态。

        ```shell
        docker run -it  {训练镜像名称}:tag /bin/bash
        pip show taskd
        ```

        回显如下，表示镜像中已安装TaskD组件。

        ```ColdFusion
        Name: taskd
        Version: x.x.x
        Summary: Ascend MindCluster taskd is a new library for training management
        Home-page: UNKNOWN
        Author: 
        Author-email: 
        License: UNKNOWN
        Location: /usr/local/python3/lib/python3.10/site-packages
        Requires: grpcio, protobuf, pyOpenSSL, torch, torch-npu
        Required-by:
        ```

    - （可选）**检查其他组件**。参考[组件状态确认](./04_confirming_status.md)，确认集群中节点是否安装了相应组件。

4. （可选）若尚未安装MindCluster集群调度组件，请参考[安装部署](./03_installation.md)章节先安装组件，TaskD的安装步骤请参考[制作镜像](../usage/resumable_training.md#制作镜像)章节。

## 升级Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226420"></a>

仅Ascend Docker Runtime支持通过命令行进行升级，其他集群调度组件可通过卸载后重新安装进行升级。

目前只支持root用户升级Ascend Docker Runtime。

**前提条件<a name="section176591058124515"></a>**

已完成[升级环境检查](#升级说明)。

**升级步骤<a name="section520182224617"></a>**

1. 下载新版本组件安装包，详情请参见参考[获取软件包](./03_installation.md#获取软件包)章节。
2. <a name="li12599722163212"></a>进入安装包（run包）所在路径，在该路径下执行以下命令为软件包添加可执行权限。

    ```shell
    cd <path to run package>
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. 通过以下命令升级Ascend Docker Runtime。
    - （可选）在默认路径下升级Ascend Docker Runtime，需要依次执行以下命令。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade
        ```

    - （可选）在指定路径下升级Ascend Docker Runtime，需要依次执行以下命令。“--install-path”参数为指定的升级路径。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade --install-path=<path>
        ```

        回显示例如下，表示升级成功。

        ```ColdFusion
        Uncompressing ascend-docker-runtime  100%
        ...
        [INFO] ascend-docker-runtime upgrade success
        ```

4. （可选）执行以下命令重启容器，使新版Ascend Docker Runtime生效。如不涉及安装路径、安装参数变更，可跳过本步骤。
    - Docker场景（或K8s集成Docker场景）

        ```shell
        systemctl daemon-reload && systemctl restart docker
        ```

    - Containerd场景（或K8s集成Containerd场景）

        ```shell
        systemctl daemon-reload && systemctl restart containerd
        ```

5. <a name="li76002022113215"></a>参考[组件状态确认](./04_confirming_status.md)章节，检查新版本Ascend Docker Runtime是否升级成功状态。
6. （可选）恢复旧版本。下载旧版本安装包，依次重新执行[步骤2](#li12599722163212)到[步骤5](#li76002022113215)。

## 升级TaskD<a name="ZH-CN_TOPIC_0000002479226444"></a>

TaskD组件安装在训练镜像内部，在训练镜像内部重新安装该whl包即可完成升级。

**前提条件<a name="section18616132394915"></a>**

已完成[升级环境检查](#升级说明)。

**升级步骤<a name="section1720814439492"></a>**

1. 参考[获取软件包](./03_installation.md#获取软件包)章节，下载新版本组件安装包。
2. 下载完成后，进入安装包所在路径并解压安装包。
3. 执行**ls -l**命令，回显示例如下。

    ```ColdFusion
    -rw-r--r-- 1 root root 1493228 Mar 14 02:09 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 1506842 Mar 12 18:07 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4. 基于已有的训练镜像，安装新版本TaskD组件。
    1. 执行以下命令运行训练镜像。

        ```shell
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest bash
        ```

    2. 执行以下命令卸载已安装的TaskD组件。

        ```shell
        pip uninstall taskd -y
        ```

        回显示例如下表示卸载成功。

        ```ColdFusion
        Successfully uninstalled taskd-{version}
        ```

    3. 执行以下命令安装新版本TaskD。

        ```shell
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        回显如下。

        ```ColdFusion
        Successfully installed taskd-{version}
        ```

    4. 安装了新版本TaskD后，将容器保存为新镜像。

        ```shell
        docker ps
        ```

        回显示例如下。

        ```ColdFusion
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES
        8b70390775f2   fd6acb527bad           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        将该容器提交为新版本训练容器镜像，注意新镜像的tag与旧镜像不一致。示例如下。

        ```shell
        docker commit 8b70390775f2 newimage:latest
        ```

5. 检查新版TaskD是否升级完成，参考[检查TaskD组件](#升级说明)章节，检查组件状态是否正常。
6. （可选）回退老版本。若旧版镜像仍然存在，无需回退操作；若不存在则按上述步骤，重新安装旧版本TaskD软件包即可。

## 升级Container Manager<a name="ZH-CN_TOPIC_0000002524548731"></a>

在物理机上直接替换Container Manager二进制升级组件。

1. 以root用户登录Container Manager组件部署的节点。
2. 将获取到的Container Manager软件包上传至服务器的任意目录（如“/tmp/container-manager”）。
3. 进入“/tmp/container-manager”目录并进行解压操作。

    ```shell
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

    >[!NOTE] 
    ><i>\<version\></i>为软件包的版本号；<i>\<arch\></i>为CPU架构。

4. 依次执行以下命令，升级Container Manager组件。

    ```shell
    # 停止Container Manager系统服务，并删除对应Container Manager二进制文件
    systemctl stop container-manager.service
    chattr -i /usr/local/bin/container-manager
    rm -f /usr/local/bin/container-manager
    
    # 从解压文件中获取新二进制文件，替换旧Container Manager二进制文件
    cp /tmp/container-manager/container-manager /usr/local/bin
    chmod 500 /usr/local/bin/container-manager
    
    # 重启Container Manager系统服务
    systemctl daemon-reload
    systemctl start container-manager.service
    ```

5. 验证Container Manager组件的升级状态。
    1. 查看组件服务的状态，需要满足组件状态为active \(running\)。

        ```shell
        systemctl status container-manager.service
        ```

        回显示例：

        ```ColdFusion
        ● container-manager.service - Ascend container manager
             Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
             Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
            Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
           Main PID: 41464 (container-manag)
              Tasks: 10 (limit: 629145)
             Memory: 13.3M
             CGroup: /system.slice/container-manager.service
                     └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
        ...
        ```

    2. 查看组件日志。

        ```shell
        cat /var/log/mindx-dl/container-manager/container-manager.log
        ```

        回显以Atlas 800I A3 超节点服务器为例：

        ```ColdFusion
        [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
        [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
        [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
        [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
        [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
        [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
        ...
        ```

        如果出现如下打印信息，表示组件运行正常。

        ```ColdFusion
        ...
        [INFO]     2025/11/25 22:46:59.289352 1       devmgr/workflow.go:57    init module <hwDev manager> success
        [INFO]     2025/11/25 22:46:59.293773 1       app/config.go:40    load fault config from /home/faultCode.json success
        [INFO]     2025/11/25 22:46:59.293866 1       app/workflow.go:50    init module <fault manager> success
        [INFO]     2025/11/25 22:46:59.293901 1       app/workflow.go:76    init module <container controller> success
        [INFO]     2025/11/25 22:46:59.293930 1       app/workflow.go:64    init module <reset-manager> success
        [INFO]     2025/11/25 22:46:59.315101 378     devmgr/hwdevmgr.go:365    subscribe device fault event success
        ...
        ```

## 升级其他组件<a name="ZH-CN_TOPIC_0000002511346401"></a>

**前提条件<a name="section176591058124515"></a>**

- 已完成[升级环境检查](#升级说明)。

- 如需升级NPU Exporter、Ascend Device Plugin、Volcano、ClusterD、Ascend Operator、Infer Operator、NodeD和Resilience Controller组件，需卸载旧版本后，再执行新版本的安装步骤。

**升级步骤<a name="section65996266718"></a>**

1. 卸载MindCluster旧版本组件。详情请参见[卸载其他组件](./06_uninstallation.md)中"卸载组件"步骤。
2. 参考[获取软件包](./03_installation.md#获取软件包)章节，下载新版本组件安装包。
3. （可选）准备MindCluster集群调度组件新版本镜像。若新版本组件采用二进制方式安装，可跳过本步骤。

    参考[准备镜像](./03_installation.md#准备镜像)章节，从昇腾镜像仓库拉取新版本镜像或者制作新版本镜像。注意新版本组件镜像tag要与旧版本组件镜像tag不一致，避免覆盖旧版本组件镜像。

4. <a name="li147194506333"></a>请根据要升级的组件，重新执行手动安装步骤。详细步骤请参见[安装MindCluster新版本组件](./03_installation.md)。
5. （可选）如需回退老版本，依次执行[卸载](./06_uninstallation.md)中"卸载其他组件"中卸载组件步骤和[步骤4](#li147194506333)，卸载新版本组件后安装旧版本组件即可。

## 升级镜像<a name="ZH-CN_TOPIC_0000002511346311"></a>

本章节仅指导用户在同一个版本内对容器镜像中二进制文件版本进行升级，升级过程中不会修改权限及启动参数。如需了解关于升级方式的更详细说明，请参见[升级说明](#升级说明)。

- 如需升级Volcano、ClusterD、Ascend Operator、Infer Operator、Resilience Controller组件的镜像，可参考[升级管理节点组件](#section1292111716589)。
- 如需升级NPU Exporter、Ascend Device Plugin和NodeD组件镜像，可参考[升级计算节点组件](#section231311416588)。
- TaskD暂不支持此种升级方式。

**升级管理节点组件<a name="section1292111716589"></a>**

1. 参考[准备镜像](./03_installation.md#准备镜像)章节，使用新的软件包制作镜像。

    >[!NOTE]
    >请保持镜像名称一致，否则可能导致原配置文件无法拉起Pod。

2. 执行以下命令，查询旧版本Deployment配置。

    ```shell
    kubectl get deployment -A|grep {组件名称}
    ```

    以ClusterD组件为例，回显示例如下。

    ```ColdFusion
    mindx-dl         clusterd        1/1     1      1       45h
    ```

3. 执行以下命令，重启Deployment。

    ```shell
    kubectl rollout restart deployment -n {命名空间名称} {deployment名称}
    ```

    以ClusterD组件为例，回显示例如下。

    ```ColdFusion
    deployment.apps/clusterd restarted
    ```

4. 检查新版本Pod是否已拉起。

    ```shell
    kubectl get pod -A|grep {组件名称}
    ```

    以ClusterD为例，回显示例如下，表示Pod成功拉起。

    ```ColdFusion
    mindx-dl   clusterd-99f8795c8-drqb4  1/1  Running 0       1m
    ```

**升级计算节点组件<a name="section231311416588"></a>**

1. 参考[准备镜像](./03_installation.md#准备镜像)章节，使用新的软件包制作镜像。

    >[!NOTE] 
    >请保持镜像名称一致，否则可能导致原配置文件无法拉起Pod。

2. 执行以下命令，查询旧版本DaemonSet配置。

    ```shell
    kubectl get ds -A|grep {组件名称}
    ```

    以NodeD组件为例，回显示例如下。

    ```ColdFusion
    mindx-dl         noded        1/1     1      1       45h
    ```

3. 执行以下命令，重启DaemonSet。

    ```shell
    kubectl rollout restart ds -n {命名空间名称} {ds名称}
    ```

    以NodeD组件为例，回显示例如下。

    ```ColdFusion
    daemonsets.apps/noded restarted
    ```

4. 检查新版本Pod是否已拉起。

    ```shell
    kubectl get pod -A|grep {组件名称}
    ```

    以NodeD为例，回显示例如下，表示Pod已拉起。

    ```ColdFusion
    mindx-dl   noded- m4j4r  1/1  Running 0     1m
    ```

## Elastic Agent升级TaskD<a name="ZH-CN_TOPIC_0000002515202401"></a>

Elastic Agent组件已经日落，本章节提供将Elastic Agent组件升级为TaskD组件的操作指导。

**前提条件<a name="section565512391204"></a>**

- 已完成升级环境检查。
- 训练镜像已安装Elastic Agent。

**操作步骤<a name="section1643711813"></a>**

1. 参考[获取软件包](./03_installation.md#获取软件包)章节，下载新版本TaskD组件安装包。
2. 下载完成后，进入安装包所在路径并解压安装包。
3. 执行**ls -l**命令，回显示例如下。

    ```ColdFusion
    -rw-r--r-- 1 root root 6134726 Nov 10 10:32 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 6205642 Nov  5 23:38 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4. 基于已有的训练镜像，卸载Elastic Agent并安装新版本TaskD。
    1. 运行训练镜像。示例如下：

        ```shell
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest /bin/bash
        ```

    2. 卸载已安装的Elastic Agent组件。

        ```shell
        pip uninstall mindx-elastic -y
        ```

        回显示例如下，表示卸载成功。

        ```ColdFusion
        Successfully uninstalled mindx_elastic-{version}
        ```

    3. 删除Elastic Agent使能代码。

        ```shell
        sed -i '/mindx_elastic.api/d' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

        （可选）执行以下命令，查看对应文件是否已经删除Elastic Agent嵌入代码。

        ```shell
        vi $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    4. 安装新版本TaskD。

        ```shell
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        回显示例如下，表示安装成功。

        ```ColdFusion
        Successfully installed taskd-{version}
        ```

        执行以下命令，使能TaskD。

        ```shell
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    5. 安装了新版本TaskD后，将容器保存为新镜像。

        ```shell
        docker ps
        ```

        回显示例如下。

        ```ColdFusion
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES 
        bb118ca00041    f76142d63d3a           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        将该容器提交为新版本训练容器镜像，注意新镜像的tag与旧镜像不一致。示例如下：

        ```shell
        docker commit bb118ca00041 newimage:latest
        ```

5. 检查TaskD是否替换完成。参考[检查TaskD](#升级说明)章节，检查组件状态是否正常。
6. 修改训练脚本（例如train\_start.sh）和任务YAML。
    1. 创建manager.py文件，放在调用训练脚本时的当前目录下，manager.py文件内容如下所示。

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os 
         
        job_id=os.getenv("MINDX_TASK_ID") 
        node_nums=XX         # 用户填入任务节点总数
        proc_per_node=XX     # 用户填入任务每个节点的训练进程数量
          
        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node}) 
        start_taskd_manager()
        ```

        >[!NOTE]
        >manager.py文件中的参数详细说明请参见[def init\_taskd\_manager\(config:dict\) -\> bool:](../api/taskd.md#def-init_taskd_managerconfigdict---bool)。

    2. 在训练脚本中增加以下代码拉起TaskD  Manager。

        <pre codetype="Python">
        <strong>export TASKD_PROCESS_ENABLE="on" 
        # 以PyTorch框架为例
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP} 
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # 具体执行路径由当前路径决定
        fi</strong> 
              
        torchrun ...</pre>

    3. 在任务YAML中修改容器端口，在所有的Pod下增加TaskD通信使用的端口9601。

        <pre codetype="yaml">
        ...
                spec:
        ...
                   containers:
        ...
                     <strong>ports:                          
                       - containerPort: 9601              
                         name: taskd-port</strong>
        ...</pre>
