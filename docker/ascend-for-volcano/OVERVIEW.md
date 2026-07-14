# Cluster Scheduling Component Atlas for Volcano

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Atlas for Volcano is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Atlas for Volcano

Atlas for Volcano is built on the open-source Volcano scheduling plugin mechanism, adding features such as Atlas AI processor (NPU) affinity scheduling and virtual device scheduling. It is deployed on management nodes. Volcano is a scheduler component that leverages the interconnect topology and processing logic of Atlas AI processors to achieve optimal utilization, maximizing the computing performance of Atlas AI processors.

### Use Cases

The basic Kubernetes scheduler can only perform resource scheduling based on the number of Atlas chips. To achieve affinity scheduling and maximize resource utilization, it is necessary to be aware of the network connectivity between Atlas chips and select the network-optimal resources. MindCluster provides the Volcano service deployed on management nodes, offering network affinity scheduling for different Atlas devices and network topologies.

### Features

- **Available Device Calculation**: Calculates cluster available device information based on fault information and node information reported by underlying cluster scheduling components. (`self-maintain-available-card` is enabled by default. When disabled, available device information is obtained from the underlying cluster scheduling components.)
- **Optimal Resource Allocation**: Retrieves the user's desired resource quantity from the Kubernetes job object, and selects the optimal resource allocation for the job based on the cluster's device count, device types, and device network topology.
- **Fault Rescheduling**: Reschedules jobs when job resources encounter faults.
- **NPU Affinity Scheduling**: Based on the interconnect topology of Atlas AI processors, prioritizes scheduling jobs to processors within the same card, then to HCCS-interconnected processors, and finally to PCIe-interconnected processors, reducing resource fragmentation and network congestion.
- **Switch Affinity Scheduling**: Achieves optimal node utilization based on the network configuration under switches and the parameter plane network configuration. Supports various network topologies including Spine-Leaf dual-layer interconnect and single-layer switch interconnect.
- **Logical Super Node Affinity Scheduling**: Divides physical super nodes into logical super nodes based on partitioning strategies, achieving optimal node utilization.
- **Multi-level Scheduling Strategy**: Abstracts cluster resources into a multi-level structure based on the NPU network topology hierarchy, configurable via the `resource-level-config` parameter.
- **Multiple Scheduling Modes**: Supports whole-card scheduling, static vNPU scheduling, dynamic vNPU scheduling, and soft-partition scheduling.

### Affinity Policies

Based on the characteristics of Atlas 910 AI processors and resource utilization rules, the following affinity policies are established (listed by priority):

1. **HCCS Affinity Scheduling Principle**: The requested Atlas 910 AI processors must be within the same HCCS ring, prioritizing the HCCS with the most matching number of remaining available processors.
2. **Fill-First Scheduling Principle**: Prioritizes allocating AI servers that have already been assigned Atlas 910 AI processors, reducing fragmentation.
3. **Even-Remaining Priority Principle**: Prioritizes HCCS that satisfy the above conditions, then selects HCCS with an even number of remaining processors.

### Upstream and Downstream Dependencies

1. Calculates cluster resource information based on information reported by ClusterD (when using ClusterD by default).
2. Receives job launch configurations from third parties, and selects optimal node resources based on cluster resource information.
3. Passes specific resource selection information to Atlas Device Plugin on compute nodes to complete device mounting.

### Image Description

Atlas for Volcano contains two images:

- **volcano-scheduler**: Volcano scheduler image, containing the Atlas NPU affinity scheduling plugin (`volcano-npu_*.so`).
- **volcano-controller**: Volcano controller image.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Starting from Atlas NPU Scheduler Plugin v26.1.0, tags follow the format below:

```text
<component-version>-<ascend-scheduling-plugin-version>-<os>
```

| 字段                                 | 示例值            | 说明                                           |
|------------------------------------|----------------|----------------------------------------------|
| `component-version`                | `v1.7.0`       | Version Number of Volcano component          |
| `ascend-scheduling-plugin-version` | `v26.1.0`      | Version Number of Atlas NPU Scheduler Plugin |
| `os`                               | `alpinelatest` | Operating System for Volcano Images          |

### Atlas for Volcano 26.1.0（Volcano v1.12.0）

| Tag                              | Dockerfile                                                                                                                                                               | Image Content                                                                                                                 |
|----------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------|
| `v1.12.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.12.0, Base Image: Alpine latest)   |
| `v1.12.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano Controller v26.1.0 Image (based on Volcano v1.12.0, Base Image: Alpine latest)                                        |
| `v1.12.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.12.0, Base Image: openEuler 24.03) |
| `v1.12.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.12.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano Controller v26.1.0 Image (based on Volcano v1.12.0, Base Image: openEuler 24.03)                                      |

### Atlas for Volcano 26.1.0（Volcano v1.9.0）

| Tag                             | Dockerfile                                                                                                                                                              | Image Content                                                                                                                |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|
| `v1.9.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.9.0, Base Image: Alpine latest)   |
| `v1.9.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano Controller v26.1.0 Image (based on Volcano v1.9.0, Base Image: Alpine latest)                                        |
| `v1.9.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.9.0, Base Image: openEuler 24.03) |
| `v1.9.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.9.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano Controller v26.1.0 Image (based on Volcano v1.9.0, Base Image: openEuler 24.03)                                      |

### Atlas for Volcano 26.1.0（Volcano v1.7.0）

| Tag                             | Dockerfile                                                                                                                                                              | Image Content                                                                                                                |
|---------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------|
| `v1.7.0-v26.1.0-alpinelatest`   | [Dockerfile-scheduler.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-scheduler.alpine)         | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.7.0, Base Image: Alpine latest)   |
| `v1.7.0-v26.1.0-alpinelatest`   | [Dockerfile-controller.alpine](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-controller.alpine)       | Volcano Controller v26.1.0 Image (based on Volcano v1.7.0, Base Image: Alpine latest)                                        |
| `v1.7.0-v26.1.0-openeuler24.03` | [Dockerfile-scheduler.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-scheduler.openeuler)   | Volcano Scheduler v26.1.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.7.0, Base Image: openEuler 24.03) |
| `v1.7.0-v26.1.0-openeuler24.03` | [Dockerfile-controller.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-for-volcano/volcano-v1.7.0/v26.1.0/Dockerfile-controller.openeuler) | Volcano Controller v26.1.0 Image (based on Volcano v1.7.0, Base Image: openEuler 24.03)                                      |

---

Tags for Atlas NPU Scheduler Plugin v26.0.0 and earlier versions follow the format below:

```text
<component-version>-<ascend-scheduling-plugin-version>
```

| Field                              | Example   | Description                                  |
|------------------------------------|-----------|----------------------------------------------|
| `component-version`                | `v1.7.0`  | Version Number of Volcano component          |
| `ascend-scheduling-plugin-version` | `v26.0.0` | Version Number of Atlas NPU Scheduler Plugin |

### Atlas for Volcano 26.0.0 (Volcano v1.9.0)

Using linux-aarch64 architecture as an example: Atlas for Volcano package download: [Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip](https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip)

| Tag              | Dockerfile (path within package)     | Image Content                                                                                                              |
|------------------|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| `v1.9.0-v26.0.0` | volcano-v1.9.0/Dockerfile-scheduler  | Volcano Scheduler v26.0.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.9.0, Base Image: Alpine latest) |
| `v1.9.0-v26.0.0` | volcano-v1.9.0/Dockerfile-controller | Volcano Controller v26.0.0 Image (based on Volcano v1.9.0, Base Image: Alpine latest)                                      |

### Atlas for Volcano 26.0.0 (Volcano v1.7.0)

Using linux-aarch64 architecture as an example: Atlas for Volcano package download: [Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip](https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip)

| Tag              | Dockerfile (path within package)     | Image Content                                                                                                              |
|------------------|--------------------------------------|----------------------------------------------------------------------------------------------------------------------------|
| `v1.7.0-v26.0.0` | volcano-v1.7.0/Dockerfile-scheduler  | Volcano Scheduler v26.0.0 Image (Including Atlas NPU Scheduler Plugin, based on Volcano v1.7.0, Base Image: Alpine latest) |
| `v1.7.0-v26.0.0` | volcano-v1.7.0/Dockerfile-controller | Volcano Controller v26.0.0 Image (based on Volcano v1.7.0, Base Image: Alpine latest)                                      |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.19.x~1.34.x | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Atlas Device Plugin | Same version as Volcano | Compute nodes | Volcano depends on Atlas Device Plugin to report device information |
| ClusterD | Same version as Volcano | Management nodes | Volcano depends on ClusterD to aggregate cluster fault information |

#### Hardware Requirements

| Resource | Up to 100 Nodes | 500 Nodes | 1000 Nodes |
|---|---|---|---|
| Volcano Scheduler CPU | 2.5 cores | 4 cores | 5.5 cores |
| Volcano Scheduler Memory | 2.5 GB | 5 GB | 8 GB |
| Volcano Controller CPU | 2 cores | 2 cores | 2.5 cores |
| Volcano Controller Memory | 2.5 GB | 3 GB | 4 GB |

### Obtain Atlas for Volcano Images Online

1. Pull images

   Pull the Atlas for Volcano images from AscendHub, replacing {tag} with the actual version tag (v1.9.0-v26.0.0 recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:{tag}
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:{tag}
   ```

2. Retag images

   Retag the official images with local tags for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:{tag} volcanosh/vc-scheduler:{tag}
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:{tag} volcanosh/vc-controller-manager:{tag}
   ```

### Build Locally (Optional)

#### Local Image Build Process for Atlas NPU Scheduler Plugin v26.1.0 and Later Versions

Example: Build a Volcano v1.9.0 component image with linux-aarch64 architecture based on Alpine latest, integrated with
Atlas NPU Scheduler Plugin v26.1.0.

1. Obtain the target Dockerfile

   Navigate to the chapter [Supported Tags and Dockerfile Links](#Supported-Tags-and-Dockerfile-Links), open the
   Dockerfile-scheduler.alpine and Dockerfile-controller.alpine link corresponding to your target version, and save the
   file to a local directory on your aarch64 environment.

2. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   # Build scheduler image
   docker build --no-cache -t volcanosh/vc-scheduler:v1.9.0 ./ -f Dockerfile-scheduler.alpine

   # Build controller image
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./ -f Dockerfile-controller.alpine
   ```

> **Important Notes**
> If your Docker version is earlier than 18.09 or BuildKit is not manually enabled, the TARGETPLATFORM variable cannot
> be read during image building, which will cause the image build to fail.
> 1. TARGETPLATFORM is a built-in global variable of Docker BuildKit for identifying the target build platform, e.g.
     linux/amd64, linux/arm64.
> 2. This variable is automatically injected only after BuildKit is enabled. It cannot be used in legacy Docker
     environments or environments where BuildKit is disabled by default.
> 3. Run the following command before building to enable BuildKit temporarily:
> ```bash
> export DOCKER_BUILDKIT=1
> ```

#### Local Image Build Process for Atlas NPU Scheduler Plugin v26.0.0 and Earlier Versions

Example: Build a Volcano v1.9.0 component image based on Alpine latest with linux-aarch64 architecture, integrated with
Atlas NPU Scheduler Plugin v26.0.0.

1. Download the officially released component package

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip
   ```

2. Extract the package to a custom directory

   ```shell
   unzip Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-volcano_26.0.0_linux-aarch64
   ```

3. Enter the extracted working directory

   ```shell
   cd Ascend-mindxdl-volcano_26.0.0_linux-aarch64/volcano-v1.9.0
   ```

4. Build Docker images locally (disable cache to ensure a clean build)

   ```bash
   # Build scheduler image
   docker build --no-cache -t volcanosh/vc-scheduler:v1.9.0 ./ -f Dockerfile-scheduler

   # Build controller image
   docker build --no-cache -t volcanosh/vc-controller-manager:v1.9.0 ./ -f Dockerfile-controller
   ```

### Deploy Atlas for Volcano

1. Start Volcano

   Replace `{version}` in the YAML filename with the actual version (currently volcano v1.9.0). Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

   ```bash
   kubectl apply -f volcano-{version}.yaml
   ```

2. Verify deployment

   ```bash
   kubectl get pods -A | grep volcano
   ```

   Expected result: The volcano-related Pods in the corresponding namespace should be in Running state.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
