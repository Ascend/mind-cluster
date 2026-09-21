# K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_0000002524312661"></a>

- When using the UB RDMA network feature, it is recommended to install the K8s RDMA Shared Dev Plugin.
- Users who only use basic containerization support and resource monitoring can skip installing this component and proceed directly to the next chapter.

## Procedure<a name="section135381552125415"></a>

1. Log in to each compute node as the root user and run the following command to check whether the image and version number are correct.

    ```shell
    docker images | grep k8s-rdma-shared-dp
    ```

   The output is similar to the following:

    ```ColdFusion
    k8s-rdma-shared-dp         v26.1.0              ef801847acd2        29 minutes ago      133MB
    ```

   - If yes, go to [Step 2](#li26221441299).
   - If no, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to create and distribute the image.

2. <a name="li26221441299"></a>Copy the YAML file in the directory where the K8s RDMA Shared Dev Plugin software package is extracted to any directory on the K8s management node.
3. If you do not modify the component startup parameters, you can skip this step. Otherwise, modify the startup parameters of K8s RDMA Shared Dev Plugin in the YAML file based on the actual situation. For details about the startup parameters, see [Table 1](#table1862682843615). You can run <b>./k8s-rdma-shared-dp -h</b> to view the parameter description.
4. In the path where the YAML file resides on the management node, run the following command to start K8s RDMA Shared Dev Plugin.

    ```shell
    kubectl apply -f k8s-rdma-shared-dp-v{version}.yaml
    ```

   Startup example:

    ```ColdFusion
    serviceaccount/k8s-rdma-shared-dp created
    clusterrole.rbac.authorization.k8s.io/pods-rdma-role created
    clusterrolebinding.rbac.authorization.k8s.io/pods-rdma-rolebinding created
    daemonset.apps/rdma-shared-dp-ds created
    ```

5. Run the following command to check whether the component is started successfully.

    ```shell
    kubectl get pod -n kube-system
    ```

   The output is similar to the following. **Running** indicates that the component started successfully.

    ```ColdFusion
    NAME                                             READY   STATUS    RESTARTS    AGE
    ...
    rdma-shared-dp-ds-fd6t8                           1/1    Running      0        74s
    ...
    ```

> [!NOTE]
>
>- After the component is installed, if the Pod status of the component is not
   Running, refer to [Component Pod Status Is Not Running](https://gitcode.com/Ascend/mind-cluster/issues/342) for troubleshooting.
>- After the component is installed, if the Pod status of the component is
   ContainerCreating, refer to [Cluster Scheduling Component Pod Is in ContainerCreating State](https://gitcode.com/Ascend/mind-cluster/issues/343) for troubleshooting.

## Parameter Description<a name="section1851191618363"></a>

**Table 1** K8s RDMA Shared Dev Plugin startup parameters

<a name="table1862682843615"></a>
<table><thead align="left"><tr id="row462602873615"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="p14626028143612"><a name="p14626028143612"></a><a name="p14626028143612"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="p1362692863611"><a name="p1362692863611"></a><a name="p1362692863611"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="p126271528193620"><a name="p126271528193620"></a><a name="p126271528193620"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="p13627192820363"><a name="p13627192820363"></a><a name="p13627192820363"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row162762819363"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p126271328193612"><a name="p126271328193612"></a><a name="p126271328193612"></a>-version / -v</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p2062718289368"><a name="p2062718289368"></a><a name="p2062718289368"></a>Flag bit</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p1962732833612"><a name="p1962732833612"></a><a name="p1962732833612"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p146279281369"><a name="p146279281369"></a><a name="p146279281369"></a>Queries the version number of the current K8s RDMA Shared Dev Plugin. This parameter is a flag bit and does not require a value. Example: ./k8s-rdma-shared-dp -version or ./k8s-rdma-shared-dp -v</p>
</td>
</tr>
<tr id="row15627928153619"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1627328103617"><a name="p1627328103617"></a><a name="p1627328103617"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p56272028193612"><a name="p56272028193612"></a><a name="p56272028193612"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p4627172833617"><a name="p4627172833617"></a><a name="p4627172833617"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p13627628113616"><a name="p13627628113616"></a><a name="p13627628113616"></a>Log level:</p>
<a name="ul262712284363"></a><a name="ul262712284363"></a><ul id="ul262712284363"><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>
</td>
</tr>
<tr id="row126271928143620"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p13627132863615"><a name="p13627132863615"></a><a name="p13627132863615"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p6627828173610"><a name="p6627828173610"></a><a name="p6627828173610"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p062752813613"><a name="p062752813613"></a><a name="p062752813613"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1262712893611"><a name="p1262712893611"></a><a name="p1262712893611"></a>Log backup retention period. The value range is 7 to 700, in days.</p>
</td>
</tr>
<tr id="row862732873610"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p1962772813620"><a name="p1962772813620"></a><a name="p1962772813620"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p162772823620"><a name="p162772823620"></a><a name="p162772823620"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p962817282369"><a name="p962817282369"></a><a name="p962817282369"></a>/var/log/mindx-dl/k8s-rdma-shared-dp/k8s-rdma-shared-dp.log</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p1862816283367"><a name="p1862816283367"></a><a name="p1862816283367"></a>Log file. When a single log file exceeds 20 MB, automatic rotation is triggered. The maximum file size cannot be modified. The naming format of the rotated file is: k8s-rdma-shared-dp-<i>{rotation_time}</i>.log, for example: k8s-rdma-shared-dp-2023-10-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row1862892813365"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p10628202814367"><a name="p10628202814367"></a><a name="p10628202814367"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p4628828173618"><a name="p4628828173618"></a><a name="p4628828173618"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p16628182814364"><a name="p16628182814364"></a><a name="p16628182814364"></a>3</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p10628172873610"><a name="p10628172873610"></a><a name="p10628172873610"></a>Maximum number of log files retained after dumping. The value range is 1 to 180, in units of files.</p>
</td>
</tr>
<tr id="row68317556189"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101521"><a name="p0894319101521"></a><a name="p0894319101521"></a><span id="ph96781327191518"><a name="ph96781327191518"></a><a name="ph96781327191518"></a>-config-file</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151516"><a name="p108941719151516"></a><a name="p108941719151516"></a><span id="ph1899563312155"><a name="ph1899563312155"></a><a name="ph1899563312155"></a>string</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961514"><a name="p19894131961514"></a><a name="p19894131961514"></a><span id="ph67327379153"><a name="ph67327379153"></a><a name="ph67327379153"></a>/k8s-rdma-shared-dev-plugin/config.json</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971512"><a name="p589551971512"></a><a name="p589551971512"></a><span id="ph4556742141518"><a name="ph4556742141518"></a><a name="ph4556742141518"></a>Configuration file path, used to specify the selector configuration for RDMA devices.</span></p>
</td>
</tr>
<tr id="row68317556190"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p0894319101522"><a name="p0894319101522"></a><a name="p0894319101522"></a><span id="ph96781327191520"><a name="ph96781327191520"></a><a name="ph96781327191520"></a>-use-cdi</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p108941719151518"><a name="p108941719151518"></a><a name="p108941719151518"></a><span id="ph1899563312157"><a name="ph1899563312157"></a><a name="ph1899563312157"></a>Flag bit</span></p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p19894131961516"><a name="p19894131961516"></a><a name="p19894131961516"></a><span id="ph67327379155"><a name="ph67327379155"></a><a name="ph67327379155"></a>false</span></p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p589551971514"><a name="p589551971514"></a><a name="p589551971514"></a><span id="ph4556742141520"><a name="ph4556742141520"></a><a name="ph4556742141520"></a>Whether to use CDI (Container Device Interface) mode to register devices with containers. This parameter is a flag bit and does not require a value. Usage example: ./k8s-rdma-shared-dp -use-cdi</span></p>
<div class="note"><span class="notetitle">[!NOTE]</span><div class="notebody"><p>UB-type RDMA devices do not support CDI mode. When a UB device is detected, CDI is automatically disabled.</p></div></div>
</td>
</tr>
<tr id="row10282191492319"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172319"><a name="p8283714172319"></a><a name="p8283714172319"></a>--enable-healthz</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472319"><a name="p828381472319"></a><a name="p828381472319"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482319"><a name="p828341482319"></a><a name="p828341482319"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432319"><a name="p828311432319"></a><a name="p828311432319"></a>Whether to enable the health check service. During K8s deployment, it is enabled (true) through the component YAML configuration.<ul><li>true: enabled.</li><li>false: disabled.</li></ul></p>
</td>
</tr>
<tr id="row10282191492320"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172320"><a name="p8283714172320"></a><a name="p8283714172320"></a>--healthz-address</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472320"><a name="p828381472320"></a><a name="p828381472320"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482320"><a name="p828341482320"></a><a name="p828341482320"></a>11251</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432320"><a name="p828311432320"></a><a name="p828311432320"></a>Port number on which the health check service listens. The value range is 1025 to 65535. During K8s deployment, it is configured as 11257 by the component YAML. If the specified port is occupied, the component fails to start.</p>
</td>
</tr>
<tr id="row10282191492321"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172321"><a name="p8283714172321"></a><a name="p8283714172321"></a>--tls-cert-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472321"><a name="p828381472321"></a><a name="p828381472321"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482321"><a name="p828341482321"></a><a name="p828341482321"></a>""</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432321"><a name="p828311432321"></a><a name="p828311432321"></a>Path to the HTTPS certificate file. If empty, the HTTP protocol is used. It must be configured together with --tls-private-key-file, or both must be empty. For the configuration method and security precautions, see <a href="../../../07_references/04_security_hardening.md">Health Probe Security Hardening</a>.</p>
</td>
</tr>
<tr id="row10282191492322"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172322"><a name="p8283714172322"></a><a name="p8283714172322"></a>--tls-private-key-file</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p828381472322"><a name="p828381472322"></a><a name="p828381472322"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482322"><a name="p828341482322"></a><a name="p828341482322"></a>""</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432322"><a name="p828311432322"></a><a name="p828311432322"></a>Path to the HTTPS private key file. If empty, the HTTP protocol is used. It must be configured together with --tls-cert-file or left empty together with it.</p>
</td>
</tr>
<tr id="row10282191492318"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="p8283714172318"><a name="p8283714172318"></a><a name="p8283714172318"></a>-h or -help</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="p82838147235"><a name="p82838147235"></a><a name="p82838147235"></a>N/A</p>
</td>
<td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="p828341482318"><a name="p828341482318"></a><a name="p828341482318"></a>N/A</p>
</td>
<td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="p828311432320"><a name="p828311432320"></a><a name="p828311432320"></a>Displays help information.</p>
</td>
</tr>
</tbody>
</table>
