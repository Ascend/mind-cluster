# Cluster Scheduling Component Atlas Dynamic Resource Allocation

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Atlas Dynamic Resource Allocation is maintained by the [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/en/mindcluster/latest/clustersched/schedulingug/docs/en/scheduling/01_introduction/00_overview.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Atlas Dynamic Resource Allocation

Atlas Dynamic Resource Allocation (Atlas DRA) is one of the core components of the MindCluster cluster scheduling suite. It is deployed as a kubelet plugin on each compute node and, based on the Kubernetes Dynamic Resource Allocation (DRA) mechanism, provides NPU device discovery, reporting, and allocation capabilities.

### Use Cases

The traditional Kubernetes Device Plugin mechanism reports extended resources through kubelet and performs coarse-grained allocation by integer counts, and cannot express the device attributes required by the scheduler in scenarios such as multi-card training, vNPU splitting, HCCS topology awareness, and fine-grained health filtering. To address this, Kubernetes introduced the DRA mechanism: node-level drivers report the attributes of each device through `ResourceSlice`, workloads request devices declaratively through `ResourceClaim`, and kubelet calls back the driver to prepare devices before container startup. MindCluster provides Atlas DRA to implement this mechanism on Atlas NPUs.

> [!NOTE]
> Atlas DRA and Atlas Device Plugin are two different device access methods and are mutually exclusive. Nodes using DRA do not need to deploy Atlas Device Plugin. If a cluster contains nodes using both access methods, distinguish them with node labels.

### Features

- **Device Discovery and Reporting**: On startup, enumerates all NPUs on the local node through DCMI and reports them
  to the API Server as `ResourceSlice`, scoped to the local node. Each device carries stable attributes
  (`type=npu`, `physicId`, `chipName`), allowing the scheduler to filter by real hardware attributes
  instead of scheduling by opaque integer counts.

- **Device Allocation (Prepare)**: Implements the kubelet DRA callback `PrepareResourceClaims`. When a workload's
  `ResourceClaim` is bound to this driver and the Pod is scheduled to the local node, kubelet invokes this callback.
  The driver allocates devices from the local inventory, persists the allocation result to a checkpoint
  (idempotent after restart), writes the CDI spec file, and returns the fully qualified CDI device ID to kubelet.

- **Device Release (Unprepare)**: Implements the kubelet DRA callback `UnprepareResourceClaims`. When a Pod is
  deleted, kubelet invokes this callback, and the driver removes the corresponding CDI spec file and
  cleans up the local checkpoint entry.

- **CDI Spec Generation**: Converts the allocated device name (for example, `npu-0`) into a fully qualified CDI
  device ID and writes a CDI spec file for each claim under the CDI root directory. The container runtime then
  merges the CDI specs and injects `/dev/davinci<id>`, shared device nodes, the `LD_LIBRARY_PATH` environment
  variable, and driver library mounts into the business container.

- **Multi-Generation Support**: Device-related logic is isolated through the `DraGenerationInterface`. Each chip
  generation (Atlas 910, Atlas 950) independently implements enumeration, attribute reporting, and ID conversion.
  Adding a new chip generation only requires adding a new generation implementation, without modifying common code.

- **Health Check**: Exposes an HTTP `healthz` endpoint (default `:11258`, path `/`) for kubelet liveness probes.
  The service is decoupled from business logic and supports both HTTP and HTTPS, rate-limited to 1 QPS with
  a burst limit of 5.

> [!NOTE]
> The DRA mechanism currently supports only basic whole-card scheduling based on Kubernetes and Volcano 1.15. Advanced scheduling features such as affinity scheduling and dynamic virtualization are not yet supported.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Starting from version v26.2.0, tags follow the format below:

```text
<version>-<os>
```

| Field     | Example         | Description                                 |
|-----------|-----------------|---------------------------------------------|
| `version` | `v26.2.0`       | Atlas DRA version number                   |
| `os`      | `ubuntu22.04`   | OS of the Atlas DRA image                  |

### Atlas Dynamic Resource Allocation Latest Version 26.2.0

The following are all images of the latest released 26.2.0 version of Atlas DRA. For all historical version Tags, please refer to [supported_tags.md](./supported_tags.md).

| Tag                      | Dockerfile                                                                                                                                          | Content                                                |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------|
| `v26.2.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-dynamic-resource-allocation/v26.2.0/Dockerfile.ubuntu)        | Atlas DRA v26.2.0 (Base Image: Ubuntu 22.04)          |
| `v26.2.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-dynamic-resource-allocation/v26.2.0/Dockerfile.openeuler) | Atlas DRA v26.2.0 (Base Image: openEuler 24.03)       |

---

## Getting Started

### Prerequisites

#### Software Dependencies

| Software                          | Supported Versions                          | Installation Location | Description                                                                                                              |
|-----------------------------------|---------------------------------------------|----------------------|--------------------------------------------------------------------------------------------------------------------------|
| Kubernetes                        | 1.36.x                                      | All nodes            | For Kubernetes usage, see the [Kubernetes documentation](https://kubernetes.io/docs/)                                    |
| Containerd                        | 2.0+                                        | All nodes            | CDI support is required. Available from the [Containerd official website](https://containerd.io/downloads/)               |
| Atlas AI processor driver and firmware | See the version compatibility table    | Compute nodes       | See the "Installing the NPU Driver and Firmware" section in the [CANN Software Installation Guide](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/softwareinst/instg/instg_0005.html?OS=openEuler&InstallType=local) |
| UMDK software package             | See the version compatibility table         | Compute nodes        | Required when building images for the Atlas 850E supernode, Atlas 650E server, and Atlas 950 SuperPoD. For details, see [Guide to Integrating UMDK into the Base Container Image](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/07_references/02_common_operations.html#容器基础镜像集成umdk安装指导) |

#### Hardware Requirements

| Name   | Requirement |
|--------|-------------|
| CPU    | 0.5 cores   |
| Memory | 0.5 GB      |

#### Installing the Driver

The host machine must have the driver and firmware installed. For details, see the "Installing the NPU Driver and Firmware" section in the [CANN Software Installation Guide](https://www.hiascend.com/document/detail/zh/CANNCommunityEdition/910/softwareinst/instg/instg_0005.html?OS=openEuler&InstallType=local).

### Obtaining the Atlas DRA Image Online

1. Pull the official image

   Pull the Atlas DRA image from AscendHub, replacing {tag} with the actual version number (the latest version is recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-dra:{tag}
   ```

2. Re-tag the image

   Re-tag the pulled official image locally to maintain a consistent local image naming convention and facilitate subsequent O&M management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-dra:{tag} ascend-dra:{tag}
   ```

### Building Locally (Optional)

Example scenario: build an Atlas DRA component image of version v26.2.0 based on Ubuntu 22.04 for the linux-aarch64 architecture.

1. Obtain the Dockerfile for the target architecture

   Go to the [Supported Tags and Dockerfile Links](#supported-tags-and-dockerfile-links) section, open the Dockerfile.ubuntu link for the target version, and save the file to a local directory on an aarch64 environment.

2. Build the Docker image locally (disable caching to ensure a clean build)

   ```bash
   docker build --no-cache -t ascend-dra:v26.2.0 ./ -f Dockerfile.ubuntu
   ```

> [!NOTE]
> If the Docker version is earlier than 18.09, or BuildKit is not manually enabled, the TARGETPLATFORM variable cannot be read during image building, causing the build to fail.
>
> 1. TARGETPLATFORM is a built-in global variable of Docker BuildKit used to identify the target build platform, for example: linux/amd64, linux/arm64.
> 2. This variable is automatically injected only after BuildKit is enabled. It cannot be used on older Docker environments or environments where BuildKit is disabled by default.
> 3. Run the following command to temporarily enable BuildKit before building:
>
> ```bash
> export DOCKER_BUILDKIT=1
> ```

### Deploying Atlas DRA

1. Deploy Atlas DRA

   The deployment manifest is a DaemonSet in the `kube-system` namespace, along with a ServiceAccount, ClusterRole,
   ClusterRoleBinding, `DeviceClass` (`npu.huawei.com`), and `ValidatingAdmissionPolicy`. Before deployment, replace
   the image `{tag}` in `ascend-dra-driver.yaml` with the actual image version.

   ```bash
   kubectl apply -f ascend-dra-driver.yaml
   ```

   Example output:

   ```text
   serviceaccount/ascend-dra-driver-service-account created
   clusterrole.rbac.authorization.k8s.io/ascend-dra-driver-role created
   clusterrolebinding.rbac.authorization.k8s.io/ascend-dra-driver-role-binding created
   daemonset.apps/ascend-dra-driver-kubeletplugin created
   deviceclass.resource.k8s.io/npu.huawei.com created
   validatingadmissionpolicy.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
   validatingadmissionpolicybinding.admissionregistration.k8s.io/resourceslices-policy-ascend-dra-driver created
   ```

2. Verify the deployment

   ```bash
   kubectl get pods -n kube-system | grep ascend-dra-driver
   ```

   Expected result: the `ascend-dra-driver-kubeletplugin` Pod on each compute node is in the `Running` state.

   ```text
   NAME                                        READY   STATUS    RESTARTS   AGE
   ascend-dra-driver-kubeletplugin-5m2xv       1/1     Running   0          74s
   ```

3. Verify ResourceSlice reporting

   ```bash
   kubectl get deviceclass npu.huawei.com
   kubectl get resourceslices
   ```

   Expected result: a DeviceClass named `npu.huawei.com` is present, and each compute node has a corresponding
   ResourceSlice, with the `DEVICECOUNT` column showing the number of discovered NPUs.

   ```text
   NAME            AGE
   npu.huawei.com  2m

   NAME                                  POOLNAME     DEVICECOUNT   AGE
   npu.huawei.com-node1-1a2b3c           node1        8             2m
   ```

> For the format of the `ResourceSlice` fields reported by the component, the semantics of device attributes,
> health check endpoint behavior, and the complete list of component startup parameters, see the official
> documentation: [Atlas Dynamic Resource Allocation API Reference](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/06_api/17_ascend_dynamic_resource_allocation.md)
> and [Manual Installation](../../docs/zh/scheduling/05_developer_guide/00_installation_deployment/00_manual_installation/13_ascend_dynamic_resource_allocation.md).
> For a complete demonstration of the Prepare/Unprepare workflow, see
> [DRA Best Practices](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/04_usage/13_ascend_dynamic_resource_allocation_best_practice/01_deploying_dra_and_running_task.md).

---

## License

View the [license information](https://www.hiascend.com/zh/legal/softlicense) for the Mind series software contained in these images.

As with all container images, the pre-installed software packages (Python, system libraries, etc.) may be governed by their own licenses.
