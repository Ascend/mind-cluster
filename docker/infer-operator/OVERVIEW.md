# Cluster Scheduling Component Infer Operator

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Infer Operator is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Infer Operator

Infer Operator is one of the MindCluster cluster scheduling components, deployed on management nodes. It is a Kubernetes Operator used to deploy and manage multi-role collaborative inference jobs. Infer Operator defines three CRDs: InferServiceSet, InferService, and InstanceSet, and implements controllers for these three resource types to reconcile their instance states.

### Use Cases

MindCluster provides the Infer Operator component to launch inference services based on instance configurations of inference services, and supports manual scaling of inference instances.

### Features

- Creates inference instance Workloads and Services.
- Supports manual scaling of inference instances.

### Upstream and Downstream Dependencies

1. Creates inference instance Workloads based on user-configured job YAML.
2. After the Workload Controller creates Pods, Volcano performs the final resource selection.
3. If the Workload requests NPU cards, Atlas Device Plugin obtains NPU information and completes device mounting.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow this format:

```text
<version>
```

| Field | Example | Description |
|---|---|---|
| `version` | `v26.0.0` | Infer Operator version |

### Infer Operator 26.1.0

| Tag | Dockerfile | Image Content |
|-----|------------|---------------|
| `v26.1.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/infer-operator/build/Dockerfile) | Infer Operator v26.1.0 (Ubuntu 22.04) |

---

## Quick Start

### Prerequisites (Optional)

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Volcano | See [Volcano Kubernetes compatibility](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | Management nodes | Infer Operator depends on Volcano for resource scheduling |
| Atlas Device Plugin | Same version as Infer Operator | Compute nodes | Required when inference jobs use NPU resources |

#### Hardware Requirements

| Resource | Requirement |
|---|---|
| CPU | 2 cores |
| Memory | 2 GB |

### Obtain Infer Operator Image Online

1. Pull the official image

   Pull the Infer Operator image from the Atlas image repository, replacing {tag} with the actual version (v26.0.0 recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/infer-operator:{tag} infer-operator:{tag}
   ```

### How to Build Locally (Optional)

The following example uses linux-aarch64 architecture and v26.0.0 version:

1. Download the officially released component package

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64.zip
   ```

2. Extract the package to a custom directory

   ```shell
   unzip Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64
   ```

3. Enter the extracted working directory

   ```shell
   cd Ascend-mindxdl-infer-operator_26.0.0_linux-aarch64
   ```

4. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t infer-operator:v26.0.0 ./ -f Dockerfile
   ```

### Deploy Infer Operator

1. Start Infer Operator

   Replace the image `{tag}` in the infer-operator-{version}.yaml file with the actual tag.

   ```bash
   kubectl apply -f infer-operator-{version}.yaml
   ```

2. Verify deployment

   ```bash
   kubectl get pods -A | grep infer-operator
   ```

   Expected result: The infer-operator related Pods in the corresponding namespace should be in Running state.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/zh/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
