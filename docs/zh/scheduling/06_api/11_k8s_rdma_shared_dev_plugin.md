# K8s RDMA Shared Dev Plugin<a name="ZH-CN_TOPIC_k8s_rdma_shared_dev_plugin"></a>

## 健康探针<a name="ZH-CN_TOPIC_healthz_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin启动组件内的HTTP健康探针服务，用于K8s livenessProbe机制探测组件存活状态。

**表 1**  健康探针接口

| 项目 | 说明 |
|------|------|
| 路径 | `/` |
| 方法 | GET |
| 默认端口 | 11257 |
| 协议 | HTTP（正确配置--tls-cert-file和--tls-private-key-file参数时为HTTPS） |

**表 2**  响应说明

| 状态码 | 触发条件 | 说明 |
|--------|---------|------|
| 200 OK | 组件正常运行 | 响应体为 `ok` |
| 404 Not Found | 请求路径非 `/` | 探针只响应根路径 |
| 405 Method Not Allowed | 请求方法非 GET | K8s livenessProbe默认使用GET，正常不会触发 |
| 503 Service Unavailable | 注册了自定义健康检查回调且检查失败 | 响应体包含具体错误信息 |

**K8s livenessProbe配置示例：**

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

探针启动参数说明详见[K8s RDMA Shared Dev Plugin 启动参数](../05_developer_guide/00_installation_deployment/00_manual_installation/12_k8s_rdma_shared_dev_plugin.md#参数说明)。

## 配置文件说明<a name="ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin通过`-config-file`参数指定的JSON配置文件配置RDMA设备的选择器，用于发现并上报节点上的RDMA设备资源。配置文件默认路径为`/k8s-rdma-shared-dev-plugin/config.json`，通过ConfigMap（名称为`rdma-devices`）挂载进容器。

**配置文件示例：**

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

**表 3**  顶层配置字段

| 字段 | 是否必选 | 类型 | 默认值 | 说明                                                           |
|------|---------|------|--------|--------------------------------------------------------------|
| periodicUpdateInterval | 否 | int | 60 | 周期性更新设备资源的时间间隔，单位为秒。取值为0时禁用周期性更新设备资源功能；未设置时使用默认值60秒；取值不能小于0。 |
| faultDetectPeriod | 否 | int | 0 | 故障检测周期，单位为秒，仅对UB类型设备生效。未设置或取值小于1时禁用故障检测功能。                   |
| configList | 是 | object列表 | - | 资源配置列表，每个元素描述一组RDMA设备的上报规则，列表至少包含1个元素。                       |

configList中每个配置对象的字段说明详见[表4](#table_config_list_k8s_rdma_shared_dev_plugin)。

**表 4**  configList配置字段<a name="table_config_list_k8s_rdma_shared_dev_plugin"></a>

| 字段 | 是否必选 | 类型 | 默认值 | 说明 |
|------|---------|------|--------|------|
| resourceName | 是 | string | - | 设备资源名称，在resourcePrefix作用域内必须唯一，仅支持大小写字母、数字和下划线。最终上报给K8s的资源名称格式为`<resourcePrefix>/<resourceName>`。 |
| resourcePrefix | 否 | string | rdma | 设备资源前缀，需为合法DNS子域名（仅支持小写字母、数字、连字符和点）。 |
| rdmaHcaMax | 是 | int | - | 设备插件可提供的RDMA资源最大数量，取值不能小于0。 |
| selectors | 否 | object | - | 设备选择器，用于过滤目标设备，字段说明详见[表5](#table_selectors_k8s_rdma_shared_dev_plugin)。 |

**表 5**  selectors选择器字段<a name="table_selectors_k8s_rdma_shared_dev_plugin"></a>

| 字段 | 类型 | 说明 |
|------|------|------|
| buses | string列表 | 设备总线类型。取值包含`ub`时启用UB设备模式；未配置或取值非`ub`时启用PCI设备模式。UB设备不支持CDI模式。 |
| vendors | string列表 | 设备厂商十六进制编码，例如`["0xcc08"]`。 |
| deviceIDs | string列表 | 设备型号十六进制编码，例如`["0x8200"]`。 |
| drivers | string列表 | 设备驱动名称，例如`["mlx5_core"]`。 |
| ifNames | string列表 | 网络接口名称，例如`["ib0"]`。 |
| linkTypes | string列表 | 网络接口链路类型，例如`["ether"]`、`["infiniband"]`。 |

**选择器匹配规则：**

- 同一选择器字段内的多个元素之间为逻辑“或”关系，例如`"vendors": ["15b3", "0xcc08"]`表示匹配厂商为`15b3`或`0xcc08`的设备。
- 不同选择器字段之间为逻辑“与”关系，例如同时配置`vendors`和`deviceIDs`时，设备需同时满足厂商和型号条件才会被选中。
- 未配置选择器字段将被忽略，不参与过滤。

## ConfigMap说明<a name="ZH-CN_TOPIC_configmap_k8s_rdma_shared_dev_plugin"></a>

K8s RDMA Shared Dev Plugin通过配置文件中的`faultDetectPeriod`参数启用故障检测功能，仅对UB类型设备生效。启动故障检测功能后，会将DPU设备的故障信息上报到Kubernetes ConfigMap中。ConfigMap位于`kube-system`命名空间下，名称为`dpuinfo-<node-name>`（其中`<node-name>`为节点名称），Label为`huawei.com/consumer.clusterd=true`。ConfigMap采用强制更新策略：当检测到故障信息发生变化或距离上次更新超过5分钟时，会触发ConfigMap更新。

ConfigMap中Data字段的Key为`DpuInfoCfg`，Value为JSON格式的DPU故障信息，详细说明请参见[表6](#table_dpuconfigmap_k8s_rdma_shared_dev_plugin)。

**表 6**  dpuinfo-\<node-name\> ConfigMap
<a name="table_dpuconfigmap_k8s_rdma_shared_dev_plugin"></a>

|字段|类型|说明|
|--|--|--|
|DPUInfo|对象|DPU设备故障信息。|
|-DPUList|列表|DPU设备列表。数组中的每个元素描述一个DPU设备的故障信息，详细说明请参见[表7](#table_dpuitem_k8s_rdma_shared_dev_plugin)。|
|-NodeEvent|对象|节点级故障事件，例如DPU卡脱落等。详细说明请参见[表8](#table_nodeevent_k8s_rdma_shared_dev_plugin)。|
|UpdateTime|RFC 3339 格式时间戳|当前DPU信息的更新时间，用于标识故障信息的最新上报时间。|

**表 7**  DPUList元素字段说明
<a name="table_dpuitem_k8s_rdma_shared_dev_plugin"></a>

|字段|类型|说明|
|--|--|--|
|HcaName|字符串|HCA设备名称，例如`mlx5_0`。|
|EthName|字符串|关联的以太网接口名称。|
|IpAddr|字符串|DPU设备的IP地址。|
|DeviceID|字符串|设备ID，十六进制格式。|
|VendorID|字符串|厂商ID，十六进制格式。|
|FaultList|列表|该DPU设备上的故障明细列表。数组中的每个元素描述一条故障信息，详细说明请参见[表9](#table_faultdetail_k8s_rdma_shared_dev_plugin)。|

**表 8**  NodeEvent字段说明
<a name="table_nodeevent_k8s_rdma_shared_dev_plugin"></a>

|字段|类型|说明|
|--|--|--|
|NodeName|字符串|节点名称。|
|FaultList|列表|节点级故障明细列表，详细说明请参见[表9](#table_faultdetail_k8s_rdma_shared_dev_plugin)。|

**表 9**  FaultList字段说明
<a name="table_faultdetail_k8s_rdma_shared_dev_plugin"></a>

| 字段          |类型|说明|
|-------------|--|--|
| FaultCode   |字符串|故障码，用于标识故障类型。|
| Time        |Unix毫秒时间戳|故障首次检测时间。|
| Description |字符串|故障描述信息。|
| FaultLevel  |字符串|故障等级。|

## 业务Pod使用及挂载资源说明<a name="ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin"></a>

业务Pod使用RDMA共享设备时，K8s RDMA Shared Dev Plugin会自动将所有RDMA设备挂载到Pod中。以下步骤用于验证业务Pod的资源申请和设备挂载状态。

### 业务Pod资源申请配置<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

业务Pod使用RDMA共享设备需要在Pod配置中声明资源请求，配置示例（申请1份RDMA设备资源，最大值可配参考[配置文件说明](#配置文件说明)rdmaHcaMax）如下：

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
> `hostNetwork`必须配置为`true`。由于业务Pod需要访问宿主机的网络命名空间来使用RDMA设备，因此必须启用hostNetwork模式。
> 资源名称格式为`<resourcePrefix>/<resourceName>`，需要在K8s RDMA Shared Dev Plugin的配置文件中定义（详见[配置文件说明](#配置文件说明)）。

### 业务Pod状态检查步骤<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

> [!NOTE]
>
> 业务容器使用1825 DPU设备时，除了需要组件挂载外，还需要：
>
> - 配置主机网络 `hostNetwork: true`
> - 配置用户态驱动，两种方式任选其一：
>   1. 在镜像中安装1825 DPU的OFED驱动
>   2. 启动容器后从主机挂载1825 DPU的OFED驱动

1. **查看Pod状态**

   执行以下命令，查看业务Pod是否创建成功：

    ```shell
    kubectl get pod rdma-app -o wide
    ```

   回显示例如下，出现 **Running** 表示Pod创建成功：

    ```ColdFusion
    NAME       READY   STATUS    RESTARTS   AGE   IP            NODE
    rdma-app   1/1     Running   0          10s   10.244.1.*   compute-node-1
    ```

### Pod内RDMA设备验证<a name="ZH-CN_TOPIC_biz_pod_rdma_verify"></a>

业务Pod创建成功后，可以通过以下步骤验证RDMA设备是否被组件正确挂载：

1. **进入Pod内部**

    ```shell
    kubectl exec -it rdma-app -- /bin/bash
    ```

2. **检查RDMA设备节点**

    ```shell
    ls -la /dev/infiniband/
    ```

   正常情况下应显示`uverbs0`、`uverbs1`等设备节点文件。如果设备节点为空或不存在，说明RDMA设备未正确挂载。

业务Pod创建成功后，还需检查RDMA网卡设备及对应的网络接口挂载情况：

1**检查Infiniband设备信息**

    ```shell
    ls -la /sys/class/infiniband/
    ```

   正常情况下应显示当前节点上的RDMA网卡设备，如`hrn5_0`、`hrn5_1`。

2**检查网络接口信息**

     ```shell
     ls -la /sys/class/net/
     ```

   正常情况下应显示节点上的网络接口设备，包括RDMA网卡对应的网络接口（如`ens***`）。如果网络没出现在Pod内，需要检查hostNetwork是否配置为true。
