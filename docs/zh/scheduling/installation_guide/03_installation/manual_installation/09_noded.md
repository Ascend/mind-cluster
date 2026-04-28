# NodeD<a name="ZH-CN_TOPIC_0000002479226406"></a>

- 使用整卡调度、静态vNPU调度、动态vNPU调度、推理卡故障恢复、推理卡故障重调度、断点续训或弹性训练时，必须安装NodeD。
- 仅使用容器化支持和资源监测的用户，可以不安装NodeD，请直接跳过本章节。
- 使用慢节点&慢网络故障功能前，需安装NodeD，详细说明请参见[慢节点&慢网络故障](../../../usage/resumable_training/01_solutions_principles.md#慢节点慢网络故障)。

## 操作步骤<a name="section135381552125414"></a>

1. 以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```shell
    docker images | grep noded
    ```

    回显示例如下：

    ```ColdFusion
    noded                               v26.0.0              ef801847acd2        29 minutes ago      133MB
    ```

    - 是，执行[步骤2](#li26221447455)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li26221447455"></a>将NodeD软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中NodeD的启动参数。启动参数请参见[表1](#table1862682843614)，可执行<b>./noded -h</b>查看参数说明。
4. （可选）使用**断点续训**或者**弹性训练**时，需要配置节点状态上报间隔。在NodeD启动YAML文件的“args”行增加“-reportInterval”参数，如下所示：

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
    >- K8s[默认40秒未收到节点响应时](https://kubernetes.io/zh-cn/docs/concepts/architecture/nodes/)将该节点置为NotReady。
    >- 当K8s APIServer请求压力变大时，可根据实际情况增大间隔时间，以减轻APIServer压力。

5. 在管理节点的YAML所在路径，执行以下命令，启动NodeD。
    - 不使用[dpc故障检测](../../../usage/resumable_training/01_solutions_principles.md#节点故障)功能，请执行以下命令。

        ```shell
        kubectl apply -f noded-v{version}.yaml
        ```

    - 如果环境已部署Scale-Out Storage DPC 24.2.0及以上版本，并且使用[dpc故障检测](../../../usage/resumable_training/01_solutions_principles.md#节点故障)功能，则执行以下命令，启动NodeD。

        ```shell
        kubectl apply -f noded-dpc-v{version}.yaml
        ```

        启动示例如下：

        ```ColdFusion
        serviceaccount/noded created
        clusterrole.rbac.authorization.k8s.io/pods-noded-role created
        clusterrolebinding.rbac.authorization.k8s.io/pods-noded-rolebinding created
        daemonset.apps/noded created
        ```

6. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                              READY   STATUS    RESTARTS   AGE
    ...
    noded-fd6t8                  1/1    Running  0        74s
    ...
    ```

>[!NOTE]
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

## 参数说明<a name="section1851191618362"></a>

**表 1** NodeD启动参数

<a name="table1862682843614"></a>
<table><thead align="left"><tr id="row462602873614"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p14626028143611"><a name="p14626028143611"></a><a name="p14626028143611"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p136269286369"><a name="p136269286369"></a><a name="p136269286369"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p126271528193618"><a name="p126271528193618"></a><a name="p126271528193618"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="p13627192820361"><a name="p13627192820361"></a><a name="p13627192820361"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row162762819362"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p126271328193610"><a name="p126271328193610"></a><a name="p126271328193610"></a>-reportInterval</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p2062718289366"><a name="p2062718289366"></a><a name="p2062718289366"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1962732833610"><a name="p1962732833610"></a><a name="p1962732833610"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><a name="ul49338283283"></a><a name="ul49338283283"></a><ul id="ul49338283283"><li>上报节点故障信息的最小间隔，如果节点状态有变化，那么在5s内就会上报，如果节点状态持续没有变化，那么上报周期为30分钟。</li><li>取值范围为1~300，单位为秒。</li><li>当K8s APIServer请求压力变大时，可根据实际情况增大间隔时间，以减轻APIServer压力。</li></ul>
</td>
</tr>
<tr id="row1240181274312"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1691522724316"><a name="p1691522724316"></a><a name="p1691522724316"></a>-monitorPeriod</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p17916227194316"><a name="p17916227194316"></a><a name="p17916227194316"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1491652715431"><a name="p1491652715431"></a><a name="p1491652715431"></a>60</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p139161227154317"><a name="p139161227154317"></a><a name="p139161227154317"></a>节点硬件故障的轮询检测周期，取值范围为60~600，单位为秒。</p>
</td>
</tr>
<tr id="row562722803619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p862732803617"><a name="p862732803617"></a><a name="p862732803617"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p166271328153612"><a name="p166271328153612"></a><a name="p166271328153612"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p176271728143613"><a name="p176271728143613"></a><a name="p176271728143613"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p146279281367"><a name="p146279281367"></a><a name="p146279281367"></a>是否查询当前<span id="ph1437310218483"><a name="ph1437310218483"></a><a name="ph1437310218483"></a>NodeD</span>的版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row15627928153617"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1627328103615"><a name="p1627328103615"></a><a name="p1627328103615"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p56272028193610"><a name="p56272028193610"></a><a name="p56272028193610"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p4627172833615"><a name="p4627172833615"></a><a name="p4627172833615"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p13627628113614"><a name="p13627628113614"></a><a name="p13627628113614"></a>日志级别：</p>
<a name="ul262712284361"></a><a name="ul262712284361"></a><ul id="ul262712284361"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row126271928143618"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13627132863613"><a name="p13627132863613"></a><a name="p13627132863613"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p662782817368"><a name="p662782817368"></a><a name="p662782817368"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p062752813611"><a name="p062752813611"></a><a name="p062752813611"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p126271289369"><a name="p126271289369"></a><a name="p126271289369"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row0896102832513"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p178963287252"><a name="p178963287252"></a><a name="p178963287252"></a>-resultMaxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p18896202818250"><a name="p18896202818250"></a><a name="p18896202818250"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p48961228192511"><a name="p48961228192511"></a><a name="p48961228192511"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p198961128162516"><a name="p198961128162516"></a><a name="p198961128162516"></a>pingmesh结果备份文件保留的天数。取值范围为[7, 700]，单位为天。</p>
<div class="note" id="note1058610517274"><a name="note1058610517274"></a><a name="note1058610517274"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p946415413280"><a name="p946415413280"></a><a name="p946415413280"></a>该参数仅支持在<span id="ph077885871817"><a name="ph077885871817"></a><a name="ph077885871817"></a>Atlas 900 A3 SuperPoD 超节点</span>上使用。且所使用的驱动版本需≥24.1.RC1。</p>
</div></div>
</td>
</tr>
<tr id="row86273287368"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1962772813618"><a name="p1962772813618"></a><a name="p1962772813618"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p162772823618"><a name="p162772823618"></a><a name="p162772823618"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p962817282367"><a name="p962817282367"></a><a name="p962817282367"></a>/var/log/mindx-dl/noded/noded.log</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1862816283365"><a name="p1862816283365"></a><a name="p1862816283365"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：noded-触发转储的时间.log，如：noded-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row1862892813363"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10628202814365"><a name="p10628202814365"></a><a name="p10628202814365"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p4628828173616"><a name="p4628828173616"></a><a name="p4628828173616"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p16628182814362"><a name="p16628182814362"></a><a name="p16628182814362"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1062817287368"><a name="p1062817287368"></a><a name="p1062817287368"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row68317556187"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101519"><a name="p0894319101519"></a><a name="p0894319101519"></a><span id="ph96781327191516"><a name="ph96781327191516"></a><a name="ph96781327191516"></a>-deviceResetTimeout</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151514"><a name="p108941719151514"></a><a name="p108941719151514"></a><span id="ph1899563312153"><a name="ph1899563312153"></a><a name="ph1899563312153"></a>int</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961512"><a name="p19894131961512"></a><a name="p19894131961512"></a><span id="ph67327379151"><a name="ph67327379151"></a><a name="ph67327379151"></a>60</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971510"><a name="p589551971510"></a><a name="p589551971510"></a><span id="ph4556742141516"><a name="ph4556742141516"></a><a name="ph4556742141516"></a>组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600</span><span id="ph124041056151513"><a name="ph124041056151513"></a><a name="ph124041056151513"></a>。</span></p>
<a name="ul1354220213192"></a><a name="ul1354220213192"></a><ul id="ul1354220213192"><li><span id="ph278017516257"><a name="ph278017516257"></a><a name="ph278017516257"></a><term id="zh-cn_topic_0000001519959665_term57208119917"><a name="zh-cn_topic_0000001519959665_term57208119917"></a><a name="zh-cn_topic_0000001519959665_term57208119917"></a>Atlas A2 训练系列产品</term></span>、<span id="ph13163257131918"><a name="ph13163257131918"></a><a name="ph13163257131918"></a>Atlas 800I A2 推理服务器</span>、<span id="ph10930753142211"><a name="ph10930753142211"></a><a name="ph10930753142211"></a>A200I A2 Box 异构组件</span>：建议配置为150秒。</li><li><span id="ph531432344210"><a name="ph531432344210"></a><a name="ph531432344210"></a><term id="zh-cn_topic_0000001519959665_term26764913715"><a name="zh-cn_topic_0000001519959665_term26764913715"></a><a name="zh-cn_topic_0000001519959665_term26764913715"></a>Atlas A3 训练系列产品</term></span>、<span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>A200T A3 Box8 超节点服务器</span>、<span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 800I A3 超节点服务器</span>：建议配置为360秒。</li><li><span>Atlas 350 标卡</span>、<span id="ph1692713816224"><a name="ph1692713816224"></a><a name="ph1692713816224"></a>Atlas 850 系列硬件产品</span>、<span id="ph18760103420211"><a name="ph18760103420211"></a><a name="ph18760103420211"></a>Atlas 950 SuperPoD</span>：建议配置为600秒。</li></ul>
</td>
</tr>
<tr id="row10282191492316"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p4283714172316"><a name="p4283714172316"></a><a name="p4283714172316"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p82838147233"><a name="p82838147233"></a><a name="p82838147233"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482316"><a name="p828341482316"></a><a name="p828341482316"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432318"><a name="p828311432318"></a><a name="p828311432318"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>
