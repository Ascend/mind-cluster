# Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426341"></a>

- 使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须在计算节点安装Ascend Device Plugin。
- 仅使用容器化支持和资源监测的用户，可以不安装Ascend Device Plugin，请直接跳过本章节。
- 安装Ascend Device Plugin之前，需要先安装Ascend Docker Runtime。Ascend Device Plugin会自动感知节点中是否已经安装Ascend Docker Runtime。

## 使用约束<a name="section1362795652416"></a>

在安装Ascend Device Plugin前，需要提前了解相关约束，具体说明请参见[表1](#table113813012140)。

**表 1**  约束说明

<a name="table113813012140"></a>
<table><thead align="left"><tr id="row193815031414"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p13383051411"><a name="p13383051411"></a><a name="p13383051411"></a>约束场景</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p73814015146"><a name="p73814015146"></a><a name="p73814015146"></a>约束说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row738802142"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p13388019145"><a name="p13388019145"></a><a name="p13388019145"></a>NPU驱动</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p73819019145"><a name="p73819019145"></a><a name="p73819019145"></a><span id="ph11461134318147"><a name="ph11461134318147"></a><a name="ph11461134318147"></a>Ascend Device Plugin</span>会周期性调用NPU驱动的相关接口。如果要升级驱动，请先停止业务任务，再停止<span id="ph1546116433149"><a name="ph1546116433149"></a><a name="ph1546116433149"></a>Ascend Device Plugin</span>容器服务。</p>
</td>
</tr>
<tr id="row5531349229"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p6691413112218"><a name="p6691413112218"></a><a name="p6691413112218"></a>配合<span id="ph14695135229"><a name="ph14695135229"></a><a name="ph14695135229"></a>Ascend Docker Runtime</span>使用</p>
<p id="p6920163951110"><a name="p6920163951110"></a><a name="p6920163951110"></a></p>
<p id="p1920153951110"><a name="p1920153951110"></a><a name="p1920153951110"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p175159351335"><a name="p175159351335"></a><a name="p175159351335"></a>组件安装顺序要求如下：</p>
<p id="p1745811135313"><a name="p1745811135313"></a><a name="p1745811135313"></a><span id="ph197011318223"><a name="ph197011318223"></a><a name="ph197011318223"></a>Ascend Device Plugin</span>容器化运行时会自动识别是否安装了<span id="ph18701713102214"><a name="ph18701713102214"></a><a name="ph18701713102214"></a>Ascend Docker Runtime</span>，需要优先安装<span id="ph167081312219"><a name="ph167081312219"></a><a name="ph167081312219"></a>Ascend Docker Runtime</span>后<span id="ph207041311225"><a name="ph207041311225"></a><a name="ph207041311225"></a>Ascend Device Plugin</span>才能正确识别<span id="ph11701013152214"><a name="ph11701013152214"></a><a name="ph11701013152214"></a>Ascend Docker Runtime</span>的安装情况。</p>
<p id="p019819298377"><a name="p019819298377"></a><a name="p019819298377"></a><span id="ph06381714397"><a name="ph06381714397"></a><a name="ph06381714397"></a>Ascend Device Plugin</span>若部署在<span id="ph66321793918"><a name="ph66321793918"></a><a name="ph66321793918"></a>Atlas 200I SoC A1 核心板</span>上，无需安装<span id="ph116721035125612"><a name="ph116721035125612"></a><a name="ph116721035125612"></a>Ascend Docker Runtime</span>。</p>
</td>
</tr>
<tr id="row1648094416218"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p9484175210212"><a name="p9484175210212"></a><a name="p9484175210212"></a>组件版本要求如下：</p>
<p id="p44813447213"><a name="p44813447213"></a><a name="p44813447213"></a>该功能要求<span id="ph196135501025"><a name="ph196135501025"></a><a name="ph196135501025"></a>Ascend Docker Runtime</span>与<span id="ph1161319502212"><a name="ph1161319502212"></a><a name="ph1161319502212"></a>Ascend Device Plugin</span>版本保持一致且需要为5.0.RC1及以上版本，安装或卸载<span id="ph1361319501123"><a name="ph1361319501123"></a><a name="ph1361319501123"></a>Ascend Docker Runtime</span>之后需要重启容器引擎才能使<span id="ph11613850625"><a name="ph11613850625"></a><a name="ph11613850625"></a>Ascend Device Plugin</span>正确识别。</p>
</td>
</tr>
<tr id="row1449218752210"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p14704133226"><a name="p14704133226"></a><a name="p14704133226"></a>以下2种场景不支持<span id="ph1371171332212"><a name="ph1371171332212"></a><a name="ph1371171332212"></a>Ascend Device Plugin</span>和<span id="ph071513132214"><a name="ph071513132214"></a><a name="ph071513132214"></a>Ascend Docker Runtime</span>配合使用。<a name="ul1771141362211"></a><a name="ul1771141362211"></a><ul id="ul1771141362211"><li>混插场景。</li><li><span id="ph1471111314226"><a name="ph1471111314226"></a><a name="ph1471111314226"></a>Atlas 200I SoC A1 核心板</span>。</li></ul>
</div>
</td>
</tr>
<tr id="row5381205148"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p16384020141"><a name="p16384020141"></a><a name="p16384020141"></a>DCMI动态库</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p67821743113213"><a name="p67821743113213"></a><a name="p67821743113213"></a>DCMI动态库目录权限要求如下：</p>
<p id="p1238120191413"><a name="p1238120191413"></a><a name="p1238120191413"></a><span id="ph285261461515"><a name="ph285261461515"></a><a name="ph285261461515"></a>Ascend Device Plugin</span>调用的DCMI动态库及其所有父目录，需要满足属主为root，其他属主程序无法运行；同时，这些文件及其目录需满足group和other不具备写权限。</p>
</td>
</tr>
<tr id="row1138160191419"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1138180101418"><a name="p1138180101418"></a><a name="p1138180101418"></a>DCMI动态库路径深度必须小于20。</p>
</td>
</tr>
<tr id="row338407145"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1739170161413"><a name="p1739170161413"></a><a name="p1739170161413"></a>如果通过设置LD_LIBRARY_PATH设置动态库路径，LD_LIBRARY_PATH环境变量总长度不能超过1024。</p>
</td>
</tr>
<tr id="row11391707149"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p133919013143"><a name="p133919013143"></a><a name="p133919013143"></a><span id="ph1078193611515"><a name="ph1078193611515"></a><a name="ph1078193611515"></a>Atlas 200I SoC A1 核心板</span></p>
<p id="p1918223205014"><a name="p1918223205014"></a><a name="p1918223205014"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p786843510309"><a name="p786843510309"></a><a name="p786843510309"></a><span id="ph1480005781518"><a name="ph1480005781518"></a><a name="ph1480005781518"></a>Atlas 200I SoC A1 核心板</span>节点上如果使用容器化部署<span id="ph080185715158"><a name="ph080185715158"></a><a name="ph080185715158"></a>Ascend Device Plugin</span>，需要配置多容器共享模式，具体请参考<span id="ph3957123242310"><a name="ph3957123242310"></a><a name="ph3957123242310"></a>《Atlas 200I SoC A1 核心板 25.0.RC1 NPU驱动和固件安装指南》中“<a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100468901/55e9d968" target="_blank" rel="noopener noreferrer">容器内运行</a>”章节</span>。</p>
</td>
</tr>
<tr id="row4248144116153"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p5840775161"><a name="p5840775161"></a><a name="p5840775161"></a><span id="ph697712515161"><a name="ph697712515161"></a><a name="ph697712515161"></a>Atlas 200I SoC A1 核心板</span>使用<span id="ph99771752169"><a name="ph99771752169"></a><a name="ph99771752169"></a>Ascend Device Plugin</span>组件，需要遵循以下配套关系：<a name="ul2977251161"></a><a name="ul2977251161"></a><ul id="ul2977251161"><li>5.0.RC2版本的<span id="ph49779571614"><a name="ph49779571614"></a><a name="ph49779571614"></a>Ascend Device Plugin</span>需要配合<span id="ph5977135101614"><a name="ph5977135101614"></a><a name="ph5977135101614"></a>Atlas 200I SoC A1 核心板</span>的23.0.RC2及其之后的驱动一起使用。</li><li>5.0.RC2之前版本的<span id="ph59771512164"><a name="ph59771512164"></a><a name="ph59771512164"></a>Ascend Device Plugin</span>只能和<span id="ph1977115181612"><a name="ph1977115181612"></a><a name="ph1977115181612"></a>Atlas 200I SoC A1 核心板</span>的23.0.RC2之前的驱动一起使用。</li></ul>
</div>
</td>
</tr>
<tr id="row14538194431511"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p45382449151"><a name="p45382449151"></a><a name="p45382449151"></a>虚拟机场景</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p8538144420153"><a name="p8538144420153"></a><a name="p8538144420153"></a>如果在虚拟机场景下部署<span id="ph142915347164"><a name="ph142915347164"></a><a name="ph142915347164"></a>Ascend Device Plugin</span>，需要在<span id="ph0429634121617"><a name="ph0429634121617"></a><a name="ph0429634121617"></a>Ascend Device Plugin</span>的镜像中安装systemd，推荐在Dockerfile中加入<strong id="b93339419563"><a name="b93339419563"></a><a name="b93339419563"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong>命令进行安装。</p>
</td>
</tr>
<tr id="row1150514563377"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p450675616371"><a name="p450675616371"></a><a name="p450675616371"></a>重启场景</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p105070566371"><a name="p105070566371"></a><a name="p105070566371"></a>若用户在安装<span id="ph444301153912"><a name="ph444301153912"></a><a name="ph444301153912"></a>Ascend Device Plugin</span>后，又重新修改了NPU的基础信息，例如修改了device ip，则需要重启<span id="ph52417305424"><a name="ph52417305424"></a><a name="ph52417305424"></a>Ascend Device Plugin</span>，否则<span id="ph23611038174213"><a name="ph23611038174213"></a><a name="ph23611038174213"></a>Ascend Device Plugin</span>不能正确识别NPU的相关信息。</p>
</td>
</tr>
</tbody>
</table>

## 操作步骤<a name="section71204451253"></a>

1. 以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```shell
    docker images | grep k8sdeviceplugin
    ```

    回显示例如下：

    ```ColdFusion
    ascend-k8sdeviceplugin               v26.0.0              29eec79eb693        About an hour ago   105MB
    ```

    - 是，执行[步骤2](#zh-cn_topic_0000001497364849_li922154411117)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="zh-cn_topic_0000001497364849_li922154411117"></a>将Ascend Device Plugin软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。请注意此处需使用适配具体处理器型号的YAML文件，并且为了避免自动识别Ascend Docker Runtime功能出现异常，请勿修改YAML文件中DaemonSet.metadata.name字段，详见下表。

    **表 2** Ascend Device Plugin的YAML文件列表

    <a name="zh-cn_topic_0000001497364849_table58619457211"></a>

    | YAML文件列表                                                  | 说明                                                                                                  |
    |-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
    | device-plugin-310-v<i>\{version\}</i>.yaml                | 推理服务器（插Atlas 300I 推理卡）上不使用Volcano的配置文件。                                                             |
    | device-plugin-310-volcano-v<i>\{version\}</i>.yaml        | 推理服务器（插Atlas 300I 推理卡）上使用Volcano的配置文件。                                                              |
    | device-plugin-310P-1usoc-v<i>\{version\}</i>.yaml         | Atlas 200I SoC A1 核心板上不使用Volcano的配置文件。                                                              |
    | device-plugin-310P-1usoc-volcano-v<i>\{version\}</i>.yaml | Atlas 200I SoC A1 核心板上使用Volcano的配置文件。                                                               |
    | device-plugin-310P-v<i>\{version\}</i>.yaml               | 除了Atlas 200I SoC A1 核心板之外的Atlas 推理系列产品上不使用Volcano的配置文件。                                             |
    | device-plugin-310P-volcano-v<i>\{version\}</i>.yaml       | 除了Atlas 200I SoC A1 核心板之外的Atlas 推理系列产品上使用Volcano的配置文件。                                              |
    | device-plugin-910-v<i>\{version\}</i>.yaml                | Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件上不使用Volcano的配置文件。 | 
    | device-plugin-volcano-v<i>\{version\}</i>.yaml            | Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件上使用Volcano的配置文件。  |
    | device-plugin-npu-v<i>\{version\}</i>.yaml                | Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD上不使用Volcano的配置文件。                                  |    
    | device-plugin-npu-volcano-v<i>\{version\}</i>.yaml        | Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD上使用Volcano的配置文件。                                   |    

3. 如不修改组件启动参数，可跳过本步骤。否则，根据实际情况修改Ascend Device Plugin的启动参数。启动参数请参见[表3](#table1064314568229)，可执行<b>./device-plugin -h</b>查看参数说明。
    - 在Atlas 200I SoC A1 核心板节点上，修改启动脚本“run\_for\_310P\_1usoc.sh”中Ascend Device Plugin的启动参数。修改完后需在所有Atlas 200I SoC A1 核心板节点上重新制作镜像，或者将本节点镜像重新制作后分发到其余所有Atlas 200I SoC A1 核心板节点。

        >[!NOTE] 
        >如果不使用Volcano作为调度器，在启动Ascend Device Plugin的时候，需要修改“run\_for\_310P\_1usoc.sh”中Ascend Device Plugin的启动参数，将“-volcanoType”参数设置为false。

    - 其他类型节点，修改对应启动YAML文件中Ascend Device Plugin的启动参数。

4. （可选）使用**断点续训**（包括进程级恢复）或**弹性训练**时，根据需要使用的故障处理模式，修改Ascend Device Plugin组件的启动YAML。

    <pre codetype="yaml">
    ...
          containers:
          - image: ascend-k8sdeviceplugin:v26.0.0
            name: device-plugin-01
            resources:
              requests:
                memory: 500Mi
                cpu: 500m
              limits:
                memory: 500Mi
                cpu: 500m
            command: [ "/bin/bash", "-c", "--"]
            args: [ "device-plugin  
                     -useAscendDocker=true 
                     <strong>-volcanoType=true                    # 重调度场景下必须使用Volcano
                     -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealthy变为healthy后，不会自动加入到可调度资源池中；关闭自动纳管，当芯片参数面网络故障恢复后，不会自动加入到可调度资源池中。该特性仅适用于Atlas 训练系列产品
                     -listWatchPeriod=5                   # 设置健康状态检查周期，范围[3,1800]，单位为秒
                     -hotReset=2 # 使用进程级恢复时，请将hotReset参数值设置为2，开启离线恢复模式</strong>
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                     -logLevel=0" ]
            securityContext:
              privileged: true
              readOnlyRootFilesystem: true
    ...</pre>

5. （可选）使用推理卡故障恢复时，需要配置热复位功能。

    <pre codetype="yaml">
          containers:
          - image: ascend-k8sdeviceplugin:v26.0.0
            name: device-plugin-01
            resources:
              requests:
                memory: 500Mi
                cpu: 500m
              limits:
                memory: 500Mi
                cpu: 500m
            command: [ "/bin/bash", "-c", "--"]
            args: [ "device-plugin  
    ...
                     <strong>-hotReset=0 # 使用推理卡故障恢复时，开启热复位功能</strong>
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log 
                     -logLevel=0" ]
    ...</pre>

6. （可选）如需更改kubelet的默认端口，则需要修改Ascend Device Plugin组件的启动YAML。示例如下。

    <pre codetype="yaml">
      env:
         - name: NODE_NAME
           valueFrom:
             fieldRef:
               fieldPath: spec.nodeName
         - name: HOST_IP
           valueFrom:
             fieldRef:
               fieldPath: status.hostIP
         <strong>- name: KUBELET_PORT   # 通知Ascend Device Plugin组件当前节点kubelet默认端口号，若未自定义kubelet默认端口号则无需传入本字段
           value: "10251"</strong>      
    volumes:
       - name: device-plugin
         hostPath:
           path: /var/lib/kubelet/device-plugins
    ...</pre>

7. （可选）根据容器运行时类型，修改Ascend Device Plugin组件的启动YAML中的挂载配置。

    - 如果容器运行时为Docker，保留docker-sock和docker-dir挂载配置，示例如下：

        ```Yaml
        volumeMounts:
          ...
          - name: docker-sock
            mountPath: /run/docker.sock
            readOnly: true
          - name: docker-dir
            mountPath: /run/docker
            readOnly: true
          - name: containerd
            mountPath: /run/containerd
            readOnly: true
        volumes:
          ...
          - name: docker-sock
            hostPath:
              path: /run/docker.sock
          - name: docker-dir
            hostPath:
              path: /run/docker
          - name: containerd
            hostPath:
              path: /run/containerd
        ```

    - 如果容器运行时为containerd，删除docker-sock和docker-dir挂载配置，保留containerd挂载配置。示例如下：

        ```Yaml
        volumeMounts:
            ...
            - name: containerd
            mountPath: /run/containerd
            readOnly: true
        volumes:
            ...
            - name: containerd
            hostPath:
                path: /run/containerd
        ```

    >[!NOTE] 
    >- 如果docker.sock文件路径不是/run/docker.sock，请在volumes中修改为实际路径，不支持使用符号链接。
    >- 如果docker目录不是/var/run/docker，请在volumes中修改为实际路径，不支持使用符号链接。
    >- 如果containerd目录不是/run/containerd，请在volumes中修改为实际路径，不支持使用符号链接。

8. 在K8s管理节点上各YAML对应路径下执行以下命令，启动Ascend Device Plugin。

    - K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件的节点（配合Volcano使用，支持虚拟化实例，YAML默认开启静态虚拟化）。

        ```shell
        kubectl apply -f device-plugin-volcano-v{version}.yaml
        ```

    - K8s集群中存在使用Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas A3 训练系列产品或Atlas 800I A2 推理服务器、A200I A2 Box 异构组件的节点（Ascend Device Plugin独立工作，不配合Volcano使用）。

        ```shell
        kubectl apply -f device-plugin-910-v{version}.yaml
        ```

    - K8s集群中存在使用推理服务器（插Atlas 300I 推理卡）的节点（使用Volcano调度器）。

        ```shell
        kubectl apply -f device-plugin-310-volcano-v{version}.yaml
        ```

    - K8s集群中存在使用推理服务器（插Atlas 300I 推理卡）的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```shell
        kubectl apply -f device-plugin-310-v{version}.yaml
        ```

    - K8s集群中存在使用Atlas 推理系列产品的节点（使用Volcano调度器，支持虚拟化实例，YAML默认开启静态虚拟化）。

        ```shell
        kubectl apply -f device-plugin-310P-volcano-v{version}.yaml
        ```

    - K8s集群中存在使用Atlas 推理系列产品的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```shell
        kubectl apply -f device-plugin-310P-v{version}.yaml
        ```

    - K8s集群中存在使用Atlas 200I SoC A1 核心板的节点（使用Volcano调度器）。

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-volcano-v{version}.yaml
        ```

    - K8s集群中存在使用Atlas 200I SoC A1 核心板的节点（Ascend Device Plugin独立工作，不使用Volcano调度器）。

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-v{version}.yaml
        ```

    >[!NOTE]
    >如果K8s集群使用了多种类型的昇腾AI处理器，请分别执行对应命令。

    启动示例如下：

    ```ColdFusion
    serviceaccount/ascend-device-plugin-sa created
    clusterrole.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-role created
    clusterrolebinding.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-rolebinding created
    daemonset.apps/ascend-device-plugin-daemonset created
    ```

9. 在K8s管理节点执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n kube-system
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                                        READY   STATUS    RESTARTS   AGE
    ...
    ascend-device-plugin-daemonset-d5ctz  1/1   Running   0        11s
    ...
    ```

>[!NOTE]
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

## 参数说明<a name="section479917441223"></a>

**表 3** Ascend Device Plugin启动参数

<a name="table1064314568229"></a>

|参数|类型|默认值|说明|
|--|--|--|--|
|-fdFlag|bool|false|边缘场景标志，是否使用FusionDirector系统来管理设备。<ul><li>true：使用FusionDirector。</li><li>false：不使用FusionDirector。</li></ul>|
|-shareDevCount|uint|1|共享设备特性开关，取值范围为1~100。<ul><li>默认值为1，代表不开启共享设备；取值为2~100，表示单颗芯片虚拟化出来的共享设备个数。</li><li>当开启软切分功能，即-softShareDevConfigDir不为空时，该参数取值必须为100。</li></ul><p>支持以下设备，其余设备该参数无效，不影响组件正常启动。</p><ul><li>Atlas 500 A2 智能小站</li><li>Atlas 200I A2 加速模块</li><li>Atlas 200I DK A2 开发者套件</li><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li></ul><p>若用户使用的是以上支持的Atlas 推理系列产品，需要注意以下问题：</p><ul><li>不支持在使用静态vNPU调度、动态vNPU调度、推理卡故障恢复和推理卡故障重调度等特性下使用共享设备功能。</li><li>单任务的请求资源数必须为1，不支持分配多芯片和跨芯片使用的场景。</li><li>依赖驱动开启共享模式，设置device-share为true，详细操作步骤和说明请参见《Atlas 中心推理卡 25.5.0 npu-smi 命令参考》中的“[设置指定设备的指定芯片的容器共享模式](https://support.huawei.com/enterprise/zh/doc/EDOC1100540373/af78d7e5)”章节。</li></ul>|
|-edgeLogFile|string|/var/alog/AtlasEdge_log/devicePlugin.log|边缘场景日志文件。fdFlag设置为true时生效。<p>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。</p>|
|-useAscendDocker|bool|true|默认为true，容器引擎是否使用Ascend Docker Runtime。开启K8s的CPU绑核功能时，需要卸载Ascend Docker Runtime并重启容器引擎。取值说明如下：<ul><li>true：使用Ascend Docker Runtime。</li><li>false：不使用Ascend Docker Runtime。</li></ul><p>MindCluster 5.0.RC1及以上版本只支持自动获取运行模式，不接受指定。</p>|
|-use310PMixedInsert|bool|false|是否使用混插模式。<ul><li>true：使用混插模式。</li><li>false：不使用混插模式。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>仅支持服务器混插Atlas 300I Pro 推理卡、Atlas 300V 视频解析卡、Atlas 300V Pro 视频解析卡。</li><li>服务器混插模式下不支持Volcano调度模式。</li><li>服务器混插模式不支持虚拟化实例。</li><li>服务器混插模式不支持故障重调度场景。</li><li>服务器混插模式不支持Ascend Docker Runtime。</li><li>非混插模式下，上报给K8s资源名称不变。<ul><li>非混插模式上报的资源名称格式为huawei.com/Ascend310P。</li><li>混插模式上报的资源名称格式为：huawei.com/Ascend310P-V、huawei.com/Ascend310P-VPro和huawei.com/Ascend310P-IPro。</li></ul></li></ul></div></div>|
|-volcanoType|bool|false|是否使用Volcano进行调度，当前已支持Atlas 训练系列产品、Atlas A2 训练系列产品、Atlas 推理系列产品和推理服务器（插Atlas 300I 推理卡）芯片。<ul><li>true：使用Volcano。</li><li>false：不使用Volcano。</li></ul>|
|-presetVirtualDevice|bool|true|虚拟化功能开关。<ul><li>设置为true时，表示使用静态虚拟化。</li><li>设置为false时，表示使用动态虚拟化。需要同步开启Volcano，即设置-volcanoType参数为true。</li></ul>|
|-version|bool|false|是否查看当前Ascend Device Plugin的版本号。<ul><li>true：查询。</li><li>false：不查询。</li></ul>|
|-listWatchPeriod|int|5|<p>设置健康状态检查周期，取值范围为[3,1800]，单位为秒。</p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>每个周期内会进行如下检查，并将检查结果写入ConfigMap中。</p><ul><li>如果设备信息没有变化且距离上次更新ConfigMap未超过5min，则不会更新ConfigMap。</li><li>如果距离上次更新ConfigMap超过5min，则无论设备信息是否发生变化，都会更新ConfigMap。</li></ul></div></div>|
|-autoStowing|bool|true|是否自动纳管已修复设备，volcanoType为true时生效。<ul><li>true：自动纳管。</li><li>false：不会自动纳管。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>设备故障后，会自动从K8s里面隔离。如果设备恢复正常，默认会自动加入K8s集群资源池。如果设备不稳定，可以设置为false，此时需要手动纳管。</p><ul><li>用户可以使用以下命令，将健康状态由unhealthy恢复为healthy的芯片重新放入资源池。<p>kubectl label nodes <i>node_name</i> huawei.com/Ascend910-Recover-</p><p>当使用 Ascend 950 系列产品时，需使用： </p><p>kubectl label nodes <i>node_name</i> huawei.com/NPU-Recover-</p></li><li>用户可以使用以下命令，将参数面网络健康状态由unhealthy恢复为healthy的芯片重新放入资源池。<p>kubectl label nodes <i>node_name</i> huawei.com/Ascend910-NetworkRecover-</p><p>当使用 Ascend 950 系列产品时，需使用： </p><p>kubectl label nodes <i>node_name</i> huawei.com/NPU-NetworkRecover-</p></li></ul>|
|-logLevel|int|0|日志级别：<ul><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul></div></div>|
|-maxAge|int|7|日志备份时间限制，取值范围为7~700，单位为天。|
|-logFile|string|/var/log/mindx-dl/devicePlugin/devicePlugin.log|非边缘场景日志文件。fdFlag设置为false时生效。<p>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：devicePlugin-触发转储的时间.log，如：devicePlugin-2023-10-07T03-38-24.402.log。</p>|
|-hotReset|int|-1|设备热复位功能参数。开启此功能，芯片发生故障后，Ascend Device Plugin会进行热复位操作，使芯片恢复健康。<ul><li>-1：关闭芯片复位功能</li><li>0：开启推理设备复位功能</li><li>1：开启训练设备在线复位功能</li><li>2：开启训练/推理设备离线复位功能</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>取值为1对应的功能已经日落，请配置其他取值。</p></div></div><p>该参数支持的训练设备：</p><ul><li>Atlas 800 训练服务器（型号 9000）（NPU满配）</li><li>Atlas 800 训练服务器（型号 9010）（NPU满配）</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD（型号 9000）</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A2 PoD 集群基础单元</li><li>Atlas 900 A3 SuperPoD 超节点</li><li>Atlas 800T A3 超节点服务器</li></ul><p>该参数支持的推理设备：</p><ul><li>Atlas 300I Pro 推理卡</li><li>Atlas 300V 视频解析卡</li><li>Atlas 300V Pro 视频解析卡</li><li>Atlas 300I Duo 推理卡</li><li>Atlas 300I 推理卡（型号 3000）（整卡）</li><li>Atlas 300I 推理卡（型号 3010）</li><li>Atlas 800I A2 推理服务器</li><li>A200I A2 Box 异构组件</li><li>Atlas 800I A3 超节点服务器</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><ul><li>针对Atlas 300I Duo 推理卡形态硬件，仅支持按卡复位，即两颗芯片会同时复位。</li><li>Atlas 800I A2 推理服务器存在以下两种热复位方式，一台Atlas 800I A2 推理服务器只能使用一种热复位方式，由集群调度组件自动识别使用哪种热复位方式。<ul><li>方式一：若设备上不存在HCCS环，执行推理任务中，当NPU出现故障，Ascend Device Plugin等待该NPU空闲后，对该NPU进行复位操作。</li><li>方式二：若设备上存在HCCS环，执行推理任务中，当服务器出现一个或多个故障NPU，Ascend Device Plugin等待环上的NPU全部空闲后，一次性复位环上所有的NPU。</li></ul></li></ul></div></div>|
|-linkdownTimeout|int|30|网络linkdown超时时间，单位秒，取值范围为1~30。<p>该参数取值建议与用户在训练脚本中配置的HCCL_RDMA_TIMEOUT时间一致。如果是多任务，建议设置为多任务中HCCL_RDMA_TIMEOUT的最小值。</p>|
|-enableSlowNode|bool|false|是否启用慢节点检测（劣化诊断）功能。<ul><li>true：开启。</li><li>false：关闭。</li></ul><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>关于劣化诊断的详细说明请参见《iMaster CCAE 产品文档》的“[劣化诊断](https://support.huawei.com/hedex/hdx.do?docid=EDOC1100445519&amp;id=ZH-CN_TOPIC_0000002147436540)”章节。</p></div></div>|
|-dealWatchHandler|bool|false|当informer链接因异常结束时，是否需要刷新本地的Pod informer缓存。<ul><li>true：刷新Pod informer缓存。</li><li>false：不刷新Pod informer缓存。</li></ul>|
|-checkCachedPods|bool|true|是否定期检查缓存中的Pod。默认取值为true，当缓存中的Pod超过1小时没有被更新，Ascend Device Plugin将会主动请求api-server查看Pod情况。<ul><li>true：检查。</li><li>false：不检查。</li></ul>|
|-maxBackups|int|30|转储后日志文件保留个数上限，取值范围为1~30，单位为个。|
|-thirdPartyScanDelay|int|300|<p>Ascend Device Plugin组件启动重新扫描的等待时长。</p><p>Ascend Device Plugin自动复位芯片失败后，会将失败信息写到节点annotation上，三方平台可以根据该信息复位失败的芯片。Ascend Device Plugin组件根据本参数设置的等待时长，等待一段时间后，重新扫描设备。</p><p>仅Atlas 800T A3 超节点服务器支持使用本参数。</p><p>单位：秒。</p>|
|-deviceResetTimeout|int|60|组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600。<ul><li>Atlas A2 训练系列产品、Atlas 800I A2 推理服务器、A200I A2 Box 异构组件：建议配置为150秒。</li><li>Atlas A3 训练系列产品、A200T A3 Box8 超节点服务器、Atlas 800I A3 超节点服务器：建议配置为360秒。</li><li>Atlas 350 标卡、Atlas 850 系列硬件产品、Atlas 950 SuperPoD：建议配置为600秒。</li></ul>|
|-softShareDevConfigDir|string|""|软切分虚拟化场景配置目录。该配置目录需要在安装Ascend Device Plugin之前在根目录下手动创建。使用软切分功能时，需要配置该参数。|
|-useSingleDieMode|bool|false|Atlas A3 推理系列产品是否开启单die直通模式。<ul><li>true：开启单die直通模式。</li><li>false：关闭单die直通模式。</li></ul>使用软切分虚拟化功能时，该参数必须配置为true。|
|-h或者-help|无|无|显示帮助信息。|
