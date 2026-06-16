# Cluster Scheduling Component NPU Exporter

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- NPU Exporter is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## NPU Exporter

NPU Exporter is one of the MindCluster cluster scheduling components, deployed on compute nodes. It reports various chip data metrics and supports two monitoring integration methods: Prometheus and Telegraf.

### Use Cases

During job execution, in addition to chip faults, it is often necessary to monitor chip network and computing utilization to identify performance bottlenecks and find directions for improving job performance. MindCluster provides the NPU Exporter component deployed on compute nodes for reporting various chip data metrics.

### Features

- Retrieves various chip and network data metrics from the driver.
- Adapts Prometheus hook functions, providing standard interfaces for Prometheus services to call.
- Adapts Telegraf hook functions, providing standard interfaces for Telegraf services to call.
- Supports real-time monitoring of Atlas AI processor utilization, temperature, voltage, memory, and other data metrics.
- Supports monitoring of virtual NPU (vNPU) AI Core utilization, vNPU total memory, and vNPU used memory.
- Supports custom metric development. Users can refer to the provided demo to develop custom metric plugins.

### Upstream and Downstream Dependencies

1. Retrieves chip and network information from the driver and stores it in a local cache.
2. Retrieves container information from the Kubernetes standard CRI interface and stores it in a local cache.
3. Implements Prometheus or Telegraf interfaces for them to periodically retrieve cached data metrics.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```
<version>
```

| Field | Example | Description |
|---|---|---|
| `version` | `v26.0.0` | NPU Exporter version |

### NPU Exporter 26.0.0

| Tag | Dockerfile | Image Content |
|-----|------------|---------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/npu-exporter/build/Dockerfile) | NPU Exporter v26.0.0 (Ubuntu 22.04) |

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Prometheus | Latest stable version recommended | Monitoring nodes | NPU Exporter adapts Prometheus hook functions to provide monitoring data |
| Atlas AI Processor Driver and Firmware | See version compatibility table | Compute nodes | See "Installing NPU Driver and Firmware" in the CANN Software Installation Guide |
| UMDK software package | See version compatibility table | Compute nodes | Necessary for Atlas 850、Atlas 950 SuperPod Products |

#### Hardware Requirements

| Resource | Requirement |
|---|---|
| CPU | 1 core |
| Memory | 1 GB |

### Obtain NPU Exporter Image Online

1. Pull the official image

Pull the NPU Exporter image from AscendHub, replacing {tag} with the actual version (v26.0.0 recommended).

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag}
```

2. Retag the image

Retag the official image with a local tag for consistent naming and easier operations management.

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/npu-exporter:{tag} npu-exporter:{tag}
```

### Build Locally (Optional)

The following example uses linux-aarch64 architecture and v26.0.0 version:

1. Download the officially released component package

```shell
wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-npu-exporter_26.0.0_linux-aarch64.zip
```

2. Extract the package to a custom directory

```shell
unzip Ascend-mindxdl-npu-exporter_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-npu-exporter_26.0.0_linux-aarch64
```

3. Enter the extracted working directory

```shell
cd Ascend-mindxdl-npu-exporter_26.0.0_linux-aarch64
```

4. Build the Docker image locally (disable cache to ensure a clean build)

```bash
docker build --no-cache -t npu-exporter:v26.0.0 ./ -f Dockerfile
```

### Deploy NPU Exporter

1. Label Kubernetes nodes

Add labels to the corresponding nodes for cluster scheduling matching. Replace <node-name> with the actual node name.

```bash
kubectl label nodes <node-name> workerselector=dls-worker-node
```

2. Start NPU Exporter

Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

```bash
kubectl apply -f npu-exporter-{version}.yaml
```

3. Verify deployment

```bash
kubectl get pods -A | grep npu-exporter
```

Expected result: The npu-exporter related Pods in the corresponding namespace should be in Running state.

4. Access monitoring metrics

```bash
curl http://<pod-ip>:8082/metrics
```

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
