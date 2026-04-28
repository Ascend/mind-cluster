# Infer Operator<a name="ZH-CN_TOPIC_0000002479386414"></a>

下发InferServiceSet类型的推理任务，必须安装Infer Operator。

## 操作步骤<a name="section209273712583"></a>

1. 以root用户登录K8s管理节点，并执行以下命令，查看Infer Operator镜像和版本号是否正确。

    ```shell
    docker images | grep infer-operator
    ```

    回显示例如下：

    ```ColdFusion
    infer-operator                      v26.0.0              c7221984e8ae        About an hour ago         140MB
    ```

    - 是，执行[步骤2](#li19793191914420)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li19793191914420"></a>将Infer Operator软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中Infer Operator的启动参数。启动参数请参见[表1](#table11614104894618)，可执行<b>./infer-operator -h</b>查看参数说明。
4. 在管理节点的YAML所在路径，执行以下命令，启动Infer Operator。

    ```shell
    kubectl apply -f infer-operator-v{version}.yaml
    ```

    启动示例如下：

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

5. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n mindx-dl
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                                         READY   STATUS    RESTARTS   AGE
    ...
    infer-operator-8322455ba7b-hwmjw      1/1    Running  0         11s
    ```

>[!NOTE]
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](../../../faq.md#组件pod状态不为running)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](../../../faq.md#集群调度组件pod处于containercreating状态)章节进行处理。
>- 启动组件失败，可参考[启动集群调度组件失败，日志打印“get sem errno =13”](../../../faq.md#启动集群调度组件失败日志打印get-sem-errno-13)章节信息。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](../../../faq.md#组件启动yaml执行成功找不到组件对应的pod)章节信息。

## 参数说明<a name="section91521925121114"></a>

**表 1** Infer Operator启动参数

<a name="table11614104894618"></a>
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
<tbody><tr><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p>-version</p>
</td>
<td class="cellrowborder" valign="top" width="14.979999999999999%" headers="mcps1.2.5.1.2 "><p>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p>false</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p>查询<span>ClusterD</span>版本号。</p><ul><li>true：查询。</li><li>false：不查询。</li></ul>
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
<td class="cellrowborder" valign="top" width="15.02%" headers="mcps1.2.5.1.3 "><p id="p56159486469"><a name="p56159486469"></a><a name="p56159486469"></a>/var/log/mindx-dl/infer-operator/infer-operator.log</p>
</td>
<td class="cellrowborder" valign="top" width="40%" headers="mcps1.2.5.1.4 "><p id="p16151448144613"><a name="p16151448144613"></a><a name="p16151448144613"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：infer-operator-触发转储的时间.log，如：infer-operator-2023-10-07T03-38-24.402.log。</p>
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
