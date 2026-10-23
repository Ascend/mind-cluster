# Ascend Dynamic Resource Allocation<a name="ZH-CN_TOPIC_0000002511426333"></a>

- Users who apply for NPU resources through the K8s Dynamic Resource Allocation (DRA) mechanism must install Ascend Dynamic Resource Allocation (hereinafter referred to as the DRA driver) on compute nodes. The DRA driver is deployed as a kubelet plugin. It is responsible for reporting the NPU resources on a node to K8s in the form of ResourceSlices, and completing device allocation and injection when a service container applies for NPU resources through a ResourceClaim.
- Users who do not use dynamic resource allocation do not need to install this component. Skip this chapter directly.

## Constraints<a name="section1362795652417"></a>

Before installing the DRA driver, familiarize yourself with the related constraints in advance. For details, see [Table 1](#table124141536dra01).

**Table 1** Constraint description

<a name="table124141536dra01"></a>
<table><thead align="left"><tr id="row0414dra240011"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.3.1.1"><p id="p0414dra240011"><a name="p0414dra240011"></a><a name="p0414dra240011"></a>Constraint Scenario</p>
</th>
<th class="cellrowborder" valign="top" width="75%" id="mcps1.2.3.1.2"><p id="p0414dra240012"><a name="p0414dra240012"></a><a name="p0414dra240012"></a>Constraint Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row0414dra240013"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240013"><a name="p0414dra240013"></a><a name="p0414dra240013"></a>K8s version</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240014"><a name="p0414dra240014"></a><a name="p0414dra240014"></a>The component YAML file uses the resource.k8s.io/v1 and admissionregistration.k8s.io/v1 APIs, which require K8s 1.34.x to 1.36.x. The component cannot be installed on a K8s cluster of an earlier version.</p>
</td>
</tr>
<tr id="row0414dra240015"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240015"><a name="p0414dra240015"></a><a name="p0414dra240015"></a>Container runtime</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240016"><a name="p0414dra240016"></a><a name="p0414dra240016"></a>The component injects NPU devices into service containers through CDI (Container Device Interface), which requires the container runtime on the compute node to support CDI, for example, containerd 1.7.0 or later.</p>
</td>
</tr>
<tr id="row0414dra240017"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240017"><a name="p0414dra240017"></a><a name="p0414dra240017"></a>NPU driver</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240018"><a name="p0414dra240018"></a><a name="p0414dra240018"></a>The component calls related interfaces of the NPU driver. To upgrade the driver, stop service jobs first, and then stop the component container service.</p>
</td>
</tr>
<tr id="row0414dra240019"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240019"><a name="p0414dra240019"></a><a name="p0414dra240019"></a>Deployment form</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240020"><a name="p0414dra240020"></a><a name="p0414dra240020"></a>The component runs as a privileged container under the root user and mounts the NPU driver directory, kubelet plugin directory, and device directory of the host. Do not modify the DaemonSet name or its mount configuration in the YAML file. Otherwise, the component cannot work properly.</p>
</td>
</tr>
<tr id="row0414dra240021"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240021"><a name="p0414dra240021"></a><a name="p0414dra240021"></a>Supported chip</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240022"><a name="p0414dra240022"></a><a name="p0414dra240022"></a>Ascend910 and Ascend950 NPUs are supported.</p>
</td>
</tr>
<tr id="row0414dra240023"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.3.1.1 "><p id="p0414dra240023"><a name="p0414dra240023"></a><a name="p0414dra240023"></a>Health check port</p>
</td>
<td class="cellrowborder" valign="top" width="75%" headers="mcps1.2.3.1.2 "><p id="p0414dra240024"><a name="p0414dra240024"></a><a name="p0414dra240024"></a>The component enables the health check service by default and occupies port 11258. Ensure that this port is not occupied on the compute node. If the port conflicts, modify the "--healthz-address" parameter in the YAML file as well as the port number in livenessProbe.</p>
</td>
</tr>
</tbody>
</table>

## Procedure<a name="section83111543151613"></a>

1. Log in to each compute node as the `root` user and run the following command to check whether the image and version number are correct.

    ```shell
    docker images | grep ascend-dra
    ```

    Command output:

    ```ColdFusion
    ascend-dra                         v26.2.0              29eec79eb693        About an hour ago   105MB
    ```

    - If yes, go to [step 2](#li2093514717dra01).
    - If no, see [Preparing an Image](./01_preparing_for_installation.md#preparing-an-image) to build and distribute the image.

2. <a name="li2093514717dra01"></a>Copy the YAML file (ascend-dra-driver-v<i>{version}</i>.yaml) in the Ascend Dynamic Resource Allocation software package decompression directory to any directory on the K8s management node.

3. If you do not need to modify the component startup parameters, skip this step. Otherwise, modify the DRA driver startup parameters in the YAML file as required. The startup parameters are located in the "command" line of the DaemonSet, as shown below. For the startup parameters, see [Table 2](#table20426115dra02). You can also run <b>./ascend-dra -h</b> to view the parameter description.

    ```Yaml
    ...
          containers:
          - name: plugin
            securityContext:
              privileged: true
            image: ascend-dra:v26.2.0
            command: ["/bin/bash", "-c", "exec ascend-dra --enable-healthz --healthz-address=11258"]
    ...
    ```

4. In the path where the YAML file is stored on the management node, run the following command to start the DRA driver.

    ```shell
    kubectl apply -f ascend-dra-driver-v{version}.yaml
    ```

    Command output:

    ```ColdFusion
    serviceaccount/ascend-dra-driver-service-account created
    clusterrole.rbac.authorization.k8s.io/ascend-dra-driver-role created
    clusterrolebinding.rbac.authorization.k8s.io/ascend-dra-driver-role-binding created
    daemonset.apps/ascend-dra-driver-kubeletplugin created
    deviceclass.resource.k8s.io/npu.huawei.com created
    validatingadmissionpolicy.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
    validatingadmissionpolicybinding.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
    ```

5. Run the following command to check whether the component is started successfully.

    ```shell
    kubectl get pod -n kube-system
    ```

    The following is the output example. **Running** indicates that the component is started successfully.

    ```ColdFusion
    NAME                                        READY   STATUS    RESTARTS   AGE
    ascend-dra-driver-kubeletplugin-5m2xv       1/1     Running   0          74s
    ```

6. Run the following command to verify whether NPU resources are reported successfully.

    ```shell
    kubectl get deviceclass npu.huawei.com
    kubectl get resourceslices
    ```

    The following is the output example. If a DeviceClass named "npu.huawei.com" can be queried and each compute node has a corresponding ResourceSlice, resources are reported successfully.

    ```ColdFusion
    NAME            AGE
    npu.huawei.com  2m

    NAME                                  POOLNAME     DEVICECOUNT   AGE
    npu.huawei.com-node1-1a2b3c           node1        8             2m
    ```

>[!NOTE]
>
>- If the Pod status of the component is not `Running` after the component is installed, handle it by referring to [Component Pod status is not Running](https://gitcode.com/Ascend/mind-cluster/issues/342).
>- If the Pod status of the component is `ContainerCreating` after the component is installed, handle it by referring to [Cluster scheduling component Pod is in ContainerCreating state](https://gitcode.com/Ascend/mind-cluster/issues/343).
>- If the component is started successfully but the Pod corresponding to the component cannot be found, refer to [The component startup YAML is executed successfully but the Pod corresponding to the component cannot be found](https://gitcode.com/Ascend/mind-cluster/issues/345).

## Parameter Description<a name="section2042611570393"></a>

**Table 2** DRA driver startup parameters

<a name="table20426115dra02"></a>
<table><thead align="left"><tr id="row0414dra240031"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.5.1.1"><p id="p0414dra240031"><a name="p0414dra240031"></a><a name="p0414dra240031"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="10%" id="mcps1.2.5.1.2"><p id="p0414dra240032"><a name="p0414dra240032"></a><a name="p0414dra240032"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.3"><p id="p0414dra240033"><a name="p0414dra240033"></a><a name="p0414dra240033"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="35%" id="mcps1.2.5.1.4"><p id="p0414dra240034"><a name="p0414dra240034"></a><a name="p0414dra240034"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row0414dra240039"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240039"><a name="p0414dra240039"></a><a name="p0414dra240039"></a>-cdi-root</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240040"><a name="p0414dra240040"></a><a name="p0414dra240040"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240041"><a name="p0414dra240041"></a><a name="p0414dra240041"></a>/var/run/cdi</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240042"><a name="p0414dra240042"></a><a name="p0414dra240042"></a>Directory where CDI files are generated. After the component allocates an NPU device, a CDI description file is generated in this directory, and kubelet injects the NPU device into the service container based on this file.</p>
</td>
</tr>
<tr id="row0414dra240043"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240043"><a name="p0414dra240043"></a><a name="p0414dra240043"></a>-kubelet-registrar-directory-path</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240044"><a name="p0414dra240044"></a><a name="p0414dra240044"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240045"><a name="p0414dra240045"></a><a name="p0414dra240045"></a>/var/lib/kubelet/plugins_registry</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240046"><a name="p0414dra240046"></a><a name="p0414dra240046"></a>kubelet plugin registration directory.</p>
</td>
</tr>
<tr id="row0414dra240047"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240047"><a name="p0414dra240047"></a><a name="p0414dra240047"></a>-kubelet-plugins-directory-path</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240048"><a name="p0414dra240048"></a><a name="p0414dra240048"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240049"><a name="p0414dra240049"></a><a name="p0414dra240049"></a>/var/lib/kubelet/plugins</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240050"><a name="p0414dra240050"></a><a name="p0414dra240050"></a>kubelet plugin data directory.</p>
</td>
</tr>
<tr id="row0414dra240051"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240051"><a name="p0414dra240051"></a><a name="p0414dra240051"></a>-deviceResetTimeout</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240052"><a name="p0414dra240052"></a><a name="p0414dra240052"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240053"><a name="p0414dra240053"></a><a name="p0414dra240053"></a>600</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240054"><a name="p0414dra240054"></a><a name="p0414dra240054"></a>When the component is started, if the number of chips is insufficient, the maximum duration for waiting for the driver to report all chips. The unit is second, and the value range is 10 to 600.</p>
</td>
</tr>
<tr id="row0414dra240055"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240055"><a name="p0414dra240055"></a><a name="p0414dra240055"></a>-kubeconfig</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240056"><a name="p0414dra240056"></a><a name="p0414dra240056"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240057"><a name="p0414dra240057"></a><a name="p0414dra240057"></a>N/A</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240058"><a name="p0414dra240058"></a><a name="p0414dra240058"></a>Path of the kubeconfig file. It does not need to be configured for in-cluster deployment, and the component uses the in-cluster authentication mode by default.</p>
</td>
</tr>
<tr id="row0414dra240059"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240059"><a name="p0414dra240059"></a><a name="p0414dra240059"></a>-kube-api-qps</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240060"><a name="p0414dra240060"></a><a name="p0414dra240060"></a>float</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240061"><a name="p0414dra240061"></a><a name="p0414dra240061"></a>5</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240062"><a name="p0414dra240062"></a><a name="p0414dra240062"></a>QPS for communication with the K8s APIServer. When the request pressure on the K8s APIServer increases, you can decrease this value based on actual conditions.</p>
</td>
</tr>
<tr id="row0414dra240063"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240063"><a name="p0414dra240063"></a><a name="p0414dra240063"></a>-kube-api-burst</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240064"><a name="p0414dra240064"></a><a name="p0414dra240064"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240065"><a name="p0414dra240065"></a><a name="p0414dra240065"></a>10</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240066"><a name="p0414dra240066"></a><a name="p0414dra240066"></a>Burst for communication with the K8s APIServer.</p>
</td>
</tr>
<tr id="row0414dra240067"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240067"><a name="p0414dra240067"></a><a name="p0414dra240067"></a>-logFile</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240068"><a name="p0414dra240068"></a><a name="p0414dra240068"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240069"><a name="p0414dra240069"></a><a name="p0414dra240069"></a>/var/log/mindx-dl/ascend-dra/ascend-dra.log</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240070"><a name="p0414dra240070"></a><a name="p0414dra240070"></a>Log file path. When a single log file exceeds 20 MB, automatic dump is triggered. The upper limit of the file size cannot be modified.</p>
</td>
</tr>
<tr id="row0414dra240071"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240071"><a name="p0414dra240071"></a><a name="p0414dra240071"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240072"><a name="p0414dra240072"></a><a name="p0414dra240072"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240073"><a name="p0414dra240073"></a><a name="p0414dra240073"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240074"><a name="p0414dra240074"></a><a name="p0414dra240074"></a>Log level:</p>
<a name="ul0414dra240071"></a><a name="ul0414dra240071"></a><ul id="ul0414dra240071"><li>Value -1: debug</li><li>Value 0: info</li><li>Value 1: warning</li><li>Value 2: error</li><li>Value 3: critical</li></ul>
</td>
</tr>
<tr id="row0414dra240075"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240075"><a name="p0414dra240075"></a><a name="p0414dra240075"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240076"><a name="p0414dra240076"></a><a name="p0414dra240076"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240077"><a name="p0414dra240077"></a><a name="p0414dra240077"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240078"><a name="p0414dra240078"></a><a name="p0414dra240078"></a>Log backup retention. The value range is 7 to 700, and the unit is day.</p>
</td>
</tr>
<tr id="row0414dra240079"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240079"><a name="p0414dra240079"></a><a name="p0414dra240079"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240080"><a name="p0414dra240080"></a><a name="p0414dra240080"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240081"><a name="p0414dra240081"></a><a name="p0414dra240081"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240082"><a name="p0414dra240082"></a><a name="p0414dra240082"></a>Upper limit of the number of log files retained after dumping. The value range is 1 to 180, and the unit is file.</p>
</td>
</tr>
<tr id="row0414dra240083"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240083"><a name="p0414dra240083"></a><a name="p0414dra240083"></a>-enable-healthz</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240084"><a name="p0414dra240084"></a><a name="p0414dra240084"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240085"><a name="p0414dra240085"></a><a name="p0414dra240085"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240086"><a name="p0414dra240086"></a><a name="p0414dra240086"></a>Whether to enable the health check service. It is enabled by the component YAML configuration for K8s deployment (true).</p>
<a name="ul0414dra240083"></a><a name="ul0414dra240083"></a><ul id="ul0414dra240083"><li>true: enabled.</li><li>false: disabled.</li></ul>
</td>
</tr>
<tr id="row0414dra240087"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240087"><a name="p0414dra240087"></a><a name="p0414dra240087"></a>-healthz-address</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240088"><a name="p0414dra240088"></a><a name="p0414dra240088"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240089"><a name="p0414dra240089"></a><a name="p0414dra240089"></a>11258</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240090"><a name="p0414dra240090"></a><a name="p0414dra240090"></a>Listening port of the health check service. When this parameter is modified, the port number configured in livenessProbe in the YAML file must be modified accordingly.</p>
</td>
</tr>
<tr id="row0414dra240091"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240091"><a name="p0414dra240091"></a><a name="p0414dra240091"></a>-tls-cert-file</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240092"><a name="p0414dra240092"></a><a name="p0414dra240092"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240093"><a name="p0414dra240093"></a><a name="p0414dra240093"></a>N/A</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240094"><a name="p0414dra240094"></a><a name="p0414dra240094"></a>Path of the TLS certificate file used when the health check service enables HTTPS. If it is not configured, the health check service uses HTTP.</p>
</td>
</tr>
<tr id="row0414dra240095"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240095"><a name="p0414dra240095"></a><a name="p0414dra240095"></a>-tls-private-key-file</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240096"><a name="p0414dra240096"></a><a name="p0414dra240096"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240097"><a name="p0414dra240097"></a><a name="p0414dra240097"></a>N/A</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240098"><a name="p0414dra240098"></a><a name="p0414dra240098"></a>Path of the TLS private key file used when the health check service enables HTTPS. It must be configured together with -tls-cert-file.</p>
</td>
</tr>
<tr id="row0414dra240099"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.5.1.1 "><p id="p0414dra240099"><a name="p0414dra240099"></a><a name="p0414dra240099"></a>-version / -v</p>
</td>
<td class="cellrowborder" valign="top" width="10%" headers="mcps1.2.5.1.2 "><p id="p0414dra240100"><a name="p0414dra240100"></a><a name="p0414dra240100"></a>bool</p>
</td>
<td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.3 "><p id="p0414dra240101"><a name="p0414dra240101"></a><a name="p0414dra240101"></a>false</p>
</td>
<td class="cellrowborder" valign="top" width="35%" headers="mcps1.2.5.1.4 "><p id="p0414dra240102"><a name="p0414dra240102"></a><a name="p0414dra240102"></a>Queries details such as the version number and commit ID of the current Ascend DRA. This parameter is a flag and does not need a value. Usage example: ./ascend-dra -version or ./ascend-dra -v</p>
</td>
</tr>
</tbody>
</table>
