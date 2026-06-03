# NodeD<a name="ZH-CN_TOPIC_0000002511346795"></a>

## 节点资源<a name="ZH-CN_TOPIC_0000002511426761"></a>

**mindx-dl-nodeinfo-_<nodename\>_<a name="section1119586114219"></a>**

当节点上存在节点故障时，NodeD将创建node-info-cm，进行故障上报。

**表 1**  mindx-dl-nodeinfo-_<nodename\>_

|参数名|描述|
|--|--|
|NodeInfo|节点维度的故障信息。|
|FaultDevList|节点故障设备列表。|
|- DeviceType|故障设备类型。|
|- DeviceId|故障设备ID。|
|- FaultCode|故障码，由英文和数组拼接而成的字符串，字符串表示故障码的十六进制。|
|- FaultLevel|故障处理等级。<li>NotHandleFault：无需处理。</li><li>PreSeparateFault：该节点上有任务则不处理，后续调度时不调度任务到该节点。</li><li>SeparateFault：任务重调度。</li>|
|NodeStatus|节点健康状态，由本节点故障处理等级最严重的设备决定。<li>Healthy：该节点故障处理等级存在且不超过NotHandleFault，该节点为健康节点，可以正常训练。</li><li>PreSeparate：该节点故障处理等级存在且不超过PreSeparateFault，该节点为预隔离节点，暂时可能对任务无影响，待任务受到影响退出后，后续不会再调度任务到该节点。</li><li>UnHealthy：该节点故障处理等级存在SeparateFault，该节点为故障节点，将影响训练任务，立即将任务调离该节点。</li>|
|CheckCode|校验码。|

## 自定义节点故障<a name="ZH-CN_TOPIC_0000002479386802"></a>

NodeD组件的配置文件NodeDConfiguration.json为系统配置文件，若用户无特殊需求，请勿随意修改。若用户需要修改故障码的故障级别，可以通过由NodeDConfiguration.json创建的mindx-dl-node-fault-config文件实现，操作指导请参见[（可选）配置节点硬件故障级别](../04_usage/resumable_training/configuration/01_configuring_fault_detection_levels.md#可选配置节点硬件故障级别)。

**表 1**  故障说明

|故障级别|故障处理策略|说明|
|--|--|--|
|NotHandleFault|无需处理|对任务无影响|
|PreSeparateFault|该节点上有任务则不处理，后续调度时不调度任务到该节点|可能导致任务受到影响|
|SeparateFault|任务重调度|任务一定会受到影响|

>[!NOTE]
>故障级别的高低为NotHandleFault < PreSeparateFault < SeparateFault。

**表 2**  节点状态说明

|节点状态|最高故障级别|故障处理策略|说明|
|--|--|--|--|
|Healthy|NotHandleFault|无需处理|该节点为健康节点，可以正常训练。|
|PreSeparate|PreSeparateFault|该节点上有任务则不处理，后续调度时不调度任务到该节点|该节点为预隔离状态，暂时可能对任务无影响，待任务受到影响退出后，后续不会再调度任务到该节点。|
|UnHealthy|SeparateFault|任务重调度|该节点为故障节点，将影响训练任务，立即将任务调离该节点。|

>[!NOTE]
>
>- 当前节点的健康状态，主要通过本节点硬件故障的最高故障级别判断。
>- Healthy、PreSeparate和UnHealthy是MindCluster自定义的节点状态，主要是用于后续任务的调度和处理。
>- PreSeparate节点上任务异常退出后如果需要断点续训，需要开启无条件重试功能。

## 健康探针<a name="ZH-CN_TOPIC_healthz_noded"></a>

NodeD 启动组件内的HTTP健康探针服务，用于K8s livenessProbe机制探测组件存活状态。

**表 2**  健康探针接口

| 项目 | 说明 |
|------|------|
| 路径 | `/` |
| 方法 | GET |
| 默认端口 | 11255 |
| 协议 | HTTP（正确配置--tls-cert-file和--tls-private-key-file参数时为HTTPS） |

**表 3**  响应说明

| 状态码 | 触发条件 | 说明 |
|--------|---------|------|
| 200 OK | 组件正常运行 | 响应体为 `ok` |
| 404 Not Found | 请求路径非 `/` | 探针只响应根路径 |
| 405 Method Not Allowed | 请求方法非 GET | K8s livenessProbe默认使用GET |
| 503 Service Unavailable | 自定义健康检查失败 | 响应体包含具体错误信息 |

**K8s livenessProbe配置示例：**

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

> 探针参数说明详见[启动参数表](../07_developer_guide/installation_deployment/manual_installation/09_noded.md#参数说明)。
