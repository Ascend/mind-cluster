# Ascend Dynamic Resource Allocation<a name="ZH-CN_TOPIC_0000002511426dra"></a>

The Ascend Dynamic Resource Allocation (Ascend DRA) component developed  is based on the K8s Dynamic Resource Allocation mechanism. It runs on each compute node as a kubelet plugin and is responsible for enumerating the NPU devices on the node, reporting ResourceSlices to the API Server, and completing the Prepare/Unprepare of ResourceClaims driven by kubelet. This document describes its health check interfaces and the reported ResourceSlice format.

## Health Check<a name="section_dra_healthz"></a>

Ascend DRA starts a built-in HTTP health probe service, which is used by the K8s livenessProbe mechanism to detect the component liveness status. The probe service is decoupled from the service logic, uses a dedicated port, and supports both HTTP and HTTPS.

**Table 1** Health probe interfaces

<a name="table_dra_healthz_interface"></a>

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default port | 11258 |
| Protocol | HTTP (HTTPS when the `--tls-cert-file` and `--tls-private-key-file` parameters are correctly configured) |
| Request throttling | 1 QPS, with a burst limit of 5; 429 is returned when the limit is exceeded |

**Table 2**  Response description

<a name="table_dra_healthz_response"></a>

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | The component is running properly | The response body is `ok` |
| 404 Not Found | The request path is not `/` | The probe responds only to the root path |
| 405 Method Not Allowed | The request method is not GET | K8s livenessProbe uses GET by default, so this is not triggered normally |
| 429 Too Many Requests | The request frequency exceeds the throttling threshold | For the throttling configuration, see [Table 1](#table_dra_healthz_interface) |
| 503 Service Unavailable | A custom health check callback is registered and the check fails | The response body contains the specific error information in the format `unhealthy: <error>` |

Ascend DRA has a built-in health check callback `draHealthChecker`, which returns "healthy" as long as the plugin process is reachable. Therefore, the probe always returns 200 under normal conditions.

**Table 3**  Health probe startup parameters

<a name="table_dra_healthz_flags"></a>

| Parameter | Meaning | Default Value |
|------|------|--------|
| --enable-healthz | Whether to enable the health check service | false |
| --healthz-address | Listening port of the health check service, value range [1025, 65535] | 11258 |
| --tls-cert-file | Path of the HTTPS certificate file. It must be configured together with --tls-private-key-file, or both must be left empty | Empty |
| --tls-private-key-file | Path of the HTTPS private key file. It must be configured together with --tls-cert-file, or both must be left empty | Empty |

> [!NOTE]
> `--enable-healthz` is set to `false` by default, and it must be explicitly enabled in the image startup command. The recommended configuration is `--enable-healthz --healthz-address=11258`.

**K8s livenessProbe configuration example:**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11258
    scheme: HTTP
  failureThreshold: 3
  periodSeconds: 10
```

## ResourceSlice Reporting Format<a name="section_dra_resourceslice"></a>

When Ascend DRA is started, it enumerates all NPU devices on the node and assembles `DriverResources` as required by the K8s resource API (`resource.k8s.io/v1`), which is published as a ResourceSlice object in the cluster through the kubelet plugin Helper. Each node corresponds to one Pool, and the Pool name is the `NodeName` of the node. Each Pool contains one Slice, and the Slice contains all NPU devices on the node.

**Table 4** ResourceSlice object description

<a name="table_dra_resourceslice"></a>

| Name | Meaning | Description |
|--|--|--|
| spec.nodeName | Name of the node that reports this ResourceSlice | Obtained from the DRA startup parameter --node-name or the environment variable NODE_NAME. |
| spec.driver | DRA driver name | Fixed value `npu.huawei.com`. |
| spec.pool.name | Resource pool name | Equal to the NodeName of the node. Each node exclusively occupies one Pool. |
| spec.pool.generation | Resource pool generation | Maintained by the kubelet Helper and incremented on each update. |
| spec.slices[].devices[] | Device list | All NPU devices enumerated on the node, with one Device object per card. |
| spec.slices[].devices[].name | Device name | Format: `npu-<phyID>`, all lowercase. For example, `npu-0` and `npu-12`. |
| spec.slices[].devices[].attributes | Device attributes | Set of device attributes. For details, see [Table 5](#table_dra_device_attributes). |

**Table 5** Device.attributes description

<a name="table_dra_device_attributes"></a>

| Attribute Name | Type | Meaning | Description |
|--|--|--|--|
| type | string | Device type | Fixed value `npu`. |
| physicId | int | Device physical ID | Obtained from the return value of GetPhysicIDFromLogicID of the device manager, consistent with the DeviceName suffix. |
| chipName | string | Chip name | Obtained from the return value of GetChipInfo of the device manager. If the acquisition fails, the value is an empty string, but the attribute is still reported. |

> [!NOTE]
> Currently, Ascend910 and Ascend950 generations report the same set of attributes, namely `type`, `physicId`, and `chipName`. If attributes need to be extended for a new chip generation, you only need to return them in the `DeviceAttributes` method of the corresponding generation, which does not affect the reporting link.

**ResourceSlice example:**

The following example shows a ResourceSlice reported by a node with two NPUs (some fields automatically filled by the cluster are omitted):

```yaml
apiVersion: resource.k8s.io/v1
kind: ResourceSlice
metadata:
  name: node-alpha-<suffix>
spec:
  nodeName: node-alpha
  driver: npu.huawei.com
  pool:
    name: node-alpha
    generation: 1
  slices:
  - devices:
    - name: npu-0
      attributes:
        type: "npu"
        physicId: 0
        chipName: "Ascend910B..."
    - name: npu-1
      attributes:
        type: "npu"
        physicId: 1
        chipName: "Ascend910B..."
```

> [!NOTE]
> A ValidatingAdmissionPolicy is also deployed in the cluster to restrict each DRA Pod to modifying only the ResourceSlice corresponding to its own node, preventing cross-node miswrites. For details, see `resourceslices-policy-ascend-dra-driver` in the deployment file `ascend-dra-driver.yaml`.
