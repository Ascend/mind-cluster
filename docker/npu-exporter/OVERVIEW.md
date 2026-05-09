# Cluster Scheduling Component NPU Exporter

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- NPU Exporter is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## NPU Exporter

NPU Exporter is a MindCluster cluster scheduling component deployed on compute nodes. It reports various chip data metrics and supports both Prometheus and Telegraf monitoring integration.

### Use Cases

During task execution, in addition to chip faults, it is often necessary to monitor chip network and compute utilization to identify performance bottlenecks and find directions for improving task performance. MindCluster provides the NPU Exporter component deployed on compute nodes for reporting various chip data metrics.

### Features

- Retrieves chip and network data metrics from the driver.
- Adapts Prometheus hook functions, providing standard interfaces for Prometheus service calls.
- Adapts Telegraf hook functions, providing standard interfaces for Telegraf service calls.
- Supports real-time monitoring of Ascend AI processor utilization, temperature, voltage, memory, and other data metrics.
- Supports monitoring of virtual NPU (vNPU) AI Core utilization, vNPU total memory, and vNPU used memory.
- Supports custom metric development — users can develop custom metric plugins using the provided demo.

### Upstream and Downstream Dependencies

1. Retrieves chip and network information from the driver and stores it in a local cache.
2. Retrieves container information from the K8s standard CRI interface and stores it in a local cache.
3. Implements Prometheus or Telegraf interfaces for periodic retrieval of cached data.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | NPU Exporter component version |
| `os` | `ubuntu22.04` | NPU Exporter image operating system |

### NPU Exporter 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | NPU Exporter v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | NPU Exporter v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Prometheus | Latest stable version recommended | Monitoring nodes | NPU Exporter adapts Prometheus hook functions to provide monitoring data |
| Ascend AI Processor Driver and Firmware | See version compatibility table | Compute nodes | See "Installing NPU Driver and Firmware" in the CANN Software Installation Guide |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 1 core |
| Memory | 1 GB |

### How to Build Locally

```bash
docker build --no-cache -t npu-exporter:{tag} ./ -f Dockerfile.{os}
```

### Deploy NPU Exporter

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag} npu-exporter:{tag}
```

3. Start NPU Exporter

Replace `{tag}` in the npu-exporter-{version}.yaml file with the actual image tag.

```bash
kubectl apply -f npu-exporter-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep npu-exporter
```

5. Access monitoring metrics

```bash
curl http://<pod-ip>:8082/metrics
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
