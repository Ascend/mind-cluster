# NodeD<a name="ZH-CN_TOPIC_0000002479226406"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:15:19.099Z pushedAt=2026-06-09T06:22:06.979Z -->

- NodeD must be installed when full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, inference card fault recovery, rescheduling upon inference card faults, resumable training, or elastic training are required.
- Users who only use containerization support and resource monitoring can skip installing NodeD and proceed directly past this chapter.
- NodeD must be installed before using the slow node & slow network fault detection feature. For details, see [Slow Node & Slow Network Fault](../../../usage/resumable_training/01_solutions_principles.md#slow-nodes--slow-network-faults).

## Procedure<a name="section135381552125414"></a>

1. Log in to each compute node as the `root` user and run the following command to check whether the image and version number are correct.

    ```shell
    docker images | grep noded
    ```

    The following is an example of the command output:

    ```ColdFusion
    noded                               v26.0.0              ef801847acd2        29 minutes ago      133MB
    ```

    - If correct, proceed to [Step 2](#li26221447455).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

2. <a name="li26221447455"></a>Copy the YAML file in the directory where the NodeD software package is decompressed to any directory on the K8s management node.
3. If you do not modify the component startup parameters, you can skip this step. Otherwise, modify the NodeD startup parameters in the YAML file according to the actual situation. For startup parameters, see [Table 1](#table1862682843614). You can run `./noded -h` to view parameter descriptions.
4. (Optional) When using **resumable training** or **elastic training**, you need to configure the node status report interval. Add the `-reportInterval` parameter in the `args` line of the NodeD startup YAML file, as shown below:

    ```Yaml
    ...
              env:
                - name: NODE_NAME
                  valueFrom:
                    fieldRef:
                      fieldPath: spec.nodeName
              imagePullPolicy: Never
              command: [ "/bin/bash", "-c", "--"]
              args: [ "/usr/local/bin/noded -logFile=/var/log/mindx-dl/noded/noded.log -logLevel=0 -reportInterval=5" ]
              securityContext:
                readOnlyRootFilesystem: true
                allowPrivilegeEscalation: true
              volumeMounts:
                - name: log-noded
    ...
    ```

    >[!NOTE]
    >- K8s [marks a node as NotReady by default if no response is received within 40 seconds](https://kubernetes.io/docs/concepts/architecture/nodes/).
    >- When the request pressure on the K8s API Server increases, you can increase the interval time according to the actual situation to reduce the load on the API Server.

5. In the path where the YAML file is located on the management node, run the following command to start NodeD.
    - If you do not use the [DPC fault detection](../../../usage/resumable_training/01_solutions_principles.md#node-faults) function, run the following command.

        ```shell
        kubectl apply -f noded-v{version}.yaml
        ```

    - If the environment requires [DPC fault detection](../../../usage/resumable_training/01_solutions_principles.md#node-faults), run the following command to start NodeD.

        ```shell
        kubectl apply -f noded-dpc-v{version}.yaml
        ```

        Example:

        ```ColdFusion
        serviceaccount/noded created
        clusterrole.rbac.authorization.k8s.io/pods-noded-role created
        clusterrolebinding.rbac.authorization.k8s.io/pods-noded-rolebinding created
        daemonset.apps/noded created
        ```

6. Run the following command to check whether the component has started successfully.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is an example of the output. `Running` indicates that the component startup is successful.

    ```ColdFusion
    NAME                              READY   STATUS    RESTARTS   AGE
    ...
    noded-fd6t8                  1/1    Running  0        74s
    ...
    ```

>[!NOTE]
>
>- If the component Pod status is not `Running` after installation, refer to [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for troubleshooting.
>- If the component Pod status is `ContainerCreating` after installation, refer to [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343) for troubleshooting.
>- If the component fails to start, refer to [Cluster Scheduling Component Startup Failure with Log "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390) for information.
>- If the component starts successfully but the corresponding Pod cannot be found, refer to [Component Startup YAML Executed Successfully but Corresponding Pod Not Found](https://gitcode.com/Ascend/mind-cluster/issues/345) for information.

## Parameter Description<a name="section1851191618362"></a>

**Table 1** NodeD startup parameters

<a name="table1862682843614"></a>
<table><thead align="left"><tr id="row462602873614"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p14626028143611"><a name="p14626028143611"></a><a name="p14626028143611"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p136269286369"><a name="p136269286369"></a><a name="p136269286369"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p126271528193618"><a name="p126271528193618"></a><a name="p126271528193618"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="p13627192820361"><a name="p13627192820361"></a><a name="p13627192820361"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row162762819362"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p126271328193610"><a name="p126271328193610"></a><a name="p126271328193610"></a>-reportInterval</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p2062718289366"><a name="p2062718289366"></a><a name="p2062718289366"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1962732833610"><a name="p1962732833610"></a><a name="p1962732833610"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><a name="ul49338283283"></a><a name="ul49338283283"></a><ul id="ul49338283283"><li>Minimum interval for reporting node fault information. If the node status changes, it will be reported within 5s. If the node status remains unchanged, the reporting period is 30 minutes.</li><li>Value range: 1 to 300, unit: seconds.</li><li>When the request pressure on the K8s APIServer increases, you can increase the interval based on the actual situation to reduce the pressure on the APIServer.</li></ul>
</td>
</tr>
<tr id="row1240181274312"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1691522724316"><a name="p1691522724316"></a><a name="p1691522724316"></a>-monitorPeriod</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p17916227194316"><a name="p17916227194316"></a><a name="p17916227194316"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1491652715431"><a name="p1491652715431"></a><a name="p1491652715431"></a>60</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p139161227154317"><a name="p139161227154317"></a><a name="p139161227154317"></a>Polling detection period for node hardware faults. Value range: 60 to 600, unit: seconds.</p>
</td>
</tr>
<tr id="row562722803619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p862732803617"><a name="p862732803617"></a><a name="p862732803617"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p166271328153612"><a name="p166271328153612"></a><a name="p166271328153612"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p176271728143613"><a name="p176271728143613"></a><a name="p176271728143613"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p146279281367"><a name="p146279281367"></a><a name="p146279281367"></a>Whether to query the current version number of <span id="ph1437310218483"><a name="ph1437310218483"></a><a name="ph1437310218483"></a>NodeD</span>.</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true: Query.</li><li>false: Do not query.</li></ul>
</td>
</tr>
<tr id="row15627928153617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1627328103615"><a name="p1627328103615"></a><a name="p1627328103615"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p56272028193610"><a name="p56272028193610"></a><a name="p56272028193610"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p4627172833615"><a name="p4627172833615"></a><a name="p4627172833615"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p13627628113614"><a name="p13627628113614"></a><a name="p13627628113614"></a>Log level:</p>
<a name="ul262712284361"></a><a name="ul262712284361"></a><ul id="ul262712284361"><li>Value is -1: debug</li><li>Value is 0: info</li><li>Value is 1: warning</li><li>Value is 2: error</li><li>Value is 3: critical</li></ul>
</td>
</tr>
<tr id="row126271928143618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13627132863613"><a name="p13627132863613"></a><a name="p13627132863613"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p662782817368"><a name="p662782817368"></a><a name="p662782817368"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p062752813611"><a name="p062752813611"></a><a name="p062752813611"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p126271289369"><a name="p126271289369"></a><a name="p126271289369"></a>Log backup retention time. Value range: 7 to 700, unit: days.</p>
</td>
</tr>
<tr id="row0896102832513"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p178963287252"><a name="p178963287252"></a><a name="p178963287252"></a>-resultMaxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p18896202818250"><a name="p18896202818250"></a><a name="p18896202818250"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p48961228192511"><a name="p48961228192511"></a><a name="p48961228192511"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p198961128162516"><a name="p198961128162516"></a><a name="p198961128162516"></a>Number of days to retain pingmesh result backup files. Value range: [7, 700], unit: days.</p>
<div class="note" id="note1058610517274"><a name="note1058610517274"></a><a name="note1058610517274"></a><span class="notetitle">[!NOTE] Description</span><div class="notebody"><p id="p946415413280"><a name="p946415413280"></a><a name="p946415413280"></a>This parameter is supported only on <span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD</span>. The driver version must be ≥ 24.1.RC1.</p>
</div></div>
</td>
</tr>
<tr id="row86273287368"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1962772813618"><a name="p1962772813618"></a><a name="p1962772813618"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p162772823618"><a name="p162772823618"></a><a name="p162772823618"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p962817282367"><a name="p962817282367"></a><a name="p962817282367"></a>/var/log/mindx-dl/noded/noded.log</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1862816283365"><a name="p1862816283365"></a><a name="p1862816283365"></a>Log file. When a single log file exceeds 20 MB, automatic rotation is triggered. The maximum file size cannot be modified. The naming format of the rotated file is: noded-rotation_time.log, for example: noded-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row1862892813363"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10628202814365"><a name="p10628202814365"></a><a name="p10628202814365"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p4628828173616"><a name="p4628828173616"></a><a name="p4628828173616"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p16628182814362"><a name="p16628182814362"></a><a name="p16628182814362"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1062817287368"><a name="p1062817287368"></a><a name="p1062817287368"></a>Maximum number of rotated log files to retain. Value range: 1 to 30, unit: files.</p>
</td>
</tr>
<tr id="row68317556187"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101519"><a name="p0894319101519"></a><a name="p0894319101519"></a><span id="ph96781327191516"><a name="ph96781327191516"></a><a name="ph96781327191516"></a>-deviceResetTimeout</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151514"><a name="p108941719151514"></a><a name="p108941719151514"></a><span id="ph1899563312153"><a name="ph1899563312153"></a><a name="ph1899563312153"></a>int</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961512"><a name="p19894131961512"></a><a name="p19894131961512"></a><span id="ph67327379151"><a name="ph67327379151"></a><a name="ph67327379151"></a>60</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971510"><a name="p589551971510"></a><a name="p589551971510"></a><span id="ph4556742141516"><a name="ph4556742141516"></a><a name="ph4556742141516"></a>During component startup, if the number of chips is insufficient, the maximum time to wait for the driver to report the complete chips. Unit: seconds, value range: 10 to 600</span><span id="ph124041056151513"><a name="ph124041056151513"></a><a name="ph124041056151513"></a>.</span></p>
<a name="ul1354220213192"></a><a name="ul1354220213192"></a><ul id="ul1354220213192"><li><span id="ph278017516257"><a name="ph278017516257"></a><a name="ph278017516257"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 training series products</term></span>, <span id="ph13163257131918"><a name="ph13163257131918"></a><a name="ph13163257131918"></a>Atlas 800I A2 inference server</span>, <span id="ph10930753142211"><a name="ph10930753142211"></a><a name="ph10930753142211"></a>A200I A2 Box heterogeneous component</span>: Suggested configuration is 150 seconds.</li><li><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 training series products</term></span>, <span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>A200T A3 Box8 SuperPoD</span>, <span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 800I A3 SuperPoD</span>: Suggested configuration is 360 seconds.</li><li><span>Atlas 350 PCIe card</span>, <span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>Atlas 850 series hardware products</span>, <span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 950 SuperPoD</span>: Suggested configuration is 600 seconds.</li></ul>
</td>
</tr>
<tr id="row10282191492316"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p4283714172316"><a name="p4283714172316"></a><a name="p4283714172316"></a>-h or -help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p82838147233"><a name="p82838147233"></a><a name="p82838147233"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482316"><a name="p828341482316"></a><a name="p828341482316"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432318"><a name="p828311432318"></a><a name="p828311432318"></a>Display help information.</p>
</td>
</tr>
</tbody>
</table>
