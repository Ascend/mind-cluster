# UB Host Device CNI<a name="ZH-CN_TOPIC_ub_host_device_cni"></a>

UB Host Device CNI is a CNI plugin that adapts UB (DPU) network devices based on Host Device CNI. It is deployed as a binary in `/opt/cni/bin` on each work host and is called by Multus CNI to mount the UB network device of the host into the network namespace of the Pod, so that the container can perform RDMA high-speed communication through the UB device.

UB devices are attached under `/sys/bus/ub/devices/<addr>/net` (UB). UB Host Device CNI enables the UB mounting logic through `ubMode`.

## Configuration File Description<a name="ZH-CN_TOPIC_config_ub_host_device_cni"></a>

UB Host Device CNI configures plugin parameters through `spec.config` in the `NetworkAttachmentDefinition` (NAD).

**Configuration example:**

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

**Table 1** Configuration field description

| Field | Mandatory | Type | Description |
|------|---------|------|------|
| type | Yes | string | Fixed to `ub-host-device`. |
| ubMode | No | bool | Whether to enable the UB device mounting logic. When enabled, only the UB path is used, and PCI classification and DPDK detection are not performed. The default value is false. |
| device | No | string | Device name. Enter the **host NIC name** (for example, `eth0`). The semantics are the same in non-UB mode and UB mode. In non-UB mode, the NIC is directly found by NIC name. In UB mode, the UB device to which the NIC name belongs is found by reverse lookup (by scanning `/sys/bus/ub/devices`). For details, see [UB Device Allocation Source](#ZH-CN_TOPIC_ub_device_source). |
| inheritHostIP | No | bool | Uses the IP address and route of the host NIC during mounting instead of applying for new ones from the IPAM. The default value is false. |
| capabilities | No | object | Declares the runtime capabilities supported by the plugin. After `"deviceID": true` is configured, the `runtimeConfig.deviceID` delivered by Multus CNI is received as the UB device mounting source. For details, see [UB Device Allocation Source](#ZH-CN_TOPIC_ub_device_source). |
| ipam | No | object | IPAM plugin configuration (for example, calico-ipam). It can be left unconfigured when `inheritHostIP=true`. |

## UB Device Allocation Source<a name="ZH-CN_TOPIC_ub_device_source"></a>

The sequence of obtaining the device ID in `ubMode` is as follows:

1. **`device` configured in the NAD (standard host-device mechanism)**: The device is directly specified in the NAD, and the CNI finds the device on the UB based on this information, without relying on kubelet/Multus injection. `device` is the host NIC name, and the plugin scans `/sys/bus/ub/devices` to find the UB device to which the NIC name belongs.
2. **runtimeConfig.deviceID**: When `device` is not configured in the NAD, configure `"capabilities": { "deviceID": true }` to use the delivered `runtimeConfig.deviceID` and mount it directly as the UB device address.

## NAD Configuration Example<a name="ZH-CN_TOPIC_nad_config_example"></a>

The following is an example of a NAD that uses the device delivered by K8s RDMA Shared Device Plugin in UB mode and uses `inheritHostIP` to inherit the host NIC IP address:

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

## Service Pod Usage and Mounted Resources<a name="ZH-CN_TOPIC_biz_pod_check_ub_host_device_cni"></a>

When a service Pod uses a UB device, it needs to mount the network corresponding to the NAD through the Multus CNI annotation and apply for the corresponding RDMA device resource. The following steps are used to verify the resource application and device mounting status of the service Pod.

### Service Pod Resource Allocation<a name="ZH-CN_TOPIC_biz_pod_resource_config"></a>

A service Pod that uses a UB device needs to declare a resource request in the Pod configuration and specify the NAD network to be mounted through the `k8s.v1.cni.cncf.io/networks` annotation. The configuration example is as follows:

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
> - The value of the `k8s.v1.cni.cncf.io/networks` annotation is the name of the NAD (which can have a namespace prefix `<namespace>/<name>`). It must correspond to the NAD in [NAD Configuration Example](#ZH-CN_TOPIC_nad_config_example). Write as many names as the number of devices to be requested.
> - The resource name format is `<resourcePrefix>/<resourceName>`, which must be defined in the configuration file of the device plugin (for example, K8s RDMA Shared Dev Plugin).

### Service Pod Status Check Procedure<a name="ZH-CN_TOPIC_biz_pod_status_check"></a>

1. Check the Pod status.

   Run the following command to check whether the service Pod is created successfully:

   ```shell
   kubectl get pod testpod1 -o wide
   ```

   The following is the output example. **Running** indicates that the Pod is created successfully:

   ```ColdFusion
   NAME      READY   STATUS    RESTARTS   AGE   IP            NODE
   testpod1  1/1     Running   0          10s   10.244.102.*   localhost.localdomain
   ```

2. Check the network mounting status.

   Run the following command to view the Pod network mounting information (the `k8s.v1.cni.cncf.io/network-status` annotation):

   ```shell
   kubectl describe pod testpod1
   ```

   Normally, the `network-status` annotation should contain `roce-network` and the corresponding IP address (for example, `10.244.102.182`).

### Device Verification in the Pod<a name="ZH-CN_TOPIC_biz_pod_device_verify"></a>

After the service Pod is created successfully, perform the following steps to verify whether the UB device is mounted correctly:

1. Access the Pod.

   ```shell
   kubectl exec -it testpod1 -- /bin/bash
   ```

2. Check the network interface.

   ```shell
   ip link show
   ```

   Normally, the mounted UB NIC should be displayed (the host NIC name is retained, for example, `ens2f0`), and the link status should be UP.

3. Check the IP address.

   ```shell
   ip addr show
   ```

   In `inheritHostIP` mode, the original IP address of the host NIC should be displayed. In IPAM mode, the IP address allocated by the IPAM should be displayed.
