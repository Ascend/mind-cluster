# Ascend Operator<a name="ZH-CN_TOPIC_0000002479386414"></a>

- 使用整卡调度（训练）、静态vNPU调度（训练）、断点续训或弹性训练的用户，必须安装Ascend Operator组件。如果使用Volcano组件作为调度器，需要先安装Volcano组件，否则Ascend Operator会启动失败。
- 使用整卡调度（推理）和推理卡故障重调度特性，下发acjob类型的分布式推理任务，必须安装Ascend Operator。
- 仅使用容器化支持和资源监测、推理卡故障恢复或推理卡故障重调度（单机任务）的用户，可以不安装Ascend Operator，请直接跳过本章节。
- 启动Ascend Operator之前，需要先启动Volcano。Ascend Operator需要感知PodGroup资源。
- Ascend Operator组件允许创建的单个AscendJob任务的最大副本数量为20000。

## 操作步骤<a name="section209273712583"></a>

1. 以root用户登录K8s管理节点，并执行以下命令，查看Ascend Operator镜像和版本号是否正确。

    ```shell
    docker images | grep ascend-operator
    ```

    回显示例如下：

    ```ColdFusion
    ascend-operator                      v26.0.0              c532e9d0889c        About an hour ago         137MB
    ```

    - 是，执行[步骤2](#li19793191914420)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li19793191914420"></a>将Ascend Operator软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Ascend Operator的启动参数。启动参数请参见[表1](#table11614104894617)，可执行<b>./ascend-operator -h</b>查看参数说明。
4. （可选）使用Ascend Operator为PyTorch和MindSpore框架下的训练任务生成集合通信配置文件（RankTable File，也叫[hccl.json](../../../api/hccl.json_file_description.md)文件），缩短集群通信建链时间。使用其他框架的用户，可跳过本步骤。
    1. 启动YAML中已经默认挂载了hccl.json文件的父目录，用户可以根据实际情况进行修改。

        <pre codetype="yaml">
        ...
                - name: ranktable-dir
                  mountPath: /user/mindx-dl/ranktable        # 容器内路径，不可修改
        ...
              volumes:
                - name: ascend-operator-log
                  hostPath:
                    path: /var/log/mindx-dl/ascend-operator
                    type: Directory
                <strong>- name: ranktable-dir</strong>
                  <strong>hostPath:</strong>
                    <strong>path: /user/mindx-dl/ranktable    # 宿主机路径，任务YAML中hccl.json文件保存路径的根目录必须和宿主机路径保持一致</strong>
                    <strong>type: DirectoryOrCreate                                      # 用于检查给定文件夹是否存在，若不存在，则会创建空文件夹</strong>
        ...</pre>

        >[!NOTE] 
        >- 容器内RankTable根目录路径不可修改，宿主机路径可以修改。用户部署任务时，任务YAML中hccl.json文件保存路径的**根目录**必须和宿主机路径保持一致。
        >- RankTable根目录文件夹权限，必须满足以下任意一个条件。
        >    - 所属的用户和用户组为hwMindX（集群调度组件默认的运行用户）。
        >    - RankTable根目录文件夹权限为777。

    2. 执行以下命令，在父目录下创建hccl.json文件的具体挂载路径。

        ```shell
        mkdir -m 777 /user/mindx-dl/ranktable/{具体挂载路径}
        ```

5. 在管理节点的YAML所在路径，执行以下命令，启动Ascend Operator。

    ```shell
    kubectl apply -f ascend-operator-v{version}.yaml
    ```

    启动示例如下：

    ```ColdFusion
    deployment.apps/ascend-operator-manager created
    serviceaccount/ascend-operator-manager created
    clusterrole.rbac.authorization.k8s.io/ascend-operator-manager-role created
    clusterrolebinding.rbac.authorization.k8s.io/ascend-operator-manager-rolebinding created
    customresourcedefinition.apiextensions.k8s.io/ascendjobs.mindxdl.gitee.com created
    ...
    ```

6. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                                         READY   STATUS    RESTARTS   AGE
    ...
    ascend-operator-7667495b6b-hwmjw      1/1    Running  0         11s
    ```

>[!NOTE] 
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

## 参数说明<a name="section91521925121114"></a>

**表 1** Ascend Operator启动参数

<a name="table11614104894617"></a>
<table><thead align="left"><tr id="row2614114884616"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p961416489463"><a name="p961416489463"></a><a name="p961416489463"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="14.979999999999999%" id="mcps1.2.5.1.2"><p id="p6614174812464"><a name="p6614174812464"></a><a name="p6614174812464"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15.02%" id="mcps1.2.5.1.3"><p id="p12614194844618"><a name="p12614194844618"></a><a name="p12614194844618"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="40%" id="mcps1.2.5.1.4"><p id="p261454810466"><a name="p261454810466"></a><a name="p261454810466"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row14614134874619"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p86145488460"><a name="p86145488460"></a><a name="p86145488460"></a>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p20614848194617"><a name="p20614848194617"></a><a name="p20614848194617"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p26141489467"><a name="p26141489467"></a><a name="p26141489467"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>是否查询<span id="ph446121313413"><a name="ph446121313413"></a><a name="ph446121313413"></a>Ascend Operator</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2312104517312"><a name="p2312104517312"></a><a name="p2312104517312"></a>日志级别支持如下几种取值：</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>日志备份时间限制，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p56159486469"><a name="p56159486469"></a><a name="p56159486469"></a>/var/log/mindx-dl/ascend-operator/ascend-operator.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：ascend-operator-触发转储的时间.log，如：ascend-operator-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row8615248184611"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p361516481465"><a name="p361516481465"></a><a name="p361516481465"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row25282845417"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p155314286546"><a name="p155314286546"></a><a name="p155314286546"></a>-enableGangScheduling</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p19531128135415"><a name="p19531128135415"></a><a name="p19531128135415"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p55362825414"><a name="p55362825414"></a><a name="p55362825414"></a>true</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p3537285549"><a name="p3537285549"></a><a name="p3537285549"></a>是否启用“gang”策略调度，默认开启。开启时根据任务指定的调度器进行任务调度。“gang”策略调度说明请参见<a href="https://volcano.sh/zh/docs/v1-7-0/plugins/" target="_blank" rel="noopener noreferrer">开源Volcano官方文档</a>。</p>
<a name="ul1161205685015"></a><a name="ul1161205685015"></a><ul id="ul1161205685015"><li>true：启用“gang”策略调度。<p id="p1469315258274"><a name="p1469315258274"></a><a name="p1469315258274"></a>使用Job级别弹性扩缩容功能时，需将本字段的取值设置为true。</p>
</li><li>false：不启用“gang”策略调度。</li></ul>
</td>
</tr>
<tr id="row1758314497918"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1231919131198"><a name="p1231919131198"></a><a name="p1231919131198"></a>-isCompress</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p631912134197"><a name="p631912134197"></a><a name="p631912134197"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p031911314198"><a name="p031911314198"></a><a name="p031911314198"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1131914134197"><a name="p1131914134197"></a><a name="p1131914134197"></a>当日志文件大小达到转储阈值时，是否对日志文件进行压缩转储（该参数后面将会弃用）。</p>
<a name="ul1529912275516"></a><a name="ul1529912275516"></a><ul id="ul1529912275516"><li>true：压缩转储。</li><li>false：不压缩转储。</li></ul>
</td>
</tr>
<tr id="row1636910277610"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p836962710617"><a name="p836962710617"></a><a name="p836962710617"></a>-kubeconfig</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p1536942715620"><a name="p1536942715620"></a><a name="p1536942715620"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p7628122811369"><a name="p7628122811369"></a><a name="p7628122811369"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p1231013103910"><a name="p1231013103910"></a><a name="p1231013103910"></a>kubeconfig的路径，当程序运行于集群外时必须配置。</p>
</td>
</tr>
<tr id="row57381540134219"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14739164054215"><a name="p14739164054215"></a><a name="p14739164054215"></a>-kubeApiBurst</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p0739184012425"><a name="p0739184012425"></a><a name="p0739184012425"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1273915409420"><a name="p1273915409420"></a><a name="p1273915409420"></a>100</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5739114017421"><a name="p5739114017421"></a><a name="p5739114017421"></a>与K8s通信时使用的突发流量。取值范围为（0,10000]，不在取值范围内使用默认值100。</p>
</td>
</tr>
<tr id="row182053596442"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p4205165917447"><a name="p4205165917447"></a><a name="p4205165917447"></a>-kubeApiQps</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p42051059174419"><a name="p42051059174419"></a><a name="p42051059174419"></a>float32</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p17205159154412"><a name="p17205159154412"></a><a name="p17205159154412"></a>50</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p172054590444"><a name="p172054590444"></a><a name="p172054590444"></a>与K8s通信时使用的QPS（每秒请求率）。取值范围为（0,10000]，不在取值范围内使用默认值50。</p>
</td>
</tr>
<tr id="row2615144813463"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p1061594884617"><a name="p1061594884617"></a><a name="p1061594884617"></a>-h或者-help</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16151748144614"><a name="p16151748144614"></a><a name="p16151748144614"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13615048184615"><a name="p13615048184615"></a><a name="p13615048184615"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16616174834615"><a name="p16616174834615"></a><a name="p16616174834615"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>
