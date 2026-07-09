# Cluster Scheduling Component Atlas Device Plugin

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Atlas Device Plugin is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Atlas Device Plugin

Atlas Device Plugin is one of the core components of the MindCluster cluster scheduling suite, deployed on compute nodes to provide resource discovery and reporting strategies tailored for Atlas devices.

### Use Cases

Kubernetes needs to be aware of resource information for scheduling. Beyond basic CPU and memory information, the Kubernetes device plugin mechanism allows users to define custom resource types and customize resource discovery and reporting strategies. MindCluster provides the Atlas Device Plugin service deployed on compute nodes to offer resource discovery and reporting strategies suitable for Atlas devices.

### Features

- **Device Discovery**: Obtains chip type and model information from the driver and reports it to kubelet and the upper-level ClusterD service. Supports discovering the number of devices from the Atlas device driver and reporting the count to the Kubernetes system. Supports discovering virtual devices split from physical devices and reporting them to the Kubernetes system.

- **Health Check**: Subscribes to chip fault information from the driver, reports chip status to kubelet, and reports chip status along with specific fault details to the upper-level scheduling service. Supports detecting the health status of Atlas devices. When a device is in an unhealthy state, it is reported to the Kubernetes system, which automatically removes the unhealthy device from the available list. The health status of virtual devices is determined by the physical devices from which they are split.

- **Device Allocation**: Supports allocating Atlas devices in the Kubernetes system. Supports NPU device rescheduling — when a device fails, a new container is automatically started, a healthy device is mounted, and the training task is rebuilt. During the resource mounting phase, it retrieves the chip information selected by the cluster scheduler and passes it to Atlas Docker Runtime via environment variables for mounting.

- **Fault Handling**: Configurable fault handling levels, with the ability to escalate fault handling levels when faults recur or persist for extended periods. If a faulty chip is idle and can recover after a restart, a hot reset is performed on the chip.

- **Network Fault Monitoring**: Subscribes to Lingqu network fault information from the Lingqu driver, reports network status to kubelet, and reports Lingqu network status along with specific fault details to the upper-level scheduling service.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```text
<version>
```

| Field | Example | Description |
|---|---|---|
| `version` | `v26.0.0` | Atlas Device Plugin version |

### Atlas Device Plugin 26.0.0

| Tag | Dockerfile | Image Content |
|-----|------------|---------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/ascend-device-plugin/build/Dockerfile) | Atlas Device Plugin v26.0.0 (Ubuntu 22.04) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Docker | 18.09.x~28.5.1 | All nodes | Available from [Docker](https://docs.docker.com/engine/install/) |
| Containerd | 1.4.x~2.1.4 (1.6.x recommended) | All nodes | Available from [Containerd](https://containerd.io/downloads/) |
| Atlas AI Processor Driver and Firmware | See version compatibility table | Compute nodes | See "Installing NPU Driver and Firmware" in the CANN Software Installation Guide |
| UMDK software package                  | See version compatibility table | Compute nodes | Necessary for Atlas 850、Atlas 950 SuperPod Products                              |

#### Hardware Requirements

| Resource | Requirement |
|---|---|
| CPU | 0.5 cores |
| Memory | 0.5 GB |

#### Install Driver

The host machine must have the driver and firmware installed. For details, see "Installing NPU Driver and Firmware" in the [CANN Software Installation Guide (Commercial Edition)](https://www.hiascend.com/document/detail/en/canncommercial/850/softwareinst/instg/instg_0005.html?Mode=PmIns&InstallType=local&OS=Debian).

### Obtain Atlas Device Plugin Image Online

1. Pull the official image

   Pull the Atlas Device Plugin image from AscendHub, replacing {tag} with the actual version (v26.0.0 recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:{tag} ascend-k8sdeviceplugin:{tag}
   ```

### Build Locally (Optional)

The following example uses the linux-aarch64 architecture and v26.0.0 version to demonstrate the complete local image build steps:

1. Download the officially released component package

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip
   ```

2. Extract the package to a custom directory

   ```shell
   unzip Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64
   ```

3. Enter the extracted working directory

   ```shell
   cd Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64
   ```

4. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t ascend-k8sdeviceplugin:v26.0.0 ./ -f Dockerfile
   ```

### Deploy Atlas Device Plugin

1. Label Kubernetes nodes

   Label nodes according to the Atlas processor model for cluster scheduling. Replace `<node-name>` with the actual node name.

   ```bash
   # Example: Label an Atlas 910 node
   kubectl label nodes <node-name> accelerator=huawei-Ascend910
   ```

2. Start Atlas Device Plugin

   Select the appropriate YAML resource file based on the device model and scheduling requirements. Replace `{tag}` in the YAML file with the actual image version.

   ```bash
   # Configuration file for products excluding Atlas 200I SoC A1 core board without Volcano.
   kubectl apply -f device-plugin-{version}.yaml

   # Configuration file for products excluding Atlas 200I SoC A1 core board with Volcano.
   kubectl apply -f device-plugin-volcano-{version}.yaml
   ```

3. Verify deployment

   ```bash
   kubectl get pods -A | grep device-plugin
   ```

   Expected result: The device-plugin related Pods in the corresponding namespace should be in Running state.

4. Check node resources

   ```bash
   kubectl describe node <npu-node-name> | grep "huawei.com/Ascend"
   ```

   Expected result: The huawei.com/Ascend resource capacity and allocatable resources should be displayed correctly.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
