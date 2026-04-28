# ClusterD<a name="ZH-CN_TOPIC_0000002511346341"></a>

- 使用整卡调度、静态vNPU调度、动态vNPU调度、断点续训、弹性训练、推理卡故障恢复或推理卡故障重调度的用户，必须安装ClusterD。集群中同时存在Ascend Device Plugin和NodeD组件时，ClusterD才能提供全量的信息收集服务。
- 在安装ClusterD时，建议提前安装Volcano。若ClusterD先于Volcano安装，ClusterD所在的Pod可能会CrashLoopBackOff，需等待Volcano的Pod启动后，ClusterD才会恢复正常。
- 仅使用容器化支持和资源监测的用户，可以不安装ClusterD，请直接跳过本章节。
- 启动ClusterD之前，需要先启动Ascend Operator。ClusterD需要感知Ascend Job资源。
- 使用慢节点&慢网络故障功能前，需安装ClusterD，详细说明请参见[慢节点&慢网络故障](../../../usage/resumable_training/01_solutions_principles.md#慢节点慢网络故障)。

## 操作步骤<a name="section20114193212615"></a>

1. 以root用户登录K8s管理节点，并执行以下命令，查看ClusterD镜像和版本号是否正确。

    ```shell
    docker images | grep clusterd
    ```

    回显示例如下：

    ```ColdFusion
    clusterd                   v26.0.0              c532e9d0889c        About an hour ago         126MB
    ```

    - 是，执行[步骤2](#li615118054419)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li615118054419"></a>将ClusterD软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中ClusterD的启动参数。启动参数请参见[表2](#table11614104894617)，可以在ClusterD二进制包的目录下执行<b>./clusterd -h</b>查看参数说明。
4. （可选）在“clusterd-v<i>\{version\}</i>.yaml”中，配置人工隔离芯片检测开关及故障频率、解除隔离时间等。

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

    **表 1**  manually_separate_policy.conf的参数说明

    <a name="table208901"></a>

    |一级参数|二级参数|类型|说明|
    |--|--|--|--|
    |enabled|-|bool|人工隔离芯片的检测开关。取值包括：<ul><li>true：开启人工隔离芯片检测功能。</li><li>false：关闭人工隔离芯片检测功能。</li></ul><p>默认值为true。若关闭该开关，会将所有ClusterD人工隔离的芯片及相关缓存都清除。</p><div class="note"><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p>YAML规范支持多种布尔值的写法（含大小写变体），但不同解析器（如K8s、Go、Python）的兼容度不同，不是所有写法都支持。推荐统一使用小写true/false。</p></div></div>|
    |separate|fault_window_hours|int|人工隔离芯片的时间。在该时间内，同一个故障码的故障次数达到fault_threshold取值，ClusterD会将故障芯片进行人工隔离。取值范围为[1, 720]，默认值为24，单位为h（小时）。|
    |-|fault_threshold|int|人工隔离芯片的阈值。取值范围为[1, 50]，默认值为3，单位为次。|
    |release|fault_free_hours|int|解除隔离的时间，表示距离最后一次达到频率进行隔离的时间，超过该时间会解除隔离。取值范围为[1, 240]或-1，默认值为48，单位为h（小时）。<ul><li>最后一次达到频率的时间即为clusterd-manual-info-cm中的LastSeparateTime。clusterd-manual-info-cm的说明请参见[clusterd-manual-info-cm](../../../api/clusterd/00_cluster_resources.md#clusterd-manual-info-cm)。</li><li>配置为-1，表示关闭解除隔离功能。</li><li>达到解除隔离时间进行自动解除隔离时，无论故障是否恢复，都会解除。</li></ul>|

    >[!NOTE]
    >若enabled字段缺失，ClusterD会识别为false；若其他int类型字段缺失，ClusterD会识别为0。
 
5. 在管理节点的YAML所在路径，执行以下命令，启动ClusterD。

    ```shell
    kubectl apply -f clusterd-v{version}.yaml
    ```

    启动示例如下：

    ```ColdFusion
    clusterrolebinding.rbac.authorization.k8s.io/pods-clusterd-rolebinding created
    lease.coordination.k8s.io/cluster-info-collector created
    deployment.apps/clusterd created
    service/clusterd-grpc-svc created
    ```

6. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                          READY   STATUS              RESTARTS   AGE
    clusterd-7844cb867d-fwcj7     0/1     Running            0          45s
    ```

>[!NOTE] 
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

## 参数说明<a name="section1250239182212"></a>

**表 2** ClusterD启动参数

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
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p5614648134615"><a name="p5614648134615"></a><a name="p5614648134615"></a>查询<span id="ph1950137183918"><a name="ph1950137183918"></a><a name="ph1950137183918"></a>ClusterD</span>版本号。</p>
<a name="ul178554235168"></a><a name="ul178554235168"></a><ul id="ul178554235168"><li>true：查询。</li><li>false：不查询。</li></ul>
</td>
</tr>
<tr id="row56148488467"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p661410488460"><a name="p661410488460"></a><a name="p661410488460"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p561484814469"><a name="p561484814469"></a><a name="p561484814469"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p8614134874616"><a name="p8614134874616"></a><a name="p8614134874616"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p6614174884615"><a name="p6614174884615"></a><a name="p6614174884615"></a>日志级别：</p>
<a name="ul76142481467"></a><a name="ul76142481467"></a><ul id="ul76142481467"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row1961574813469"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p14615748144617"><a name="p14615748144617"></a><a name="p14615748144617"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p76151484466"><a name="p76151484466"></a><a name="p76151484466"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row1061520484468"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p161512481467"><a name="p161512481467"></a><a name="p161512481467"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p16615048134619"><a name="p16615048134619"></a><a name="p16615048134619"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p1668892293119"><a name="p1668892293119"></a><a name="p1668892293119"></a>/var/log/mindx-dl/clusterd/clusterd.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：clusterd-触发转储的时间.log，如：clusterd-2024-06-07T03-38-24.402.log。</p>
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
<tr id="row147481810102010"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p15748191011204"><a name="p15748191011204"></a><a name="p15748191011204"></a>-useProxy</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p id="p17830536152010"><a name="p17830536152010"></a><a name="p17830536152010"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p13748141013205"><a name="p13748141013205"></a><a name="p13748141013205"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p2748131042020"><a name="p2748131042020"></a><a name="p2748131042020"></a>是否使用代理转发gRPC请求。</p>
<a name="ul71770166215"></a><a name="ul71770166215"></a><ul id="ul71770166215"><li>true：是</li><li>false：否
</li></ul><div class="note" id="note12300045132119"><a name="note12300045132119"></a><a name="note12300045132119"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p17300245162118"><a name="p17300245162118"></a><a name="p17300245162118"></a>建议在启动YAML中将本参数取值配置为“true”，并对ClusterD进行安全加固，详细说明请参见<a href="../../../security_hardening.md#clusterd安全加固">ClusterD安全加固</a>章节。</p>
</div></div>
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
