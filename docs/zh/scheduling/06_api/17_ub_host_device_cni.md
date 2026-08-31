# UB Host Device CNI<a name="ZH-CN_TOPIC_ub_host_device_cni"></a>

UB Host Device CNI是基于Host Device CNI适配UB（DPU）网络设备的CNI插件，以二进制方式部署在各work宿主机`/opt/cni/bin`下，由Multus CNI调用，将宿主机的UB网络设备挂载进Pod的网络命名空间，使容器能够通过UB设备进行RDMA高速通信。

UB设备挂在`/sys/bus/ub/devices/<addr>/net`下（UB总线）。UB Host Device CNI通过`ubMode`开启UB挂载逻辑。

## 配置文件说明<a name="ZH-CN_TOPIC_config_ub_host_device_cni"></a>

UB Host Device CNI通过`NetworkAttachmentDefinition`（NAD）中的`spec.config`配置插件参数。

**配置示例：**

```json
{
  "cniVersion": "0.3.1",
  "name": "roce-network",
  "type": "ub-host-device",
  "ubMode": true,
  "inheritHostIP": true,
  "capabilities": { "deviceID": true }
}
```

**表 1**  配置字段说明

| 字段 | 是否必选 | 类型 | 说明 |
|------|---------|------|------|
| type | 是 | string | 固定为`ub-host-device`。 |
| ubMode | 否 | bool | 是否开启UB设备挂载逻辑。开启后只走UB路径，不做PCI归类及DPDK检测。默认值false。 |
| device | 否 | string | 设备名，填写**宿主网卡名**（如`eth0`），非UB与UB模式语义一致。非UB模式按网卡名直接查找；UB模式按网卡名反查其所属的UB设备（扫描`/sys/bus/ub/devices`），详见[UB设备分配来源](#ZH-CN_TOPIC_ub_device_source)。 |
| inheritHostIP | 否 | bool | 挂载时沿用宿主网卡的IP/路由，而不是向IPAM重新申请。默认值false。 |
| capabilities | 否 | object | 声明插件支持的运行时能力。配置`"deviceID": true`后，接收Multus CNI下发的`runtimeConfig.deviceID`作为UB设备挂载来源，详见[UB设备分配来源](#ZH-CN_TOPIC_ub_device_source)。 |
| ipam | 否 | object | IPAM插件配置（如calico-ipam）。`inheritHostIP=true`时可不配置。 |

## UB设备分配来源<a name="ZH-CN_TOPIC_ub_device_source"></a>

`ubMode`下设备ID的获取顺序如下：

1. **NAD配置的`device`（标准host-device机制）**：在NAD里直接指定设备，CNI根据该信息在UB总线上查找设备，不依赖kubelet/multus注入。`device`填写宿主网卡名，插件扫描`/sys/bus/ub/devices`反查该网卡名所属的UB设备。
2. **runtimeConfig.deviceID**：NAD未配置`device`时，通过配置`"capabilities": { "deviceID": true }`采用下发的`runtimeConfig.deviceID`，作为UB设备地址直接挂载。

## NAD配置示例<a name="ZH-CN_TOPIC_nad_config_example"></a>

以下为UB模式下通过K8s RDMA Shared Device Plugin下发的设备，使用`inheritHostIP`、沿用宿主网卡IP的NAD示例：

```yaml
apiVersion: "k8s.cni.cncf.io/v1"
kind: NetworkAttachmentDefinition
metadata:
  name: roce-network
  namespace: default
  annotations:
    k8s.v1.cni.cncf.io/resourceName: "huawei.com/ub_rdma"
spec:
  config: |
    {
      "cniVersion": "0.3.1",
      "name": "roce-network",
      "type": "ub-host-device",
      "ubMode": true,
      "inheritHostIP": true,
      "capabilities": { "deviceID": true }
    }
```

## 业务Pod使用及挂载资源说明<a name="ZH-CN_TOPIC_biz_pod_check_ub_host_device_cni"></a>

业务Pod使用UB设备时，需要通过Multus CNI注解挂载NAD对应网络，并申请对应的RDMA设备资源。以下步骤用于验证业务Pod的资源申请和设备挂载状态。

### 业务Pod资源申请配置<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

业务Pod使用UB设备需要在Pod配置中声明资源请求，并通过`k8s.v1.cni.cncf.io/networks`注解指定挂载的NAD网络，配置示例如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: testpod1
  annotations:
    k8s.v1.cni.cncf.io/networks: roce-network
spec:
  restartPolicy: OnFailure
  containers:
    - image: ubuntu:22.04
      name: appcntr1
      imagePullPolicy: IfNotPresent
      resources:
        requests:
          huawei.com/ub_rdma: '1'
        limits:
          huawei.com/ub_rdma: '1'
      command:
        - /bin/bash
        - -c
        - sleep 300000
```

> [!NOTE]
>
> - `k8s.v1.cni.cncf.io/networks`注解值为NAD的名称（可带命名空间前缀`<namespace>/<name>`），需与[NAD配置示例](#ZH-CN_TOPIC_nad_config_example)中的NAD对应，申请多少个设备就要写多少个名称。
> - 资源名称格式为`<resourcePrefix>/<resourceName>`，需要在设备插件（如K8s RDMA Shared Dev Plugin）的配置文件中定义。

### 业务Pod状态检查步骤<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

1. **查看Pod状态**

   执行以下命令，查看业务Pod是否创建成功：

   ```shell
   kubectl get pod testpod1 -o wide
   ```

   回显示例如下，出现**Running**表示Pod创建成功：

   ```ColdFusion
   NAME      READY   STATUS    RESTARTS   AGE   IP            NODE
   testpod1  1/1     Running   0          10s   10.244.102.*   localhost.localdomain
   ```

2. **查看网络挂载状态**

   执行以下命令，查看Pod网络挂载信息（`k8s.v1.cni.cncf.io/network-status`注解）：

   ```shell
   kubectl describe pod testpod1
   ```

   正常情况下，`network-status`注解中应包含`roce-network`网络及对应的IP地址（如`10.244.102.182`）。

### Pod内设备验证<a name="ZH-CN_TOPIC_biz_pod_device_verify"></a>

业务Pod创建成功后，可以通过以下步骤验证UB设备是否被正确挂载：

1. **进入Pod内部**

   ```shell
   kubectl exec -it testpod1 -- /bin/bash
   ```

2. **检查网络接口**

   ```shell
   ip link show
   ```

   正常情况下应显示挂载的UB网卡（保持宿主网卡名，如`ens2f0`），且链路状态为UP。

3. **检查IP地址**

   ```shell
   ip addr show
   ```

   `inheritHostIP`模式下应显示宿主网卡原有的IP地址；IPAM模式下应显示IPAM分配的IP地址。
