# Quick Start<a name="ZH-CN_TOPIC_0000002511346939"></a>

<!-- md-trans-meta sourceCommit=cc2549abde258f50a946dbab123b7ff2035ed91e translatedAt=2026-08-27T02:28:28.575Z pushedAt=2026-08-27T06:46:31.972Z -->

This document provides two quick start scenarios to help you get started with Ascend NPU cluster scheduling:

- **10-Minute Quick Start**: Deploy only Ascend Device Plugin and Ascend Docker Runtime, use the Kubernetes native scheduler to schedule normal Pods, and quickly verify NPU resource scheduling capabilities. This is suitable for beginners to get a quick experience.
- **Training Service Quick Start**: Deploy the complete cluster scheduling components (NodeD, Ascend Device Plugin, Ascend Docker Runtime, Volcano, ClusterD, and Ascend Operator), and use a PyTorch training job as an example to experience the end-to-end training process.

You can select an appropriate quick start path based on your actual requirements.

## Environment Preparation<a name="section159013591917"></a>

The quick start examples require that the cluster environment has been set up.

- Kubernetes has been installed on all nodes, with supported versions 1.17.x to 1.34.x. (If you need to install Volcano, install Kubernetes 1.19.x or later. For the specific Kubernetes version, see [the corresponding Kubernetes version on the Volcano official website](https://github.com/volcano-sh/volcano/blob/master/README.md#kubernetes-compatibility)). To obtain the software package, see the [Kubernetes community](https://kubernetes.io/docs/concepts/).
- Docker has been installed on all nodes, with supported versions 18.09.x to 28.5.1. To obtain the software package, see the [Docker community or official website](https://docs.docker.com/engine/install/).
- The matching firmware and driver have been installed on all nodes.
- Check whether npu-smi and the hccn_tool can run properly on the host.
- Pulling images and downloading component installation packages may require a network environment. Ensure that the network is normal or prepare the relevant offline image packages and component installation packages by yourself.

  >[!NOTE]
  >
  >- See [*Ascend Training Solution Version Mapping*](https://support.huawei.com/enterprise/en/ascend-computing/ascend-training-solution-pid-258915853/software) to confirm whether the firmware and driver versions are compatible with the cluster scheduling components.
  >- The NPU driver and firmware versions can be queried by running the **npu-smi info -t board -i** <i>NPU ID</i> command. In the command output, the "Software Version" field indicates the NPU driver version, and the "Firmware Version" field indicates the NPU firmware version.

## 10-Minute Quick Start

This tutorial guides you through setting up the most simplified Ascend NPU cluster scheduling environment **within 10 minutes**, using only:

- **Ascend Device Plugin**: NPU device discovery and resource reporting
- **Ascend Docker Runtime**: Resource mounting for NPU devices and others
- **Kubernetes native scheduler**: No additional scheduling components required
- **Normal Pod**: Quickly verify NPU scheduling capability

### Install Components

The following Atlas 800T A2 training server (compute node) with an AArch64 CPU architecture as an example.

1. Check the NPU status to ensure that the NPU driver matching the server has been correctly installed.

    ```shell
    npu-smi info
    ```

   If chip information is displayed, the NPU driver has been correctly installed.

2. Add a label to the NPU node.

    Execute the following command to create a node label for the compute node.

    ```shell
    kubectl label nodes --all workerselector=dls-worker-node
    ```

   >[!NOTE]
   >
   > The `workerselector=dls-worker-node` label is used to identify compute nodes so that cluster scheduling components (such as NodeD and Ascend Device Plugin) can identify and manage NPU resources.

3. Deploy Ascend Docker Runtime and Ascend Device Plugin.

   >[!NOTE]
   >
   > The `VERSION` environment variable specifies the Ascend component version. This document uses `26.1.0` as an example. This variable must be set in each independent code block.

    1. Deploy Ascend Docker Runtime.

        ```shell
        VERSION=26.1.0
        mkdir -p /tmp/Ascend-docker-runtime
        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        chmod +x Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        echo Y | ./Ascend-docker-runtime_${VERSION}_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```

       The example output is as follows, indicating successful installation.

        ```output
        Uncompressing ascend-docker-runtime  100%
        Please read the End User License Agreement carefully. Your use of the Huawei Software
        will be deemed as your acceptance of the constraints mentioned in the Agreement.
        The full text of the EULA is available at:
        https://www.hiascend.com/en/legal/softlicense

        Do you accept the EULA to install Ascend-docker-runtime? [y/N]
        [INFO] user accepted EULA
        [INFO] installing ascend-docker-runtime
        ...
        [INFO] ascend-docker-runtime install success
        ```

    2. Pull the Ascend Device Plugin image.

        ```shell
        VERSION=26.1.0
        # Pull the Ascend Device Plugin image from the Huawei Cloud image repository.
        docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v${VERSION}

        # Add a local label to the image.
        docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-k8sdeviceplugin:v${VERSION} ascend-k8sdeviceplugin:v${VERSION}
        ```

    3. Deploy Ascend Device Plugin.

        ```shell
        VERSION=26.1.0
        # Pull the configuration file.
        mkdir -p /tmp/devicePlugin
        cd /tmp/devicePlugin
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-mindxdl-device-plugin_${VERSION}_linux-aarch64.zip
        unzip Ascend-mindxdl-device-plugin_${VERSION}_linux-aarch64.zip

        # Deploy Device Plugin. If VERSION is earlier than 26.1.0, the yaml file is device-plugin-910-v${VERSION}.yaml.
        kubectl apply -f device-plugin-v${VERSION}.yaml
        ```

       Check the status of the Ascend Device Plugin Pod.

        ```shell
        kubectl get pod -n kube-system
        ```

       The example output is as follows, indicating that the status is normal.

        ```output
        NAME                                  READY   STATUS    RESTARTS   AGE
        ...
        ascend-device-plugin-daemonset-d5ctz  1/1     Running   0          11s
        ...
        ```

    4. View the NPU resources of the node.

        ```shell
        kubectl describe node -A | grep "huawei.com/Ascend910"
        ```

       The example output is as follows, showing the number of available NPUs.

        ```output
        huawei.com/Ascend910:     8
        huawei.com/Ascend910:     8
        ```

### Scheduling NPU Pods

1. Create the test Pod configuration file `npu-test-pod.yaml`.

    ```yaml
    apiVersion: v1
    kind: Pod
    metadata:
      name: npu-test
    spec:
      containers:
      - name: npu-container
        image: ubuntu:22.04          # Test Pod image, which can be customized
        command: ["/bin/bash", "-c", "sleep 3600"]
        resources:
          limits:
            huawei.com/Ascend910: 1  # Request one NPU
          requests:
            huawei.com/Ascend910: 1
    ```

2. Deploy the test Pod.

    ```shell
    kubectl apply -f npu-test-pod.yaml
    ```

3. Verify Pod scheduling.

    ```shell
    # View the Pod status.
    kubectl get pods npu-test -o wide

    # Expected output (STATUS is Running, indicating successful scheduling)
    NAME      READY   STATUS    RESTARTS   AGE   IP           NODE      NOMINATED NODE
    npu-test  1/1     Running   0          10s   10.244.1.2   worker01  <none>
    ```

4. Verify NPU access.

    ```shell
    # Enter the container to verify NPU availability.
    kubectl exec -it npu-test  -- /bin/bash

    # Execute the npu-smi info command in the container to correctly display chip information
    export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/driver/lib64/driver:${LD_LIBRARY_PATH}
    npu-smi info
    ```

5. Clean up test resources.

    ```shell
    VERSION=26.1.0
    # Delete the test Pod
    kubectl delete pod npu-test

    # Delete Ascend Device Plugin. If VERSION is earlier than 26.1.0, the yaml file is device-plugin-910-v${VERSION}.yaml
    kubectl delete -f device-plugin-v${VERSION}.yaml
    ```

**Common Issues**

| Issue | Cause | Solution |
|------|------|---------|
| Pod remains Pending | Insufficient NPU resources or mismatched node labels | Check `kubectl describe pod` and node labels |
| Ascend Device Plugin fails to start | Incorrect driver path | Check whether `/usr/local/Ascend/driver` exists |

## Training Service Quick Start

This section still uses an Atlas 800T A2 training server with an AArch64 CPU architecture as an example to guide developers in quickly completing the installation of NodeD, Ascend Device Plugin, Ascend Docker Runtime, Volcano, ClusterD, and Ascend Operator, and in using the full-NPU scheduling feature to quickly submit a training job.

### Procedure<a name="section17940333114314"></a>

**Table 1** Key procedure description

|Procedure|Description|More References|
|--|--|--|
|[Installing Components](#section1837511531098)|Using an Atlas 800T A2 training server as an example, this section walks you through quickly installing cluster scheduling components on Ascend devices.|For more parameter descriptions and procedures for installing cluster scheduling components, see the [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md) chapter.|
|[Submitting a Training Job](#section106493419399)|Using a simple PyTorch training job as an example, this section helps you quickly understand the workflow for submitting a training job.|For more parameter descriptions and procedures for submitting training jobs, see the [Basic Scheduling](../04_usage/03_basic_scheduling/00_feature_description.md) chapter.|

### Installing Components<a name="section1837511531098"></a>

The Atlas 800T A2 training server is used as an example in the following steps. For detailed installation steps and parameter descriptions of all components, see [Installation and Deployment](../03_installation_guide/02_installation/00_helm_installation.md).

1. Create a node label.

    Execute the following command to create a node label for the compute node (for example, the node name is "worker01").

    ```shell
    kubectl label nodes worker01 node-role.kubernetes.io/worker=worker workerselector=dls-worker-node masterselector=dls-master-node --overwrite
    ```

2. Install components. Using the AArch64 architecture as an example, download the software packages that match your actual architecture.
    >[!NOTE]
    >
    >Helm is used as an example for quick deployment, which requires MindCluster version 26.1.0 or later. See [Installing with Helm](../03_installation_guide/02_installation/00_helm_installation.md).

    1. Install Ascend Docker Runtime.

        ```shell
        VERSION=26.1.0
        mkdir -p /tmp/Ascend-docker-runtime
        cd /tmp/Ascend-docker-runtime
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        chmod +x Ascend-docker-runtime_${VERSION}_linux-aarch64.run
        echo Y | ./Ascend-docker-runtime_${VERSION}_linux-aarch64.run --install
        systemctl daemon-reload && systemctl restart docker
        ```

    2. Install NodeD, Ascend Device Plugin, Volcano, ClusterD, and Ascend Operator through Helm.

        ```shell
        VERSION=26.1.0
        mkdir /tmp/helm
        cd /tmp/helm
        wget https://gitcode.com/Ascend/mind-cluster/releases/download/v${VERSION}/Ascend-helm-deploy-tool_${VERSION}_linux.zip
        unzip Ascend-helm-deploy-tool_${VERSION}_linux.zip
        helm install mindcluster-crds mindcluster-crds-deploy-tool-*.tgz
        helm install mindcluster mindcluster-deploy-tool-*.tgz
        ```

        The example output is as follows, indicating that the installation is successful.

        ```output
        Release "mindcluster-crds" does not exist. Installing it now.
        NAME: mindcluster-crds
        LAST DEPLOYED: ...
        NAMESPACE: mindx-dl
        STATUS: deployed
        REVISION: 1
        TEST SUITE: None
        ```

        ```output
        Release "mindcluster" does not exist. Installing it now.
        NAME: mindcluster
        LAST DEPLOYED: ...
        NAMESPACE: mindx-dl
        STATUS: deployed
        REVISION: 1
        TEST SUITE: None
        ```

    3. Verify whether the components run normally, using NodeD as an example.

        ```shell
        kubectl get pod -n mindx-dl
        ```

        The example output is as follows, indicating that NodeD is running normally.

        ```shell
        NAME                                  READY   STATUS    RESTARTS   AGE
        ...
        noded-694474f599-54w6b                1/1     Running   0          11s
        ...
        ```

### Submitting a Training Job<a name="section106493419399"></a>

1. Prepare the image.

   Download the ascend-pytorch training image of version 24.0.x from the [Ascend image repository](https://www.hiascend.com/en/developer/ascendhub). The image does not contain files such as training scripts and code. During training, files such as training scripts and code are usually mapped into the container by mounting.

   >[!NOTE]
   >
   > The image version used in this example is 24.0.0-A2-2.1.0. To obtain the latest image version, visit the [Ascend image repository](https://www.hiascend.com/en/developer/ascendhub) to view the list of available versions, or contact Huawei technical support to obtain version compatibility information.

    ```shell
    docker pull swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04
    docker tag swr.cn-south-1.myhuaweicloud.com/ascendhub/ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04 ascend-pytorch:24.0.0-A2-2.1.0-ubuntu20.04
    ```

2. Prepare the training job.

    1. Execute the following command to download "ResNet50_ID4149_for_PyTorch" from the master branch of the [PyTorch code repository](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch) as the training code, and decompress it to the `/data/atlas_dls/public/code/` directory.

        ```shell
        mkdir -p /data/atlas_dls/public/code/
        cd /data/atlas_dls/public/code/
        wget https://raw.gitcode.com/Ascend/ModelZoo-PyTorch/archive/refs/heads/master.zip?path=PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch -O ResNet50_ID4149_for_PyTorch.zip
        unzip ResNet50_ID4149_for_PyTorch.zip
        mv ModelZoo-PyTorch-master-PyTorch-built-in-cv-classification-ResNet50_ID4149_for_PyTorch/PyTorch/built-in/cv/classification/ResNet50_ID4149_for_PyTorch ResNet50_ID4149_for_PyTorch
        ```

    2. Execute the following command to obtain `train_start.sh` from the `samples/train/basic-training/without-ranktable/pytorch` directory of the [MindCluster-Samples](https://gitcode.com/Ascend/mindcluster-deploy) repository, and place it in the `/data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts` directory.

        ```shell
        mkdir /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        cd /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        wget https://raw.gitcode.com/Ascend/mindcluster-deploy/raw/master/samples/train/basic-training/without-ranktable/pytorch/train_start.sh
        ```

    3. Execute the following command to obtain the `pytorch_standalone_acjob_quickstart.yaml` file from the `samples/train/basic-training/without-ranktable/pytorch` directory of the [MindCluster-Samples](https://gitcode.com/Ascend/mindcluster-deploy) repository. The example defaults to a single-server single-device job.

        ```shell
        cd /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts
        wget https://raw.gitcode.com/Ascend/mindcluster-deploy/raw/master/samples/train/basic-training/without-ranktable/pytorch/pytorch_standalone_acjob_quickstart.yaml
        ```

    4. (Optional) Prepare the dataset. The `--dummy` parameter is set by default in `pytorch_standalone_acjob_quickstart.yaml`, which automatically generates a random dataset for the training job, so the training job can be started without a real dataset. If you need to use a real dataset, delete the `--dummy` parameter from this yaml file, prepare the dataset corresponding to ResNet-50 by yourself, and upload the dataset to `/data/atlas_dls/public/dataset/resnet50/imagenet` in compliance with the corresponding specifications.

        ```shell
        mkdir /data/atlas_dls/public/dataset/resnet50/imagenet
        cd /data/atlas_dls/public/dataset/resnet50/imagenet
        ```

3. Submit a single-server single-device job.

    ```shell
    kubectl apply -f /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/pytorch_standalone_acjob_quickstart.yaml
    ```

4. View the Pod running status.

    ```shell
    kubectl get pod -A -o wide
    ```

   The example output is as follows. The presence of `Running` indicates that the job is running normally.

   >[!NOTE]
   >
   > In the output, `192.168.244.xxx` is the actual IP address assigned to the Pod, and `worker01` is the actual node name. Refer to the actual output.

    ```output
    NAMESPACE        NAME                                       READY   STATUS    RESTARTS   AGE     IP                NODE      NOMINATED NODE   READINESS GATES
    default          default-test-pytorch-master-0              1/1     Running   0          6s      192.168.244.xxx   worker01   <none>           <none>
    ```

   >[!NOTE]
   >
   >If the job remains in the `Pending` status after submission, see [Training job stuck in Pending status, cause: nodes are unavailable](https://gitcode.com/Ascend/mind-cluster/issues/352) or [Job stuck in Pending status due to insufficient resources](https://gitcode.com/Ascend/mind-cluster/issues/355) for handling.

5. View the training results.

    1. Execute the following command on any node to view the training result.

        ```shell
        kubectl logs -n default default-test-pytorch-master-0
        ```

    2. View the training log. If the following content appears, the training is successful.

       >[!NOTE]
       >
       > In the output, `10.106.227.xxx` is the actual IP address assigned by the cluster. Refer to the actual output.

        ```output
        [20260724-11:16:23] [MindXDL Service Log]Training start at 2026-07-24-11:16:23
        /usr/local/python3.9.2/lib/python3.9/site-packages/torchvision/io/image.py:13: UserWarning: Failed to load image Python extension: 'libc10_cuda.so: cannot open shared object file: No such file or directory'If you don't plan on using image functionality from `torchvision.io`, you can ignore this warning. Otherwise, there might be something wrong with your environment. Did you have `libjpeg` or `libpng` installed before building `torchvision` from source?
          warn(
        /job/code/main.py:215: UserWarning: You have chosen to seed training. This will turn on the CUDNN deterministic setting, which can slow down your training considerably! You may see unexpected behavior when restarting from checkpoints.
          warnings.warn('You have chosen to seed training. '
        /job/code/main.py:222: UserWarning: You have chosen a specific GPU. This will completely disable data parallelism.
          warnings.warn('You have chosen a specific GPU. This will completely '
        Use GPU: 0 for training
        => creating model 'resnet50'
        ```

6. Clean up the test resources.

    ```shell
    # Delete the training job Pod
    kubectl delete -f /data/atlas_dls/public/code/ResNet50_ID4149_for_PyTorch/scripts/pytorch_standalone_acjob_quickstart.yaml

    # Uninstall the components deployed by Helm (optional).
    helm uninstall mindcluster
    helm uninstall mindcluster-crds
    ```
