# Quick Start<a name="ZH-CN_TOPIC_0000002511346939"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:16:19.835Z pushedAt=2026-06-09T06:22:06.954Z -->

This document provides two quick start scenarios to help you quickly get started with cluster scheduling powered by Ascend NPUs.

- **10-Minute Quick Start**: Deploy only Ascend Device Plugin, use the Kubernetes native scheduler to schedule regular pods, and quickly verify NPU resource scheduling capabilities. This is suitable for beginners who want a quick experience.
- **E2E Training Service Quick Start**: Deploy all the cluster scheduling components (NodeD, Ascend Device Plugin, Ascend Docker Runtime, Volcano, ClusterD, and Ascend Operator). This scenario takes a PyTorch training job as an example to describe the E2E distributed training process.

You can choose the appropriate entry path based on your actual needs.

## Environment Preparation<a name="section159013591917"></a>

Ensure that the cluster environment has been set up.

- Kubernetes has been installed on all nodes, with supported versions 1.17.x~1.34.x. If you need to install Volcano, install Kubernetes version 1.19.x or later. For specific Kubernetes versions, see [the corresponding Kubernetes versions on the Volcano official website](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility). To obtain the software package, see the [Kubernetes community](https://kubernetes.io/docs/setup/).
- Docker has been installed on all nodes, with supported versions 18.09.x~28.5.1. To obtain the software package, see the [Docker community or official website](https://docs.docker.com/engine/install/).
- The corresponding firmware and drivers have been installed on all nodes. For the firmware and driver installation steps for the Atlas 800T A2 training server, see the [Atlas A2 Center Inference and Training Hardware 26.0.RC1 NPU Driver and Firmware Installation Guide](https://support.huawei.com/enterprise/en/doc/EDOC1100568434/426cffd9).
- Check whether [npu-smi](https://support.huawei.com/enterprise/en/doc/EDOC1100568421/426cffd9) and [hccn_tool](https://support.huawei.com/enterprise/en/doc/EDOC1100568362/426cffd9) tools can run normally on the host.

    >[!NOTE]
    >
    >- Refer to the [Ascend Training Solution Version Mapping](https://support.huawei.com/enterprise/en/ascend-computing/ascend-training-solution-pid-258915853/software) to confirm whether the firmware and driver versions are compatible with the cluster scheduling components.
    >- The NPU driver and firmware versions can be queried using the `npu-smi info -t board -i NPU ID` command. In the example output, the `Software Version` field indicates the NPU driver version, and the `Firmware Version` field indicates the NPU firmware version.
    >- In the following text, `{xxx}` takes the value `910` as the chip model.

## 10-Minute Quick Start

### Overview

This tutorial guides you through setting up the most simplified Ascend NPU cluster scheduling environment in 10 minutes, using only:

- **Ascend Device Plugin**: Responsible for NPU device discovery and resource reporting.
- **Kubernetes native scheduler**: No additional scheduling components required.
- **Regular pod**: Responsible for verifying the NPU scheduling capability.

### Environment Requirements

| Requirement | Description                |
|------|-------------------|
| Compute node | Atlas 800T A2 training server (Arm64) as an example    |
| Driver version | Ascend driver matching the server  |

### Pre-check

Ensure that the NPU driver is correctly installed:

```shell
# Check NPU status. The chip information will be displayed in the command output
npu-smi info
```

### Adding Labels to NPU Nodes

```shell
# Get the node name
kubectl get nodes

# Add necessary labels to NPU nodes (replace worker01 with the actual node name)
kubectl label nodes worker01 workerselector=dls-worker-node
kubectl label nodes worker01 accelerator=huawei-Ascend910
```

### Deploying Ascend Device Plugin

#### 1. Pull the Ascend Device Plugin image

```shell
# Pull the Ascend Device Plugin image from the Huawei Cloud image repository
docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0

# Add a local label to the image
docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0
```

#### 2 Deploy Ascend Device Plugin

```shell
# Pull Configuration File
mkdir /tmp/devicePlugin
cd /tmp/devicePlugin
wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip
unzip Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip

# Deploy Ascend Device Plugin
kubectl apply -f device-plugin-910-v26.0.0.yaml
```

#### 3 Verify the Deployment

```shell
# Check the Pod Status
kubectl get pod -n kube-system

# Expected Output
NAME                                  READY   STATUS    RESTARTS   AGE
...
ascend-device-plugin-daemonset-d5ctz  1/1     Running   0          11s
...
```

### Verifying NPU Resources

```shell
# View the NPU resources of the node
kubectl describe node worker01 | grep -A 10 "huawei.com/Ascend910"

# Expected output (showing the number of available NPUs)
huawei.com/Ascend910:     8
huawei.com/Ascend910:     8
```

### Scheduling an NPU Pod

#### 1 Create a test pod configuration file

Create `npu-test-pod.yaml`:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: npu-test
spec:
  nodeSelector:
    workerselector: dls-worker-node
  containers:
  - name: npu-container
    image: ubuntu:22.04
    command: ["/bin/bash", "-c", "sleep 3600"]
    resources:
      limits:
        huawei.com/Ascend910: 1  # Request 1 NPU
      requests:
        huawei.com/Ascend910: 1
    volumeMounts:
    - name: ascend-driver
      mountPath: /usr/local/Ascend/driver
      readOnly: true
  volumes:
  - name: ascend-driver
    hostPath:
      path: /usr/local/Ascend/driver
```

#### 2 Deploy the test pod

```shell
kubectl apply -f npu-test-pod.yaml
```

#### 3 Verify pod scheduling

```shell
# Check pod status
kubectl get pods npu-test -o wide

# Expected output (STATUS = Running indicates successful scheduling)
NAME      READY   STATUS    RESTARTS   AGE   IP           NODE      NOMINATED NODE
npu-test  1/1     Running   0          10s   10.244.1.2   worker01  <none>
```

### Verifying NPU Access

```shell
# Enter the container to verify NPU availability
kubectl exec -it npu-test  -- /bin/bash

# Execute `npu-smi info` inside the container to check NPU information
export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:${LD_LIBRARY_PATH}
npu-smi info
```

### Cleaning Up Test Resources

```shell
# Delete the test pod
kubectl delete pod npu-test

# Delete Ascend Device Plugin (if needed)
kubectl delete -f device-plugin-910-v26.0.0.yaml
```

### FAQs

| Issue | Cause | Solution |
|------|------|---------|
| Pod remains in Pending state | Insufficient NPU resources or mismatched node labels | Check `kubectl describe pod` and node labels. |
| Ascend Device Plugin boot failure | Incorrect driver path | Check whether `/usr/local/Ascend/driver` exists. |

## E2E Training Service Quick Start

This section uses two Atlas 800T A2 training servers (one as the management node and one as the compute node) as an example to guide developers through quickly installing NodeD, Ascend Device Plugin, Ascend Docker Runtime, Volcano, ClusterD, and Ascend Operator, and using the full-NPU scheduling feature to quickly submit a training job.

### Procedure<a name="section17940333114314"></a>

**Table 1**  Key procedures

|Procedure|Description|For More Information|
|--|--|--|
|[Installing Components](#section1837511531098)|Using Atlas 800T A2 training servers as an example, this walks you through quickly installing cluster scheduling components on Ascend devices.| [Installation and Deployment](../developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md)|
|[Delivering a Training Job](#section106493419399)|Using a simple PyTorch training job as an example, this helps you quickly understand the workflow for submitting a training job.| [Basic Scheduling](../usage/basic_scheduling/00_feature_description.md)|

### Installing Components <a name="section1837511531098"></a>

The following uses an Atlas 800T A2 training server as an example. For detailed installation steps and parameter descriptions for all components, see [Installation and Deployment](../developer_guide/installation_deployment/manual_installation/00_obtaining_software_packages.md).

1. Log in to the compute node or management node as the `root` user and create the component installation directories.
    1. Run the following commands in sequence to create the installation directories on the compute node. The following directories are examples only.

        ```shell
        mkdir /tmp/noded
        mkdir /tmp/devicePlugin
        mkdir /tmp/Ascend-docker-runtime
        ```

    2. Run the following commands in sequence to create the installation directories on the management node. The following directories are examples only.

        ```shell
        mkdir /tmp/ascend-volcano
        mkdir /tmp/ascend-operator
        mkdir /tmp/clusterd
        ```

2. Download software packages with your desired architecture. The AArch64 architecture is used as an example.
    1. Run the following commands in sequence to obtain the NodeD, Ascend Device Plugin, and Ascend Docker Runtime installation packages on the compute node and decompress them.

        ```shell
        cd /tmp/noded
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-noded_26.0.0_linux-aarch64.zip

        cd /tmp/devicePlugin
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-device-plugin_26.0.0_linux-aarch64.zip

        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-docker-runtime_26.0.0_linux-aarch64.run
        ```

    2. Run the following commands in sequence on the management node to obtain the Volcano, ClusterD, and Ascend Operator installation packages.

        ```shell
        cd /tmp/ascend-volcano
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-volcano_26.0.0_linux-aarch64.zip

        cd /tmp/ascend-operator
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-ascend-operator_26.0.0_linux-aarch64.zip

        cd /tmp/clusterd
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v26.0.0/Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip
        unzip Ascend-mindxdl-clusterd_26.0.0_linux-aarch64.zip
        ```

3. Pull component images.
    1. Run the following commands in sequence to pull the component images on the compute node.

        ```shell
        cd /tmp/noded
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/noded:v26.0.0 noded:v26.0.0

        cd /tmp/devicePlugin
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v26.0.0 ascend-k8sdeviceplugin:v26.0.0
        ```

    2. Run the following commands in sequence to create component images on the management node.

        ```shell
        cd /tmp/ascend-volcano/volcano-v1.7.0
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-scheduler:v1.7.0-v26.0.0 volcanosh/vc-scheduler:v1.7.0-v26.0.0

        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/vc-controller-manager:v1.7.0-v26.0.0 volcanosh/vc-controller-manager:v1.7.0-v26.0.0

        cd /tmp/ascend-operator
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-operator:v26.0.0 ascend-operator:v26.0.0

        cd /tmp/clusterd
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/clusterd:v26.0.0 clusterd:v26.0.0
        ```

4. Create node labels.
    >[!NOTE]
    >
    >If the message "already has a value... and --overwrite is false" is displayed when you run the node label creation command, the label already exists. You can use `--overwrite` to overwrite it.
    1. Run the following command on the Kubernetes management node to query the node name.

        ```shell
        kubectl get node
        ```

        Example output:

        ```ColdFusion
        NAME       STATUS   ROLES           AGE   VERSION
        worker01   Ready    worker    23h   v1.17.3
        ```

    2. Run the following commands in sequence to create a node label for the compute node (`worker01` as an example).

        ```shell
        kubectl label nodes worker01 node-role.kubernetes.io/worker=worker
        kubectl label nodes worker01 workerselector=dls-worker-node
        kubectl label nodes worker01 host-arch=huawei-arm
        kubectl label nodes worker01 accelerator=huawei-Ascend910
        kubectl label nodes worker01 accelerator-type=module-{xxx}b-8     #Enter the chip model
        kubectl label nodes worker01 nodeDEnable=on
        ```

    3. Run the following command to create a node label for the management node (`master01` as an example).

        ```shell
        kubectl label nodes master01 masterselector=dls-master-node
        ```

5. Create a user.
    >[!NOTE]
    >
    >Sudo privileges are required for user creation.
    1. Run the following commands in sequence to create a username on the compute node.

        ```shell
        useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
        usermod -a -G HwHiAiUser hwMindX
        ```

    2. Run the following command to create a username on the management node.

        ```shell
        useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX
        ```

6. Create log directories. Custom log directories are not supported.
    >[!NOTE]
    >
    >Sudo privileges are required for log directory creation.
    1. Run the following commands in sequence to create log directories on the compute node.

        ```shell
        mkdir -m 755 /var/log/mindx-dl
        chown root:root /var/log/mindx-dl
        mkdir -m 750 /var/log/mindx-dl/devicePlugin
        chown root:root /var/log/mindx-dl/devicePlugin
        mkdir -m 750 /var/log/mindx-dl/noded
        chown hwMindX:hwMindX /var/log/mindx-dl/noded
        ```

    2. Run the following commands in sequence to create log directories on the management node.

        ```shell
        mkdir -m 755 /var/log/mindx-dl
        chown root:root /var/log/mindx-dl
        mkdir -m 750 /var/log/mindx-dl/volcano-controller
        chown hwMindX:hwMindX /var/log/mindx-dl/volcano-controller
        mkdir -m 750 /var/log/mindx-dl/volcano-scheduler
        chown hwMindX:hwMindX /var/log/mindx-dl/volcano-scheduler
        mkdir -m 750 /var/log/mindx-dl/ascend-operator
        chown hwMindX:hwMindX /var/log/mindx-dl/ascend-operator
        mkdir -m 750 /var/log/mindx-dl/clusterd
        chown hwMindX:hwMindX /var/log/mindx-dl/clusterd
        ```

7. Run the following command on any node to create the namespace.

    ```shell
    kubectl create ns mindx-dl
    ```

8. Install components.
    1. Run the following commands in sequence to install Ascend Docker Runtime on the host of the compute node.

        ```shell
        cd /tmp/Ascend-docker-runtime
        chmod u+x Ascend-docker-runtime_26.0.0_linux-aarch64.run
        ./Ascend-docker-runtime_26.0.0_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```

    2. On the compute node, run the following commands in sequence to install components.

        ```shell
        cd /tmp/noded
        kubectl apply -f noded-v26.0.0.yaml

        cd /tmp/devicePlugin
        kubectl apply -f device-plugin-volcano-v26.0.0.yaml
        ```

    3. On the *management node, run the following commands in sequence to install components.

        ```shell
        cd /tmp/ascend-operator
        kubectl apply -f ascend-operator-v26.0.0.yaml

        cd /tmp/ascend-volcano/volcano-v1.7.0  # If you are using Volcano version 1.9.0, change it to v1.9.0.
        kubectl apply -f volcano-v1.7.0.yaml

        cd /tmp/clusterd
        kubectl apply -f clusterd-v26.0.0.yaml
        ```

    4. Run the following command to check whether the components have started successfully.

        ```shell
        kubectl get pod -A
        ```

        Taking ClusterD as an example, the example output is as follows. `Running` indicates that the component has started successfully.

        ```ColdFusion
        NAME                              READY   STATUS    RESTARTS   AGE
        ...
        clusterd-fd6t8                       1/1     Running   0          74s
        ...
        ```

### Delivering a Training Job<a name="section106493419399"></a>

1. Prepare an image.

    Download the ascend-pytorch training image (24.0.X) from the [Ascend Image Repository](https://www.hiascend.com/developer/ascendhub) according to the system architecture (Arm/x86_64). Modify the training base image by changing the default user in the container to `root`. The image does not contain training scripts, code, or other files. During training, files such as training scripts and code are typically mapped into the container using the mount method.

2. Perform script adaptation.
    1. <a name="zh-cn_topic_0000001558834814_li1298552813512"></a>Download "ResNet50_ID4149_for_PyTorch" from the master branch of the [PyTorch Code Repository](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch) as the training code.
    2. Prepare the dataset corresponding to ResNet-50 on your own, and comply with the corresponding specifications when using it.
    3. The administrator uploads the dataset to the storage node. Go to the `/data/atlas_dls/public` directory and upload the dataset to any location, such as `/data/atlas_dls/public/dataset/resnet50/imagenet`.

        ```shell
        root@ubuntu:/data/atlas_dls/public/dataset/resnet50/imagenet# pwd
        ```

    4. Decompress the training code downloaded in [Step 1](#zh-cn_topic_0000001558834814_li1298552813512) to the local machine, and upload the `ModelZoo-PyTorch/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch` directory from the decompressed training code to the environment, for example, to the `/data/atlas_dls/public/code/` path.
    5. In the `/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch` path, comment out the following code in `main.py`.

        ```Python
        def main():
            args = parser.parse_args()
            os.environ['MASTER_ADDR'] = args.addr
            #os.environ['MASTER_PORT'] = '29501'  # Comment out this line of code
            if os.getenv('ALLOW_FP32', False) and os.getenv('ALLOW_HF32', False):
                raise RuntimeError('ALLOW_FP32 and ALLOW_HF32 cannot be set at the same time!')
            elif os.getenv('ALLOW_HF32', False):
                torch.npu.conv.allow_hf32 = True
            elif os.getenv('ALLOW_FP32', False):
                torch.npu.conv.allow_hf32 = False
                torch.npu.matmul.allow_hf32 = False
        ```

    6. Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository, switch to the corresponding version branch according to *mindcluster-deploy Open-Source Repository Version Description*, obtain the `train_start.sh` file from the `samples/train/basic-training/without-ranktable/pytorch` directory, and construct the following directory structure under the `/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts` path.

        ```text
        root@ubuntu:/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts#
        scripts/
             ├── train_start.sh
        ```

3. Prepare the job YAML.
    1. Go to the [mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy) repository. Based on the *mindcluster-deploy Open-Source Repository Version Description*, switch to the corresponding version branch and obtain the `pytorch_standalone_acjob_<i>\{xxx\}</i>.yaml` file from the `samples/train/basic-training/without-ranktable/pytorch` directory (*{xxx}* indicates the chip model). The example defaults to a single-server single-device job.
    2. Modify the example YAML and upload it to any file path after modification. For detailed descriptions of each parameter in the following YAML, see [Table 1](../api/ascend_operator.md).

        ```Yaml
        apiVersion: mindxdl.gitee.com/v1
        kind: AscendJob
        ...
        spec:
        ...
          replicaSpecs:
            Master:
        ...
                spec:
                  nodeSelector:
                    host-arch: huawei-arm
                    accelerator-type: module-{xxx}b-8   # Change from the original `card-{xxx}b-2` to `module-{xxx}b-8`, where `{xxx}` indicates the chip model.
                  containers:
                  - name: ascend
                    image: pytorch-test:latest     # Modify to the image name obtained in Step 1.
        ...
                    resources:
                      limits:
                        huawei.com/Ascend910: 1
                      requests:
                        huawei.com/Ascend910: 1
        ...
                  volumes:
                  - name: code
                    nfs:      #If the NFS service is not installed, change nfs to hostPath and delete server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/"
                  - name: data
                    nfs:     #If the NFS service is not installed, change nfs to hostPath and delete server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/public/dataset/"
                  - name: output
                    nfs:     #If the NFS service is not installed, change nfs to hostPath and delete server: 127.0.0.1
                      server: 127.0.0.1
                      path: "/data/atlas_dls/output/"
        ...
        ```

4. Run the following command to deliver a single-server single-device job.

    ```shell
    kubectl apply -f pytorch_standalone_acjob_{xxx}.yaml
    ```

5. Run the following command to check the pod running status.

    ```shell
    kubectl get pod --all-namespaces -o wide
    ```

    The example output is as follows. If `Running` appears, the job is running normally.

    ```ColdFusion
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE      NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          6s      192.168.244.xxx   worker01   <none>           <none>
    ```

    >[!NOTE]
    >
    >If the training job remains in the `Pending` state after being delivered, refer to [Training Job in Pending State, Cause: nodes are unavailable](https://gitcode.com/Ascend/mind-cluster/issues/352) or [Job in Pending State Due to Insufficient Resources](https://gitcode.com/Ascend/mind-cluster/issues/355) for troubleshooting.

6. View the training results.
    1. Run the following command on any node to view the training results.

        ```shell
        kubectl logs -n  Namespace Pod name
        ```

        For example:

        ```shell
        kubectl logs -n default default-test-pytorch-master-0
        ```

    2. View the training log. If the following content appears, training succeeds.

        ```ColdFusion
        [20251218-20:31:57] [MindXDL Service Log]server id is: 0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=7 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=6 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=5 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=4 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=3 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=2 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/bin/python /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py --data=/job/data/resnet50/imagenet --amp --arch=resnet50 --seed=49 -j=128 --world-size=1 --lr=1.6 --dist-backend=hccl --multiprocessing-distributed --epochs=1 --batch-size=512 --gpu=1 --multiprocessing-distributed --addr=10.106.227.104 --world-size=1 --rank=0
        /usr/local/python3.10.5/lib/python3.10/site-packages/torchvision/io/image.py:13: UserWarning: Failed to load image Python extension: ''If you don't plan on using image functionality from `torchvision.io`, you can ignore this warning. Otherwise, there might be something wrong with your environment. Did you have `libjpeg` or `libpng` installed before building `torchvision` from source?
          warn(
        [2025-12-18 20:32:02] [WARNING] [470] profiler.py: Invalid parameter export_type: None, reset it to text.
        /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py:201: UserWarning: You have chosen to seed training. This will turn on the CUDNN deterministic setting, which can slow down your training considerably! You may see unexpected behavior when restarting from checkpoints.
        warnings.warn('You have chosen to seed training. '
        /job/code/No_Rank_ResNet50_ID4149_for_PyTorch/main.py:208: UserWarning: You have chosen a specific GPU. This will completely disable data parallelism.
        warnings.warn('You have chosen a specific GPU. This will completely '
        Use GPU: 0 for training
        => creating model 'resnet50'
        ```
