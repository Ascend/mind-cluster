# DPU调度<a name="ZH-CN_TOPIC_dpu_scheduling"></a>

K8s RDMA Shared Dev Plugin将节点上的DPU（UB RDMA设备）上报为`<resourcePrefix>/<resourceName>`资源（如`huawei.com/ub_rdma`），业务Pod申请该资源后即可使用，支持以下两种模式：

- **共享模式**：业务Pod申请资源后，组件将节点上的所有RDMA设备挂载到Pod中，多个Pod共享设备。
- **独占模式（beta特性）**：组件根据业务Pod的NPU注解与NPU-DPU映射关系，为Pod分配节点上发现的真实UB设备，单个设备同一时间仅分配给一个Pod。

## 使用前必读<a name="ZH-CN_TOPIC_dpu_scheduling_prerequisite"></a>

- K8s RDMA Shared Dev Plugin已安装，安装方法请参见[K8s RDMA Shared Dev Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md)。
- 组件上报的资源名称`<resourcePrefix>/<resourceName>`通过组件配置文件定义，参见[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)。
- 业务容器使用DPU设备所需的字符设备、驱动配置文件与动态库说明，参见[容器场景配置](../../06_api/11_k8s_rdma_shared_dev_plugin.md#容器场景配置)。
- 使用独占模式前，需在集群中安装Multus CNI与UB Host Device CNI，方法参见[独占模式](#ZH-CN_TOPIC_ub_excl_alloc_k8s_rdma_shared_dev_plugin)。

## 共享模式<a name="ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin"></a>

业务Pod使用RDMA共享设备时，K8s RDMA Shared Dev Plugin会自动将所有RDMA设备挂载到Pod中。以下步骤用于验证业务Pod的资源申请和设备挂载状态。

### 配置业务Pod资源<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

业务Pod使用RDMA共享设备需要在Pod配置中声明资源请求，配置示例（申请1份RDMA设备资源，最大值配置可参考[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)中的`rdmaHcaMax`）如下：

```yaml
apiVersion: v1
kind: Pod
metadata:
   name: mofed-test-pod
spec:
   restartPolicy: OnFailure
   hostNetwork: true
   containers:
      - image: ubuntu:22.04
        name: mofed-test-ctr
        imagePullPolicy: IfNotPresent
        securityContext:
           capabilities:
              add: [ "NET_ADMIN", "SYS_ADMIN", "IPC_LOCK" ]
        resources:
           requests:
              huawei.com/npu: 8
              huawei.com/ub_rdma: 1
           limits:
              huawei.com/npu: 8
              huawei.com/ub_rdma: 1
        command:
           - sh
           - -c
           - |
              ls -l /dev/infiniband /sys/class/infiniband
              sleep 1000000
```

> [!NOTE]
>
>- `hostNetwork`必须配置为`true`。由于业务Pod需要访问宿主机的网络命名空间来使用RDMA设备，因此必须启用hostNetwork模式。
>- 资源名称格式为`<resourcePrefix>/<resourceName>`，需要在K8s RDMA Shared Dev Plugin的配置文件中定义（详见[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)）。
>- 完整任务yaml参考[acjob-8npu.yaml](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/train/with-dpu-training/rdma-shared/acjob-8npu.yaml)。

### 检查业务Pod状态<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

> [!NOTE]
>
> 业务容器使用1825 DPU设备时，除了需要组件挂载外，还需要：
>
> - 配置主机网络`hostNetwork: true`
> - 配置用户态驱动，两种方式任选其一：
    >   - 在镜像中安装1825 DPU的OFED驱动
>   - 启动容器后从主机挂载1825 DPU的OFED驱动（挂载方式参见[容器场景配置](../../06_api/11_k8s_rdma_shared_dev_plugin.md#容器场景配置)）

执行以下命令，查看业务Pod是否创建成功：

```shell
kubectl get pod rdma-app -o wide
```

回显示例如下，出现**Running**表示Pod创建成功：

```ColdFusion
NAME       READY   STATUS    RESTARTS   AGE   IP            NODE
rdma-app   1/1     Running   0          10s   10.244.1.*   compute-node-1
```

## 独占模式（beta特性）<a name="ZH-CN_TOPIC_ub_excl_alloc_k8s_rdma_shared_dev_plugin"></a>

为组件增加`-ub-excl-mode`启动参数后，组件工作在独占模式，开启方法与参数说明请参见[独占模式配置](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md#section187410285361)。

独占模式下，设备挂载依赖Multus CNI读取Pod的`k8s.v1.cni.cncf.io/device-status`注解，并通过`runtimeConfig.deviceID`将设备地址下发给UB Host Device CNI，因此使用独占模式前需在集群中安装Multus CNI。

### Multus CNI安装

1. 可参考官方[Multus CNI](https://github.com/k8snetworkplumbingwg/multus-cni)安装Multus CNI。
2. 执行以下命令，以DaemonSet方式部署Multus CNI。

    ```shell
    kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml
    ```

3. 执行以下命令，确认Multus CNI是否部署成功。

    ```shell
    kubectl get pod -n kube-system -l app=multus
    ```

   回显示例如下，所有Multus Pod为**Running**表示部署成功，此时各节点的`/opt/cni/bin`目录下已生成`multus`二进制文件。

    ```ColdFusion
    NAME                READY   STATUS    RESTARTS   AGE
    kube-multus-ds-8xkl2   1/1   Running   0        74s
    kube-multus-ds-pq6vz   1/1   Running   0        68s
    ```

4. 参见[UB Host Device CNI](../../06_api/19_ub_host_device_cni.md)中的NAD配置示例创建NAD，供业务Pod通过`k8s.v1.cni.cncf.io/networks`注解挂载网络。

### 独占模式的设备分配流程

1. Volcano在调度时将分配给Pod的NPU卡信息写入Pod的`huawei.com/npu`注解，格式为NPU ID列表，例如`"0,1"`。
2. Pod创建后，组件通过kubelet `/pods`端点读取本节点业务Pod信息，获取`huawei.com/npu`注解中的NPU ID。
3. 组件通过镜像内置的`/etc/rdma-plugin/npu-nic-mapping.json`配置文件，将NPU ID映射为以该NPU作为主网卡的DPU设备。
4. 组件完成设备分配后，将分配结果写入Pod的`k8s.v1.cni.cncf.io/device-status`注解，供Multus CNI与UB Host Device CNI完成设备挂载。

### 独占模式下业务Pod的资源配置示例

```yaml
apiVersion: v1
kind: Pod
metadata:
   name: ub-rdma-test-pod
   annotations:
      k8s.v1.cni.cncf.io/networks: roce-network, roce-network
spec:
   restartPolicy: OnFailure
   containers:
      - image: ubuntu:22.04
        name: ub-rdma-test-ctr
        imagePullPolicy: IfNotPresent
        securityContext:
           capabilities:
              add: [ "IPC_LOCK" ]
        resources:
           requests:
              huawei.com/npu: 2
              huawei.com/ub_rdma: 2
           limits:
              huawei.com/npu: 2
              huawei.com/ub_rdma: 2
        command:
           - sh
           - -c
           - |
              ls -l /dev/infiniband /sys/class/infiniband
              sleep 1000000
```

> [!NOTE]
>
>- `k8s.v1.cni.cncf.io/networks`注解值为NAD的名称，NAD需按UB Host Device CNI的要求配置（参见[UB Host Device CNI](../../06_api/19_ub_host_device_cni.md)），申请多少个设备就写多少个名称。
>- 独占模式基于UB类型设备。
>- 独占模式的开启方法与参数说明请参见[独占模式配置](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md#section187410285361)。
>- 完整任务yaml参考[独占模式任务](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/train/with-dpu-training/rdma-exclusive)

Pod创建成功后，可查看Pod的`k8s.v1.cni.cncf.io/device-status`注解确认分配结果：

```shell
kubectl get pod ub-rdma-test-pod -o jsonpath='{.metadata.annotations.k8s\.v1\.cni\.cncf\.io/device-status}'
```

注解值为JSON格式，Key为申请了独占模式资源的容器名称，Value为分配给该容器的设备ID列表。若注解不存在，表示设备尚未分配，可稍后重试或检查组件日志。

## 验证Pod内RDMA设备<a name="ZH-CN_TOPIC_biz_pod_rdma_verify"></a>

业务Pod创建成功后，可以通过以下步骤验证RDMA设备是否被组件正确挂载：

1. 进入Pod内部

    ```shell
    kubectl exec -it rdma-app -- /bin/bash
    ```

2. 检查RDMA设备节点

    ```shell
    ls -la /dev/infiniband/
    ```

   正常情况下显示`uverbs0`、`uverbs1`等设备节点文件。如果设备节点为空或不存在，说明RDMA设备未正确挂载。

业务Pod创建成功后，还需检查RDMA网卡设备及对应的网络接口挂载情况：

1. 检查Infiniband设备信息

    ```shell
    ls -la /sys/class/infiniband/
    ```

   正常情况下显示当前节点上的RDMA网卡设备，如`hrn5_0`、`hrn5_1`。

2. 检查网络接口信息

   ```shell
   ls -la /sys/class/net/
   ```

   正常情况下显示节点上的网络接口设备，包括RDMA网卡对应的网络接口（如`ens***`）。
3. 检查nic设备数量

   ```shell
   hinicadm5 info
   ```

   正常情况下显示当前节点的card信息，包含'Card num'和'Device Information'信息。
