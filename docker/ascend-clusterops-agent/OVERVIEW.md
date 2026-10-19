# Cluster Scheduling Component Atlas ClusterOps Agent

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- Atlas ClusterOps Agent is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/en/mindcluster/latest/clustersched/schedulingug/docs/en/scheduling/01_introduction/00_overview.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## Atlas ClusterOps Agent

Atlas ClusterOps Agent is one of the MindCluster cluster scheduling components. It wraps ascend-faultdiag (offline log
parsing and diagnosis) into a cluster diagnosis agent. Users only need to specify the name of a faulty job to
automatically complete the full **collection → parsing → diagnosis → reporting** workflow.

The component consists of three modules. Agent Core and Node Collector share the same Docker image (the startup entry
is selected by the `SERVICE` environment variable):

- **Agent Core**: The diagnosis controller, deployed on management nodes. It maintains the central relationship cache
  of job pods and the mount relationship table, dispatches collection instructions to each node concurrently,
  aggregates parsed artifacts, invokes `ascend-fd diag` to generate the diagnosis report, and caches results as needed.
- **Node Collector**: The node-side collector, deployed on compute nodes. It collects job logs from the host and
  statically mounted shared storage according to the collection manifest, and runs `ascend-fd parse` locally for log
  parsing.
- **Kubectl plugins** (`kubectl-ascend_diag` / `kubectl-clusterops`): The user-facing entry point, installed on the
  user machine. `ascend_diag` invokes Agent Core via port-forward; `clusterops` manages LLM configuration.

### Use Cases

When a training or inference job fails at runtime, the fault root cause must be located quickly. Jobs are typically
distributed across multiple nodes, and fault-related logs (system logs, device logs, service logs, plogs, etc.) are
scattered across the hosts of each node, making manual troubleshooting costly and inefficient. MindCluster provides the
cluster ops agent feature: a single `kubectl ascend_diag` command triggers job-level diagnosis and automatically
completes node log collection, log parsing, and centralized diagnosis, helping users quickly locate the fault root
cause.

### Features

- One-click job-level diagnosis that automatically schedules node log collection, local parsing, and centralized
  diagnosis.
- Maintains the central relationship cache of job pods and the mount relationship table, supporting mount path lookup
  for job logs such as plogs.
- Caches diagnosis results after each run; repeated diagnoses return the cache directly, with the `--refresh` parameter
  to force a refresh.
- Optionally integrates an LLM service to intelligently summarize diagnosis reports and produce root cause reports.

### Upstream and Downstream Dependencies

1. Agent Core obtains job requests from the K8s API Server.
2. Agent Core obtains job cache information, mount pairs, and env information of training/inference jobs from
   ConfigMaps.
3. Agent Core obtains mount information, env information, and metadata from training/inference job containers.
4. Agent Core sends collection instructions to Node Collector.
5. Node Collector returns the parsing results to Agent Core.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow the format below:

```text
<version>
```

| Field     | Example   | Description                          |
|-----------|-----------|--------------------------------------|
| `version` | `v26.2.0` | Version Number of Atlas ClusterOps Agent |

> [!NOTE]
> Unlike components such as ClusterD, this image is built on the unified python:3.13-slim base image (the component
> whls are pure Python packages and architecture-independent), so the tag does not contain an OS field.

### Atlas ClusterOps Agent Latest Version 26.2.0

The following are all images of the latest released 26.2.0 version of Atlas ClusterOps Agent. For all historical
version Tags, please refer to
[supported_tags.md](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-clusterops-agent/supported_tags.md).

| Tag       | Dockerfile                                                                                                                              | Image Content                                                       |
|-----------|-----------------------------------------------------------------------------------------------------------------------------------------|---------------------------------------------------------------------|
| `v26.2.0` | [Dockerfile](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/ascend-clusterops-agent/v26.2.0/Dockerfile) | Atlas ClusterOps Agent v26.2.0 (Base Image: python 3.13-slim) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software    | Supported Versions                          | Installation Location | Description                                                       |
|-------------|---------------------------------------------|-----------------------|-------------------------------------------------------------------|
| Kubernetes  | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes             | See [Kubernetes Documentation](https://kubernetes.io/docs/)        |
| kubectl     | Matching the Kubernetes version             | User machine          | Required by the kubectl plugins                                   |
| socat       | -                                           | User machine          | Ensures `kubectl port-forward` works properly                     |

#### Hardware Requirements

| Resource | Management Node (Agent Core) | Compute Node (Node Collector) |
|----------|------------------------------|-------------------------------|
| CPU      | 2 cores                      | 4 cores                       |
| Memory   | 2 GB                         | 4 GB                          |

### Obtain the Atlas ClusterOps Agent Image Online

1. Pull the official image

   Pull the Atlas ClusterOps Agent image from AscendHub, replacing {tag} with the actual version.

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-clusterops-agent:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-clusterops-agent:{tag} ascend-clusterops-agent:{tag}
   ```

### Build Locally (Optional)

Example: build an Atlas ClusterOps Agent image of architecture linux-aarch64, version v26.2.0.

1. Obtain the Dockerfile

   Navigate to the [Supported Tags and Dockerfile Links](#supported-tags-and-dockerfile-links) section, open the
   Dockerfile link corresponding to your target version, and save the file to a local directory on your aarch64
   environment.

2. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t ascend-clusterops-agent:v26.2.0 ./
   ```

> [!NOTE]
>
> - If your Docker version is earlier than 18.09 or BuildKit is not manually enabled, the TARGETPLATFORM variable
>   cannot be read during image building, which will cause the image build to fail. Run the following command before
>   building to enable BuildKit temporarily:
>
>   ```bash
>   export DOCKER_BUILDKIT=1
>   ```
>
> - The build process requires internet access: the Dockerfile automatically downloads the component package
>   (`Ascend-mindxdl-ascend-clusterops-agent_{version}_linux.zip`) and the log parsing dependency package
>   (`Ascend-mindxdl-faultdiag_{version}_linux-{arch}.zip`) from the MindCluster releases, selects the wheel matching
>   the target architecture, and pip inside the image automatically installs runtime dependencies such as grpcio and
>   fastapi.

### Deploy Atlas ClusterOps Agent

Before deployment, replace the image `{tag}` in the YAML files with the actual image version.

1. Deploy Agent Core (Deployment, on management nodes).
2. Deploy Node Collector (DaemonSet, one per NPU node).
3. Install the kubectl plugins (on the user machine).

For detailed installation steps, see
[MindCluster Cluster Scheduling Component - Manual Installation](https://gitcode.com/Ascend/mind-cluster/blob/master/docs/en/scheduling/03_installation_guide/02_installation/01_manual_installation.md).

Verify the deployment:

```bash
kubectl get pods -A | grep -E "agent-core|node-collector"
```

Expected result: the `agent-core` Pod and the `node-collector` Pods on each compute node are in Running state.

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in
these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their
respective license agreements.
