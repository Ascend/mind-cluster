# Cluster Scheduling Component Atlas Operator

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Atlas Operator is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/2600/clustersched/dlug/docs/zh/scheduling/introduction.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Atlas Operator

Atlas Operator is one of the MindCluster cluster scheduling components, deployed on management nodes. It supports distributed training on Kubernetes using the MindSpore and PyTorch AI frameworks. Through the CRD (Custom Resource Definition) mechanism, AscendJob tasks are defined, allowing users to easily implement distributed training by simply configuring YAML files.

### Use Cases

MindCluster provides the Atlas Operator component to input information required for collective communication, including the master process IP, RankTable information for static network collective communication, and the rankId of the current Pod.

### Features

- Creates Pods and injects collective communication parameters via environment variables.
- Creates RankTable files and mounts them to containers via shared storage or ConfigMap, optimizing collective communication link establishment performance.

### Upstream and Downstream Dependencies

1. Perceives whether the resources required by the current job are satisfied through Volcano.
2. After resources are satisfied, creates the corresponding Pods for the job and injects collective communication parameters as environment variables.
3. After Pod creation, Volcano performs the final resource selection.
4. Obtains the chip ID, IP, and rankId information of the job from Atlas Device Plugin, aggregates them, and generates the collective communication file.
5. Mounts the collective communication file into the container via shared storage or ConfigMap.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Starting from version v26.1.0, tags follow the format below:

```text
<version>-<os>
```

| Field     | Example       | Description                                |
|-----------|---------------|--------------------------------------------|
| `version` | `v26.1.0`     | Version Number of Atlas Operator           |
| `os`      | `ubuntu22.04` | Operating System for Atlas Operator Images |

### Atlas Operator 26.1.0

| Tag                      | Dockerfile                                                                                                                      | Image Content                                        |
|--------------------------|---------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-operator/v26.1.0/Dockerfile.ubuntu)       | Atlas Operator v26.1.0 (Base Image: Ubuntu 22.04)    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-operator/v26.1.0/Dockerfile.openeuler) | Atlas Operator v26.1.0 (Base Image: openEuler 24.03) |

---

Tags for v26.0.0 and earlier versions follow the format below:

```text
<version>
```

| Field     | Example   | Description                      |
|-----------|-----------|----------------------------------|
| `version` | `v26.0.0` | Version Number of Atlas Operator |

### Atlas Operator 26.0.0

| Tag       | Dockerfile                                                                                                    | Image Content                         |
|-----------|---------------------------------------------------------------------------------------------------------------|---------------------------------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/ascend-operator/build/Dockerfile) | Atlas Operator v26.0.0 (Ubuntu 22.04) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| Volcano | See [Volcano Kubernetes compatibility](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility) | Management nodes | Atlas Operator depends on Volcano for resource scheduling |

#### Hardware Requirements

| Resource | Requirement |
|---|---|
| CPU | 2 cores |
| Memory | 2.5 GB |

### Obtain Atlas Operator Image Online

1. Pull the official image

   Pull the Atlas Operator image from AscendHub, replacing {tag} with the actual version (v26.0.0 recommended).

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:{tag} ascend-operator:{tag}
   ```

### Build Locally (Optional)

#### Local Build Steps for v26.1.0 and Later Versions

Example: build an Atlas Operator image of architecture linux-aarch64, version v26.1.0, based on Ubuntu 22.04.

1. Obtain the target Dockerfile

   Navigate to the chapter [Supported Tags and Dockerfile Links](#Supported-Tags-and-Dockerfile-Links), open the
   Dockerfile.ubuntu link corresponding to your target version, and save the file to a local directory on your aarch64
   environment.

2. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t ascend-operator:v26.1.0 ./ -f Dockerfile.ubuntu
   ```

> **Important Notes**
> If your Docker version is earlier than 18.09 or BuildKit is not manually enabled, the TARGETPLATFORM variable cannot
> be read during image building, which will cause the image build to fail.
> 1. TARGETPLATFORM is a built-in global variable of Docker BuildKit for identifying the target build platform, e.g.
     linux/amd64, linux/arm64.
> 2. This variable is automatically injected only after BuildKit is enabled. It cannot be used in legacy Docker
     environments or environments where BuildKit is disabled by default.
> 3. Run the following command before building to enable BuildKit temporarily:
> ```bash
> export DOCKER_BUILDKIT=1
> ```

#### Local Build Steps for v26.0.0 and Earlier Versions

Example: Build an Atlas Operator image of architecture linux-aarch64, version v26.0.0, based on Ubuntu 22.04.

1. Download the officially released component package

   ```shell
   wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip
   ```

2. Extract the package to a custom directory

   ```shell
   unzip Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip -d Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64
   ```

3. Enter the extracted working directory

   ```shell
   cd Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64
   ```

4. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t ascend-operator:v26.0.0 ./ -f Dockerfile
   ```

### Deploy Atlas Operator

1. Start Atlas Operator

   Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

   ```bash
   kubectl apply -f ascend-operator-{version}.yaml
   ```

2. Verify deployment

   ```bash
   kubectl get pods -A | grep ascend-operator
   ```

   Expected result: The ascend-operator related Pods in the corresponding namespace should be in Running state.

---

## Supported Hardware

For descriptions of currently supported Atlas hardware models, please refer to the official documentation:
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
