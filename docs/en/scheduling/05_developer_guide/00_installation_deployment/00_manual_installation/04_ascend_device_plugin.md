# Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511426341"></a>

- Users who use full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, resumable training, elastic training, inference card failure recovery, or rescheduling upon inference card faults must install Ascend Device Plugin on compute nodes.
- Users who only use containerization support and resource monitoring do not need to install Ascend Device Plugin and can skip this chapter directly.
- Before installing Ascend Device Plugin, install Ascend Docker Runtime first. Ascend Device Plugin automatically detects whether Ascend Docker Runtime has been installed on a node.

## Constraints<a name="section1362795652416"></a>

Before installing Ascend Device Plugin, you need to understand the relevant constraints. For details, see <a href="#table113813012140">Table 1</a>.

**Table 1**  Constraints

<a name="table113813012140"></a>
<table><thead align="left"><tr id="row193815031414"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p13383051411"><a name="p13383051411"></a><a name="p13383051411"></a>Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p73814015146"><a name="p73814015146"></a><a name="p73814015146"></a>Constraint Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row738802142"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p13388019145"><a name="p13388019145"></a><a name="p13388019145"></a>NPU driver</p>
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
<p id="p019819298377"><a name="p019819298377"></a><a name="p019819298377"></a>If <span id="ph06381714397"><a name="ph06381714397"></a><a name="ph06381714397"></a>Ascend Device Plugin</span> is deployed on the <span id="ph66321793918"><a name="ph66321793918"></a><a name="ph66321793918"></a>Atlas 200I SoC A1 core board</span>, <span id="ph116721035125612"><a name="ph116721035125612"></a><a name="ph116721035125612"></a>Ascend Docker Runtime</span> is not required.</p>
</td>
</tr>
<tr id="row1648094416218"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p9484175210212"><a name="p9484175210212"></a><a name="p9484175210212"></a>The component version requirements are as follows:</p>
<p id="p44813447213"><a name="p44813447213"></a><a name="p44813447213"></a>This function requires that the versions of <span id="ph196135501025"><a name="ph196135501025"></a><a name="ph196135501025"></a>Ascend Docker Runtime</span> and <span id="ph1161319502212"><a name="ph1161319502212"></a><a name="ph1161319502212"></a>Ascend Device Plugin</span> be consistent and be 5.0.RC1 or later. After installing or uninstalling <span id="ph1361319501123"><a name="ph1361319501123"></a><a name="ph1361319501123"></a>Ascend Docker Runtime</span>, you need to restart the container engine for <span id="ph11613850625"><a name="ph11613850625"></a><a name="ph11613850625"></a>Ascend Device Plugin</span> to correctly identify it.</p>
</td>
</tr>
<tr id="row1449218752210"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p14704133226"><a name="p14704133226"></a><a name="p14704133226"></a><span id="ph1371171332212"><a name="ph1371171332212"></a><a name="ph1371171332212"></a>Ascend Device Plugin</span> and <span id="ph071513132214"><a name="ph071513132214"></a><a name="ph071513132214"></a>Ascend Docker Runtime</span> cannot be used together in the following two scenarios.<a name="ul1771141362211"></a><a name="ul1771141362211"></a><ul id="ul1771141362211"><li>Hybrid deployment scenarios.</li><li><span id="ph1471111314226"><a name="ph1471111314226"></a><a name="ph1471111314226"></a>Atlas 200I SoC A1 core board</span>.</li></ul>
</div>
</td>
</tr>
<tr id="row5381205148"><td class="cellrowborder" rowspan="3" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p16384020141"><a name="p16384020141"></a><a name="p16384020141"></a>DCMI dynamic library</p>
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
<tr id="row11391707149"><td class="cellrowborder" rowspan="2" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p133919013143"><a name="p133919013143"></a><a name="p133919013143"></a><span id="ph1078193611515"><a name="ph1078193611515"></a><a name="ph1078193611515"></a>Atlas 200I SoC A1 core board</span></p>
<p id="p1918223205014"><a name="p1918223205014"></a><a name="p1918223205014"></a></p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p786843510309"><a name="p786843510309"></a><a name="p786843510309"></a>If <span id="ph1480005781518"><a name="ph1480005781518"></a><a name="ph1480005781518"></a>Ascend Device Plugin</span> is deployed in image mode on an <span id="ph080185715158"><a name="ph080185715158"></a><a name="ph080185715158"></a>Atlas 200I SoC A1 core board</span> node, you need to configure the multi-container sharing mode.</p>
</td>
</tr>
<tr id="row4248144116153"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><div class="p" id="p5840775161"><a name="p5840775161"></a><a name="p5840775161"></a>When using <span id="ph697712515161"><a name="ph697712515161"></a><a name="ph697712515161"></a>Ascend Device Plugin</span> with the <span id="ph99771752169"><a name="ph99771752169"></a><a name="ph99771752169"></a>Atlas 200I SoC A1 core board</span>, the following compatibility relationships must be observed:<a name="ul2977251161"></a><a name="ul2977251161"></a><ul id="ul2977251161"><li><span id="ph49779571614"><a name="ph49779571614"></a><a name="ph49779571614"></a>Ascend Device Plugin</span> version 5.0.RC2 must be used with the driver of <span id="ph5977135101614"><a name="ph5977135101614"></a><a name="ph5977135101614"></a>Atlas 200I SoC A1 Core Board</span> version 23.0.RC2 or later.</li><li><span id="ph59771512164"><a name="ph59771512164"></a><a name="ph59771512164"></a>Ascend Device Plugin</span> versions earlier than 5.0.RC2 can only be used with the driver of <span id="ph1977115181612"><a name="ph1977115181612"></a><a name="ph1977115181612"></a>Atlas 200I SoC A1 core board</span> earlier than 23.0.RC2.</li></ul>
</div>
</td>
</tr>
<tr id="row14538194431511"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p45382449151"><a name="p45382449151"></a><a name="p45382449151"></a>Virtual machine</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p8538144420153"><a name="p8538144420153"></a><a name="p8538144420153"></a>If <span id="ph142915347164"><a name="ph142915347164"></a><a name="ph142915347164"></a>Ascend Device Plugin</span> is deployed in a virtual machine, systemd needs to be installed in the image of <span id="ph0429634121617"><a name="ph0429634121617"></a><a name="ph0429634121617"></a>Ascend Device Plugin</span>. It is recommended to add the <strong id="b93339419563"><a name="b93339419563"></a><a name="b93339419563"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong> command to the Dockerfile for installation.</p>
</td>
</tr>
<tr id="row1150514563377"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p450675616371"><a name="p450675616371"></a><a name="p450675616371"></a>Restart</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p105070566371"><a name="p105070566371"></a><a name="p105070566371"></a>If the user modifies the basic information of NPU after installing <span id="ph444301153912"><a name="ph444301153912"></a><a name="ph444301153912"></a>Ascend Device Plugin</span>, for example, modifying the device IP, <span id="ph52417305424"><a name="ph52417305424"></a><a name="ph52417305424"></a>Ascend Device Plugin</span> needs to be restarted. Otherwise, <span id="ph23611038174213"><a name="ph23611038174213"></a><a name="ph23611038174213"></a>Ascend Device Plugin</span> cannot correctly identify the relevant information of the NPU.</p>
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
    ascend-k8sdeviceplugin               v26.1.0              29eec79eb693        About an hour ago   105MB
    ```

    - If correct, go to [Step 2](#zh-cn_topic_0000001497364849_li922154411117).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

2. <a name="zh-cn_topic_0000001497364849_li922154411117"></a>Copy the YAML files in the decompressed directory of the Ascend Device Plugin package to any directory on the K8s management node. Note that you need to use the YAML files that are adapted to the specific processor model. To prevent exceptions in the automatic identification of Ascend Docker Runtime, do not modify the `DaemonSet.metadata.name` field in the YAML files. For details, see the following table.

    **Table 2** YAML file list of Ascend Device Plugin

    <a name="zh-cn_topic_0000001497364849_table58619457211"></a>

    | YAML File List                                                  | Description                                                                                                  |
    |-----------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
    | device-plugin-310P-1usoc-v<i>\{version\}</i>.yaml         | Configuration file for Atlas 200I SoC A1 core board without Volcano                                                               |
    | device-plugin-310P-1usoc-volcano-v<i>\{version\}</i>.yaml | Configuration file for Atlas 200I SoC A1 core board with Volcano                                                               |
    | device-plugin-v<i>\{version\}</i>.yaml                | Configuration file for products without Volcano (other than the Atlas 200I SoC A1 core board) |
    | device-plugin-volcano-v<i>\{version\}</i>.yaml            | Configuration file for products with Volcano (other than the Atlas 200I SoC A1 core board)  |

3. If you do not modify the component startup parameters, skip this step. Otherwise, modify the startup parameters of Ascend Device Plugin based on the actual situation. For details about the startup parameters, see [Table 3](#table1064314568229). You can run `./device-plugin -h` to view parameter descriptions.
    - On a node with Atlas 200I SoC A1 core board, modify the startup parameters of Ascend Device Plugin in the startup script `run_for_310P_1usoc.sh`. After modification, you need to rebuild the image on all nodes with Atlas 200I SoC A1 core board, or rebuild the image on this node and distribute it to all other nodes with Atlas 200I SoC A1 core board.

        >[!NOTE]
        >If Volcano is not used as the scheduler, when starting Ascend Device Plugin, you need to modify the startup parameters of Ascend Device Plugin in `run_for_310P_1usoc.sh` and set the `-volcanoType` parameter to `false`.

    - For other types of nodes, modify the startup parameters of Ascend Device Plugin in the corresponding startup YAML file.

4. <a name="step_npu_nic_mapping"></a> (Optional for A5) If the compute node has a 1825 NIC (DPU), Ascend Device Plugin reads the connection relationship between NPUs and 1825 NICs through `npu-nic-mapping.json`, which is used to generate RoCE communication addresses for A5 devices. The default configuration in the software package may not match the actual topology of the node. It is recommended to create the mapping file manually according to the node's actual conditions and mount it into the container.

    - Obtain the physical connection relationship between NPUs and 1825 NICs from the driver side.

        The 1825 NIC is a UB network device that exists in the system in the form of an Ethernet interface. Run the following command to query the Ethernet interface name (e.g., ens0f0) corresponding to each RDMA device.

        ```shell
        ls /sys/class/infiniband/*/device/net/
        ```

    - Create the `npu-nic-mapping.json` file for this node, and fill it in according to the obtained connection relationships. The file format is as follows.

        ```json
        {
          "npuNics": [
            {
              "npuId": 0,
              "nicNames": ["ens0f0", "ens0f2", "ens1f0"]
            },
            {
              "npuId": 1,
              "nicNames": ["ens1f0", "ens1f2", "ens0f0"]
            }
          ]
        }
        ```

        Field descriptions:
        - `npuNics`: The list of mappings between NPUs and NICs.
        - `npuId`: The physical ID of the NPU. Each NPU on the node must have one corresponding record configured.
        - `nicNames`: The list of 1825 NIC interface names connected to this NPU, arranged in descending order of priority. The component will traverse this list in order and use the first NIC with a valid IP address to generate the RoCE communication address.

    - Add the following mount declarations to the Ascend Device Plugin YAML to mount the created file to the fixed path `/user/mindx-dl/npu/npu-nic-mapping.json` inside the container (this path cannot be modified).

        ```yaml
        volumeMounts:
          ...
          - name: npu-nic-mapping
            mountPath: /user/mindx-dl/npu/npu-nic-mapping.json
            readOnly: true
          ...
        volumes:
          ...
          - name: npu-nic-mapping
            hostPath:
              path: /user/mindx-dl/npu/npu-nic-mapping.json
              type: File
          ...
        ```

        `hostPath.path` above is the actual path of the file on the compute node (the host path can be adjusted, but symbolic links are not supported).

    >[!NOTE]
    >- The physical connection relationships between NPUs and 1825 NICs may differ across nodes. Please create and mount the file separately according to the actual topology of each node.
    >- This configuration file is read when the component starts. After modification, the component needs to be restarted for the configuration to take effect.

5. (Optional) To use **resumable training** (including process-level recovery) or **elastic training**, modify the startup YAML of the Ascend Device Plugin component based on the fault handling mode to be used.

    <pre codetype="yaml">
    ...
          containers:
          - image: ascend-k8sdeviceplugin:v26.1.0
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
                     <strong>-volcanoType=true                    # Volcano must be used in rescheduling scenarios
                     -autoStowing=true                    # This field is deprecated. Whether to enable automatic management. The default value is true. Setting it to false disables automatic management. When the chip health status changes from unhealthy to healthy, it will not be automatically added to the schedulable resource pool. When automatic management is disabled and the chip parameter-plane network recovers from a failure, it will not be automatically added to the schedulable resource pool. This feature only applies to <term>Atlas training products</term>
                     -listWatchPeriod=5                   # Set the health check interval, range [3,1800], in seconds
                     -hotReset=2 # When using process-level recovery, set the hotReset parameter value to 2 to enable offline recovery mode</strong>
                     -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
                     -logLevel=0" ]
            securityContext:
              privileged: true
              readOnlyRootFilesystem: true
    ...</pre>

6. (Optional) When using inference card failure recovery, you need to configure the hot reset function.

    <pre codetype="yaml">
          containers:
          - image: ascend-k8sdeviceplugin:v26.1.0
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

7. (Optional) If you need to change the default port of kubelet, modify the startup YAML of the Ascend Device Plugin component. The following is an example.

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

8. (Optional) If Ascend Docker Runtime is not installed, you need to manually mount the Docker or Containerd sock file. Examples are as follows depending on the runtime type.

    - If the container runtime is Docker, retain the `docker-sock` and `docker-dir` mount configurations. Example:

        ```yaml
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

9. On the K8s management node, run the following command in the corresponding YAML path to start Ascend Device Plugin.

    - Nodes in the K8s cluster that use products other than Atlas 200I SoC A1 core board (used with Volcano, virtualization instances supported, and static virtualization enabled by default in the YAML)

        ```shell
        kubectl apply -f device-plugin-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use products other than Atlas 200I SoC A1 core board (Ascend Device Plugin works independently, not used with Volcano)

        ```shell
        kubectl apply -f device-plugin-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas 200I SoC A1 core board (using Volcano)

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-volcano-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use Atlas 200I SoC A1 core board (Ascend Device Plugin works independently, not using Volcano)

        ```shell
        kubectl apply -f device-plugin-310P-1usoc-v{version}.yaml
        ```

    >[!NOTE]
    >If the K8s cluster uses multiple types of Ascend AI Processors, run the corresponding commands separately.

    Startup example:

    ```ColdFusion
    serviceaccount/ascend-device-plugin-sa created
    clusterrole.rbac.authorization.k8s.io/pods-node-ascend-device-plugin-role created
    clusterrolebinding.rbac.authorization.k8s.io/pods-node-ascend-device-plugin-rolebinding created
    daemonset.apps/ascend-device-plugin-daemonset created
    ```

10. Run the following command on the K8s management node to check whether the component has started successfully.

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
| `-fdFlag` | bool | false | Edge scenario flag, whether to use the FusionDirector system to manage devices.<ul><li>`true`: Use FusionDirector.</li><li>`false`: Do not use FusionDirector.</li></ul> |
| `-shareDevCount` | uint | 1 | Shared device feature switch, value range 1–100.<ul><li>The default value is `1`, which means shared devices are not enabled. Values from 2 to 100 indicate the number of shared devices virtualized from a single chip.</li><li>When soft partitioning is enabled (i.e., `-softShareDevConfigDir` is not empty), this parameter must be set to `100`.</li></ul><p>Supported on the following devices; for other devices, this parameter is invalid and does not affect normal component startup.</p><ul><li>Atlas 500 A2 intelligent edge station</li><li>Atlas 200I A2 accelerator module</li><li>Atlas 200I DK A2 </li><li>Atlas 300I Pro inference card</li><li>Atlas 300V video analysis card</li><li>Atlas 300V Pro video analysis card</li></ul><p>If the user is using any of the above supported <term>Atlas inference products</term>, note the following:</p><ul><li>The shared device feature is not supported when using static vNPU scheduling, dynamic vNPU scheduling, inference card fault recovery, or rescheduling upon inference card faults.</li><li>The requested resource count for a single task must be 1. Scenarios involving multi-chip allocation or cross-chip usage are not supported.</li><li>This feature depends on the driver enabling shared mode by setting `device-share` to `true`.</li></ul> |
| `-edgeLogFile` | string | `/var/alog/AtlasEdge_log/devicePlugin.log` | Edge scenario log file. Takes effect when `fdFlag` is set to `true`.<p>Automatic log rotation is triggered when a single log file exceeds 20 MB. The maximum file size is not configurable.</p> |
| `-use310PMixedInsert` | bool | false | Whether to use mixed insertion mode.<ul><li>`true`: Use mixed insertion mode.</li><li>`false`: Do not use mixed insertion mode.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>Only supports mixed insertion of Atlas 300I Pro inference card, Atlas 300V Video Parsing Card, and Atlas 300V Pro video analysis card in servers.</li><li>Volcano scheduling mode is not supported in mixed insertion mode.</li><li>Virtual instances are not supported in mixed insertion mode.</li><li>Fault rescheduling is not supported in mixed insertion mode.</li><li>Ascend Docker Runtime is not supported in mixed insertion mode.</li><li>In non-mixed insertion mode, the resource name reported to K8s remains unchanged.<ul><li>In non-mixed insertion mode, the resource name format is `huawei.com/Ascend310P`.</li><li>In mixed insertion mode, the resource name formats are: `huawei.com/Ascend310P-V`, `huawei.com/Ascend310P-VPro`, and `huawei.com/Ascend310P-IPro`.</li></ul></li></ul></div></div> |
| `-volcanoType` | bool | false | Whether to use Volcano for scheduling. Currently supported for <term>Atlas training products</term>, <term>Atlas A2 training products</term>, <term>Atlas inference products</term>, and inference servers (with Atlas 300I inference cards installed).<ul><li>`true`: Use Volcano.</li><li>`false`: Do not use Volcano.</li></ul> |
| `-presetVirtualDevice` | bool | true | Virtualization switch.<ul><li>When set to `true`, static virtualization is used.</li><li>When set to `false`, dynamic virtualization is used. Volcano must also be enabled by setting the `-volcanoType` parameter to `true`.</li></ul> |
| `-version` | bool | false | Whether to query the version number of the current Ascend Device Plugin.<ul><li>`true`: Query.</li><li>`false`: Do not query.</li></ul> |
| `-listWatchPeriod` | int | 5 | <p>Sets the health check interval, in seconds, with a value range of [3, 1800].</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>Within each interval, the following checks are performed and the results are written to the ConfigMap:</p><ul><li>If the device information has not changed and it has been less than 5 minutes since the last ConfigMap update, the ConfigMap will not be updated.</li><li>If more than 5 minutes have passed since the last ConfigMap update, the ConfigMap will be updated regardless of whether the device information has changed.</li></ul></div></div> |
| `-autoStowing` | bool | true | Whether to automatically manage repaired devices. Takes effect when `volcanoType` is `true`.<ul><li>`true`: Automatic management.</li><li>`false`: Do not automatically manage.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>This field is deprecated.</p><p>After a device failure, it is automatically isolated from K8s. If the device returns to normal, it is automatically added back to the K8s cluster resource pool by default. If the device is unstable, this can be set to `false`, in which case manual management is required.</p><ul><li>Users can use the following command to return a chip whose health status has changed from `unhealthy` to `healthy` back to the resource pool:<p>`kubectl label nodes <i>node_name</i> huawei.com/Ascend910-Recover-`</p><p>When using <term>Ascend 950 products</term>, use:</p><p>`kubectl label nodes <i>node_name</i> huawei.com/NPU-Recover-`</p></li><li>Users can use the following command to return a chip whose parameter-plane network health status has changed from `unhealthy` to `healthy` back to the resource pool:<p>`kubectl label nodes <i>node_name</i> huawei.com/Ascend910-NetworkRecover-`</p><p>When using <term>Ascend 950 products</term>, use:</p><p>`kubectl label nodes <i>node_name</i> huawei.com/NPU-NetworkRecover-`</p></li></ul></div></div> |
| `-logLevel` | int | 0 | Log level:<ul><li>`-1`: debug</li><li>`0`: info</li><li>`1`: warning</li><li>`2`: error</li><li>`3`: critical</li></ul> |
| `-maxAge` | int | 7 | Log backup retention period, in days, value range 7–700. |
| `-logFile` | string | `/var/log/mindx-dl/devicePlugin/devicePlugin.log` | Non-edge scenario log file. Takes effect when `fdFlag` is set to `false`.<p>Automatic log rotation is triggered when a single log file exceeds 20 MB. The maximum file size is not configurable. Rotated files are named in the format: devicePlugin-<i>{rotation_time}</i>.log, e.g., `devicePlugin-2023-10-07T03-38-24.402.log`.</p> |
| `-hotReset` | int | -1 | Device hot reset feature parameter. When enabled, if a chip fails, Ascend Device Plugin performs a hot reset to recover the chip.<ul><li>`-1`: Disable chip reset</li><li>`0`: Enable inference device reset</li><li>`1`: Enable training device online reset</li><li>`2`: Enable training/inference device offline reset</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>The functionality corresponding to value `1` is deprecated. Please configure other values.</p></div></div><p>Supported devices for this parameter:</p><ul><li>Atlas 800 training server (model 9000) (fully populated with NPUs)</li><li>Atlas 800 training server (model 9010) (fully populated with NPUs)</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD (model 9000)</li><li>Atlas 800T A2 training server</li><li>Atlas 900 A2 PoD cluster basic unit</li><li>Atlas 900 A3 SuperPoD</li><li>Atlas 800T A3 SuperPoD server</li><li>Atlas 850E SuperPoD</li><li>Atlas 650E server</li><li>Atlas 950 SuperPoD</li><li>Atlas 350 accelerator card</li><li>Atlas 300I Pro inference card</li><li>Atlas 300V video analysis card</li><li>Atlas 300V Pro video analysis card</li><li>Atlas 300I Duo inference card</li><li>Atlas 300I inference card (model 3000) (entire card)</li><li>Atlas 300I inference card (model 3010)</li><li>Atlas 800I A2 inference server</li><li>A200I A2 Box heterogeneous subrack</li><li>Atlas 800I A3 SuperPoD server</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><ul><li>For the Atlas 300I Duo inference card form factor hardware, only card-level reset is supported, meaning both chips are reset simultaneously.</li><li>For the Atlas 800I A2 inference server, there are two hot reset modes. Each server can only use one mode, and the cluster scheduling components automatically identify which mode to use.<ul><li>Mode 1: If no HCCS ring exists on the device, during inference tasks, when an NPU fails, Ascend Device Plugin waits for that NPU to become idle and then performs a reset on that NPU.</li><li>Mode 2: If an HCCS ring exists on the device, during inference tasks, when one or more NPUs on the server fail, Ascend Device Plugin waits for all NPUs on the ring to become idle and then resets all NPUs on the ring at once.</li></ul></li><li>Hot reset recovery cannot cover all faults. Some faults may fail to recover, e.g., faults causing card removal, device OS hang, etc.</li></ul></div></div> |
| `-linkdownTimeout` | int | 30 | Network link-down timeout, in seconds, value range 1–30.<p>It is recommended that this parameter value be consistent with the `HCCL_RDMA_TIMEOUT` configured in the training script. For multi-task scenarios, it is recommended to set it to the minimum `HCCL_RDMA_TIMEOUT` value among the tasks.</p> |
| `-enableSlowNode` | bool | false | Whether to enable the slow node detection (degradation diagnosis) feature.<ul><li>`true`: Enable.</li><li>`false`: Disable.</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>For detailed information on degradation diagnosis, refer to the "[Degradation Diagnosis](https://support.huawei.com/hedex/hdx.do?docid=EDOC1100445519&id=ZH-CN_TOPIC_0000002147436540)" section in the *iMaster CCAE Product Documentation*.</p></div></div> |
| `-dealWatchHandler` | bool | false | Whether to refresh the local Pod informer cache when the informer connection terminates abnormally.<ul><li>`true`: Refresh the Pod informer cache.</li><li>`false`: Do not refresh the Pod informer cache.</li></ul> |
| `-checkCachedPods` | bool | true | Whether to periodically check Pods in the cache. Default is `true`. If a Pod in the cache has not been updated for more than 1 hour, Ascend Device Plugin will actively request the api-server to check the Pod status.<ul><li>`true`: Check.</li><li>`false`: Do not check.</li></ul> |
| `-maxBackups` | int | 30 | Maximum number of retained log files after rotation, value range 1–180. |
| `-thirdPartyScanDelay` | int | 300 | <p>The waiting duration for Ascend Device Plugin to re-scan devices after startup.</p><p>After Ascend Device Plugin fails to automatically reset a chip, it writes the failure information to the node annotation. Third-party platforms can use this information to reset failed chips. After waiting for the duration specified by this parameter, Ascend Device Plugin re-scans the devices.</p><p>This parameter is only supported on the Atlas 800T A3 SuperPoD server.</p><p>Unit: seconds.</p> |
| `-deviceResetTimeout` | int | 600 | When the component starts, if the number of chips is insufficient, this is the maximum time to wait for the driver to report the complete chip set, in seconds, value range 10–600.<ul><li><term>Atlas A2 training products</term>, Atlas 800I A2 inference server, A200I A2 Box heterogeneous subrack: Recommended value is 150 seconds.</li><li><term>Atlas A3 training products</term>, A200T A3 Box8 SuperPoD server, Atlas 800I A3 SuperPoD server: Recommended value is 360 seconds.</li><li>Atlas 350 Accelerator Card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD: Recommended value is 600 seconds.</li></ul> |
| `-softShareDevConfigDir` | string | `""` | Configuration directory for soft-partitioning virtualization scenarios. This directory must be manually created in the root directory before installing Ascend Device Plugin. This parameter must be configured when using the soft partitioning feature. |
| `-useSingleDieMode` | bool | false | Whether to enable single-die passthrough mode for <term>Atlas A3 inference products</term>.<ul><li>`true`: Enable single-die passthrough mode.</li><li>`false`: Disable single-die passthrough mode.</li></ul>This parameter must be set to `true` when using the soft partitioning virtualization feature. |
| `--enable-healthz` | bool | false | Whether to enable the health check service. Enabled (`true`) via the component YAML configuration during K8s deployment.<ul><li>`true`: Enable.</li><li>`false`: Disable.</li></ul> |
| `--healthz-address` | string | 11251 | The port number on which the health check service listens, value range 1025–65535. Configured as 11251 via the component YAML during K8s deployment. If the specified port is occupied, the component fails to start. |
| `--tls-cert-file` | string | `""` | Path to the HTTPS certificate file. If empty, the HTTP protocol is used. Must be configured together with `--tls-private-key-file`, or both left empty. For configuration methods and security considerations, refer to [Health Probe Security Hardening](../../../07_references/04_security_hardening.md). |
| `--tls-private-key-file` | string | `""` | Path to the HTTPS private key file. If empty, the HTTP protocol is used. Must be configured together with `--tls-cert-file`, or both left empty. |
| `-h` or `--help` | N/A | N/A | Display help information. |
