# Manual uninstallation<a name="ZH-CN_TOPIC_0000002511426389"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-30T12:20:55.738Z pushedAt=2026-06-30T12:23:24.363Z -->

- To uninstall Ascend Docker Runtime, see [Uninstalling Ascend Docker Runtime](#section6134163311244).
- To uninstall Container Manager, see [Uninstalling the Container Manager component](#section1461059103619).
- To uninstall NPU Exporter, Ascend Device Plugin, Volcano, ClusterD, Ascend Operator, Infer Operator, NodeD, and Resilience Controller, see [Uninstalling other components](#section6361146202520).

## Uninstalling Ascend Docker Runtime<a name="section6134163311244"></a>

- Scenario 1: Different installation paths are used.

    When uninstalling Ascend Docker Runtime, perform the uninstallation operation twice for different container engines according to [Step 2](#li345320287225). Each uninstallation requires specifying the corresponding installation path, i.e., the `--install-path` parameter.

- Scenario 2: The same installation path is used.

    When uninstalling Ascend Docker Runtime, perform the uninstallation operation once according to [Step 2](#li345320287225). After the uninstallation is complete, you need to manually restore the `daemon.json` file of the other engine to its content before Ascend Docker Runtime was installed.

If you need to retain one of the container engines, reinstall it for the corresponding scenario after Ascend Docker Runtime is uninstalled.

1. (Optional) Disable pingmesh network detection.
    1. Log in to the environment and go to the NodeD extraction directory.
    2. Run the following command to edit the `pingmesh-config` file.

        ```shell
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3. Modify the value of the `activate` field.
        - If the SuperPoD node ID is in the `pingmesh-config` file, set `activate` to `off` under the corresponding SuperPoD ID field.
        - If the SuperPoD ID is not in the `pingmesh-config` file, you can set it in the following two ways.
            - Add the SuperPoD information in the configuration file and set `activate` to `off`.
            - Delete all SuperPoD information in the `pingmesh-config` file and set the value of the `activate` field in the global configuration to `off`.

2. <a name="li345320287225"></a>You can choose one of the following methods to uninstall Ascend Docker Runtime.
    - Method 1: (Recommended) Use the software package to uninstall
        1. First, go to the path where the installation package (run package) is located.

            ```shell
            cd <path to run package>
            ```

        2. Run the following uninstall command to uninstall Ascend Docker Runtime in the default installation path.

            - Docker scenario (or K8s integrated with Docker scenario)

                ```shell
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall
                ```

            - Containerd scenario (or K8s integrated with Containerd scenario)

                ```shell
                ./Ascend-docker-runtime_{version}_linux-{arch}.run --uninstall --install-scene=containerd
                ```

            >[!NOTE]
            >- If the Docker configuration file path is not the default `/etc/docker/daemon.json`, add the `--config-file-path` parameter to specify the configuration file path.
            >- If the Containerd configuration file path is not the default `/etc/containerd/config.toml`, add the `--config-file-path` parameter to specify the configuration file path.
            >- To uninstall Ascend Docker Runtime in a specified installation path, add the `--install-path=<path>` parameter to the uninstall command.

            The echo example is as follows, indicating a successful uninstallation.

            ```ColdFusion
            Uncompressing ascend-docker-runtime  100%
            ...
            [INFO] ascend-docker-runtime uninstall success
            ```

    - Method 2: Use the script to uninstall

        1. First, go to the `script` directory under the installation path of Ascend Docker Runtime (the default installation path is: `/usr/local/Ascend/Ascend-Docker-Runtime`):

            ```shell
            cd /usr/local/Ascend/Ascend-Docker-Runtime/script
            ```

        2. Run the uninstallation script to uninstall.

            - Docker scenario (or K8s integration with Docker scenario)

                ```shell
                uninstall.sh docker docker <daemon.json file path>
                ```

            - Containerd scenario (or K8s integrated with Containerd scenario)

                ```shell
                uninstall.sh containerd containerd <config.toml file path>
                ```

            >[!NOTE]
            >- You can omit the path to the Docker configuration file `daemon.json`. If not specified, `/etc/docker/daemon.json` is used by default.
            >- You can omit the path to the Containerd configuration file `config.toml`. If not specified, `/etc/containerd/config.toml` is used by default.

        The command output example is as follows, indicating a successful uninstallation.

        ```ColdFusion
        [INFO]: You will recover Docker's daemon
        ...
        [INFO] uninstall.sh exec success
        ```

3. (Optional) In the K8s integrated with Containerd scenario, if you need to restore the modified `kubeadm-flags.env`, see the [K8s official documentation](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/) to restore the `kubeadm-flags.env` configuration file. This step can be skipped in other scenarios.
4. Restart the service.
    - Docker scenario (or K8s integrated with Docker scenario)

        ```shell
        systemctl daemon-reload && systemctl restart docker
        ```

    - Containerd scenario (or K8s integrated with Containerd scenario)

        ```shell
        systemctl daemon-reload && systemctl restart containerd
        ```

## Uninstalling Container Manager<a name="section1461059103619"></a>

Use the deployment script (`deploy.sh`) for uninstallation. The script automatically completes operations such as stopping and disabling the service, deleting the systemd unit file, and removing binary files.

1. Log in to the node where Container Manager is deployed as the `root` user.

2. Run the following commands in sequence to uninstall the Container Manager service.

    ```shell
    # Stop the Container Manager service.
    systemctl stop container-manager.timer
    systemctl disable container-manager.timer
    systemctl stop container-manager.service
    systemctl disable container-manager.service
    
    # Delete the Container Manager service.
    rm -f /etc/systemd/system/container-manager.service
    rm -f /etc/systemd/system/container-manager.timer
    systemctl daemon-reload
    systemctl reset-failed
    
    # Delete the corresponding Container Manager binary file.
    chattr -i /usr/local/bin/container-manager
    rm -f /usr/local/bin/container-manager
    ```

3. Delete the log files. Confirm the actual path before deletion.

    ```shell
    rm -rf /var/log/mindx-dl/container-manager
    ```

## Uninstalling Other Components<a name="section6361146202520"></a>

Uninstalling cluster scheduling components is supported. You can uninstall components and then reinstall the latest version. By uninstalling each component one by one and deleting the corresponding namespace, log directory, configuration file, etc., please select the corresponding uninstallation method based on the installation method.

1. (Optional) Disable pingmesh network detection.
    1. Log in to the environment and enter the NodeD decompression directory.
    2. Run the following command to edit the `pingmesh-config` file.

        ```shell
        kubectl edit cm -n cluster-system   pingmesh-config
        ```

    3. Modify the value of the `activate` field.
        - If the SuperPoD ID is in the `pingmesh-config` file, modify the activate field under this SuperPoD ID to off.
        - If the SuperPoD ID is not in the `pingmesh-config` file, you can configure it using the following two methods.
            - Add the SuperPoD information in the configuration file and set `activate` to `off`.
            - Delete the information of all SuperPoDs in the `pingmesh-config` file and set the value of the `activate` field in the global configuration to `off`.

2. Uninstall the component. Choose the corresponding uninstallation method based on how the component was installed.
    - Uninstall via image method. The uninstallation method is similar for each component: navigate to the directory containing the component's YAML configuration file and perform a delete operation. This operation must be performed on the K8s management node. The following uses uninstalling Ascend Device Plugin as an example; you should complete the uninstallation of the remaining components on your own.

        1. Log in to the management node as the root user.
        2. Navigate to the directory containing the Ascend Device Plugin YAML configuration file (e.g., `/home/ascend-device-plugin`).

            ```shell
            cd /home/ascend-device-plugin
            ```

        3. In the Ascend Device Plugin component installation environment, run the following command to uninstall Ascend Device Plugin.

            ```shell
            kubectl delete -f device-plugin-volcano-v{version}.yaml
            ```

            Command output:

            ```ColdFusion
            serviceaccount "ascend-device-plugin-sa-910" deleted
            clusterrole.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-role-910" deleted
            clusterrolebinding.rbac.authorization.k8s.io "pods-node-ascend-device-plugin-rolebinding-910" deleted
            deployment.apps "ascend-device-plugin-daemonset-910" deleted
            ```

        >[!NOTE]
        >When Ascend Device Plugin is used with Volcano, a ConfigMap is created. Run the following command to delete it.
        >
        >```shell
        >kubectl delete cm mindx-dl-deviceinfo-<node-name> -n kube-system
        >```

    - Uninstall via the binary method. The following uses uninstalling NPU Exporter as an example. You need to uninstall the remaining components yourself.
        1. Log in to the node where the component is deployed as the root user.
        2. In the NPU Exporter installation environment, run the following commands in sequence to uninstall NPU Exporter.

            ```shell
            systemctl stop npu-exporter.service
            systemctl disable npu-exporter.service
            chattr -i /etc/systemd/system/npu-exporter.service
            rm -f /etc/systemd/system/npu-exporter.service
            systemctl daemon-reload
            systemctl reset-failed
            chattr -i /usr/local/bin/npu-exporter
            rm -f /usr/local/bin/npu-exporter
            ```

3. Delete the namespaces. The namespace `npu-exporter` for NPU Exporter and the namespace `volcano-system` for Volcano are already deleted synchronously when the components are uninstalled, so you can skip this step.

    Run the following command to delete the namespace created when installing the cluster scheduling components. Deleting a namespace will delete all resources under that namespace. Confirm before executing.

    ```shell
    kubectl delete ns mindx-dl
    ```

    The echo example is as follows:

    ```ColdFusion
    namespace "mindx-dl" deleted
    ```

4. Delete log files. Refer to the [Creating Log Directories](./03_installation/manual_installation/01_preparing_for_installation.md#creating-log-directories) section and delete the log directories of the cluster scheduling components on the corresponding nodes. Take ClusterD as an example. Confirm before deleting.

    ```shell
    rm -rf /var/log/mindx-dl/clusterd
    ```

5. (Optional) When uninstalling Resilience Controller, if certificates and KubeConfig files are imported, you need to delete them. Confirm before deleting.

    ```shell
    rm -rf /etc/mindx-dl/resilience-controller
    ```
