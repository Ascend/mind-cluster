# DPU Job Scheduling<a name="ZH-CN_TOPIC_dpu_scheduling"></a>

K8s RDMA Shared Dev Plugin reports the DPU (UB RDMA device) on a node as a `<resourcePrefix>/<resourceName>` resource (for example, `huawei.com/ub_rdma`). A service Pod can use the resource after applying for it. The following two modes are supported:

- **Shared mode**: After a service Pod applies for the resource, the component mounts all RDMA devices on the node to the Pod, and multiple Pods share the devices.
- **Exclusive mode (beta feature)**: Based on the NPU annotation of the service Pod and the NPU-DPU mapping, the component allocates the real UB device discovered on the node to the Pod. A single device is allocated to only one Pod at a time.

## Before You Start<a name="ZH-CN_TOPIC_dpu_scheduling_prerequisite"></a>

- K8s RDMA Shared Dev Plugin has been installed. For the installation method, see [K8s RDMA Shared Dev Plugin](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md).
- The resource name `<resourcePrefix>/<resourceName>` reported by the component is defined in the component configuration file. For details, see [Configuration File Description](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin).
- For details about the character devices, driver configuration files, and dynamic libraries required for a service container to use DPU devices, see [Configuration in Container Scenarios](../../06_api/11_k8s_rdma_shared_dev_plugin.md#configuration-in-container-scenarios).
- Before using exclusive mode, install Multus CNI and UB Host Device CNI in the cluster. For details, see [Exclusive Mode](#ZH-CN_TOPIC_ub_excl_alloc_k8s_rdma_shared_dev_plugin).

## Shared Mode<a name="ZH-CN_TOPIC_biz_pod_check_k8s_rdma_shared_dev_plugin"></a>

When a service Pod uses an RDMA shared device, K8s RDMA Shared Dev Plugin automatically mounts all RDMA devices to the Pod. The following steps are used to verify the resource application and device mounting status of the service Pod.

### Configuring Service Pod Resources<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

A service Pod that uses an RDMA shared device needs to declare a resource request in the Pod configuration. The configuration example is as follows (applying for one RDMA device resource; for the maximum value configuration, see `rdmaHcaMax` in [Configuration File Description](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)):

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
>- `hostNetwork` must be set to `true`. Because the service Pod needs to access the network namespace of the host to use RDMA devices, the `hostNetwork` mode must be enabled.
>- The resource name format is `<resourcePrefix>/<resourceName>`, which must be defined in the configuration file of K8s RDMA Shared Dev Plugin (for details, see [Configuration File Description](../../06_api/11_k8s_rdma_shared_dev_plugin.md#ZH-CN_TOPIC_config_k8s_rdma_shared_dev_plugin)).
>- For the complete job yaml, see [acjob-8npu.yaml](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/train/with-dpu-training/rdma-shared/acjob-8npu.yaml).

### Checking the Service Pod Status<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

> [!NOTE]
>
> When a service container uses a 1825 DPU device, in addition to component mounting, you also need to:
>
> - Configure the host network `hostNetwork: true`.
> - Configure the user-mode driver. Choose either of the following methods:
    >   - Install the OFED driver of the 1825 DPU in the image.
    >   - After the container is started, mount the OFED driver of the 1825 DPU from the host (for the mounting method, see [Configuration in Container Scenarios](../../06_api/11_k8s_rdma_shared_dev_plugin.md#configuration-in-container-scenarios)).

Run the following command to check whether the service Pod is created successfully:

```shell
kubectl get pod rdma-app -o wide
```

The following is an example. **Running** indicates that the Pod is created successfully:

```ColdFusion
NAME       READY   STATUS    RESTARTS   AGE   IP            NODE
rdma-app   1/1     Running   0          10s   10.244.1.*   compute-node-1
```

## Exclusive Mode (Beta Feature)<a name="ZH-CN_TOPIC_ub_excl_alloc_k8s_rdma_shared_dev_plugin"></a>

After the `-ub-excl-mode` startup parameter is added for the component, the component works in exclusive mode. For the enabling method and parameter description, see [Exclusive Mode Configuration](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md#section187410285361).

In exclusive mode, device mounting depends on Multus CNI that reads the `k8s.v1.cni.cncf.io/device-status` annotation of the Pod and delivers the device address to UB Host Device CNI through `runtimeConfig.deviceID`. Therefore, Multus CNI must be installed in the cluster before exclusive mode is used.

### Installing Multus CNI

1. You can install Multus CNI by referring to the official [Multus CNI documentation](https://github.com/k8snetworkplumbingwg/multus-cni).
2. Run the following command to deploy Multus CNI as a DaemonSet.

    ```shell
    kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset.yml
    ```

3. Run the following command to confirm whether Multus CNI is deployed successfully.

    ```shell
    kubectl get pod -n kube-system -l app=multus
    ```

   The following is an example. If all Multus Pods are in the **Running** state, the deployment is successful. In this case, the `multus` binary file has been generated in the `/opt/cni/bin` directory on each node.

    ```ColdFusion
    NAME                READY   STATUS    RESTARTS   AGE
    kube-multus-ds-8xkl2   1/1   Running   0        74s
    kube-multus-ds-pq6vz   1/1   Running   0        68s
    ```

4. Create a NAD by referring to the NAD configuration example in [UB Host Device CNI](../../06_api/19_ub_host_device_cni.md), so that service Pods can mount the network through the `k8s.v1.cni.cncf.io/networks` annotation.

### Device Allocation Process in Exclusive Mode

1. During scheduling, Volcano writes the NPU information allocated to the Pod to the `huawei.com/npu` annotation of the Pod, in the format of an NPU ID list, for example, `"0,1"`.
2. After the Pod is created, the component reads the service Pod information of the node through the kubelet `/pods` endpoint to obtain the NPU ID in the `huawei.com/npu` annotation.
3. The component maps the NPU ID to the DPU device that uses this NPU as the primary NIC, through the `/etc/rdma-plugin/npu-nic-mapping.json` configuration file built into the image.
4. After the component completes device allocation, it writes the allocation result to the `k8s.v1.cni.cncf.io/device-status` annotation of the Pod, so that Multus CNI and UB Host Device CNI can complete device mounting.

### Service Pod Resource Configuration Example in Exclusive Mode

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
>- The value of the `k8s.v1.cni.cncf.io/networks` annotation is the name of the NAD, and the NAD must be configured as required by UB Host Device CNI (see [UB Host Device CNI](../../06_api/19_ub_host_device_cni.md)). Write as many names as the number of devices to be requested.
>- Exclusive mode is based on UB type devices.
>- For the enabling method and parameter description of exclusive mode, see [Exclusive Mode Configuration](../../05_developer_guide/00_installation_deployment/00_manual_installation/11_k8s_rdma_shared_dev_plugin.md#section187410285361).
>- For the complete job yaml, see [Exclusive Job](https://gitcode.com/Ascend/mindcluster-deploy/tree/master/samples/train/with-dpu-training/rdma-exclusive).

After the Pod is created successfully, you can view the `k8s.v1.cni.cncf.io/device-status` annotation of the Pod to confirm the allocation result:

```shell
kubectl get pod ub-rdma-test-pod -o jsonpath='{.metadata.annotations.k8s\.v1\.cni\.cncf\.io/device-status}'
```

The annotation value is in JSON format. The `Key` is the name of the container that applied for the exclusive mode resource, and the `Value` is the list of device IDs allocated to the container. If the annotation does not exist, the device has not been allocated yet. You can retry later or check the component logs.

## Verifying RDMA Devices in a Pod<a name="ZH-CN_TOPIC_biz_pod_rdma_verify"></a>

After the service Pod is created successfully, perform the following steps to verify whether the RDMA device is mounted correctly by the component:

1. Access the Pod.

    ```shell
    kubectl exec -it rdma-app -- /bin/bash
    ```

2. Check the RDMA device nodes.

    ```shell
    ls -la /dev/infiniband/
    ```

   Normally, device node files such as `uverbs0` and `uverbs1` are displayed. If the device node is empty or does not exist, the RDMA device is not mounted correctly.

After the service Pod is created successfully, you also need to check the mounting of RDMA NIC devices and the corresponding network interfaces:

1. Check the InfiniBand device information.

    ```shell
    ls -la /sys/class/infiniband/
    ```

   Normally, the RDMA NIC devices on the current node are displayed, such as `hrn5_0` and `hrn5_1`.

2. Check the network interface information.

   ```shell
   ls -la /sys/class/net/
   ```

   Normally, the network interface devices on the node are displayed, including the network interface corresponding to the RDMA NIC (for example, `ens***`).
3. Check the number of NIC devices.

   ```shell
   hinicadm5 info
   ```

   Normally, the card information of the current node is displayed, including `Card num` and `Device Information`.
