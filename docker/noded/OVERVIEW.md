# Cluster Scheduling Component NodeD

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.1.0-ubuntu22.04
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.1.0-openeuler24.03
```

---

## NodeD

NodeD is a MindCluster cluster scheduling component deployed on compute nodes. It detects node abnormal states, retrieves CPU, memory, and disk fault information from IPMI, and reports it to ClusterD.

### Use Cases

When a node's CPU, memory, or disk experiences certain faults, training tasks will fail. To allow training tasks to exit quickly when a node fault occurs and prevent new tasks from being scheduled to faulty nodes, MindCluster provides the NodeD component for detecting node abnormalities.

### Features

- Retrieves node abnormalities from IPMI and reports them to the upper-level scheduling service.
- Periodically sends node fault information to the upper-level scheduling service.

### Upstream and Downstream Dependencies

1. Retrieves CPU, memory, and disk fault information from IPMI on compute nodes.
2. Reports CPU, memory, and disk fault information of compute nodes to ClusterD.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | NodeD component version |
| `os` | `ubuntu22.04` | NodeD image operating system |

### NodeD 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | NodeD v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | NodeD v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| ClusterD | Same version as NodeD | Management nodes | Fault information reported by NodeD is aggregated by ClusterD |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 0.5 cores |
| Memory | 0.3 GB |

### How to Build Locally

```bash
docker build --no-cache -t noded:{tag} ./ -f Dockerfile.{os}
```

### Deploy NodeD

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag} noded:{version}
```

3. Start NodeD

Replace `{tag}` in the noded-{version}.yaml file with the actual image tag.

```bash
kubectl apply -f noded-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep noded
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
