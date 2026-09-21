# NPU Exporter<a name="ZH-CN_TOPIC_0000002511426331"></a>

<!-- md-trans-meta sourceCommit=a277c409db3c3340f95d7c4831c0d54fa24e71a7 translatedAt=2026-09-02T01:25:38.351Z pushedAt=2026-09-02T01:30:48.464Z -->

- When using **resource monitoring**, NPU Exporter must be installed. This component supports integration with Prometheus or Telegraf.
    - When integrating with Prometheus, NPU Exporter can be deployed in either image mode or binary mode. For deployment differences, see [Image and Binary Deployment Differences](../../../07_references/05_appendix.md#differences-between-image-and-binary-deployments).
    - When integrating with Telegraf, see the [Working with Telegraf](../../../04_usage/01_resource_monitoring/03_working_with_telegraf.md) section to install NPU Exporter and Telegraf.

- Users who do not use **resource monitoring** can skip installing NPU Exporter and proceed directly past this chapter.

## Constraints<a name="section1362795652416"></a>

Before installing NPU Exporter, read the relevant constraints described in [Table 1](#table105071852271).

**Table 1** Constraints

<a name="table105071852271"></a>
<table><thead align="left"><tr id="row2050719520272"><th class="cellrowborder" valign="top" width="29.970000000000002%" id="mcps1.2.3.1.1"><p id="p1950795152711"><a name="p1950795152711"></a><a name="p1950795152711"></a>Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="70.03%" id="mcps1.2.3.1.2"><p id="p75071151277"><a name="p75071151277"></a><a name="p75071151277"></a>Constraints</p>
</th>
</tr>
</thead>
<tbody><tr id="row115077513271"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p17925222411"><a name="p17925222411"></a><a name="p17925222411"></a>NPU driver</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p450745142712"><a name="p450745142712"></a><a name="p450745142712"></a><span id="ph10112356112713"><a name="ph10112356112713"></a><a name="ph10112356112713"></a>NPU Exporter</span> periodically invokes the NPU driver's related interfaces to detect the NPU status. To upgrade the driver, stop tasks first, and then stop the <span id="ph154413248375"><a name="ph154413248375"></a><a name="ph154413248375"></a>NPU Exporter</span> container service.</p>
<div class="note" id="note1993172317415"><a name="note1993172317415"></a><a name="note1993172317415"></a><span class="notetitle">Note</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226378_p18934232419"><a name="zh-cn_topic_0000002479226378_p18934232419"></a><a name="zh-cn_topic_0000002479226378_p18934232419"></a>To ensure that <span id="zh-cn_topic_0000002479226378_ph7206429154119"><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a><a name="zh-cn_topic_0000002479226378_ph7206429154119"></a>NPU Exporter</span> can be installed by a non-root user (such as "hwMindX") when deployed in binary mode, use the "--install-for-all" parameter when installing the driver. An example is as follows.<pre class="screen" id="zh-cn_topic_0000002479226378_screen15239164112445"><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a><a name="zh-cn_topic_0000002479226378_screen15239164112445"></a>./Ascend-hdk-&lt;chip_type&gt;-npu-driver_&lt;version&gt;_linux-&lt;arch&gt;.run --full --install-for-all</pre>
</div>
</div></div>
</td>
</tr>
<tr id="row54685525282"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p5249201634114"><a name="p5249201634114"></a><a name="p5249201634114"></a><span id="ph1461172794116"><a name="ph1461172794116"></a><a name="ph1461172794116"></a>K8s</span> version</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p5468852142813"><a name="p5468852142813"></a><a name="p5468852142813"></a>Before using <span id="ph98079531286"><a name="ph98079531286"></a><a name="ph98079531286"></a>NPU Exporter</span>, check the <span id="ph18807253152810"><a name="ph18807253152810"></a><a name="ph18807253152810"></a>K8s</span> version of the environment. If the <span id="ph6808453102813"><a name="ph6808453102813"></a><a name="ph6808453102813"></a>K8s</span> version is 1.24.x or later, you need to <a href="https://github.com/mirantis/cri-dockerd#build-and-install" target="_blank" rel="noopener noreferrer">install cri-dockerd</a> as a dependency.</p>
</td>
</tr>
<tr id="row7507135142716"><td class="cellrowborder" rowspan="3" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p45071516276"><a name="p45071516276"></a><a name="p45071516276"></a>DCMI dynamic library</p>
<p id="p14507145152714"><a name="p14507145152714"></a><a name="p14507145152714"></a></p>
<p id="p9507651272"><a name="p9507651272"></a><a name="p9507651272"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p6555101612381"><a name="p6555101612381"></a><a name="p6555101612381"></a>The directory permission requirements for the DCMI dynamic library are as follows:</p>
<p id="p950745102715"><a name="p950745102715"></a><a name="p950745102715"></a><span id="ph1496251019288"><a name="ph1496251019288"></a><a name="ph1496251019288"></a>NPU Exporter</span> calls the DCMI dynamic library. All parent directories of the library must be owned by root, and programs owned by other users cannot run. In addition, these files and their directories must not have write permission for "group" and "other".</p>
</td>
</tr>
<tr id="row1650710572715"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p195079518272"><a name="p195079518272"></a><a name="p195079518272"></a>The path depth of the DCMI dynamic library must be less than 20.</p>
</td>
</tr>
<tr id="row35071553276"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p18507205192711"><a name="p18507205192711"></a><a name="p18507205192711"></a>If the dynamic library path is set through "LD_LIBRARY_PATH", the total length of "LD_LIBRARY_PATH" cannot exceed 1024.</p>
</td>
</tr>
<tr id="row75074519275"><td class="cellrowborder" rowspan="2" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p050719519271"><a name="p050719519271"></a><a name="p050719519271"></a><span id="ph13135203152812"><a name="ph13135203152812"></a><a name="ph13135203152812"></a>Atlas 200I SoC A1 core board</span></p>
<p id="p35076552719"><a name="p35076552719"></a><a name="p35076552719"></a></p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p209012054192411"><a name="p209012054192411"></a><a name="p209012054192411"></a>To use <span id="ph56561935182816"><a name="ph56561935182816"></a><a name="ph56561935182816"></a>NPU Exporter</span> on the <span id="ph1865633562811"><a name="ph1865633562811"></a><a name="ph1865633562811"></a>Atlas 200I SoC A1 core board</span>, ensure that the NPU driver of the <span id="ph10656153513282"><a name="ph10656153513282"></a><a name="ph10656153513282"></a>Atlas 200I SoC A1 core board</span> is version 23.0.RC2 or later.</p>
</td>
</tr>
<tr id="row165073518272"><td class="cellrowborder" valign="top" headers="mcps1.2.3.1.1 "><p id="p95251515257"><a name="p95251515257"></a><a name="p95251515257"></a>To deploy the <span id="ph19614124172819"><a name="ph19614124172819"></a><a name="ph19614124172819"></a>NPU Exporter</span> on an <span id="ph136141041142813"><a name="ph136141041142813"></a><a name="ph136141041142813"></a>Atlas 200I SoC A1 core board</span> node in image mode, configure the multi-container sharing mode.</p>
</td>
</tr>
<tr id="row1044710113298"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p1144701142912"><a name="p1144701142912"></a><a name="p1144701142912"></a>Virtual machine</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p14473110297"><a name="p14473110297"></a><a name="p14473110297"></a>If the <span id="ph6368151492319"><a name="ph6368151492319"></a><a name="ph6368151492319"></a>NPU Exporter</span> is deployed in a virtual machine, systemd must be installed in the <span id="ph24388313372"><a name="ph24388313372"></a><a name="ph24388313372"></a>NPU Exporter</span> image. It is recommended to add the <strong id="b14813193310547"><a name="b14813193310547"></a><a name="b14813193310547"></a>RUN apt-get update &amp;&amp; apt-get install -y systemd</strong> command to the Dockerfile for installation.</p>
</td>
</tr>
<tr id="row_container_metrics"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p_container_metrics_cat">Container-related metrics</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><ul id="ol_container_metrics"><li>Container-related metrics are displayed only in the Prometheus scenario when NPUs are mounted to Pods in K8s. The Telegraf scenario does not support displaying container-related metrics.</li><li>In the Prometheus scenario, if NPUs are not mounted to Pods in K8s, the container_name, namespace, and pod_name labels of each metric are empty.</li><li>When parsing the NPU information mounted to a Pod, the "ASCEND_VISIBLE_DEVICE" environment variable in the container is parsed first. Ensure that this environment variable is not preset in the image to avoid interference.</li><li>Do not start service containers in privileged mode. If a service container is started in privileged mode, the NPUs actually used by service processes may be inconsistent with the NPUs allocated by MindCluster components.</li></ul>
</td>
</tr>
</tbody>
</table>

## Compatibility Notes

### 26.1.0 Compatibility Notes

- Configuration file path change: the paths of `metricConfiguration.json` and `pluginConfiguration.json` are changed from `/usr/local/` to `/user/mind-cluster/npu-exporter-config`. Users migrating from earlier versions (26.0 and earlier) need to copy the configuration files to `/user/mind-cluster/npu-exporter-config`.

- The following metrics add a `-1` status, indicating that the DCMI interface/hccn_tool invocation failed. For specific metric information, see [NPU Data Information](../../../06_api/00_npu_exporter/01_prometheus_metrics_api.md#section1379685784314):
    - `npu_chip_info_health_status`
    - `npu_chip_info_network_status`
    - `npu_chip_info_link_status`
    - `npu_chip_info_link_status_X_Y`

- Migrate the following metrics from the npu metric group to the utilization metric group, with a default collection interval of 1s. For details about the metrics, see [Utilization Data Information](../../../06_api/00_npu_exporter/01_prometheus_metrics_api.md#section1379685784315):
    - `npu_chip_info_utilization`
    - `container_npu_utilization`
    - `npu_chip_info_vector_utilization`
    - `npu_chip_info_cube_utilization`
    - `npu_chip_info_overall_utilization`

- The default `-updateTime=5` parameter is removed. The collection interval of each metric group is now controlled by the `intervalSeconds` field in the configuration file.

  If the `-updateTime` parameter is still configured and its value is within the valid range, that parameter value takes precedence.

- Adjust the default collection interval of each metric group from the original global 5s to the following:

  |Collection Interval|Metric Group|
  |------------|----------------|
  | Collected once  | version                                                |
  | 1s         | utilization                                            |
  | 5s         | npu                                                    |
  | 10s        | ddr                                                    |
  | 60s        | sio, hbm, hccs, pcie, vnpu, roce, optical, network, ub |
  | 86400s (one day) | nodeBase                                               |

## Procedure<a name="section83111543151612"></a>

NPU Exporter supports two installation modes. You can select either one based on actual conditions. This component provides only HTTP services. To use the more secure HTTPS service, modify the source code for adaptation.

- (Recommended) Run in image mode. For installation steps, see [Running in Image Mode](#section2035402135914).
- When security requirements are strict, it is recommended to run in binary mode on a physical machine. For installation steps, see [Running in Binary Mode](#section103551921135917).

## Running in Image Mode<a name="section2035402135914"></a>

1. Log in to each compute node as the root user.
2. (Optional) Modify the `metricConfiguration.json` or `pluginConfiguration.json` file to configure the collection switch and collection interval of the default metric group or custom metric group.
    1. Enter the NPU Exporter package extraction directory.
    2. <a name="li11364381194"></a>Open the `metricConfiguration.json` file.

        ```shell
        vi metricConfiguration.json
        ```

    3. Press `i` to enter edit mode, and configure the collection switch and collection interval of the default metric group as needed.

        The following is an example of the configuration file:

        ```json
            [
                {"metricsGroup": "version", "state": "ON", "intervalSeconds": -1},
                {"metricsGroup": "utilization", "state": "ON", "intervalSeconds": 1},
                {"metricsGroup": "npu", "state": "ON", "intervalSeconds": 5},
                {"metricsGroup": "ddr", "state": "ON", "intervalSeconds": 10},
                {"metricsGroup": "sio", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hbm", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hccs", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "pcie", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "vnpu", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "nodeBase", "state": "ON", "intervalSeconds": 86400},

                {"metricsGroup": "roce", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "optical", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "network", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "ub", "state": "ON", "intervalSeconds": 60}
            ]
        ```

        <a name="table192202574406"></a>

        |Name|Description|
        |---|---|
        |metricsGroup|Name of the default metric group.<ul><li>Collected through DCMI:<ul><li>version: version data information</li><li>utilization: utilization data information</li><li>npu: NPU data information</li><li>ddr: DDR data information</li><li>sio: SIO data information</li><li>hbm: on-chip memory data information</li><li>hccs: HCCS data information</li><li>pcie: PCIe data information</li><li>vnpu: vNPU data information</li><li>nodeBase: basic node information</li></ul></li><li>Collected through hccn_tool:<ul><li>roce: RoCE data information</li><li>optical: optical module data information</li><li>network: Network data information</li><li>ub: NPU UB data information</li></ul></li></ul>|
        |state|Switch for metric group collection and reporting. The default value is ON.<ul><li>ON: enabled. After the switch of the corresponding metric group is enabled, the metrics of this metric group are collected and reported.</li><li>OFF: disabled. After the switch of the corresponding metric group is disabled, the metrics of this metric group are not collected or reported.</li></ul>|
        |intervalSeconds|Collection interval of the metric group, in seconds.<ul><li>Must be configured as an integer value.</li><li>Value range: -1, 1 to 86400 seconds.</li><li>If this configuration item is missing, the default value of 60 seconds is used.</li><li>If configured as -1, the metric group is collected only once and is not collected repeatedly.</li></ul>|

    4. <a name="li151815494115"></a>Press `Esc`, type `:wq!` to save and exit.
    5. Refer to [2.b](#li11364381194) through [2.d](#li151815494115), modify the `pluginConfiguration.json` file, and configure the collection switch and collection interval of custom metric groups as needed.

        <a name="table970154420512"></a>

        |Name|Description|
        |---|---|
        |metricsGroup|Name of the custom metric group registered with NPU Exporter. For details about custom metrics, see [Custom Metrics Development](../../01_custom_metrics_development.md).|
        |state|Switch for metric group collection and reporting. The default value is OFF.<ul><li>ON: indicates enabled. After the switch of the corresponding metric group is enabled, the metrics of this metric group are collected and reported.</li><li>OFF: indicates disabled. After the switch of the corresponding metric group is disabled, the metrics of this metric group are not collected or reported.</li></ul>|
        |intervalSeconds|Collection interval of the custom metric group, in seconds.<ul><li>Must be configured as an integer value.</li><li>The value range is -1 and 1 to 86400 seconds.</li><li>If this configuration item is missing, the default value of 60 seconds is used.</li><li>If it is configured as -1, the metric group is collected only once and is not collected repeatedly.</li></ul>|

    6. If custom metrics are developed through the plugin method, the binary file must be rebuilt and recompiled.
    7. See [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to rebuild and distribute the image.

3. (Optional) Mount configuration files.

    There are three configuration file mount scenarios. For specific configuration methods, refer to the [Dynamic Configuration Loading Instructions](#dynamic-configuration-loading-instructions) section.

    - **Default**: By default, the configuration file is not mounted to the host. The built-in default configuration in the image is used, and **dynamic configuration modification is not supported**.
    - **HostPath mount**: Mount the configuration files on the host to the container at `/user/mind-cluster/npu-exporter-config/metricConfiguration.json` and `/user/mind-cluster/npu-exporter-config/pluginConfiguration.json`. This method supports independent configuration for each node, and can also achieve unified global configuration through a shared directory.
    - **ConfigMap mount**: Manage configuration files uniformly through the K8s ConfigMap, so that all nodes use the same configuration.

4. Based on the container runtime actually in use, verify that the NPU Exporter image and version number are correct.

    >[!NOTE]
    >You can run the `kubectl get nodes -o wide` command and check the output in the **CONTAINER-RUNTIME** column to determine the container runtime:
    >- If the output is `docker://xxx`, the Docker scenario applies.
    >- If the output is `containerd://xxx`, the Containerd scenario applies.

    - **Docker**: Run the following command.

        ```shell
        docker images | grep npu-exporter
        ```

        Command output:

        ```ColdFusion
        npu-exporter                         v26.1.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    - **Containerd**: Run the following command.

        ```shell
        ctr -n k8s.io c ls | grep npu-exporter
        ```

        Command output:

        ```ColdFusion
        docker.io/library/npu-exporter:v26.1.0                                                         application/vnd.docker.distribution.manifest.v2+json      sha256:38fd69ee9f5753e73a55a216d039f6ed4ea8a5de15c0e6b3bb503022db470c7b 91.5 MiB  linux/arm64
        ```

    - Yes, proceed to [Step 5](#li0640635114211).
    - No, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

5. <a name="li0640635114211"></a>Copy the YAML file in the NPU Exporter package extraction directory to any directory on the Kubernetes management node.
6. Select the following steps based on the container runtime actually in use.
    - **Containerd**: Set `containerMode` to `containerd` and modify the following bolded code.

        If the default NPU Exporter startup parameter `-containerMode=docker` is used, skip this step.

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
                <strong>- name: docker                                       # Delete when containerd is used</strong>
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
                <strong>- name: docker                                # Delete when containerd is used</strong>
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

    - **Docker**: Delete the mount files of the original container runtime, add the mount directory for the `dockershim.sock` file, and modify the following bold code.

        If the NPU Exporter startup parameter `-containerMode=containerd` is used, you can skip this step.

        >[!NOTICE]
        >This step can effectively resolve the NPU Exporter data loss caused by kubelet restart. After adding the mount directory, many mount files, such as `docker.sock`, are also added, which poses a risk of container escape.

        <pre codetype="yaml">
        ...
                volumeMounts:
                  - name: log-npu-exporter
        ...
                  - name: sys
                    mountPath: /sys
                    readOnly: true
                  <strong>- name: docker-shim                        # Delete the fields in bold</strong>
                    <strong>mountPath: /var/run/dockershim.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: docker</strong>
                    <strong>mountPath: /var/run/docker</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: cri-dockerd</strong>
                    <strong>mountPath: /var/run/cri-dockerd.sock</strong>
                    <strong>readOnly: true</strong>
                  <strong>- name: sock                   # Add the fields in bold</strong>
                    <strong>mountPath: /var/run        # Use the actual dockershim.sock directory</strong>
                  - name: containerd
                    mountPath: /run/containerd
        ...
              volumes:
                - name: log-npu-exporter
        ...
                - name: sys
                  hostPath:
                    path: /sys
                <strong>- name: docker-shim                    # Delete the fields in bold</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/dockershim.sock</strong>
                <strong>- name: docker</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/docker</strong>
                <strong>- name: cri-dockerd</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run/cri-dockerd.sock</strong>
                <strong>- name: sock                 # dd the fields in bold</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /var/run                    # Use the actual dockershim.sock directory</strong>
                - name: containerd
                  hostPath:
                    path: /run/containerd
         ...</pre>

7. If you do not need to modify other startup parameters of the component, skip this step. Otherwise, modify the startup parameters of NPU Exporter in the YAML file as required. The startup parameters are described in [Table 2](#table872410431914). You can also run <b>./npu-exporter -h</b> to view the parameter description.
8. In the directory where the YAML file resides on the management node, run the following command to start NPU Exporter.

    - Nodes in the K8s cluster that use Atlas 200I SoC A1 core board, run the following command.

        ```shell
        kubectl apply -f npu-exporter-310P-1usoc-v{version}.yaml
        ```

    - Nodes in the K8s cluster that use other products, run the following command.

        ```shell
        kubectl apply -f npu-exporter-v{version}.yaml
        ```

    A startup example is as follows:

    ```ColdFusion
    namespace/npu-exporter created
    networkpolicy.networking.k8s.io/exporter-network-policy created
    daemonset.apps/npu-exporter created
    ```

    >[!NOTE]
    >The error "Error from server (NotFound): error when creating "npu-exporter-<i>x.x.x</i>.yaml":namespaces "npu-exporter" not found" displayed during NPU Exporter boot indicates that the namespace for NPU Exporter was not created successfully. Run the following command to create it manually.
    >
    >```shell
    >kubectl create ns npu-exporter
    >```

9. Run the following command on any node to check whether the component starts successfully.

    ```shell
    kubectl get pod -n npu-exporter
    ```

    An example of the output is as follows. **Running** indicates that the component starts successfully. If the status is **CrashLoopBackOff**, it may be caused by incorrect directory permissions. For details, see [NPU Exporter fails to check the dynamic path, and the log shows check uid or mode failed](https://gitcode.com/Ascend/mind-cluster/issues/350).

    ```ColdFusion
    NAME                            READY   STATUS    RESTARTS   AGE
    ...
    npu-exporter-hqpxl        1/1    Running   0        11s
    ```

    >[!NOTE]
    >
    >- The use of NPU Exporter has requirements on the process environment. When running in image mode, ensure that the "/sys" directory and the container runtime communication socket file are mounted into the NPU Exporter container. If the NPU container information is not obtained by calling the Metrics API of NPU Exporter, the issue may be caused by an incorrect socket file path. For details, see [The log shows connecting to container runtime failed](https://gitcode.com/Ascend/mind-cluster/issues/346).
    >- After the component is installed, if the Pod status is not Running, see [Component Pod status is not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
    >- After the component is installed, if the Pod status is ContainerCreating, see [Cluster scheduling component Pod is in ContainerCreating state](https://gitcode.com/Ascend/mind-cluster/issues/343).
    >- If the component fails to start, see [Cluster scheduling component fails to start, and the log prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
    >- If the component starts successfully but the corresponding Pod cannot be found, see [Component startup YAML is executed successfully, but the corresponding Pod cannot be found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Binary Mode<a name="section103551921135917"></a>

When NPU Exporter runs in image mode, it requires a privileged container, the root user, and the socket file of docker-shim or Containerd to be mounted. If the container is maliciously exploited, there is a risk of container escape. When higher security is required, you can run it directly on the physical machine in binary mode.

>[!NOTE]
>
>- When deploying NPU Exporter in binary mode, you can use a non-root user (for example, `hwMindX`) for deployment. Change the permission of the log directory to hwMindX. The command example is as follows: **chown <i>hwMindX:hwMindX</i> /var/log/mindx-dl/npu-exporter**.
>- The user in the following steps is `hwMindX`.

1. Log in to the server as the `root` user.
2. Upload the NPU Exporter package to any directory on the server (for example, `/home/ascend-npu-exporter`) and extract it.
3. (Optional) Create the configuration file directory `/user/mind-cluster/npu-exporter-config` (this directory cannot be changed to another one), and copy the `metricConfiguration.json` and `pluginConfiguration.json` files from the NPU Exporter package extraction directory to this directory.

    ```shell
    mkdir -p /user/mind-cluster/npu-exporter-config
    cp metricConfiguration.json pluginConfiguration.json /user/mind-cluster/npu-exporter-config
    ```

4. (Optional) Modify the `metricConfiguration.json` or `pluginConfiguration.json` file to configure the collection switch and collection interval of the default metric group or custom metric group.
    1. Enter the `/user/mind-cluster/npu-exporter-config` directory.
    2. <a name="li1445835411478"></a>Open the `metricConfiguration.json` file.

        ```shell
        vi metricConfiguration.json
        ```

    3. Press `i` to enter edit mode, and configure the collection switch and collection interval of the default metric group as needed.

        The following is an example of the configuration file:

        ```json
            [
                {"metricsGroup": "version", "state": "ON", "intervalSeconds": -1},
                {"metricsGroup": "utilization", "state": "ON", "intervalSeconds": 1},
                {"metricsGroup": "npu", "state": "ON", "intervalSeconds": 5},
                {"metricsGroup": "ddr", "state": "ON", "intervalSeconds": 10},
                {"metricsGroup": "sio", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hbm", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "hccs", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "pcie", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "vnpu", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "nodeBase", "state": "ON", "intervalSeconds": 86400},

                {"metricsGroup": "roce", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "optical", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "network", "state": "ON", "intervalSeconds": 60},
                {"metricsGroup": "ub", "state": "ON", "intervalSeconds": 60}
            ]
        ```

        <a name="zh-cn_topic_0000002511426331_table192202574406"></a>

        |Name|Description|
        |---|---|
        |metricsGroup|Name of the default metric group.<ul><li>Collected through DCMI:<ul><li>version: version data information</li><li>utilization: utilization data information</li><li>npu: NPU data information</li><li>ddr: DDR data information</li><li>sio: SIO data information</li><li>hbm: on-chip memory data information</li><li>hccs: HCCS data information</li><li>pcie: PCIe data information</li><li>vnpu: vNPU data information</li><li>nodeBase: basic node information</li></ul></li><li>Collected through hccn_tool:<ul><li>roce: RoCE data information</li><li>optical: optical module data information</li><li>network: Network data information</li><li>ub: NPU UB data information</li></ul></li></ul>|
        |state|Switch for metric group collection and reporting. The default value is ON.<ul><li>ON: enabled. After the switch of the corresponding metric group is enabled, the metrics of this metric group are collected and reported.</li><li>OFF: disabled. After the switch of the corresponding metric group is disabled, the metrics of this metric group are not collected or reported.</li></ul>|
        |intervalSeconds|Collection interval of the metric group, in seconds.<ul><li>Must be configured as an integer value.</li><li>The value range is -1 and 1 to 86400 seconds.</li><li>If this configuration item is missing, the default value of 60 seconds is used.</li><li>If it is set to -1, the metric group is collected only once and is not collected repeatedly.</li></ul>|

    4. <a name="li18459954104718"></a>Press `Esc`, type `:wq!` to save and exit.
    5. Refer to [4.b](#li1445835411478) to [4.d](#li18459954104718), modify the `pluginConfiguration.json` file, and configure the collection switch and collection interval of custom metric groups as needed.

        <a name="table16459165464719"></a>

        |Name|Description|
        |---|---|
        |metricsGroup|Name of the custom metric group registered with NPU Exporter. For details about how to develop custom metrics, see [Custom Metrics Development](../../01_custom_metrics_development.md).|
        |state|Switch for metric group collection and reporting. The default value is OFF.<ul><li>ON: enabled. After the switch of the corresponding metric group is enabled, the metrics of that metric group are collected and reported.</li><li>OFF: disabled. After the switch of the corresponding metric group is disabled, the metrics of that metric group are not collected or reported.</li></ul>|
        |intervalSeconds|Collection interval of the custom metric group, in seconds.<ul><li>Must be configured as an integer value.</li><li>Value range: -1, or 1 to 86400 seconds.</li><li>If this configuration item is missing, the default value of 60 seconds is used.</li><li>If configured as -1, the metric group is collected only once and is not collected repeatedly.</li></ul>|

    6. If custom metrics are developed through the plugin method, the binary file must be rebuilt and recompiled.

5. Create and edit the `npu-exporter.service` file.
    1. Run the following command to create the `npu-exporter.service` file.

        ```shell
        vi /home/ascend-npu-exporter/npu-exporter.service
        ```

    2. Refer to the following content and write it into the `npu-exporter.service` file.

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

        By default, NPU Exporter listens only on 127.0.0.1. You can modify the IP address to listen on by changing the startup parameters `-ip` and the `ExecStart` field in the `npu-exporter.service` file.

    3. Press `Esc` and type `:wq!` to save and exit.

6. Create and edit the `npu-exporter.timer` file. By configuring a delayed startup through the timer, you can ensure that the NPU cards are in place when NPU Exporter starts.
    1. Run the following command to create the `npu-exporter.timer` file.

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

    3. Press `Esc`, type `:wq!` to save and exit.

7. If the deployment node is an Atlas 200I SoC A1 core board, run the following commands in sequence to add the `hwMindX` user to the `HwBaseUser` and `HwDmUser` user groups on the node. Users of non-Atlas 200I SoC A1 core boards can skip this step.

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
    >If you need to obtain container-related data information, NPU Exporter requires temporary privilege escalation to establish connections with the CRI and OCI sockets. Run the following commands:
    >
    >```shell
    >chattr -i /usr/local/bin/npu-exporter
    >setcap cap_setuid+ep /usr/local/bin/npu-exporter
    >chattr +i /usr/local/bin/npu-exporter
    >systemctl restart npu-exporter
    >```

## Parameter Description<a name="section2042611570392"></a>

**Table 2** NPU Exporter startup parameters

<a name="table872410431914"></a>

|Name|Type|Default Value|Description|
|--|--|--|--|
|-port|int|8082|Listening port. Value range: 1025 to 40000.|
|-updateTime|int|None|**To be sunset soon and not recommended.** Global configuration metric update interval. Value range: 1 to 60 seconds. It is recommended to configure the metric update interval by group. For details, see [Configuration File](#section103551921135917).<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody">If the updateTime parameter is configured, it remains effective and takes precedence over intervalSeconds in the metricConfiguration.json/pluginConfiguration.json configuration files.</div></div>|
|-ip|string|None|This parameter has no default value and must be configured.<p>Listening IP address. It must be in valid IPv4 or IPv6 format. On hosts with multiple NICs, configuring it as 0.0.0.0 is not recommended.</p>|
|-version|bool|false|Whether to query the NPU Exporter version number.<ul><li>true: Query.</li><li>false: Do not query.</li></ul>|
|-concurrency|int|5|Rate limiting size of the HTTP service. The default is 5 concurrent requests. Value range: 1 to 512.|
|-logLevel|int|0|Log level:<ul><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>|
|-maxAge|int|7|Log backup retention time. Value range: 7 to 700, in days.|
|-logFile|string|/var/log/mindx-dl/npu-exporter/npu-exporter.log|Log file.<p>When a single log file exceeds 20 MB, automatic rotation is triggered. The maximum file size cannot be modified. The naming format of rotated files is: npu-exporter-<i>{rotation_time}</i>.log, for example: npu-exporter-2023-10-07T03-38-24.402.log.</p>|
|-maxBackups|int|30|Maximum number of rotated log files to retain. Value range: 1 to 180, in files.|
|-containerMode|string|docker|Sets the container runtime type.<ul><li>Set to docker to indicate that the current environment uses Docker as the container runtime.</li><li>Set to containerd to indicate that the current environment uses Containerd as the container runtime.</li><li>Set to "isula" to indicate that the current environment uses iSula as the container runtime.</li></ul>|
|-containerd|string|<ul><li>(Docker)unix:/run/docker/containerd/docker-containerd.sock</li><li>(Containerd)unix:///run/containerd/containerd.sock</li><li>(iSula)unix:///run/isulad.sock</li></ul>|Endpoint of the containerd daemon process, used to communicate with Containerd.<ul><li>If containerMode=docker, the default value is /run/docker/containerd/docker-containerd.sock. If the connection fails, it automatically attempts to connect to unix:///run/containerd/containerd.sock and unix:///run/docker/containerd/containerd.sock.</li><li>If containerMode=containerd, the default value is /run/containerd/containerd.sock.</li><li>If containerMode=isula, the default value is /run/isulad.sock.</li></ul><p>In general, use the default value. If you have modified the sock file path of Containerd, modify the corresponding path accordingly.</p><p>You can run the **ps aux \| grep containerd** command to check whether the sock file path of Containerd has been modified.</p>|
|-endpoint|string|<ul><li>(Docker)unix:///var/run/dockershim.sock</li><li>(Containerd)unix:///run/containerd/containerd.sock</li><li>(iSula)unix:///run/isulad.sock</li></ul>|Sock address of the CRI server:<ul><li>If containerMode=docker, it connects to Dockershim to obtain the container list. The default value is /var/run/dockershim.sock.</li><li>If containerMode=containerd, the default value is /run/containerd/containerd.sock.</li><li>If containerMode=isula, the default value is /run/isulad.sock.</li></ul><p>In general, use the default value unless you have modified the sock file path of Dockershim or Containerd.</p><p>If the connection fails, it automatically attempts to connect to unix:///run/cri-dockerd.sock.</p>|
|-limitIPConn|int|5|Value range of the TCP connection limit per IP: 1 to 128.|
|-limitTotalConn|int|20|Value range of the total TCP connection limit of the program: 1 to 512.|
|-limitIPReq|string|20/1|Request limit per IP. 20/1 indicates a limit of 20 requests per second. Each side of "/" supports a maximum of three digits.|
|-cacheSize|int|102400|Limit on the number of cache keys. Value range: 1 to 1024000.|
|--enable-healthz|bool|false|Whether to enable the health check service. During K8s deployment, it is enabled (true) by the component YAML configuration.<ul><li>true: Enable.</li><li>false: Disable.</li></ul>|
|--healthz-address|string|11251|Listening port number of the health check service. Value range: 1025 to 65535. During K8s deployment, it is configured as 11256 by the component YAML. If the specified port is occupied, the component fails to start.|
|--tls-cert-file|string|""|Path of the HTTPS certificate file. If empty, the HTTP protocol is used. It must be configured together with --tls-private-key-file or both left empty. For the configuration method and security precautions, see [Health Probe Security Hardening](../../../07_references/04_security_hardening.md).|
|--tls-private-key-file|string|""|Path of the HTTPS private key file. If empty, the HTTP protocol is used. It must be configured together with --tls-cert-file or both left empty.|
|-h or -help|None|None|Displays help information.|
|-platform|string|Prometheus|Specifies the platform to integrate with.<ul><li>Prometheus: Integrate with Prometheus</li><li>Telegraf: Integrate with Telegraf</li></ul>|
|-poll_interval|duration(int)|1|Interval for Telegraf data reporting, in seconds. This parameter takes effect only when integrating with the Telegraf platform, that is, it takes effect only when -platform=Telegraf is specified. Otherwise, this parameter does not take effect.|
|-profilingTime|int|200|Configures the PCIe bandwidth collection time, in milliseconds. Value range: 1 to 2000.|
|-hccsBWProfilingTime|int|200|Sampling duration of the HCCS link bandwidth. Value range: 1 to 1000, in milliseconds.|
|-deviceResetTimeout|int|600|When the component starts, if the number of chips is insufficient, the maximum duration to wait for the driver to report the complete chips, in seconds. Value range: 10 to 600.<ul><li><term>Atlas A2 training products</term>, Atlas 800I A2 inference server, A200I A2 Box heterogeneous subrack: 150 seconds is recommended.</li><li><term>Atlas A3 training products</term>, A200T A3 Box8 SuperPoD server, Atlas 800I A3 SuperPoD server: 360 seconds is recommended.</li><li>Atlas 350 accelerator card, Atlas 850E SuperPoD, Atlas 650E server, Atlas 950 SuperPoD: 600 seconds is recommended.</li></ul>|
|-textMetricsFilePath|string|None|Specifies the path of the custom metrics file. For details about its constraints, see [Constraints](../../../06_api/00_npu_exporter/03_custom_metrics_file.md#constraints).|
|-enableLegacyMetrics|bool|false|Specifies whether to enable the legacy Prometheus network metric format of the Atlas 350 accelerator card. The default value is false, which means disabled.|

## Dynamic Configuration Loading Instructions<a name="dynamic-configuration-loading-instructions"></a>

NPU Exporter supports dynamically loading configuration files, allowing configuration changes to take effect without restarting the component.

### Binary Deployment

The configuration file paths are `/user/mind-cluster/npu-exporter-config/metricConfiguration.json` and `/user/mind-cluster/npu-exporter-config/pluginConfiguration.json`.

Modify these two files directly. NPU Exporter automatically detects file changes and reloads the configuration.

### K8s HostPath Mount

- Advantages of HostPath mount:
  - Configuration changes take effect immediately.
  - Each node can be configured independently, or a shared directory can be used to achieve unified global configuration.
- Disadvantages of HostPath mount: configuration changes are difficult to track.

Configure HostPath mount in the deployment YAML to mount the configuration file on the host to the `/user/mind-cluster/npu-exporter-config` path in the container:

1. Enter the NPU Exporter package extraction directory and prepare the configuration file.

    ```bash
    # Create a configuration directory on each node. The directory can be customized, as long as it is consistent with the mount path in the YAML.
    mkdir -p /user/mind-cluster/npu-exporter-config
    cp metricConfiguration.json /user/mind-cluster/npu-exporter-config
    cp pluginConfiguration.json /user/mind-cluster/npu-exporter-config
    ```

2. Mount to the Pod.

    ```yaml
    volumeMounts:
      - name: npu-config
        mountPath: /user/mind-cluster/npu-exporter-config
        readOnly: true

    volumes:
      - name: npu-config
        hostPath:
          path: /user/mind-cluster/npu-exporter-config
          type: DirectoryOrCreate
    ```

Modify the configuration file on the host. NPU Exporter automatically detects file changes and reloads the configuration.

### K8s ConfigMap Mount

- Advantages of ConfigMap mounting:
  - Centrally manages the configuration of all nodes and supports updating the configuration of all nodes with one click.
  - Configuration changes can be tracked and version-controlled.
- Disadvantages of ConfigMap mounting:
  - All nodes use the same configuration, and a single node cannot be configured independently.
  - Configuration changes take effect with a certain delay (the time for K8s to update the ConfigMap to the container).

1. Enter the NPU Exporter package extraction directory and create the ConfigMap.

    ```bash
    kubectl create ns npu-exporter
    kubectl create cm -n npu-exporter npu-exporter-metric-config \
      --from-file=metricConfiguration.json=./metricConfiguration.json \
      --from-file=pluginConfiguration.json=./pluginConfiguration.json
    ```

2. Mount it to the Pod.

    ```yaml
    volumeMounts:
      - name: npu-config
        mountPath: /user/mind-cluster/npu-exporter-config
        readOnly: true

    volumes:
      - name: npu-config
        configMap:
          name: npu-exporter-metric-config
    ```

    >[!NOTICE]
    >
    >- When uninstalling the component directly through the YAML in the NPU Exporter package, the `npu-exporter-metric-config` ConfigMap will be deleted. If you need to retain the configuration, back it up in advance:
    >
    >   ```bash
    >   kubectl get cm -n npu-exporter npu-exporter-metric-config -o yaml > npu-exporter-metric-config.yaml
    >   ```
    >
    >- The ConfigMap must be mounted directly to the directory, and **`subPath` must not be used**:
    >   - Using `subPath` prevents the ConfigMap from being automatically synchronized into the container after updates, requiring a restart to take effect.
    >   - After modifying the ConfigMap, the files in the container cannot be updated in real time. A certain period of time is required (K8s mechanism, up to about 10 minutes) before the file changes can be detected.

After the ConfigMap is updated, Kubernetes automatically updates the configuration file in the container, and NPU Exporter automatically detects and reloads the configuration.

### Configuration Change Verification

After the configuration is changed, you can check the NPU Exporter logs to confirm whether the configuration is loaded successfully:

```shell
# Log in to the corresponding server first.
# Prometheus scenario
tail -100f /var/log/mindx-dl/npu-exporter/npu-exporter.log

# Telegraf scenario
tail -100f /var/log/mindx-dl/npu-exporter/npu-plugin.log
```

A successful configuration load prints logs similar to the following:

```text
detected config change: ...
reloading configuration...
```
