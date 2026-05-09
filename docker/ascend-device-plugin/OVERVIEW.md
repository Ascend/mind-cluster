# Cluster Scheduling Component Ascend Device Plugin

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Ascend Device Plugin is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend Device Plugin

Ascend Device Plugin is one of the core components of the MindCluster cluster scheduling suite, deployed on compute nodes to provide resource discovery and reporting strategies tailored for Ascend devices.

### Use Cases

Kubernetes needs to be aware of resource information for scheduling. Beyond basic CPU and memory information, the Kubernetes device plugin mechanism allows users to define custom resource types and customize resource discovery and reporting strategies. MindCluster provides the Ascend Device Plugin service deployed on compute nodes to offer resource discovery and reporting strategies suitable for Ascend devices.

### Features

- **Device Discovery**: Obtains chip type and model information from the driver and reports it to kubelet and the upper-level ClusterD service. Supports discovering the number of devices from the Ascend device driver and reporting the count to the Kubernetes system. Supports discovering virtual devices split from physical devices and reporting them to the Kubernetes system.

- **Health Check**: Subscribes to chip fault information from the driver, reports chip status to kubelet, and reports chip status along with specific fault details to the upper-level scheduling service. Supports detecting the health status of Ascend devices. When a device is in an unhealthy state, it is reported to the Kubernetes system, which automatically removes the unhealthy device from the available list. The health status of virtual devices is determined by the physical devices from which they are split.

- **Device Allocation**: Supports allocating Ascend devices in the Kubernetes system. Supports NPU device rescheduling — when a device fails, a new container is automatically started, a healthy device is mounted, and the training task is rebuilt. During the resource mounting phase, it retrieves the chip information selected by the cluster scheduler and passes it to Ascend Docker Runtime via environment variables for mounting.

- **Fault Handling**: Configurable fault handling levels, with the ability to escalate fault handling levels when faults recur or persist for extended periods. If a faulty chip is idle and can recover after a restart, a hot reset is performed on the chip.

- **Network Fault Monitoring**: Subscribes to Lingqu network fault information from the Lingqu driver, reports network status to kubelet, and reports Lingqu network status along with specific fault details to the upper-level scheduling service.

### Upstream and Downstream Dependencies

1. Obtains chip type, quantity, and health status information from DCMI, or issues chip reset commands.
2. Reports chip type, quantity, and status to kubelet.
3. Reports chip type, quantity, and specific fault information to ClusterD.
4. Passes the chip information selected by the scheduler to Ascend Docker Runtime via environment variables.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | Ascend Device Plugin component version |
| `os` | `ubuntu22.04` | Ascend Device Plugin image operating system |

### Ascend Device Plugin 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Ascend Device Plugin v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Ascend Device Plugin v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Docker | 18.09.x~28.5.1 | All nodes | Available from [Docker](https://docs.docker.com/engine/install/) |
| Containerd | 1.4.x~2.1.4 (1.6.x recommended) | All nodes | Available from [Containerd](https://containerd.io/downloads/) |
| Ascend AI Processor Driver and Firmware | See version compatibility table | Compute nodes | See "Installing NPU Driver and Firmware" in the CANN Software Installation Guide |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 0.5 cores |
| Memory | 0.5 GB |

#### Install Driver

The host machine must have the driver and firmware installed. For details, see "Installing NPU Driver and Firmware" in the [CANN Software Installation Guide (Commercial Edition)](https://www.hiascend.com/document/detail/en/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian) or [CANN Software Installation Guide (Community Edition)](https://www.hiascend.com/document/detail/en/CANNCommunityEdition/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=openEuler).

### How to Build Locally

```bash
docker build --no-cache -t ascend-k8sdeviceplugin:{tag} ./ -f Dockerfile.{os}
```

### Deploy Ascend Device Plugin

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{tag} ascend-k8sdeviceplugin:{tag}
```

3. Label the nodes

Label nodes according to the Ascend processor model:

```bash
# Example: Label an Ascend 910 node
kubectl label nodes <node-name> accelerator=huawei-Ascend910
kubectl label nodes <node-name> host-arch=huawei-arm
```

4. Start Ascend Device Plugin

Select the appropriate YAML file based on the Ascend processor model:

Replace `{tag}` in the YAML file with the actual image tag.

```bash
# Configuration file for inference servers (with Atlas 300I inference cards) without Volcano.
kubectl apply -f device-plugin-310-{version}.yaml

# Configuration file for inference servers (with Atlas 300I inference cards) with Volcano.
kubectl apply -f device-plugin-310-volcano-{version}.yaml

# Configuration file for Atlas inference series products (excluding Atlas 200I SoC A1 core board) without Volcano.
kubectl apply -f device-plugin-310P-{version}.yaml

# Configuration file for Atlas inference series products (excluding Atlas 200I SoC A1 core board) with Volcano.
kubectl apply -f device-plugin-310P-volcano-{version}.yaml

# Configuration file for Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas 800I A2 inference servers, or A200I A2 Box heterogeneous components without Volcano.
kubectl apply -f device-plugin-910-{version}.yaml

# Configuration file for Atlas training series products, Atlas A2 training series products, Atlas A3 training series products, Atlas 800I A2 inference servers, or A200I A2 Box heterogeneous components with Volcano.
kubectl apply -f device-plugin-910-volcano-{version}.yaml

# Configuration file for Atlas 350 standard cards, Atlas 850 series hardware products, and Atlas 950 SuperPoD without Volcano.
kubectl apply -f device-plugin-npu-{version}.yaml

# Configuration file for Atlas 350 standard cards, Atlas 850 series hardware products, and Atlas 950 SuperPoD with Volcano.
kubectl apply -f device-plugin-npu-volcano-{version}.yaml
```

5. Verify deployment

```bash
kubectl get pods -A | grep device-plugin
```

6. Check node resources

```bash
kubectl describe node <npu-node-name> | grep "huawei.com/Ascend"
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
