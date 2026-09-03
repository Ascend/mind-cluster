# K8s RDMA Shared Device Plugin

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- K8s RDMA Shared Device Plugin is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
        - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## K8s RDMA Shared Device Plugin

K8s RDMA Shared Device Plugin is a Kubernetes device plugin for managing RDMA devices in a shared manner. It enables
containers to share RDMA devices, providing high-performance networking for distributed applications.

### Use Cases

When running distributed training or high-performance computing workloads that require RDMA (Remote Direct Memory
Access), the K8s RDMA Shared Device Plugin allows multiple containers to share RDMA devices efficiently.

### Features

- Manages RDMA devices on Kubernetes nodes, supporting UB device types
- Supports device sharing among multiple containers
- Provides device selection based on bus, vendor, device ID, driver, and interface name
- Supports shared and exclusive work modes; in exclusive mode, devices are allocated based on the NPU-to-DPU mapping
- Integrates with Kubernetes device plugin framework
- Supports UB device fault detection and writes fault information to a ConfigMap
- Writes DPU resource annotations to nodes and device status annotations to Pods

### Upstream and Downstream Dependencies

1. Detects RDMA devices on compute nodes and performs periodic fault detection
2. Registers with Kubernetes kubelet device plugin framework
3. Reports device availability to Kubernetes scheduler
4. Writes fault detection information to Kubernetes as a ConfigMap
5. Writes the DPU resource annotation `huawei.com/dpu.resource.name` to nodes
6. In exclusive mode, allocates devices to Pods based on the NPU-to-DPU mapping and writes the result to the `k8s.v1.cni.cncf.io/device-status` annotation for Multus CNI

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field     | Example       | Description                                                         |
|-----------|---------------|---------------------------------------------------------------------|
| `version` | `v26.1.0`     | Version Number of K8s RDMA Shared Device Plugin Component           |
| `os`      | `ubuntu22.04` | Operating System for K8s RDMA Shared Device Plugin Component Images |

### K8s RDMA Shared Device Plugin 26.1.0

| Tag                      | Dockerfile                                                                                                                                 | Image Content                                                       |
|--------------------------|--------------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.ubuntu)       | K8s RDMA Shared Device Plugin v26.1.0 (Base Image: Ubuntu 22.04)    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.openeuler) | K8s RDMA Shared Device Plugin v26.1.0 (Base Image: openEuler 24.03) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software     | Supported Versions                          | Installation Location | Description                                                 |
|--------------|---------------------------------------------|-----------------------|-------------------------------------------------------------|
| Kubernetes   | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes             | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| RDMA Drivers | OFED 5.6 or later                           | Compute nodes         | RDMA device drivers                                         |

#### Hardware Requirements

| Resource | Requirement |
|----------|-------------|
| CPU      | 0.1 cores   |
| Memory   | 0.1 GB      |

### Build Locally (Optional)

Example: build an K8s RDMA Shared Device Plugin image of architecture linux-aarch64, version v26.1.0, based on Ubuntu
22.04.

1. Obtain the target Dockerfile

   Navigate to the chapter Supported Tags and Dockerfile Links (for example,
   [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/k8s-rdma-shared-dev-plugin/v26.1.0/Dockerfile.ubuntu)),
   open the Dockerfile.ubuntu link corresponding to your target version, and save the file to a local directory on your
   aarch64 environment.

2. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t k8s-rdma-shared-dev-plugin:v26.1.0 ./ -f Dockerfile.ubuntu
   ```

> [!NOTE]
>
> If your Docker version is earlier than 18.09 or BuildKit is not manually enabled, the TARGETPLATFORM variable cannot
> be read during image building, which will cause the image build to fail.
>
> 1. TARGETPLATFORM is a built-in global variable of Docker BuildKit for identifying the target build platform, e.g.
     linux/amd64, linux/arm64.
> 2. This variable is automatically injected only after BuildKit is enabled. It cannot be used in legacy Docker
     environments or environments where BuildKit is disabled by default.
> 3. Run the following command before building to enable BuildKit temporarily:
>
> ```bash
> export DOCKER_BUILDKIT=1
> ```

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

### Startup Parameters

The K8s RDMA Shared Device Plugin supports the following startup parameters:

| Parameter           | Description                                                                                | Default                                                           |
|---------------------|--------------------------------------------------------------------------------------------|-------------------------------------------------------------------|
| `--config-file`     | Path to the device plugin config file                                                      | `/k8s-rdma-shared-dev-plugin/config.json`                         |
| `--use-cdi`         | Use CDI to expose devices in containers (UB devices do not support CDI)                    | `false`                                                           |
| `--ub-excl-mode`    | Enable exclusive mode for UB devices (default: shared mode)                                | `false`                                                           |
| `--logLevel`        | Log level: `-1` debug, `0` info, `1` warning, `2` error, `3` critical                      | `0`                                                               |
| `--maxBackups`      | Maximum number of backup log files, range `(0, 180]`                                       | `3`                                                               |
| `--maxAge`          | Number of days to retain backup log files, range `[7, 700]`                                | `7`                                                               |
| `--logFile`         | Log file path; the file is rotated when its size exceeds 20MB                              | `/var/log/mindx-dl/k8s-rdma-shared-dp/k8s-rdma-shared-dp.log`     |
| `--version`/`-v`    | Show the application version                                                               | -                                                                 |

### Configuration File Parameters

The K8s RDMA Shared Device Plugin can be configured with the following parameters:

| Parameter                | Type   | Description                                                                    | Default                       |
|--------------------------|--------|--------------------------------------------------------------------------------|-------------------------------|
| `periodicUpdateInterval` | int    | Interval (seconds) for periodic device updates; defaults to 60 if not set, 0 disables it | 60                            |
| `faultDetectPeriod`      | int    | Periodic fault detection interval (seconds); not set disables detection, minimum is 1 | 0 (disabled)                  |
| `configList`             | array  | List of device configurations                                                    | []                            |
| `resourceName`           | string | Resource name for the device plugin                                             | rdma                          |
| `resourcePrefix`         | string | Resource prefix                                                                 | rdma-ub     |
| `rdmaHcaMax`             | int    | Maximum number of RDMA HCA devices                                              | 1000                          |
| `devices`                | array  | List of device names to include (deprecated, use `selectors.ifNames` instead)   | []                            |
| `selectors.buses`        | array  | Bus types to filter devices (e.g., `ub` to enable UB devices)                   | []                            |
| `selectors.vendors`      | array  | Vendor IDs to filter devices                                                    | []                            |
| `selectors.deviceIDs`    | array  | Device IDs to filter devices                                                    | []                            |
| `selectors.drivers`      | array  | Driver names to filter devices                                                  | []                            |
| `selectors.ifNames`      | array  | Interface names to filter devices                                               | []                            |
| `selectors.linkTypes`    | array  | Link types to filter devices                                                    | []                            |

Configuration example:

```json
{
  "periodicUpdateInterval": 300,
  "faultDetectPeriod": 5,
  "configList": [
    {
      "resourceName": "rdma-ub-devices",
      "rdmaHcaMax": 1000,
      "selectors": {
        "buses": ["ub"]
      }
    }
  ]
}
```

### Work Modes

- **Shared mode (default)**: All Pods in the cluster share all RDMA devices on the node. In shared mode, virtual devices are created based on the `rdmaHcaMax` parameter and reported to kubelet; Pods requesting the resource can use any device.
- **Exclusive mode**: Each NPU corresponds to a specific DPU device, and devices are allocated through the NPU-NIC mapping. In exclusive mode, the real UB devices discovered on the node are reported. When a Pod requests resources, the plugin looks up the corresponding DPU device via the mapping file (`/etc/rdma-plugin/npu-nic-mapping.json`) based on the NPUs requested by the Pod, and mounts it into the container, achieving device-level isolation.
  - **Enablement**: Start the component with the `--ub-excl-mode` parameter to enable exclusive mode.
  - **Allocation result**: In exclusive mode, the plugin writes the allocated DPU devices to the Pod's `k8s.v1.cni.cncf.io/device-status` annotation. Multus CNI reads this annotation to obtain the devices and passes them to UB Host Device CNI to complete device mounting.

---

## Supported Hardware

Supports UB type RDMA network cards

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in
these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their
respective license agreements.
