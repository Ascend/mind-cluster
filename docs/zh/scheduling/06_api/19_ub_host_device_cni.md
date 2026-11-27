# UB Host Device CNI<a name="ZH-CN_TOPIC_ub_host_device_cni"></a>

UB Host Device CNI由Multus CNI调用，将宿主机UB网卡移入Pod网络命名空间。安装方法请参见[UB Host Device CNI安装部署](../05_developer_guide/00_installation_deployment/00_manual_installation/17_ub_host_device_cni.md)。

## 配置文件说明<a name="ZH-CN_TOPIC_config_ub_host_device_cni"></a>

插件通过NetworkAttachmentDefinition（NAD）的`spec.config`接收JSON格式的CNI配置。容器化网络使用ROCE网卡时，使用按网卡名称配置的NAD，每个网络接口对应一个NAD。

NAD的通用字段配置字段的含义请参见上游官方文档：Multus CNI（`https://github.com/k8snetworkplumbingwg/multus-cni`）及host-device CNI（`https://github.com/k8snetworkplumbingwg/host-device-cni`）。

**表 1**  UB Host Device CNI字段约束

| 字段 | 类型 | 说明 |
|------|------|------|
| spec.config.type | string | CNI插件类型，固定为`ub-host-device`，Multus以此在CNI二进制目录中查找插件。 |
| spec.config.ubMode | bool | 是否开启UB设备挂载逻辑。开启后只走UB路径，不做PCI归类及DPDK检测。默认值false。 |
| spec.config.inheritHostIP | bool | 网卡移入Pod时是否继承宿主机IP及用户态路由，未开启时需配置ipam。不会迁移宿主机的策略路由规则，容器内需单独配置。 |

## UB设备分配来源<a name="ZH-CN_TOPIC_ub_device_source"></a>

`ubMode`下设备ID的获取顺序如下：

1. **NAD配置的`device`（标准host-device机制）**：在NAD里直接指定设备，CNI根据该信息在UB总线上查找设备，不依赖kubelet/multus注入。`device`填写宿主网卡名，插件扫描`/sys/bus/ub/devices`反查该网卡名所属的UB设备。
2. **runtimeConfig.deviceID**：NAD未配置`device`时，通过配置`"capabilities": { "deviceID": true }`采用下发的`runtimeConfig.deviceID`，作为UB设备地址直接挂载。

## NAD配置示例<a name="ZH-CN_TOPIC_nad_config_example"></a>

NAD资源YAML文件（`k8s-rdma-shared-dp-nad.yaml`）中预定义了以下16个NAD，每个NAD的`metadata.name`、`spec.config.name`和`spec.config.device`需采用相同的网卡名称。

**表 2**  预定义NAD名称

| 分组 | NAD名称（与宿主机网卡名一致） |
|------|---------------------------|
| ens0 | ens0f0、ens0f1、ens0f2、ens0f3 |
| ens1 | ens1f0、ens1f1、ens1f2、ens1f3 |
| ens2 | ens2f0、ens2f1、ens2f2、ens2f3 |
| ens3 | ens3f0、ens3f1、ens3f2、ens3f3 |

以下为`ens0f0`的完整模板。

```yaml
apiVersion: k8s.cni.cncf.io/v1
kind: NetworkAttachmentDefinition
metadata:
  name: ens0f0
  namespace: default
spec:
  config: |
    {
      "cniVersion": "0.3.1",
      "name": "ens0f0",
      "type": "ub-host-device",
      "ubMode": true,
      "inheritHostIP": true,
      "device": "ens0f0"
    }
```

创建前，在每个目标计算节点执行以下命令查询真实名称。网卡已移入Pod时，需在相应Pod网络命名空间内查询。

```shell
rdma link show
```

若节点实际网卡名称与预定义NAD名称不一致，请以`rdma link show`查询到的实际名称为准调整NAD。以实际网卡名为`ens5f0`为例，将`ens0f0`模板中的`metadata.name`、`spec.config.name`和`spec.config.device`三处统一替换为`ens5f0`，其余字段保持不变。

## 业务Pod使用及挂载资源说明<a name="ZH-CN_TOPIC_biz_pod_check_ub_host_device_cni"></a>

### 业务Pod资源申请配置<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

Pod通过`k8s.v1.cni.cncf.io/networks`引用NAD，格式为`<NAD名称>@<容器内接口名称>`。例如`ens0f0@ens0f0`。

UB Host Device CNI会按Multus传入的接口名重命名网卡，此处须通过配置显式指定与宿主机网络设备一致的接口名。

容器化网络使用ROCE网卡的详细指导请参见[容器化网络使用ROCE网卡](../04_usage/00_containerization/05_1825_container_network.md)。

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

   正常情况下，`network-status`注解中应包含申请的网络设备（如`ens0f0`）及对应的IP地址（如`10.244.102.182`）。

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
