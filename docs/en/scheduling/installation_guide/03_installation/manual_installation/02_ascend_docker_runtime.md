# Ascend Docker Runtime<a name="ZH-CN_TOPIC_0000002479226434"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:12:58.171Z pushedAt=2026-06-09T06:22:06.801Z -->

- Users who demand containerization support, full-NPU scheduling, static vNPU scheduling, dynamic vNPU scheduling, resumable training, elastic training, inference card fault recovery, or rescheduling upon inference card faults must install Ascend Docker Runtime.
- Users who only use resource monitoring do not need to install Ascend Docker Runtime and can skip this chapter directly.

## Prerequisites<a name="section137058405153"></a>

- Before installation, ensure that the user ID of the `runc` file is 0.
- Before installation, in the Containerd scenario, ensure that the user and user group of the `/etc/containerd/config.toml` file are `root`; in the Docker scenario or Isula scenario, ensure that the user and user group of the `/etc/docker/daemon.json` file are `root`.
- Before installation, ensure that the `RPATH` and `RUNPATH` of the so files configured in `/etc/ld.so.preload` do not contain the HDK driver directory with a relative path (you can check this using the `readelf -d xxx.so` command).

## Confirming the Installation Scenario<a name="zh-cn_topic_0000001930317932_section1235447163310"></a>

Currently, only the `root` user is supported for installing Ascend Docker Runtime. Please select the corresponding installation method based on the actual situation.

1. Run the following command on the K8s management node to query the node name.

    ```shell
    kubectl get node
    ```

    Example output:

    ```ColdFusion
    NAME       STATUS   ROLES           AGE   VERSION
    ubuntu     Ready    worker          23h   v1.17.3
    ```

2. Check the container runtime of the current node, where `node-name` is the node name.
    - Non-K8s scenario: Run the following command on any node.

        ```shell
        docker --version      # Docker
        containerd --version     # Containerd
        ```

        - If the output shows Docker version information, the current scenario is [Docker scenario](#zh-cn_topic_0000001930317932_section1443063532919).
        - If the output shows Containerd version information, the current scenario is [Containerd scenario](#zh-cn_topic_0000001930317932_section196591123133116).
        - If the output contains version information for both Docker and Containerd, determine which container runtime to use for job execution.

    - K8s integrated container runtime scenario: Run the following command on the management node.

        ```shell
        kubectl describe node <node-name> | grep -i runtime
        ```

        - If the output contains Docker information, this indicates a [K8s integrated Docker scenario](#zh-cn_topic_0000001930317932_section1443063532919).
        - If the output contains Containerd information, this indicates a [K8s integrated Containerd scenario](#zh-cn_topic_0000001930317932_section14600174633116).

## Installing Ascend Docker Runtime in a Docker Scenario<a name="zh-cn_topic_0000001930317932_section1443063532919"></a>

The installation procedure of Ascend Docker Runtime in the K8s-Docker integrated scenario is the same as that in the Docker scenario.

1. After downloading the software package, go to the path where the software package (run package) is located on all compute nodes.

    ```shell
    cd <path to run package>
    ```

2. Run the following command to add executable permission to the software package.

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. Run the following command to verify the consistency and integrity of the software package installation file.

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    Example output:

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. Install Ascend Docker Runtime using the following command.

    - To install to the default path, run the following command.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install
        ```

    - To specify installation to a custom path, run the following command. The `--install-path` parameter specifies the installation path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-path=<path>
        ```

    >[!NOTE]
    >- An absolute path must be used for specifying the installation path.
    >- If the Docker configuration file path is not the default `/etc/docker/daemon.json`, add the `--config-file-path` parameter to specify the configuration file path.

    The following example output indicates a successful installation.

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. Run the following command to make Ascend Docker Runtime take effect.

    ```shell
    systemctl daemon-reload && systemctl restart docker
    ```

    Ascend Device Plugin automatically checks whether Ascend Docker Runtime exists during startup. Therefore, you need to start Ascend Docker Runtime first and then start Ascend Device Plugin. If you start Ascend Device Plugin first and then start Ascend Docker Runtime, you need to restart Ascend Device Plugin by referring to [Ascend Device Plugin](./04_ascend_device_plugin.md).

## Installing Ascend Docker Runtime in the Containerd Scenario<a name="zh-cn_topic_0000001930317932_section196591123133116"></a>

1. After the software package is downloaded, go to the path where the software package (run package) is located.

    ```shell
    cd <path to run package>
    ```

2. Run the following command to add executable permission to the software package.

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. Run the following command to verify the consistency and integrity of the software package installation file.

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    Example output:

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. Install Ascend Docker Runtime using the following command.

    - Install to the default path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    - To install to a specified path, run the following command. The `--install-path` parameter specifies the installation path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE]
        >- An absolute path must be used for specifying the installation path.
        >- If the Containerd configuration file path is not the default `/etc/containerd/config.toml`, you need to add the `--config-file-path` parameter to specify the configuration file path.

    The following example output indicates a successful installation.

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. <a name="zh-cn_topic_0000001930317932_section19659112313311605"></a>(Optional) If the installation fails, refer to the following steps to modify the Containerd configuration file.
    1. Open the configuration file.
        - **Scenario where Containerd has no default configuration file**: Run the following commands in sequence to create and modify the configuration file.

            ```shell
            mkdir /etc/containerd
            containerd config default > /etc/containerd/config.toml
            vim /etc/containerd/config.toml
            ```

        - **Scenario with an existing Containerd configuration file**: Open and modify the configuration file.

            ```shell
            vim /etc/containerd/config.toml
            ```

    2. Add `ascend runtime` and set it as the default runtime. An example is shown below.
       1. Locate the following runc configuration in the configuration file (note that `io.containerd.cri.v1.runtime` may vary across different Containerd versions; use the actual value):

            <pre>
             [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc]
               runtime_type = 'io.containerd.runc.v2'
               runtime_path = ''
               pod_annotations = []
               container_annotations = []
               privileged_without_host_devices = false
               privileged_without_host_devices_all_devices_allowed = false
               cgroup_writable = false
               base_runtime_spec = ''
               cni_conf_dir = ''
               cni_max_conf_num = 0
               snapshotter = ''
               sandboxer = 'podsandbox'
               io_type = ''

               [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.runc.options]
                 BinaryName = ''
                 CriuImagePath = ''
                 CriuWorkPath = ''
                 IoGid = 0
                 IoUid = 0
                 NoNewKeyring = false
                 Root = ''
                 ShimCgroup = ''
                 SystemdCgroup = true
            ...</pre>

       2. Copy the above configuration content, change `runc` to `ascend` and set the value of `BinaryName` to the installation path of the ascend-docker-runtime executable file. Refer to the following (where `io.containerd.cri.v1.runtime` may vary across different Containerd versions; use the actual value):

            <pre>
             [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.ascend]
               runtime_type = 'io.containerd.runc.v2'
               runtime_path = ''
               pod_annotations = []
               container_annotations = []
               privileged_without_host_devices = false
               privileged_without_host_devices_all_devices_allowed = false
               cgroup_writable = false
               base_runtime_spec = ''
               cni_conf_dir = ''
               cni_max_conf_num = 0
               snapshotter = ''
               sandboxer = 'podsandbox'
               io_type = ''

               [plugins.'io.containerd.cri.v1.runtime'.containerd.runtimes.ascend.options]
                 BinaryName = '/usr/local/Ascend/Ascend-Docker-Runtime/ascend-docker-runtime'
                 CriuImagePath = ''
                 CriuWorkPath = ''
                 IoGid = 0
                 IoUid = 0
                 NoNewKeyring = false
                 Root = ''
                 ShimCgroup = ''
                 SystemdCgroup = true
            ...</pre>

       3. Locate the following configuration item and change the value of `default_runtime_name` to `ascend` (where `io.containerd.cri.v1.runtime` may vary across different Containerd versions; use the actual value):

            Before modification:
            <pre>
            [plugins.'io.containerd.cri.v1.runtime'.containerd]
              default_runtime_name = 'runc'
              ignore_blockio_not_enabled_errors = false
              ignore_rdt_not_enabled_errors = false
              ...</pre>

            After modification:
            <pre>
            [plugins.'io.containerd.cri.v1.runtime'.containerd]
              default_runtime_name = 'ascend'
              ignore_blockio_not_enabled_errors = false
              ignore_rdt_not_enabled_errors = false
            ...</pre>

6. Run the following command to restart Containerd.

    ```shell
    systemctl daemon-reload && systemctl restart containerd
    ```

## Installing Ascend Docker Runtime in the K8s-Containerd Integrated Scenario<a name="zh-cn_topic_0000001930317932_section14600174633116"></a>

1. After the software package is downloaded, first go to the path where the installation package (run package) is located.

    ```shell
    cd <path to run package>
    ```

2. Run the following command to add executable permission to the software package.

    ```shell
    chmod u+x Ascend-docker-runtime_{version}_linux-{arch}.run
    ```

3. Run the following command to verify the consistency and integrity of the software package installation file.

    ```shell
    ./Ascend-docker-runtime_{version}_linux-{arch}.run --check
    ```

    Example output:

    ```ColdFusion
    [WARNING]: --check is meaningless for ascend-docker-runtime and will be discarded in the future
    Verifying archive integrity...  100%   SHA256 checksums are OK.
    ...
     All good.
    ```

4. You can install Ascend Docker Runtime using the following commands.

    - Install to the default path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd
        ```

    - To install to a specified path, run the following command. The `--install-path` parameter specifies the installation path.

        ```shell
        ./Ascend-docker-runtime_{version}_linux-{arch}.run --install --install-scene=containerd --install-path=<path>
        ```

        >[!NOTE]
        >You must use an absolute path when specifying the installation path.

    The following example output indicates a successful installation.

    ```ColdFusion
    Uncompressing ascend-docker-runtime  100%
    [INFO]: installing ascend-docker-runtime
    ...
    [INFO] ascend-docker-runtime install success
    ```

5. (Optional) If the installation fails, refer to [Step 5](#zh-cn_topic_0000001930317932_section19659112313311605) in [Installing Ascend Docker Runtime in the Containerd Scenario](#zh-cn_topic_0000001930317932_section196591123133116).

6. If you need to change the container runtime on the node from Docker to Containerd, you must modify the kubelet configuration file `kubeadm-flags.env` on the node. For details, see the [K8s official documentation](https://kubernetes.io/docs/tasks/administer-cluster/migrating-from-dockershim/change-runtime-containerd/).

7. If a Docker service exists, run the following command to stop the corresponding service.

    ```shell
    systemctl stop docker
    ```

8. Run the command to restart Containerd and kubelet. The following is an example output.

    ```shell
    systemctl daemon-reload && systemctl restart containerd kubelet
    ```

## Parameter Description <a name="zh-cn_topic_0000001930317932_section425619177219"></a>

[Table 1](#zh-cn_topic_0000001930317932_table35676204212) lists parameters supported by the Ascend Docker Runtime installation commands.

**Table 1** Parameters

<a name="zh-cn_topic_0000001930317932_table35676204212"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000001930317932_row1856732017219"><th class="cellrowborder" valign="top" width="32.43%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0000001930317932_p155677203214"><a name="zh-cn_topic_0000001930317932_p155677203214"></a><a name="zh-cn_topic_0000001930317932_p155677203214"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="67.57%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0000001930317932_p1456712016216"><a name="zh-cn_topic_0000001930317932_p1456712016216"></a><a name="zh-cn_topic_0000001930317932_p1456712016216"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000001930317932_row2568112072119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p05681620192117"><a name="zh-cn_topic_0000001930317932_p05681620192117"></a><a name="zh-cn_topic_0000001930317932_p05681620192117"></a>--help | -h</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1356892011218"><a name="zh-cn_topic_0000001930317932_p1356892011218"></a><a name="zh-cn_topic_0000001930317932_p1356892011218"></a>Query help information.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681520112117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p3568122042118"><a name="zh-cn_topic_0000001930317932_p3568122042118"></a><a name="zh-cn_topic_0000001930317932_p3568122042118"></a>--info</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p15568720122112"><a name="zh-cn_topic_0000001930317932_p15568720122112"></a><a name="zh-cn_topic_0000001930317932_p15568720122112"></a>Queries software package build information. This option will be deprecated in a later version.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row756832062117"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p6568142052120"><a name="zh-cn_topic_0000001930317932_p6568142052120"></a><a name="zh-cn_topic_0000001930317932_p6568142052120"></a>--list</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p4568182018212"><a name="zh-cn_topic_0000001930317932_p4568182018212"></a><a name="zh-cn_topic_0000001930317932_p4568182018212"></a>Queries the software package list. This option will be deprecated in a later version.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row2568520172112"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p856882092110"><a name="zh-cn_topic_0000001930317932_p856882092110"></a><a name="zh-cn_topic_0000001930317932_p856882092110"></a>--check</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p185681720182113"><a name="zh-cn_topic_0000001930317932_p185681720182113"></a><a name="zh-cn_topic_0000001930317932_p185681720182113"></a>Checks the consistency and integrity of packages. This option will be deprecated in a later version.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row165681920202119"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15568122012120"><a name="zh-cn_topic_0000001930317932_p15568122012120"></a><a name="zh-cn_topic_0000001930317932_p15568122012120"></a>--quiet</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p256818204217"><a name="zh-cn_topic_0000001930317932_p256818204217"></a><a name="zh-cn_topic_0000001930317932_p256818204217"></a>Indicates silent installation, which skips interactive messages. It must be used together with install, uninstall, or upgrade. This option will be deprecated in a later version.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19568182011213"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p55691220202114"><a name="zh-cn_topic_0000001930317932_p55691220202114"></a><a name="zh-cn_topic_0000001930317932_p55691220202114"></a>--tar arg1 [arg2 ...]</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p25691320162114"><a name="zh-cn_topic_0000001930317932_p25691320162114"></a><a name="zh-cn_topic_0000001930317932_p25691320162114"></a>Runs the <strong>tar</strong> command on the software package. Use the parameters following tar as the command parameters. For example, the <strong>--tar xvf</strong> command indicates that the .run package will be decompressed to the current directory. This option will be deprecated in a later version.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row156942092116"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p75697203214"><a name="zh-cn_topic_0000001930317932_p75697203214"></a><a name="zh-cn_topic_0000001930317932_p75697203214"></a>--install</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p1357015208213"><a name="zh-cn_topic_0000001930317932_p1357015208213"></a><a name="zh-cn_topic_0000001930317932_p1357015208213"></a>Install the software package. You can specify the installation path using --install-path=&lt;path&gt;, or install directly to the default path without specifying an installation path.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row19570122010217"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p15570172014213"><a name="zh-cn_topic_0000001930317932_p15570172014213"></a><a name="zh-cn_topic_0000001930317932_p15570172014213"></a>--install-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="p369633161410"><a name="p369633161410"></a><a name="p369633161410"></a>Specify the installation path.</p>
<a name="zh-cn_topic_0000001930317932_ul29611936455"></a><a name="zh-cn_topic_0000001930317932_ul29611936455"></a><ul id="zh-cn_topic_0000001930317932_ul29611936455"><li>An absolute path must be used as the installation path.</li><li>If the global configuration file "ascend_docker_runtime_install.info" exists in the environment, the specified installation path must be consistent with the installation path saved in the global configuration file.</li><li>If you want to change the installation path, you must first uninstall the <span id="zh-cn_topic_0000001930317932_ph1528115352583"><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a><a name="zh-cn_topic_0000001930317932_ph1528115352583"></a>Ascend Docker Runtime</span> software package from the original path and ensure that the global configuration file "ascend_docker_runtime_install.info" has been deleted.</li><li>If the <span id="zh-cn_topic_0000001930317932_ph93781522588"><a name="zh-cn_topic_0000001930317932_ph93781522588"></a><a name="zh-cn_topic_0000001930317932_ph93781522588"></a>Ascend Docker Runtime</span> version prior to 5.0.RC1 was installed via the ToolBox installation package, this file does not exist and does not need to be deleted.</li><li>If no installation path is specified, it will be installed to the default path <span class="filepath" id="zh-cn_topic_0000001930317932_filepath7570102017212"><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a><a name="zh-cn_topic_0000001930317932_filepath7570102017212"></a>"/usr/local/Ascend"</span>.</li><li>If the installation directory is specified using this parameter, the running user must have read and write permissions on the specified installation path.</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1444404185013"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1144584125019"><a name="zh-cn_topic_0000001930317932_p1144584125019"></a><a name="zh-cn_topic_0000001930317932_p1144584125019"></a>--install-scene=&lt;scene&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p153510190174"><a name="zh-cn_topic_0000001930317932_p153510190174"></a><a name="zh-cn_topic_0000001930317932_p153510190174"></a><span id="zh-cn_topic_0000001930317932_ph1308455195116"><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a><a name="zh-cn_topic_0000001930317932_ph1308455195116"></a>Ascend Docker Runtime</span> installation scenario. <span id="zh-cn_topic_0000001930317932_ph1641213426170"><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a><a name="zh-cn_topic_0000001930317932_ph1641213426170"></a>The default value is</span> <span id="zh-cn_topic_0000001930317932_ph8821719135318"><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a><a name="zh-cn_topic_0000001930317932_ph8821719135318"></a>docker,</span> and the value descriptions are as follows.</p>
<a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><a name="zh-cn_topic_0000001930317932_ul8352122811918"></a><ul id="zh-cn_topic_0000001930317932_ul8352122811918"><li><span id="zh-cn_topic_0000001930317932_ph3371331161710"><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a><a name="zh-cn_topic_0000001930317932_ph3371331161710"></a>docker</span>: Indicates installation in a <span id="zh-cn_topic_0000001930317932_ph1159416519530"><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a><a name="zh-cn_topic_0000001930317932_ph1159416519530"></a>Docker</span> (or <span id="zh-cn_topic_0000001930317932_ph5391475179"><a name="zh-cn_topic_0000001930317932_ph5391475179"></a><a name="zh-cn_topic_0000001930317932_ph5391475179"></a>K8s integration with Docker</span>) scenario.</li><li><span id="zh-cn_topic_0000001930317932_ph7743733115213"><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a><a name="zh-cn_topic_0000001930317932_ph7743733115213"></a>c</span><span id="zh-cn_topic_0000001930317932_ph1274373385212"><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a><a name="zh-cn_topic_0000001930317932_ph1274373385212"></a>ontainerd: Indicates installation in a</span> Containerd (or K8s integration with Containerd) scenario.</li><li>isula: Indicates installation in an iSula container engine scenario.</li></ul><p>--install-scene cannot be used alone and must be used together with --install, --uninstall, or --upgrade.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row16570162013216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1457092012114"><a name="zh-cn_topic_0000001930317932_p1457092012114"></a><a name="zh-cn_topic_0000001930317932_p1457092012114"></a>--uninstall</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p35701320182115"><a name="zh-cn_topic_0000001930317932_p35701320182115"></a><a name="zh-cn_topic_0000001930317932_p35701320182115"></a>Uninstall the software. If an installation path was specified during installation, the installation path must also be specified during uninstallation using the parameter --install-path=&lt;path&gt;.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row757019209212"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p11570122092117"><a name="zh-cn_topic_0000001930317932_p11570122092117"></a><a name="zh-cn_topic_0000001930317932_p11570122092117"></a>--upgrade</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p5570720152111"><a name="zh-cn_topic_0000001930317932_p5570720152111"></a><a name="zh-cn_topic_0000001930317932_p5570720152111"></a>Upgrade the software. If an installation path was specified during installation, the installation path must also be specified during upgrade using the parameter --install-path=&lt;path&gt;.</p>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row106534178110"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p17661618012"><a name="zh-cn_topic_0000001930317932_p17661618012"></a><a name="zh-cn_topic_0000001930317932_p17661618012"></a>--config-file-path=&lt;path&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p18661121811111"><a name="zh-cn_topic_0000001930317932_p18661121811111"></a><a name="zh-cn_topic_0000001930317932_p18661121811111"></a>Configuration file path for <span id="zh-cn_topic_0000001930317932_ph86621218919"><a name="zh-cn_topic_0000001930317932_ph86621218919"></a><a name="zh-cn_topic_0000001930317932_ph86621218919"></a>Docker</span> or <span id="zh-cn_topic_0000001930317932_ph196625181110"><a name="zh-cn_topic_0000001930317932_ph196625181110"></a><a name="zh-cn_topic_0000001930317932_ph196625181110"></a>Containerd</span>. If this parameter is not specified, the following default paths are used.</p>
<a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><a name="zh-cn_topic_0000001930317932_ul1666216181816"></a><ul id="zh-cn_topic_0000001930317932_ul1666216181816"><li><span id="zh-cn_topic_0000001930317932_ph146627186110"><a name="zh-cn_topic_0000001930317932_ph146627186110"></a><a name="zh-cn_topic_0000001930317932_ph146627186110"></a>Docker</span>: /etc/docker/daemon.json</li><li><span id="zh-cn_topic_0000001930317932_ph4662118513"><a name="zh-cn_topic_0000001930317932_ph4662118513"></a><a name="zh-cn_topic_0000001930317932_ph4662118513"></a>Containerd</span>: /etc/containerd/config.toml</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1857082012216"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p65701620122117"><a name="zh-cn_topic_0000001930317932_p65701620122117"></a><a name="zh-cn_topic_0000001930317932_p65701620122117"></a>--install-type=&lt;type&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><div class="p" id="zh-cn_topic_0000001930317932_p155774343616"><a name="zh-cn_topic_0000001930317932_p155774343616"></a><a name="zh-cn_topic_0000001930317932_p155774343616"></a>This parameter is only supported when installing or upgrading <span id="zh-cn_topic_0000001930317932_ph1796213135594"><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a><a name="zh-cn_topic_0000001930317932_ph1796213135594"></a>Ascend Docker Runtime</span> on the following products: <a name="zh-cn_topic_0000001930317932_ul760551653710"></a><a name="zh-cn_topic_0000001930317932_ul760551653710"></a><ul id="zh-cn_topic_0000001930317932_ul760551653710"><li><span id="zh-cn_topic_0000001930317932_ph87811154145311"><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a><a name="zh-cn_topic_0000001930317932_ph87811154145311"></a>Atlas 200 AI acceleration module (RC scenario)</span></li><li><span id="zh-cn_topic_0000001930317932_ph1851111042012"><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a><a name="zh-cn_topic_0000001930317932_ph1851111042012"></a>Atlas 200I A2 acceleration module</span> (RC scenario)</li><li><span id="zh-cn_topic_0000001930317932_ph225916251208"><a name="zh-cn_topic_0000001930317932_ph225916251208"></a><a name="zh-cn_topic_0000001930317932_ph225916251208"></a>Atlas 200I DK A2</span></li><li><span id="zh-cn_topic_0000001930317932_ph271718714435"><a name="zh-cn_topic_0000001930317932_ph271718714435"></a><a name="zh-cn_topic_0000001930317932_ph271718714435"></a>Atlas 200I SoC A1 core board</span></li><li><span id="zh-cn_topic_0000001930317932_ph12573124613552"><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a><a name="zh-cn_topic_0000001930317932_ph12573124613552"></a>Atlas 500 intelligent edge station (model 3000)</span></li><li><span id="zh-cn_topic_0000001930317932_ph11710328131520"><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a><a name="zh-cn_topic_0000001930317932_ph11710328131520"></a>Atlas 500 A2 intelligent edge station</span></li></ul>
</div>
<div class="p" id="zh-cn_topic_0000001930317932_p157201431201014"><a name="zh-cn_topic_0000001930317932_p157201431201014"></a><a name="zh-cn_topic_0000001930317932_p157201431201014"></a>This parameter is used to set the default mount content for <span id="zh-cn_topic_0000001930317932_ph118353873517"><a name="zh-cn_topic_0000001930317932_ph118353873517"></a><a name="zh-cn_topic_0000001930317932_ph118353873517"></a>Ascend Docker Runtime</span>, and must be used together with "--install" in the format --install --install-type=&lt;type&gt;. The optional values for &lt;type&gt; are: <a name="zh-cn_topic_0000001930317932_ul848511715115"></a><a name="zh-cn_topic_0000001930317932_ul848511715115"></a><ul id="zh-cn_topic_0000001930317932_ul848511715115"><li>A200</li><li>A200ISoC</li><li>A200IA2 (supports <span id="zh-cn_topic_0000001930317932_ph1323354011201"><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a><a name="zh-cn_topic_0000001930317932_ph1323354011201"></a>Atlas 200I A2 acceleration module</span> (RC scenario) and <span id="zh-cn_topic_0000001930317932_ph192331940102018"><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a><a name="zh-cn_topic_0000001930317932_ph192331940102018"></a>Atlas 200I DK A2</span>)</li><li>A500</li><li>A500A2</li></ul>
</div>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row14570162052115"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p1857042012112"><a name="zh-cn_topic_0000001930317932_p1857042012112"></a><a name="zh-cn_topic_0000001930317932_p1857042012112"></a>--ce=&lt;ce&gt;</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><a name="zh-cn_topic_0000001930317932_ul4752351238"></a><ul id="zh-cn_topic_0000001930317932_ul4752351238"><li>This option needs to be specified only when iSula is used to start the container. The value needs to be set to isula. It must be used together with --install or --uninstall.</li><li>It cannot be used together with --install-scene. You are advised to use --install-scene instead of --ce. --ce will be discarded in the near future.</li></ul>
</td>
</tr>
<tr id="zh-cn_topic_0000001930317932_row1633572102619"><td class="cellrowborder" valign="top" width="32.43%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0000001930317932_p733611211268"><a name="zh-cn_topic_0000001930317932_p733611211268"></a><a name="zh-cn_topic_0000001930317932_p733611211268"></a>--version</p>
</td>
<td class="cellrowborder" valign="top" width="67.57%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0000001930317932_p83361215264"><a name="zh-cn_topic_0000001930317932_p83361215264"></a><a name="zh-cn_topic_0000001930317932_p83361215264"></a>Query the <span id="zh-cn_topic_0000001930317932_ph7723132765210"><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a><a name="zh-cn_topic_0000001930317932_ph7723132765210"></a>Ascend Docker Runtime</span> version.</p>
</td>
</tr>
</tbody>
</table>
