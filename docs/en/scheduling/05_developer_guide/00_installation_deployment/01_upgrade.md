# Manual Upgrade<a name="ZH-CN_TOPIC_0000002479226452"></a>

## Upgrade Instructions<a name="ZH-CN_TOPIC_0000002511346381"></a>

This chapter is intended to guide you through upgrading the MindCluster cluster scheduling components to a new version. The upgrade supports the following two methods.

- Full upgrade: This type of upgrade not only upgrades the binary image files of each component, but also allows modification of the component's configuration files after the upgrade. This type of upgrade supports cross-version upgrades; for example, you can upgrade from version 5.0.x to version 7.0.x.
- Image upgrade: This type of upgrade only upgrades the binary files of each component. It does not support modifying permissions, startup parameters, etc., and does not require a pre-upgrade environment check. This type of upgrade only supports upgrades within the same version.

    **Table 1**  Upgrade method description

    <a name="table1527494117524"></a>

    |Upgrade Method|Cross-Version Upgrade Supported|Training/Inference Jobs Need to Be Stopped| Reference|
    | --- | --- | --- | --- |
    |Full upgrade|Yes|Yes|[Upgrade Instructions](#upgrade-instructions)\-[Upgrading Other Components](#upgrading-other-components)|
    |Image upgrade|No|No|[Upgrading Image](#upgrading-container-manager)|

    >[!NOTE]
    >This chapter does not apply to the following scenario: The user has modified the source code (excluding configuration files) of the old MindCluster cluster scheduling components. In this case, analyze the version code differences before upgrading.

**Upgrade Environment Check<a name="section19242859587a"></a>**

Before performing the upgrade steps for each component, select the corresponding component for checking based on the actual installation scenario.

1. Check whether there are any running jobs. If a job is currently being executed, wait for the job to complete or stop the job in advance before upgrading MindCluster components.
    1. Execute the following command to check whether there are any running jobs.

        ```shell
        kubectl get pods -A
        ```

        The output example is as follows.

        ```ColdFusion
        NAMESPACE        NAME                                       READY   STATUS    RESTARTS         AGE
        default          ubuntu-pod                                 1/1     Running   32 (118m ago)    3d18h ...
        ```

    2. Enter the path where the job YAML is located and execute the following command to stop the job.

        ```shell
        kubectl delete -f  xxx.yaml              # xxx represents the name of the job YAML. Fill it in according to the actual situation.
        ```

2. (Optional) Check whether the pingmesh UnifiedBus network detection switch is turned off.
    1. Execute the following command to edit the `pingmesh-config` ConfigMap.

        ```shell
        kubectl edit cm -n cluster-system pingmesh-config
        ```

       If the output is as follows, it indicates that the pingmesh UnifiedBus network detection switch is off. There is no need to perform [Step 2](#li1427143773119).

        ```ColdFusion
        Error from server (NotFound): configmaps "pingmesh-config" not found
        ```

    2. <a name="li1427143773119"></a>(Optional) Modify the value of the `activate` field.
        - If the SuperPoD ID is in the `pingmesh-config` file, modify the `activate` field under this SuperPoD ID to `off`.
        - If the SuperPoD ID is not in the `pingmesh-config` file, you can configure it using the following two methods.
            - Add the SuperPoD information in the configuration file and set `activate` to `off`.
            - Delete all SuperPoD information in the `pingmesh-config` file, and set the value of the `activate` field in the global configuration to `off`.

3. Check the installed MindCluster components.
    - (Optional) Check TaskD. Execute the following command to enter the container and view the TaskD installation status.

        ```shell
        docker run -it  {Training_image_name}:tag /bin/bash
        pip show taskd
        ```

        The output example as follows indicates that TaskD has been installed in the image.

        ```ColdFusion
        Name: taskd
        Version: x.x.x
        Summary: Ascend MindCluster taskd is a new library for training management
        Home-page: UNKNOWN
        Author:
        Author-email:
        License: UNKNOWN
        Location: /usr/local/python3/lib/python3.10/site-packages
        Requires: grpcio, protobuf, pyOpenSSL, torch, torch-npu
        Required-by:
        ```

    - (Optional)Check other components. Refer to [Component Status Confirmation](../../03_installation_guide/03_confirming_status.md) to confirm whether the corresponding components are installed on the nodes in the cluster.

4. (Optional) If MindCluster cluster scheduling components have not been installed, refer to the [Installation and Deployment](00_manual_installation/00_obtaining_software_packages.md) chapter to install them first. For the installation steps of TaskD, refer to the [Preparing an Image](../../04_usage/04_resumable_training/04_using_resumable_training_on_the_cli.md#building-an-image) chapter.

## Upgrading Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226420"></a>

Only Ascend Docker Runtime supports upgrade via the command line. Other cluster scheduling components can be upgraded by uninstalling them and then reinstalling them.

Currently, only the `root` user is supported for upgrading Ascend Docker Runtime.

**Prerequisites<a name="section176591058124515"></a>**

The [upgrade environment check](#upgrade-instructions) has been completed.

**Upgrade Steps<a name="section520182224617"></a>**

1. Download the new version of the component installation package. For details, see the [Obtain Software Packages](00_manual_installation/00_obtaining_software_packages.md) chapter.
2. <a name="li12599722163212"></a>Enter the path where the installation package (run package) is located, and execute the following command in this path to add executable permissions to the software package.

    ```shell
    cd <path to run package>
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. Upgrade Ascend Docker Runtime using the following command.
    - (Optional) To upgrade Ascend Docker Runtime in the default path, execute the following command.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade
        ```

    - (Optional) To upgrade Ascend Docker Runtime in a specified path, execute the following command. The `--install-path` parameter specifies the upgrade path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --upgrade --install-path=<path>
        ```

        The output example is as follows, indicating a successful upgrade.

        ```ColdFusion
        Uncompressing ascend-docker-runtime  100%
        ...
        [INFO] ascend-docker-runtime upgrade success
        ```

4. (Optional) Execute the following command to restart the container so that the new Ascend Docker Runtime takes effect. If no changes are involved in the installation path or installation parameters, you can skip this step.
    - Docker scenario (or K8s integrated with Docker)

        ```shell
        systemctl daemon-reload && systemctl restart docker
        ```

    - Containerd scenario (or K8s integrated with Containerd)

        ```shell
        systemctl daemon-reload && systemctl restart containerd
        ```

5. <a name="li76002022113215"></a>Refer to the [Component Status](../../03_installation_guide/03_confirming_status.md) chapter to check whether the new version of Ascend Docker Runtime has been upgraded successfully.
6. (Optional) Restore the old version. Download the old version installation package, and re-execute [Step 2](#li12599722163212) to [Step 5](#li76002022113215) in sequence.

## Upgrading TaskD<a name="ZH-CN_TOPIC_0000002479226444"></a>

TaskD is installed inside the training image. Re-install the whl package inside the training image to complete the upgrade.

**Prerequisites<a name="section18616132394915"></a>**

The [upgrade environment check](#upgrade-instructions) has been completed.

**Upgrade Steps<a name="section1720814439492"></a>**

1. Refer to the [Obtain Software Packages](00_manual_installation/00_obtaining_software_packages.md) chapter to download the new version of the component installation package.
2. After the download is complete, enter the path where the installation package is located and decompress the installation package.
3. Execute the `ls -l` command. The output example is as follows.

    ```ColdFusion
    -rw-r--r-- 1 root root 1493228 Mar 14 02:09 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 1506842 Mar 12 18:07 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4. Based on the existing training image, install the new version of TaskD.
    1. Execute the following command to run the training image.

        ```shell
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest bash
        ```

    2. Execute the following command to uninstall the installed TaskD.

        ```shell
        pip uninstall taskd -y
        ```

        The output example as follows indicates a successful uninstallation.

        ```ColdFusion
        Successfully uninstalled taskd-{version}
        ```

    3. Execute the following command to install the new version of TaskD.

        ```shell
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        The output is as follows:

        ```ColdFusion
        Successfully installed taskd-{version}
        ```

    4. After installing the new version of TaskD, save the container as a new image.

        ```shell
        docker ps
        ```

        The output is as follows:

        ```ColdFusion
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES
        8b70390775f2   fd6acb527bad           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        Commit the container as a new version of the training container image. Note that the tag of the new image is inconsistent with the old image.

        ```shell
        docker commit 8b70390775f2 newimage:latest
        ```

5. Check whether the new version of TaskD has been upgraded. Refer to the [Checking TaskD](#upgrade-instructions) chapter to verify that the component status is normal.
6. (Optional) Roll back to the old version. If the old image still exists, no rollback operation is required; if it does not exist, follow the steps above to reinstall the old version of the TaskD package.

## Upgrading Container Manager<a name="ZH-CN_TOPIC_0000002524548731"></a>

Use the deployment script (`deploy.sh`) to perform the upgrade. The script automatically completes binary replacement and service restart, keeping the existing service configuration (startup parameters, etc.) unchanged.

>[!IMPORTANT]
>Please use the `upgrade` command for the upgrade. Do not use the `install` command. Repeated use of the install command will overwrite the existing service configuration (such as startup parameters), which may cause the service configuration to be lost.

1. Log in to the node where the Container Manager component is deployed as the `root` user.

2. Upload the obtained new version Container Manager package to any directory on the server. The following uses the `/home/container-manager` directory as an example (version 26.1.0 and later support deployment script upgrade).

    >[!NOTE]
    >If the server can access the network, you can also download the package using the following command:
    >
    >```shell
    >wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    >```
    >
    ><i>\<version\></i> is the version number of the package; <i>\<arch\></i> is the CPU architecture (such as x86_64, aarch64).

3. Enter the directory where the package is located and decompress the package.

    ```shell
    cd /home/container-manager
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

4. Enter the decompressed directory of the package and execute the `deploy.sh` script to upgrade Container Manager.

    ```shell
    bash deploy.sh upgrade
    ```

    The upgrade script will execute the following operations in sequence: stop the service, replace the binary files, and restart the service.

    After a successful upgrade, the output example is as follows:

    ```shell
    [INFO] Upgrading container-manager...
    [INFO] Current version : container-manager version: v26.0.0_linux-x86-64
    [INFO] Target version  : container-manager version: v26.1.0_linux-x86-64
    [INFO] Stopping service...
    [INFO] Replacing binary...
    [INFO] Starting service...
    [INFO] Binary upgraded to: container-manager version: v26.1.0_linux-x86-64
    [INFO] Upgrade completed successfully
    ```

5. Verify the upgrade status of Container Manager.
    1. Check the status of the component service. The component status must be `active (running)`.

        ```shell
        systemctl status container-manager.service
        ```

        Output example:

        ```ColdFusion
        ● container-manager.service - Ascend container manager
             Loaded: loaded (/etc/systemd/system/container-manager.service; disabled; vendor preset: enabled)
             Active: active (running) since Wed 2025-11-26 20:56:50 UTC; 16s ago
            Process: 41459 ExecStart=/bin/bash -c container-manager run  -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null 2>&1 & (code=exited, status=0/SUCCESS)
           Main PID: 41464 (container-manag)
              Tasks: 10 (limit: 629145)
             Memory: 13.3M
             CGroup: /system.slice/container-manager.service
                     └─41464 /home/container-manager/container-manager run -ctrStrategy ringRecover
        ...
        ```

    2. Check the component logs.

        ```shell
        cat /var/log/mindx-dl/container-manager/container-manager.log
        ```

        Output example for the Atlas 800I A3 SuperPoD server:

        ```ColdFusion
        [INFO]     2025/11/25 22:46:59.007163 1       hwlog/api.go:108    container-manager.log's logger init success
        [INFO]     2025/11/25 22:46:59.007288 1       command/run.go:150    init log success
        [INFO]     2025/11/25 22:46:59.007506 1       devmanager/devmanager.go:134    get card list from dcmi reset timeout is 60
        [INFO]     2025/11/25 22:46:59.250103 1       devmanager/devmanager.go:142    deviceManager get cardList is [0 1 2 3 4 5 6 7], cardList length equal to cardNum: 8
        [INFO]     2025/11/25 22:46:59.250267 1       devmanager/devmanager.go:171    the dcmi version is 25.5.0.b030
        [INFO]     2025/11/25 22:46:59.250405 1       devmanager/devmanager.go:235    chipName: Ascend910, devType: Ascend910A3
        ...
        ```

        If the following print information appears, it indicates that the component is running normally.

        ```ColdFusion
        ...
        [INFO]     2025/11/25 22:46:59.289352 1       devmgr/workflow.go:57    init module <hwDev manager> success
        [INFO]     2025/11/25 22:46:59.293773 1       app/config.go:40    load fault config from /home/faultCode.json success
        [INFO]     2025/11/25 22:46:59.293866 1       app/workflow.go:50    init module <fault manager> success
        [INFO]     2025/11/25 22:46:59.293901 1       app/workflow.go:76    init module <container controller> success
        [INFO]     2025/11/25 22:46:59.293930 1       app/workflow.go:64    init module <reset-manager> success
        [INFO]     2025/11/25 22:46:59.315101 378     devmgr/hwdevmgr.go:365    subscribe device fault event success
        ...
        ```

## Upgrading Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511346402"></a>

**Prerequisites**

The [upgrade environment check](#upgrade-instructions) has been completed.

**Precautions**

- Before upgrading Ascend Device Plugin, you must first upgrade Ascend Docker Runtime to version 26.1.0 or later.
- If Ascend Docker Runtime is not upgraded, or is not used, or a non-default mount path is used, you need to uncomment the following `volumeMounts` and `volumes` sections in the Ascend Device Plugin deployment YAML and modify them as needed:

  `volumeMounts`:

  ```yaml
          - name: docker-sock  # if container runtime is containerd, delete this volumeMount, else do not modify
            mountPath: /run/docker.sock
            readOnly: true
          - name: docker-dir  # if container runtime is containerd, delete this volumeMount, else do not modify
            mountPath: /run/docker
            readOnly: true
          - name: containerd
            mountPath: /run/containerd
            readOnly: true
  ```

  `volumes`:

  ```yaml
        - name: docker-sock    # if container runtime is containerd, delete this volumeMount, else do not modify
          hostPath:            # if your sock file is not in /run/docker.sock, please modify it, not support symbolic link
            path: /run/docker.sock
        - name: docker-dir     # if container runtime is containerd, delete this volumeMount, else do not modify
          hostPath:            # if your docker dir is not in /run/docker, please modify it, not support symbolic link
            path: /run/docker
        - name: containerd
          hostPath:
            path: /run/containerd
  ```

**Upgrade Steps**

1. Log in to the management node as the `root` user.
2. Navigate to the directory where the Ascend Device Plugin YAML configuration file is located (e.g., `/home/ascend-device-plugin`).

    ```shell
    cd /home/ascend-device-plugin
    ```

3. In the Ascend Device Plugin component installation environment, run the following command to uninstall the old version of Ascend Device Plugin.

    ```shell
    kubectl delete -f device-plugin-volcano-v{version}.yaml
    ```

4. Delete the DaemonSet of the old version of Ascend Device Plugin.

    First, query the DaemonSet name of Ascend Device Plugin:

    ```shell
    kubectl get ds -A | grep ascend-device-plugin
    ```

    - If the query result shows that no DaemonSet for Ascend Device Plugin exists, skip this step.
    - If a DaemonSet for Ascend Device Plugin still exists, delete the corresponding DaemonSet based on the query result. For example:

      ```shell
      kubectl delete ds -n kube-system ascend-device-plugin-daemonset-910
      ```

5. Install the new version of Ascend Device Plugin.

    For detailed steps, refer to [Manually Installing Ascend Device Plugin](00_manual_installation/04_ascend_device_plugin.md).

## Upgrading Other Components<a name="ZH-CN_TOPIC_0000002511346401"></a>

**Prerequisites<a name="section176591058124515"></a>**

- The [upgrade environment check](#upgrade-instructions) has been completed.

- If you need to upgrade NPU Exporter, Volcano, K8s RDMA Shared Dev Plugin, ClusterD, Ascend Operator, Infer Operator, and NodeD, you must uninstall the old versions first, and then execute the installation steps for the new version.

**NOTE**

If a job with resumable training enabled is running, upgrading ClusterD will cause resumable training to become invalid. In this case, after completing the new version image creation, install the new version of ClusterD directly without uninstalling the old version.

**Upgrade Steps<a name="section65996266718"></a>**

1. Uninstall the old version of MindCluster components. For details, see [Uninstalling Other Components](02_uninstallation.md#uninstalling-other-components).
2. Refer to the [Obtain Software Packages](00_manual_installation/00_obtaining_software_packages.md) chapter to download the new of the component installation packages.
3. (Optional) Prepare the new image of the MindCluster cluster scheduling components. If the new component is installed using binary files, you can skip this step.

    Refer to the [Preparing an Image](00_manual_installation/01_preparing_for_installation.md#preparing-an-image) chapter to pull the new version image from the Ascend image repository or create a new version image. Note that the tag of the new version component image must be inconsistent with that of the old component image to avoid overwriting the old component image.

4. <a name="li147194506333"></a>Re-execute the manual installation steps based on the component to be upgraded. For detailed steps, see [Installing New MindCluster Components](00_manual_installation/00_obtaining_software_packages.md).
5. (Optional) If you need to roll back to an older version, execute "[Uninstall Other Components](02_uninstallation.md#uninstalling-other-components) > Step 2" and [Step 4](#li147194506333) in sequence to uninstall the new components and then install the old components.
6. When upgrading the Infer Operator component from version 26.0.0 or earlier to the latest version, you need to delete the log directory and recreate it with `root` permissions, or change the permissions of the log directory and log files to `root`.

## Upgrading Image<a name="ZH-CN_TOPIC_0000002511346311"></a>

This chapter only guides users on upgrading the binary file version within a container image under the same version. Permissions and startup parameters will not be modified during the upgrade process. For more detailed instructions on upgrade methods, see [Upgrade Instructions](#upgrade-instructions).

- If you need to upgrade the images of Volcano, ClusterD, Ascend Operator, and Infer Operator, refer to [Upgrading Management Node Components](#section1292111716589).
- If you need to upgrade the images of NPU Exporter, Ascend Device Plugin, and NodeD, refer to [Upgrading Compute Node Components](#section231311416588).
- TaskD does not currently support this type of upgrade.

**Upgrading Management Node Components<a name="section1292111716589"></a>**

1. Refer to the [Preparing an Image](./00_manual_installation/01_preparing_for_installation.md#preparing-an-image) chapter and use the new software package to create the image.

    >[!NOTE]
    >Keep the image name consistent; otherwise, the original service configuration may fail to start the Pod.

2. Execute the following command to query the old Deployment configuration.

    ```shell
    kubectl get deployment -A|grep {component_name}
    ```

    Taking ClusterD as an example, the output example is as follows.

    ```ColdFusion
    mindx-dl         clusterd        1/1     1      1       45h
    ```

3. Execute the following command to restart Deployment.

    ```shell
    kubectl rollout restart deployment -n {namespace_name} {deployment_name}
    ```

    Taking ClusterD as an example, the output example is as follows.

    ```ColdFusion
    deployment.apps/clusterd restarted
    ```

4. Check whether the new Pod has been launched.

    ```shell
    kubectl get pod -A|grep {component_name}
    ```

    Taking ClusterD as an example, the output example is as follows, indicating that the Pod has been launched successfully.

    ```ColdFusion
    mindx-dl   clusterd-99f8795c8-drqb4  1/1  Running 0       1m
    ```

**Upgrading the Compute Node Components<a name="section231311416588"></a>**

1. Refer to the [Preparing an Image](00_manual_installation/01_preparing_for_installation.md#preparing-an-image) chapter and use the new software package to create the image.

    >[!NOTE]
    >Keep the image name consistent; otherwise, the original configuration file may fail to start the Pod.

2. Execute the following command to query the old DaemonSet configuration.

    ```shell
    kubectl get ds -A|grep {component_name}
    ```

    Taking NodeD as an example, the output example is as follows.

    ```ColdFusion
    mindx-dl         noded        1/1     1      1       45h
    ```

3. Execute the following command to restart the DaemonSet.

    ```shell
    kubectl rollout restart ds -n {namespace_name} {DaemonSet_name}
    ```

    Taking NodeD as an example, the output example is as follows.

    ```ColdFusion
    daemonsets.apps/noded restarted
    ```

4. Check whether the new version Pod has been started.

    ```shell
    kubectl get pod -A|grep {component_name}
    ```

    Taking NodeD as an example, the output example is as follows, indicating that the Pod has been started.

    ```ColdFusion
    mindx-dl   noded- m4j4r  1/1  Running 0     1m
    ```

## Elastic Agent (Deprecated) Upgrade to TaskD<a name="ZH-CN_TOPIC_0000002515202401"></a>

>[!NOTE]
>Elastic Agent has been deprecated as of version 7.3.0. This chapter is retained solely as a historical migration reference. If Elastic Agent is no longer used in your current environment, you may skip this chapter.

Elastic Agent has been deprecated. This chapter provides operational guidance for upgrading Elastic Agent to TaskD.

**Prerequisites<a name="section565512391204"></a>**

- The upgrade environment check has been completed.
- Elastic Agent is installed in the training image.

**Steps<a name="section1643711813"></a>**

1. Refer to the [Obtain Software Packages](00_manual_installation/00_obtaining_software_packages.md) chapter to download the new version of the TaskD installation package.
2. After the download is completed, enter the path where the installation package is located and decompress the installation package.
3. Execute the `ls -l` command. The output example is as follows.

    ```ColdFusion
    -rw-r--r-- 1 root root 6134726 Nov 10 10:32 Ascend-mindxdl-taskd_{version}_linux-aarch64.zip
    -r-------- 1 root root 6205642 Nov  5 23:38 taskd-{version}-py3-none-linux_aarch64.whl
    ```

4. Based on the existing training image, uninstall Elastic Agent and install the new version of TaskD.
    1. Run the training image.

        ```shell
        docker run -it  -v /host/packagepath:/container/packagepath training_image:latest /bin/bash
        ```

    2. Uninstall the installed Elastic Agent.

        ```shell
        pip uninstall mindx-elastic -y
        ```

        The output example as follows indicates a successful uninstallation.

        ```ColdFusion
        Successfully uninstalled mindx_elastic-{version}
        ```

    3. Delete the Elastic Agent startup code.

        ```shell
        sed -i '/mindx_elastic.api/d' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

        (Optional) Execute the following command to check whether the Elastic Agent embedded code has been deleted from the corresponding file.

        ```shell
        vi $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    4. Install the new version of TaskD.

        ```shell
        pip install taskd-{version}-py3-none-linux_aarch64.whl
        ```

        The output example is as follows, indicating a successful installation.

        ```ColdFusion
        Successfully installed taskd-{version}
        ```

        Execute the following command to start TaskD.

        ```shell
        sed -i '/import os/i import taskd.python.adaptor.patch' $(pip3 show torch | grep Location | awk -F ' ' '{print $2}')/torch/distributed/run.py
        ```

    5. After installing the new version of TaskD, save the container as a new image.

        ```shell
        docker ps
        ```

        The output example is as follows:

        ```ColdFusion
        CONTAINER ID   IMAGE                  COMMAND                  CREATED        STATUS        PORTS     NAMES
        bb118ca00041    f76142d63d3a           "/bin/bash -c 'sleep…"   2 hours ago    Up 2 hours              k8s_ascend_default-last-test-deepseek2-60b
        ```

        Commit this container as the new version training container image. Note that the tag of the new image is inconsistent with the old image.

        ```shell
        docker commit bb118ca00041 newimage:latest
        ```

5. Check whether TaskD replacement is completed. Refer to the [Checking TaskD](#upgrade-instructions) chapter to check whether the component status is normal.
6. Modify the training script (for example, `train_start.sh`) and the job YAML.
    1. Create the `manager.py` file and place it in the current directory when calling the training script. The content of the `manager.py` file is as follows.

        ```Python
        from taskd.api import init_taskd_manager, start_taskd_manager
        import os

        job_id=os.getenv("MINDX_TASK_ID")
        node_nums=XX         # Total number of nodes
        proc_per_node=XX     # Training processes per node

        init_taskd_manager({"job_id":job_id, "node_nums": node_nums, "proc_per_node": proc_per_node})
        start_taskd_manager()
        ```

        >[!NOTE]
        >For detailed parameter descriptions in the manager.py file, see [def init\_taskd\_manager\(config:dict\) -\> bool:](../../06_api/07_taskd/04_taskd_manager_apis.md#def-init_taskd_managerconfigdict---bool).

    2. Add the following code to the training script to start TaskD Manager.

        <pre codetype="Python">
        <strong>export TASKD_PROCESS_ENABLE="on"
        # Take PyTorch framework as an example
        if [[ "${RANK}" == 0 ]]; then
            export MASTER_ADDR=${POD_IP}
            python /job/code/manager.py 2>> /job/code/alllogs/$MINDX_TASK_ID/taskd/error.log &           # The specific execution path of manager.py is determined by the current path, and the error.log path must be created in advance.
        fi</strong>

        torchrun ...</pre>

    3. Modify the container port in the job YAML, and add `port 9601` for TaskD communication under all Pods.

        <pre codetype="yaml">
        ...
                spec:
        ...
                   containers:
        ...
                     <strong>ports:
                       - containerPort: 9601
                         name: taskd-port</strong>
        ...</pre>
