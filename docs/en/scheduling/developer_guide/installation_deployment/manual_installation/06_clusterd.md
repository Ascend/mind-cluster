# ClusterD<a name="ZH-CN_TOPIC_0000002511346341"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:14:06.865Z pushedAt=2026-06-09T06:22:06.841Z -->

- Users who use full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, resumable training, elastic training, inference card fault recovery, or rescheduling upon inference card faults must install ClusterD. ClusterD can provide full information collection services only when both Ascend Device Plugin and NodeD components exist in the cluster.
- When you install ClusterD, it is recommended to install Volcano in advance. If ClusterD is installed before Volcano, the Pod where ClusterD resides may experience `CrashLoopBackOff`. ClusterD will return to normal only after the Volcano Pod starts.
- Users who only use containerization support and resource monitoring do not need to install ClusterD. Please skip this chapter directly.
- Before starting ClusterD, you need to start Ascend Operator first. ClusterD needs to perceive Ascend Job resources.
- To detect slow node and slow network faults, you need to install ClusterD. For details, see [Slow Node and Slow Network Faults](../../../usage/resumable_training/00_feature_description.md).

## Procedure<a name="section20114193212615"></a>

1. Log in to the K8s management node as the `root` user and run the following command to check whether the ClusterD image and version number are correct.

    ```shell
    docker images | grep clusterd
    ```

    The following is an example of the output:

    ```ColdFusion
    clusterd                   v26.0.0              c532e9d0889c        About an hour ago         126MB
    ```

    - If correct, perform [Step 2](#li615118054419).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image building and distribution.

2. <a name="li615118054419"></a>Copy the YAML files from the extracted ClusterD software package directory to any directory on the K8s management node.
3. If you do not need to modify the component startup parameters, skip this step. Otherwise, modify the ClusterD startup parameters in the YAML file based on the actual situation. For startup parameters, see [Table 2](#table11614104894617). You can run `./clusterd -h` in the ClusterD binary package directory to view the parameter description.
4. (Optional) In `clusterd-v{version}.yaml`, configure the detection switch for manually isolated chips, fault frequency, de-isolation time, etc.

    ```Yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: clusterd-config-cm
      namespace: cluster-system
    data:
      manually_separate_policy.conf: |
        enabled: true
        separate:
          fault_window_hours: 24
          fault_threshold: 3
        release:
          fault_free_hours: 48

    ```

    **Table 1** Parameter description of manually_separate_policy.conf

    <a name="table208901"></a>

    |Primary Parameter|Secondary Parameter|Type|Description|
    |--|--|--|--|
    |enabled|-|bool|Detection switch for manually isolated chips. Values include: <ul><li>true: enabled.</li><li>false: disabled.</li></ul><p>The default value is true. If this switch is turned off, all chips manually isolated by ClusterD and related caches will be cleared.</p><div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>The YAML specification supports multiple boolean value formats (including case variants), but different parsers (such as K8s, Go, and Python) have varying compatibility, and not all formats are supported. It is recommended to use lowercase true/false uniformly.</p></div></div>|
    |separate|fault_window_hours|int|Duration for manual chip isolation. Within this duration, if the number of faults with the same fault code reaches the fault_threshold value, ClusterD will manually isolate the faulty chip. The value range is [1, 720], the default value is 24, and the unit is h (hour).|
    |-|fault_threshold|int|Threshold for manual chip isolation. The value range is [1, 50], the default value is 3, and the unit is times.|
    |release|fault_free_hours|int|Unisolation time, indicating the time elapsed since the last isolation triggered by reaching the frequency threshold. After this time, the isolation will be removed. The value range is [1, 240] or -1, the default value is 48, and the unit is h (hour).<ul><li>The time when the frequency threshold was last reached is the LastSeparateTime in clusterd-manual-info-cm. For details about clusterd-manual-info-cm, see [clusterd-manual-info-cm](../../../api/clusterd/00_cluster_resources.md#clusterd-manual-info-cm).</li><li>Configuring -1 disables the unisolation function.</li><li>When automatic unisolation is performed upon reaching the unisolation time, the isolation will be removed regardless of whether the fault has recovered.</li></ul>|

    >[!NOTE]
    >If the `enabled` field is missing, ClusterD will recognize it as `false`; if other int type fields are missing, ClusterD will recognize them as `0`.

5. On the management node, in the path where the YAML file is located, run the following command to start ClusterD.

    ```shell
    kubectl apply -f clusterd-v{version}.yaml
    ```

    Example:

    ```ColdFusion
    clusterrolebinding.rbac.authorization.k8s.io/pods-clusterd-rolebinding created
    lease.coordination.k8s.io/cluster-info-collector created
    deployment.apps/clusterd created
    service/clusterd-grpc-svc created
    ```

6. Run the following command to check whether the component startup is successful.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is an example of the output. `Running` indicates that the component startup is successful.

    ```ColdFusion
    NAME                          READY   STATUS              RESTARTS   AGE
    clusterd-7844cb867d-fwcj7     0/1     Running            0          45s
    ```

>[!NOTE]
>
>- If the Pod status of the component is not `Running` after installation, refer to [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for handling.
>- If the Pod status of the component is `ContainerCreating` after installation, refer to [Cluster Scheduling Component Pod in ContainerCreating Status](https://gitcode.com/Ascend/mind-cluster/issues/343) for handling.
>- If the component fails to start, refer to [Failed to Start Cluster Scheduling Component, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390) for information.
>- If the component starts successfully but the corresponding Pod cannot be found, refer to [Component Startup YAML Executed Successfully but Corresponding Pod Not Found](https://gitcode.com/Ascend/mind-cluster/issues/345) for information.

## Parameter Description<a name="section1250239182212"></a>

**Table 2** ClusterD startup parameters

<a name="table11614104894617"></a>
<table><thead align="left"><tr id="row2614114884616"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p961416489463"><a name="p961416489463"></a><a name="p961416489463"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="14.979999999999999%" id="mcps1.2.5.1.2"><p id="p6614174812464"><a name="p6614174812464"></a><a name="p6614174812464"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.5.1.3"><p id="p12614194844618"><a name="p12614194844618"></a><a name="p12614194844618"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p261454810466"><a name="p261454810466"></a><a name="p261454810466"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row14614134874619"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p86145488460"><a name="p86145488460"></a><a name="p86145488460"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p20614848194617"><a name="p20614848194617"></a><a name="p20614848194617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p26141489467"><a name="p26141489467"></a><a name="p26141489467"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>Queries the <span id="ph1950137183918"><a name="ph1950137183918"></a><a name="ph1950137183918"></a>ClusterD</span> version.</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true: Query.</li><li>false: Do not query.</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6614174884615"><a name="p6614174884615"></a><a name="p6614174884615"></a>Log level:</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>Log backup retention period. The value range is 7 to 700, in days.</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1668892293119"><a name="p1668892293119"></a><a name="p1668892293119"></a>/var/log/mindx-dl/clusterd/clusterd.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>Log file. Automatic rotation is triggered when a single log file exceeds 20 MB. The maximum file size cannot be modified. The naming format for rotated files is: clusterd-<i>rotation_time</i>.log, for example, clusterd-2024-06-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>Maximum number of rotated log files to retain. The value range is 1 to 30.</p>
</td>
</tr>
<tr id="row147481810102010"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p15748191011204"><a name="p15748191011204"></a><a name="p15748191011204"></a>-useProxy</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p17830536152010"><a name="p17830536152010"></a><a name="p17830536152010"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13748141013205"><a name="p13748141013205"></a><a name="p13748141013205"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2748131042020"><a name="p2748131042020"></a><a name="p2748131042020"></a>Whether to use a proxy to forward gRPC requests.</p>
<a name="ul71770166215"></a><a name="ul71770166215"></a><ul id="ul71770166215"><li>true: Yes</li><li>false: No
</li></ul><div class="note" id="note12300045132119"><a name="note12300045132119"></a><a name="note12300045132119"></a><span class="notetitle">[!NOTE] Note</span><div class="notebody"><p id="p17300245162118"><a name="p17300245162118"></a><a name="p17300245162118"></a>It is recommended to set this parameter to "true" in the startup YAML and perform security hardening for ClusterD. For details, see the <a href="../../../references/security_hardening.md#clusterd-security-hardening">ClusterD Security Hardening</a> section.</p>
</div></div>
</td>
</tr>
<tr id="row2615144813463"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1061594884617"><a name="p1061594884617"></a><a name="p1061594884617"></a>-h or -help</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16151748144614"><a name="p16151748144614"></a><a name="p16151748144614"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13615048184615"><a name="p13615048184615"></a><a name="p13615048184615"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16616174834615"><a name="p16616174834615"></a><a name="p16616174834615"></a>Displays help information.</p>
</td>
</tr>
</tbody>
</table>
