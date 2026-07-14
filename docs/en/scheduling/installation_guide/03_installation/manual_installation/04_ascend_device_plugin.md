# Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426341"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:14:02.062Z pushedAt=2026-06-09T06:22:06.921Z -->

- Users who use full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, resumable training, elastic training, inference card failure recovery, or rescheduling upon inference card faults must install Ascend Device Plugin on compute nodes.
- Users who only use containerization support and resource monitoring do not need to install Ascend Device Plugin and can skip this chapter directly.
- Before installing Ascend Device Plugin, install Ascend Docker Runtime first. Ascend Device Plugin automatically detects whether Ascend Docker Runtime has been installed on a node.

## Constraints<a name="section1362795652416"></a>

Before installing Ascend Device Plugin, you need to understand the relevant constraints. For details, see <a href="#table113813012140">Table 1</a>.

**Table 1**  Constraints

<a name="table113813012140"></a>
<table><thead align="left"><tr id="row193815031414"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p13383051411"><a name="p13383051411"></a><a name="p13383051411"></a>Constraint Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p73814015146"><a name="p73814015146"></a><a name="p73814015146"></a>Constraint Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row738802142"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p13388019145"><a name="p13388019145"></a><a name="p13388019145"></a>NPU Driver</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p73819019145"><a name="p73819019145"></a><a name="p73819019145"></a><span id="ph11461134318147"><a name="ph11461134318147"></a><a name="ph11461134318147"></a>Ascend Device Plugin</span> periodically calls the relevant interfaces of the NPU driver. To upgrade the driver, stop the service tasks first, and then stop the <span id="ph1546116433149"><a name="ph1546116433149"></a><a name="ph1546116433149"></a>Ascend Device Plugin</span> container service.</p>
</td>
</tr>
<tr id="row5531349229"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p6691413112218"><a name="p6691413112218"></a><a name="p6691413112218"></a>Used with <span id="ph14695135229"><a name="ph14695135229"></a><a name="ph14695135229"></a>Ascend Docker Runtime</span></p>
<p id="p6920163951110"><a name="p6920163951110"></a><a name="p6920163951110"></a></p>
<p id="p1920153951110"><a name="p1920153951110"></a><a name="p1920153951110"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p175159351335"><a name="p175159351335"></a><a name="p175159351335"></a>The component installation sequence requirements are as follows:</p>
<p id="p1745811135313"><a name="p1745811135313"></a><a name="p1745811135313"></a>When running in image mode, <span id="ph197011318223"><a name="ph197011318223"></a><a name="ph197011318223"></a>Ascend Device Plugin</span> automatically identifies whether <span id="ph18701713102214"><a name="ph18701713102214"></a><a name="ph18701713102214"></a>Ascend Docker Runtime</span> is installed. <span id="ph167081312219"><a name="ph167081312219"></a><a name="ph167081312219"></a>Ascend Docker Runtime</span> must be installed first so that <span id="ph207041311225"><a name="ph207041311225"></a><a name="ph207041311225"></a>Ascend Device Plugin</span> can correctly identify the installation status of <span id="ph11701013152214"><a name="ph11701013152214"></a><a name="ph11701013152214"></a>Ascend Docker Runtime</span>.</p>
<p id="p019819298377"><a name="p019819298377"></a><a name="p019819298377"></a>If <span id="ph06381714397"><a name="ph06381714397"></a><a name="ph06381714397"></a>Ascend Device Plugin</span> is deployed on the <span id="ph66321793918"><a name="ph66321793918"></a><a name="ph66321793918"></a>Atlas 200I SoC A1 Core Board</span>, <span id="ph116721035125612"><a name="ph116721035125612"></a><a name="ph116721035125612"></a>Ascend Docker Runtime</span> is not required.</p>
</td>
</tr>
<tr id="row1648094416218"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p9484175210212"><a name="p9484175210212"></a><a name="p9484175210212"></a>The component version requirements are as follows:</p>
<p id="p44813447213"><a name="p44813447213"></a><a name="p44813447213"></a>This function requires that the versions of <span id="ph196135501025"><a name="ph196135501025"></a><a name="ph196135501025"></a>Ascend Docker Runtime</span> and <span id="ph1161319502212"><a name="ph1161319502212"></a><a name="ph1161319502212"></a>Ascend Device Plugin</span> be consistent and be 5.0.RC1 or later. After installing or uninstalling <span id="ph1361319501123"><a name="ph1361319501123"></a><a name="ph1361319501123"></a>Ascend Docker Runtime</span>, you need to restart the container engine for <span id="ph11613850625"><a name="ph11613850625"></a><a name="ph11613850625"></a>Ascend Device Plugin</span> to correctly identify it.</p>
</td>
</tr>
<tr id="row1449218752210"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p14704133226"><a name="p14704133226"></a><a name="p14704133226"></a><span id="ph1371171332212"><a name="ph1371171332212"></a><a name="ph1371171332212"></a>Ascend Device Plugin</span> and <span id="ph071513132214"><a name="ph071513132214"></a><a name="ph071513132214"></a>Ascend Docker Runtime</span> cannot be used together in the following two scenarios.<a name="ul1771141362211"></a><a name="ul1771141362211"></a><ul id="ul1771141362211"><li>Hybrid deployment scenarios.</li><li><span id="ph1471111314226"><a name="ph1471111314226"></a><a name="ph1471111314226"></a>Atlas 200I SoC A1 Core Board</span>.</li></ul>
</div>
</td>
</tr>
<tr id="row5381205148"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p16384020141"><a name="p16384020141"></a><a name="p16384020141"></a>DCMI Dynamic Library</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p67821743113213"><a name="p67821743113213"></a><a name="p67821743113213"></a>The directory permission requirements for the DCMI dynamic library are as follows:</p>
<p id="p1238120191413"><a name="p1238120191413"></a><a name="p1238120191413"></a>The DCMI dynamic library called by <span id="ph285261461515"><a name="ph285261461515"></a><a name="ph285261461515"></a>Ascend Device Plugin</span> and all its parent directories must be owned by root, and programs owned by other users cannot run. In addition, the group and other permissions for these files and directories must not include write permission.</p>
</td>
</tr>
<tr id="row1138160191419"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1138180101418"><a name="p1138180101418"></a><a name="p1138180101418"></a>The path depth of the DCMI dynamic library must be less than 20.</p>
</td>
</tr>
<tr id="row338407145"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p1739170161413"><a name="p1739170161413"></a><a name="p1739170161413"></a>If the dynamic library path is set by setting LD_LIBRARY_PATH, the total length of the LD_LIBRARY_PATH environment variable cannot exceed 1024.</p>
</td>
</tr>
<tr id="row11391707149"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p133919013143"><a name="p133919013143"></a><a name="p133919013143"></a><span id="ph1078193611515"><a name="ph1078193611515"></a><a name="ph1078193611515"></a>Atlas 200I SoC A1 Core Board</span></p>
<p id="p1918223205014"><a name="p1918223205014"></a><a name="p1918223205014"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p786843510309"><a name="p786843510309"></a><a name="p786843510309"></a>If <span id="ph1480005781518"><a name="ph1480005781518"></a><a name="ph1480005781518"></a>Ascend Device Plugin</span> is deployed in image mode on an <span id="ph080185715158"><a name="ph080185715158"></a><a name="ph080185715158"></a>Atlas 200I SoC A1 Core Board</span> node, you need to configure the multi-container sharing mode. For details, see the "Running in a Container" section in <span id="ph3957123242310"><a name="ph3957123242310"></a><a name="ph3957123242310"></a><em>Atlas 200I SoC A1 Core Board NPU Driver and Firmware Installation Guide</em><a href="https://support.huawei.com/enterprise/zh/doc/EDOC1100493510/55e9d968" target="_blank" rel="noopener noreferrer"></a></span>.</p>
</td>
</tr>
<tr id="row4248144116153"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p5840775161"><a name="p5840775161"></a><a name="p5840775161"></a>When using the <span id="ph697712515161"><a name="ph697712515161"></a><a name="ph697712515161"></a>Ascend Device Plugin</span> component with the <span id="ph99771752169"><a name="ph99771752169"></a><a name="ph99771752169"></a>Atlas 200I SoC A1 Core Board</span>, the following compatibility relationships must be observed:<a name="ul2977251161"></a><a name="ul2977251161"></a><ul id="ul2977251161"><li><span id="ph49779571614"><a name="ph49779571614"></a><a name="ph49779571614"></a>Ascend Device Plugin</span> version 5.0.RC2 must be used with the driver of <span id="ph5977135101614"><a name="ph5977135101614"></a><a name="ph5977135101614"></a>Atlas 200I SoC A1 Core Board</span> version 23.0.RC2 or later.</li><li><span id="ph59771512164"><a name="ph59771512164"></a><a name="ph59771512164"></a>Ascend Device Plugin</span> versions earlier than 5.0.RC2 can only be used with the driver of <span id="ph1977115181612"><a name="ph1977115181612"></a><a name="ph1977115181612"></a>Atlas 200I SoC A1 Core Board</span> earlier than 23.0.RC2.</li></ul>
</div>
</td>
</tr>
<tr id="row14538194431511"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p45382449151"><a name="p45382449151"></a><a name="p45382449151"></a>VM Scenario</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p8538144420153"><a name="p8538144420153"></a><a name="p8538144420153"></a>If <span id="ph142915347164"><a name="ph142915347164"></a><a name="ph142915347164"></a>Ascend Device Plugin</span> is deployed in a VM scenario, systemd needs to be installed in the image of <span id="ph0429634121617"><a name="ph0429634121617"></a><a name="ph0429634121617"></a>Ascend Device Plugin</span>. It is recommended to add the <strong id="b93339419563"><a name="b93339419563"></a><a name="b93339419563"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong> command to the Dockerfile for installation.</p>
</td>
</tr>
<tr id="row1150514563377"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p450675616371"><a name="p450675616371"></a><a name="p450675616371"></a>Restart Scenario</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p105070566371"><a name="p105070566371"></a><a name="p105070566371"></a>If the user modifies the basic information of the NPU after installing <span id="ph444301153912"><a name="ph444301153912"></a><a name="ph444301153912"></a>Ascend Device Plugin</span>, for example, modifying the device IP, <span id="ph52417305424"><a name="ph52417305424"></a><a name="ph52417305424"></a>Ascend Device Plugin</span> needs to be restarted. Otherwise, <span id="ph23611038174213"><a name="ph23611038174213"></a><a name="ph23611038174213"></a>Ascend Device Plugin</span> cannot correctly identify the relevant information of the NPU.</p>
</td>
</tr>
</tbody>
</table>

## Procedure<a name="section71204451253"></a>

1. Log in to each compute node as the `root` user and run the following command to check whether the image and version number are correct.

    ```shell
    docker images | grep k8sdeviceplugin
    ```

    Command output:

    ```ColdFusion
    ascend-k8sdeviceplugin               v26.0.0              29eec79eb693        About an hour ago   105MB
    ```

    - If correct, go to [Step 2](#zh-cn_topic_0000001497364849_li922154411117).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

2. <a name="zh-cn_topic_0000001497364849_li922154411117"></a>Copy the YAML files in the decompressed directory of the Ascend Device Plugin package to any directory on the K8s management node. Note that you need to use the YAML files that are adapted to the specific processor model. To prevent exceptions in the automatic identification of Ascend Docker Runtime, do not modify the `DaemonSet.metadata.name` field in the YAML files. For details, see the following table.

    **Table 2** YAML file list of Ascend Device Plugin

    <a name="zh-cn_topic_0000001497364849_table58619457211"></a>

    | YAML File List                                                  | Description                                                                                                  |
    |-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
    | device-plugin-310-v<i>\{version\}</i>.yaml                | Configuration file for inference servers (with Atlas 300I inference cards) without Volcano.                                                             |
    | device-plugin-310-volcano-v<i>\{version\}</i>.yaml        | Configuration file for inference servers (with Atlas 300I inference cards) using Volcano.                                                              |
    | device-plugin-310P-1usoc-v<i>\{version\}</i>.yaml         | Configuration file for Atlas 200I SoC A1 core boards without Volcano.                                                              |
    | device-plugin-310P-1usoc-volcano-v<i>\{version\}</i>.yaml | Configuration file for Atlas 200I SoC A1 core boards using Volcano.                                                               |
    | device-plugin-310P-v<i>\{version\}</i>.yaml               | Configuration file for Atlas inference series products other than Atlas 200I SoC A1 core boards without Volcano.                                             |
    | device-plugin-310P-volcano-v<i>\{version\}</i>.yaml       | Configuration file for Atlas inference series products other than Atlas 200I SoC A1 core boards using Volcano.                                              |
    | device-plugin-910-v<i>\{version\}</i>.yaml                | Configuration file for Atlas training series products, <term>Atlas A2 training series products</term>, <term>Atlas A3 training series products</term>, or Atlas 800I A2 inference servers and A200I A2 Box heterogeneous components without Volcano. |
    | device-plugin-volcano-v<i>\{version\}</i>.yaml            | Configuration file for Atlas training series products, <term>Atlas A2 training series products</term>, <term>Atlas A3 training series products</term>, or Atlas 800I A2 inference servers and A200I A2 Box heterogeneous components using Volcano.  |
    | device-plugin-npu-v<i>\{version\}</i>.yaml                | Configuration file for Atlas 350 standard cards, Atlas 850 series hardware products, and Atlas 950 SuperPoD without Volcano.                                  |
    | device-plugin-npu-volcano-v<i>\{version\}</i>.yaml        | Configuration file for Atlas 350 standard cards, Atlas 850 series hardware products, and Atlas 950 SuperPoD using Volcano.                                   |

3. If you do not modify the component startup parameters, skip this step. Otherwise, modify the startup parameters of Ascend Device Plugin based on the actual situation. For details about the startup parameters, see [Table 3](#table1064314568229). You can run `./device-plugin -h` to view parameter descriptions.
    - On a node with Atlas 200I SoC A1 core board, modify the startup parameters of Ascend Device Plugin in the startup script `run_for_310P_1usoc.sh`. After modification, you need to rebuild the image on all nodes with Atlas 200I SoC A1 core board, or rebuild the image on this node and distribute it to all other nodes with Atlas 200I SoC A1 core board.

        >[!NOTE]
        >If Volcano is not used as the scheduler, when starting Ascend Device Plugin, you need to modify the startup parameters of Ascend Device Plugin in `run_for_310P_1usoc.sh` and set the `-volcanoType` parameter to `false`.

    - For other types of nodes, modify the startup parameters of Ascend Device Plugin in the corresponding startup YAML file.

4. (Optional) To use **resumable training** (including process-level recovery) or **elastic training**, modify the startup YAML of the Ascend Device Plugin component based on the fault handling mode to be used.

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
                     <strong>-volcanoType=true                    # Volcano must be used in rescheduling scenarios
                     -autoStowing=true                    # Whether to enable automatic admission. The default value is true. Setting it to false disables automatic admission. When the chip health status changes from unhealthy to healthy, it will not be automatically added to the schedulable resource pool. When automatic admission is disabled, the chip will not be automatically added to the schedulable resource pool after the parameter plane network fault is recovered. This feature is only applicable to Atlas Training Series products.
                     -listWatchPeriod=5                   # Set the health status check period, value range [3,1800], unit: seconds
                     -hotReset=2 # When using process-level recovery, set the hotReset parameter value to 2 to enable offline recovery mode</strong>
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                     -logLevel=0" ]
            securityContext:
              privileged: true
              readOnlyRootFilesystem: true
    ...</pre>

5. (Optional) When using inference card failure recovery, you need to configure the hot reset function.

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
                     <strong>-hotReset=0 # Enable the hot reset function when using inference card failure recovery</strong>
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                     -logLevel=0" ]
    ...</pre>

6. (Optional) If you need to change the default port of kubelet, modify the startup YAML of the Ascend Device Plugin component. The following is an example.

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
         <strong>- name: KUBELET_PORT   # Notifies the Ascend Device Plugin component of the default kubelet port number on the current node. If the default kubelet port number is not customized, this field does not need to be passed.
           value: "10251"</strong>
    volumes:
       - name: device-plugin
         hostPath:
           path: /var/lib/kubelet/device-plugins
    ...</pre>

7. (Optional) Modify the mount configuration in the startup YAML of the Ascend Device Plugin component based on the container runtime type.

    - If the container runtime is Docker, retain the `docker-sock` and `docker-dir` mount configurations. Example:

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

    - If the container runtime is containerd, delete the `docker-sock` and `docker-dir` mount configurations and retain the containerd mount configuration. Example:

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
    >- If the `docker.sock` file path is not `/run/docker.sock`, modify it to the actual path in `volumes`. Symbolic links are not supported.
    >- If the docker directory is not `/var/run/docker`, modify it to the actual path in `volumes`. Symbolic links are not supported.
    >- If the containerd directory is not `/run/containerd`, modify it to the actual path in `volumes`. Symbolic links are not supported.

8. On the K8s management node, run the following command in the corresponding YAML path to start Ascend Device Plugin.

    - Nodes in the K8s cluster that use other  Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas 800I A2 inference server, or A200I A2 Box heterogeneous subrack (used with Volcano, virtualization instances supported, and static Virtualization enabled by default in the YAML)

        ```shell
        kubectl apply -f device-plugin-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas 800I A2 inference server, or A200I A2 Box heterogeneous subrack (Ascend Device Plugin works independently, not used with Volcano)

        ```shell
        kubectl apply -f device-plugin-910-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use inference servers (with Atlas 300I inference cards) (using Volcano)

        ```shell
        kubectl apply -f device-plugin-310-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use inference servers (with Atlas 300I inference cards) (Ascend Device Plugin works independently, not using Volcano)

        ```shell
        kubectl apply -f device-plugin-310-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas inference series products (using Volcano, virtualization supported, static virtualization enabled by default in the YAML)

        ```shell
        kubectl apply -f device-plugin-310P-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas inference series products (Ascend Device Plugin works independently, not using the Volcano Scheduler)

        ```shell
        kubectl apply -f device-plugin-310P-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas 200I SoC A1 Core Boards (using Volcano)

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas 200I SoC A1 Core Boards (Ascend Device Plugin works independently, not using Volcano)

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-v{version}.yaml
        ```

    >[!NOTE]
    >If the K8s cluster uses multiple types of Ascend AI Processors, run the corresponding commands separately.

    Startup example:

    ```ColdFusion
    serviceaccount/ascend-device-plugin-sa created
    clusterrole.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-role created
    clusterrolebinding.rbac.authorization.K8s.io/pods-node-ascend-device-plugin-rolebinding created
    daemonset.apps/ascend-device-plugin-daemonset created
    ```

9. Run the following command on the K8s management node to check whether the component has started successfully.

    ```shell
    kubectl get pod -n kube-system
    ```

    The following is an example of the output. `Running` indicates that the component startup is successful.

    ```ColdFusion
    NAME                                        READY   STATUS    RESTARTS   AGE
    ...
    ascend-device-plugin-daemonset-d5ctz  1/1   Running   0        11s
    ...
    ```

>[!NOTE]
>
>- If the pod status of the component is not `Running` after installation, see [Component Pod Status Is Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
>- If the pod status of the component is `ContainerCreating` after installation, see [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343).
>- If the component fails to start, see [Cluster Scheduling Component Startup Failure, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
>- If the component starts successfully but the corresponding pod cannot be found, see [Component Startup YAML Executed Successfully but Corresponding Pod Not Found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Parameter Description<a name="section479917441223"></a>

**Table 3** Ascend Device Plugin startup parameters

<a name="table1064314568229"></a>

|Parameter|Type|Default Value|Description|
|--|--|--|--|
|-fdFlag|bool|false|Edge scenario flag. Enable FusionDirector system to manage devices.<ul><li>true: Use FusionDirector.</li><li>false: Do not use FusionDirector.</li></ul>|
|-shareDevCount|uint|1|Shared device feature switch. Value range: 1–100.<ul><li>The default value is 1, indicating that shared devices are not enabled. A value from 2 to 100 indicates the number of shared devices virtualized from a single chip.</li><li>When the soft partitioning function is enabled, that is, -softShareDevConfigDir is not empty, this parameter must be set to 100.</li></ul><p>The following devices are supported. For other devices, this parameter is invalid and does not affect normal Component Startup.</p><ul><li>Atlas 500 A2 Intelligent Station</li><li>Atlas 200I A2 Acceleration Module</li><li>Atlas 200I DK A2 </li><li>Atlas 300I Pro Inference Card</li><li>Atlas 300V Video Analysis Card</li><li>Atlas 300V Pro Video Analysis Card</li></ul><p>If you are using the supported Atlas Inference Series products mentioned above, note the following:</p><ul><li>The shared device function is not supported when using features such as static vNPU scheduling, dynamic vNPU scheduling, Inference Card Failure recovery, and Inference Card Failure rescheduling.</li><li>The number of requested resources for a single task must be 1. Scenarios involving multi-chip allocation and cross-chip usage are not supported.</li><li>This function depends on the driver enabling shared mode and setting device-share to true. </li></ul>|
|-edgeLogFile|string|/var/alog/AtlasEdge_log/devicePlugin.log|Edge scenario log file. This parameter takes effect when fdFlag is Set To true.<p>When a single log file exceeds 20 MB, the automatic dump function is triggered. The maximum file size cannot be modified.</p>|
|-useAscendDocker|bool|true|The default value is true. Whether the container engine uses Ascend Docker Runtime. When enabling the K8s CPU core binding function, you need to uninstall Ascend Docker Runtime and restart the container engine. The values are described as follows:<ul><li>true: Use Ascend Docker Runtime.</li><li>false: Do not use Ascend Docker Runtime.</li></ul><p>MindCluster 5.0.RC1 and later versions only support automatic acquisition of the running mode and do not accept manual specification.</p>|
|-use310PMixedInsert|bool|false|Enable hybrid deployment mode.<ul><li>true: Use hybrid deployment mode.</li><li>false: Do not use hybrid deployment mode.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>Only Server Hybrid Deployment of Atlas 300I Pro Inference Card, Atlas 300V Video Analysis Card, and Atlas 300V Pro Video Analysis Card is supported.</li><li>Volcano scheduling mode is not supported in server hybrid deployment mode.</li><li>Virtualization instances are not supported in server hybrid deployment mode.</li><li>Fault rescheduling scenarios are not supported in server hybrid deployment mode.</li><li>Ascend Docker Runtime is not supported in server hybrid deployment mode.</li><li>In non-hybrid deployment mode, the resource name reported to K8s remains unchanged.<ul><li>The resource name format reported in non-hybrid deployment mode is huawei.com/Ascend310P.</li><li>The resource name format reported in hybrid deployment mode is huawei.com/Ascend310P-V, huawei.com/Ascend310P-VPro, and huawei.com/Ascend310P-IPro.</li></ul></li></ul></div></div>|
|-volcanoType|bool|false|Whether to use Volcano for scheduling. Currently, Atlas Training Series products, Atlas A2 Training Series products, Atlas Inference Series products, and Inference Server (with Atlas 300I Inference Card) chips are supported.<ul><li>true: Use Volcano.</li><li>false: Do not use Volcano.</li></ul>|
|-presetVirtualDevice|bool|true|Virtualization function switch.<ul><li>When Set To true, it indicates Static Virtualization.</li><li>When Set To false, it indicates dynamic virtualization. Volcano needs to be enabled synchronously, that is, set the -volcanoType parameter to true.</li></ul>|
|-version|bool|false|Whether to view the current version number of Ascend Device Plugin.<ul><li>true: Query.</li><li>false: Do not query.</li></ul>|
|-listWatchPeriod|int|5|<p>Set the health status check period. Value range: [3,1800]. Unit: Seconds.</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>During each period, the following checks are performed, and the check results are written to the ConfigMap.</p><ul><li>If the device information has not changed and it has been less than 5 minutes since the last ConfigMap update, the ConfigMap will not be updated.</li><li>If it has been more than 5 minutes since the last ConfigMap update, the ConfigMap will be updated regardless of whether the device information has changed.</li></ul></div></div>|
|-autoStowing|bool|true|Whether to enable Automatic Admission of repaired devices. This parameter takes effect when volcanoType is true.<ul><li>true: Automatic Admission.</li><li>false: No Automatic Admission.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>After a device fault occurs, it will be automatically isolated from K8s. If the device returns to normal, it will be automatically added to the K8s cluster resource pool by default. If the device is unstable, you can set this parameter to false, in which case manual admission is required.</p><ul><li>You can use the following command to put a chip whose health status has recovered from unhealthy to healthy back into the resource pool.<p>kubectl label nodes <i>node_name</i> huawei.com/Ascend910-Recover-</p><p>When using Ascend 950 series products, use: </p><p>kubectl label nodes <i>node_name</i> huawei.com/NPU-Recover-</p></li><li>You can use the following command to put a chip whose parameter plane network health status has recovered from unhealthy to healthy back into the resource pool.<p>kubectl label nodes <i>node_name</i> huawei.com/Ascend910-NetworkRecover-</p><p>When using Ascend 950 series products, use: </p><p>kubectl label nodes <i>node_name</i> huawei.com/NPU-NetworkRecover-</p></li></ul>|
|-logLevel|int|0|Log level:<ul><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul></div></div>|
|-maxAge|int|7|Log backup retention period. Value range: 7–700. Unit: days.|
|-logFile|string|/var/log/mindx-dl/devicePlugin/devicePlugin.log|Non-edge scenario log file. This parameter takes effect when fdFlag is Set To false.<p>When a single log file exceeds 20 MB, the automatic dump function is triggered. The maximum file size cannot be modified. The naming format of the dumped file is: devicePlugin-<dump_trigger_time>.log, for example: devicePlugin-2023-10-07T03-38-24.402.log.</p>|
|-hotReset|int|-1|Device hot reset function parameter. When this function is enabled, after a chip fault occurs, Ascend Device Plugin will perform a hot reset operation to restore the chip to health.<ul><li>-1: Disable the chip reset function</li><li>0: Enable the inference device reset function</li><li>1: Enable the training device online reset function</li><li>2: Enable the training/inference device offline reset function</li></ul><div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><p>The function corresponding to the value 1 has been deprecated. Please configure other values.</p></div></div><p>Training devices supported by this parameter:</p><ul><li>Atlas 800 Training Server (Model 9000) (Fully Configured NPU)</li><li>Atlas 800 Training Server (Model 9010) (Fully Configured NPU)</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD (Model 9000)</li><li>Atlas 800T A2 Training Server</li><li>Atlas 900 A2 PoD Cluster Base Unit</li><li>Atlas 900 A3 SuperPoD</li><li>Atlas 800T A3 SuperPoD</li></ul><p>Inference devices supported by this parameter:</p><ul><li>Atlas 300I Pro Inference Card</li><li>Atlas 300V Video Analysis Card</li><li>Atlas 300V Pro Video Analysis Card</li><li>Atlas 300I Duo Inference Card</li><li>Atlas 300I Inference Card (Model 3000) (Full Card)</li><li>Atlas 300I Inference Card (Model 3010)</li><li>Atlas 800I A2 Inference Server</li><li>A200I A2 Box Heterogeneous Component</li><li>Atlas 800I A3 SuperPoD</li></ul><div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><ul><li>For the Atlas 300I Duo Inference Card form factor hardware, only per-card reset is supported, meaning both chips will be reset simultaneously.</li><li>Atlas 800I A2 Inference Server has the following two hot reset modes. One Atlas 800I A2 Inference Server can only use one hot reset mode, which is automatically identified by the cluster scheduling component.<ul><li>Mode 1: If there is no HCCS ring on the device, during inference task execution, when an NPU fault occurs, Ascend Device Plugin waits for the NPU to become idle and then resets the NPU.</li><li>Mode 2: If there is an HCCS ring on the device, during inference task execution, when one or more faulty NPUs appear on the server, Ascend Device Plugin waits for all NPUs on the ring to become idle and then resets all NPUs on the ring at once.</li></ul></li></ul></div></div>|
|-linkdownTimeout|int|30|Network linkdown timeout duration. Unit: seconds. Value range: 1–30.<p>It is recommended that the value of this parameter be consistent with the HCCL_RDMA_TIMEOUT configured in the training script. For multi-task scenarios, it is recommended to set it to the minimum value of HCCL_RDMA_TIMEOUT among the multiple tasks.</p>|
|-enableSlowNode|bool|false|Enable the slow node detection (degradation diagnosis) function.<ul><li>true: Enable.</li><li>false: Disable.</li></ul><div class="note"><span class="notetitle">[!NOTE] Note</span><div class="notebody"><p>For details about degradation diagnosis, see the "[Degradation Diagnosis](https://support.huawei.com/hedex/hdx.do?docid=EDOC1100445519&amp;id=ZH-CN_TOPIC_0000002147436540)" section in the *iMaster CCAE Product Documentation*.</p></div></div>|
|-dealWatchHandler|bool|false|Whether to refresh the local Pod informer cache when the informer connection ends abnormally.<ul><li>true: Refresh the Pod informer cache.</li><li>false: Do not refresh the Pod informer cache.</li></ul>|
|-checkCachedPods|bool|true|Whether to periodically check the Pods in the cache. The default value is true. When a Pod in the cache has not been updated for more than 1 hour, Ascend Device Plugin will actively request the api-server to check the Pod status.<ul><li>true: Check.</li><li>false: Do not check.</li></ul>|
|-maxBackups|int|30|Maximum number of log files retained after dump. Value range: 1–30. Unit: files.|
|-thirdPartyScanDelay|int|300|<p>Waiting duration for Ascend Device Plugin component startup rescan.</p><p>After Ascend Device Plugin fails to automatically reset a chip, it writes the failure information to the node annotation. The third-party platform can reset the failed chip based on this information. The Ascend Device Plugin component waits for a period of time set by this parameter and then rescans the device.</p><p>This parameter is only supported by the Atlas 800T A3 SuperPoD.</p><p>Unit: seconds.</p>|
|-deviceResetTimeout|int|60|During component startup, if the number of chips is insufficient, the maximum duration to wait for the driver to report the complete chips. Unit: seconds. Value range: 10–600.<ul><li>Atlas A2 Training Series products, Atlas 800I A2 Inference Server, A200I A2 Box Heterogeneous Component: Recommended Configuration is 150 seconds.</li><li>Atlas A3 Training Series products, A200T A3 Box8 SuperPoD, Atlas 800I A3 SuperPoD: Recommended Configuration is 360 seconds.</li><li>Atlas 350 Standard Card, Atlas 850 Series Hardware products, Atlas 950 SuperPoD: Recommended Configuration is 600 seconds.</li></ul>|
|-softShareDevConfigDir|string|""|Configuration directory for the soft partitioning virtualization scenario. This configuration directory needs to be manually created in the root directory before installing Ascend Device Plugin. This parameter needs to be configured when using the soft partitioning function.|
|-useSingleDieMode|bool|false|Whether to enable single-die Passthrough Mode for <term>Atlas A3 Inference Series products</term>.<ul><li>true: Enable single-die Passthrough Mode.</li><li>false: Disable single-die Passthrough Mode.</li></ul>When using the soft partitioning virtualization function, this parameter must be configured as true.|
|-h or -help|N/A|N/A|Display help information.|
