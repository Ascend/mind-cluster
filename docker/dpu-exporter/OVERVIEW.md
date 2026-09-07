# Cluster Scheduling Component DPU Exporter

> English | [中文](./OVERVIEW.zh.md)

## Quick Reference

- DPU Exporter is maintained by [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
- Where to get help
    - [MindCluster Repository](https://gitcode.com/Ascend/mind-cluster)
    - [MindCluster Atlas Community](https://www.hiascend.com/document/detail/zh/mindcluster/latest/clustersched/dlug/docs/zh/scheduling/01_introduction/00_overview.md)
    - [Issue Tracker](https://gitcode.com/Ascend/mind-cluster/issues)

---

## DPU Exporter

### Use Cases

During job execution, the health status of the DPU directly affects job stability. MindCluster provides the DPU
Exporter component to monitor the running status and statistics metrics of the DPU.

### Features

- Obtains the running status and statistics metrics of the DPU from the NIC management tool and file interfaces.
- Provides a Prometheus metrics interface for monitoring the running status and statistics metrics of the DPU.

### Upstream and Downstream Dependencies

1. Obtains DPU global metrics and interface-level metrics from the NIC management tool and file interfaces
   respectively.
2. Converts the obtained metrics into Prometheus metric format.
3. Provides a Prometheus metrics interface for monitoring the running status and statistics metrics of the DPU.

---

## Supported Tags and Dockerfile Links

### Tag Convention

Tags follow the format below:

```text
<version>-<os>
```

| Field     | Example       | Description                              |
|-----------|---------------|------------------------------------------|
| `version` | `v26.2.0`     | Version Number of DPU Exporter           |
| `os`      | `ubuntu22.04` | Operating System for DPU Exporter Images |

### DPU Exporter 26.2.0

| Tag                      | Dockerfile                                                                                                                   | Image Content                                      |
|--------------------------|------------------------------------------------------------------------------------------------------------------------------|----------------------------------------------------|
| `v26.2.0-ubuntu22.04`    | [Dockerfile.ubuntu](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/dpu-exporter/v26.2.0/Dockerfile.ubuntu)       | DPU Exporter v26.2.0 (Base Image: Ubuntu 22.04)    |
| `v26.2.0-openeuler24.03` | [Dockerfile.openeuler](https://gitcode.com/Ascend/mind-cluster/blob/master/docker/dpu-exporter/v26.2.0/Dockerfile.openeuler) | DPU Exporter v26.2.0 (Base Image: openEuler 24.03) |

---

## Quick Start

### Prerequisites

#### Software Dependencies

| Software                           | Supported Versions                          | Installation Location | Description                                                                      |
|------------------------------------|---------------------------------------------|-----------------------|----------------------------------------------------------------------------------|
| Kubernetes                         | 1.17.x~1.34.x (1.19.x or later recommended) | All nodes             | See [Kubernetes Documentation](https://kubernetes.io/docs/)                      |
| Prometheus                          | Latest stable version recommended           | Monitoring nodes      | DPU Exporter adapts Prometheus hook functions to provide monitoring data         |
| DPU Driver, Firmware and hinicadm5  | See version compatibility table             | Compute nodes         | hinicadm5 is the management tool delivered with the DPU driver package           |

#### Hardware Requirements

| Resource | Requirement |
|----------|-------------|
| CPU      | 1 core      |
| Memory   | 512 MB      |

### Obtain DPU Exporter Image Online

1. Pull the official image

   Pull the DPU Exporter image from AscendHub, replacing {tag} with the actual version.

   ```bash
   docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/dpu-exporter:{tag}
   ```

2. Retag the image

   Retag the official image with a local tag for consistent naming and easier operations management.

   ```bash
   docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/dpu-exporter:{tag} dpu-exporter:{tag}
   ```

### Build Locally (Optional)

Example: build a DPU Exporter image of architecture linux-aarch64, version v26.2.0, based on Ubuntu 22.04.

1. Obtain the build artifacts

   Run `build.sh` in the `component/dpu-exporter/build` directory of the source repository. The `output` directory
   will contain the `dpu-exporter` binary, `config.json`, and the deployment YAML file.

2. Obtain the target Dockerfile

   Navigate to the chapter Supported Tags and Dockerfile Links, open the Dockerfile link corresponding to your
   target version, and save the file together with the build artifacts to a local directory on your aarch64
   environment.

3. Build the Docker image locally (disable cache to ensure a clean build)

   ```bash
   docker build --no-cache -t dpu-exporter:v26.2.0 ./ -f Dockerfile
   ```

### Deploy DPU Exporter

1. Label Kubernetes nodes

   Add labels to the corresponding nodes for cluster scheduling matching. Replace `<node-name>` with the actual node
   name.

   ```bash
   kubectl label nodes <node-name> workerselector=dls-worker-node
   ```

2. Start DPU Exporter

   Before deployment, replace the image `{tag}` in the YAML file with the actual image version.

   ```bash
   kubectl apply -f dpu-exporter-{version}.yaml
   ```

3. Verify deployment

   ```bash
   kubectl get pods -A | grep dpu-exporter
   ```

   Expected result: The dpu-exporter related Pods in the corresponding namespace should be in Running state.

4. Access monitoring metrics

   ```bash
   curl http://<pod-ip>:8080/metrics
   ```

---

## License

View the [license information](https://www.hiascend.com/en/legal/softlicense) for the Mind series software contained in
these images.

As with all container images, pre-installed software packages (Python, system libraries, etc.) may be subject to their
respective license agreements.
