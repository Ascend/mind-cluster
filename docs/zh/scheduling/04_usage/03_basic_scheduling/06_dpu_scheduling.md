# DPU调度<a name="ZH-CN_TOPIC_dpu_scheduling"></a>

K8s RDMA Shared Dev Plugin将节点上的DPU（UB RDMA设备）上报为`<resourcePrefix>/<resourceName>`资源（如`huawei.com/ub_rdma`），业务Pod申请该资源后即可使用，支持以下两种模式：

- **共享模式**：业务Pod申请资源后，组件将节点上的所有RDMA设备挂载到Pod中，多个Pod共享设备。
- **独占模式（beta特性）**：业务Pod申请资源后，组件将kubelet为该Pod分配的RDMA设备挂载到Pod中，单个设备同一时间仅分配给一个Pod。

## 使用前必读<a name="ZH-CN_TOPIC_dpu_scheduling_prerequisite"></a>

- K8s RDMA Shared Dev Plugin已安装，安装方法请参见[K8s RDMA Shared Dev Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md)。
- 组件上报的资源名称`<resourcePrefix>/<resourceName>`通过组件配置文件定义，参见[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)。
- 业务容器使用DPU设备所需的字符设备、驱动配置文件与动态库说明，请参见[容器场景配置](../../06_api/11_k8s_rdma_shared_dev_plugin.md#容器场景配置)。
- 容器化网络使用RDMA设备时，需在集群中安装Multus CNI与UB Host Device CNI，详情请参见[容器化网络使用ROCE网卡](../00_containerization/05_1825_container_network.md)。

## 共享模式<a name="ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin"></a>

业务Pod使用RDMA共享设备时，K8s RDMA Shared Dev Plugin会自动将所有RDMA设备挂载到Pod中。以下步骤用于验证业务Pod的资源申请和设备挂载状态。

### 配置业务Pod资源<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

业务Pod使用RDMA共享设备需要在Pod配置中声明资源请求，配置示例（申请1个RDMA设备资源，最大值配置可参考[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)中的`rdmaHcaMax`）如下：

<pre codetype="yaml">
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
              <strong>huawei.com/ub_rdma: 1</strong>
           limits:
              huawei.com/npu: 8
              <strong>huawei.com/ub_rdma: 1</strong>
        command:
           - sh
           - -c
           - |
              ls -l /dev/infiniband /sys/class/infiniband
              sleep 1000000
</pre>

> [!NOTE]
>
>- 资源名称格式为`<resourcePrefix>/<resourceName>`，需要在K8s RDMA Shared Dev Plugin的配置文件中定义（详见[配置文件说明](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)）。
>- 完整任务yaml参考[acjob-8npu.yaml](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/train/with-dpu-training/rdma-shared/acjob-8npu.yaml)。

### 检查业务Pod状态<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

> [!NOTE]
>
> 业务容器使用1825 DPU设备时，除了需要组件挂载外，还需要配置用户态驱动，两种方式任选其一：
>
>- 在镜像中安装1825 DPU的OFED驱动
>- 从主机侧挂载1825 DPU的OFED驱动（挂载方式参见[容器场景配置](../../06_api/11_k8s_rdma_shared_dev_plugin.md#容器场景配置)）

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

独占模式下，kubelet选择设备，K8s RDMA Shared Dev Plugin按Allocate请求中的设备ID挂载相应的设备，一个DPU字符设备只会分配给一个Pod使用。

## 验证Pod内RDMA设备<a name="ZH-CN_TOPIC_biz_pod_rdma_verify"></a>

业务Pod创建成功后，可以通过以下步骤验证Pod内RDMA设备。

1. 验证RDMA字符设备是否被组件正确挂载。

   1. 进入Pod内部。

      ```shell
      kubectl exec -it rdma-app -- /bin/bash
      ```

   2. 检查RDMA设备节点。

      ```shell
      ls -la /dev/infiniband/
      ```

      正常情况下显示`uverbs0`、`uverbs1`等设备节点文件。如果设备节点为空或不存在，说明RDMA字符设备未正确挂载。

2. 检查RDMA网卡设备及对应的网络接口挂载情况。

   1. 检查Infiniband设备信息。

      ```shell
      ibv_devices
      ```

      正常情况下显示当前节点上的RDMA网卡设备，如`hrn5_0`、`hrn5_1`。

   2. 检查网络接口信息。

      ```shell
      ls -la /sys/class/net
      ```

      正常情况下显示任务申请的RDMA网卡对应的网络接口（如ens***）。

   3. 检查NIC设备数量。

      ```shell
      hinicadm5 info
      ```

      正常情况下显示当前节点的card信息，包含`Card num`和`Device Information`信息（需自行安装或挂载该工具）。
