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

For details about the probe startup parameters, see [Parameter Description](../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md#parameter-description).

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
| configList | Yes | object list | - | Resource configuration list. Each element describes the reporting rules for a group of RDMA devices. The list contains at least 1 element.<p>For the field description of each configuration object in configList, see [Table 4](#table_config_list_k8s_rdma_shared_dev_plugin). </p>|

**Table 4** configList fields<a name="table_config_list_k8s_rdma_shared_dev_plugin"></a>

| Name | Mandatory | Type | Default | Description |
|------|---------|------|--------|------|
| resourceName | Yes | string | - | Device resource name. It must be unique within the resourcePrefix scope and supports only uppercase and lowercase letters, digits, and underscores. The final resource name format reported to K8s is `<resourcePrefix>/<resourceName>`. |
| resourcePrefix | No | string | rdma | Device resource prefix. It must be a valid DNS subdomain (supporting only lowercase letters, digits, hyphens, and dots). |
| rdmaHcaMax | Yes | int | - | Maximum number of RDMA resources that the device plugin can provide. The value cannot be less than 0. In UB scenarios, this field does not take effect. The number of RDMA resources that can be provided is determined by the number of UB devices actually discovered.|
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
|UpdateTime|Unix timestamp in milliseconds|The update time of the current DPU information, used to identify the latest reporting time of the fault information.|

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
|AffectedNPU|list|List of NPU device IDs affected by the fault of this device. Only NPU device IDs that use the current UB NIC as their primary NIC are included.|

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

## Configuration in Container Scenarios

When a service needs to use a UB NIC inside a container, the RoCE device and driver must be mounted; otherwise, the UB NIC cannot be initialized in the container. For an example of the resources required in a container in the UB NIC scenario, see [Table 10](#table1011k8srdmashareddevplugin).

**Table 10** Resources required to use a UB NIC in a container
<a name="table1011k8srdmashareddevplugin"></a>

| Resource Path | Type |
|---------|------|
| /dev/infiniband/rdma_cm | Character device |
| /dev/infiniband/uverbs{N} | Character device (N indicates the device number) |
| /etc/libibverbs.d/hrn5.driver | Driver configuration file |
| /usr/lib64/libhrn5-rdmav34.so | Dynamic library |
| /usr/lib64/libhrn5-rdmav.so | Dynamic library |
| /usr/lib64/libibv_extend.so | Dynamic library |
| /usr/lib64/libibv_extend.so.4 | Dynamic library |
| /usr/lib64/libibv_extend.so.4.2.0 | Dynamic library |

How to obtain each resource:

- **Character devices**: No mounting configuration is required. The DaemonSet of the component has mounted the host `/dev` directory and discovers UB devices through the host sysfs. After a service Pod requests the `<resourcePrefix>/<resourceName>` resource, kubelet automatically injects the `/dev/infiniband/rdma_cm` and the corresponding `uverbs{N}` character devices when creating the container, based on the device information returned by the component.
- **Driver configuration files and dynamic libraries**: They are not automatically mounted by the component. They must be included in the service image, or mounted from the host through hostPath. You are advised to install them in advance in the service image. If a non-UB NIC is used, consult the NIC provider to confirm the files to be mounted.
- **rdma-core**: An open-source RDMA user-mode component that provides basic libraries such as libibverbs and librdmacm, and tools such as `ibv_devinfo`. The dynamic libraries provided by the UB NIC driver in Table 10 (`libhrn5-*` and `libibv_extend*`) are vendor extensions of libibverbs and must run on top of libibverbs. Ensure that they are installed in the training image, or directly use a training image with rdma-core preinstalled.

In the container scenario, you also need to configure the `NET_ADMIN`, `SYS_ADMIN`, and `IPC_LOCK` permissions for the service container. For the configuration method, see the YAML example below.

The following is a configuration example of mounting driver configuration files and dynamic libraries through `hostPath` (for the other Pod configurations, see [Business Pod Usage and Mounted Resource Description](../04_usage/03_basic_scheduling/07_dpu_scheduling.md#ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin)):

```yaml
apiVersion: v1
kind: AscendJob
metadata:
   name: rdma-test
spec:
   restartPolicy: OnFailure
   hostNetwork: true
   containers:
      - image: rdma-test:latest
        name: ub-rdma-test-ctr
        imagePullPolicy: IfNotPresent
        securityContext:
           capabilities:
              add: [ "NET_ADMIN", "SYS_ADMIN", "IPC_LOCK" ]
        resources:
           requests:
              huawei.com/ub_rdma: '1'
           limits:
              huawei.com/ub_rdma: '1'
        volumeMounts:
           - name: hrn5-driver-config
             mountPath: /etc/libibverbs.d/hrn5.driver
           - name: hrn5-rdmav34
             mountPath: /usr/lib64/libhrn5-rdmav34.so
           - name: hrn5-rdmav
             mountPath: /usr/lib64/libhrn5-rdmav.so
           - name: ibv-extend
             mountPath: /usr/lib64/libibv_extend.so
           - name: ibv-extend-4
             mountPath: /usr/lib64/libibv_extend.so.4
           - name: ibv-extend-4-2-0
             mountPath: /usr/lib64/libibv_extend.so.4.2.0
   volumes:
      - name: hrn5-driver-config
        hostPath:
           path: /etc/libibverbs.d/hrn5.driver
           type: File
      - name: hrn5-rdmav34
        hostPath:
           path: /usr/lib64/libhrn5-rdmav34.so
           type: File
      - name: hrn5-rdmav
        hostPath:
           path: /usr/lib64/libhrn5-rdmav.so
           type: File
      - name: ibv-extend
        hostPath:
           path: /usr/lib64/libibv_extend.so
           type: File
      - name: ibv-extend-4
        hostPath:
           path: /usr/lib64/libibv_extend.so.4
           type: File
      - name: ibv-extend-4-2-0
        hostPath:
           path: /usr/lib64/libibv_extend.so.4.2.0
           type: File
```

> [!NOTE]
>
> - When mounting dynamic libraries, the mount path in the container must be the same as the host path (for example, `/usr/lib64`); otherwise, the dynamic libraries may fail to find each other due to dependencies.
> - The file paths and version numbers in Table 10 and the preceding example (for example, `libibv_extend.so.4.2.0`) are examples only in the UB NIC scenario. As the driver version is upgraded, the minor version number of the dynamic library may change. Use the driver files actually installed on the host as the basis.
