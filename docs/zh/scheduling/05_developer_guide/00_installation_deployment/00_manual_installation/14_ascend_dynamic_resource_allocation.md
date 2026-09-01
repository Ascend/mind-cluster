# Ascend Dynamic Resource Allocation<a name="ZH-CN_TOPIC_0000002511426333"></a>

- 使用K8s动态资源分配（Dynamic Resource Allocation，DRA）机制申请NPU资源的用户，必须在计算节点安装Ascend Dynamic Resource Allocation组件（以下简称DRA驱动）。DRA驱动以kubelet插件形式部署，负责将节点上的NPU资源以ResourceSlice的形式上报至K8s，并在业务容器通过ResourceClaim申请NPU资源时，完成设备分配与注入。
- 不使用动态资源分配的用户，可以不安装该组件，请直接跳过本章节。

## 使用约束<a name="section1362795652417"></a>

在安装DRA驱动前，需要提前了解相关约束，具体说明请参见[表1](#table124141536dra01)。

**表 1**  约束说明

<a name="table124141536dra01"></a>
<table><thead align="left"><tr id="row0414dra240011"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p0414dra240011"><a name="p0414dra240011"></a><a name="p0414dra240011"></a>约束场景</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p0414dra240012"><a name="p0414dra240012"></a><a name="p0414dra240012"></a>约束说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row0414dra240013"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240013"><a name="p0414dra240013"></a><a name="p0414dra240013"></a>K8s版本</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240014"><a name="p0414dra240014"></a><a name="p0414dra240014"></a>组件的YAML文件中使用resource.k8s.io/v1与admissionregistration.k8s.io/v1 API，要求K8s版本为1.34.x~1.36.x，低版本K8s集群无法安装该组件。</p>
</td>
</tr>
<tr id="row0414dra240015"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240015"><a name="p0414dra240015"></a><a name="p0414dra240015"></a>容器运行时</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240016"><a name="p0414dra240016"></a><a name="p0414dra240016"></a>组件通过CDI（Container Device Interface）向业务容器注入NPU设备，要求计算节点的容器运行时支持CDI，如containerd 1.7.0及以上版本。</p>
</td>
</tr>
<tr id="row0414dra240017"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240017"><a name="p0414dra240017"></a><a name="p0414dra240017"></a>NPU驱动</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240018"><a name="p0414dra240018"></a><a name="p0414dra240018"></a>组件会调用NPU驱动的相关接口。如果要升级驱动，请先停止业务任务，再停止组件容器服务。</p>
</td>
</tr>
<tr id="row0414dra240019"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240019"><a name="p0414dra240019"></a><a name="p0414dra240019"></a>部署形态</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240020"><a name="p0414dra240020"></a><a name="p0414dra240020"></a>组件以特权容器方式运行，运行用户为root，并且挂载了宿主机的NPU驱动目录、kubelet插件目录与设备目录，请勿修改YAML文件中DaemonSet的名称及其挂载配置，否则组件无法正常工作。</p>
</td>
</tr>
<tr id="row0414dra240021"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240021"><a name="p0414dra240021"></a><a name="p0414dra240021"></a>支持芯片</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240022"><a name="p0414dra240022"></a><a name="p0414dra240022"></a>支持Ascend910系列和Ascend950系列处理器。</p>
</td>
</tr>
<tr id="row0414dra240023"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240023"><a name="p0414dra240023"></a><a name="p0414dra240023"></a>健康检查端口</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240024"><a name="p0414dra240024"></a><a name="p0414dra240024"></a>组件默认启用健康检查服务并占用11251端口，请确保计算节点上该端口未被占用。若端口冲突，可修改YAML文件中“--healthz-address”参数以及livenessProbe中的端口号。</p>
</td>
</tr>
</tbody>
</table>

## 操作步骤<a name="section83111543151613"></a>

1. 以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```shell
    docker images | grep ascend-dra
    ```

    回显示例如下：

    ```ColdFusion
    ascend-dra                         v26.2.0              29eec79eb693        About an hour ago   105MB
    ```

    - 是，执行[步骤2](#li2093514717dra01)。
    - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li2093514717dra01"></a>将Ascend Dynamic Resource Allocation软件包解压目录下的YAML文件（ascend-dra-driver-v<i>{version}</i>.yaml），拷贝到K8s管理节点上任意目录。

3. 如不修改组件启动参数，可跳过本步骤。否则，根据实际情况修改YAML文件中DRA驱动的启动参数，启动参数位于DaemonSet的“command”行，如下所示。启动参数请参见[表2](#table20426115dra02)，可执行<b>./ascend-dra -h</b>查看参数说明。

    ```Yaml
    ...
          containers:
          - name: plugin
            securityContext:
              privileged: true
            image: ascend-dra:v6.0.0
            command: ["/bin/bash", "-c", "exec ascend-dra --enable-healthz --healthz-address=11251"]
    ...
    ```

4. 在管理节点的YAML所在路径，执行以下命令，启动DRA驱动。

    ```shell
    kubectl apply -f ascend-dra-driver-v{version}.yaml
    ```

    回显示例如下：

    ```ColdFusion
    serviceaccount/ascend-dra-driver-service-account created
    clusterrole.rbac.authorization.k8s.io/ascend-dra-driver-role created
    clusterrolebinding.rbac.authorization.k8s.io/ascend-dra-driver-role-binding created
    daemonset.apps/ascend-dra-driver-kubeletplugin created
    deviceclass.resource.k8s.io/npu.huawei.com created
    validatingadmissionpolicy.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
    validatingadmissionpolicybinding.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
    ```

5. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n kube-system
    ```

    回显示例如下，出现**Running**表示组件启动成功。

    ```ColdFusion
    NAME                                        READY   STATUS    RESTARTS   AGE
    ascend-dra-driver-kubeletplugin-5m2xv       1/1     Running   0          74s
    ```

6. 执行以下命令，验证NPU资源是否上报成功。

    ```shell
    kubectl get deviceclass npu.huawei.com
    kubectl get resourceslices
    ```

    回显示例如下，能查询到名为“npu.huawei.com”的DeviceClass，且每个计算节点存在对应的ResourceSlice，表示资源上报成功。

    ```ColdFusion
    NAME            AGE
    npu.huawei.com  2m

    NAME                                  POOLNAME     DEVICECOUNT   AGE
    npu.huawei.com-node1-1a2b3c           node1        8             2m
    ```

>[!NOTE]
>
>- 安装组件后，组件的Pod状态不为Running，可参考[组件Pod状态不为Running](https://gitcode.com/Ascend/mind-cluster/issues/342)章节进行处理。
>- 安装组件后，组件的Pod状态为ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](https://gitcode.com/Ascend/mind-cluster/issues/343)章节进行处理。
>- 组件启动成功，找不到组件对应的Pod，可参考[组件启动YAML执行成功，找不到组件对应的Pod](https://gitcode.com/Ascend/mind-cluster/issues/345)章节信息。

## 参数说明<a name="section2042611570393"></a>

**表 2** DRA驱动启动参数

<a name="table20426115dra02"></a>
<table><thead align="left"><tr id="row0414dra240031"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p0414dra240031"><a name="p0414dra240031"></a><a name="p0414dra240031"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="10%" id="mcps1.2.5.1.2"><p id="p0414dra240032"><a name="p0414dra240032"></a><a name="p0414dra240032"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p0414dra240033"><a name="p0414dra240033"></a><a name="p0414dra240033"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="35%" id="mcps1.2.5.1.4"><p id="p0414dra240034"><a name="p0414dra240034"></a><a name="p0414dra240034"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row0414dra240039"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240039"><a name="p0414dra240039"></a><a name="p0414dra240039"></a>-cdi-root</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240040"><a name="p0414dra240040"></a><a name="p0414dra240040"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240041"><a name="p0414dra240041"></a><a name="p0414dra240041"></a>/var/run/cdi</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240042"><a name="p0414dra240042"></a><a name="p0414dra240042"></a>CDI文件的生成目录。组件分配NPU设备后会在该目录下生成CDI描述文件，kubelet依据该文件向业务容器注入NPU设备。</p>
</td>
</tr>
<tr id="row0414dra240043"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240043"><a name="p0414dra240043"></a><a name="p0414dra240043"></a>-kubelet-registrar-directory-path</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240044"><a name="p0414dra240044"></a><a name="p0414dra240044"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240045"><a name="p0414dra240045"></a><a name="p0414dra240045"></a>/var/lib/kubelet/plugins_registry</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240046"><a name="p0414dra240046"></a><a name="p0414dra240046"></a>kubelet插件注册目录。</p>
</td>
</tr>
<tr id="row0414dra240047"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240047"><a name="p0414dra240047"></a><a name="p0414dra240047"></a>-kubelet-plugins-directory-path</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240048"><a name="p0414dra240048"></a><a name="p0414dra240048"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240049"><a name="p0414dra240049"></a><a name="p0414dra240049"></a>/var/lib/kubelet/plugins</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240050"><a name="p0414dra240050"></a><a name="p0414dra240050"></a>kubelet插件数据目录。</p>
</td>
</tr>
<tr id="row0414dra240051"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240051"><a name="p0414dra240051"></a><a name="p0414dra240051"></a>-deviceResetTimeout</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240052"><a name="p0414dra240052"></a><a name="p0414dra240052"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240053"><a name="p0414dra240053"></a><a name="p0414dra240053"></a>600</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240054"><a name="p0414dra240054"></a><a name="p0414dra240054"></a>组件启动时，若芯片数量不足，等待驱动上报完整芯片的最大时长，单位为秒，取值范围为10~600。</p>
</td>
</tr>
<tr id="row0414dra240055"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240055"><a name="p0414dra240055"></a><a name="p0414dra240055"></a>-kubeconfig</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240056"><a name="p0414dra240056"></a><a name="p0414dra240056"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240057"><a name="p0414dra240057"></a><a name="p0414dra240057"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240058"><a name="p0414dra240058"></a><a name="p0414dra240058"></a>kubeconfig文件路径。集群内部署时无需配置，组件默认使用集群内认证方式。</p>
</td>
</tr>
<tr id="row0414dra240059"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240059"><a name="p0414dra240059"></a><a name="p0414dra240059"></a>-kube-api-qps</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240060"><a name="p0414dra240060"></a><a name="p0414dra240060"></a>float</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240061"><a name="p0414dra240061"></a><a name="p0414dra240061"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240062"><a name="p0414dra240062"></a><a name="p0414dra240062"></a>与K8s APIServer通信的QPS。当K8s APIServer请求压力变大时，可根据实际情况减小该值。</p>
</td>
</tr>
<tr id="row0414dra240063"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240063"><a name="p0414dra240063"></a><a name="p0414dra240063"></a>-kube-api-burst</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240064"><a name="p0414dra240064"></a><a name="p0414dra240064"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240065"><a name="p0414dra240065"></a><a name="p0414dra240065"></a>10</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240066"><a name="p0414dra240066"></a><a name="p0414dra240066"></a>与K8s APIServer通信的Burst。</p>
</td>
</tr>
<tr id="row0414dra240067"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240067"><a name="p0414dra240067"></a><a name="p0414dra240067"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240068"><a name="p0414dra240068"></a><a name="p0414dra240068"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240069"><a name="p0414dra240069"></a><a name="p0414dra240069"></a>/var/log/mindx-dl/ascend-dra/ascend-dra.log</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240070"><a name="p0414dra240070"></a><a name="p0414dra240070"></a>日志文件路径。单个日志文件超过20MB时会触发自动转储功能，文件大小上限不支持修改。</p>
</td>
</tr>
<tr id="row0414dra240071"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240071"><a name="p0414dra240071"></a><a name="p0414dra240071"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240072"><a name="p0414dra240072"></a><a name="p0414dra240072"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240073"><a name="p0414dra240073"></a><a name="p0414dra240073"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240074"><a name="p0414dra240074"></a><a name="p0414dra240074"></a>日志级别：</p>
<a name="ul0414dra240071"></a><a name="ul0414dra240071"></a><ul id="ul0414dra240071"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row0414dra240075"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240075"><a name="p0414dra240075"></a><a name="p0414dra240075"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240076"><a name="p0414dra240076"></a><a name="p0414dra240076"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240077"><a name="p0414dra240077"></a><a name="p0414dra240077"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240078"><a name="p0414dra240078"></a><a name="p0414dra240078"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row0414dra240079"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240079"><a name="p0414dra240079"></a><a name="p0414dra240079"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240080"><a name="p0414dra240080"></a><a name="p0414dra240080"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240081"><a name="p0414dra240081"></a><a name="p0414dra240081"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240082"><a name="p0414dra240082"></a><a name="p0414dra240082"></a>转储后日志文件保留个数上限，取值范围为1~180，单位为个。</p>
</td>
</tr>
<tr id="row0414dra240083"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240083"><a name="p0414dra240083"></a><a name="p0414dra240083"></a>-enable-healthz</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240084"><a name="p0414dra240084"></a><a name="p0414dra240084"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240085"><a name="p0414dra240085"></a><a name="p0414dra240085"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240086"><a name="p0414dra240086"></a><a name="p0414dra240086"></a>是否启用健康检查服务。K8s部署时由组件YAML配置启用（true）。</p>
<a name="ul0414dra240083"></a><a name="ul0414dra240083"></a><ul id="ul0414dra240083"><li>true：启用。</li><li>false：禁用。</li></ul>
</td>
</tr>
<tr id="row0414dra240087"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240087"><a name="p0414dra240087"></a><a name="p0414dra240087"></a>-healthz-address</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240088"><a name="p0414dra240088"></a><a name="p0414dra240088"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240089"><a name="p0414dra240089"></a><a name="p0414dra240089"></a>11251</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240090"><a name="p0414dra240090"></a><a name="p0414dra240090"></a>健康检查服务的监听端口。修改该参数时，需同步修改YAML文件中livenessProbe配置的端口号。</p>
</td>
</tr>
<tr id="row0414dra240091"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240091"><a name="p0414dra240091"></a><a name="p0414dra240091"></a>-tls-cert-file</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240092"><a name="p0414dra240092"></a><a name="p0414dra240092"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240093"><a name="p0414dra240093"></a><a name="p0414dra240093"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240094"><a name="p0414dra240094"></a><a name="p0414dra240094"></a>健康检查服务启用HTTPS时使用的TLS证书文件路径，未配置时健康检查服务使用HTTP。</p>
</td>
</tr>
<tr id="row0414dra240095"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240095"><a name="p0414dra240095"></a><a name="p0414dra240095"></a>-tls-private-key-file</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240096"><a name="p0414dra240096"></a><a name="p0414dra240096"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240097"><a name="p0414dra240097"></a><a name="p0414dra240097"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240098"><a name="p0414dra240098"></a><a name="p0414dra240098"></a>健康检查服务启用HTTPS时使用的TLS私钥文件路径，需与-tls-cert-file同时配置。</p>
</td>
</tr>
</tbody>
</table>
