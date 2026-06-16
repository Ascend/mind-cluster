# K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_0000002524312661"></a>

- 使用UB RDMA网络功能时，建议安装K8s RDMA Shared Dev Plugin。
- 仅使用基础容器化支持和资源监测的用户，可以不安装该组件，请直接跳过本章节。

## 操作步骤<a name="section135381552125415"></a>

1. 以root用户登录各计算节点，并执行以下命令查看镜像和版本号是否正确。

    ```shell
    docker images | grep k8s-rdma-shared-dev-plugin
    ```

   回显示例如下：

    ```ColdFusion
    k8s-rdma-shared-dev-plugin         v26.1.0              ef801847acd2        29 minutes ago      133MB
    ```

   - 是，执行[步骤2](#li26221441299)。
   - 否，请参见[准备镜像](./01_preparing_for_installation.md#准备镜像)，完成镜像制作和分发。

2. <a name="li26221441299"></a>将K8s RDMA Shared Dev Plugin软件包解压目录下的YAML文件，拷贝到K8s管理节点上任意目录。
3. 如不修改组件启动参数，可跳过本步骤。否则，请根据实际情况修改YAML文件中K8s RDMA Shared Dev Plugin的启动参数。启动参数请参见[表1](#table1862682843615)，可执行<b>./k8s-rdma-shared-dp -h</b>查看参数说明。
4. 在管理节点的YAML所在路径，执行以下命令，启动K8s RDMA Shared Dev Plugin。

    ```shell
    kubectl apply -f k8s-rdma-shared-dev-plugin-v{version}.yaml
    ```

   启动示例如下：

    ```ColdFusion
    serviceaccount/k8s-rdma-shared-dev-plugin created
    clusterrole.rbac.authorization.k8s.io/pods-rdma-role created
    clusterrolebinding.rbac.authorization.k8s.io/pods-rdma-rolebinding created
    daemonset.apps/k8s-rdma-shared-dev-plugin created
    ```

5. 执行以下命令，查看组件是否启动成功。

    ```shell
    kubectl get pod -n kube-system
    ```

   回显示例如下，出现 **Running** 表示组件启动成功。

    ```ColdFusion
    NAME                                             READY   STATUS    RESTARTS    AGE
    ...
    k8s-rdma-shared-dev-plugin-fd6t8                  1/1    Running      0        74s
    ...
    ```

> [!NOTE]
>
>- 安装组件后，组件的Pod状态不为
   Running，可参考[组件Pod状态不为Running](https://gitcode.com/Ascend/mind-cluster/issues/342)章节进行处理。
>- 安装组件后，组件的Pod状态为
   ContainerCreating，可参考[集群调度组件Pod处于ContainerCreating状态](https://gitcode.com/Ascend/mind-cluster/issues/343)章节进行处理。

## 参数说明<a name="section1851191618363"></a>

**表 1** K8s RDMA Shared Dev Plugin启动参数

<a name="table1862682843615"></a>
<table><thead align="left"><tr id="row462602873615"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p14626028143612"><a name="p14626028143612"></a><a name="p14626028143612"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p1362692863611"><a name="p1362692863611"></a><a name="p1362692863611"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p126271528193620"><a name="p126271528193620"></a><a name="p126271528193620"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="p13627192820363"><a name="p13627192820363"></a><a name="p13627192820363"></a>说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row162762819363"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p126271328193612"><a name="p126271328193612"></a><a name="p126271328193612"></a>-version / -v</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p2062718289368"><a name="p2062718289368"></a><a name="p2062718289368"></a>标志位</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1962732833612"><a name="p1962732833612"></a><a name="p1962732833612"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p146279281369"><a name="p146279281369"></a><a name="p146279281369"></a>查询当前K8s RDMA Shared Dev Plugin的版本号，该参数为标志位，无需跟值。使用示例：./k8s-rdma-shared-dp -version或./k8s-rdma-shared-dp -v</p>
</td>
</tr>
<tr id="row15627928153619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1627328103617"><a name="p1627328103617"></a><a name="p1627328103617"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p56272028193612"><a name="p56272028193612"></a><a name="p56272028193612"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p4627172833617"><a name="p4627172833617"></a><a name="p4627172833617"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p13627628113616"><a name="p13627628113616"></a><a name="p13627628113616"></a>日志级别：</p>
<a name="ul262712284363"></a><a name="ul262712284363"></a><ul id="ul262712284363"><li>取值为-1：debug</li><li>取值为0：info</li><li>取值为1：warning</li><li>取值为2：error</li><li>取值为3：critical</li></ul>
</td>
</tr>
<tr id="row126271928143620"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13627132863615"><a name="p13627132863615"></a><a name="p13627132863615"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p6627828173610"><a name="p6627828173610"></a><a name="p6627828173610"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p062752813613"><a name="p062752813613"></a><a name="p062752813613"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1262712893611"><a name="p1262712893611"></a><a name="p1262712893611"></a>日志备份时间，取值范围为7~700，单位为天。</p>
</td>
</tr>
<tr id="row862732873610"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1962772813620"><a name="p1962772813620"></a><a name="p1962772813620"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p162772823620"><a name="p162772823620"></a><a name="p162772823620"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p962817282369"><a name="p962817282369"></a><a name="p962817282369"></a>/var/log/mindx-dl/k8s-rdma-shared-dp/k8s-rdma-shared-dp.log</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1862816283367"><a name="p1862816283367"></a><a name="p1862816283367"></a>日志文件。单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为：k8s-rdma-shared-dp-触发转储的时间.log，如：k8s-rdma-shared-dp-2023-10-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row1862892813365"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10628202814367"><a name="p10628202814367"></a><a name="p10628202814367"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p4628828173618"><a name="p4628828173618"></a><a name="p4628828173618"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p16628182814364"><a name="p16628182814364"></a><a name="p16628182814364"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p10628172873610"><a name="p10628172873610"></a><a name="p10628172873610"></a>转储后日志文件保留个数上限，取值范围为1~30，单位为个。</p>
</td>
</tr>
<tr id="row68317556189"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101521"><a name="p0894319101521"></a><a name="p0894319101521"></a><span id="ph96781327191518"><a name="ph96781327191518"></a><a name="ph96781327191518"></a>-config-file</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151516"><a name="p108941719151516"></a><a name="p108941719151516"></a><span id="ph1899563312155"><a name="ph1899563312155"></a><a name="ph1899563312155"></a>string</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961514"><a name="p19894131961514"></a><a name="p19894131961514"></a><span id="ph67327379153"><a name="ph67327379153"></a><a name="ph67327379153"></a>/etc/kubernetes/kubelet-plugins.d/device-plugins/rdma_shared_dev_plugin.json</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971512"><a name="p589551971512"></a><a name="p589551971512"></a><span id="ph4556742141518"><a name="ph4556742141518"></a><a name="ph4556742141518"></a>配置文件路径，用于指定RDMA设备的选择器配置。</span></p>
</td>
</tr>
<tr id="row68317556190"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101522"><a name="p0894319101522"></a><a name="p0894319101522"></a><span id="ph96781327191520"><a name="ph96781327191520"></a><a name="ph96781327191520"></a>-use-cdi</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151518"><a name="p108941719151518"></a><a name="p108941719151518"></a><span id="ph1899563312157"><a name="ph1899563312157"></a><a name="ph1899563312157"></a>标志位</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961516"><a name="p19894131961516"></a><a name="p19894131961516"></a><span id="ph67327379155"><a name="ph67327379155"></a><a name="ph67327379155"></a>false</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971514"><a name="p589551971514"></a><a name="p589551971514"></a><span id="ph4556742141520"><a name="ph4556742141520"></a><a name="ph4556742141520"></a>是否使用CDI（Container Device Interface）模式向容器注册设备，该参数为标志位，无需跟值。使用示例：./k8s-rdma-shared-dp -use-cdi</span></p>
<p>UB类型的RDMA设备不支持CDI模式，当检测到UB设备时会自动禁用CDI。</p>
</td>
</tr>
<tr id="row10282191492319"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172319"><a name="p8283714172319"></a><a name="p8283714172319"></a>--enable-healthz</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472319"><a name="p828381472319"></a><a name="p828381472319"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482319"><a name="p828341482319"></a><a name="p828341482319"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432319"><a name="p828311432319"></a><a name="p828311432319"></a>是否启用健康检查服务。K8s部署时由组件YAML配置启用（true）。<ul><li>true：启用。</li><li>false：禁用。</li></ul></p>
</td>
</tr>
<tr id="row10282191492320"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172320"><a name="p8283714172320"></a><a name="p8283714172320"></a>--healthz-address</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472320"><a name="p828381472320"></a><a name="p828381472320"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482320"><a name="p828341482320"></a><a name="p828341482320"></a>11251</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432320"><a name="p828311432320"></a><a name="p828311432320"></a>健康检查服务侦听端口号，取值范围 1025~65535。K8s部署时由组件YAML配置为11257。若指定端口被占用，组件启动失败。</p>
</td>
</tr>
<tr id="row10282191492321"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172321"><a name="p8283714172321"></a><a name="p8283714172321"></a>--tls-cert-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472321"><a name="p828381472321"></a><a name="p828381472321"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482321"><a name="p828341482321"></a><a name="p828341482321"></a>""</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432321"><a name="p828311432321"></a><a name="p828311432321"></a>HTTPS 证书文件路径。为空则使用 HTTP 协议。与 --tls-private-key-file 必须同时配置或同时为空。</p>
</td>
</tr>
<tr id="row10282191492322"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172322"><a name="p8283714172322"></a><a name="p8283714172322"></a>--tls-private-key-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472322"><a name="p828381472322"></a><a name="p828381472322"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482322"><a name="p828341482322"></a><a name="p828341482322"></a>""</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432322"><a name="p828311432322"></a><a name="p828311432322"></a>HTTPS 私钥文件路径。为空则使用 HTTP 协议。与 --tls-cert-file 必须同时配置或同时为空。</p>
</td>
</tr>
<tr id="row10282191492318"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172318"><a name="p8283714172318"></a><a name="p8283714172318"></a>-h 或者 -help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p82838147235"><a name="p82838147235"></a><a name="p82838147235"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482318"><a name="p828341482318"></a><a name="p828341482318"></a>无</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432320"><a name="p828311432320"></a><a name="p828311432320"></a>显示帮助信息。</p>
</td>
</tr>
</tbody>
</table>
