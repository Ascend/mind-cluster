# K8s RDMA Shared Device Plugin

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- K8s RDMA Shared Device Plugin is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## K8s RDMA Shared Device Plugin

K8s RDMA Shared Device Plugin is a Kubernetes device plugin for managing RDMA devices in a shared manner. It enables containers to share RDMA devices, providing high-performance networking for distributed applications.

### Use Cases

When running distributed training or high-performance computing workloads that require RDMA (Remote Direct Memory Access), the K8s RDMA Shared Device Plugin allows multiple containers to share RDMA devices efficiently.

### Features

- Manages RDMA devices on Kubernetes nodes
- Supports device sharing among multiple containers
- Provides device selection based on vendor, device ID, driver, and interface name
- Integrates with Kubernetes device plugin framework
- Supports Container Device Interface (CDI)

### Upstream and Downstream Dependencies

1. Detects RDMA devices on compute nodes
2. Registers with Kubernetes kubelet device plugin framework
3. Reports device availability to Kubernetes scheduler

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | K8s RDMA Shared Device Plugin component version |
| `os` | `ubuntu22.04` | Image operating system |

### K8s RDMA Shared Device Plugin 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | K8s RDMA Shared Device Plugin v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | K8s RDMA Shared Device Plugin v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| RDMA Drivers | OFED 5.6 or later | Compute nodes | RDMA device drivers |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 0.1 cores |
| Memory | 0.1 GB |

### How to Build Locally

```bash
docker build --no-cache -t k8s-rdma-shared-dev-plugin:{tag} ./ -f Dockerfile.{os}
```

### Deploy K8s RDMA Shared Device Plugin

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dev-plugin:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/k8s-rdma-shared-dev-plugin:{tag} k8s-rdma-shared-dev-plugin:{version}
```

3. Create configuration file

Create a ConfigMap with the device plugin configuration.

4. Deploy using DaemonSet

```bash
kubectl apply -f k8s-rdma-shared-dev-plugin-{version}.yaml
```

5. Verify deployment

```bash
kubectl get pods -A | grep k8s-rdma-shared-dev-plugin
```

---

## Configuration

The K8s RDMA Shared Device Plugin can be configured with the following parameters:

| Parameter | Type | Description | Default |
| -- | -- | -- | -- |
| `periodicUpdateInterval` | int | Interval (seconds) for periodic device updates | 0 (disabled) |
| `configList` | array | List of device configurations | [] |
| `resourceName` | string | Resource name for the device plugin | rdma |
| `resourcePrefix` | string | Resource prefix | huawei.com |
| `rdmaHcaMax` | int | Maximum number of RDMA HCA devices | 1000 |
| `devices` | array | List of device names to include | [] |
| `selectors.buses` | array | Bus types to filter devices (e.g., "ub" to enable UB devices) | [] |
| `selectors.vendors` | array | Vendor IDs to filter devices | [] |
| `selectors.deviceIDs` | array | Device IDs to filter devices | [] |
| `selectors.drivers` | array | Driver names to filter devices | [] |
| `selectors.ifNames` | array | Interface names to filter devices | [] |
| `selectors.linkTypes` | array | Link types to filter devices | [] |

---

## Supported Hardware

PCI and UB type DPU network cards

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
