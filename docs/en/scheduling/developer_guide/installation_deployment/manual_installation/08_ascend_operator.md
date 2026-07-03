# Ascend Operator<a name="ZH-CN_TOPIC_0000002479386414"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:20:58.639Z pushedAt=2026-06-30T12:23:24.381Z -->

- Users who use full-NPU scheduling (training), static vNPU scheduling (training), resumable training, or elastic training must install Ascend Operator. If Volcano is used as the scheduler, it must be installed first; otherwise, Ascend Operator will fail to start.
- Users who use full-NPU scheduling (inference) and scheduling upon inference card faults to submit distributed inference job of the AscendJob type must install Ascend Operator.
- Users who only use containerization and resource monitoring, inference card fault recovery, or scheduling upon inference card faults (single-node jobs) do not need to install Ascend Operator. Please skip this chapter directly.
- Before starting Ascend Operator, Volcano must be started first. Ascend Operator needs to perceive PodGroup resources.
- The maximum number of replicas for a single AscendJob that Ascend Operator allows to create is 20,000.

## Procedure<a name="section209273712583"></a>

1. Log in to the K8s management node as the `root` user and run the following command to check whether the Ascend Operator image and version number are correct.

    ```shell
    docker images | grep ascend-operator
    ```

    The following is an example of the output:

    ```ColdFusion
    ascend-operator                      v26.0.0              c532e9d0889c        About an hour ago         137MB
    ```

    - If correct, proceed to [Step 2](#li19793191914420).
    - If not correct, see [Preparing Images](./01_preparing_for_installation.md) to complete image creation and distribution.

2. <a name="li19793191914420"></a>Copy the YAML files from the extracted Ascend Operator software package directory to any directory on the K8s management node.
3. If you do not modify the component startup parameters, you can skip this step. Otherwise, modify the Ascend Operator startup parameters in the YAML file based on the actual situation. For startup parameters, see [Table 1](#table11614104894617). You can run `./ascend-operator -h` to view parameter descriptions.
4. (Optional) Use Ascend Operator to generate the collective communication configuration file (RankTable File, also known as the [hccl.json](../../../api/hccl.json_file_description.md)) for training jobs under the PyTorch and MindSpore frameworks, to shorten the cluster communication link establishment time. Users of other frameworks can skip this step.
    1. The parent directory of the hccl.json file is already mounted by default in the startup YAML. Modify it based on the actual situation.

        <pre codetype="yaml">
        ...
                - name: ranktable-dir
                  mountPath: /user/mindx-dl/ranktable        # Container path, cannot be modified
        ...
              volumes:
                - name: ascend-operator-log
                  hostPath:
                    path: /var/log/mindx-dl/ascend-operator
                    type: Directory
                <strong>- name: ranktable-dir</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /user/mindx-dl/ranktable    # Host path. The root directory of the hccl.json file save path in the job YAML must be consistent with the host path.</strong>
                    <strong>type: DirectoryOrCreate                                      # Used to check whether the given folder exists. If it does not exist, an empty folder will be created.</strong>
        ...</pre>

        >[!NOTE]
        >- The root directory path of the RankTable inside the container cannot be modified, but the host path can be modified. When deploying a job, the root directory of the `hccl.json` file save path in the job YAML must be consistent with the host path.
        >- The permissions of the RankTable root directory folder must meet one of the following conditions:
        >    - The owning user and user group are `hwMindX` (the default running user of cluster scheduling components).
        >    - The permissions of the RankTable root directory folder are `777`.

    2. Run the following command to create the specific mount path for the `hccl.json` file in the parent directory.

        ```shell
        mkdir -m 777 /user/mindx-dl/ranktable/{Mount path}
        ```

5. In the path where the YAML file is located on the management node, run the following command to start Ascend Operator.

    ```shell
    kubectl apply -f ascend-operator-v{version}.yaml
    ```

    The following is a startup example:

    ```ColdFusion
    deployment.apps/ascend-operator-manager created
    serviceaccount/ascend-operator-manager created
    clusterrole.rbac.authorization.k8s.io/ascend-operator-manager-role created
    clusterrolebinding.rbac.authorization.k8s.io/ascend-operator-manager-rolebinding created
    customresourcedefinition.apiextensions.k8s.io/ascendjobs.mindxdl.gitee.com created
    ...
    ```

6. Run the following command to check whether the component has started successfully.

    ```shell
    kubectl get pod -n mindx-dl
    ```

    The following is an example of the output. **Running** indicates that the component startup is successful.

    ```ColdFusion
    NAME                                         READY   STATUS    RESTARTS   AGE
    ...
    ascend-operator-7667495b6b-hwmjw      1/1    Running  0         11s
    ```

>[!NOTE]
>
>- If the component Pod status is not `Running` after installation, see [Component Pod Status Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
>- If the component Pod status is `ContainerCreating` after installation, see [Cluster Scheduling Component Pod in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343).
>- If the component fails to start, see [Cluster Scheduling Component Startup Failure, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390).
>- If the component starts successfully but the corresponding Pod cannot be found, see [Component Startup YAML Executed Successfully but Corresponding Pod Not Found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Parameter Description<a name="section91521925121114"></a>

**Table 1** Ascend Operator startup parameters

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
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>Whether to query the <span id="ph446121313413"><a name="ph446121313413"></a><a name="ph446121313413"></a>Ascend Operator</span> version.</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true: Query.</li><li>false: Do not query.</li></ul>
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
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p56159486469"><a name="p56159486469"></a><a name="p56159486469"></a>/var/log/mindx-dl/ascend-operator/ascend-operator.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>Log file. When a single log file exceeds 20 MB, the automatic dump function is triggered. The maximum file size cannot be modified. The naming format of the dumped file is: ascend-operator-.log, for example: ascend-operator-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>Maximum number of log files retained after dumping. The value range is 1 to 30, in units of file count.</p>
</td>
</tr>
<tr id="row25282845417"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p155314286546"><a name="p155314286546"></a><a name="p155314286546"></a>-enableGangScheduling</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p19531128135415"><a name="p19531128135415"></a><a name="p19531128135415"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p55362825414"><a name="p55362825414"></a><a name="p55362825414"></a>true</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p3537285549"><a name="p3537285549"></a><a name="p3537285549"></a>Whether to enable "gang" policy scheduling. It is enabled by default. When enabled, task scheduling is performed based on the scheduler specified by the task. For details about "gang" policy scheduling, see the <a href="https://volcano.sh/zh/docs/v1-7-0/plugins/" target="_blank" rel="noopener noreferrer">open-source Volcano official documentation</a>.</p>
<a name="ul1161205685015"></a><a name="ul1161205685015"></a><ul id="ul1161205685015"><li>true: Enable "gang" policy scheduling.<p id="p1469315258274"><a name="p1469315258274"></a><a name="p1469315258274"></a>When using the job-level elastic scaling feature, the value of this field must be set to true.</p>
</li><li>false: Disable "gang" policy scheduling.</li></ul>
</td>
</tr>
<tr id="row1758314497918"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1231919131198"><a name="p1231919131198"></a><a name="p1231919131198"></a>-isCompress</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p631912134197"><a name="p631912134197"></a><a name="p631912134197"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p031911314198"><a name="p031911314198"></a><a name="p031911314198"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1131914134197"><a name="p1131914134197"></a><a name="p1131914134197"></a>Whether to compress and dump log files when the log file size reaches the dump threshold (this parameter will be deprecated later).</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true: compress and dump.</li><li>false: do not compress and dump.</li></ul>
</td>
</tr>
<tr id="row1636910277610"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p836962710617"><a name="p836962710617"></a><a name="p836962710617"></a>-kubeconfig</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p1536942715620"><a name="p1536942715620"></a><a name="p1536942715620"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p7628122811369"><a name="p7628122811369"></a><a name="p7628122811369"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1231013103910"><a name="p1231013103910"></a><a name="p1231013103910"></a>Path to kubeconfig. Must be configured when the program runs outside the cluster.</p>
</td>
</tr>
<tr id="row57381540134219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14739164054215"><a name="p14739164054215"></a><a name="p14739164054215"></a>-kubeApiBurst</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p0739184012425"><a name="p0739184012425"></a><a name="p0739184012425"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1273915409420"><a name="p1273915409420"></a><a name="p1273915409420"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5739114017421"><a name="p5739114017421"></a><a name="p5739114017421"></a>Burst traffic used when communicating with K8s. The value range is (0, 10000]. If the value is outside this range, the default value 100 is used.</p>
</td>
</tr>
<tr id="row182053596442"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p4205165917447"><a name="p4205165917447"></a><a name="p4205165917447"></a>-kubeApiQps</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p42051059174419"><a name="p42051059174419"></a><a name="p42051059174419"></a>float32</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17205159154412"><a name="p17205159154412"></a><a name="p17205159154412"></a>50</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p172054590444"><a name="p172054590444"></a><a name="p172054590444"></a>QPS (queries per second) used for communication with K8s. The value range is (0,10000]. If the value is outside this range, the default value of 50 is used.</p>
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
