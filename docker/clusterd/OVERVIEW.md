# Cluster Scheduling Component ClusterD

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- ClusterD is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## ClusterD

ClusterD is one of the MindCluster cluster scheduling components, deployed on management nodes. It collects and aggregates cluster job, resource, and fault information along with their impact scope, performs statistical analysis from the dimensions of jobs, chips, and faults, and uniformly determines fault handling levels and policies.

### Use Cases

A single node may experience multiple faults. If each node handles faults independently, jobs may simultaneously be subject to multiple recovery strategies. To coordinate job handling levels, MindCluster provides the ClusterD service deployed on management nodes. ClusterD collects and aggregates cluster job, resource, and fault information along with their impact scope, performs statistical analysis from the dimensions of jobs, chips, and faults, and uniformly determines fault handling levels and policies.

### Features

- Obtains chip, node, and network information from Atlas Device Plugin and NodeD components, and retrieves public fault information from ConfigMap or gRPC.
- Aggregates the above fault information for upper-level cluster scheduling services to query.
- Establishes connections with training containers to control training processes for recomputation actions.
- Interacts with out-of-band services to transmit job information.

### Upstream and Downstream Dependencies

1. Obtains chip information from Atlas Device Plugin on each compute node.
2. Obtains CPU, memory, and disk health status information, DPC shared storage fault information, and Lingqu network fault information from NodeD on each compute node.
3. Retrieves public fault information from ConfigMap or gRPC.
4. Aggregates resource information across the entire cluster and reports it to Ascend-volcano-plugin.
5. Monitors cluster job information and reports job status, resource usage, and other information to CCAE.
6. Interacts with in-container processes to control training processes for recomputation.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```text
<version>
```

| Field | Example | Description |
|---|---|---|
| `version` | `v26.0.0` | ClusterD version |

### ClusterD 26.0.0

| Tag | Dockerfile | Image Content |
|-----|------------|---------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/clusterd/build/Dockerfile) | ClusterD v26.0.0 (Ubuntu 22.04) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Atlas Device Plugin | Same version as ClusterD | Compute nodes | ClusterD depends on Atlas Device Plugin to report chip information |
| NodeD | Same version as ClusterD | Compute nodes | ClusterD depends on NodeD to report node fault information |

#### Hardware Requirements

| Resource | Up to 100 Nodes | 500 Nodes | 1000 Nodes |
|---|---|---|---|
| CPU | 1 core | 2 cores | 4 cores |
| Memory | 1 GB | 2 GB | 8 GB |

### Obtain ClusterD Image Online

1. Pull the official image

   Pull the ClusterD image from AscendHub, replacing {tag} with the actual version (v26.0.0 recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:{tag} clusterd:{tag}
   ```

### Build Locally (Optional)

The following example uses linux-aarch64 architecture and v26.0.0 version:

1. Download the officially released component package

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip
   ```

2. Extract the package to a custom directory

   ```shell
   unzip Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-clusterd_26.0.0_linux-aarch64
   ```

3. Enter the extracted working directory

   ```shell
   cd Ascend-mindxdl-clusterd_26.0.0_linux-aarch64
   ```

4. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t clusterd:v26.0.0 ./ -f Dockerfile
   ```

### Deploy ClusterD

1. Start ClusterD

   Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

   ```bash
   kubectl apply -f clusterd-{version}.yaml
   ```

2. Verify deployment

   ```bash
   kubectl get pods -A | grep clusterd
   ```

   Expected result: The clusterd related Pods in the corresponding namespace should be in Running state.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
