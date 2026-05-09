# Cluster Scheduling Component ClusterD

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- ClusterD is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## ClusterD

ClusterD is a MindCluster cluster scheduling component deployed on management nodes. It collects and aggregates cluster task, resource, and fault information along with their impact scope, performs statistical analysis from task, chip, and fault dimensions, and uniformly determines fault handling levels and strategies.

### Use Cases

A single node may experience multiple faults. If each node handles faults independently, tasks may be subject to conflicting recovery strategies simultaneously. To coordinate fault handling levels, MindCluster provides the ClusterD service deployed on management nodes. ClusterD collects and aggregates cluster task, resource, and fault information along with their impact scope, performs statistical analysis from task, chip, and fault dimensions, and uniformly determines fault handling levels and strategies.

### Features

- Obtains chip, node, and network information from Ascend Device Plugin and NodeD components, and retrieves public fault information from ConfigMap or gRPC.
- Aggregates the above fault information for upper-level cluster scheduling services to query.
- Establishes connections with training containers to control training processes for recomputation.
- Interacts with out-of-band services to transmit task information.

### Upstream and Downstream Dependencies

1. Obtains chip information from Ascend Device Plugin on each compute node.
2. Obtains CPU, memory, and disk health status information, DPC shared storage fault information, and Lingqu network fault information from NodeD on each compute node.
3. Retrieves public fault information from ConfigMap or gRPC.
4. Aggregates resource information across the entire cluster and reports it to Ascend-volcano-plugin.
5. Monitors cluster task information and reports task status, resource usage, and other information to CCAE.
6. Interacts with in-container processes to control training processes for recomputation.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | ClusterD component version |
| `os` | `ubuntu22.04` | ClusterD image operating system |

### ClusterD 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | ClusterD v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | ClusterD v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Ascend Device Plugin | Same version as ClusterD | Compute nodes | ClusterD depends on Ascend Device Plugin to report chip information |
| NodeD | Same version as ClusterD | Compute nodes | ClusterD depends on NodeD to report node fault information |

#### Hardware Requirements

| Resource | Up to 100 Nodes | 500 Nodes | 1000 Nodes |
| -- | -- | -- | -- |
| CPU | 1 core | 2 cores | 4 cores |
| Memory | 1 GB | 2 GB | 8 GB |

### How to Build Locally

```bash
docker build --no-cache -t ascend-k8sclusterd:{tag} ./ -f Dockerfile.{os}
```

### Deploy ClusterD

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:{tag} ascend-k8sclusterd:{tag}
```

3. Start ClusterD

Replace `{tag}` in the clusterd-{version}.yaml file with the actual image tag.

```bash
kubectl apply -f clusterd-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep clusterd
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
