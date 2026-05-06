# Cluster Scheduling Component Infer Operator

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.1.0-ubuntu22.04
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:v26.1.0-openeuler24.03
```

---

## Infer Operator

Infer Operator is a MindCluster cluster scheduling component deployed on management nodes. It is a Kubernetes Operator used to deploy and manage multi-role collaborative inference tasks. Infer Operator defines three CRDs — InferServiceSet, InferService, and InstanceSet — and implements controllers for these three resource types to reconcile their instance states.

### Use Cases

MindCluster provides the Infer Operator component to launch inference services based on instance configuration and supports manual scaling of inference instances.

### Features

- Creates inference instance Workloads and Services.
- Supports manual scaling of inference instances.

### Upstream and Downstream Dependencies

1. Creates inference instance Workloads based on user-configured task YAML.
2. After the Workload Controller creates Pods, Volcano performs the final resource selection.
3. If the Workload requests NPU cards, Ascend Device Plugin obtains NPU information and completes device mounting.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```shell
<version>-<os>
```

| Field | Example | Description |
| -- | -- | -- |
| `version` | `v26.1.0` | Infer Operator component version |
| `os` | `ubuntu22.04` | Infer Operator image operating system |

### Infer Operator 26.1.0

| Tag | Dockerfile | Image Content |
| --- | ----------- | -------- |
| `v26.1.0-ubuntu22.04` | [Dockerfile.ubuntu](v26.1.0/Dockerfile.ubuntu) | Infer Operator v26.1.0 image for Ubuntu 22.04 |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](v26.1.0/Dockerfile.openeuler) | Infer Operator v26.1.0 image for openEuler 24.03 |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
| -- | -- | -- | -- |
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Volcano | See [Volcano Kubernetes compatibility](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | Management nodes | Infer Operator depends on Volcano for resource scheduling |
| Ascend Device Plugin | Same version as Infer Operator | Compute nodes | Required when inference tasks use NPU |

#### Hardware Requirements

| Resource | Requirement |
| -- | -- |
| CPU | 2 cores |
| Memory | 2 GB |

### How to Build Locally

```bash
docker build --no-cache -t infer-operator:{tag} ./ -f Dockerfile.{os}
```

### Deploy Infer Operator

1. Pull the image

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag}
```

2. Retag the image

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag} infer-operator:{tag}
```

3. Start Infer Operator

Replace `{tag}` in the infer-operator-{version}.yaml file with the actual image tag.

```bash
kubectl apply -f infer-operator-{version}.yaml
```

4. Verify deployment

```bash
kubectl get pods -A | grep infer-operator
```

---

## Supported Hardware

For descriptions of supported Ascend hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/document/detail/en/mindcluster/600/clustersched/introduction/schedulingsd/mxdlug_005.html) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
