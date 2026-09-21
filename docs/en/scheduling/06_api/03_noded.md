# NodeD<a name="ZH-CN_TOPIC_0000002511346795"></a>

## Node Resources<a name="ZH-CN_TOPIC_0000002511426761"></a>

**mindx-dl-nodeinfo-_<nodename\>_<a name="section1119586114219"></a>**

When a fault occurs on a node, NodeD will create `node-info-cm` to report the fault.

**Table 1**  mindx-dl-nodeinfo-_<nodename\>_

|Parameter Name|Description|
|--|--|
|NodeInfo|Fault information at the node level.|
|FaultDevList|List of faulty devices on the node.|
|- DeviceType|Type of the faulty device.|
|- DeviceId|ID of the faulty device.|
|- FaultCode|Fault code, a string composed of English letters and numbers, representing the fault code in hexadecimal.|
|- FaultLevel|Fault handling level.<li>NotHandleFault: No Action Required.</li><li>PreSeparateFault: If a job exists on this node, ignore; no jobs will be scheduled to this node during subsequent scheduling.</li><li>SeparateFault: Job rescheduling.</li>|
|NodeStatus|Node health status, determined by the device with the most severe fault handling level on this node.<li>Healthy: The fault handling level on this node exists and does not exceed NotHandleFault. The node is healthy and can perform training normally.</li><li>PreSeparate: The fault handling level on this node exists and does not exceed PreSeparateFault. The node is in a pre-isolated state, which may temporarily have no impact on jobs. After the job is affected and exits, no further jobs will be scheduled to this node.</li><li>UnHealthy: The fault handling level on this node includes SeparateFault. The node is a Faulty Node, which will affect training jobs. Jobs will be immediately migrated away from this node.</li>|
|CheckCode|Check code.|

## Customizing Node Faults<a name="ZH-CN_TOPIC_0000002479386802"></a>

The configuration file `NodeDConfiguration.json` of the NodeD component is a system configuration file. Do not modify it arbitrarily unless you have special requirements. If you need to modify the fault level of a fault code, you can do so through the `mindx-dl-node-fault-config` file created from `NodeDConfiguration.json`. For details, see [(Optional) Configuring Node Hardware Fault Levels](../04_usage/04_resumable_training/03_configuration/01_configuring_fault_detection_levels.md#optional-configuring-node-hardware-fault-levels).

**Table 1**  Fault description

|Fault Level|Fault Handling Policy|Description|
|--|--|--|
|NotHandleFault|No action required.|Have no impact on jobs.|
|PreSeparateFault|If a job exists on this node, the fault is ignored; no jobs will be scheduled to this node during subsequent scheduling.|May cause jobs to be affected.|
|SeparateFault|Job rescheduling|Jobs will definitely be affected.|

> [!NOTE]
> The fault levels, from lowest to highest, are NotHandleFault < PreSeparateFault < SeparateFault.

**Table 2**  Node status description

|Node Status|Highest Fault Level|Fault Handling Policy|Description|
|--|--|--|--|
|Healthy|NotHandleFault|No action required.|The node is healthy and can perform training normally.|
|PreSeparate|PreSeparateFault|If a job exists on this node, the fault is ignored; no jobs will be scheduled to this node during subsequent scheduling.|The node is in a pre-isolated state. It may not affect jobs temporarily. After a job is affected and exits, no further jobs will be scheduled to this node.|
|UnHealthy|SeparateFault|Job rescheduling|The node is a faulty node that will affect training jobs. Jobs are immediately migrated away from this node.|

> [!NOTE]
>
>- The current health status of a node is primarily determined by the highest fault level of its hardware faults.
>- `Healthy`, `PreSeparate`, and `UnHealthy` are node statuses defined by MindCluster, primarily used for subsequent job scheduling and handling.
>- If a job on a `PreSeparate` node exits abnormally and requires resumable training, the unconditional retry function must be enabled.

## Health Probe<a name="ZH-CN_TOPIC_healthz_noded"></a>

NodeD starts an HTTP health probe service within the component, which is used by the K8s livenessProbe mechanism to detect the component's liveness status.

**Table 3** Health probe interface

| Item | Description |
|------|------|
| Path | `/` |
| Method | GET |
| Default Port | 11255 |
| Protocol | HTTP (HTTPS when `--tls-cert-file` and `--tls-private-key-file` are correctly configured) |

**Table 4** Response descriptions

| Status Code | Trigger Condition | Description |
|--------|---------|------|
| 200 OK | Component is running normally | Response body is `ok` |
| 404 Not Found | Request path is not `/` | The probe only responds to the root path|
| 405 Method Not Allowed | Request method is not GET  | K8s livenessProbe uses GET by default, so this is not normally triggered |
| 503 Service Unavailable | Custom health check callback is registered and the check fails | Response body contains specific error information |

**K8s LivenessProbe Configuration Example:**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11255
    scheme: HTTP
  initialDelaySeconds: 20
  periodSeconds: 15
  timeoutSeconds: 5
  failureThreshold: 3
```

For detailed probe parameter descriptions, see [Parameter Description](../05_developer_guide/00_installation_deployment/00_manual_installation/09_noded.md#parameter-description).
