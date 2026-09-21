# Configuring Offline Reset for Inference Jobs<a name="ZH-CN_TOPIC_0000002479226442"></a>

Currently, offline reset is only supported for Atlas 800I A2 inference servers and Atlas 800I A3 SuperPoD servers. When this function is enabled, a hot reset operation is performed after a chip fault occurs, restoring the chip to a healthy state.

## Prerequisites

[Ascend Device Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/04_ascend_device_plugin.md) has been installed.

## Procedure

To enable the offline reset feature for MindIE Motor inference jobs, simply set the Ascend Device Plugin startup parameter `-hotReset` to `0` or `2`.

If Ascend Device Plugin has not been started, you can directly modify the YAML startup parameters of Ascend Device Plugin to enable the offline reset feature. If Ascend Device Plugin has already been started, you can use the following method to enable the offline reset feature.

1. View Ascend Device Plugin DaemonSet information.

    ```shell
    kubectl get daemonsets -A
    ```

    Command output:

    ```CodeFusion
    NAMESPACE      NAME                             DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR                    AGE
    ...
    kube-system    ascend-device-plugin-daemonset   2         2         2       2            2           workerselector=dls-worker-node   7h21m
    ...
    ```

2. Edit the DaemonSet of Ascend Device Plugin.

    ```shell
    kubectl edit daemonset ascend-device-plugin-daemonset -n kube-system
    ```

    Add the parameter `-hotReset=2` to the container arguments and save.

    <pre codetype="yaml">
    ...
    containers:
      - args:
        - device-plugin -volcanoType=true -presetVirtualDevice=true -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log
          -logLevel=0 --enable-healthz=true --healthz-address=11251 -hotReset=2
        command:
    ...
    ...</pre>

**Table 1**  Parameter description

<a name="table173461839165111"></a>

|Parameter|Type|Default Value|Description|
|--|--|--|--|
|-hotReset|int|-1|Device hot reset. When this function is enabled, if a chip fault occurs, Ascend Device Plugin performs a hot reset operation to restore the chip to a healthy state.<ul><li>-1: Disable the chip reset function</li><li>0: Enable the inference device reset function</li><li>1: Enable the training device online reset function</li><li>2: Enable the training/inference device offline reset function</li></ul><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>The function corresponding to the value 1 has been deprecated. Configure other values.</p></div></div>Supported devices for this parameter: <ul><li>Atlas 800 training server (model 9000) (fully populated with NPUs)</li><li>Atlas 800 training server (model 9010) (fully populated with NPUs)</li><li>Atlas 900T PoD Lite</li><li>Atlas 900 PoD (model 9000)</li><li>Atlas 800T A2 training server</li><li>Atlas 900 A2 PoD cluster basic unit</li><li>Atlas 900 A3 SuperPod</li><li>Atlas 800T A3 SuperPoD server</li><li>Atlas 850E SuperPoD</li><li>Atlas 650E server</li><li>Atlas 950 SuperPoD</li><li>Atlas 350 accelerator card</li><li>Atlas 300I Pro inference card</li><li>Atlas 300V video analysis card</li><li>Atlas 300V Pro video analysis card</li><li>Atlas 300I Duo inference card</li><li>Atlas 300I inference card (model 3000) (entire card)</li><li>Atlas 300I inference card (model 3010)</li><li>Atlas 800I A2 inference server</li><li>A200I A2 Box heterogeneous subrack</li><li>Atlas 800I A3 SuperPoD server</li></ul>|

>[!NOTE]
>Atlas 800I A2 inference server supports the following two fault recovery methods. An Atlas 800I A2 inference server can use only one fault recovery method, which is automatically identified by cluster scheduling components.
>
>- Method 1: If no HCCS ring exists on the device, during inference job execution, when an NPU fault occurs, Ascend Device Plugin waits for the NPU to become idle and then resets the NPU.
>- Method 2: If an HCCS ring exists on the device, during inference job execution, when one or more faulty NPUs occur on the server, Ascend Device Plugin waits for all NPUs on the ring to become idle and then resets all NPUs on the ring at once.

## Viewing Offline Reset Results

Run the following command to query basic device information.

```shell
/usr/local/bin/npu-smi info
```

When an NPU device has a fault, the output is as follows:

```ColdFusion
+------------------------------------------------------------------------------------------------+
| npu-smi 23.0.5                   Version: 23.0.5                                               |
+---------------------------+---------------+----------------------------------------------------+
| NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
| Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
+===========================+===============+====================================================+
| 0     xxx                 | Warning            | 73.1        37                0    / 0             |
| 0                         | 0000:61:00.0  | 0           920  / 13553      0    / 32768         |
+===========================+===============+====================================================+
...
+===========================+===============+====================================================+
| 7     xxx                 | OK            | 67.0        38                0    / 0             |
| 0                         | 0000:3D:00.0  | 0           2346 / 15567      0    / 32768         |
+===========================+===============+====================================================+
+---------------------------+---------------+----------------------------------------------------+
| NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
+===========================+===============+====================================================+
| No running processes found in NPU 0                                                            |
+===========================+===============+====================================================+
...
+===========================+===============+====================================================+
| No running processes found in NPU 7                                                            |
+===========================+===============+====================================================+
```

After the offline reset feature is successfully executed, the NPU device status will change from "Warning" to "OK".

```ColdFusion
+------------------------------------------------------------------------------------------------+
| npu-smi 23.0.5                   Version: 23.0.5                                               |
+---------------------------+---------------+----------------------------------------------------+
| NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
| Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
+===========================+===============+====================================================+
| 0     xxx                 | OK            | 73.1        37                0    / 0             |
| 0                         | 0000:61:00.0  | 0           920  / 13553      0    / 32768         |
+===========================+===============+====================================================+
...
+===========================+===============+====================================================+
| 7     xxx                 | OK            | 67.0        38                0    / 0             |
| 0                         | 0000:3D:00.0  | 0           2346 / 15567      0    / 32768         |
+===========================+===============+====================================================+
+---------------------------+---------------+----------------------------------------------------+
| NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
+===========================+===============+====================================================+
| No running processes found in NPU 0                                                            |
+===========================+===============+====================================================+
...
+===========================+===============+====================================================+
| No running processes found in NPU 7                                                            |
+===========================+===============+====================================================+
```
