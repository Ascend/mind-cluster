# K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_k8s_rdma_shared_dev_plugin"></a>

## Health Probe<a name="ZH-CN_TOPIC_healthz_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin starts an HTTP health probe service within the component, which is used by the K8s livenessProbe mechanism to detect the liveness status of the component.

**Table 1**  Health probe interface

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default Port | 11257 |
| Protocol | HTTP (HTTPS when `--tls-cert-file` and `--tls-private-key-file` are correctly configured) |

**Table 2** Response descriptions

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | Component is running normally | Response body is `ok` |
| 404 Not Found | Request path is not `/` | The probe only responds to the root path|
| 405 Method Not Allowed | Request method is not GET  | K8s livenessProbe uses GET by default, so this is not normally triggered |
| 503 Service Unavailable | Custom health check callback is registered and the check fails | Response body contains specific error information |

**K8s LivenessProbe Configuration Example**:

```yaml
livenessProbe:
   httpGet:
      path: /
      port: 11257
      scheme: HTTP
   initialDelaySeconds: 10
   periodSeconds: 10
   timeoutSeconds: 3
   failureThreshold: 3
```

For details about the probe startup parameters, see [Parameter Description](../05_developer_guide/00_installation_deployment/00_manual_installation/12_k8s_rdma_shared_dev_plugin.md#parameter-description).

## Configuration File Description<a name="ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin configures a selector for RDMA devices through the JSON configuration file specified by the `-config-file` parameter, which is used to discover and report RDMA device resources on a node. The default path of the configuration file is `/k8s-rdma-shared-dev-plugin/config.json`, and it is mounted into the container through a ConfigMap (named `rdma-devices`).

**Configuration File Example**:

```json
{
   "periodicUpdateInterval": 300,
   "faultDetectPeriod": 5,
   "configList": [
      {
         "resourcePrefix": "huawei.com",
         "resourceName": "ub_rdma",
         "rdmaHcaMax": 8,
         "selectors": {
            "buses": ["ub"],
            "vendors": ["0xcc08"],
            "deviceIDs": ["0x8200"]
         }
      }
   ]
}
```

**Table 3** Top-level fields

| Name | Mandatory | Type | Default | Description                                                           |
|------|---------|------|--------|--------------------------------------------------------------|
| periodicUpdateInterval | No | int | 60 | Interval for periodically updating device resources, in seconds. When the value is 0, the periodic device resource update function is disabled. When not set, the default value of 60 seconds is used. The value cannot be less than 0. |
| faultDetectPeriod | No | int | 0 | Fault detection period, in seconds. It takes effect only for UB devices. When not set or the value is less than 1, the fault detection function is disabled.                   |
| configList | Yes | object list | - | Resource configuration list. Each element describes the reporting rules for a group of RDMA devices. The list contains at least 1 element.                       |

For the field description of each configuration object in configList, see [Table 4](#table_config_list_k8s_rdma_shared_dev_plugin).

**Table 4** configList fields<a name="table_config_list_k8s_rdma_shared_dev_plugin"></a>

| Name | Mandatory | Type | Default | Description |
|------|---------|------|--------|------|
| resourceName | Yes | string | - | Device resource name. It must be unique within the resourcePrefix scope and supports only uppercase and lowercase letters, digits, and underscores. The final resource name format reported to K8s is `<resourcePrefix>/<resourceName>`. |
| resourcePrefix | No | string | rdma | Device resource prefix. It must be a valid DNS subdomain (supporting only lowercase letters, digits, hyphens, and dots). |
| rdmaHcaMax | Yes | int | - | Maximum number of RDMA resources that the device plugin can provide. The value cannot be less than 0. |
| selectors | No | object | - | Device selector, used to filter target devices. For details, see [Table 5](#table_selectors_k8s_rdma_shared_dev_plugin) for the field description. |

**Table 5** selectors fields<a name="table_selectors_k8s_rdma_shared_dev_plugin"></a>

| Name | Type | Description |
|------|------|------|
| buses | string list | Device bus type. When the value contains `ub`, UB device mode is enabled; when it is not configured or the value is not `ub`, PCI device mode is enabled. UB devices do not support CDI mode. |
| vendors | string list | Device vendor hexadecimal code, for example, `["0xcc08"]`. |
| deviceIDs | string list | Device model hexadecimal code, for example, `["0x8200"]`. |
| drivers | string list | Device driver name, for example, `["mlx5_core"]`. |
| ifNames | string list | Network interface name, for example, `["ib0"]`. |
| linkTypes | string list | Network interface link type, for example, `["ether"]`, `["infiniband"]`. |

**Selector matching rules**:

- Multiple elements within the same selector field are in a logical OR relationship. For example, `"vendors": ["15b3", "0xcc08"]` means matching devices whose vendor is `15b3` or `0xcc08`.
- Different selector fields are in a logical AND relationship. For example, when both `vendors` and `deviceIDs` are configured, a device must satisfy both the vendor and model conditions to be selected.
- Selector fields that are not configured are ignored and do not participate in filtering.

## ConfigMap Description<a name="ZH-CN_TOPIC_configmap_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin enables the fault detection feature through the `faultDetectPeriod` parameter in the configuration file, which takes effect only for UB-type devices. After the fault detection feature is enabled, the fault information of DPU devices is reported to a Kubernetes ConfigMap. The ConfigMap is located in the `kube-system` namespace, named `dpuinfo-<node-name>` (where `<node-name>` is the node name), with the Label `huawei.com/consumer.clusterd=true`. The ConfigMap uses a forced update policy: an update is triggered when a change in fault information is detected or when more than 5 minutes have elapsed since the last update.

In the ConfigMap, the Key of the Data field is `DpuInfoCfg`, and the Value is the DPU fault information in JSON format. For a detailed description, see [Table 6](#table_dpuconfigmap_k8s_rdma_shared_dev_plugin).

**Table 6** dpuinfo-\<node-name\> ConfigMap
<a name="table_dpuconfigmap_k8s_rdma_shared_dev_plugin"></a>

|Name|Type|Description|
|--|--|--|
|DPUInfo|object|DPU device fault information.|
|-DPUList|list|DPU device list. Each element in the array describes the fault information of a DPU device. For a detailed description, see [Table 7](#table_dpuitem_k8s_rdma_shared_dev_plugin).|
|-NodeEvent|object|Node-level fault events, such as DPU card removal. For a detailed description, see [Table 8](#table_nodeevent_k8s_rdma_shared_dev_plugin).|
|UpdateTime|RFC 3339 timestamp|The update time of the current DPU information, used to identify the latest reporting time of the fault information.|

**Table 7** DPUList element description
<a name="table_dpuitem_k8s_rdma_shared_dev_plugin"></a>

|Field|Type|Description|
|--|--|--|
|HcaName|string|HCA device name, for example, `mlx5_0`.|
|EthName|string|Name of the associated Ethernet network interface.|
|IpAddr|string|IP address of the DPU device.|
|DeviceID|string|Device ID in hexadecimal format.|
|VendorID|string|Vendor ID in hexadecimal format.|
|FaultList|list|List of fault details on this DPU device. Each element in the array describes one piece of fault information. For details, see [Table 9](#table_faultdetail_k8s_rdma_shared_dev_plugin).|

**Table 8** NodeEvent field description
<a name="table_nodeevent_k8s_rdma_shared_dev_plugin"></a>

|Field|Type|Description|
|--|--|--|
|NodeName|string|Node name.|
|FaultList|list|List of node-level fault details. For details, see [Table 9](#table_faultdetail_k8s_rdma_shared_dev_plugin).|

**Table 9** FaultList field description
<a name="table_faultdetail_k8s_rdma_shared_dev_plugin"></a>

| Field       |Type|Description|
|-------------|--|--|
| FaultCode   |string|Fault code, used to identify the fault type.|
| Time        |Unix millisecond timestamp|Time when the fault was first detected.|
| Description |string|Fault description information.|
| FaultLevel  |string|Fault level.|

## Business Pod Usage and Mounted Resource Description<a name="ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin"></a>

When a business Pod uses RDMA shared devices, K8s RDMA Shared Dev Plugin automatically mounts all RDMA devices into the Pod. The following steps are used to verify the resource request and device mounting status of the business Pod.

### Business Pod Resource Request Configuration<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

To use RDMA shared devices, a business Pod must declare resource requests in its Pod configuration. The following configuration example requests one RDMA device resource (for the maximum configurable value, see [Configuration File Description](#configuration-file-description) rdmaHcaMax):

```yaml
apiVersion: v1
kind: Pod
metadata:
   name: mofed-test-pod
spec:
   restartPolicy: OnFailure
   hostNetwork: true
   containers:
      - image: rdma-test:latest
        name: mofed-test-ctr
        imagePullPolicy: IfNotPresent
        securityContext:
           capabilities:
              add: [ "IPC_LOCK" ]
        resources:
           requests:
              huawei.com/ub_rdma: '1'
           limits:
              huawei.com/ub_rdma: '1'
        command:
           - sh
           - -c
           - |
              ls -l /dev/infiniband /sys/class/infiniband
              sleep 1000000
```

> [!NOTICE]
> `hostNetwork` must be set to `true`. Because a business Pod needs to access the host's network namespace to use RDMA devices, hostNetwork mode must be enabled.
> The resource name format is `<resourcePrefix>/<resourceName>`, which must be defined in the K8s RDMA Shared Dev Plugin configuration file (for details, see [Configuration File Description](#configuration-file-description)).

### Checking the Business Pod Status<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

> [!NOTE]
>
> When a business container uses a 1825 DPU, in addition to component mounting, the following are also required:
>
> - Configure `hostNetwork: true`
> - Configure the user-mode driver. Choose either of the following two methods:
>   1. Install the OFED driver for the 1825 DPU in the image.
>   2. Mount the OFED driver for the 1825 DPU from the host after the container starts.

1. **Check the Pod status.**

   Run the following command to check whether the business Pod is created successfully:

    ```shell
    kubectl get pod rdma-app -o wide
    ```

   The following is an example of the output. **Running** indicates that the Pod is created successfully:

    ```ColdFusion
    NAME       READY   STATUS    RESTARTS   AGE   IP            NODE
    rdma-app   1/1     Running   0          10s   10.244.1.*   compute-node-1
    ```

### RDMA Device Verification in a Pod<a name="ZH-CN_TOPIC_biz_pod_rdma_verify"></a>

After the service Pod is created successfully, you can verify whether the RDMA devices are correctly mounted by the component through the following steps:

1. **Enter the Pod.**

    ```shell
    kubectl exec -it rdma-app -- /bin/bash
    ```

2. **Check the RDMA device nodes.**

    ```shell
    ls -la /dev/infiniband/
    ```

   Under normal circumstances, device node files such as `uverbs0` and `uverbs1` should be displayed. If the device nodes are empty or do not exist, the RDMA devices are not correctly mounted.

After the service Pod is created successfully, you also need to check the mounting status of the RDMA NIC devices and the corresponding network interfaces:

1. **Check Infiniband device information.**

    ```shell
    ls -la /sys/class/infiniband/
    ```

   Under normal circumstances, the RDMA NIC devices on the current node should be displayed, such as `hrn5_0` and `hrn5_1`.

2. **Check network interface information.**

     ```shell
     ls -la /sys/class/net/
     ```

   Under normal circumstances, the network interface devices on the node should be displayed, including the network interfaces corresponding to the RDMA NICs (such as `ens***`). If the network does not appear in the Pod, check whether hostNetwork is set to true.
