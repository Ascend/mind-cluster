# Cluster Scheduling Component NodeD

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- NodeD is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## NodeD

NodeD is one of the MindCluster cluster scheduling components, deployed on compute nodes. It detects abnormal node states, retrieves CPU, memory, and disk fault information from IPMI, and reports it to ClusterD.

### Use Cases

When certain faults occur in a node's CPU, memory, or disk, training jobs will fail. To enable training jobs to exit quickly when node faults occur and prevent new jobs from being scheduled to faulty nodes, MindCluster provides the NodeD component for detecting node abnormalities.

### Features

- Retrieves node abnormalities from IPMI and reports them to the upper-level resource scheduling service.
- Periodically sends node fault information to the upper-level resource scheduling service.

### Upstream and Downstream Dependencies

1. Retrieves CPU, memory, and disk fault information of compute nodes from IPMI.
2. Reports CPU, memory, and disk fault information of compute nodes to ClusterD.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```
<version>
```

| Field | Example | Description |
|---|---|---|
| `version` | `v26.0.0` | NodeD version |

### NodeD 26.0.0

| Tag | Dockerfile | Image Content |
|-----|------------|---------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/noded/build/Dockerfile) | NodeD v26.0.0 (Ubuntu 22.04) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| ClusterD | Same version as NodeD | Management nodes | Fault information reported by NodeD is aggregated and processed by ClusterD |

#### Hardware Requirements

| Resource | Requirement |
|---|---|
| CPU | 0.5 cores |
| Memory | 0.3 GB |

### Obtain NodeD Image Online

1. Pull the official image

Pull the NodeD image from AscendHub, replacing {tag} with the actual version (v26.0.0 recommended).

```bash
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag}
```

2. Retag the image

Retag the official image with a local tag for consistent naming and easier operations management.

```bash
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:{tag} noded:{tag}
```

### Build Locally (Optional)

The following example uses linux-aarch64 architecture and v26.0.0 version:

1. Download the officially released component package

```shell
wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip
```

2. Extract the package to a custom directory

```shell
unzip Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-noded_26.0.0_linux-aarch64
```

3. Enter the extracted working directory

```shell
cd Ascend-mindxdl-noded_26.0.0_linux-aarch64
```

4. Build the Docker image locally (disable cache to ensure a clean build)

```bash
docker build --no-cache -t noded:v26.0.0 ./ -f Dockerfile
```

### Deploy NodeD

1. Label Kubernetes nodes

Add labels to the corresponding nodes for cluster scheduling matching. Replace <node-name> with the actual node name.

```bash
kubectl label nodes <node-name> workerselector=dls-worker-node
```

2. Start NodeD

Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

```bash
kubectl apply -f noded-{version}.yaml
```

3. Verify deployment

```bash
kubectl get pods -A | grep noded
```

Expected result: The noded related Pods in the corresponding namespace should be in Running state.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/introduction.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
