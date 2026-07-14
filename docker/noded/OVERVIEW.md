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

Starting from version v26.1.0, tags follow the format below:

```text
<version>-<os>
```

| Field     | Example       | Description                       |
|-----------|---------------|-----------------------------------|
| `version` | `v26.1.0`     | Version Number of NodeD           |
| `os`      | `ubuntu22.04` | Operating System for NodeD Images |

### NodeD 26.1.0

| Tag                      | Dockerfile                                                                                                            | Image Content                               |
|--------------------------|-----------------------------------------------------------------------------------------------------------------------|---------------------------------------------|
| `v26.1.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/noded/v26.1.0/Dockerfile.ubuntu)       | NodeD v26.1.0 (Base Image: Ubuntu 22.04)    |
| `v26.1.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/noded/v26.1.0/Dockerfile.openeuler) | NodeD v26.1.0 (Base Image: openEuler 24.03) |

---

Tags for v26.0.0 and earlier versions follow the format below:

```text
<version>
```

| Field     | Example   | Description             |
|-----------|-----------|-------------------------|
| `version` | `v26.0.0` | Version Number of NodeD |

### NodeD 26.0.0

| Tag       | Dockerfile                                                                                          | Image Content                            |
|-----------|-----------------------------------------------------------------------------------------------------|------------------------------------------|
| `v26.0.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/v26.0.0/component/noded/build/Dockerfile) | NodeD v26.0.0 (Base Image: Ubuntu 22.04) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software | Supported Versions | Installation Location | Description |
|---|---|---|---|
| Kubernetes | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes | See [Kubernetes Documentation](https://kubernetes.io/docs/) |
| ClusterD | Same version as NodeD | Management nodes | Fault information reported by NodeD is aggregated and processed by ClusterD |
| UMDK software package| See version compatibility table | Compute nodes | Necessary for Atlas 850、Atlas 950 SuperPod Products |

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

#### Local Build Steps for v26.1.0 and Later Versions

Example: build an NodeD image of architecture linux-aarch64, version v26.1.0, based on Ubuntu 22.04.

1. Obtain the target Dockerfile

   Navigate to the chapter [Supported Tags and Dockerfile Links](#Supported-Tags-and-Dockerfile-Links), open the
   Dockerfile.ubuntu link corresponding to your target version, and save the file to a local directory on your aarch64
   environment.

2. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t noded:v26.1.0 ./ -f Dockerfile.ubuntu
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

Example: Build an NodeD image of architecture linux-aarch64, version v26.0.0, based on Ubuntu 22.04.

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

   Add labels to the corresponding nodes for cluster scheduling matching. Replace `<node-name>` with the actual node name.

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
[Supported Product Formats and OS List](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/zh/scheduling/01_introduction/03_supported_product_models_and_os.md#%E6%94%AF%E6%8C%81%E7%9A%84%E4%BA%A7%E5%93%81%E5%BD%A2%E6%80%81%E5%92%8Cos%E6%B8%85%E5%8D%95)

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their respective license agreements.
