# Ascend Dynamic Resource Allocation<a name="ZH-CN_TOPIC_0000002511426dra"></a>

Ascend Dynamic Resource Allocation（简称Ascend DRA）组件基于K8s动态资源分配（Dynamic Resource Allocation）机制，以kubelet插件方式运行在每个计算节点上，负责枚举本节点NPU设备、向API Server上报ResourceSlice、并在kubelet驱动下完成ResourceClaim的Prepare/Unprepare。本文说明其健康检查接口与上报的ResourceSlice格式。

## 健康检查<a name="section_dra_healthz"></a>

Ascend DRA启动内置的HTTP健康探针服务，用于K8s livenessProbe机制探测组件存活状态。探针服务与业务逻辑解耦，使用独立端口，支持HTTP与HTTPS两种协议。

**表 1**  健康探针接口

<a name="table_dra_healthz_interface"></a>

| 项目 | 说明 |
|------|------|
| 路径 | `/` |
| 方法 | GET |
| 默认端口 | 11251 |
| 协议 | HTTP（正确配置--tls-cert-file和--tls-private-key-file参数时为HTTPS） |
| 请求限流 | 1 QPS，突发上限5；超限返回429 |

**表 2**  响应说明

<a name="table_dra_healthz_response"></a>

| 状态码 | 触发条件 | 说明 |
|--------|---------|------|
| 200 OK | 组件正常运行 | 响应体为 `ok` |
| 404 Not Found | 请求路径非 `/` | 探针只响应根路径 |
| 405 Method Not Allowed | 请求方法非 GET | K8s livenessProbe默认使用GET，正常不会触发 |
| 429 Too Many Requests | 请求频率超出限流阈值 | 限流配置见[表1](#table_dra_healthz_interface) |
| 503 Service Unavailable | 注册了自定义健康检查回调且检查失败 | 响应体包含具体错误信息，格式为 `unhealthy: <error>` |

Ascend DRA内置一个健康检查回调`draHealthChecker`，只要插件进程可达即返回健康，因此正常情况下探针恒为200。

**表 3**  健康探针启动参数

<a name="table_dra_healthz_flags"></a>

| 参数 | 含义 | 默认值 |
|------|------|--------|
| --enable-healthz | 是否启用健康检查服务 | false |
| --healthz-address | 健康检查服务监听端口，取值范围[1025, 65535] | 11251 |
| --tls-cert-file | HTTPS证书文件路径，需与--tls-private-key-file同时配置或同时留空 | 空 |
| --tls-private-key-file | HTTPS私钥文件路径，需与--tls-cert-file同时配置或同时留空 | 空 |

> [!NOTE]
> --enable-healthz默认为false，镜像启动命令需显式开启。推荐配置为 `--enable-healthz --healthz-address=11251`。

**K8s livenessProbe配置示例：**

```yaml
livenessProbe:
  httpGet:
    path: /
    port: 11251
    scheme: HTTP
  failureThreshold: 3
  periodSeconds: 10
```

## ResourceSlice上报格式<a name="section_dra_resourceslice"></a>

Ascend DRA启动时枚举本节点全部NPU设备，按K8s资源API（resource.k8s.io/v1）要求组装`DriverResources`，经kubelet插件Helper发布为集群中的ResourceSlice对象。每个节点对应一个Pool，Pool名称为本节点NodeName；每个Pool内包含一个Slice，Slice中包含本节点全部NPU设备。

**表 4**  ResourceSlice对象字段说明

<a name="table_dra_resourceslice"></a>

| 名称 | 含义 | 说明 |
|--|--|--|
| spec.nodeName | 上报该ResourceSlice的节点名称 | 取自DRA启动参数--node-name或环境变量NODE_NAME。 |
| spec.driver | DRA驱动名称 | 固定值 `npu.huawei.com`。 |
| spec.pool.name | 资源池名称 | 等于本节点NodeName，每个节点独占一个Pool。 |
| spec.pool.generation | 资源池代际 | 由kubelet Helper维护，每次更新递增。 |
| spec.slices[].devices[] | 设备列表 | 本节点枚举到的全部NPU设备，每张卡一个Device对象。 |
| spec.slices[].devices[].name | 设备名称 | 格式为 `npu-<phyID>`，全小写。例如 `npu-0`、`npu-12`。 |
| spec.slices[].devices[].attributes | 设备属性 | 设备属性集合，详细说明见[表5](#table_dra_device_attributes)。 |

**表 5**  设备属性（Device.attributes）说明

<a name="table_dra_device_attributes"></a>

| 属性名 | 类型 | 含义 | 说明 |
|--|--|--|--|
| type | string | 设备类型 | 固定值 `npu`。 |
| physicId | int | 设备物理ID | 取自设备管理器GetPhysicIDFromLogicID的返回值，与DeviceName后缀一致。 |
| chipName | string | 芯片名称 | 取自设备管理器GetChipInfo的返回值。获取失败时为空字符串，但属性仍会上报。 |

> [!NOTE]
> 当前Ascend910与Ascend950两类芯片代际上报的属性集合一致，均为type、physicId、chipName三项。新增芯片代际如需扩展属性，在对应代际的DeviceAttributes方法中返回即可，不影响上报链路。

**ResourceSlice示例：**

以下示例展示一个包含两块NPU的节点上报的ResourceSlice（省略部分集群自动填充字段）：

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
> 集群中同时部署了ValidatingAdmissionPolicy，限制每个DRA Pod只能修改本节点对应的ResourceSlice，防止跨节点误写。详见部署文件ascend-dra-driver.yaml中的resourceslices-policy-ascend-dra-driver。
