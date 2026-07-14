# NPU Exporter<a name="ZH-CN_TOPIC_0000002511426331"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:12:31.569Z pushedAt=2026-06-09T06:22:06.772Z -->

- To use **Resource Monitoring**, NPU Exporter must be installed. This component supports integration with Prometheus or Telegraf.
    - When integrating with Prometheus, NPU Exporter can be deployed in both Image Mode and Binary Mode. For deployment differences, see [Image and Binary Deployment Differences](../../../appendix.md#differences-between-image-and-binary-deployments).
    - When integrating with Telegraf, refer to the [Working with Telegraf](../../../usage/resource_monitoring/03_working_with_telegraf.md) section to install NPU Exporter and Telegraf.

- Users who do not use **Resource Monitoring** can skip installing NPU Exporter.

## Constraints<a name="section1362795652416"></a>

Before installing NPU Exporter, you need to understand the relevant constraints in advance. For details, see [Table 1](#table105071852271).

**Table 1**  Constraints

<a name="table105071852271"></a>
<table><thead align="left"><tr id="row2050719520272"><th class="cellrowborder" valign="top" width="29.970000000000002%" id="mcps1.2.3.1.1"><p id="p1950795152711"><a name="p1950795152711"></a><a name="p1950795152711"></a>Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="70.03%" id="mcps1.2.3.1.2"><p id="p75071151277"><a name="p75071151277"></a><a name="p75071151277"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row115077513271"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p17925222411"><a name="p17925222411"></a><a name="p17925222411"></a>NPU Driver</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p450745142712"><a name="p450745142712"></a><a name="p450745142712"></a><span id="ph10112356112713"><a name="ph10112356112713"></a><a name="ph10112356112713"></a>NPU Exporter</span> periodically calls the relevant interfaces of the NPU driver to detect the NPU status. If you need to upgrade the driver, stop the task first, and then stop the <span id="ph154413248375"><a name="ph154413248375"></a><a name="ph154413248375"></a>NPU Exporter</span> container service.</p>
<div class="note" id="note1993172317415"><a name="note1993172317415"></a><a name="note1993172317415"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226378_p18934232419"><a name="zh-cn_topic_0000002479226378_p18934232419"></a><a name="zh-cn_topic_0000002479226378_p18934232419"></a>To ensure that <span id="zh-cn_topic_0000002479226378_ph7206429154119"><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a>NPU Exporter</span> can be installed by a non-root user (such as hwMindX) when deployed in Binary Mode, use the --install-for-all parameter when installing the driver. An example is as follows.<pre class="screen" id="zh-cn_topic_0000002479226378_screen15239164112445"><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a>./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre>
</div>
</div></div>
</td>
</tr>
<tr id="row54685525282"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p5249201634114"><a name="p5249201634114"></a><a name="p5249201634114"></a><span id="ph1461172794116"><a name="ph1461172794116"></a><a name="ph1461172794116"></a>K8s</span> Version</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p5468852142813"><a name="p5468852142813"></a><a name="p5468852142813"></a>Before using <span id="ph98079531286"><a name="ph98079531286"></a><a name="ph98079531286"></a>NPU Exporter</span>, ensure the environment's <span id="ph18807253152810"><a name="ph18807253152810"></a><a name="ph18807253152810"></a>K8s</span> version is compatible. If the <span id="ph6808453102813"><a name="ph6808453102813"></a><a name="ph6808453102813"></a>K8s</span> version is 1.24.x or later, you need to <a href="https://github.com/mirantis/cri-dockerd#build-and-install" target="_blank" rel="noopener noreferrer">install cri-dockerd</a> as a dependency.</p>
</td>
</tr>
<tr id="row7507135142716"><td class="cellrowborder" rowspan="3" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p45071516276"><a name="p45071516276"></a><a name="p45071516276"></a>DCMI Dynamic Library</p>
<p id="p14507145152714"><a name="p14507145152714"></a><a name="p14507145152714"></a></p>
<p id="p9507651272"><a name="p9507651272"></a><a name="p9507651272"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p6555101612381"><a name="p6555101612381"></a><a name="p6555101612381"></a>The directory permission requirements for the DCMI dynamic library are as follows:</p>
<p id="p950745102715"><a name="p950745102715"></a><a name="p950745102715"></a>All parent directories of the DCMI dynamic library called by <span id="ph1496251019288"><a name="ph1496251019288"></a><a name="ph1496251019288"></a>NPU Exporter</span> must be owned by root, and programs owned by other users cannot run. Additionally, these files and their directories must not have write permissions for group and other.</p>
</td>
</tr>
<tr id="row1650710572715"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p195079518272"><a name="p195079518272"></a><a name="p195079518272"></a>The path depth of the DCMI dynamic library must be less than 20.</p>
</td>
</tr>
<tr id="row35071553276"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18507205192711"><a name="p18507205192711"></a><a name="p18507205192711"></a>If you set the dynamic library path by configuring LD_LIBRARY_PATH, the total length of the LD_LIBRARY_PATH environment variable cannot exceed 1,024.</p>
</td>
</tr>
<tr id="row75074519275"><td class="cellrowborder" rowspan="2" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p050719519271"><a name="p050719519271"></a><a name="p050719519271"></a><span id="ph13135203152812"><a name="ph13135203152812"></a><a name="ph13135203152812"></a>Atlas 200I SoC A1 Core Board</span></p>
<p id="p35076552719"><a name="p35076552719"></a><a name="p35076552719"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p209012054192411"><a name="p209012054192411"></a><a name="p209012054192411"></a>To use the <span id="ph56561935182816"><a name="ph56561935182816"></a><a name="ph56561935182816"></a>Atlas 200I SoC A1 Core Board</span> with <span id="ph1865633562811"><a name="ph1865633562811"></a><a name="ph1865633562811"></a>NPU Exporter</span>, ensure that the NPU driver of the <span id="ph10656153513282"><a name="ph10656153513282"></a><a name="ph10656153513282"></a>Atlas 200I SoC A1 Core Board</span> is version 23.0.RC2 or later. </p>
</td>
</tr>
<tr id="row165073518272"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p95251515257"><a name="p95251515257"></a><a name="p95251515257"></a>To deploy <span id="ph19614124172819"><a name="ph19614124172819"></a><a name="ph19614124172819"></a>NPU Exporter</span> in Image Mode on an <span id="ph136141041142813"><a name="ph136141041142813"></a><a name="ph136141041142813"></a>Atlas 200I SoC A1 Core Board</span> node, you need to configure the multi-container sharing mode. </p>
</td>
</tr>
<tr id="row1044710113298"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p1144701142912"><a name="p1144701142912"></a><a name="p1144701142912"></a>VM Scenario</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p14473110297"><a name="p14473110297"></a><a name="p14473110297"></a>If you deploy <span id="ph6368151492319"><a name="ph6368151492319"></a><a name="ph6368151492319"></a>NPU Exporter</span> in a VM scenario, you need to install systemd in the <span id="ph24388313372"><a name="ph24388313372"></a><a name="ph24388313372"></a>NPU Exporter</span> image. It is recommended to add the <strong id="b14813193310547"><a name="b14813193310547"></a><a name="b14813193310547"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong> command in the Dockerfile for installation.</p>
</td>
</tr>
</tbody>
</table>

## Procedure <a name="section83111543151612"></a>

NPU Exporter supports two installation modes. You can choose either one based on the actual situation. This component provides only HTTP services. If you need to use the more secure HTTPS service, modify the source code for adaptation.

- (Recommended) Run in Image Mode. For installation steps, see [Running in Image Mode](#section2035402135914).
- If high security is required, it is recommended to run in Binary Mode on a physical machine. For installation steps, see [Running in Binary Mode](#section103551921135917).

## Running in Image Mode <a name="section2035402135914"></a>

1. Log in to each compute node as the `root` user.
2. (Optional) Modify the `metricConfiguration.json` or `pluginConfiguration.json` file to configure the collection and reporting switches for default metric groups or custom metric groups.
    1. Go to the NPU Exporter package decompression directory.
    2. <a name="li11364381194"></a>Open the `metricConfiguration.json` file.

        ```shell
        vi metricConfiguration.json
        ```

    3. Press `i` to enter the insert mode and configure the collection and reporting switches for default metric groups as required.

        <a name="table192202574406"></a>
        <table><thead align="left"><tr id="row152204575408"><th class="cellrowborder" valign="top" width="30.12%" id="mcps1.1.3.1.1"><p id="p1220125712404"><a name="p1220125712404"></a><a name="p1220125712404"></a>Parameter</p>
        </th>
        <th class="cellrowborder" valign="top" width="69.88%" id="mcps1.1.3.1.2"><p id="p622019575401"><a name="p622019575401"></a><a name="p622019575401"></a>Description</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row182201357164014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p152201573404"><a name="p152201573404"></a><a name="p152201573404"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p222035704018"><a name="p222035704018"></a><a name="p222035704018"></a>Default metric group name.</p>
        <a name="ul222055714012"></a><a name="ul222055714012"></a><ul id="ul222055714012"><li>ddr: DDR data information</li><li>hccs: HCCS data information</li><li>npu: NPU data information</li><li>network: Network data information</li><li>pcie: PCIe data information</li><li>roce: RoCE data information</li><li>sio: SIO data information</li><li>vnpu: vNPU data information</li><li>version: Version data information</li><li>optical: Optical module data information</li><li>hbm: On-chip memory data information</li><li>ub: NPU UB data information</li></ul>
        </td>
        </tr>
        <tr id="row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="p182201657134015"><a name="p182201657134015"></a><a name="p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="p722015718403"><a name="p722015718403"></a><a name="p722015718403"></a>Switch for metric group collection and reporting. The default value is ON.</p>
        <a name="ul14220557134016"></a><a name="ul14220557134016"></a><ul id="ul14220557134016"><li>ON: Enabled. After the switch for the corresponding metric group is turned on, the metrics of this metric group will be collected and reported.</li><li>OFF: Disabled. After the switch for the corresponding metric group is turned off, the metrics of this metric group will not be collected or reported.</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4. <a name="li151815494115"></a>Press `Esc` and enter `:wq!` to save and exit.
    5. Refer to [2.b](#li11364381194) to [2.d](#li151815494115), modify the `pluginConfiguration.json` file, and configure the switch for custom metric group collection and reporting as needed.

        <a name="table970154420512"></a>
        <table><thead align="left"><tr id="row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="p15701444553"><a name="p15701444553"></a><a name="p15701444553"></a>Parameter</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="p117011441156"><a name="p117011441156"></a><a name="p117011441156"></a>Description</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p170118446517"><a name="p170118446517"></a><a name="p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p9568105120719"><a name="p9568105120719"></a><a name="p9568105120719"></a>Name of the custom metric group registered with <span id="ph18671058476"><a name="ph18671058476"></a><a name="ph18671058476"></a>NPU Exporter</span>. For details about custom metrics, see <a href="../../../appendix.md#custom-metric-development">Custom Metric Development</a>.</p>
        </td>
        </tr>
        <tr id="row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="p157021644755"><a name="p157021644755"></a><a name="p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="p2148170172513"><a name="p2148170172513"></a><a name="p2148170172513"></a>Switch for metric group collection and reporting. The default value is OFF.</p>
        <a name="ul1870217441514"></a><a name="ul1870217441514"></a><ul id="ul1870217441514"><li>ON: Enabled. After the switch for the corresponding metric group is turned on, the metrics of this metric group will be collected and reported.</li><li>OFF: Disabled. After the switch for the corresponding metric group is turned off, the metrics of this metric group will not be collected or reported.</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6. If custom metrics are developed through the plugin method, you need to rebuild and compile the binary file.
    7. See [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to remake and distribute the image.

3. Check whether the NPU Exporter image and version number are correct.
    - **Docker scenario**: Run the following command.

        ```shell
        docker images | grep npu-exporter
        ```

        Command output:

        ```ColdFusion
        npu-exporter                         v26.0.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    - **Containerd scenario**: Run the following command.

        ```shell
        ctr -n k8s.io c ls | grep npu-exporter
        ```

        Command output:

        ```ColdFusion
        docker.io/library/npu-exporter:v26.0.0                                                         application/vnd.docker.distribution.manifest.v2+json      sha256:38fd69ee9f5753e73a55a216d039f6ed4ea8a5de15c0e6b3bb503022db470c7b 91.5 MiB  linux/arm64
        ```

    - If correct, perform [Step 4](#li0640635114211).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

4. <a name="li0640635114211"></a>Copy the YAML files in the NPU Exporter software package decompression directory to any directory on the K8s management node.
5. Perform the following steps based on the actual image mode used.
    - **Containerd**: Set `containerMode` to `containerd` and modify the bold code below.

        If the default NPU Exporter startup parameter `-containerMode=docker` is used, you can skip this step.

        <pre codetype="yaml">
        apiVersion: apps/v1
        kind: DaemonSet
        metadata:
          name: npu-exporter
          namespace: npu-exporter
        spec:
          selector:
            matchLabels:
              app: npu-exporter
        ...
            spec:
        ...
              args: [ "umask 027;npu-exporter -port=8082 -ip=0.0.0.0  -updateTime=5
                         -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log -logLevel=0 <strong>-containerMode=containerd</strong>" ]
        ...
              volumeMounts:
        ...
                - name: docker-shim
                  mountPath: /var/run/dockershim.sock
                  readOnly: true
                <strong>- name: docker                                       # Delete when using containerd only</strong>
                  <strong>mountPath: /var/run/docker</strong>
                  <strong>readOnly: true</strong>
                - name: cri-dockerd
                  mountPath: /var/run/cri-dockerd.sock
                  readOnly: true
                - name: containerd
                  mountPath: /run/containerd
                  readOnly: true
                - name: isulad
                  mountPath: /run/isulad.sock
                  readOnly: true
        ...
              volumes:
        ...
                - name: docker-shim
                  hostPath:
                    path: /var/run/dockershim.sock
                <strong>- name: docker                                # Delete this when using only containerd</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/docker</strong>
                - name: cri-dockerd
                  hostPath:
                    path: /var/run/cri-dockerd.sock
                - name: containerd
                  hostPath:
                    path: /run/containerd
                - name: isulad
                  hostPath:
                    path: /run/isulad.sock

        ...</pre>

    - **Docker**: Delete the mount files of the original container runtime, add the mount directory for the `dockershim.sock` file, and modify the bolded code below.

        If the NPU Exporter startup parameter `-containerMode=containerd` is used, you can skip this step.

        >[!NOTE]
        >This step can effectively resolve the issue of NPU Exporter data loss caused by kubelet restart. After adding the mount directory, many mount files, such as `docker.sock`, will also be added, which poses a risk of container escape.

        <pre codetype="yaml">
        ...
                volumeMounts:
                  - name: log-npu-exporter
        ...
                  - name: sys
                    mountPath: /sys
                    readOnly: true
                  <strong>- name: docker-shim                        # Delete the following bold fields</strong>
                    <strong>mountPath: /var/run/dockershim.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: docker</strong>
                    <strong>mountPath: /var/run/docker</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: cri-dockerd</strong>
                    <strong>mountPath: /var/run/cri-dockerd.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: sock                   # Add the following bold fields</strong>
                    <strong>mountPath: /var/run        # Subject to the actual dockershim.sock file directory</strong>
                  - name: containerd
                    mountPath: /run/containerd
        ...
              volumes:
                - name: log-npu-exporter
        ...
                - name: sys
                  hostPath:
                    path: /sys
                <strong>- name: docker-shim                    # Delete the following bold fields</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/dockershim.sock</strong>
                <strong>- name: docker</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/docker</strong>
                <strong>- name: cri-dockerd</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/cri-dockerd.sock</strong>
                <strong>- name: sock                 # Add the following bold fields</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run                    # Use the actual dockershim.sock file directory</strong>
                - name: containerd
                  hostPath:
                    path: /run/containerd
         ...</pre>

6. If you do not need to modify other startup parameters of the component, you can skip this step. Otherwise, modify the startup parameters of NPU Exporter in the YAML file based on the actual situation. For details about the startup parameters, see [Table 2](#table872410431914). You can also run `./npu-exporter -h` to view the parameter description.
7. In the path where the YAML file is located on the management node, run the following command to start NPU Exporter.

    - If the K8s cluster uses Atlas 200I SoC A1 core board nodes, run the following command.

        ```shell
        kubectl apply -f npu-exporter-310P-1usoc-v{version}.yaml
        ```

    - If the K8s cluster uses node types other than the Atlas 200I SoC A1 core board, run the following command.

        ```shell
        kubectl apply -f npu-exporter-v{version}.yaml
        ```

    Example:

    ```ColdFusion
    namespace/npu-exporter created
    networkpolicy.networking.K8s.io/exporter-network-policy created
    daemonset.apps/npu-exporter created
    ```

    >[!NOTE]
    >When NPU Exporter is started, if the error `Error from server (NotFound): error when creating "npu-exporter-x.x.x.yaml":namespaces "npu-exporter" not found` appears, it indicates that the namespace for NPU Exporter was not created successfully. You need to run the following command to create it manually.
    >
    >```shell
    >kubectl create ns npu-exporter
    >```

8. Run the following command on any node to check whether the component starts successfully.

    ```shell
    kubectl get pod -n npu-exporter
    ```

    The command output is as follows. **Running** indicates that the component starts successfully. If the status is **CrashLoopBackOff**, it may be caused by incorrect directory permissions. For details, see [NPU Exporter Dynamic Path Check Fails, Log Shows check uid or mode failed](https://gitcode.com/Ascend/mind-cluster/issues/350).

    ```ColdFusion
    NAME                            READY   STATUS    RESTARTS   AGE
    ...
    npu-exporter-hqpxl        1/1    Running   0        11s
    ```

    >[!NOTE]
    >
    >- The use of NPU Exporter has requirements for the process environment. When running in Image Mode, ensure that the `/sys` directory and the container runtime communication socket file are mounted to the NPU Exporter container. If NPU container-related information is not obtained by calling the Metrics API of NPU Exporter, the issue may be caused by an incorrect socket file path. For details, see [Log Shows connecting to container runtime failed](https://gitcode.com/Ascend/mind-cluster/issues/346).
    >- After installing the component, if the Pod status of the component is not `Running`, see [Component Pod Status Is Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
    >- After installing the component, if the Pod status of the component is `ContainerCreating`, see [Cluster Scheduling Component Pod Is in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343).
    >- If the component fails to start, see [Cluster Scheduling Component Fails to Start, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
    >- If the component starts successfully but the corresponding Pod cannot be found, see [Component Startup YAML Executes Successfully but the Corresponding Pod Cannot Be Found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Run in Binary Mode<a name="section103551921135917"></a>

When NPU Exporter runs in Image Mode, it requires a privileged container, the `root` user, and the mounting of the docker-shim or Containerd socket file. If the container is maliciously exploited, there is a risk of container escape. When high security is required, you can run it directly on the physical machine in Binary Mode.

>[!NOTE]
>
>- When deploying NPU Exporter in Binary Mode, you can use a non-root user (for example, `hwMindX`) for deployment. Change the permission of the log directory to `hwMindX`. The command example is as follows: `chown <i>hwMindX:hwMindX</i> /var/log/mindx-dl/npu-exporter`.
>- The users in the following steps are all `hwMindX`.

1. Log in to the server as the `root` user.
2. Upload the NPU Exporter software package to any directory on the server (for example, `/home/ascend-npu-exporter`) and decompress it.
3. Copy the `metricConfiguration.json` and `pluginConfiguration.json` files from the decompressed directory of the NPU Exporter software package to the `/usr/local` directory.
4. (Optional) Modify the `metricConfiguration.json` or `pluginConfiguration.json` file to configure the switches for collecting and reporting default metric groups or custom metric groups.
    1. Go to the /usr/local directory.
    2. <a name="li1445835411478"></a>Open the `metricConfiguration.json` file.

        ```shell
        vi metricConfiguration.json
        ```

    3. Press `i` to enter the insert mode and configure the switches for collecting and reporting default metric groups based on the actual requirements.

        <a name="zh-cn_topic_0000002511426331_table192202574406"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row152204575408"><th class="cellrowborder" valign="top" width="30.12%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p1220125712404"><a name="zh-cn_topic_0000002511426331_p1220125712404"></a><a name="zh-cn_topic_0000002511426331_p1220125712404"></a>Parameter</p>
        </th>
        <th class="cellrowborder" valign="top" width="69.88%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p622019575401"><a name="zh-cn_topic_0000002511426331_p622019575401"></a><a name="zh-cn_topic_0000002511426331_p622019575401"></a>Description</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row182201357164014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p152201573404"><a name="zh-cn_topic_0000002511426331_p152201573404"></a><a name="zh-cn_topic_0000002511426331_p152201573404"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p222035704018"><a name="zh-cn_topic_0000002511426331_p222035704018"></a><a name="zh-cn_topic_0000002511426331_p222035704018"></a>Default metric group name.</p>
        <a name="zh-cn_topic_0000002511426331_ul222055714012"></a><a name="zh-cn_topic_0000002511426331_ul222055714012"></a><ul id="zh-cn_topic_0000002511426331_ul222055714012"><li>ddr: DDR data information</li><li>hccs: HCCS data information</li><li>npu: NPU data information</li><li>network: Network data information</li><li>pcie: PCIe data information</li><li>roce: RoCE data information</li><li>sio: SIO data information</li><li>vnpu: vNPU data information</li><li>version: Version data information</li><li>optical: Optical module data information</li><li>hbm: On-chip memory data information</li><li>ub: NPU UB data information</li></ul>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row5220257114014"><td class="cellrowborder" valign="top" width="30.12%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p182201657134015"><a name="zh-cn_topic_0000002511426331_p182201657134015"></a><a name="zh-cn_topic_0000002511426331_p182201657134015"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="69.88%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p722015718403"><a name="zh-cn_topic_0000002511426331_p722015718403"></a><a name="zh-cn_topic_0000002511426331_p722015718403"></a>Switch for metric group collection and reporting. The default value is ON.</p>
        <a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><a name="zh-cn_topic_0000002511426331_ul14220557134016"></a><ul id="zh-cn_topic_0000002511426331_ul14220557134016"><li>ON: Enabled. After the switch for the corresponding metric group is turned on, the metrics of this metric group will be collected and reported.</li><li>OFF: Disabled. After the switch for the corresponding metric group is turned off, the metrics of this metric group will not be collected or reported.</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    4. <a name="li18459954104718"></a>Press `Esc` and enter `:wq!` to save and exit.
    5. Refer to [4.b](#li1445835411478) to [4.d](#li18459954104718), modify the `pluginConfiguration.json` file, and configure the switch for custom metric group collection and reporting as needed.

        <a name="table16459165464719"></a>
        <table><thead align="left"><tr id="zh-cn_topic_0000002511426331_row157015443510"><th class="cellrowborder" valign="top" width="23.14%" id="mcps1.1.3.1.1"><p id="zh-cn_topic_0000002511426331_p15701444553"><a name="zh-cn_topic_0000002511426331_p15701444553"></a><a name="zh-cn_topic_0000002511426331_p15701444553"></a>Parameter</p>
        </th>
        <th class="cellrowborder" valign="top" width="76.86%" id="mcps1.1.3.1.2"><p id="zh-cn_topic_0000002511426331_p117011441156"><a name="zh-cn_topic_0000002511426331_p117011441156"></a><a name="zh-cn_topic_0000002511426331_p117011441156"></a>Description</p>
        </th>
        </tr>
        </thead>
        <tbody><tr id="zh-cn_topic_0000002511426331_row47010440518"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p170118446517"><a name="zh-cn_topic_0000002511426331_p170118446517"></a><a name="zh-cn_topic_0000002511426331_p170118446517"></a>metricsGroup</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p9568105120719"><a name="zh-cn_topic_0000002511426331_p9568105120719"></a><a name="zh-cn_topic_0000002511426331_p9568105120719"></a>Name of the custom metric group registered with <span id="zh-cn_topic_0000002511426331_ph18671058476"><a name="zh-cn_topic_0000002511426331_ph18671058476"></a><a name="zh-cn_topic_0000002511426331_ph18671058476"></a>NPU Exporter</span>. For details about how to customize metrics, see <a href="../../../appendix.md#custom-metric-development">Custom Metric Development</a>.</p>
        </td>
        </tr>
        <tr id="zh-cn_topic_0000002511426331_row157010441654"><td class="cellrowborder" valign="top" width="23.14%" headers="mcps1.1.3.1.1 "><p id="zh-cn_topic_0000002511426331_p157021644755"><a name="zh-cn_topic_0000002511426331_p157021644755"></a><a name="zh-cn_topic_0000002511426331_p157021644755"></a>state</p>
        </td>
        <td class="cellrowborder" valign="top" width="76.86%" headers="mcps1.1.3.1.2 "><p id="zh-cn_topic_0000002511426331_p2148170172513"><a name="zh-cn_topic_0000002511426331_p2148170172513"></a><a name="zh-cn_topic_0000002511426331_p2148170172513"></a>Switch for metric group collection and reporting. The default value is OFF.</p>
        <a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><a name="zh-cn_topic_0000002511426331_ul1870217441514"></a><ul id="zh-cn_topic_0000002511426331_ul1870217441514"><li>ON: Enabled. After the switch for the corresponding metric group is turned on, the metrics of this metric group will be collected and reported.</li><li>OFF: Disabled. After the switch for the corresponding metric group is turned off, the metrics of this metric group will not be collected or reported.</li></ul>
        </td>
        </tr>
        </tbody>
        </table>

    6. If custom metrics are developed through the plugin method, you need to rebuild and compile the binary file.

5. Create and edit a `npu-exporter.service` file.
    1. Run the following command to create a `npu-exporter.service` file.

        ```shell
        vi /home/ascend-npu-exporter/npu-exporter.service
        ```

    2. Write the following content into the `npu-exporter.service` file.

        <pre>
        [Unit]
        Description=Ascend npu exporter
        Documentation=hiascend.com

        [Service]
        ExecStart=/bin/bash -c "/usr/local/bin/npu-exporter -ip=127.0.0.1 -port=8082 -logFile=/var/log/mindx-dl/npu-exporter/npu-exporter.log>/dev/null  2>&1 &"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking
        User=hwMindX
        Group=hwMindX

        [Install]
        WantedBy=multi-user.target</pre>

        By default, NPU Exporter listens only on 127.0.0.1. You can modify the IP address to listen on by changing the startup parameter "-ip" and the "ExecStart" field in the "npu-exporter.service" file.

    3. Press `Esc` and enter `:wq!` to save and exit.

6. Create and edit the `npu-exporter.timer file`. Configuring a delayed start via the timer ensures that the NPUs are ready when NPU Exporter starts.
    1. Run the following command to create a `npu-exporter.timer` file.

        ```shell
         vi /home/ascend-npu-exporter/npu-exporter.timer
        ```

    2. Refer to the following example and write it into the `npu-exporter.timer` file.

        <pre>
        [Unit]
        Description=Timer for NPU Exporter Service

        [Timer]
        OnBootSec=60s            # Set the delayed startup time for NPU Exporter. Adjust it based on the actual situation.
        Unit=npu-exporter.service

        [Install]
        WantedBy=timers.target</pre>

    3. Press `Esc` and enter `:wq!` to save and exit.

7. If the deployment node is an Atlas 200I SoC A1 core board, run the following commands in sequence to add the `hwMindX` user to the `HwBaseUser` and `HwDmUser` user groups on the node. Users who are not using the Atlas 200I SoC A1 core board can skip this step.

    ```shell
    usermod -a -G HwBaseUser hwMindX
    usermod -a -G HwDmUser hwMindX
    ```

8. Run the following commands in sequence to enable the NPU Exporter service.

    ```shell
    cd /home/ascend-npu-exporter
    cp npu-exporter /usr/local/bin
    cp npu-exporter.service /etc/systemd/system
    chattr +i /etc/systemd/system/npu-exporter.service
    cp npu-exporter.timer /etc/systemd/system
    chattr +i /etc/systemd/system/npu-exporter.timer
    chmod 500 /usr/local/bin/npu-exporter
    chown hwMindX:hwMindX /usr/local/bin/npu-exporter
    chattr +i /usr/local/bin/npu-exporter
    systemctl enable npu-exporter.timer
    systemctl start npu-exporter
    systemctl start npu-exporter.timer
    ```

    > [!NOTE]
    > If you need to obtain container-related data information, NPU Exporter requires temporary privilege escalation to establish connections with the CRI and OCI sockets. Run the following commands:
    >
    > ```shell
    > chattr -i /usr/local/bin/npu-exporter
    > setcap cap_setuid+ep /usr/local/bin/npu-exporter
    > chattr +i /usr/local/bin/npu-exporter
    > systemctl restart npu-exporter
    > ```

## Parameter Description<a name="section2042611570392"></a>

**Table 2** NPU Exporter startup parameters

<a name="table872410431914"></a>

|Parameter|Type|Default Value|Description|
|--|--|--|--|
|-port|int|8082|Listening port. Value range: 1025 to 40,000.|
|-updateTime|int|5|Information update interval, ranging from 1 to 60 seconds. If the interval is set too long, some containers with a lifespan shorter than the update interval may fail to report.|
|-ip|string|None|This parameter has no default value and must be configured.<p>Listening IP address, which must be in valid IPv4 or IPv6 format. Configuring it as 0.0.0.0 is not recommended on multi-NIC hosts.</p>|
|-version|bool|false|Whether to query the NPU Exporter version.<ul><li>true: Query.</li><li>false: Do not query.</li></ul>|
|-concurrency|int|5|HTTP service throttling limit. The default value is 5 concurrent requests. Value range: 1 to 512.|
|-logLevel|int|0|Log level:<ul><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>|
|-maxAge|int|7|Log backup retention period. Value range: 7 to 700, in days.|
|-logFile|string|/var/log/mindx-dl/npu-exporter/npu-exporter.log|Log file.<p>Automatic log rotation is triggered when a single log file exceeds 20 MB. The maximum file size cannot be modified. The naming format for rotated files is: npu-exporter-<rotation trigger time/>.log, for example, npu-exporter-2023-10-07T03-38-24.402.log.</p>|
|-maxBackups|int|30|Maximum number of rotated log files to retain. Value range: 1 to 30, in units.|
|-containerMode|string|docker|Sets the container runtime type.<ul><li>Set to docker to indicate that the current environment uses Docker as the container runtime.</li><li>Set to containerd to indicate that the current environment uses Containerd as the container runtime.</li><li>Set to isula to indicate that the current environment uses iSula as the container runtime.</li></ul>|
|-containerd|string|<ul><li>(Docker) unix:///run/docker/containerd/docker-containerd.sock</li><li>(Containerd) unix:///run/containerd/containerd.sock</li><li>(iSula) unix:///run/isulad.sock</li></ul>|Endpoint of the containerd daemon process, used for communication with Containerd.<ul><li>If containerMode=docker, the default value is /run/docker/containerd/docker-containerd.sock. If the connection fails, it automatically attempts to connect to unix:///run/containerd/containerd.sock and unix:///run/docker/containerd/containerd.sock.</li><li>If containerMode=containerd, the default value is /run/containerd/containerd.sock.</li><li>If containerMode=isula, the default value is /run/isulad.sock.</li></ul><p>In most cases, the default value can be used. If you have modified the sock file path of Containerd, you need to modify the path accordingly.</p><p>You can run the **ps aux \| grep containerd** command to check whether the sock file path of Containerd has been modified.</p>|
|-endpoint|string|<ul><li>(Docker) unix:///var/run/dockershim.sock</li><li>(Containerd) unix:///run/containerd/containerd.sock</li><li>(iSula) unix:///run/isulad.sock</li></ul>|Sock address of the CRI server:<ul><li>If containerMode=docker, it connects to Dockershim to obtain the container list. The default value is /var/run/dockershim.sock.</li><li>If containerMode=containerd, the default value is /run/containerd/containerd.sock.</li><li>If containerMode=isula, the default value is /run/isulad.sock.</li></ul><p>In most cases, the default value can be used unless you have modified the sock file path of Dockershim or Containerd.</p><p>If the connection fails, it automatically attempts to connect to unix:///run/cri-dockerd.sock.</p>|
|-limitIPConn|int|5|TCP connection limit per IP. Value range: 1 to 128.|
|-limitTotalConn|int|20|Total TCP connection limit for the program. Value range: 1 to 512.|
|-limitIPReq|string|20/1|Request limit per IP. 20/1 means 20 requests are limited per second. The values on both sides of the "/" support a maximum of three digits.|
|-cacheSize|int|102400|Limit on the number of cached keys. Value range: 1 to 1,024,000.|
|-h or -help|None|None|Displays help information.|
|-platform|string|Prometheus|Specifies the target platform.<ul><li>Prometheus: Connects to Prometheus.</li><li>Telegraf: Connects to Telegraf.</li></ul>|
|-poll_interval|duration(int)|1|Interval for Telegraf data reporting, in seconds. This parameter takes effect only when connecting to the Telegraf platform, meaning -platform=Telegraf must be specified; otherwise, this parameter does not take effect.|
|-profilingTime|int|200|Configures the PCIe bandwidth collection duration, in milliseconds. Value range: 1 to 2,000.|
|-hccsBWProfilingTime|int|200|HCCS link bandwidth sampling duration. Value range: 1 to 1,000, in milliseconds.|
|-deviceResetTimeout|int|60|Maximum waiting time for the driver to report complete chips if the number of chips is insufficient during component startup, in seconds. Value range: 10 to 600.<ul><li>Atlas A2 training series products, Atlas 800I A2 inference servers, A200I A2 Box heterogeneous components: Recommended value is 150 seconds.</li><li>Atlas A3 training series products, A200T A3 Box8 super node servers, Atlas 800I A3 super node servers: Recommended value is 360 seconds.</li><li>Atlas 350 standard cards, Atlas 850 series hardware products, Atlas 950 SuperPoD: Recommended value is 600 seconds.</li></ul>|
|-textMetricsFilePath|string|None|Specifies the file path for custom metric files. For details about its constraints, see [Constraint Description](../../../api/npu_exporter/03_custom_metrics_file.md).|
