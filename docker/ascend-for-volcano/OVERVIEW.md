# Cluster Scheduling Component Ascend for Volcano

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.1.0
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.1.0
```

***

## Ascend for Volcano

Ascend for Volcano is built on the open-source Volcano scheduling plugin mechanism, adding features such as Ascend AI processor (NPU) affinity scheduling and virtual device scheduling. It is deployed on management nodes. Volcano is a scheduler component based on the interconnect topology and processing logic of Ascend AI processors, achieving optimal utilization of Ascend AI processors and maximizing their computing performance.

### Use Cases

Kubernetes basic scheduling can only perform resource scheduling by sensing the number of Ascend chips. To achieve affinity scheduling and maximize resource utilization, it is necessary to be aware of the network connection methods between Ascend chips and select network-optimal resources. MindCluster provides the Volcano service deployed on management nodes, offering network affinity scheduling for different Ascend devices and network topologies.

### Features

- **Available Device Calculation**: Calculates available device information for the cluster based on fault information and node information reported by lower-level cluster scheduling components. (`self-maintain-available-card` is enabled by default. When disabled, available device information is obtained from lower-level cluster scheduling components.)
- **Optimal Resource Allocation**: Retrieves the user's desired resource quantity from K8s task objects, combines it with the cluster's device quantity, device type, and device network topology to select optimal resources for task allocation.
- **Fault Rescheduling**: Reschedules tasks when resource faults occur.
- **NPU Affinity Scheduling**: Based on the interconnect topology of Ascend AI processors, prioritizes scheduling tasks to processors within the same card, then to HCCS-interconnected processors, and finally to PCIe-interconnected processors, reducing resource fragmentation and network congestion.
- **Switch Affinity Scheduling**: Based on the network configuration of nodes under switches and the parameter plane network configuration, achieves optimal node utilization. Supports various network topologies including Spine-Leaf dual-layer interconnect and single-layer switch interconnect.
- **Logical Super Node Affinity Scheduling**: Partitions physical super nodes into logical super nodes based on splitting strategies to achieve optimal node utilization.
- **Multi-level Scheduling Policy**: Abstracts cluster resources into a multi-level structure based on NPU network topology hierarchy, configurable via the `resource-level-config` parameter.
- **Multiple Scheduling Modes**: Supports whole-card scheduling, static vNPU scheduling, dynamic vNPU scheduling, and soft-partition scheduling.

### Affinity Policies

The following affinity policies are established based on the characteristics of Ascend 910 AI processors and resource utilization rules (listed in priority order):

1. **HCCS Affinity Scheduling Principle**: Requested Ascend 910 AI processors must be within the same HCCS ring. Prioritize the HCCS with the most matching number of available processors.
2. **Fill-first Scheduling Principle**: Prioritize allocating AI servers that already have Ascend 910 AI processors assigned, reducing fragmentation.
3. **Even-remainder Priority Principle**: Prioritize HCCS that satisfy the above conditions, then select HCCS with an even number of remaining processors.

### Upstream and Downstream Dependencies

1. Calculates cluster resource information based on information reported by ClusterD (default scenario using ClusterD).
2. Receives task launch configurations from third parties, selects optimal node resources based on cluster resource information.
3. Passes specific resource selection information to Ascend Device Plugin on compute nodes to complete device mounting.

### Image Description

Ascend for Volcano consists of two images:

- **volcano-scheduler**: Volcano scheduler image, containing the Ascend NPU affinity scheduling plugin (`volcano-npu_*.so`).
- **volcano-controller**: Volcano controller image.

***

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<component-version>-<ascend-plugin-version>
```

| Field                   | Example   | Description                          |
| ----------------------- | --------- | ------------------------------------ |
| `component-version`     | `v1.7.0`  | Volcano component version            |
| `ascend-plugin-version` | `v26.1.0` | Ascend NPU scheduling plugin version |

### Ascend for Volcano 26.1.0 (Volcano v1.9.0)

| Tag              | Dockerfile                                                            | Image Content                                                                                |
| ---------------- | --------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `v1.9.0-v26.1.0` | [Dockerfile-scheduler](volcano-v1.9.0/v26.1.0/Dockerfile-scheduler)   | Volcano scheduler v26.1.0 image (with Ascend NPU scheduling plugin, based on Volcano v1.9.0) |
| `v1.9.0-v26.1.0` | [Dockerfile-controller](volcano-v1.9.0/v26.1.0/Dockerfile-controller) | Volcano controller v26.1.0 image (based on Volcano v1.9.0)                                   |

### Ascend for Volcano 26.1.0 (Volcano v1.7.0)

| Tag              | Dockerfile                                                            | Image Content                                                                                |
| ---------------- | --------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `v1.7.0-v26.1.0` | [Dockerfile-scheduler](volcano-v1.7.0/v26.1.0/Dockerfile-scheduler)   | Volcano scheduler v26.1.0 image (with Ascend NPU scheduling plugin, based on Volcano v1.7.0) |
| `v1.7.0-v26.1.0` | [Dockerfile-controller](volcano-v1.7.0/v26.1.0/Dockerfile-controller) | Volcano controller v26.1.0 image (based on Volcano v1.7.0)                                   |

***

## Quick Start

### Prerequisites

#### Software Dependencies

| Software             | Supported Versions      | Installation Location | Description                                                          |
| -------------------- | ----------------------- | --------------------- | -------------------------------------------------------------------- |
| Kubernetes           | 1.19.x\~1.34.x          | All nodes             | See [Kubernetes Documentation](https://kubernetes.io/docs/)          |
| Ascend Device Plugin | Same version as Volcano | Compute nodes         | Volcano depends on Ascend Device Plugin to report device information |
| ClusterD             | Same version as Volcano | Management nodes      | Volcano depends on ClusterD to aggregate cluster fault information   |

#### Hardware Requirements

| Resource                  | Up to 100 Nodes | 500 Nodes | 1000 Nodes |
| ------------------------- | --------------- | --------- | ---------- |
| Volcano Scheduler CPU     | 2.5 cores       | 4 cores   | 5.5 cores  |
| Volcano Scheduler Memory  | 2.5 GB          | 5 GB      | 8 GB       |
| Volcano Controller CPU    | 2 cores         | 2 cores   | 2.5 cores  |
| Volcano Controller Memory | 2.5 GB          | 3 GB      | 4 GB       |

### How to Build Locally

```bash
# Build scheduler image
docker build --no-cache -t volcanosh/vc-scheduler:{tag} ./ -f Dockerfile-scheduler

# Build controller image
docker build --no-cache -t volcanosh/vc-controller-manager:{tag} ./ -f Dockerfile-controller
```

### Deploy Ascend for Volcano

1. Pull the images

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:{version}
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/volcano-controller:{version}
```

2. Retag the images

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/volcano-scheduler:{version} volcanosh/vc-scheduler:{tag}
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/volcano-controller:{version} volcanosh/vc-controller-manager:{tag}
```

3. Start Volcano

Replace `{tag}` in the YAML file with the actual image tag.

```bash
kubectl apply -f volcano-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep volcano
```

***

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

***

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
