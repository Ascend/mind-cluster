# Volcano<a name="ZH-CN_TOPIC_0000002479226394"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:14:22.672Z pushedAt=2026-06-09T06:22:06.854Z -->

## Installing Volcano<a name="ZH-CN_TOPIC_0000002511426351"></a>

- Users who use full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, resumable training, elastic training, inference card fault recovery, or inference card fault rescheduling must install a **scheduler** on the master node. This scheduler can be Volcano or another scheduler.
- If Volcano is used for job scheduling, it is not recommended to create or mount NPU containers through Docker or Containerd instructions and run jobs inside the containers. Otherwise, Volcano scheduling issues may be triggered.
- Users who only use containerization support and resource monitoring do not need to install Volcano. Please skip this chapter directly.

This chapter provides installation guidance for the Volcano components (vc-scheduler and vc-controller-manager) for cluster scheduling. If you need to use other components of open-source Volcano, install them yourself and ensure their security.

>[!NOTE]
>
>- In this document, Volcano refers to  Volcano provided by the cluster scheduler components by default. Other schedulers based on open-source Volcano can integrate the Ascend-volcano-plugin for developers to implement NPU scheduling-related functions. For details, see  [(Optional) Integrating Ascend Plugins to Extend Open-Source Volcano](#optional-integrating-ascend-plugins-to-extend-open-source-volcano).
>
>- NodeD versions 6.0.RC1 and later are incompatible with earlier versions of Volcano. If you are using NodeD 6.0.RC1 or later, you must use Volcano 6.0.RC1 or later.
>
>- For Volcano 6.0.RC2 and later, ClusterD must be installed by default. If ClusterD is not installed, the startup parameters of Volcano must be modified; otherwise, Volcano will not be able to schedule jobs normally.

- Managing the same node resources with Volcano, and other schedulers is not supported.

**Procedure<a name="section57241227172819"></a>**

1. Log in to the K8s master node as the `root` user and run the following command to check whether the Volcano image and version number are correct.

    ```shell
    docker images | grep volcanosh
    ```

    The following is an example of the response:

    ```ColdFusion
    volcanosh/vc-controller-manager      v1.7.0              84c73128cc55        3 days ago          44.5MB
    volcanosh/vc-scheduler               v1.7.0              e90c114c75b1        3 days ago          188MB
    ```

    - If correct, perform [Step 2](#li823273914318).
    - If not correct, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to complete image creation.

2. <a name="li823273914318"></a>Copy the YAML files from the extracted Volcano package directory to any directory on the Kubernetes master node.
3. If you do not modify the component startup parameters, you can skip this step. Otherwise, modify the Volcano startup parameters in the corresponding startup YAML file based on the actual situation. For common startup parameters, see [Table 4](#table5305150122116) and [Table 5](#table203077022111).
4. Configure Volcano log dumping.

    During installation, Volcano logs will be mounted to the disk space (`/var/log/mindx-dl`). By default, when the daily log write volume reaches 1.8 GB, Volcano will clear the log file. To prevent the space from being fully occupied, configure log dumping for Volcano. For information about the configuration items, see [Table 1](#table1123141112311), or choose a proper log dumping policy to avoid log loss.

    1. On the master node, in the `/etc/logrotate.d` directory, run the following command to create a log dumping configuration file.

        ```shell
        vi /etc/logrotate.d/File name
        ```

        For example:

        ```shell
        vi /etc/logrotate.d/volcano
        ```

        Write the following content, and then run the `:wq` command to save it.

        <pre>
        /var/log/mindx-dl/volcano-*/*.log{
             daily
             rotate 8
             size 50M
             compress
             dateext
             missingok
             notifempty
             copytruncate
             create 0640 hwMindX hwMindX
             sharedscripts
             postrotate
                 chmod 640 /var/log/mindx-dl/volcano-*/*.log
                 chmod 440 /var/log/mindx-dl/volcano-*/*.log-*
             endscript
        }</pre>

    2. Run the following commands in sequence to set the configuration file permissions to `640` and the owner to `root`.

        ```shell
        chmod 640 /etc/logrotate.d/File name
        chown root /etc/logrotate.d/File name
        ```

        For example:

        ```shell
        chmod 640 /etc/logrotate.d/volcano
        chown root /etc/logrotate.d/volcano
        ```

    **Table 1** Volcano log dumping configuration items

    <a name="table1123141112311"></a>
    <table><thead align="left"><tr id="row412371119316"><th class="cellrowborder" valign="top" width="20.352035203520348%" id="mcps1.2.4.1.1"><p id="p12123811163113"><a name="p12123811163113"></a><a name="p12123811163113"></a>Config Item</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.82338233823382%" id="mcps1.2.4.1.2"><p id="p8123141118315"><a name="p8123141118315"></a><a name="p8123141118315"></a>Description</p>
    </th>
    <th class="cellrowborder" valign="top" width="45.82458245824582%" id="mcps1.2.4.1.3"><p id="p16123121111319"><a name="p16123121111319"></a><a name="p16123121111319"></a>Optional Values</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row612391119318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p1012481112315"><a name="p1012481112315"></a><a name="p1012481112315"></a>daily</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p16124101111317"><a name="p16124101111317"></a><a name="p16124101111317"></a>Log dumping frequency.</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul1512431163118"></a><a name="ul1512431163118"></a><ul id="ul1512431163118"><li>daily: Performs a dumping check once a day.</li><li>weekly: Performs a dumping check once a week.</li><li>monthly: Performs a dumping check once a month.</li><li>yearly: Performs a dumping check once a year.</li></ul>
    </td>
    </tr>
    <tr id="row912511118314"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p111261711103117"><a name="p111261711103117"></a><a name="p111261711103117"></a>rotate <em id="i20126171193115"><a name="i20126171193115"></a><a name="i20126171193115"></a>x</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p13126191153117"><a name="p13126191153117"></a><a name="p13126191153117"></a>The number of times the log file is dumped before being deleted.</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p161261911153111"><a name="p161261911153111"></a><a name="p161261911153111"></a><em id="i161266119317"><a name="i161266119317"></a><a name="i161266119317"></a>x</em> is the number of backups.</p>
    <p id="p19126101103110"><a name="p19126101103110"></a><a name="p19126101103110"></a>For example:</p>
    <a name="ul151261211103117"></a><a name="ul151261211103117"></a><ul id="ul151261211103117"><li>rotate 0: No backup.</li><li>rotate 8: Retain 8 backups.</li></ul>
    </td>
    </tr>
    <tr id="row1912641115318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p15127111153115"><a name="p15127111153115"></a><a name="p15127111153115"></a>size <em id="i912731113110"><a name="i912731113110"></a><a name="i912731113110"></a>xx</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p412751115312"><a name="p412751115312"></a><a name="p412751115312"></a>The log file is dumped only when it reaches the specified size.</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p131273112314"><a name="p131273112314"></a><a name="p131273112314"></a>The size unit can be specified as:</p>
    <a name="ul1012771118311"></a><a name="ul1012771118311"></a><ul id="ul1012771118311"><li>byte (default)</li><li>K</li><li>M</li></ul>
    <p id="p1112761173115"><a name="p1112761173115"></a><a name="p1112761173115"></a>For example, size 50M means the log file is dumped when it reaches 50 MB.</p>
    <div class="note" id="note191277111311"><a name="note191277111311"></a><a name="note191277111311"></a><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><p id="p17127191193120"><a name="p17127191193120"></a><a name="p17127191193120"></a>logrotate periodically checks the log file size based on the configured dumping frequency. Only files whose size exceeds the size limit during the check will trigger dumping.</p>
    <p id="p112771153111"><a name="p112771153111"></a><a name="p112771153111"></a>This means that logrotate does not dump the log file immediately when it reaches the size limit.</p>
    </div></div>
    </td>
    </tr>
    <tr id="row4127111173111"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p161282115316"><a name="p161282115316"></a><a name="p161282115316"></a>compress</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p112851110317"><a name="p112851110317"></a><a name="p112851110317"></a>Whether to compress the dumped logs using gzip.</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul712814118312"></a><a name="ul712814118312"></a><ul id="ul712814118312"><li>compress: Use gzip compression.</li><li>nocompress: Do not use gzip compression.</li></ul>
    </td>
    </tr>
    <tr id="row18128511203117"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p412801153117"><a name="p412801153117"></a><a name="p412801153117"></a>notifempty</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p9128141123113"><a name="p9128141123113"></a><a name="p9128141123113"></a>Whether to dump empty files.</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul31281611163118"></a><a name="ul31281611163118"></a><ul id="ul31281611163118"><li>ifempty: Empty files are also dumped.</li><li>notifempty: Empty files do not trigger dumping.</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

5. (Optional) In `volcano-v{version}.yaml`, configure the CPU and memory required by Volcano. For recommended CPU and memory values, refer to the suggested values in the volcano-controller and volcano-scheduler tables in the [Open-Source Volcano Official Documentation](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html).

    <pre codetype="yaml">
    ...
    kind: Deployment
    ...
      labels:
        app: volcano-scheduler
    spec:
      replicas: 1
    ...
        spec:
    ...
              imagePullPolicy: "IfNotPresent"
              <strong>resources:
                requests:
                  memory: 4Gi
                  cpu: 5500m
                limits:
                  memory: 8Gi
                  cpu: 5500m</strong>
    ...
    kind: Deployment
    ...
      labels:
        app: volcano-controller
    spec:
    ...
        spec:
    ...
              <strong>resources:
                requests:
                  memory: 3Gi
                  cpu: 2000m
                limits:
                  memory: 3Gi
                  cpu: 2000m</strong>
    ...</pre>

6. (Optional) Tune scheduling time performance. It is supported to configure the plugins used by Volcano in `volcano-v{version}</i>.yaml`. Refer to the tables of Volcano advanced configuration parameters and supported plugins list in the [Open-Source Volcano Official Documentation](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html) for operation.

    ```Yaml
    ...
    data:
      volcano-scheduler.conf: |
        actions: "enqueue, allocate, backfill"
        tiers:
        - plugins:
          - name: priority
            enableNodeOrder: false
          - name: gang
            enableNodeOrder: false
          - name: conformance
            enableNodeOrder: false
          - name: volcano-npu_v26.0.0_linux-aarch64   # Here, v26.0.0 is the version number of MindCluster, and this value varies depending on the version.
        - plugins:
          - name: drf
            enableNodeOrder: false
          - name: predicates
            enableNodeOrder: false
            arguments:
              predicate.GPUSharingEnable: false
              predicate.GPUNumberEnable: false
          - name: proportion
            enableNodeOrder: false
          - name: nodeorder
          - name: binpack
            enableNodeOrder: false
    ....
    ```

7. (Optional) In `volcano-v{version}.yaml`, enable the Volcano health check interface and the Prometheus information collection interface.

    <pre codetype="yaml">
    ...
    kind: Deployment
    metadata:
      name: volcano-scheduler
      namespace: volcano-system
      labels:
        app: volcano-scheduler
    spec:
      ...
      template:
    ...
            - name: volcano-scheduler
              image: volcanosh/vc-scheduler:v1.7.0
              args: [ ...
                  ...
                  <strong>--enable-healthz=true   # To ensure normal access to the Volcano health check port, this parameter value must be "true"
                  --enable-metrics=true   # To ensure normal access to the Prometheus information collection port, this parameter value must be "true"</strong>
                  ...
    ...</pre>

    **Table 2** Description of open interfaces for the cluster scheduling Volcano component

    <a name="zh-cn_topic_0000001446965056_table173071368477"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001446965056_row153077618473"><th class="cellrowborder" valign="top" width="34.68346834683469%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001446965056_p3307116134715"><a name="zh-cn_topic_0000001446965056_p3307116134715"></a><a name="zh-cn_topic_0000001446965056_p3307116134715"></a>Access Method</p>
    </th>
    <th class="cellrowborder" valign="top" width="6.0906090609060906%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001446965056_p1525244211493"><a name="zh-cn_topic_0000001446965056_p1525244211493"></a><a name="zh-cn_topic_0000001446965056_p1525244211493"></a>Protocol</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.741174117411742%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001446965056_p04543391867"><a name="zh-cn_topic_0000001446965056_p04543391867"></a><a name="zh-cn_topic_0000001446965056_p04543391867"></a>Method</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.89168916891689%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001446965056_p23071468473"><a name="zh-cn_topic_0000001446965056_p23071468473"></a><a name="zh-cn_topic_0000001446965056_p23071468473"></a>Purpose</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.59305930593059%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001446965056_p730796134713"><a name="zh-cn_topic_0000001446965056_p730796134713"></a><a name="zh-cn_topic_0000001446965056_p730796134713"></a>Component</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001446965056_row23070613479"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p1730717615477"><a name="zh-cn_topic_0000001446965056_p1730717615477"></a><a name="zh-cn_topic_0000001446965056_p1730717615477"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252142154917"><a name="zh-cn_topic_0000001446965056_p10252142154917"></a><a name="zh-cn_topic_0000001446965056_p10252142154917"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p64546391612"><a name="zh-cn_topic_0000001446965056_p64546391612"></a><a name="zh-cn_topic_0000001446965056_p64546391612"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p151727316530"><a name="zh-cn_topic_0000001446965056_p151727316530"></a><a name="zh-cn_topic_0000001446965056_p151727316530"></a>Health check port</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p103089610475"><a name="zh-cn_topic_0000001446965056_p103089610475"></a><a name="zh-cn_topic_0000001446965056_p103089610475"></a>volcano-controller</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row1308176144715"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p2308865475"><a name="zh-cn_topic_0000001446965056_p2308865475"></a><a name="zh-cn_topic_0000001446965056_p2308865475"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p162523428499"><a name="zh-cn_topic_0000001446965056_p162523428499"></a><a name="zh-cn_topic_0000001446965056_p162523428499"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p245433916610"><a name="zh-cn_topic_0000001446965056_p245433916610"></a><a name="zh-cn_topic_0000001446965056_p245433916610"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p53084617475"><a name="zh-cn_topic_0000001446965056_p53084617475"></a><a name="zh-cn_topic_0000001446965056_p53084617475"></a>Health check port</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p14308176104718"><a name="zh-cn_topic_0000001446965056_p14308176104718"></a><a name="zh-cn_topic_0000001446965056_p14308176104718"></a>volcano-scheduler</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001446965056_row830812614472"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p19308116104714"><a name="zh-cn_topic_0000001446965056_p19308116104714"></a><a name="zh-cn_topic_0000001446965056_p19308116104714"></a>http://volcano-scheduler-serviceIP:8080/metrics</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252104224912"><a name="zh-cn_topic_0000001446965056_p10252104224912"></a><a name="zh-cn_topic_0000001446965056_p10252104224912"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p134543391765"><a name="zh-cn_topic_0000001446965056_p134543391765"></a><a name="zh-cn_topic_0000001446965056_p134543391765"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p193087624718"><a name="zh-cn_topic_0000001446965056_p193087624718"></a><a name="zh-cn_topic_0000001446965056_p193087624718"></a>Prometheus information collection port</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p3308166154716"><a name="zh-cn_topic_0000001446965056_p3308166154716"></a><a name="zh-cn_topic_0000001446965056_p3308166154716"></a>volcano-scheduler</p>
    </td>
    </tr>
    </tbody>
    </table>

8. (Optional) In `volcano-v{version}.yaml`, configure the mode for deleting Pods during rescheduling, virtualization mode, switch affinity scheduling, and whether to self-maintain available chip status.

    <pre codetype="yaml">
    ...
    data:
      volcano-scheduler.conf: |
    ...
        configurations:
          - name: init-params
            arguments: {<strong>"grace-over-time":"900","presetVirtualDevice":"true","nslb-version":"1.0","shared-tor-num":"2","useClusterInfoManager":"false","self-maintain-available-card":"true","super-pod-size": "48","reserve-nodes": "2","forceEnqueue":"true"</strong>}
    ...</pre>

    **Table 3**  Parameter description

    <a name="table208981646194315"></a>
    <table><thead align="left"><tr id="row08991746174316"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.1"><p id="p132621494445"><a name="p132621494445"></a><a name="p132621494445"></a>Parameter Name</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.2"><p id="p194862061467"><a name="p194862061467"></a><a name="p194862061467"></a>Default Value</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%" id="mcps1.2.4.1.3"><p id="p18991846144317"><a name="p18991846144317"></a><a name="p18991846144317"></a>Parameter Description</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1788817373541"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1888143725417"><a name="p1888143725417"></a><a name="p1888143725417"></a>grace-over-time</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p7888103725412"><a name="p7888103725412"></a><a name="p7888103725412"></a>900</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p145262285146"><a name="p145262285146"></a><a name="p145262285146"></a>Maximum time required to delete a Pod in graceful deletion mode during rescheduling, in seconds. The value ranges from 2 to 3600. Configuring this field indicates the use of graceful deletion mode for rescheduling. Graceful deletion means that during rescheduling, <span id="ph8305245165813"><a name="ph8305245165813"></a><a name="ph8305245165813"></a>Volcano</span> is waited on to perform related cleanup work. If the Pod is not successfully deleted after 900 seconds, the Pod is forcibly deleted directly without cleanup.</p>
    </td>
    </tr>
    <tr id="row95211735125411"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1352203555411"><a name="p1352203555411"></a><a name="p1352203555411"></a>presetVirtualDevice</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p552293515412"><a name="p552293515412"></a><a name="p552293515412"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p135221235105419"><a name="p135221235105419"></a><a name="p135221235105419"></a>Virtualization mode used.</p>
    <a name="ul206451443111219"></a><a name="ul206451443111219"></a><ul id="ul206451443111219"><li>true: Static virtualization</li><li>false: Dynamic virtualization</li></ul>
    </td>
    </tr>
    <tr id="row1589974674320"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p6899114619435"><a name="p6899114619435"></a><a name="p6899114619435"></a>nslb-version</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6484146114619"><a name="p6484146114619"></a><a name="p6484146114619"></a>1.0</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p7830165414514"><a name="p7830165414514"></a><a name="p7830165414514"></a>Version of switch affinity scheduling. The value can be 1.0 or 2.0.</p>
    <div class="note" id="note882315541054"><a name="note882315541054"></a><a name="note882315541054"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><a name="ul59831535122714"></a><a name="ul59831535122714"></a><ul id="ul59831535122714"><li>Switch affinity scheduling 1.0 supports <span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas training series products</span> and <span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 training series products</term></span>; supports <span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span> and <span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>.</li><li>Switch affinity scheduling 2.0 supports <span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 training series products</term></span>; supports the <span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span> framework.</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row8899946174318"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p168998463434"><a name="p168998463434"></a><a name="p168998463434"></a>shared-tor-num</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p989910464439"><a name="p989910464439"></a><a name="p989910464439"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p925113215214"><a name="p925113215214"></a><a name="p925113215214"></a>Maximum number of shared switches that a single task can use in switch affinity scheduling 2.0. The value can be 1 or 2. This parameter takes effect only when nslb-version is set to 2.0.</p>
    <p id="p1856962434719"><a name="p1856962434719"></a><a name="p1856962434719"></a>For details about switch affinity scheduling (1.0 or 2.0), see the <a href="../../../usage/basic_scheduling/01_affinity_scheduling/04_node_based_affinity.md">Node-based Affinity</a> chapter.</p>
    </td>
    </tr>
    <tr id="row797916276295"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p9621013114312"><a name="p9621013114312"></a><a name="p9621013114312"></a>useClusterInfoManager</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p1024875418187"><a name="p1024875418187"></a><a name="p1024875418187"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1797510751015"><a name="p1797510751015"></a><a name="p1797510751015"></a>Method by which <span id="ph18393155819297"><a name="ph18393155819297"></a><a name="ph18393155819297"></a>Volcano</span> obtains cluster information. The value description is as follows:</p>
    <a name="ul675021361014"></a><a name="ul675021361014"></a><ul id="ul675021361014"><li>true: Reads the <span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ConfigMap</span> reported by <span id="ph1921415457302"><a name="ph1921415457302"></a><a name="ph1921415457302"></a>ClusterD</span>.</li><li>false: Reads the <span id="ph19274234236"><a name="ph19274234236"></a><a name="ph19274234236"></a>ConfigMap</span> reported by <span id="ph144095321390"><a name="ph144095321390"></a><a name="ph144095321390"></a>Ascend Device Plugin</span> and <span id="ph039324431114"><a name="ph039324431114"></a><a name="ph039324431114"></a>NodeD</span> respectively.</li></ul>
    <div class="note" id="note1466414341216"><a name="note1466414341216"></a><a name="note1466414341216"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p1166463181214"><a name="p1166463181214"></a><a name="p1166463181214"></a>By default, the <span id="ph19101421151220"><a name="ph19101421151220"></a><a name="ph19101421151220"></a>ConfigMap</span> reported by the <span id="ph139579361121"><a name="ph139579361121"></a><a name="ph139579361121"></a>ClusterD</span> component is read. In future versions, reading the <span id="ph3588183951516"><a name="ph3588183951516"></a><a name="ph3588183951516"></a>ConfigMap</span> reported by <span id="ph1758893981514"><a name="ph1758893981514"></a><a name="ph1758893981514"></a>Ascend Device Plugin</span> and <span id="ph75887392157"><a name="ph75887392157"></a><a name="ph75887392157"></a>NodeD</span> will no longer be supported.</p>
    </div></div>
    </td>
    </tr>
    <tr id="row1913114164518"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p91454144514"><a name="p91454144514"></a><a name="p91454144514"></a>self-maintain-available-card</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p814241174517"><a name="p814241174517"></a><a name="p814241174517"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p151414164516"><a name="p151414164516"></a><a name="p151414164516"></a>Whether Volcano self-maintains the available chip status. The value description is as follows:</p>
    <a name="ul299044019472"></a><a name="ul299044019472"></a><ul id="ul299044019472"><li>true: Volcano self-maintains the available chip status.</li><li>false: Volcano obtains the available chip status based on the <span id="ph98552414486"><a name="ph98552414486"></a><a name="ph98552414486"></a>ConfigMap</span> reported by ClusterD or <span id="ph1185824104819"><a name="ph1185824104819"></a><a name="ph1185824104819"></a>Ascend Device Plugin</span>.</li></ul>
    </td>
    </tr>
    <tr id="row4612538250"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p26130381510"><a name="p26130381510"></a><a name="p26130381510"></a>super-pod-size</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6613173818516"><a name="p6613173818516"></a><a name="p6613173818516"></a>48</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p461323812519"><a name="p461323812519"></a><a name="p461323812519"></a>Number of nodes in one SuperPoD of <span id="ph128111331314"><a name="ph128111331314"></a><a name="ph128111331314"></a>Atlas 900 A3 SuperPoD</span>.</p>
    </td>
    </tr>
    <tr id="row9561856657"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p956215565514"><a name="p956215565514"></a><a name="p956215565514"></a>reserve-nodes</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p125626567516"><a name="p125626567516"></a><a name="p125626567516"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p25637568515"><a name="p25637568515"></a><a name="p25637568515"></a>Number of reserved nodes in one SuperPoD of <span id="ph915032251212"><a name="ph915032251212"></a><a name="ph915032251212"></a>Atlas 900 A3 SuperPoD</span>.</p>
    <div class="note" id="note1514175285210"><a name="note1514175285210"></a><a name="note1514175285210"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p96481321115510"><a name="p96481321115510"></a><a name="p96481321115510"></a>If the configured reserve-nodes is greater than super-pod-size, the following scenarios exist.</p>
    <a name="ul13842528165510"></a><a name="ul13842528165510"></a><ul id="ul13842528165510"><li>If super-pod-size is greater than 2, reserve-nodes is reset to 2 by default.</li><li>If super-pod-size is less than or equal to 2, reserve-nodes is reset to 0 by default.</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row1890722719501"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p590882716507"><a name="p590882716507"></a><a name="p590882716507"></a><span id="ph19180940145012"><a name="ph19180940145012"></a><a name="ph19180940145012"></a>forceEnqueue</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p17908162765017"><a name="p17908162765017"></a><a name="p17908162765017"></a><span id="ph16315161885115"><a name="ph16315161885115"></a><a name="ph16315161885115"></a>true</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1790852711505"><a name="p1790852711505"></a><a name="p1790852711505"></a><span id="ph188121729115116"><a name="ph188121729115116"></a><a name="ph188121729115116"></a>Whether a task is forcibly </span><span id="ph280124455118"><a name="ph280124455118"></a><a name="ph280124455118"></a>entered into the scheduling queue</span><span id="ph1179814511514"><a name="ph1179814511514"></a><a name="ph1179814511514"></a> when the cluster NPU resources are sufficient.</span><span id="ph2278145215114"><a name="ph2278145215114"></a><a name="ph2278145215114"></a> The value description is as follows:</span></p>
    <a name="ul12820554135117"></a><a name="ul12820554135117"></a><ul id="ul12820554135117"><li>true: When Volcano enables the <span id="ph11766123385220"><a name="ph11766123385220"></a><a name="ph11766123385220"></a>Enqueue</span> action, if the cluster NPU resources meet the requirements of the current task, the task is forcibly <span id="ph3349644125319"><a name="ph3349644125319"></a><a name="ph3349644125319"></a>entered into the scheduling queue</span><span id="ph22191237135316"><a name="ph22191237135316"></a><a name="ph22191237135316"></a></span>, regardless of whether other resources are sufficient. If the current task remains in the scheduling queue for a long time, it pre-occupies resources, which may prevent other tasks from entering the queue.</li><li>Other values: When cluster NPU resources are insufficient, the task is rejected from <span id="ph6205121155415"><a name="ph6205121155415"></a><a name="ph6205121155415"></a>entering the scheduling queue. If</span> NPU resources meet the requirements of the current task, all plugins jointly decide whether to <span id="ph370210413554"><a name="ph370210413554"></a><a name="ph370210413554"></a>enter the scheduling queue</span>.</li></ul>
    <p id="p12691948115614"><a name="p12691948115614"></a><a name="p12691948115614"></a>For a detailed description of this parameter, see <a href="https://volcano.sh/docs/v1.9.0/Scheduler/Actions" target="_blank" rel="noopener noreferrer">Volcano Actions</a>.</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p>resource-level-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p>The default value is empty.</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p>Json configuration of the network resource levels corresponding to nodes in the cluster.</p><p>This parameter is used only in multi-level scheduling tasks. It is not required for non-multi-level scheduling tasks.</p><p>The value description is as follows:</p><ul><li>The first-level key of the Json file is the name of the network topology tree, and the corresponding value is the detailed definition of the network topology tree.</li><li>In the detailed definition of the network topology tree, the key is used to identify a specific network level. The value is the prefix level plus the network level sequence number n, where n is a positive integer greater than or equal to 1. The value is the corresponding network level definition structure.</li><li>In the network level definition structure, the following fields exist:<ul><li>label: Identifies the key of the node label at this network level. The value is a string.</li><li>reservedNode: Identifies the number of reserved sub-level nodes. The value is an integer and takes effect only in the level1 configuration. During multi-level scheduling task scheduling, scheduling is preferentially attempted after deducting the reserved number of nodes. If scheduling cannot be performed after deducting the reserved nodes, the reserved node resources are used normally.</li></ul></li></ul><p>For a detailed description and examples of this parameter, see <a href="../../../usage/basic_scheduling/05_multi_level_scheduling.md#configuring-volcano-startup-parameters">Configuring Volcano Startup Parameters</a>.</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE]
    >- For more configuration about open-source Volcano, see the [open-source Volcano official documentation](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html).
    >- K8s supports node affinity scheduling using the nodeAffinity field. For a detailed description of this field, see the [Kubernetes official documentation](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/). Volcano also supports this field. For operation instructions, see the [Scheduling Configuration](../../../common_operations.md#scheduling-configuration) chapter.

9. (Optional) Tune scheduling time performance. Volcano supports optimizing the scheduling time for a single job (training vcjob or acjob) with 4,000 or 5,000 Pods to approximately 5 minutes when scheduling them onto 4,000 or 5,000 nodes. If you want to use this scheduling feature, make the following modifications in `volcano-v{version}.yaml`.

    - To achieve the reference time of approximately 5 minutes, ensure that the CPU frequency is at least 2.60 GHz and the APIServer latency does not exceed 80 milliseconds.
    - If you do not use native K8s `nodeAffinity` and `podAntiAffinity` for scheduling, you can disable the `nodeorder` plugin to further reduce the scheduling time.

    <pre codetype="yaml">
    data:
      volcano-scheduler.conf: |

    ...
          - name: proportion
            enableNodeOrder: false
          - name: nodeorder
            <strong>enableNodeOrder: false     # Optional. When nodeAffinity and podAntiAffinity scheduling are not used, you can disable the nodeorder plugin.</strong>
    ...
          containers:
            - name: volcano-scheduler
              image: volcanosh/vc-scheduler:v1.7.0
              command: ["/bin/ash"]
              args: ["-c", "umask 027; <strong>GOMEMLIMIT=15000000000 GOGC=off</strong> /vc-scheduler      <strong># Add the GOMEMLIMIT=15000000000 and GOGC=off fields</strong>
                      --scheduler-conf=/volcano.scheduler/volcano-scheduler.conf
                      --plugins-dir=plugins
                      --logtostderr=false
                      --log_dir=/var/log/mindx-dl/volcano-scheduler
                      --log_file=/var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
                      -v=2 2>&1"]
              imagePullPolicy: "IfNotPresent"
              resources:
                requests:
                  <strong>memory: 10000Mi                                                                # Change 4Gi to 10000Mi</strong>
                  cpu: 5500m
                limits:
                  <strong>memory: 15000Mi                                                       # Change 8Gi to 15000Mi</strong>
                  cpu: 5500m
    ...</pre>

10. In the path where the YAML file is located on the master node, run the following command to start Volcano.

    ```shell
    kubectl apply -f volcano-v{version}.yaml
    ```

    Example:

    ```ColdFusion
    namespace/volcano-system created
    namespace/volcano-monitoring created
    configmap/volcano-scheduler-configmap created
    serviceaccount/volcano-scheduler created
    clusterrole.rbac.authorization.k8s.io/volcano-scheduler created
    clusterrolebinding.rbac.authorization.k8s.io/volcano-scheduler-role created
    deployment.apps/volcano-scheduler created
    service/volcano-scheduler-service created
    serviceaccount/volcano-controllers created
    clusterrole.rbac.authorization.k8s.io/volcano-controllers created
    clusterrolebinding.rbac.authorization.k8s.io/volcano-controllers-role created
    deployment.apps/volcano-controllers created
    customresourcedefinition.apiextensions.k8s.io/jobs.batch.volcano.sh created
    customresourcedefinition.apiextensions.k8s.io/commands.bus.volcano.sh created
    customresourcedefinition.apiextensions.k8s.io/podgroups.scheduling.volcano.sh created
    customresourcedefinition.apiextensions.k8s.io/queues.scheduling.volcano.sh created
    customresourcedefinition.apiextensions.k8s.io/numatopologies.nodeinfo.volcano.sh created
    ```

11. Run the following command to check the component status.

    ```shell
    kubectl get pod -n volcano-system
    ```

    The following is an example of the response. `Running` indicates that the component startup is successful:

    ```ColdFusion
    NAME                                          READY    STATUS     RESTARTS     AGE
    volcano-controllers-5cf8d788d5-qdpzq   1/1     Running   0          1m
    volcano-scheduler-6cffd555c9-45k7c     1/1     Running   0          1m
    ```

    >[!NOTE]
    >- If the Volcano Pod status is `CrashLoopBackOff`, see [Pod Status is CrashLoopBackOff After Manually Installing Volcano](https://gitcode.com/Ascend/mind-cluster/issues/347) for troubleshooting.
    >- If the `volcano-scheduler-6cffd555c9-45k7c` status is `Running` but scheduling is abnormal, see [Volcano Component Works Abnormally, Log Shows "Failed to get plugin"](https://gitcode.com/Ascend/mind-cluster/issues/348) for troubleshooting.
    >- If the component Pod status is not `Running` after installation, see [Component Pod Status Is Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for troubleshooting.
    >- If the component Pod status is `ContainerCreating` after installation, see [Cluster Scheduler Component Pod Is in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343) for troubleshooting.
    >- If the component fails to start, see [Failed to Start Cluster Scheduler Component, Log Prints "get sem errno =13"](https://gitcode.com/Ascend/mind-cluster/issues/390) for information.
    >- If the component starts successfully but the corresponding Pod cannot be found, see [Component Startup YAML Executes Successfully but Corresponding Pod Cannot Be Found](https://gitcode.com/Ascend/mind-cluster/issues/345) for information.

**Parameter Description<a name="section1317934882010"></a>**

**Table 4**  volcano-scheduler startup parameters

<a name="table5305150122116"></a>
<table><thead align="left"><tr id="row63052016218"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p133052042113"><a name="p133052042113"></a><a name="p133052042113"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="15.310000000000002%" id="mcps1.2.5.1.2"><p id="p330560162111"><a name="p330560162111"></a><a name="p330560162111"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="29.69%" id="mcps1.2.5.1.3"><p id="p3305600215"><a name="p3305600215"></a><a name="p3305600215"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p63067062115"><a name="p63067062115"></a><a name="p63067062115"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row12306160112118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p14306100102116"><a name="p14306100102116"></a><a name="p14306100102116"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p103066014211"><a name="p103066014211"></a><a name="p103066014211"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p13306150192120"><a name="p13306150192120"></a><a name="p13306150192120"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2030616002117"><a name="p2030616002117"></a><a name="p2030616002117"></a>Log directory. The default value in the component startup YAML is /var/log/mindx-dl/volcano-scheduler.</p>
</td>
</tr>
<tr id="row230620102115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143064002118"><a name="p143064002118"></a><a name="p143064002118"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p173067012119"><a name="p173067012119"></a><a name="p173067012119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p430620132116"><a name="p430620132116"></a><a name="p430620132116"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p930615020218"><a name="p930615020218"></a><a name="p930615020218"></a>Log file name. The default value in the component startup YAML is /var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log.</p>
<div class="note" id="note19596191219291"><a name="note19596191219291"></a><a name="note19596191219291"></a><div class="notebody"><p id="p10596012112919"><a name="p10596012112919"></a><a name="p10596012112919"></a>The naming format for dumped files is: volcano-scheduler.log-<trigger dumping time/>.gz, for example: volcano-scheduler.log-20230926.gz.</p>
</div></div>
</td>
</tr>
<tr id="row17922126205817"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p139228267582"><a name="p139228267582"></a><a name="p139228267582"></a>--scheduler-conf</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p17490165855810"><a name="p17490165855810"></a><a name="p17490165855810"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p192312261580"><a name="p192312261580"></a><a name="p192312261580"></a>/volcano.scheduler/volcano-scheduler.conf</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1192372695818"><a name="p1192372695818"></a><a name="p1192372695818"></a>Absolute path of the scheduler component configuration file.</p>
</td>
</tr>
<tr id="row630618042113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p173061701214"><a name="p173061701214"></a><a name="p173061701214"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p730613011217"><a name="p730613011217"></a><a name="p730613011217"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p0306170112117"><a name="p0306170112117"></a><a name="p0306170112117"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p430613012117"><a name="p430613012117"></a><a name="p430613012117"></a>Whether to print logs to standard output.</p>
<a name="ul582374031615"></a><a name="ul582374031615"></a><ul id="ul582374031615"><li>true: Print.</li><li>false: Do not print.</li></ul>
</td>
</tr>
<tr id="row53063062118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p330618022113"><a name="p330618022113"></a><a name="p330618022113"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p133061010218"><a name="p133061010218"></a><a name="p133061010218"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p23068042118"><a name="p23068042118"></a><a name="p23068042118"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p43067042115"><a name="p43067042115"></a><a name="p43067042115"></a>Log output level:</p>
<a name="ul03064012212"></a><a name="ul03064012212"></a><ul id="ul03064012212"><li>Value is 1: error</li><li>Value is 2: warning</li><li>Value is 3: info</li><li>Value is 4: debug</li></ul>
</td>
</tr>
<tr id="row11306140152113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p730614015211"><a name="p730614015211"></a><a name="p730614015211"></a>--plugins-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1130614013214"><a name="p1130614013214"></a><a name="p1130614013214"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p12307200192115"><a name="p12307200192115"></a><a name="p12307200192115"></a>plugins</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p173071603217"><a name="p173071603217"></a><a name="p173071603217"></a>Scheduler plugin loading path.</p>
</td>
</tr>
<tr id="row113072012113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9307140142120"><a name="p9307140142120"></a><a name="p9307140142120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p03072016212"><a name="p03072016212"></a><a name="p03072016212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p330712011215"><a name="p330712011215"></a><a name="p330712011215"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p53071209215"><a name="p53071209215"></a><a name="p53071209215"></a>Whether to query the volcano-scheduler binary version number.</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true: Query.</li><li>false: Do not query.</li></ul>
</td>
</tr>
<tr id="row62114943417"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p182349173416"><a name="p182349173416"></a><a name="p182349173416"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1221849193415"><a name="p1221849193415"></a><a name="p1221849193415"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p1021949203420"><a name="p1021949203420"></a><a name="p1021949203420"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1321749193419"><a name="p1321749193419"></a><a name="p1321749193419"></a>Maximum storage size of the log file (in MB).</p>
<div class="note" id="note1919311416364"><a name="note1919311416364"></a><a name="note1919311416364"></a><div class="notebody"><p id="p7193444361"><a name="p7193444361"></a><a name="p7193444361"></a>When the log file size exceeds the threshold, the log content will be cleared.</p>
</div></div>
</td>
</tr>
<tr id="row159867311462"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10986173174613"><a name="p10986173174613"></a><a name="p10986173174613"></a>--leader-elect</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1098619374617"><a name="p1098619374617"></a><a name="p1098619374617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p19866311462"><a name="p19866311462"></a><a name="p19866311462"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p4986143184611"><a name="p4986143184611"></a><a name="p4986143184611"></a>Start leader election mode when starting with multiple replicas.</p>
</td>
</tr>
<tr id="row1253065634617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1453015644610"><a name="p1453015644610"></a><a name="p1453015644610"></a>--percentage-nodes-to-find</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p145301156194612"><a name="p145301156194612"></a><a name="p145301156194612"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p16530175615462"><a name="p16530175615462"></a><a name="p16530175615462"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p11530165644617"><a name="p11530165644617"></a><a name="p11530165644617"></a>Percentage of available nodes to select from the total cluster nodes during task scheduling.</p>
</td>
</tr>
</tbody>
</table>

**Table 5**  volcano-controller startup parameters

<a name="table203077022111"></a>
<table><thead align="left"><tr id="row18307705217"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p193071001218"><a name="p193071001218"></a><a name="p193071001218"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p13307208218"><a name="p13307208218"></a><a name="p13307208218"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p123078062120"><a name="p123078062120"></a><a name="p123078062120"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p4307100172120"><a name="p4307100172120"></a><a name="p4307100172120"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row173077014210"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p43078015211"><a name="p43078015211"></a><a name="p43078015211"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p173071104213"><a name="p173071104213"></a><a name="p173071104213"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p113071302218"><a name="p113071302218"></a><a name="p113071302218"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330718019213"><a name="p330718019213"></a><a name="p330718019213"></a>Log directory. The default value in the component startup YAML is /var/log/mindx-dl/volcano-controller.</p>
</td>
</tr>
<tr id="row1307170112113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p17307130112117"><a name="p17307130112117"></a><a name="p17307130112117"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1630780182118"><a name="p1630780182118"></a><a name="p1630780182118"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1930714062115"><a name="p1930714062115"></a><a name="p1930714062115"></a>None</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p143077018217"><a name="p143077018217"></a><a name="p143077018217"></a>Log file name. The default value in the component startup YAML is /var/log/mindx-dl/volcano-controller/volcano-controller.log.</p>
<div class="note" id="note215144410296"><a name="note215144410296"></a><a name="note215144410296"></a><div class="notebody"><p id="p715144132910"><a name="p715144132910"></a><a name="p715144132910"></a>The naming format for files after dumping is: volcano-controller.log-<trigger dumping time/>.gz, for example: volcano-controller.log-20230926.gz.</p>
</div></div>
</td>
</tr>
<tr id="row730760202120"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p93071805219"><a name="p93071805219"></a><a name="p93071805219"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p730812011211"><a name="p730812011211"></a><a name="p730812011211"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p2308140142118"><a name="p2308140142118"></a><a name="p2308140142118"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2308170172116"><a name="p2308170172116"></a><a name="p2308170172116"></a>Whether to print logs to standard output.</p>
<a name="ul142362048125710"></a><a name="ul142362048125710"></a><ul id="ul142362048125710"><li>true: Print.</li><li>false: Do not print.</li></ul>
</td>
</tr>
<tr id="row930819012214"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p193088092115"><a name="p193088092115"></a><a name="p193088092115"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1930812016213"><a name="p1930812016213"></a><a name="p1930812016213"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p123081003218"><a name="p123081003218"></a><a name="p123081003218"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330830162118"><a name="p330830162118"></a><a name="p330830162118"></a>Log output level:</p>
<a name="ul6308150112119"></a><a name="ul6308150112119"></a><ul id="ul6308150112119"><li>1: error</li><li>2: warning</li><li>3: info</li><li>4: debug</li></ul>
</td>
</tr>
<tr id="row1330813015217"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p133085052120"><a name="p133085052120"></a><a name="p133085052120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p030814011212"><a name="p030814011212"></a><a name="p030814011212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p10308140122115"><a name="p10308140122115"></a><a name="p10308140122115"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p130818011219"><a name="p130818011219"></a><a name="p130818011219"></a>Binary version number of volcano-controller.</p>
</td>
</tr>
<tr id="row926534763719"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1413064912376"><a name="p1413064912376"></a><a name="p1413064912376"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p313074910373"><a name="p313074910373"></a><a name="p313074910373"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p31301349183714"><a name="p31301349183714"></a><a name="p31301349183714"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p111301349113715"><a name="p111301349113715"></a><a name="p111301349113715"></a>Maximum storage size of the log file (in MB).</p>
<div class="note" id="note1513064943719"><a name="note1513064943719"></a><a name="note1513064943719"></a><div class="notebody"><p id="p111317492373"><a name="p111317492373"></a><a name="p111317492373"></a>When the log file size exceeds the threshold, the log content will be cleared.</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>Volcano is open-source software. Only common startup parameters currently in use are listed here. For other detailed parameters, refer to the open-source software documentation.

## (Optional) Using Volcano Switch Affinity Scheduling<a name="ZH-CN_TOPIC_0000002479226480"></a>

Volcano supports switch affinity scheduling. To use this feature, you need to upload the mapping between switches and servers for Volcano to use. The steps are as follows.

>[!NOTE]
>Currently, only full-NPU switch affinity scheduling is supported for training and inference tasks. Static or dynamic vNPU scheduling is not supported.

**Procedure<a name="section7172163412209"></a>**

1. <a name="li6319161364017"></a>Prepare the LLD document for network design of the deployment environment, and upload it to any directory on the K8s master node (using `/home/tor-affinity` as an example).

    >[!NOTE]
    >The LLD file name must be lld.xlsx.

2. Obtain the LLD document parsing script.

    Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository, and enter the branch corresponding to the version according to [mindcluster-deploy Open-Source Repository Version Description](../../../appendix.md#mindcluster-deploy-open-source-repository-version-description). Download the `lld_to_cm.py` file from the `samples/utils` directory, and upload the file to the directory on the master node mentioned in [Step 1](#li6319161364017).

3. Run the following command to start the `lld_to_cm.py` script.

    ```shell
    python ./lld_to_cm.py --num 32
    ```

    >[!NOTE]
    >- Use the `--num` (or -`n`) subcommand to specify the number of nodes under a switch. If this parameter is not specified, the default value is `4`.
    >- Use the `--level` (or `-l`) subcommand to specify the switch networking type. If this parameter is not specified, the default value is `double_layer`.
    >    - `single_layer`: Uses single-layer switch networking.
    >    - `double_layer`: Uses double-layer switch networking.
    >- This script requires the openpyxl module. If the installation environment lacks this module, you can run the `pip install openpyxl` command to install it.

4. Run the following command to check whether the ConfigMap is created successfully.

    ```shell
    kubectl get cm -n kube-system basic-tor-node-cm
    ```

    The following is a command output example, indicating that the creation is successful.

    ```ColdFusion
    NAME                DATA   AGE
    basic-tor-node-cm   1      8s
    ```

**Configuring Switch Affinity Scheduling<a name="section125904488511"></a>**

To configure switch affinity scheduling, you need to configure the `tor-affinity` parameter in the job YAML. The location and configuration description of `tor-affinity` are shown in the following table.

**Table 1** YAML parameter description

<a name="table325141716575"></a>

|Parameter|Value|Description|
|--|--|--|
|(.kind=="AscendJob").metadata.labels.tor-affinity|<ul><li>large-model-schema: Large model job or padding job</li><li>normal-schema: Normal job</li><li>null: Switch affinity scheduling is not used<div class="note"><span class="notetitle">[!NOTE] Description</span><div class="notebody"><p>Users need to select the job type based on the number of job replicas. If the number of job replicas is less than 4, it is a padding job. If the number of job replicas is greater than or equal to 4, it is a large model job. Normal jobs have no restrictions on the number of job replicas.</p></div></div></li></ul>|<p>The default value is null, indicating that switch affinity scheduling is not used. Users need to configure it based on the job type.</p><ul><li>Switch affinity scheduling 1.0 supports Atlas training series products and <term>Atlas A2 training series products</term>; supports PyTorch and MindSpore frameworks.</li><li>Switch affinity scheduling 2.0 supports <term>Atlas A2 training series products</term>; supports the PyTorch framework.</li><li>Only full cards are supported for switch affinity scheduling. Static vNPU is not supported for switch affinity scheduling.</li></ul>|

## (Optional) Integrating Ascend Plugins to Extend Open-Source Volcano<a name="ZH-CN_TOPIC_0000002511426365"></a>

The cluster scheduling Volcano compomet adds NPU scheduling-related features on top of the open-source Volcano. This functionality can be implemented by integrating the Ascend-volcano-plugin provided by cluster scheduling for developers. The open-source [Volcano](https://volcano.sh/docs/v1.9.0/Home/Introduction) framework supports a plugin mechanism for users to register scheduling plugins and implement different scheduling policies.

>[!NOTE]
>Ascend-volcano-plugin currently supports open-source Volcano v1.7.0 and v1.9.0, and no modifications have been made to the open-source Volcano framework.

**Procedure<a name="section2672154791712"></a>**

1. Run the following commands in sequence to pull the official Volcano (using v1.7 as an example) open-source code in the `$GOPATH/src/volcano.sh/` directory.

    ```shell
    mkdir -p $GOPATH/src/volcano.sh/
    cd $GOPATH/src/volcano.sh/
    git clone -b release-1.7 https://github.com/volcano-sh/volcano.git
    ```

2. Rename the obtained [ascend-for-volcano](https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-for-volcano) source code to ascend-volcano-plugin, and upload it to the plugin path of the official open-source Volcano code (`_$GOPATH_/src/volcano.sh/volcano/pkg/scheduler/plugins/`).
3. <a name="li627818212613"></a>Run the following commands in sequence to compile the open-source Volcano binary files and the Huawei NPU scheduling plugin .so file. Select the corresponding parameter for the `build.sh` script based on the open-source code version, for example, v1.7.0.

    ```shell
    cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
    chmod +x build.sh
    ./build.sh v1.7.0
    ```

    >[!NOTE]
    >The compiled binary files and dynamic link library files are located in the `$GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output` directory.

    For the list of compiled files, see [Table 1](#table5623201371819).

    **Table 1** Files in the output path

    <a name="table5623201371819"></a>

    |File Name|Description|
    |--|--|
    |volcano-npu-<em>{version}</em>.so|Huawei NPU scheduling plugin dynamic link library|
    |Dockerfile-scheduler|volcano-scheduler image build text file|
    |Dockerfile-controller|volcano-controller image build text file|
    |volcano-<em>v{version}</em>.yaml|Volcano startup configuration file|
    |vc-scheduler|volcano-scheduler component binary file|
    |vc-controller-manager|volcano-controller component binary file|

4. Choose one of the following two methods to start the volcano-scheduler component.
    - Using the startup YAML of the MindCluster Volcano
        1. Run the following command to build the Volcano image. Select the corresponding parameter for the image based on the open-source code version, such as v1.7.0.

            ```shell
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
            ```

        2. Run the following command to start the volcano-scheduler component.

            ```shell
            kubectl apply -f volcano-v{version}.yaml
            ```

            Example:

            ```ColdFusion
            namespace/volcano-system created
            namespace/volcano-monitoring created
            configmap/volcano-scheduler-configmap created
            serviceaccount/volcano-scheduler created
            clusterrole.rbac.authorization.k8s.io/volcano-scheduler created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-scheduler-role created
            deployment.apps/volcano-scheduler created
            service/volcano-scheduler-service created
            serviceaccount/volcano-controllers created
            clusterrole.rbac.authorization.k8s.io/volcano-controllers created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-controllers-role created
            deployment.apps/volcano-controllers created
            customresourcedefinition.apiextensions.k8s.io/jobs.batch.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/commands.bus.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/podgroups.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/queues.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/numatopologies.nodeinfo.volcano.sh created
            ```

    - Using the startup YAML of the open-source Volcano
        1. Copy the `volcano-npu-_{version}_.so` file compiled in [Step 3](#li627818212613) to the `$GOPATH/src/volcano.sh/volcano` directory of the open-source Volcano; add the following command to the Dockerfile of the open-source Volcano (path: `$GOPATH/src/volcano.sh/volcano/installer/dockerfile/scheduler/Dockerfile`).

            ```shell
            FROM golang:1.19.1 AS builder
            WORKDIR /go/src/volcano.sh/
            ADD . volcano
            RUN cd volcano && make vc-scheduler
            FROM alpine:latest
            COPY --from=builder /go/src/volcano.sh/volcano/_output/bin/vc-scheduler /vc-scheduler
            COPY volcano-npu_*.so plugins/     # Add
            ENTRYPOINT ["/vc-scheduler"]
            ```

        2. Run the following commands in sequence to build the Volcano image. Select the corresponding tag for the image based on the open-source code version, such as v1.7.0.

            ```shell
            cd $GOPATH/src/volcano.sh/volcano
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f installer/dockerfile/scheduler/Dockerfile
            ```

        3. Modify `volcano-development.yaml`. The file path is `$GOPATH/src/volcano.sh/volcano/installer/volcano-development.yaml`.

            <pre codetype="yaml">
            apiVersion: v1
            kind: ConfigMap
            metadata:
              name: volcano-scheduler-configmap
              namespace: volcano-system
            data:
               volcano-scheduler.conf: |
                 actions: "enqueue, allocate, backfill"
                 tiers:
                 - plugins:
                   - name: priority
                   - name: gang
                     enablePreemptable: false
                   - name: conformance
                   <strong>- name: volcano-npu_v26.0.0_linux-x86_64    # Custom scheduling plugin added in ConfigMap. Ensure version compatibility between components.</strong>
                 - plugins:
                   - name: overcommit
                   - name: drf
                     enablePreemptable: false
                   - name: predicates
                   - name: proportion
                   - name: nodeorder
                   - name: binpack
                <strong>configurations:           # Add the following bold fields, which are Volcano configuration fields</strong>
                  <strong>- name: init-params</strong>
                    <strong>arguments: {"grace-over-time":"900","presetVirtualDevice":"true","nslb-version":"1.0","shared-tor-num":"2","useClusterInfoManager":"false","super-pod-size": "48","reserve-nodes": "2"}</strong>
            ...
            kind: Deployment
            apiVersion: apps/v1
            metadata:
              name: volcano-scheduler
              namespace: volcano-system
              labels:
                app: volcano-scheduler
            spec:
              ...
              template:
            ...
                    - name: volcano-scheduler
                      image: volcanosh/vc-scheduler:v1.7.0
                      args:
                        - --logtostderr
                        - --scheduler-conf=/volcano.scheduler/volcano-scheduler.conf
                        - --enable-healthz=true
                        - --enable-metrics=true
                        <strong>- --plugins-dir=plugins       # Load custom plugins in the volcano-scheduler startup command</strong>
                        - -v=3
                        - 2>&1

            # Source: volcano/templates/scheduler.yaml
            kind: ClusterRole
            apiVersion: rbac.authorization.k8s.io/v1
            metadata:
              name: volcano-scheduler
            rules:
            ...
              - apiGroups: ["nodeinfo.volcano.sh"]
                resources: ["numatopologies"]
                verbs: ["get", "list", "watch", "delete"]
              <strong>- apiGroups: [""]                          # Add get permission for services</strong>
                <strong>resources: ["services"]</strong>
                <strong>verbs: ["get"]</strong>
              - apiGroups: [""]
                resources: ["configmaps"]
                verbs: ["get", "create", "delete", "update",<strong>"list","watch"</strong>]    # Add list and watch permissions for ConfigMap
              - apiGroups: ["apps"]
                resources: ["daemonsets", "replicasets", "statefulsets"]
                verbs: ["list", "watch", "get"]
            ...</pre>

        4. Run the following command to start the volcano-scheduler component.

            ```shell
            kubectl apply -f installer/volcano-development.yaml
            ```

            Command output:

            ```ColdFusion
            namespace/volcano-system created
            namespace/volcano-monitoring created
            serviceaccount/volcano-admission created
            configmap/volcano-admission-configmap created
            clusterrole.rbac.authorization.k8s.io/volcano-admission created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-admission-role created
            service/volcano-admission-service created
            deployment.apps/volcano-admission created
            job.batch/volcano-admission-init created
            customresourcedefinition.apiextensions.k8s.io/jobs.batch.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/commands.bus.volcano.sh created
            serviceaccount/volcano-controllers created
            clusterrole.rbac.authorization.k8s.io/volcano-controllers created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-controllers-role created
            deployment.apps/volcano-controllers created
            serviceaccount/volcano-scheduler created
            configmap/volcano-scheduler-configmap created
            clusterrole.rbac.authorization.k8s.io/volcano-scheduler created
            clusterrolebinding.rbac.authorization.k8s.io/volcano-scheduler-role created
            service/volcano-scheduler-service created
            deployment.apps/volcano-scheduler created
            customresourcedefinition.apiextensions.k8s.io/podgroups.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/queues.scheduling.volcano.sh created
            customresourcedefinition.apiextensions.k8s.io/numatopologies.nodeinfo.volcano.sh created
            mutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-pods-mutate created
            mutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-queues-mutate created
            mutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-podgroups-mutate created
            mutatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-jobs-mutate created
            validatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-jobs-validate created
            validatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-pods-validate created
            validatingwebhookconfiguration.admissionregistration.k8s.io/volcano-admission-service-queues-validate created
            ```
