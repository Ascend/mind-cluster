# Volcano<a name="ZH-CN_TOPIC_0000002479226394"></a>

## 安装Volcano<a name="ZH-CN_TOPIC_0000002511426351"></a>

- 使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须在管理节点安装**调度器**，该调度器可以是Volcano或其他调度器。
- 若使用Volcano进行任务调度，则不建议通过Docker或Containerd指令创建/挂载NPU卡的容器，并在容器内跑任务。否则可能会触发Volcano调度问题。
- 仅使用容器化支持和资源监测的用户，可以不安装Volcano，请直接跳过本章节。

    本章为集群调度提供Volcano组件（vc-scheduler和vc-controller-manager）的安装指导。如需使用开源Volcano的其他组件，请用户自行安装，并保证其安全性。

    >[!NOTE] 
    >- 本文档中Volcano默认为集群调度组件提供的Volcano组件。其他基于开源Volcano的调度器可通过[（可选）集成昇腾插件扩展开源Volcano](#可选集成昇腾插件扩展开源volcano)章节，集成集群调度组件为开发者提供的Ascend-volcano-plugin插件，实现NPU调度相关功能。
    >- 6.0.RC1及以上版本NodeD与老版本Volcano不兼容，若使用6.0.RC1及以上版本的NodeD，需要配套使用6.0.RC1及以上版本的Volcano。
    >- 6.0.RC2及以上版本使用Volcano调度器时，默认必须安装ClusterD组件，若不安装ClusterD，则必须修改Volcano的启动参数，否则Volcano将无法正常调度任务。
- 不支持Volcano调度器和其他调度器管理相同的节点资源。

**操作步骤<a name="section57241227172819"></a>**

1. 以root用户登录K8s管理节点，并执行以下命令，查看Volcano镜像和版本号是否正确。

    ```shell
    docker images | grep volcanosh
    ```

    回显示例如下。

    ```ColdFusion
    volcanosh/vc-controller-manager      v1.7.0              84c73128cc55        3 days ago          44.5MB
    volcanosh/vc-scheduler               v1.7.0              e90c114c75b1        3 days ago          188MB
    ```

    - 是，执行[步骤2](#li823273914318)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作。

2. <a name="li823273914318"></a>将Volcano软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改对应启动YAML文件中Volcano的启动参数。常用启动参数请参见[表4](#table5305150122116)和[表5](#table203077022111)。
4. 配置Volcano日志转储。

    安装过程中，Volcano日志将挂载到磁盘空间（“/var/log/mindx-dl”）。默认情况下单日日志写入达到1.8G后，Volcano将清空日志文件。为防止空间被占满，请为Volcano配置日志转储，配置项信息参见[表1](#table1123141112311)，或选择更频繁的日志转储策略，避免日志丢失。

    1. 在管理节点“/etc/logrotate.d”目录下，执行以下命令，创建日志转储配置文件。

        ```shell
        vi /etc/logrotate.d/文件名
        ```

        例如：

        ```shell
        vi /etc/logrotate.d/volcano
        ```

        写入以下内容，然后执行<b>:wq</b>命令保存。

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

    2. 依次执行以下命令，设置配置文件权限为640，属主为root。

        ```shell
        chmod 640 /etc/logrotate.d/文件名
        chown root /etc/logrotate.d/文件名
        ```

        例如：

        ```shell
        chmod 640 /etc/logrotate.d/volcano
        chown root /etc/logrotate.d/volcano
        ```

    **表 1** Volcano日志转储文件配置项

    <a name="table1123141112311"></a>
    <table><thead align="left"><tr id="row412371119316"><th class="cellrowborder" valign="top" width="20.352035203520348%" id="mcps1.2.4.1.1"><p id="p12123811163113"><a name="p12123811163113"></a><a name="p12123811163113"></a>配置项</p>
    </th>
    <th class="cellrowborder" valign="top" width="33.82338233823382%" id="mcps1.2.4.1.2"><p id="p8123141118315"><a name="p8123141118315"></a><a name="p8123141118315"></a>说明</p>
    </th>
    <th class="cellrowborder" valign="top" width="45.82458245824582%" id="mcps1.2.4.1.3"><p id="p16123121111319"><a name="p16123121111319"></a><a name="p16123121111319"></a>可选值</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row612391119318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p1012481112315"><a name="p1012481112315"></a><a name="p1012481112315"></a>daily</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p16124101111317"><a name="p16124101111317"></a><a name="p16124101111317"></a>日志转储频率。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul1512431163118"></a><a name="ul1512431163118"></a><ul id="ul1512431163118"><li>daily：每日进行一次转储检查。</li><li>weekly：每周进行一次转储检查。</li><li>monthly：每月进行一次转储检查。</li><li>yearly：每年进行一次转储检查。</li></ul>
    </td>
    </tr>
    <tr id="row912511118314"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p111261711103117"><a name="p111261711103117"></a><a name="p111261711103117"></a>rotate <em id="i20126171193115"><a name="i20126171193115"></a><a name="i20126171193115"></a>x</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p13126191153117"><a name="p13126191153117"></a><a name="p13126191153117"></a>日志文件删除之前转储的次数。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p161261911153111"><a name="p161261911153111"></a><a name="p161261911153111"></a><em id="i161266119317"><a name="i161266119317"></a><a name="i161266119317"></a>x</em>为备份次数。</p>
    <p id="p19126101103110"><a name="p19126101103110"></a><a name="p19126101103110"></a>例如：</p>
    <a name="ul151261211103117"></a><a name="ul151261211103117"></a><ul id="ul151261211103117"><li>rotate 0：没有备份。</li><li>rotate 8：保留8次备份。</li></ul>
    </td>
    </tr>
    <tr id="row1912641115318"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p15127111153115"><a name="p15127111153115"></a><a name="p15127111153115"></a>size <em id="i912731113110"><a name="i912731113110"></a><a name="i912731113110"></a>xx</em></p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p412751115312"><a name="p412751115312"></a><a name="p412751115312"></a>日志文件到达指定的大小时才转储。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><p id="p131273112314"><a name="p131273112314"></a><a name="p131273112314"></a>size单位可以指定：</p>
    <a name="ul1012771118311"></a><a name="ul1012771118311"></a><ul id="ul1012771118311"><li>byte（缺省）</li><li>K</li><li>M</li></ul>
    <p id="p1112761173115"><a name="p1112761173115"></a><a name="p1112761173115"></a>例如size 50M指日志文件达到50 MB时转储。</p>
    <div class="note" id="note191277111311"><a name="note191277111311"></a><a name="note191277111311"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p17127191193120"><a name="p17127191193120"></a><a name="p17127191193120"></a>logrotate会根据配置的转储频率，定期检查日志文件大小，检查时大小超过size的文件才会触发转储。</p>
    <p id="p112771153111"><a name="p112771153111"></a><a name="p112771153111"></a>这意味着，logrotate并不会在日志文件达到大小限制时立刻将其转储。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row4127111173111"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p161282115316"><a name="p161282115316"></a><a name="p161282115316"></a>compress</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p112851110317"><a name="p112851110317"></a><a name="p112851110317"></a>是否通过gzip压缩转储以后的日志。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul712814118312"></a><a name="ul712814118312"></a><ul id="ul712814118312"><li>compress：使用gzip压缩。</li><li>nocompress：不使用gzip压缩。</li></ul>
    </td>
    </tr>
    <tr id="row18128511203117"><td class="cellrowborder" valign="top" width="20.352035203520348%" headers="mcps1.2.4.1.1 "><p id="p412801153117"><a name="p412801153117"></a><a name="p412801153117"></a>notifempty</p>
    </td>
    <td class="cellrowborder" valign="top" width="33.82338233823382%" headers="mcps1.2.4.1.2 "><p id="p9128141123113"><a name="p9128141123113"></a><a name="p9128141123113"></a>空文件是否转储。</p>
    </td>
    <td class="cellrowborder" valign="top" width="45.82458245824582%" headers="mcps1.2.4.1.3 "><a name="ul31281611163118"></a><a name="ul31281611163118"></a><ul id="ul31281611163118"><li>ifempty：空文件也转储。</li><li>notifempty：空文件不触发转储。</li></ul>
    </td>
    </tr>
    </tbody>
    </table>

5. （可选）在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano所需的CPU和内存。CPU和内存推荐值可以参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)中volcano-controller和volcano-scheduler表格的建议值。

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

6. （可选）调度时间性能调优。支持在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano所使用的插件。请参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)中Volcano高级配置参数说明和支持的Plugins列表的表格说明进行操作。

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
          - name: volcano-npu_v26.0.0_linux-aarch64   # 其中v26.0.0为MindCluster的版本号，根据不同版本，该处取值不同
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

7. （可选）在“volcano-_v\{version\}_.yaml”中，配置开启Volcano健康检查接口和Prometheus信息收集接口。

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
                  <strong>--enable-healthz=true   # 为保证可正常访问Volcano健康检查端口，本参数取值需为"true"
                  --enable-metrics=true   # 为保证可正常访问Prometheus信息收集端口，本参数取值需为"true"</strong>
                  ...
    ...</pre>

    **表 2** 集群调度Volcano组件开放接口说明

    <a name="zh-cn_topic_0000001446965056_table173071368477"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001446965056_row153077618473"><th class="cellrowborder" valign="top" width="34.68346834683469%" id="mcps1.2.6.1.1"><p id="zh-cn_topic_0000001446965056_p3307116134715"><a name="zh-cn_topic_0000001446965056_p3307116134715"></a><a name="zh-cn_topic_0000001446965056_p3307116134715"></a>访问方式</p>
    </th>
    <th class="cellrowborder" valign="top" width="6.0906090609060906%" id="mcps1.2.6.1.2"><p id="zh-cn_topic_0000001446965056_p1525244211493"><a name="zh-cn_topic_0000001446965056_p1525244211493"></a><a name="zh-cn_topic_0000001446965056_p1525244211493"></a>协议</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.741174117411742%" id="mcps1.2.6.1.3"><p id="zh-cn_topic_0000001446965056_p04543391867"><a name="zh-cn_topic_0000001446965056_p04543391867"></a><a name="zh-cn_topic_0000001446965056_p04543391867"></a>方法</p>
    </th>
    <th class="cellrowborder" valign="top" width="16.89168916891689%" id="mcps1.2.6.1.4"><p id="zh-cn_topic_0000001446965056_p23071468473"><a name="zh-cn_topic_0000001446965056_p23071468473"></a><a name="zh-cn_topic_0000001446965056_p23071468473"></a>作用</p>
    </th>
    <th class="cellrowborder" valign="top" width="30.59305930593059%" id="mcps1.2.6.1.5"><p id="zh-cn_topic_0000001446965056_p730796134713"><a name="zh-cn_topic_0000001446965056_p730796134713"></a><a name="zh-cn_topic_0000001446965056_p730796134713"></a>所属组件</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001446965056_row23070613479"><td class="cellrowborder" valign="top" width="34.68346834683469%" headers="mcps1.2.6.1.1 "><p id="zh-cn_topic_0000001446965056_p1730717615477"><a name="zh-cn_topic_0000001446965056_p1730717615477"></a><a name="zh-cn_topic_0000001446965056_p1730717615477"></a>http://podIP:11251/healthz</p>
    </td>
    <td class="cellrowborder" valign="top" width="6.0906090609060906%" headers="mcps1.2.6.1.2 "><p id="zh-cn_topic_0000001446965056_p10252142154917"><a name="zh-cn_topic_0000001446965056_p10252142154917"></a><a name="zh-cn_topic_0000001446965056_p10252142154917"></a>http</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.741174117411742%" headers="mcps1.2.6.1.3 "><p id="zh-cn_topic_0000001446965056_p64546391612"><a name="zh-cn_topic_0000001446965056_p64546391612"></a><a name="zh-cn_topic_0000001446965056_p64546391612"></a>Get</p>
    </td>
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p151727316530"><a name="zh-cn_topic_0000001446965056_p151727316530"></a><a name="zh-cn_topic_0000001446965056_p151727316530"></a>健康检查端口</p>
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
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p53084617475"><a name="zh-cn_topic_0000001446965056_p53084617475"></a><a name="zh-cn_topic_0000001446965056_p53084617475"></a>健康检查端口</p>
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
    <td class="cellrowborder" valign="top" width="16.89168916891689%" headers="mcps1.2.6.1.4 "><p id="zh-cn_topic_0000001446965056_p193087624718"><a name="zh-cn_topic_0000001446965056_p193087624718"></a><a name="zh-cn_topic_0000001446965056_p193087624718"></a>Prometheus信息收集端口</p>
    </td>
    <td class="cellrowborder" valign="top" width="30.59305930593059%" headers="mcps1.2.6.1.5 "><p id="zh-cn_topic_0000001446965056_p3308166154716"><a name="zh-cn_topic_0000001446965056_p3308166154716"></a><a name="zh-cn_topic_0000001446965056_p3308166154716"></a>volcano-scheduler</p>
    </td>
    </tr>
    </tbody>
    </table>

8. （可选）在“volcano-v<i>\{version\}</i>.yaml“中，配置Volcano使用的集群调度组件为用户提供的重调度时删除Pod的模式、虚拟化方式、交换机亲和性调度、是否自维护可用芯片状态等。

    <pre codetype="yaml">
    ...
    data:
      volcano-scheduler.conf: |
    ...
        configurations:
          - name: init-params
            arguments: {<strong>"grace-over-time":"900","presetVirtualDevice":"true","nslb-version":"1.0","shared-tor-num":"2","useClusterInfoManager":"false","self-maintain-available-card":"true","super-pod-size": "48","reserve-nodes": "2","forceEnqueue":"true"</strong>}
    ...</pre>

    **表 3**  参数说明

    <a name="table208981646194315"></a>
    <table><thead align="left"><tr id="row08991746174316"><th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.1"><p id="p132621494445"><a name="p132621494445"></a><a name="p132621494445"></a>参数名称</p>
    </th>
    <th class="cellrowborder" valign="top" width="20%" id="mcps1.2.4.1.2"><p id="p194862061467"><a name="p194862061467"></a><a name="p194862061467"></a>默认值</p>
    </th>
    <th class="cellrowborder" valign="top" width="60%" id="mcps1.2.4.1.3"><p id="p18991846144317"><a name="p18991846144317"></a><a name="p18991846144317"></a>参数说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="row1788817373541"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1888143725417"><a name="p1888143725417"></a><a name="p1888143725417"></a>grace-over-time</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p7888103725412"><a name="p7888103725412"></a><a name="p7888103725412"></a>900</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p145262285146"><a name="p145262285146"></a><a name="p145262285146"></a>重调度优雅删除模式下删除Pod所需最大时间，单位为秒，取值范围2~3600。配置该字段表示使用重调度的优雅删除模式。优雅删除是指在重调度过程中，会等待<span id="ph8305245165813"><a name="ph8305245165813"></a><a name="ph8305245165813"></a>Volcano</span>执行相关善后工作，900秒后若Pod还未删除成功，再直接强制删除Pod，不做善后。</p>
    </td>
    </tr>
    <tr id="row95211735125411"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p1352203555411"><a name="p1352203555411"></a><a name="p1352203555411"></a>presetVirtualDevice</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p552293515412"><a name="p552293515412"></a><a name="p552293515412"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p135221235105419"><a name="p135221235105419"></a><a name="p135221235105419"></a>采用的虚拟化方式。</p>
    <a name="ul206451443111219"></a><a name="ul206451443111219"></a><ul id="ul206451443111219"><li>true：静态虚拟化</li><li>false：动态虚拟化</li></ul>
    </td>
    </tr>
    <tr id="row1589974674320"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p6899114619435"><a name="p6899114619435"></a><a name="p6899114619435"></a>nslb-version</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6484146114619"><a name="p6484146114619"></a><a name="p6484146114619"></a>1.0</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p7830165414514"><a name="p7830165414514"></a><a name="p7830165414514"></a>交换机亲和性调度的版本，可以取值为1.0和2.0。</p>
    <div class="note" id="note882315541054"><a name="note882315541054"></a><a name="note882315541054"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul59831535122714"></a><a name="ul59831535122714"></a><ul id="ul59831535122714"><li>交换机亲和性调度1.0版本支持<span id="ph1157665817140"><a name="ph1157665817140"></a><a name="ph1157665817140"></a>Atlas 训练系列产品</span>和<span id="ph168598363399"><a name="ph168598363399"></a><a name="ph168598363399"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph4181625925"><a name="ph4181625925"></a><a name="ph4181625925"></a>PyTorch</span>和<span id="ph61882510210"><a name="ph61882510210"></a><a name="ph61882510210"></a>MindSpore</span>。</li><li>交换机亲和性调度2.0版本支持<span id="ph311717506401"><a name="ph311717506401"></a><a name="ph311717506401"></a><term id="zh-cn_topic_0000001519959665_term57208119917_1"><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a><a name="zh-cn_topic_0000001519959665_term57208119917_1"></a>Atlas A2 训练系列产品</term></span>；支持<span id="ph619244413568"><a name="ph619244413568"></a><a name="ph619244413568"></a>PyTorch</span>框架。</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row8899946174318"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p168998463434"><a name="p168998463434"></a><a name="p168998463434"></a>shared-tor-num</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p989910464439"><a name="p989910464439"></a><a name="p989910464439"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p925113215214"><a name="p925113215214"></a><a name="p925113215214"></a>交换机亲和性调度2.0中单个任务可使用的最大共享交换机数量，可取值为1或2。仅在nslb-version取值为2.0时生效。</p>
    <p id="p1856962434719"><a name="p1856962434719"></a><a name="p1856962434719"></a>交换机亲和性调度（1.0或2.0）说明可以参见<a href="../../../usage/basic_scheduling/01_affinity_scheduling/04_node_based_affinity.md">基于节点的亲和性</a>章节。</p>
    </td>
    </tr>
    <tr id="row797916276295"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p9621013114312"><a name="p9621013114312"></a><a name="p9621013114312"></a>useClusterInfoManager</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p1024875418187"><a name="p1024875418187"></a><a name="p1024875418187"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1797510751015"><a name="p1797510751015"></a><a name="p1797510751015"></a><span id="ph18393155819297"><a name="ph18393155819297"></a><a name="ph18393155819297"></a>Volcano</span>获取集群信息的方式。取值说明如下：</p>
    <a name="ul675021361014"></a><a name="ul675021361014"></a><ul id="ul675021361014"><li>true：读取<span id="ph16899408574"><a name="ph16899408574"></a><a name="ph16899408574"></a>ClusterD</span>上报的<span id="ph1921415457302"><a name="ph1921415457302"></a><a name="ph1921415457302"></a>ConfigMap</span>。</li><li>false：分别读取<span id="ph19274234236"><a name="ph19274234236"></a><a name="ph19274234236"></a>Ascend Device Plugin</span>和<span id="ph144095321390"><a name="ph144095321390"></a><a name="ph144095321390"></a>NodeD</span>上报的<span id="ph039324431114"><a name="ph039324431114"></a><a name="ph039324431114"></a>ConfigMap</span>。</li></ul>
    <div class="note" id="note1466414341216"><a name="note1466414341216"></a><a name="note1466414341216"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p1166463181214"><a name="p1166463181214"></a><a name="p1166463181214"></a>默认使用读取<span id="ph19101421151220"><a name="ph19101421151220"></a><a name="ph19101421151220"></a>ClusterD</span>组件上报的<span id="ph139579361121"><a name="ph139579361121"></a><a name="ph139579361121"></a>ConfigMap</span>。后续版本将不支持读取<span id="ph3588183951516"><a name="ph3588183951516"></a><a name="ph3588183951516"></a>Ascend Device Plugin</span>和<span id="ph1758893981514"><a name="ph1758893981514"></a><a name="ph1758893981514"></a>NodeD</span>上报的<span id="ph75887392157"><a name="ph75887392157"></a><a name="ph75887392157"></a>ConfigMap</span>。</p>
    </div></div>
    </td>
    </tr>
    <tr id="row1913114164518"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p91454144514"><a name="p91454144514"></a><a name="p91454144514"></a>self-maintain-available-card</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p814241174517"><a name="p814241174517"></a><a name="p814241174517"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p151414164516"><a name="p151414164516"></a><a name="p151414164516"></a>Volcano是否自维护可用芯片状态。取值说明如下：</p>
    <a name="ul299044019472"></a><a name="ul299044019472"></a><ul id="ul299044019472"><li>true：Volcano自维护可用芯片状态。</li><li>false：Volcano根据ClusterD或<span id="ph98552414486"><a name="ph98552414486"></a><a name="ph98552414486"></a>Ascend Device Plugin</span>上报的<span id="ph1185824104819"><a name="ph1185824104819"></a><a name="ph1185824104819"></a>ConfigMap</span>获取可用芯片状态。</li></ul>
    </td>
    </tr>
    <tr id="row4612538250"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p26130381510"><a name="p26130381510"></a><a name="p26130381510"></a>super-pod-size</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p6613173818516"><a name="p6613173818516"></a><a name="p6613173818516"></a>48</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p461323812519"><a name="p461323812519"></a><a name="p461323812519"></a><span id="ph128111331314"><a name="ph128111331314"></a><a name="ph128111331314"></a>Atlas 900 A3 SuperPoD 超节点</span>中一个超节点的节点数量。</p>
    </td>
    </tr>
    <tr id="row9561856657"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p956215565514"><a name="p956215565514"></a><a name="p956215565514"></a>reserve-nodes</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p125626567516"><a name="p125626567516"></a><a name="p125626567516"></a>2</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p25637568515"><a name="p25637568515"></a><a name="p25637568515"></a><span id="ph915032251212"><a name="ph915032251212"></a><a name="ph915032251212"></a>Atlas 900 A3 SuperPoD 超节点</span>中一个超节点中预留节点数量。</p>
    <div class="note" id="note1514175285210"><a name="note1514175285210"></a><a name="note1514175285210"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p96481321115510"><a name="p96481321115510"></a><a name="p96481321115510"></a>若设置的reserve-nodes大于super-pod-size时，存在以下场景。</p>
    <a name="ul13842528165510"></a><a name="ul13842528165510"></a><ul id="ul13842528165510"><li>super-pod-size大于2，则默认重置reserve-nodes取值为2</li><li>super-pod-size小于或等于2，则默认重置reserve-nodes取值为0。</li></ul>
    </div></div>
    </td>
    </tr>
    <tr id="row1890722719501"><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p id="p590882716507"><a name="p590882716507"></a><a name="p590882716507"></a><span id="ph19180940145012"><a name="ph19180940145012"></a><a name="ph19180940145012"></a>forceEnqueue</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p id="p17908162765017"><a name="p17908162765017"></a><a name="p17908162765017"></a><span id="ph16315161885115"><a name="ph16315161885115"></a><a name="ph16315161885115"></a>true</span></p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p id="p1790852711505"><a name="p1790852711505"></a><a name="p1790852711505"></a><span id="ph188121729115116"><a name="ph188121729115116"></a><a name="ph188121729115116"></a>任务在集群NPU资源满足的情况下是否强制</span><span id="ph280124455118"><a name="ph280124455118"></a><a name="ph280124455118"></a>进入待调度队列</span><span id="ph1179814511514"><a name="ph1179814511514"></a><a name="ph1179814511514"></a>。</span><span id="ph2278145215114"><a name="ph2278145215114"></a><a name="ph2278145215114"></a>取值说明如下：</span></p>
    <a name="ul12820554135117"></a><a name="ul12820554135117"></a><ul id="ul12820554135117"><li>true：Volcano开启<span id="ph11766123385220"><a name="ph11766123385220"></a><a name="ph11766123385220"></a>Enqueue</span>这个action时，若集群NPU资源满足当前任务，则任务会<span id="ph3349644125319"><a name="ph3349644125319"></a><a name="ph3349644125319"></a>强制</span><span id="ph22191237135316"><a name="ph22191237135316"></a><a name="ph22191237135316"></a>进入待调度队列</span>，不会关心其他资源是否充足。如果当前任务长时间在待调度队列中，会预占用资源，从而可能导致其他任务无法入队。</li><li>其他值：当集群NPU资源不足时，拒绝任务<span id="ph6205121155415"><a name="ph6205121155415"></a><a name="ph6205121155415"></a>进入待调度队列。若</span>NPU资源满足当前任务，则由所有插件共同决定是否<span id="ph370210413554"><a name="ph370210413554"></a><a name="ph370210413554"></a>进入待调度队列</span>。</li></ul>
    <p id="p12691948115614"><a name="p12691948115614"></a><a name="p12691948115614"></a>关于该参数的详细说明请参见<a href="https://volcano.sh/en/docs/v1-12-0/actions/" target="_blank" rel="noopener noreferrer">Volcano Actions</a>。</p>
    </td>
    </tr>
    <tr><td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.1 "><p>resource-level-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="20%" headers="mcps1.2.4.1.2 "><p>默认值为空</p>
    </td>
    <td class="cellrowborder" valign="top" width="60%" headers="mcps1.2.4.1.3 "><p>集群中节点对应的网络资源层级的Json配置。</p><p>该参数仅在多级调度任务中使用，非多级调度任务无需配置该参数。</p><p>取值说明如下：</p><ul><li>Json文件第一层key为网络拓扑树的名称，对应的value为该网络拓扑树的详细定义。</li><li>在网络拓扑树的详细定义中，key用于标识具体的网络层级，取值为前缀level+网络层级序号n，其中n为大于等于1的正整数；value为对应的网络层级定义结构体。</li><li>在网络层级定义结构体中，存在如下字段：<ul><li>label：标识节点在该网络层级的节点标签的key，取值为字符串。</li><li>reservedNode：标识预留的子层级的节点数量，取值为整数，仅在level1层级的配置中生效。多级调度任务调度时会优先尝试扣除预留数量的节点进行调度，扣除预留节点不能执行调度时会正常使用预留节点资源。</li></ul></li></ul><p>关于该参数的详细说明和样例请参见<a href="../../../usage/basic_scheduling/05_multi_level_scheduling.md#配置volcano启动参数">配置Volcano启动参数</a>。</p>
    </td>
    </tr>
    </tbody>
    </table>

    >[!NOTE]
    >- 更多关于开源Volcano的配置，可以参见[开源Volcano官方文档](https://support.huaweicloud.com/usermanual-cce/cce_10_0193.html)进行操作。
    >- K8s支持使用nodeAffinity字段进行节点亲和性调度，该字段的详细说明请参见[Kubernetes官方文档](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/assign-pods-nodes-using-node-affinity/)；Volcano也支持使用该字段，操作指导请参见[调度配置](../../../common_operations.md#调度配置)章节。

9. （可选）调度时间性能调优。支持Volcano将单任务（训练的vcjob或acjob任务）的4000或5000个Pod调度到4000或5000个节点上的调度时间优化到5分钟左右，若用户想要使用该调度性能，需要在“volcano-v<i>\{version\}</i>.yaml”上做如下修改。
 
    - 若要达到5分钟左右的参考时间，需要保证CPU的频率至少为2.60GHz，APIServer时延不超过80毫秒。
    - 如果不使用K8s原生的nodeAffinity和podAntiAffinity进行调度，可以关闭nodeorder插件，进一步减少调度时间。

    <pre codetype="yaml">
    data:
      volcano-scheduler.conf: |
    
    ...
          - name: proportion
            enableNodeOrder: false
          - name: nodeorder
            <strong>enableNodeOrder: false     # 可选，不使用nodeAffinity和podAntiAffinity调度时，可关闭nodeorder插件</strong>
    ...
          containers:
            - name: volcano-scheduler
              image: volcanosh/vc-scheduler:v1.7.0
              command: ["/bin/ash"]
              args: ["-c", "umask 027; <strong>GOMEMLIMIT=15000000000 GOGC=off</strong> /vc-scheduler      <strong># 新增GOMEMLIMIT=15000000000和GOGC=off字段</strong>
                      --scheduler-conf=/volcano.scheduler/volcano-scheduler.conf
                      --plugins-dir=plugins
                      --logtostderr=false
                      --log_dir=/var/log/mindx-dl/volcano-scheduler
                      --log_file=/var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log
                      -v=2 2>&1"]
              imagePullPolicy: "IfNotPresent"
              resources:
                requests:
                  <strong>memory: 10000Mi                                                                # 将4Gi修改为10000Mi</strong>
                  cpu: 5500m
                limits:
                  <strong>memory: 15000Mi                                                       # 将8Gi修改为15000Mi</strong>
                  cpu: 5500m
    ...</pre>

10. 在管理节点的YAML所在路径，执行以下命令，启动Volcano。

    ```shell
    kubectl apply -f volcano-v{version}.yaml
    ```

    启动示例如下：

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

11. 执行以下命令，查看组件状态。

    ```shell
    kubectl get pod -n volcano-system
    ```

    回显示例如下，出现**Running**表示组件启动成功：

    ```ColdFusion
    NAME                                          READY    STATUS     RESTARTS     AGE
    volcano-controllers-5cf8d788d5-qdpzq   1/1     Running   0          1m
    volcano-scheduler-6cffd555c9-45k7c     1/1     Running   0          1m
    ```

    >[!NOTE] 
    >- 若Volcano的Pod状态为CrashLoopBackOff，可以参见[手动安装Volcano后，Pod状态为：CrashLoopBackOff](../../../faq.md#手动安装volcano后pod状态为crashloopbackoff)章节进行处理。
    >- 若volcano-scheduler-6cffd555c9-45k7c状态为Running，但是调度异常，可以参见[Volcano组件工作异常，日志出现Failed to get plugin](../../../faq.md#volcano组件工作异常日志出现failed-to-get-plugin)章节进行处理。
    >- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
    >- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
    >- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
    >- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

**参数说明<a name="section1317934882010"></a>**

**表 4**  volcano-scheduler启动参数

<a name="table5305150122116"></a>
<table><thead align="left"><tr id="row63052016218"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p133052042113"><a name="p133052042113"></a><a name="p133052042113"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15.310000000000002%" id="mcps1.2.5.1.2"><p id="p330560162111"><a name="p330560162111"></a><a name="p330560162111"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="29.69%" id="mcps1.2.5.1.3"><p id="p3305600215"><a name="p3305600215"></a><a name="p3305600215"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p63067062115"><a name="p63067062115"></a><a name="p63067062115"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row12306160112118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p14306100102116"><a name="p14306100102116"></a><a name="p14306100102116"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p103066014211"><a name="p103066014211"></a><a name="p103066014211"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p13306150192120"><a name="p13306150192120"></a><a name="p13306150192120"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2030616002117"><a name="p2030616002117"></a><a name="p2030616002117"></a>日志目录，组件启动YAML中默认值为/var/log/mindx-dl/volcano-scheduler。</p>
</td>
</tr>
<tr id="row230620102115"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p143064002118"><a name="p143064002118"></a><a name="p143064002118"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p173067012119"><a name="p173067012119"></a><a name="p173067012119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p430620132116"><a name="p430620132116"></a><a name="p430620132116"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p930615020218"><a name="p930615020218"></a><a name="p930615020218"></a>日志文件名称，组件启动YAML中默认值为/var/log/mindx-dl/volcano-scheduler/volcano-scheduler.log。</p>
<div class="note" id="note19596191219291"><a name="note19596191219291"></a><a name="note19596191219291"></a><div class="notebody"><p id="p10596012112919"><a name="p10596012112919"></a><a name="p10596012112919"></a>转储后文件的命名格式为：volcano-scheduler.log-触发转储的时间.gz，如：volcano-scheduler.log-20230926.gz。</p>
</div></div>
</td>
</tr>
<tr id="row17922126205817"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p139228267582"><a name="p139228267582"></a><a name="p139228267582"></a>--scheduler-conf</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p17490165855810"><a name="p17490165855810"></a><a name="p17490165855810"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p192312261580"><a name="p192312261580"></a><a name="p192312261580"></a>/volcano.scheduler/volcano-scheduler.conf</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1192372695818"><a name="p1192372695818"></a><a name="p1192372695818"></a>调度组件配置文件的绝对路径。</p>
</td>
</tr>
<tr id="row630618042113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p173061701214"><a name="p173061701214"></a><a name="p173061701214"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p730613011217"><a name="p730613011217"></a><a name="p730613011217"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p0306170112117"><a name="p0306170112117"></a><a name="p0306170112117"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p430613012117"><a name="p430613012117"></a><a name="p430613012117"></a>日志是否打印到标准输出。</p>
<a name="ul582374031615"></a><a name="ul582374031615"></a><ul id="ul582374031615"><li>true：打印。</li><li>false：不打印。</li></ul>
</td>
</tr>
<tr id="row53063062118"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p330618022113"><a name="p330618022113"></a><a name="p330618022113"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p133061010218"><a name="p133061010218"></a><a name="p133061010218"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p23068042118"><a name="p23068042118"></a><a name="p23068042118"></a>2</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p43067042115"><a name="p43067042115"></a><a name="p43067042115"></a>日志输出级别：</p>
<a name="ul03064012212"></a><a name="ul03064012212"></a><ul id="ul03064012212"><li>取值为1：error</li><li>取值为2：warning</li><li>取值为3：info</li><li>取值为4：debug</li></ul>
</td>
</tr>
<tr id="row11306140152113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p730614015211"><a name="p730614015211"></a><a name="p730614015211"></a>--plugins-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1130614013214"><a name="p1130614013214"></a><a name="p1130614013214"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p12307200192115"><a name="p12307200192115"></a><a name="p12307200192115"></a>plugins</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p173071603217"><a name="p173071603217"></a><a name="p173071603217"></a>scheduler插件加载路径。</p>
</td>
</tr>
<tr id="row113072012113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p9307140142120"><a name="p9307140142120"></a><a name="p9307140142120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p03072016212"><a name="p03072016212"></a><a name="p03072016212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p330712011215"><a name="p330712011215"></a><a name="p330712011215"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p53071209215"><a name="p53071209215"></a><a name="p53071209215"></a>是否查询volcano-scheduler二进制版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row62114943417"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p182349173416"><a name="p182349173416"></a><a name="p182349173416"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1221849193415"><a name="p1221849193415"></a><a name="p1221849193415"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p1021949203420"><a name="p1021949203420"></a><a name="p1021949203420"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p1321749193419"><a name="p1321749193419"></a><a name="p1321749193419"></a>日志文件最大存储大小（单位为M）。</p>
<div class="note" id="note1919311416364"><a name="note1919311416364"></a><a name="note1919311416364"></a><div class="notebody"><p id="p7193444361"><a name="p7193444361"></a><a name="p7193444361"></a>当日志文件大小超过阈值时，日志内容会被清空。</p>
</div></div>
</td>
</tr>
<tr id="row159867311462"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10986173174613"><a name="p10986173174613"></a><a name="p10986173174613"></a>--leader-elect</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p1098619374617"><a name="p1098619374617"></a><a name="p1098619374617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p19866311462"><a name="p19866311462"></a><a name="p19866311462"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p4986143184611"><a name="p4986143184611"></a><a name="p4986143184611"></a>多副本启动时启动选主模式。</p>
</td>
</tr>
<tr id="row1253065634617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1453015644610"><a name="p1453015644610"></a><a name="p1453015644610"></a>--percentage-nodes-to-find</p>
</td>
<td class="cellrowborder" valign="top" width="15.310000000000002%" headers="mcps1.2.5.1.2 "><p id="p145301156194612"><a name="p145301156194612"></a><a name="p145301156194612"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="29.69%" headers="mcps1.2.5.1.3 "><p id="p16530175615462"><a name="p16530175615462"></a><a name="p16530175615462"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p11530165644617"><a name="p11530165644617"></a><a name="p11530165644617"></a>任务调度时选取可用节点占集群总节点的百分比。</p>
</td>
</tr>
</tbody>
</table>

**表 5**  volcano-controller启动参数

<a name="table203077022111"></a>
<table><thead align="left"><tr id="row18307705217"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p193071001218"><a name="p193071001218"></a><a name="p193071001218"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p13307208218"><a name="p13307208218"></a><a name="p13307208218"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.3"><p id="p123078062120"><a name="p123078062120"></a><a name="p123078062120"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.4"><p id="p4307100172120"><a name="p4307100172120"></a><a name="p4307100172120"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row173077014210"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p43078015211"><a name="p43078015211"></a><a name="p43078015211"></a>--log-dir</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p173071104213"><a name="p173071104213"></a><a name="p173071104213"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p113071302218"><a name="p113071302218"></a><a name="p113071302218"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330718019213"><a name="p330718019213"></a><a name="p330718019213"></a>日志目录，组件启动YAML中默认值为/var/log/mindx-dl/volcano-controller。</p>
</td>
</tr>
<tr id="row1307170112113"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p17307130112117"><a name="p17307130112117"></a><a name="p17307130112117"></a>--log-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1630780182118"><a name="p1630780182118"></a><a name="p1630780182118"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p1930714062115"><a name="p1930714062115"></a><a name="p1930714062115"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p143077018217"><a name="p143077018217"></a><a name="p143077018217"></a>日志文件名称，组件启动YAML中默认值为/var/log/mindx-dl/volcano-controller/volcano-controller.log。</p>
<div class="note" id="note215144410296"><a name="note215144410296"></a><a name="note215144410296"></a><div class="notebody"><p id="p715144132910"><a name="p715144132910"></a><a name="p715144132910"></a>转储后文件的命名格式为：volcano-controller.log-触发转储的时间.gz，如：volcano-controller.log-20230926.gz。</p>
</div></div>
</td>
</tr>
<tr id="row730760202120"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p93071805219"><a name="p93071805219"></a><a name="p93071805219"></a>--logtostderr</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p730812011211"><a name="p730812011211"></a><a name="p730812011211"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p2308140142118"><a name="p2308140142118"></a><a name="p2308140142118"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p2308170172116"><a name="p2308170172116"></a><a name="p2308170172116"></a>日志是否打印到标准输出。</p>
<a name="ul142362048125710"></a><a name="ul142362048125710"></a><ul id="ul142362048125710"><li>true：打印。</li><li>false：不打印。</li></ul>
</td>
</tr>
<tr id="row930819012214"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p193088092115"><a name="p193088092115"></a><a name="p193088092115"></a>-v</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p1930812016213"><a name="p1930812016213"></a><a name="p1930812016213"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p123081003218"><a name="p123081003218"></a><a name="p123081003218"></a>4</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p330830162118"><a name="p330830162118"></a><a name="p330830162118"></a>日志输出级别：</p>
<a name="ul6308150112119"></a><a name="ul6308150112119"></a><ul id="ul6308150112119"><li>1：error</li><li>2：warning</li><li>3：info</li><li>4：debug</li></ul>
</td>
</tr>
<tr id="row1330813015217"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p133085052120"><a name="p133085052120"></a><a name="p133085052120"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p030814011212"><a name="p030814011212"></a><a name="p030814011212"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p10308140122115"><a name="p10308140122115"></a><a name="p10308140122115"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p130818011219"><a name="p130818011219"></a><a name="p130818011219"></a>volcano-controller二进制版本号。</p>
</td>
</tr>
<tr id="row926534763719"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1413064912376"><a name="p1413064912376"></a><a name="p1413064912376"></a>--log_file_max_size</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p313074910373"><a name="p313074910373"></a><a name="p313074910373"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.3 "><p id="p31301349183714"><a name="p31301349183714"></a><a name="p31301349183714"></a>1800</p>
</td>
<td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.4 "><p id="p111301349113715"><a name="p111301349113715"></a><a name="p111301349113715"></a>日志文件最大存储大小（单位为M）。</p>
<div class="note" id="note1513064943719"><a name="note1513064943719"></a><a name="note1513064943719"></a><div class="notebody"><p id="p111317492373"><a name="p111317492373"></a><a name="p111317492373"></a>当日志文件大小超过阈值时，日志内容会被清空。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 
>Volcano为开源软件，启动参数只罗列目前使用的常见参数，其他详细的参数请参见开源软件的说明。

## （可选）使用Volcano交换机亲和性调度<a name="ZH-CN_TOPIC_0000002479226480"></a>

Volcano组件支持交换机的亲和性调度。使用该功能需要上传交换机与服务器节点的对应关系以供Volcano使用，操作步骤如下。

>[!NOTE] 
>当前只支持训练和推理任务进行整卡的交换机亲和性调度，不支持静态或动态vNPU调度。

**操作步骤<a name="section7172163412209"></a>**

1. <a name="li6319161364017"></a>准备部署环境的网络设计LLD文档，将其上传到K8s管理节点的任意目录（以“/home/tor-affinity”为例）。

    >[!NOTE] 
    >LLD文件名需要是lld.xlsx。

2. 获取LLD文档解析脚本。

    进入[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)仓库，根据[mindcluster-deploy开源仓版本说明](../../../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支。下载“samples/utils”目录中的lld\_to\_cm.py文件，将该文件上传到管理节点[步骤1](#li6319161364017)中的目录下。

3. 执行以下命令，启动“lld\_to\_cm.py”脚本。

    ```shell
    python ./lld_to_cm.py --num 32
    ```

    >[!NOTE] 
    >- 使用--num（或-n）子命令指定一个交换机下的节点个数，不指定该参数时默认取值为4。
    >- 使用--level（或-l）子命令指定交换机组网类型，不指定该参数时默认取值为double\_layer，取值说明如下。
    >    - single\_layer：使用单层交换机组网。
    >    - double\_layer：使用双层交换机组网。
    >- 该脚本需要使用到openpyxl模块，如果安装环境缺少该模块，可以使用**pip install openpyxl**命令进行安装。

4. 执行以下命令，检查ConfigMap是否创建成功。

    ```shell
    kubectl get cm -n kube-system basic-tor-node-cm
    ```

    回显示例如下，表示创建成功。

    ```ColdFusion
    NAME                DATA   AGE
    basic-tor-node-cm   1      8s
    ```

**配置交换机亲和性调度<a name="section125904488511"></a>**

配置交换机的亲和性调度需要在任务YAML中配置tor-affinity参数，tor-affinity的位置和配置说明如下表所示。

**表 1**  YAML参数说明

<a name="table325141716575"></a>

|参数|取值|说明|
|--|--|--|
|(.kind=="AscendJob").metadata.labels.tor-affinity|<ul><li>large-model-schema：大模型任务或填充任务</li><li>normal-schema：普通任务</li><li>null：不使用交换机亲和性调度<div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>用户需要根据任务副本数，选择任务类型。任务副本数小于4为填充任务。任务副本数大于或等于4为大模型任务。普通任务不限制任务副本数。</p></div></div></li></ul>|<p>默认值为null，表示不使用交换机亲和性调度。用户需要根据任务类型进行配置。</p><ul><li>交换机亲和性调度1.0版本支持Atlas 训练系列产品和<term>Atlas A2 训练系列产品</term>；支持PyTorch和MindSpore框架。</li><li>交换机亲和性调度2.0版本支持<term>Atlas A2 训练系列产品</term>；支持PyTorch框架。</li><li>只支持整卡进行交换机亲和性调度，不支持静态vNPU进行交换机亲和性调度。</li></ul>|

## （可选）集成昇腾插件扩展开源Volcano<a name="ZH-CN_TOPIC_0000002511426365"></a>

集群调度提供的Volcano组件是在开源Volcano的基础上新增了关于NPU调度相关的功能，该功能可通过集成集群调度为开发者提供的Ascend-volcano-plugin插件实现。开源[Volcano](https://volcano.sh/zh/#home_slider)框架支持插件机制供用户注册调度插件，实现不同的调度策略。

>[!NOTE] 
>Ascend-volcano-plugin目前支持开源Volcano v1.7.0和v1.9.0版本，且未对开源Volcano框架做修改。

**操作步骤<a name="section2672154791712"></a>**

1. 依次执行以下命令，在“$GOPATH/src/volcano.sh/”目录下拉取Volcano版本（以v1.7为例）官方开源代码。

    ```shell
    mkdir -p $GOPATH/src/volcano.sh/
    cd $GOPATH/src/volcano.sh/ 
    git clone -b release-1.7 https://github.com/volcano-sh/volcano.git
    ```

2. 将获取的[ascend-for-volcano](https://gitcode.com/Ascend/mind-cluster/tree/master/component/ascend-for-volcano)源码重命名为ascend-volcano-plugin，并上传至开源Volcano官方开源代码的插件路径下（“_$GOPATH_/src/volcano.sh/volcano/pkg/scheduler/plugins/”）。
3. <a name="li627818212613"></a>依次执行以下命令，编译开源Volcano二进制文件和华为NPU调度插件so文件。根据开源代码版本，为build.sh脚本选择对应的参数，如v1.7.0。

    ```shell
    cd $GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/build
    chmod +x build.sh
    ./build.sh v1.7.0
    ```

    >[!NOTE] 
    >编译出的二进制文件和动态链接库文件在“$GOPATH/src/volcano.sh/volcano/pkg/scheduler/plugins/ascend-volcano-plugin/output”目录下。

    编译后的文件列表见[表1](#table5623201371819)。

    **表 1**  output路径下的文件

    <a name="table5623201371819"></a>

    |文件名|说明|
    |--|--|
    |volcano-npu-<em>{version}</em>.so|华为NPU调度插件动态链接库|
    |Dockerfile-scheduler|volcano-scheduler镜像构建文本文件|
    |Dockerfile-controller|volcano-controller镜像构建文本文件|
    |volcano-<em>v{version}</em>.yaml|Volcano的启动配置文件|
    |vc-scheduler|volcano-scheduler组件二进制文件|
    |vc-controller-manager|volcano-controller组件二进制文件|

4. 选择以下两种方式之一，启动volcano-scheduler组件。
    - 使用集群调度组件提供的启动YAML，启动volcano-scheduler组件。
        1. 执行以下命令，制作Volcano镜像。根据开源代码版本，为镜像选择对应的参数，如v1.7.0。

            ```shell
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f ./Dockerfile-scheduler
            ```

        2. 执行以下命令，启动volcano-scheduler组件。

            ```shell
            kubectl apply -f volcano-v{version}.yaml
            ```

            启动示例如下。

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

    - 使用开源Volcano的启动YAML，启动volcano-scheduler组件。
        1. 将[步骤3](#li627818212613)中编译出的volcano-npu-_\{version\}_.so文件拷贝到开源Volcano的“\$GOPATH/src/volcano.sh/volcano”目录下；在开源Volcano的Dockerfile（路径为“\$GOPATH/src/volcano.sh/volcano/installer/dockerfile/scheduler/Dockerfile”）中添加如下命令。

            ```shell
            FROM golang:1.19.1 AS builder
            WORKDIR /go/src/volcano.sh/
            ADD . volcano
            RUN cd volcano && make vc-scheduler
            FROM alpine:latest
            COPY --from=builder /go/src/volcano.sh/volcano/_output/bin/vc-scheduler /vc-scheduler
            COPY volcano-npu_*.so plugins/     #新增
            ENTRYPOINT ["/vc-scheduler"]
            ```

        2. 依次执行以下命令，制作Volcano镜像。根据开源代码版本，为镜像选择对应的参数，如v1.7.0。

            ```shell
            cd $GOPATH/src/volcano.sh/volcano
            docker build --no-cache -t volcanosh/vc-scheduler:v1.7.0 ./ -f installer/dockerfile/scheduler/Dockerfile
            ```

        3. 修改volcano-development.yaml，该文件路径为“$GOPATH/src/volcano.sh/volcano/installer/volcano-development.yaml”。

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
                   <strong>- name: volcano-npu_v26.0.0_linux-x86_64    # 在ConfigMap中的新增自定义调度插件，请注意保持组件的版本配套关系</strong>
                 - plugins:
                   - name: overcommit
                   - name: drf
                     enablePreemptable: false
                   - name: predicates
                   - name: proportion
                   - name: nodeorder
                   - name: binpack
                <strong>configurations:           # 新增以下加粗字段，该字段为Volcano配置字段</strong>
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
                        <strong>- --plugins-dir=plugins       # 在volcano-scheduler启动命令中加载自定义插件</strong>
                        - -v=3
                        - 2>&1
            ---
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
              <strong>- apiGroups: [""]                          # 新增services的get权限</strong>  
                <strong>resources: ["services"]</strong>
                <strong>verbs: ["get"]</strong>
              - apiGroups: [""]
                resources: ["configmaps"]
                verbs: ["get", "create", "delete", "update",<strong>"list","watch"</strong>]    # 新增ConfigMap的list和watch权限
              - apiGroups: ["apps"]
                resources: ["daemonsets", "replicasets", "statefulsets"]
                verbs: ["list", "watch", "get"]
            ...</pre>

        4. 执行以下命令，启动volcano-scheduler组件。

            ```shell
            kubectl apply -f installer/volcano-development.yaml
            ```

            回显示例如下。

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
