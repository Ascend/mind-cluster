# DPU Exporter<a name="ZH-CN_TOPIC_0000002511426332"></a>

- When you use **DPU resource monitoring**, DPU Exporter must be installed. This component can connect to Prometheus and can be deployed in container mode or binary mode.
- If you do not use **DPU resource monitoring**, you do not need to install DPU Exporter. Skip this section.

## Usage Constraints<a name="section1362795652417"></a>

Before installing DPU Exporter, familiarize yourself with the related constraints. For details, see [Table 1](#table1050718522712).

**Table 1** Constraint description

<a name="table1050718522712"></a>
<table><thead align="left"><tr id="row2050719520273"><th class="cellrowborder" valign="top" width="29.970000000000002%" id="mcps1.2.3.1.1"><p id="p1950795152712"><a name="p1950795152712"></a><a name="p1950795152712"></a>Constraint Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="70.03%" id="mcps1.2.3.1.2"><p id="p75071151278"><a name="p75071151278"></a><a name="p75071151278"></a>Constraint Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row115077513272"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p17925222412"><a name="p17925222412"></a><a name="p17925222412"></a>DPU driver and tools</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p450745142713"><a name="p450745142713"></a><a name="p450745142713"></a><span id="ph10112356112714"><a name="ph10112356112714"></a><a name="ph10112356112714"></a>DPU Exporter</span> periodically calls the <code>hinicadm5</code> tool to query DPU's global metrics and reads the sysfs interface under <code>/sys/class/net/</code> to obtain interface-level metrics. To upgrade the driver or tools, stop service jobs first and then stop the <span id="ph154413248376"><a name="ph154413248376"></a><a name="ph154413248376"></a>DPU Exporter</span> service.</p>
<div class="note" id="note1993172317416"><a name="note1993172317416"></a><a name="note1993172317416"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><div class="p" id="zh-cn_topic_0000002479226379_p18934232419"><a name="zh-cn_topic_0000002479226379_p18934232419"></a><a name="zh-cn_topic_0000002479226379_p18934232419"></a>In container deployment, ensure that <code>hinicadm5</code> on the host is located in the <code>/usr/sbin/</code> directory. The file has been mounted to the container through hostPath in the YAML. Dynamic link libraries that <code>hinicadm5</code> depends on (such as <code>libmnl</code>) have been installed in the image during image build.</div>
</div></div>
</td>
</tr>
<tr id="row54685525283"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p5249201634115"><a name="p5249201634115"></a><a name="p5249201634115"></a><span id="ph1461172794117"><a name="ph1461172794117"></a><a name="ph1461172794117"></a>K8s</span> version</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p5468852142814"><a name="p5468852142814"></a><a name="p5468852142814"></a>When the container runtime is Docker, before using <span id="ph98079531287"><a name="ph98079531287"></a><a name="ph98079531287"></a>DPU Exporter</span>, check the <span id="ph18807253152811"><a name="ph18807253152811"></a><a name="ph18807253152811"></a>K8s</span> version of the environment. If the <span id="ph6808453102814"><a name="ph6808453102814"></a><a name="ph6808453102814"></a>K8s</span> version is 1.24.x or later, you need to <a href="https://github.com/mirantis/cri-dockerd#build-and-install" target="_blank" rel="noopener noreferrer">install cri-dockerd</a> as a dependency.</p>
</td>
</tr>
<tr id="row7507135142717"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p050719519272"><a name="p050719519272"></a><a name="p050719519272"></a>hinicadm5 concurrency</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p6555101612382"><a name="p6555101612382"></a><a name="p6555101612382"></a>The <code>hinicadm5</code> tool does not support concurrent calls. DPU Exporter collects global metrics card by card in serial mode, which requires no additional action from users.</p>
</td>
</tr>
<tr id="row_dpu_sysfs"><td class="cellrowborder" valign="top" width="29.970000000000002%" headers="mcps1.2.3.1.1 "><p id="p_dpu_sysfs_cat">sysfs file interface</p>
</td>
<td class="cellrowborder" valign="top" width="70.03%" headers="mcps1.2.3.1.2 "><p id="p_dpu_sysfs_desc">DPU Exporter obtains interface-level metrics (carrier, carrier_changes, operstate, and all statistics items in the statistics directory) by reading files in the <code>/sys/class/net/&lt;interface_name&gt;/</code> directory. Ensure that the user running DPU-Exporter has the read permission on the directory. In container deployment, the container runs with <code>hostNetwork: true</code>, and <code>/sys/class/net/</code> in the container is the NIC information of the host.</p>
</td>
</tr>
</tbody>
</table>

## Procedure<a name="section83111543151613"></a>

DPU Exporter supports two installation modes. You can select either of them based on the actual situation. This component provides only HTTP services. If you need to use more secure HTTPS services, modify the source code for adaptation.

- (Recommended) Run in container mode: For the installation procedure, see [Running in Container Mode](#section2035402135915).
- When security requirements are high, you are advised to run in binary mode on a physical machine. For the installation procedure, see [Running in Binary Mode](#section103551921135918).

## Running in Container Mode<a name="section2035402135915"></a>

1. Log in to each compute node as the `root` user.

2. (Optional) Configure the collection period and metric trustlist.

    The following is a configuration file (`config.json`) example:

    ```json
    {
        "hinicadm5CollectorInterval": 40,
        "sysfsCollectorInterval": 20,
        "dpuListRefreshInterval": 60,
        "metricWhiteList": []
    }
    ```

    <a name="table192202574407"></a>

    |Parameter|Description|
    |---|---|
    |hinicadm5CollectorInterval|Collection period of global metrics (hinicadm5 counter), in seconds. The value range is 1 to 86400, and the default value is 40.|
    |sysfsCollectorInterval|Collection period of interface-level metrics (sysfs file interface), in seconds. The value range is 1 to 86400, and the default value is 20.|
    |dpuListRefreshInterval|Refresh period of the DPU device list, in seconds. The value range is 1 to 86400, and the default value is 60.|
    |metricWhiteList|Custom metric whitelist, which supports prefix matching (ending with `*`). When it is empty, the [Default Trustlist](../../../06_api/16_dpu_exporter.md#default-trustlist) is used.|

    The configuration file can be mounted in either of the following modes. The method for modifying the configuration depends on the selected mount mode. For details, see [Dynamically Loading Configurations](#dynamically-loading-configurations).

    - **Default mode (ConfigMap mount)**: The configuration file is mounted to the container through a ConfigMap, and the default configuration defined in the YAML is used. To modify the configuration, modify the `config.json` content defined by the ConfigMap in the YAML directly, or update the ConfigMap after deployment by running **kubectl edit cm -n dpu-exporter dpu-exporter-config**. The modification takes effect dynamically.
    - **HostPath mount**: Mount the configuration file on the host to the container. The path in the container is `/etc/dpu-exporter/config.json`. In this mode, modify the `config.json` file on the host. Each node can be configured independently, or a shared directory can be used for global unified configuration. If you mount the file to another path in the container, modify the `-config` startup parameter of DPU Exporter in the YAML (for details, see [Parameter Description](#parameter-description)).

3. Check whether the DPU Exporter image exists and whether its name and version number are correct based on the container runtime in use.

    - **Docker scenario**: Run the following command.

        ```shell
        docker images | grep dpu-exporter
        ```

        Command output:

        ```ColdFusion
        dpu-exporter                         v26.2.0              20185c45f1bc        About an hour ago         90.1MB
        ```

    - **Containerd scenario**: Run the following command.

        ```shell
        ctr -n k8s.io i ls | grep dpu-exporter
        ```

    If the image exists and its name and version number are correct, perform [step 4](#li0640635114212). If the image does not exist, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to build and distribute the image.

4. <a name="li0640635114212"></a>Copy the YAML file in the DPU Exporter software package decompression directory to any directory on the K8s management node.

5. If you do not need to modify other startup parameters of the component, skip this step. Otherwise, modify the DPU Exporter startup parameters in the YAML file as required. For details about the startup parameters, see [Table 2](#table872410431915). You can also run <b>./dpu-exporter -h</b> to view the parameter description.

6. In the directory where the YAML file is stored on the management node, run the following command to start DPU Exporter.

    ```shell
    kubectl apply -f dpu-exporter-v26.2.0.yaml
    ```

    Command output:

    ```ColdFusion
    namespace/dpu-exporter created
    configmap/dpu-exporter-config created
    daemonset.apps/dpu-exporter created
    service/dpu-exporter created
    ```

    >[!NOTE]
    >When starting DPU Exporter, if the error "Error from server (NotFound): error when creating "dpu-exporter-v26.2.0.yaml":namespaces "dpu-exporter" not found" occurs, the namespace of DPU Exporter fails to be created. Run the following command to create it manually.
    >
    >```shell
    >kubectl create ns dpu-exporter
    >```

7. Run the following command on any node to check whether the component has started successfully.

    ```shell
    kubectl get pod -n dpu-exporter
    ```

    The following is the output example. The presence of **Running** indicates that the component has started successfully.

    ```ColdFusion
    NAME                      READY   STATUS    RESTARTS   AGE
    dpu-exporter-xxxxx        1/1     Running   0          11s
    ```

    >[!NOTE]
    >
    >- When DPU Exporter runs in container mode, the YAML mounts "/usr/sbin/hinicadm5" (the hinicadm5 tool) and the "/var/log/hinic5" directory (hinicadm5 logs) on the host to the container through hostPath by default. Dynamic link libraries that hinicadm5 depends on (such as libmnl) have been installed in the image during image build.
    >- If the Pod status of the component is not Running after the component is installed, view the Pod logs to locate the problem:
    >
    >   ```bash
    >   kubectl logs -n dpu-exporter <pod-name>
    >   ```

8. Verify that metrics are collected properly.

    Run the following command on any node to access the metrics interface. `<node-ip>` is the IP address of the node where DPU Exporter is deployed.

    ```shell
    curl http://<node-ip>:8083/metrics
    ```

    Normally, metrics prefixed with `dpu_` and `dpu_interface_` can be seen.

## Running in Binary Mode<a name="section103551921135918"></a>

When DPU Exporter runs in container mode, privileged containers, the `root` user, and hostPath directory mounts are required. If a container is exploited maliciously, container escape risks exist. When security requirements are high, you can run the component in binary mode directly on a physical machine.

>[!NOTE]
>
>- When DPU Exporter is deployed in binary mode, the component must run as the `root` user: the code checks the consistency between the owner of the `hinicadm5` file and the current process user at startup, and `hinicadm5`, as the management tool of the DPU driver, itself requires root permissions for execution.
>- The commands in the following steps are all run as the `root` user.

1. Log in to the server as the `root` user.

2. Upload the DPU Exporter software package to any directory on the server (for example, "/home/ascend-dpu-exporter") and decompress it.

3. (Optional) Create a configuration file directory and copy the configuration file to the directory.

    ```shell
    cd /home/ascend-dpu-exporter
    mkdir -p /etc/dpu-exporter
    cp config.json /etc/dpu-exporter/
    ```

4. (Optional) Modify the `config.json` file to configure the collection period and metric trustlist.
    1. Go to the "/etc/dpu-exporter" directory.
    2. <a name="li1445835411479"></a>Open the `config.json` file.

        ```shell
        vi config.json
        ```

    3. Press `i` to enter the edit mode and configure the collection period and metric trustlist as required.

        The following is a configuration file example:

        ```json
        {
            "hinicadm5CollectorInterval": 40,
            "sysfsCollectorInterval": 20,
            "dpuListRefreshInterval": 60,
            "metricWhiteList": []
        }
        ```

        <a name="zh-cn_topic_0000002511426332_table192202574407"></a>

        |Parameter|Description|
        |---|---|
        |hinicadm5CollectorInterval|Collection period of global metrics (hinicadm5 counter), in seconds. The value range is 1 to 86400. The default value is 40.|
        |sysfsCollectorInterval|Collection period of interface-level metrics (sysfs file interface), in seconds. The value range is 1 to 86400. The default value is 20.|
        |dpuListRefreshInterval|Refresh period of the DPU device list, in seconds. The value range is 1 to 86400. The default value is 60.|
        |metricWhiteList|Custom metric trustlist, which supports prefix matching (ending with `*`). When it is empty, the [Default Trustlist](../../../06_api/16_dpu_exporter.md#default-trustlist) is used.|

    4. <a name="li18459954104719"></a>Press `Esc`, type `:wq!`, and press `Enter` to save and exit.

5. Create and edit the `dpu-exporter.service` file.
    1. Create the `dpu-exporter.service` file.

        ```shell
        vi /home/ascend-dpu-exporter/dpu-exporter.service
        ```

    2. Write the following content to the `dpu-exporter.service` file.

        <pre>
        [Unit]
        Description=Ascend dpu exporter
        Documentation=hiascend.com

        [Service]
        ExecStart=/bin/bash -c "/usr/local/bin/dpu-exporter -config=/etc/dpu-exporter/config.json -port=8083 >/dev/null 2>&1 &"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking

        [Install]
        WantedBy=multi-user.target</pre>

        By default, DPU Exporter listens on port 8083. You can change the listening port by modifying the "-port" startup parameter in the "ExecStart" field of the "dpu-exporter.service" file.

    3. <a name="li18459954104719"></a>Press `Esc`, type `:wq!`, and press `Enter` to save and exit.

6. Create and edit the `dpu-exporter.timer` file. Configuring a delayed start through the timer ensures that the DPUs are in position when DPU Exporter starts.
    1. Create the `dpu-exporter.timer` file.

        ```shell
         vi /home/ascend-dpu-exporter/dpu-exporter.timer
        ```

    2. Refer to the following example and write the content to the `dpu-exporter.timer` file.

        <pre>
        [Unit]
        Description=Timer for DPU Exporter Service

        [Timer]
        # Set the delayed start time of DPU Exporter. Adjust it based on the actual situation.
        OnBootSec=60s
        Unit=dpu-exporter.service

        [Install]
        WantedBy=timers.target</pre>

    3. <a name="li18459954104719"></a>Press `Esc`, type `:wq!`, and press `Enter` to save and exit.

7. Run the following commands in sequence to enable the DPU Exporter service.

    ```shell
    cd /home/ascend-dpu-exporter
    cp dpu-exporter /usr/local/bin
    mkdir -p /var/log/mindx-dl/dpu-exporter
    cp dpu-exporter.service /etc/systemd/system
    chattr +i /etc/systemd/system/dpu-exporter.service
    cp dpu-exporter.timer /etc/systemd/system
    chattr +i /etc/systemd/system/dpu-exporter.timer
    chmod 500 /usr/local/bin/dpu-exporter
    chattr +i /usr/local/bin/dpu-exporter
    systemctl enable dpu-exporter.timer
    systemctl start dpu-exporter
    systemctl start dpu-exporter.timer
    ```

    >[!NOTE]
    >During the first deployment, `systemctl start dpu-exporter` starts the service immediately, and the `OnBootSec` delay of the timer takes effect only when the system restarts. If you want the first start to be delayed by the timer as well, run only `systemctl start dpu-exporter.timer`.

8. Verify that the service has started successfully.

    ```shell
    curl http://127.0.0.1:8083/metrics
    ```

    Normally, metrics prefixed with `dpu_` and `dpu_interface_` can be seen.

## Parameter Description<a name="section2042611570393"></a>

**Table 2** DPU Exporter startup parameters

<a name="table872410431915"></a>

|Parameter|Type|Default Value|Description|
|--|--|--|--|
|-config|string|/etc/dpu-exporter/config.json|Configuration file path. The configuration file is in JSON format and contains configuration items such as the collection period and metric trustlist. For details, see [Configuration File Description](#table192202574407).|
|-port|int|8083|Listening port, with a value range of 1025 to 40000.|
|-version|bool|false| Whether to query details such as the DPU Exporter version number and commit ID.<ul><li>true: query.</li><li>false: do not query.</li></ul>|
|--enable-healthz|bool|false|Whether to enable the health check service. In K8s deployment, it is enabled (true) by the component YAML configuration.<ul><li>true: enabled.</li><li>false: disabled.</li></ul>|
|--healthz-address|string|11251|Listening port number of the health check service, with a value range of 1025 to 65535. In K8s deployment, it is set to 11259 by the component YAML configuration. If the specified port is occupied, the component fails to start. For details about the health probe interfaces, see [Health Probe](../../../06_api/16_dpu_exporter.md#section_dpu_health_probe).|
|--tls-cert-file|string|""|Path of the HTTPS certificate file. If it is empty, the HTTP protocol is used. This parameter and --tls-private-key-file must be configured or left empty at the same time.|
|--tls-private-key-file|string|""|Path of the HTTPS private key file. If it is empty, the HTTP protocol is used. This parameter and --tls-cert-file must be configured or left empty at the same time.|
|-logLevel|int|0|Log level: <ul><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>|
|-maxAge|int|7|Log retention time, with a value range of 7 to 700, in days.|
|-maxBackups|int|30|Upper limit of retained log files after rotation, with a value range of 1 to 180, in files.|
|-logFile|string|/var/log/mindx-dl/dpu-exporter/dpu-exporter.log|Log file path. This parameter currently cannot be modified through the command line, and the log path is fixed to the default value.<p>When a single log file exceeds 20 MB, automatic rotation is triggered, and the upper limit of the file size cannot be modified. The naming format of a rotated file is dpu-exporter-<time_when_rotation_is_triggered>.log, for example, dpu-exporter-2023-10-07T03-38-24.402.log.</p>|
|-h or -help|N/A|N/A|Help information.|

## Dynamically Loading Configurations<a name="dynamically-loading-configurations"></a>

DPU Exporter supports dynamically loading the configuration file, so that configuration changes take effect without restarting the component.

### Binary Deployment

The configuration file path is `/etc/dpu-exporter/config.json`.

You only need to modify this file directly. DPU Exporter automatically detects file changes and reloads the configuration.

### K8s HostPath Mount

- Advantages of `HostPath` mount:
  - Configuration changes take effect immediately.
  - Each node can be configured independently, or a shared directory can be used for global unified configuration.
- Disadvantage of `HostPath` mount: Configuration changes are not easy to track.

Configure `HostPath` mount in the deployment YAML to mount the configuration file on the host to the `/etc/dpu-exporter` path of the container:

1. Prepare the configuration file on each node (run the commands in the decompressed software package directory, for example, "/home/ascend-dpu-exporter").

    ```bash
    cd /home/ascend-dpu-exporter
    mkdir -p /etc/dpu-exporter
    cp config.json /etc/dpu-exporter/
    ```

2. Change the volume mount in the YAML to the `HostPath` mode.

    ```yaml
    volumeMounts:
      - name: config
        mountPath: /etc/dpu-exporter
        readOnly: true

    volumes:
      - name: config
        hostPath:
          path: /etc/dpu-exporter
          type: DirectoryOrCreate
    ```

You only need to modify this file directly. DPU Exporter automatically detects file changes and reloads the configuration.

### K8s ConfigMap Mount (Default)

- Advantages of ConfigMap mount:
  - Unified management of the configurations of all nodes, supporting one-click update of the configurations of all nodes.
  - Configuration changes are traceable and version-controlled.
- Disadvantages of ConfigMap mount:
  - All nodes use the same configuration, and a single node cannot be configured independently.
  - There is a certain delay before the configuration takes effect (the time for K8s to update the ConfigMap to the container).

By default, the YAML of DPU Exporter mounts the configuration file in ConfigMap mode. The ConfigMap has been defined in the YAML and contains the `config.json` key.

Update the configuration in the ConfigMap.

```bash
kubectl edit cm -n dpu-exporter dpu-exporter-config
```

Alternatively, modify the configuration values in the YAML and redeploy directly:

```bash
kubectl apply -f dpu-exporter-v26.2.0.yaml
```

>[!NOTICE]
>
>The ConfigMap must be mounted directly to the directory. **`subPath` cannot be used.**
>
> - Using `subPath` prevents the updated ConfigMap from being automatically synchronized to the container, and a restart is required for the change to take effect.
> - After the ConfigMap is modified, the file in the container is not updated in real time. It takes a certain period (a K8s mechanism, up to about 10 minutes) for the file change to be detected.

After the ConfigMap is updated, K8s automatically updates the configuration file in the container, and DPU Exporter automatically detects and reloads the configuration.

### Verifying Configuration Changes

After a configuration change, you can check whether the configuration is loaded successfully by viewing the DPU Exporter logs:

```shell
# Binary deployment
tail -100f /var/log/mindx-dl/dpu-exporter/dpu-exporter.log

# Container deployment
kubectl logs -n dpu-exporter <pod-name> --tail=100
```

If the configuration is loaded successfully, logs similar to the following are printed:

```text
reloading config...
config loaded from /etc/dpu-exporter/config.json
collector chains rebuilt after config reload
config reload completed
```

## Trustlist Configuration Description<a name="whitelist-configuration-description"></a>

DPU Exporter controls the collection scope of hinicadm5 global metrics through the trustlist mechanism to avoid collecting unnecessary metrics.

### Trustlist Priority

- **Custom trustlist takes precedence**: If `metricWhiteList` is configured (and not empty), the custom trustlist is used.
- **Default trustlist**: If `metricWhiteList` is empty or not configured, the built-in default trustlist is used.

### Trustlist Matching Rules

- **Exact match**: The metric matches when the metric name is exactly the same as the pattern.
- **Prefix match**: When the pattern ends with `*`, all metric names starting with the prefix are matched.

### Configuration Example

Configure a custom trustlist in `config.json`:

```json
{
    "hinicadm5CollectorInterval": 40,
    "sysfsCollectorInterval": 20,
    "dpuListRefreshInterval": 60,
    "metricWhiteList": [
        "roce_rx_*",
        "roce_tx_*",
        "roce_cmdq_ctr_roce_cmd_2rst_qp"
    ]
}
```

The preceding configuration means that only metrics starting with `roce_rx_`, metrics starting with `roce_tx_`, and metrics exactly matching `roce_cmdq_ctr_roce_cmd_2rst_qp` are collected.
