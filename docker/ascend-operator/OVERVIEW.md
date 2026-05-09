# Cluster Scheduling Component Ascend Operator

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Ascend Operator is maintained by [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Code Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Ascend Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Ascend Operator

Ascend Operator is a MindCluster cluster scheduling component deployed on management nodes. It supports distributed training on Kubernetes using three AI frameworks: MindSpore, PyTorch, and TensorFlow. The CRD (Custom Resource Definition) defines the AscendJob task type, allowing users to easily implement distributed training by simply configuring YAML files.

### Use Cases

MindCluster provides the Ascend Operator component to inject information required for collective communication, including the master process IP, RankTable information for static network topology-based collective communication, and the rankId of the current Pod.

### Features

- Creates Pods and injects collective communication parameters as environment variables.
- Creates RankTable files and mounts them to containers via shared storage or ConfigMap to optimize collective communication link setup performance.

### Upstream and Downstream Dependencies

1. Uses Volcano to determine whether the resources required by the current task are available.
2. Once resources are available, creates the corresponding Pods for the task and injects collective communication parameters as environment variables.
3. After Pod creation, Volcano performs the final resource selection.
4. Obtains chip ID, IP, and rankId information from Ascend Device Plugin, aggregates them, and generates the collective communication file.
5. Mounts the collective communication file into the container via shared storage or ConfigMap.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | Ascend Operator component version |
| `os` | `ubuntu22.04` | Ascend Operator image operating system |

### Ascend Operator 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Ascend Operator v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Ascend Operator v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Volcano | See [Volcano Kubernetes compatibility](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | Management nodes | Ascend Operator depends on Volcano for resource scheduling |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 2 cores |
| Memory | 2.5 GB |

### How to Build Locally

```bash
docker build --no-cache -t ascend-k8soperator:{tag} ./ -f Dockerfile.{os}
```

### Deploy Ascend Operator

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:{tag} ascend-k8soperator:{tag}
```

3. Start Ascend Operator

Replace `{tag}` in the ascend-operator-{version}.yaml file with the actual image tag.

```bash
kubectl apply -f ascend-operator-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep ascend-operator
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
