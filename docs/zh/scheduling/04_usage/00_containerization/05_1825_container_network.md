# 容器化网络使用ROCE网卡

本章节介绍在Kubernetes中，通过Multus CNI、UB Host Device CNI和K8s RDMA Shared Dev Plugin在容器化网络场景下使用ROCE网卡的方法。

## 实现原理

**表 1**  组件分工

| 组件或配置 | 作用 |
|-----------|------|
| K8s RDMA Shared Dev Plugin | 发现并上报RDMA资源；为业务Pod挂载RDMA字符设备。使用Ascend Docker Runtime场景下将需要挂载的设备名称写入 `ASCEND_RDMA_DEVICES`环境变量，由Ascend Docker Runtime完成挂载 |
| Ascend Docker Runtime | 读取`ASCEND_RDMA_DEVICES`环境变量，为业务Pod挂载RDMA字符设备及驱动文件。 |
| Multus CNI | 读取Pod的`k8s.v1.cni.cncf.io/networks`注解及对应NAD，调用UB Host Device CNI进行网络设备挂载。 |
| UB Host Device CNI | 将申请的宿主机网卡移入Pod网络命名空间 |

## 使用约束

- 容器化网络使用场景下，K8s RDMA Shared Dev Plugin开启独占模式时，任务必须申请节点上全部的ROCE网卡设备；K8s RDMA Shared Dev Plugin开启共享模式时，申请的ROCE网卡设备需要满足与NPU的亲和关系。
- 待迁移网卡不能作为宿主机管理网络或集群默认网络使用。网卡移入Pod后，宿主机不能继续通过该接口通信。
- NAD名称、NAD中`device`指定的宿主机网卡名称、Pod中请求的接口名称必须一致。例如`ens0f0@ens0f0`。
- 容器内需配置每张ROCE网卡的策略路由，并在业务启动前完成。

## 环境准备

> - 1\. 确保各计算节点ROCE网卡处于可用状态。可以通过`rdma link show`查看ROCE网卡网络设备的名称，例如：

    ```shell
    link hrn5_0/1 state ACTIVE physical_state LINK_UP netdev ens0f0
    ```

    其中，`ens0f0`为网络接口名称，用于NAD和任务网络注解。

> - 2\. （可选）安装Ascend Docker Runtime。K8s RDMA Shared Dev Plugin可与Ascend Docker Runtime协作进行字符设备的挂载，需先安装Ascend Docker Runtime，再部署K8s RDMA Shared Dev Plugin。安装流程请参见[Ascend Docker Runtime](../../05_developer_guide/00_installation_deployment/00_manual_installation/02_ascend_docker_runtime.md)。

    > [!NOTE]
    > Ascend Docker Runtime仅挂载ROCE网卡驱动文件及RDMA字符设备，`rdma-core`等开源用户态库不会被自动挂载，需在业务镜像中提前安装。
    > 无Ascend Docker Runtime场景下，rdma-dp仅挂载RDMA字符设备，`rdma-core`及ROCE网卡驱动文件需手动安装挂载，请参见[容器场景配置](../../06_api/11_k8s_rdma_shared_dev_plugin.md)

> - 3\. 部署Multus CNI插件，可参考官方仓库`https://github.com/k8snetworkplumbingwg/multus-cni`，执行以下命令以DaemonSet方式部署。

    ```shell
    kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml
    ```

    执行以下命令，确认Multus CNI是否部署成功。

    ```shell
    kubectl get pod -n kube-system -l app=multus
    ```

    回显示例如下，所有Multus Pod为**Running**表示部署成功，此时各节点的`/opt/cni/bin`目录下已生成`multus`二进制文件。

    ```
    NAME                READY   STATUS    RESTARTS   AGE
    kube-multus-ds-8xkl2   1/1   Running   0        74s
    kube-multus-ds-pq6vz   1/1   Running   0        68s
    ```

> - 4\. 部署UB Host Device CNI，部署流程请参见[UB Host Device CNI](../../05_developer_guide/00_installation_deployment/00_manual_installation/17_ub_host_device_cni.md)。
> - 5\. 部署[K8s RDMA Shared Dev Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md)。组件部署YAML中同时预定义了各网络端口的NAD（NetworkAttachmentDefinition）。

    > [!NOTE]
    > 部署YAML中为ROCE网卡的网络端口预定义NAD，每个端口一个，共16个（`ens0f0`至`ens3f3`）。以`ens0f0`为例，NAD内容示例如下。
    >
    > ```yaml
    > apiVersion: "k8s.cni.cncf.io/v1"
    > kind: NetworkAttachmentDefinition
    > metadata:
    >   name: ens0f0
    >   namespace: default
    > spec:
    >   config: |
    >     {
    >       "cniVersion": "0.3.1",
    >       "name": "ens0f0",
    >       "type": "ub-host-device",
    >       "ubMode": true,
    >       "inheritHostIP": true,
    >       "device": "ens0f0"
    >     }
    > ```
    >
    > NAD字段约束及调整方法请参见[NAD配置示例](../../06_api/19_ub_host_device_cni.md#ZH-CN_TOPIC_nad_config_example)。

## 配置任务YAML

### 申请网络设备（NAD）

通过网络注解引用NAD，为业务Pod挂载ROCE网络接口，注解需覆盖全部ROCE接口。

```yaml
metadata:
  annotations:
    k8s.v1.cni.cncf.io/networks:
      ens0f0@ens0f0,ens0f1@ens0f1,ens0f2@ens0f2,ens0f3@ens0f3,
      ens1f0@ens1f0,ens1f1@ens1f1,ens1f2@ens1f2,ens1f3@ens1f3,
```

- 网络注解格式为`<NAD名称>@<容器内接口名称>`，多个网络使用英文逗号分隔，如`ens0f0@ens0f0`。。
- 每个接口仅引用一次，不能重复引用同一个NAD来申请多张网卡。

### 申请字符设备

在业务容器中申请NPU和DPU资源。

```yaml
spec:
  hostNetwork: false
  containers:
    - name: ascend
      image: training-image:tag
      securityContext:
        capabilities:
          add: ["SYS_ADMIN", "NET_ADMIN", "IPC_LOCK"]
      env:
        - name: ASCEND_VISIBLE_DEVICES
          valueFrom:
            fieldRef:
              fieldPath: metadata.annotations['huawei.com/npu']
      resources:
        requests:
          huawei.com/npu: 8
          huawei.com/ub_rdma: 8
        limits:
          huawei.com/npu: 8
          huawei.com/ub_rdma: 8
```

## 配置容器内策略路由

多网卡场景下，仅依赖主路由表的默认路由无法保证报文从预期的DPU网卡发出，策略路由需要为每个独立出口的网络接口配置一张路由表，在表内设置经该接口及其网关发送的默认路由规则，同时添加源地址匹配规则，使以该接口IP作为源地址的报文查询对应路由表，从而选择预期的网络出口。规则和路由需在**业务Pod内部**设置。

策略路由配置可参考以下两条命令模板：

```shell
ip rule add from <NET_IP> table <TABLE_ID> pref <PRIORITY>
ip route add default via <GW_IP> dev <IF_NAME> table <TABLE_ID> onlink
```

**表 2**  策略路由命令参数说明

| 参数 | 说明 |
|------|------|
| `from <NET_IP>/32` | 匹配源地址为该IP的报文。 |
| `table <TABLE_ID>` | 匹配后查询的路由表编号。每张网卡使用不同的编号，例如16张网卡使用100～115，需确保未被其他配置占用。 |
| `pref <PRIORITY>` | 规则优先级，数值越小越先检查。 |
| `default via <GW_IP>` | 该表的默认路由及下一跳网关。 |
| `dev <IF_NAME>` | 出口接口，应与使用此路由表的网络接口名称一致。 |
| `onlink` | 将网关视为该接口链路上直接可达的地址，如果网关可达则不需要配置。 |

以下以IPv4为例：`ens0f0`地址为`192.0.2.10/24`，网关为`192.0.2.1`；`ens0f1`地址为`198.51.100.10/24`，网关为`198.51.100.1`。这些地址仅用于说明，需替换为实际网络规划。

```shell
# ens0f0，路由表100，规则优先级100
ip rule add from 192.0.2.10/32 table 100 pref 100
ip route add default via 192.0.2.1 dev ens0f0 table 100 onlink


ip -4 route replace table 101 default via 198.51.100.1 dev ens0f1 onlink
```

以`192.0.2.10`为源地址的报文匹配规则后查询表100，经`ens0f0`发往网关`192.0.2.1`；以`198.51.100.10`为源地址的报文查询表101，经`ens0f1`发往网关`198.51.100.1`。

对其余网卡重复上述配置，直到所有已申请的DPU网卡都有独立路由表、直连路由、默认路由和源IP规则。

根据实际组网结构配置策略路由的更多详细指导请参见[策略路由配置指导](https://support.huawei.com/carrier/productNewOffering?col=product&path=PBI1-262732867/PBI1-262735886/PBI1-22892969/PBI1-23710427/PBI1-258915853)。
