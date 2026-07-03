# Container Manager<a name="ZH-CN_TOPIC_0000002524428759"></a>

<!-- md-trans-meta sourceCommit=unknown translatedAt=2026-06-09T02:12:40.685Z pushedAt=2026-06-09T06:22:06.788Z -->

Container Manager runs directly on the physical machine in binary mode, providing container lifecycle management, fault detection, and recovery capabilities.

## Procedure

Use the deployment script (`deploy.sh`) for installation. The script automatically completes operations such as binary copying, systemd service file generation, and service start/stop, reducing manual configuration errors.

1. Log in to the server as the `root` user.

2. Upload the obtained Container Manager package to any directory on the server. The following uses the `/home/container-manager` directory as an example (version 26.1.0 and later support installation via a deployment script; for earlier versions, refer to the corresponding historical version documentation).

    >[!NOTE]
    >If the server has network access, you can also download the software package using the following command:
    >
    >```shell
    >wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    >```
    >
    >`<version>` is the version number of the software package; `<arch>` is the CPU architecture (such as x86_64 or aarch64).

3. Go to the directory where the software package is located and decompress the software package.

    ```shell
    cd /home/container-manager
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

4. (Optional) Create a custom fault code configuration file to customize the fault code handling level. For configuration and usage details, see [(Optional) Configuring Chip Fault Levels](../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#optional-configuring-chip-fault-levels). This file is not covered in the following steps.

5. Go to the decompressed directory and run the `deploy.sh` script to install the Container Manager service.

    If the service is already installed, the script will prompt that reinstallation will overwrite the existing configuration and ask for confirmation. Enter `y` to continue the installation, or enter `N` to cancel. No confirmation is required for the first installation. In automated scenarios, use the `-y` parameter to skip the confirmation.

    - Install with default parameters (Docker Runtime, not requiring automatic container recovery):

        ```shell
        bash deploy.sh install
        ```

    - Install with custom parameters. Configure the startup parameters according to the actual environment:

        ```shell
        bash deploy.sh install \
            --runtimeType=containerd \
            --ctrStrategy=ringRecover \
            --logLevel=0 \
            --timerDelay=60s \
            --logPath=/var/log/mindx-dl/container-manager/container-manager.log
        ```

        If information similar to the following is displayed, the installation is successful. The verification result is automatically output after installation. `active (running)` for `Service` indicates that the service started successfully.

        ```shell
        [INFO] Installing container-manager...
        [INFO]   Runtime type : containerd
        [INFO]   Socket path  : /run/containerd/containerd.sock
        [INFO]   CTR strategy : ringRecover (recover all related chip containers)
        [INFO]   Log level    : 0 (info)
        [INFO]   Log path     : /var/log/mindx-dl/container-manager/container-manager.log
        [INFO] Creating log directory...
        [INFO] Installing binary to /usr/local/bin/
        [INFO] Generating systemd service unit...
        [INFO] Generating systemd timer (delay=60s)...
        [INFO] Timer enabled (starts 60s after boot)
        [INFO] Installation completed, verifying...
          Service       : active (running)
          Auto-start    : enabled
          Timer         : active
          Binary        : /usr/local/bin/container-manager  (v26.1.0)

          ✓ All checks passed

        Usage:
          systemctl status container-manager   # Check running status
          journalctl -u container-manager -f   # View live logs
          bash deploy.sh uninstall             # Uninstall
        ```

        >[!NOTE]
        >If the `Service` status in the command output shows `failed` or `inactive`, you can check the service logs for troubleshooting by using the following Command:
        >
        >```shell
        >journalctl -u container-manager -f
        >```

    For details about the options supported by the install command, see [Table 1](#table_deploy_script_options). For details about the meaning of each startup parameter, see [Table 2](#table8724104319141cm).

    More usage examples:

    - Use the containerd runtime and configure the `ringRecover` policy:

        ```shell
        bash deploy.sh install --runtimeType=containerd --ctrStrategy=ringRecover
        ```

    - Use a custom fault configuration file and configure the log level:

        ```shell
        bash deploy.sh install --ctrStrategy=singleRecover --faultConfig=/etc/mindx-dl/container-manager/faultCode.json --logLevel=-1
        ```

    - Use the containerd runtime, and customize the log path and timer delay:

        ```shell
        bash deploy.sh install \
            --runtimeType=containerd \
            --ctrStrategy=ringRecover \
            --logPath=/var/log/mindx-dl/container-manager/container-manager.log \
            --timerDelay=120s \
            --maxAge=30 \
            --maxBackups=10
        ```

## Parameter Description<a name="section2042611570392"></a>

**Table 1** `deploy.sh` commands

<a name="table_deploy_script_options"></a>
<table><thead align="left"><tr id="row_deploy_cmd_header"><th class="cellrowborder" valign="top" width="10.801080108010803%" id="mcps1.2.6.1.1"><p id="p_deploy_cmd_name"><a name="p_deploy_cmd_name"></a><a name="p_deploy_cmd_name"></a>Command</p>
</th>
<th class="cellrowborder" valign="top" width="16.291629162916294%" id="mcps1.2.6.1.2"><p id="p_deploy_cmd_param"><a name="p_deploy_cmd_param"></a><a name="p_deploy_cmd_param"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="11.561156115611562%" id="mcps1.2.6.1.3"><p id="p_deploy_cmd_type"><a name="p_deploy_cmd_type"></a><a name="p_deploy_cmd_type"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="23.342334233423344%" id="mcps1.2.6.1.4"><p id="p_deploy_cmd_default"><a name="p_deploy_cmd_default"></a><a name="p_deploy_cmd_default"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="38.00380038003801%" id="mcps1.2.6.1.5"><p id="p_deploy_cmd_desc"><a name="p_deploy_cmd_desc"></a><a name="p_deploy_cmd_desc"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row_deploy_install_runtime_type"><td class="cellrowborder" rowspan="10" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p_deploy_install_name"><a name="p_deploy_install_name"></a><a name="p_deploy_install_name"></a>install</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_runtime_type_param"><a name="p_deploy_install_runtime_type_param"></a><a name="p_deploy_install_runtime_type_param"></a>--runtimeType</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_runtime_type_type"><a name="p_deploy_install_runtime_type_type"></a><a name="p_deploy_install_runtime_type_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_runtime_type_default"><a name="p_deploy_install_runtime_type_default"></a><a name="p_deploy_install_runtime_type_default"></a>docker</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_runtime_type_desc"><a name="p_deploy_install_runtime_type_desc"></a><a name="p_deploy_install_runtime_type_desc"></a>Container runtime type, corresponding to the binary startup parameter <em>-runtimeType</em>. For details, see <a href="#table8724104319141cm">Table 2</a>. When containerd is selected, if sockPath is not specified, it automatically switches to /run/containerd/containerd.sock.</p>
</td>
</tr>
<tr id="row_deploy_install_sock_path"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_sock_path_param"><a name="p_deploy_install_sock_path_param"></a><a name="p_deploy_install_sock_path_param"></a>--sockPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_sock_path_type"><a name="p_deploy_install_sock_path_type"></a><a name="p_deploy_install_sock_path_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_sock_path_default"><a name="p_deploy_install_sock_path_default"></a><a name="p_deploy_install_sock_path_default"></a>/run/docker.sock</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_sock_path_desc"><a name="p_deploy_install_sock_path_desc"></a><a name="p_deploy_install_sock_path_desc"></a>Socket file path for the container runtime, corresponding to the binary startup parameter <em>-sockPath</em>. For details, see <a href="#table8724104319141cm">Table 2</a>. The path must exist and cannot be a symbolic link.</p>
</td>
</tr>
<tr id="row_deploy_install_ctr_strategy"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_ctr_strategy_param"><a name="p_deploy_install_ctr_strategy_param"></a><a name="p_deploy_install_ctr_strategy_param"></a>--ctrStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_ctr_strategy_type"><a name="p_deploy_install_ctr_strategy_type"></a><a name="p_deploy_install_ctr_strategy_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_ctr_strategy_default"><a name="p_deploy_install_ctr_strategy_default"></a><a name="p_deploy_install_ctr_strategy_default"></a>never</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_ctr_strategy_desc"><a name="p_deploy_install_ctr_strategy_desc"></a><a name="p_deploy_install_ctr_strategy_desc"></a>Fault container start/stop strategy, corresponding to the binary startup parameter <em>-ctrStrategy</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_log_level"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_log_level_param"><a name="p_deploy_install_log_level_param"></a><a name="p_deploy_install_log_level_param"></a>--logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_log_level_type"><a name="p_deploy_install_log_level_type"></a><a name="p_deploy_install_log_level_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_log_level_default"><a name="p_deploy_install_log_level_default"></a><a name="p_deploy_install_log_level_default"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_log_level_desc"><a name="p_deploy_install_log_level_desc"></a><a name="p_deploy_install_log_level_desc"></a>Log level, corresponding to the binary startup parameter <em>-logLevel</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_log_path"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_log_path_param"><a name="p_deploy_install_log_path_param"></a><a name="p_deploy_install_log_path_param"></a>--logPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_log_path_type"><a name="p_deploy_install_log_path_type"></a><a name="p_deploy_install_log_path_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_log_path_default"><a name="p_deploy_install_log_path_default"></a><a name="p_deploy_install_log_path_default"></a>/var/log/mindx-dl/container-manager/container-manager.log</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_log_path_desc"><a name="p_deploy_install_log_path_desc"></a><a name="p_deploy_install_log_path_desc"></a>Log file path, corresponding to the binary startup parameter <em>-logPath</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_max_age"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_max_age_param"><a name="p_deploy_install_max_age_param"></a><a name="p_deploy_install_max_age_param"></a>--maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_max_age_type"><a name="p_deploy_install_max_age_type"></a><a name="p_deploy_install_max_age_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_max_age_default"><a name="p_deploy_install_max_age_default"></a><a name="p_deploy_install_max_age_default"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_max_age_desc"><a name="p_deploy_install_max_age_desc"></a><a name="p_deploy_install_max_age_desc"></a>Log backup retention time, corresponding to the binary startup parameter <em>-maxAge</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_max_backups"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_max_backups_param"><a name="p_deploy_install_max_backups_param"></a><a name="p_deploy_install_max_backups_param"></a>--maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_max_backups_type"><a name="p_deploy_install_max_backups_type"></a><a name="p_deploy_install_max_backups_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_max_backups_default"><a name="p_deploy_install_max_backups_default"></a><a name="p_deploy_install_max_backups_default"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_max_backups_desc"><a name="p_deploy_install_max_backups_desc"></a><a name="p_deploy_install_max_backups_desc"></a>Maximum number of log files retained after rotation, corresponding to the binary startup parameter <em>-maxBackups</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_fault_config"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_fault_config_param"><a name="p_deploy_install_fault_config_param"></a><a name="p_deploy_install_fault_config_param"></a>--faultConfig</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_fault_config_type"><a name="p_deploy_install_fault_config_type"></a><a name="p_deploy_install_fault_config_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_fault_config_default"><a name="p_deploy_install_fault_config_default"></a><a name="p_deploy_install_fault_config_default"></a>"" (empty)</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_fault_config_desc"><a name="p_deploy_install_fault_config_desc"></a><a name="p_deploy_install_fault_config_desc"></a>Path to the custom fault configuration file, corresponding to the binary startup parameter <em>-faultConfigPath</em>. For details, see <a href="#table8724104319141cm">Table 2</a>.</p>
</td>
</tr>
<tr id="row_deploy_install_timer_delay"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_timer_delay_param"><a name="p_deploy_install_timer_delay_param"></a><a name="p_deploy_install_timer_delay_param"></a>--timerDelay</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_timer_delay_type"><a name="p_deploy_install_timer_delay_type"></a><a name="p_deploy_install_timer_delay_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_timer_delay_default"><a name="p_deploy_install_timer_delay_default"></a><a name="p_deploy_install_timer_delay_default"></a>60s</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_timer_delay_desc"><a name="p_deploy_install_timer_delay_desc"></a><a name="p_deploy_install_timer_delay_desc"></a>Delay time for starting Container Manager after system boot, ensuring that NPU devices are ready before starting the service. Supported formats include 60s, 2min, 1h, etc.</p>
</td>
</tr>
<tr id="row_deploy_install_yes"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_yes_param"><a name="p_deploy_install_yes_param"></a><a name="p_deploy_install_yes_param"></a>-y, --yes</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_yes_type"><a name="p_deploy_install_yes_type"></a><a name="p_deploy_install_yes_type"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_yes_default"><a name="p_deploy_install_yes_default"></a><a name="p_deploy_install_yes_default"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_yes_desc"><a name="p_deploy_install_yes_desc"></a><a name="p_deploy_install_yes_desc"></a>Skip the installation confirmation prompt, used for automated script scenarios.</p>
</td>
</tr>
<tr id="row_deploy_uninstall"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p_deploy_uninstall_name"><a name="p_deploy_uninstall_name"></a><a name="p_deploy_uninstall_name"></a>uninstall</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_uninstall_param"><a name="p_deploy_uninstall_param"></a><a name="p_deploy_uninstall_param"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_uninstall_type"><a name="p_deploy_uninstall_type"></a><a name="p_deploy_uninstall_type"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_uninstall_default"><a name="p_deploy_uninstall_default"></a><a name="p_deploy_uninstall_default"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_uninstall_desc"><a name="p_deploy_uninstall_desc"></a><a name="p_deploy_uninstall_desc"></a>Uninstall the Container Manager service, including: stopping and disabling the systemd service and timer, deleting systemd unit files, and deleting binary files. The log directory is retained by default.</p>
</td>
</tr>
<tr id="row_deploy_upgrade"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p_deploy_upgrade_name"><a name="p_deploy_upgrade_name"></a><a name="p_deploy_upgrade_name"></a>upgrade</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_upgrade_param"><a name="p_deploy_upgrade_param"></a><a name="p_deploy_upgrade_param"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_upgrade_type"><a name="p_deploy_upgrade_type"></a><a name="p_deploy_upgrade_type"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_upgrade_default"><a name="p_deploy_upgrade_default"></a><a name="p_deploy_upgrade_default"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_upgrade_desc"><a name="p_deploy_upgrade_desc"></a><a name="p_deploy_upgrade_desc"></a>Upgrade the Container Manager service, including: stopping the service, replacing binary files, and restarting the service. Existing service configurations (such as startup parameters) are retained unchanged.</p>
<div class="note" id="note_deploy_upgrade"><a name="note_deploy_upgrade"></a><a name="note_deploy_upgrade"></a><span class="notetitle">[!NOTE] NOTE</span><div class="notebody"><a name="ul_deploy_upgrade_note"></a><a name="ul_deploy_upgrade_note"></a><ul id="ul_deploy_upgrade_note"><li>Use the upgrade command for upgrades. Using the install command multiple times will overwrite existing service configurations (such as startup parameters), which may cause service configuration loss.</li></ul>
</div></div>
</td>
</tr>
</tbody>
</table>

**Table 2** Container Manager startup parameters

<a name="table8724104319141cm"></a>
<table><thead align="left"><tr id="row57241434113"><th class="cellrowborder" valign="top" width="10.801080108010803%" id="mcps1.2.6.1.1"><p id="p1272416432118"><a name="p1272416432118"></a><a name="p1272416432118"></a>Command</p>
</th>
<th class="cellrowborder" valign="top" width="16.291629162916294%" id="mcps1.2.6.1.2"><p id="p18138161362918"><a name="p18138161362918"></a><a name="p18138161362918"></a>Parameter</p>
</th>
<th class="cellrowborder" valign="top" width="11.561156115611562%" id="mcps1.2.6.1.3"><p id="p1072419431419"><a name="p1072419431419"></a><a name="p1072419431419"></a>Type</p>
</th>
<th class="cellrowborder" valign="top" width="23.342334233423344%" id="mcps1.2.6.1.4"><p id="p1372464316111"><a name="p1372464316111"></a><a name="p1372464316111"></a>Default Value</p>
</th>
<th class="cellrowborder" valign="top" width="38.00380038003801%" id="mcps1.2.6.1.5"><p id="p772517434117"><a name="p772517434117"></a><a name="p772517434117"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row1450614311118"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p5507143131115"><a name="p5507143131115"></a><a name="p5507143131115"></a>help</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p15138141392917"><a name="p15138141392917"></a><a name="p15138141392917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p623516353012"><a name="p623516353012"></a><a name="p623516353012"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p3507243131112"><a name="p3507243131112"></a><a name="p3507243131112"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p15507184331111"><a name="p15507184331111"></a><a name="p15507184331111"></a>View help information.</p>
</td>
</tr>
<tr id="row1494284312299"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p19942104322911"><a name="p19942104322911"></a><a name="p19942104322911"></a>version</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p1942743162912"><a name="p1942743162912"></a><a name="p1942743162912"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p894234312917"><a name="p894234312917"></a><a name="p894234312917"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p39421343132915"><a name="p39421343132915"></a><a name="p39421343132915"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p129421643102918"><a name="p129421643102918"></a><a name="p129421643102918"></a>View the version information of <span id="ph1220617322468"><a name="ph1220617322468"></a><a name="ph1220617322468"></a>Container Manager</span>.</p>
</td>
</tr>
<tr id="row19151746182920"><td class="cellrowborder" rowspan="8" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p215164602914"><a name="p215164602914"></a><a name="p215164602914"></a>run</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p41514652911"><a name="p41514652911"></a><a name="p41514652911"></a>-logPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p106467567226"><a name="p106467567226"></a><a name="p106467567226"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p1364685612219"><a name="p1364685612219"></a><a name="p1364685612219"></a>/var/log/mindx-dl/container-manager/container-manager.log</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p46466565223"><a name="p46466565223"></a><a name="p46466565223"></a>Log file. When a single log file exceeds 20 MB, automatic rotation is triggered. The maximum file size cannot be modified. The naming format of the rotated file is container-manager-<time_of_rotation/>.log, for example: container-manager-2025-11-07T03-38-24.402.log.</p>
</td>
</tr>
<tr id="row17214348192911"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3645125662216"><a name="p3645125662216"></a><a name="p3645125662216"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13645175613228"><a name="p13645175613228"></a><a name="p13645175613228"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9645105618222"><a name="p9645105618222"></a><a name="p9645105618222"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1926353023718"><a name="p1926353023718"></a><a name="p1926353023718"></a>Log level:</p>
<a name="ul15263163018377"></a><a name="ul15263163018377"></a><ul id="ul15263163018377"><li>-1: debug</li><li>0: info</li><li>1: warning</li><li>2: error</li><li>3: critical</li></ul>
</td>
</tr>
<tr id="row14307145012915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p33071750112914"><a name="p33071750112914"></a><a name="p33071750112914"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p335715188373"><a name="p335715188373"></a><a name="p335715188373"></a>Log backup retention period. The value range is [7, 700], in days.</p>
</td>
</tr>
<tr id="row535865213293"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7358952182915"><a name="p7358952182915"></a><a name="p7358952182915"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>Maximum number of rotated log files to retain. The value range is (0, 30], in number of files.</p>
</td>
</tr>
<tr id="row8414634133110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p241417348316"><a name="p241417348316"></a><a name="p241417348316"></a>-ctrStrategy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13414234183112"><a name="p13414234183112"></a><a name="p13414234183112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p134147348319"><a name="p134147348319"></a><a name="p134147348319"></a>never</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9414134153113"><a name="p9414134153113"></a><a name="p9414134153113"></a>Faulty container start/stop strategy:</p>
<a name="ul17352545173818"></a><a name="ul17352545173818"></a><ul id="ul17352545173818"><li>never: Do not start or stop containers.</li><li>singleRecover: Only start/stop the container that mounts the faulty chip. When a fault occurs, stop the container; after the fault is recovered, restart the container.</li><li>ringRecover: Start/stop containers that mount all chips associated with the faulty chip. When a fault occurs, stop the containers; after the fault is recovered, restart the containers.</li></ul>
<div class="note" id="note16897891164"><a name="note16897891164"></a><a name="note16897891164"></a><span class="notetitle">[!NOTE] **NOTE**</span><div class="notebody"><a name="ul370062752110"></a><a name="ul370062752110"></a><ul id="ul370062752110"><li><span id="ph646865823518"><a name="ph646865823518"></a><a name="ph646865823518"></a>Container Manager</span> only performs container start/stop operations when it detects that a chip is in the RestartRequest, RestartBusiness, FreeRestartNPU, or RestartNPU fault state. For details about fault types, see "Fault Code Level Description" in <a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#fault-configuration-description">Fault Configuration Description</a>.</li><li>When the faulty container start/stop strategy is set to singleRecover or ringRecover, users are not supported to specify a container restart policy to enable automatic container restart when starting containers. Choose one of the two options.</li><li>If a container is stopped due to manual intervention, it may cause data inconsistency in the memory of <span id="ph93985387580"><a name="ph93985387580"></a><a name="ph93985387580"></a>Container Manager</span>, leading to abnormal container status.</li></ul>
</div></div>
</td>
</tr>
<tr id="row16901536173117"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1069033663113"><a name="p1069033663113"></a><a name="p1069033663113"></a>-sockPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p969043633119"><a name="p969043633119"></a><a name="p969043633119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13690153610315"><a name="p13690153610315"></a><a name="p13690153610315"></a>/run/docker.sock</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9690143653110"><a name="p9690143653110"></a><a name="p9690143653110"></a>The sock file of the container runtime. This path is not allowed to be a symbolic link.</p>
</td>
</tr>
<tr id="row11407174710314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1407174713310"><a name="p1407174713310"></a><a name="p1407174713310"></a>-runtimeType</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14407247203112"><a name="p14407247203112"></a><a name="p14407247203112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p140711477312"><a name="p140711477312"></a><a name="p140711477312"></a>docker</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p6407647193117"><a name="p6407647193117"></a><a name="p6407647193117"></a>Container runtime type:</p>
<a name="ul8283112164115"></a><a name="ul8283112164115"></a><ul id="ul8283112164115"><li>docker: The container runtime is docker.</li><li>containerd: The container runtime is containerd.
</li></ul><div class="note" id="note1244216377415"><a name="note1244216377415"></a><a name="note1244216377415"></a><span class="notetitle">[!NOTE] **NOTE**</span><div class="notebody"><a name="ul7130194664718"></a><a name="ul7130194664718"></a><ul id="ul7130194664718"><li><span id="ph14779959144911"><a name="ph14779959144911"></a><a name="ph14779959144911"></a>Container Manager</span> can only manage containers started by one container runtime.</li><li>When the container runtime is containerd, only containers whose namespace is not moby can be managed. If containers with the same name exist in multiple namespaces, the container management function may be abnormal.</li></ul>
</div></div>
</td>
</tr>
<tr id="row44581192384"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p945879163814"><a name="p945879163814"></a><a name="p945879163814"></a>-faultConfigPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6458139183820"><a name="p6458139183820"></a><a name="p6458139183820"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3949155543819"><a name="p3949155543819"></a><a name="p3949155543819"></a>""</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16458189133819"><a name="p16458189133819"></a><a name="p16458189133819"></a>Custom fault configuration file path. If not configured, the default fault code configuration is used. For details about custom fault configuration files, see <a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#fault-level-configuration">Fault Level Configuration</a>.</p>
<div class="note" id="note116910214413"><a name="note116910214413"></a><a name="note116910214413"></a><span class="notetitle">[!NOTE] **NOTE**</span><div class="notebody"><a name="ul1246612216016"></a><a name="ul1246612216016"></a><ul id="ul1246612216016"><li>This path is not allowed to be a symbolic link.</li><li>The file permission must be no higher than 640.</li></ul>
</div></div>
</td>
</tr>
<tr id="row441711302328"><td class="cellrowborder" valign="top" width="10.801080108010803%" headers="mcps1.2.6.1.1 "><p id="p0417030143218"><a name="p0417030143218"></a><a name="p0417030143218"></a>status</p>
</td>
<td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p4417103012320"><a name="p4417103012320"></a><a name="p4417103012320"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p041703019329"><a name="p041703019329"></a><a name="p041703019329"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p541719308323"><a name="p541719308323"></a><a name="p541719308323"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p541718306324"><a name="p541718306324"></a><a name="p541718306324"></a>Query container recovery progress, including container ID, status, status start time, and description. For details about container status definitions and change rules, see <a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#container-recovery">Container Recovery</a>.</p>
<div class="note" id="note18966355162717"><a name="note18966355162717"></a><a name="note18966355162717"></a><span class="notetitle">[!NOTE] **NOTE**</span><div class="notebody"><p id="p179661455192711"><a name="p179661455192711"></a><a name="p179661455192711"></a>If the container information queried by status is incorrect, check whether the run service has been terminated or more than one <span id="ph47887203387"><a name="ph47887203387"></a><a name="ph47887203387"></a>Container Manager</span> is started in the environment.</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>After the Container Manager service is started, if you need to modify the startup parameters of Container Manager, modify the startup parameters in the service configuration file and then run the following command to restart the Container Manager service.
>
>```shell
>systemctl daemon-reload && systemctl restart container-manager
>```
>
>Or use the deployment script to reinstall (the script automatically stops the old service and performs an overwrite installation):
>
>```shell
>bash deploy.sh install --ctrStrategy=singleRecover --logLevel=1
>```
