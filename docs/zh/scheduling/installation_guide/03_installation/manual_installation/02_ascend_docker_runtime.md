# Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226434"></a>

- 使用容器化支持、整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须安装Ascend Docker Runtime。
- 仅使用资源监测的用户，可以不安装Ascend Docker Runtime，请直接跳过本章节。

## 前提条件<a name="section137058405153"></a>

- 安装前，请确保runc文件的用户ID为0。
- 安装前，Containerd场景下请确保“/etc/containerd/config.toml”文件的用户及用户组为root；Docker场景或Isula场景下请确保“/etc/docker/daemon.json”文件的用户及用户组为root。
- 安装前，请确保“/etc/ld.so.preload”中配置的so文件的RPATH和RUNPATH不要包含带有相对路径的HDK驱动目录（可以通过“readelf -d xxx.so”命令查看）。

## 确认安装场景<a name="zh-cn_topic_0000001930317932_section1235447163310"></a>

目前仅支持root用户安装Ascend Docker Runtime，请根据实际情况选择对应的安装方式。

1. 在K8s管理节点执行以下命令，查询节点名称。

    ```shell
    kubectl get node
    ```

    回显示例如下：

    ```ColdFusion
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2. 查看当前节点的容器运行时。其中node-name为节点名称。
    - 不使用K8s场景：在任意节点执行以下命令。

        ```shell
        docker --version      # Docker
        containerd --version     # Containerd
        ```

        - 若回显为Docker的版本信息，表示当前是[Docker场景](#zh-cn_topic_0000001930317932_section1443063532919)。
        - 若回显为Containerd的版本信息，表示当前是[Containerd场景](#zh-cn_topic_0000001930317932_section196591123133116)。
        - 若同时有Docker和Containerd的版本信息，请用户自行确定任务所要使用的容器运行时。

    - K8s集成容器运行时场景：在管理节点执行以下命令。

        ```shell
        kubectl describe node <node-name> | grep -i runtime
        ```

        - 若回显中有Docker信息，表示当前是[K8s集成Docker场景](#zh-cn_topic_0000001930317932_section1443063532919)。
        - 若回显中有Containerd信息，表示当前是[K8s集成Containerd场景](#zh-cn_topic_0000001930317932_section14600174633116)。

## Docker场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section1443063532919"></a>

K8s集成Docker场景安装Ascend Docker Runtime，与Docker场景下安装Ascend Docker Runtime操作一致。

1. 安装包下载完成后，在所有计算节点，进入安装包（run包）所在路径。

    ```shell
    cd <path to run package>
    ```

2. 执行以下命令，为软件包添加可执行权限。

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. 执行如下命令，校验软件包安装文件的一致性和完整性。

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. 可通过以下命令安装Ascend Docker Runtime。

    - 安装到默认路径下，执行以下命令。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install
        ```

    - 安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-path=<path>
        ```

    >[!NOTE]
    >- 指定安装路径时必须使用绝对路径。
    >- Docker配置文件路径不是默认的“/etc/docker/daemon.json”时，需要新增--config-file-path参数，用于指定该配置文件路径。

    回显示例如下，表示安装成功。

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. 执行以下命令，使Ascend Docker Runtime生效。

    ```shell
    systemctl daemon-reload && systemctl restart docker
    ```

    Ascend Device Plugin在启动时会自动检测Ascend Docker Runtime是否存在，所以需要先启动Ascend Docker Runtime，再启动Ascend Device Plugin。若先启动Ascend Device Plugin，再启动Ascend Docker Runtime，需要参见[Ascend Device Plugin](./04_ascend_device_plugin.md)章节重新启动Ascend Device Plugin。

## Containerd场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section196591123133116"></a>

1. 安装包下载完成后，首先进入安装包（run包）所在路径。

    ```shell
    cd <path to run package>
    ```

2. 执行以下命令，为软件包添加可执行权限。

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. 执行如下命令，校验软件包安装文件的一致性和完整性。

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. 可通过以下命令安装Ascend Docker Runtime。

    - 安装到默认路径下。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    - 安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE] 
        >- 指定安装路径时必须使用绝对路径。
        >- Containerd的配置文件路径不是默认的“/etc/containerd/config.toml”时，需要新增--config-file-path参数，用于指定该配置文件路径。

    回显示例如下，表示安装成功。

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. <a name="zh-cn_topic_0000001930317932_section19659112313311605"></a>（可选）如果安装失败，可参照以下步骤修改Containerd配置文件。
    1. 打开配置文件。
        - **Containerd无默认配置文件场景**：依次执行以下命令，创建并修改配置文件。

            ```shell
            mkdir /etc/containerd
            containerd config default > /etc/containerd/config.toml
            vim /etc/containerd/config.toml
            ```

        - **Containerd已有配置文件场景**：打开并修改配置文件。

            ```shell
            vim /etc/containerd/config.toml
            ```

    2. 新增ascend runtime，并将其设置为默认runtime，示例如下所示。
       1. 在配置文件中找到如下runc配置内容（其中“io.containerd.cri.v1.runtime”在不同Containerd版本下可能不同，以实际为准）：

            <pre>
             [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc]
               runtime_type = 'io.containerd.runc.v2'
               runtime_path = ''
               pod_annotations = []
               container_annotations = []
               privileged_without_host_devices = false
               privileged_without_host_devices_all_devices_allowed = false
               cgroup_writable = false
               base_runtime_spec = ''
               cni_conf_dir = ''
               cni_max_conf_num = 0
               snapshotter = ''
               sandboxer = 'podsandbox'
               io_type = ''
                
               [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc.options]
                 BinaryName = ''
                 CriuImagePath = ''
                 CriuWorkPath = ''
                 IoGid = 0
                 IoUid = 0
                 NoNewKeyring = false
                 Root = ''
                 ShimCgroup = ''
                 SystemdCgroup = true
            ...</pre>

       2. 复制上述配置内容，将“runc”修改为“ascend”并配置“BinaryName”的值为ascend-docker-runtime可执行文件的安装路径，参考如下（其中“io.containerd.cri.v1.runtime”在不同Containerd版本下可能不同，以实际为准）：

            <pre>
             [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.ascend]
               runtime_type = 'io.containerd.runc.v2'
               runtime_path = ''
               pod_annotations = []
               container_annotations = []
               privileged_without_host_devices = false
               privileged_without_host_devices_all_devices_allowed = false
               cgroup_writable = false
               base_runtime_spec = ''
               cni_conf_dir = ''
               cni_max_conf_num = 0
               snapshotter = ''
               sandboxer = 'podsandbox'
               io_type = ''
                
               [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.ascend.options]
                 BinaryName = '/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime'
                 CriuImagePath = ''
                 CriuWorkPath = ''
                 IoGid = 0
                 IoUid = 0
                 NoNewKeyring = false
                 Root = ''
                 ShimCgroup = ''
                 SystemdCgroup = true
            ...</pre>
            
       3. 找到下述配置项，将其中的“default_runtime_name”的值修改为“ascend”（其中“io.containerd.cri.v1.runtime”在不同Containerd版本下可能不同，以实际为准）：

            修改前：
            <pre>
            [plugins.'io.containerd.cri.v1.runtime'.containerd]
              default_runtime_name = 'runc'
              ignore_blockio_not_enabled_errors = false
              ignore_rdt_not_enabled_errors = false
              ...</pre>
              
            修改后：
            <pre>
            [plugins.'io.containerd.cri.v1.runtime'.containerd]
              default_runtime_name = 'ascend'
              ignore_blockio_not_enabled_errors = false
              ignore_rdt_not_enabled_errors = false
            ...</pre>

6. 执行以下命令，重启Containerd。

    ```shell
    systemctl daemon-reload && systemctl restart containerd
    ```

## K8s集成Containerd场景下安装Ascend Docker Runtime<a name="zh-cn_topic_0000001930317932_section14600174633116"></a>

1. 安装包下载完成后，首先进入安装包（run包）所在路径。

    ```shell
    cd <path to run package>
    ```

2. 执行以下命令，为软件包添加可执行权限。

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. 执行如下命令，校验软件包安装文件的一致性和完整性。

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    回显示例如下：

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. 可通过以下命令安装Ascend Docker Runtime。

    - 安装到默认路径下。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    - 安装到指定路径下，执行以下命令，“--install-path”参数为指定的安装路径。

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE] 
        >指定安装路径时必须使用绝对路径。

    回显示例如下，表示安装成功。

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. （可选）如果安装失败，可参考[Containerd场景下安装Ascend Docker Runtime](#zh-cn_topic_0000001930317932_section196591123133116)中的[步骤5](#zh-cn_topic_0000001930317932_section19659112313311605)。

6. 如需将节点上的容器运行时从Docker更改为Containerd，需要修改节点上kubelet的配置文件kubeadm-flags.env。详情请参见[K8s官方文档](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/)。

7. 如果存在Docker服务，请执行以下命令停止对应服务。

    ```shell
    systemctl stop docker
    ```

8. 执行命令，重启Containerd和kubelet，示例如下。

    ```shell
    systemctl daemon-reload && systemctl restart containerd kubelet
    ```

## Ascend Docker Runtime安装包命令行参数说明<a name="zh-cn_topic_0000001930317932_section425619177219"></a>

参数说明如[表1](#zh-cn_topic_0000001930317932_table35676204212)所示。

**表 1**  安装包支持的参数说明

<a name="zh-cn_topic_0000001930317932_table35676204212"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001930317932_row1856732017219"><th class="cellrowborder" valign="top" width="32.43%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000001930317932_p155677203214"><a name="zh-cn_topic_0000001930317932_p155677203214"></a><a name="zh-cn_topic_0000001930317932_p155677203214"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="67.57%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000001930317932_p1456712016216"><a name="zh-cn_topic_0000001930317932_p1456712016216"></a><a name="zh-cn_topic_0000001930317932_p1456712016216"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001930317932_row2568112072119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p05681620192117"><a name="zh-cn_topic_0000001930317932_p05681620192117"></a><a name="zh-cn_topic_0000001930317932_p05681620192117"></a>--help | -h</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1356892011218"><a name="zh-cn_topic_0000001930317932_p1356892011218"></a><a name="zh-cn_topic_0000001930317932_p1356892011218"></a>查询帮助信息。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681520112117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p3568122042118"><a name="zh-cn_topic_0000001930317932_p3568122042118"></a><a name="zh-cn_topic_0000001930317932_p3568122042118"></a>--info</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p15568720122112"><a name="zh-cn_topic_0000001930317932_p15568720122112"></a><a name="zh-cn_topic_0000001930317932_p15568720122112"></a>查询软件包构建信息。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row756832062117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p6568142052120"><a name="zh-cn_topic_0000001930317932_p6568142052120"></a><a name="zh-cn_topic_0000001930317932_p6568142052120"></a>--list</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p4568182018212"><a name="zh-cn_topic_0000001930317932_p4568182018212"></a><a name="zh-cn_topic_0000001930317932_p4568182018212"></a>查询软件包文件列表。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row2568520172112"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p856882092110"><a name="zh-cn_topic_0000001930317932_p856882092110"></a><a name="zh-cn_topic_0000001930317932_p856882092110"></a>--check</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p185681720182113"><a name="zh-cn_topic_0000001930317932_p185681720182113"></a><a name="zh-cn_topic_0000001930317932_p185681720182113"></a>检查软件包的一致性和完整性。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681920202119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15568122012120"><a name="zh-cn_topic_0000001930317932_p15568122012120"></a><a name="zh-cn_topic_0000001930317932_p15568122012120"></a>--quiet</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p256818204217"><a name="zh-cn_topic_0000001930317932_p256818204217"></a><a name="zh-cn_topic_0000001930317932_p256818204217"></a>静默安装，跳过交互式信息，需要配合install、uninstall或者upgrade使用。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19568182011213"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p55691220202114"><a name="zh-cn_topic_0000001930317932_p55691220202114"></a><a name="zh-cn_topic_0000001930317932_p55691220202114"></a>--tar arg1 [arg2 ...]</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p25691320162114"><a name="zh-cn_topic_0000001930317932_p25691320162114"></a><a name="zh-cn_topic_0000001930317932_p25691320162114"></a>对软件包执行tar命令，使用tar后面的参数作为命令的参数。例如执行<strong id="zh-cn_topic_0000001930317932_b656982016214"><a name="zh-cn_topic_0000001930317932_b656982016214"></a><a name="zh-cn_topic_0000001930317932_b656982016214"></a>--tar xvf</strong>命令，解压run安装包的内容到当前目录。在未来某个版本将废弃该参数。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row156942092116"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p75697203214"><a name="zh-cn_topic_0000001930317932_p75697203214"></a><a name="zh-cn_topic_0000001930317932_p75697203214"></a>--install</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1357015208213"><a name="zh-cn_topic_0000001930317932_p1357015208213"></a><a name="zh-cn_topic_0000001930317932_p1357015208213"></a>安装软件包。可以指定安装路径--install-path=&lt;path&gt;，也可以不指定安装路径，直接安装到默认路径下。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19570122010217"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15570172014213"><a name="zh-cn_topic_0000001930317932_p15570172014213"></a><a name="zh-cn_topic_0000001930317932_p15570172014213"></a>--install-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="p369633161410"><a name="p369633161410"></a><a name="p369633161410"></a>指定安装路径。</p>
<a name="zh-cn_topic_0000001930317932_ul29611936455"></a><a name="zh-cn_topic_0000001930317932_ul29611936455"></a><ul id="zh-cn_topic_0000001930317932_ul29611936455"><li>必须使用绝对路径作为安装路径。</li><li>当环境上存在全局配置文件“ascend_docker_runtime_install.info”时，指定的安装路径必须与全局配置文件中保存的安装路径保持一致。</li><li>如用户想更换安装路径，需先卸载原路径下的<span id="zh-cn_topic_0000001930317932_ph1528115352583"><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a>Ascend Docker Runtime</span>软件包并确保全局配置文件“ascend_docker_runtime_install.info”已被删除。</li><li>若5.0.RC1版本之前的<span id="zh-cn_topic_0000001930317932_ph93781522588"><a name="zh-cn_topic_0000001930317932_ph93781522588"></a><a name="zh-cn_topic_0000001930317932_ph93781522588"></a>Ascend Docker Runtime</span>是通过ToolBox安装包安装的，则该文件不存在，不需要删除。</li><li>若不指定安装路径，将安装到默认路径<span class="filepath" id="zh-cn_topic_0000001930317932_filepath7570102017212"><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a>“/usr/local/Ascend”</span>。</li><li>若通过该参数指定了安装目录，运行用户需要对指定的安装路径有读写权限。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1444404185013"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1144584125019"><a name="zh-cn_topic_0000001930317932_p1144584125019"></a><a name="zh-cn_topic_0000001930317932_p1144584125019"></a>--install-scene=&lt;scene&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p153510190174"><a name="zh-cn_topic_0000001930317932_p153510190174"></a><a name="zh-cn_topic_0000001930317932_p153510190174"></a><span id="zh-cn_topic_0000001930317932_ph1308455195116"><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a>Ascend Docker Runtime</span>安装场景。<span id="zh-cn_topic_0000001930317932_ph1641213426170"><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a>默认值为</span><span id="zh-cn_topic_0000001930317932_ph8821719135318"><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a>docker，</span>取值说明如下。</p>
<a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><ul id="zh-cn_topic_0000001930317932_ul8352122811918"><li><span id="zh-cn_topic_0000001930317932_ph3371331161710"><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a>docker</span>：表示在<span id="zh-cn_topic_0000001930317932_ph1159416519530"><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a>Docker</span>（或<span id="zh-cn_topic_0000001930317932_ph5391475179"><a name="zh-cn_topic_0000001930317932_ph5391475179"></a><a name="zh-cn_topic_0000001930317932_ph5391475179"></a>K8s集成Docker</span>）场景安装。</li><li><span id="zh-cn_topic_0000001930317932_ph7743733115213"><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a>c</span><span id="zh-cn_topic_0000001930317932_ph1274373385212"><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a>ontainerd：表示在</span>Containerd（或K8s集成Containerd）场景安装。</li><li>isula：表示在iSula容器引擎场景下安装。</li></ul><p>--install-scene不能单独使用，必须和--install、--uninstall或--upgrade一起使用。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row16570162013216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1457092012114"><a name="zh-cn_topic_0000001930317932_p1457092012114"></a><a name="zh-cn_topic_0000001930317932_p1457092012114"></a>--uninstall</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p35701320182115"><a name="zh-cn_topic_0000001930317932_p35701320182115"></a><a name="zh-cn_topic_0000001930317932_p35701320182115"></a>卸载软件。如果安装时指定了安装路径，那么卸载时也需要指定安装路径，安装路径的参数为--install-path=&lt;path&gt;。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row757019209212"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p11570122092117"><a name="zh-cn_topic_0000001930317932_p11570122092117"></a><a name="zh-cn_topic_0000001930317932_p11570122092117"></a>--upgrade</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p5570720152111"><a name="zh-cn_topic_0000001930317932_p5570720152111"></a><a name="zh-cn_topic_0000001930317932_p5570720152111"></a>升级软件。如果安装时指定了安装路径，那么升级时也需要指定安装路径，安装路径的参数为--install-path=&lt;path&gt;。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row106534178110"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p17661618012"><a name="zh-cn_topic_0000001930317932_p17661618012"></a><a name="zh-cn_topic_0000001930317932_p17661618012"></a>--config-file-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p18661121811111"><a name="zh-cn_topic_0000001930317932_p18661121811111"></a><a name="zh-cn_topic_0000001930317932_p18661121811111"></a><span id="zh-cn_topic_0000001930317932_ph86621218919"><a name="zh-cn_topic_0000001930317932_ph86621218919"></a><a name="zh-cn_topic_0000001930317932_ph86621218919"></a>Docker</span>或<span id="zh-cn_topic_0000001930317932_ph196625181110"><a name="zh-cn_topic_0000001930317932_ph196625181110"></a><a name="zh-cn_topic_0000001930317932_ph196625181110"></a>Containerd</span>的配置文件路径。不指定该参数时默认使用以下路径。</p>
<a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><ul id="zh-cn_topic_0000001930317932_ul1666216181816"><li><span id="zh-cn_topic_0000001930317932_ph146627186110"><a name="zh-cn_topic_0000001930317932_ph146627186110"></a><a name="zh-cn_topic_0000001930317932_ph146627186110"></a>Docker</span>: /etc/docker/daemon.json</li><li><span id="zh-cn_topic_0000001930317932_ph4662118513"><a name="zh-cn_topic_0000001930317932_ph4662118513"></a><a name="zh-cn_topic_0000001930317932_ph4662118513"></a>Containerd</span>: /etc/containerd/config.toml</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1857082012216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p65701620122117"><a name="zh-cn_topic_0000001930317932_p65701620122117"></a><a name="zh-cn_topic_0000001930317932_p65701620122117"></a>--install-type=&lt;type&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><div class="p" id="zh-cn_topic_0000001930317932_p155774343616"><a name="zh-cn_topic_0000001930317932_p155774343616"></a><a name="zh-cn_topic_0000001930317932_p155774343616"></a>仅支持在以下产品安装或升级<span id="zh-cn_topic_0000001930317932_ph1796213135594"><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a>Ascend Docker Runtime</span>时使用该参数：<a name="zh-cn_topic_0000001930317932_ul760551653710"></a><a name="zh-cn_topic_0000001930317932_ul760551653710"></a><ul id="zh-cn_topic_0000001930317932_ul760551653710"><li><span id="zh-cn_topic_0000001930317932_ph87811154145311"><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a>Atlas 200 AI加速模块（RC场景）</span></li><li><span id="zh-cn_topic_0000001930317932_ph1851111042012"><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a>Atlas 200I A2 加速模块</span>（RC场景）</li><li><span id="zh-cn_topic_0000001930317932_ph225916251208"><a name="zh-cn_topic_0000001930317932_ph225916251208"></a><a name="zh-cn_topic_0000001930317932_ph225916251208"></a>Atlas 200I DK A2 开发者套件</span></li><li><span id="zh-cn_topic_0000001930317932_ph271718714435"><a name="zh-cn_topic_0000001930317932_ph271718714435"></a><a name="zh-cn_topic_0000001930317932_ph271718714435"></a>Atlas 200I SoC A1 核心板</span></li><li><span id="zh-cn_topic_0000001930317932_ph12573124613552"><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a>Atlas 500 智能小站（型号 3000）</span></li><li><span id="zh-cn_topic_0000001930317932_ph11710328131520"><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a>Atlas 500 A2 智能小站</span></li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001930317932_p157201431201014"><a name="zh-cn_topic_0000001930317932_p157201431201014"></a><a name="zh-cn_topic_0000001930317932_p157201431201014"></a>该参数用于设置<span id="zh-cn_topic_0000001930317932_ph118353873517"><a name="zh-cn_topic_0000001930317932_ph118353873517"></a><a name="zh-cn_topic_0000001930317932_ph118353873517"></a>Ascend Docker Runtime</span>的默认挂载内容，且需要配合“--install”一起使用，格式为--install --install-type=&lt;type&gt;。&lt;type&gt;可选值为：<a name="zh-cn_topic_0000001930317932_ul848511715115"></a><a name="zh-cn_topic_0000001930317932_ul848511715115"></a><ul id="zh-cn_topic_0000001930317932_ul848511715115"><li>A200</li><li>A200ISoC</li><li>A200IA2（支持<span id="zh-cn_topic_0000001930317932_ph1323354011201"><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a>Atlas 200I A2 加速模块</span>（RC场景）和<span id="zh-cn_topic_0000001930317932_ph192331940102018"><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a>Atlas 200I DK A2 开发者套件</span>）</li><li>A500</li><li>A500A2</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row14570162052115"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1857042012112"><a name="zh-cn_topic_0000001930317932_p1857042012112"></a><a name="zh-cn_topic_0000001930317932_p1857042012112"></a>--ce=&lt;ce&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><ul id="zh-cn_topic_0000001930317932_ul4752351238"><li>仅在使用<span id="zh-cn_topic_0000001930317932_ph137882109239"><a name="zh-cn_topic_0000001930317932_ph137882109239"></a><a name="zh-cn_topic_0000001930317932_ph137882109239"></a>iSula</span>启动容器时需要指定该参数，参数值为isula。并且需要配合--install或者--uninstall一起使用，不能单独使用。</li><li>不支持和--install-scene同时使用。建议使用--install-scene替代--ce参数。后续--ce会废弃。</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1633572102619"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p733611211268"><a name="zh-cn_topic_0000001930317932_p733611211268"></a><a name="zh-cn_topic_0000001930317932_p733611211268"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p83361215264"><a name="zh-cn_topic_0000001930317932_p83361215264"></a><a name="zh-cn_topic_0000001930317932_p83361215264"></a>查询<span id="zh-cn_topic_0000001930317932_ph7723132765210"><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a>Ascend Docker Runtime</span>版本。</p>
</td>
</tr>
</tbody>
</table>
