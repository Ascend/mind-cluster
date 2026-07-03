# Infer Operator<a name="ZH-CN_TOPIC_0000002479386414"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:13:59.832Z pushedAt=2026-06-09T06:22:06.829Z -->

To deliver inference jobs of the InferServiceSet type, you must install Infer Operator.

## Procedure<a name="section209273712583"></a>

1. Log in to the K8s management node as the `root` user and run the following command to check whether the Infer Operator image and version number are correct.

    ```shell
    docker images | grep infer-operator
    ```

    The following is an example of the command output:

    ```ColdFusion
    infer-operator                      v26.0.0              c7221984e8ae        About an hour ago         140MB
    ```

    - If correct, proceed to [Step 2](#li19793191914420).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation and distribution.

2. <a name="li19793191914420"></a>Copy the YAML file from the extracted Infer Operator package directory to any directory on the Kubernetes management node.
3. If you do not need to modify the component startup parameters, you can skip this step. Otherwise, modify the startup parameters of Infer Operator in the YAML file according to the actual situation. For details on startup parameters, see [Table 1](#table11614104894618). You can run `./infer-operator -h` to view parameter descriptions.
4. In the directory where the YAML file is located on the management node, run the following command to start Infer Operator.

    ```shell
    kubectl apply -f infer-operator-v{version}.yaml
    ```

    Example:

    ```ColdFusion
    configmap/infer-operator-config created
    deployment.apps/infer-operator-manager created
    serviceaccount/infer-operator-manager created
    clusterrole.rbac.authorization.k8s.io/infer-operator-manager-role created
    clusterrolebinding.rbac.authorization.k8s.io/infer-operator-manager-rolebinding created
    customresourcedefinition.apiextensions.k8s.io/inferservicesets.mindcluster.huawei.com created
    customresourcedefinition.apiextensions.k8s.io/inferservices.mindcluster.huawei.com created
    customresourcedefinition.apiextensions.k8s.io/instancesets.mindcluster.huawei.com created
    ...
    ```

5. Run the following command to check whether the component startup is successful.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is an example of the command output. `Running` indicates that the component startup is successful.

    ```ColdFusion
    NAME                                         READY   STATUS    RESTARTS   AGE
    ...
    infer-operator-8322455ba7b-hwmjw      1/1    Running  0         11s
    ```

> **NOTE**
>
>- If the Pod status of the component is not `Running` after installation, see [Component Pod Status Is Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
>- If the Pod status of the component is `ContainerCreating` after installation, see [Cluster Scheduling Component Pod Is in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343).
>- If the component fails to start, see [Failed to Start Cluster Scheduling Component, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
>- If the component starts successfully but the corresponding Pod cannot be found, see [Component Startup YAML Executed Successfully but Corresponding Pod Cannot Be Found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Parameter Description<a name="section91521925121114"></a>

**Table 1** Infer Operator startup parameters

<a name="table11614104894618"></a>
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
<tbody><tr><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p>Queries the <span>ClusterD</span> version number.</p><ul><li>true: Query.</li><li>false: Do not query.</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2312104517312"><a name="p2312104517312"></a><a name="p2312104517312"></a>The log level supports the following values:</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>Value -1: debug</li><li>Value 0: info</li><li>Value 1: warning</li><li>Value 2: error</li><li>Value 3: critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>Log backup time limit. The value range is 7 to 700, in days.</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p56159486469"><a name="p56159486469"></a><a name="p56159486469"></a>/var/log/mindx-dl/infer-operator/infer-operator.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>Log file. When a single log file exceeds 20 MB, the automatic dump function is triggered. The maximum file size cannot be modified. The naming format of the dumped file is: infer-operator-<time when dump is triggered/>.log, for example: infer-operator-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>Maximum number of log files retained after dump. The value range is 1 to 30, in number of files.</p>
</td>
</tr>
<tr id="row1758314497918"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1231919131198"><a name="p1231919131198"></a><a name="p1231919131198"></a>-isCompress</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p631912134197"><a name="p631912134197"></a><a name="p631912134197"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p031911314198"><a name="p031911314198"></a><a name="p031911314198"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1131914134197"><a name="p1131914134197"></a><a name="p1131914134197"></a>Whether to compress the log file for dump when the log file size reaches the dump threshold (this parameter will be deprecated in the future).</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true: Compress dump.</li><li>false: Do not compress dump.</li></ul>
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
