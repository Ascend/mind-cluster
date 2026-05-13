# Container Manager<a name="ZH-CN_TOPIC_0000002524428759"></a>

Container Manager组件直接在物理机上通过二进制方式运行，提供容器生命周期管理、故障检测与恢复功能。

## 操作步骤

使用部署脚本（deploy.sh）进行安装，脚本自动完成二进制拷贝、systemd服务文件生成、服务启停等操作，减少人工配置错误。

1. 使用root用户登录服务器。

2. 将获取到的Container Manager软件包上传至服务器的任意目录，以下以“/home/container-manager”目录为例（26.1.0及以上版本支持部署脚本安装，使用历史版本请参考对应历史版本资料说明）。

    >[!NOTE]
    >若服务器可访问网络，也可通过以下命令下载软件包：
    >
    >```shell
    >wget https://gitcode.com/Ascend/mind-cluster/releases/download/v{version}/Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    >```
    >
    ><i>\<version\></i>为软件包的版本号；<i>\<arch\></i>为CPU架构（如x86_64、aarch64）。

3. 进入软件包所在目录，解压软件包。

    ```shell
    cd /home/container-manager
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

4. （可选）创建自定义故障码配置文件，自定义故障码处理级别。配置及使用详情请参见[（可选）配置芯片故障级别](../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#可选配置芯片故障级别)，以下步骤不体现该文件。

5. 进入解压后的目录，执行deploy.sh脚本安装Container Manager服务。

    若服务已安装，脚本会提示重新安装将覆盖现有配置并要求确认，输入 **y** 继续安装，输入 **N** 取消安装。首次安装无需确认。自动化场景下可使用 **-y** 参数跳过确认。

    - 使用默认参数安装（Docker运行时，不自动恢复容器）：

        ```shell
        bash deploy.sh install
        ```

    - 使用自定义参数安装，请根据实际环境配置启动参数：

        ```shell
        bash deploy.sh install \
            --runtimeType=containerd \
            --ctrStrategy=ringRecover \
            --logLevel=0 \
            --timerDelay=60s \
            --logPath=/var/log/mindx-dl/container-manager/container-manager.log
        ```

        安装成功后，回显示例如下，安装完成会自动输出验证结果，Service为 **active (running)** 表示服务启动成功。

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
        >如果回显中Service显示为**failed**或**inactive**，可通过以下命令查看服务日志排查问题：
        >
        >```shell
        >journalctl -u container-manager -f
        >```

    install命令支持的选项请参见[表1](#table_deploy_script_options)，各启动参数的详细含义请参见[表2](#table8724104319141cm)。

    更多使用示例：

    - 使用containerd运行时，配置ringRecover恢复策略：

        ```shell
        bash deploy.sh install --runtimeType=containerd --ctrStrategy=ringRecover
        ```

    - 使用自定义故障配置文件，并配置日志级别：

        ```shell
        bash deploy.sh install --ctrStrategy=singleRecover --faultConfig=/etc/mindx-dl/container-manager/faultCode.json --logLevel=-1
        ```

    - 使用containerd运行时，自定义日志路径和定时器延迟：

        ```shell
        bash deploy.sh install \
            --runtimeType=containerd \
            --ctrStrategy=ringRecover \
            --logPath=/var/log/mindx-dl/container-manager/container-manager.log \
            --timerDelay=120s \
            --maxAge=30 \
            --maxBackups=10
        ```

## 参数说明<a name="section2042611570392"></a>

**表 1** deploy.sh脚本命令

<a name="table_deploy_script_options"></a>
<table><thead align="left"><tr id="row_deploy_cmd_header"><th class="cellrowborder" valign="top" width="10.801080108010803%" id="mcps1.2.6.1.1"><p id="p_deploy_cmd_name"><a name="p_deploy_cmd_name"></a><a name="p_deploy_cmd_name"></a>命令</p>
</th>
<th class="cellrowborder" valign="top" width="16.291629162916294%" id="mcps1.2.6.1.2"><p id="p_deploy_cmd_param"><a name="p_deploy_cmd_param"></a><a name="p_deploy_cmd_param"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.561156115611562%" id="mcps1.2.6.1.3"><p id="p_deploy_cmd_type"><a name="p_deploy_cmd_type"></a><a name="p_deploy_cmd_type"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="23.342334233423344%" id="mcps1.2.6.1.4"><p id="p_deploy_cmd_default"><a name="p_deploy_cmd_default"></a><a name="p_deploy_cmd_default"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="38.00380038003801%" id="mcps1.2.6.1.5"><p id="p_deploy_cmd_desc"><a name="p_deploy_cmd_desc"></a><a name="p_deploy_cmd_desc"></a>说明</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_runtime_type_desc"><a name="p_deploy_install_runtime_type_desc"></a><a name="p_deploy_install_runtime_type_desc"></a>容器运行时类型，对应二进制启动参数<em>-runtimeType</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。选择containerd时，sockPath未指定则自动切换为/run/containerd/containerd.sock。</p>
</td>
</tr>
<tr id="row_deploy_install_sock_path"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_sock_path_param"><a name="p_deploy_install_sock_path_param"></a><a name="p_deploy_install_sock_path_param"></a>--sockPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_sock_path_type"><a name="p_deploy_install_sock_path_type"></a><a name="p_deploy_install_sock_path_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_sock_path_default"><a name="p_deploy_install_sock_path_default"></a><a name="p_deploy_install_sock_path_default"></a>/run/docker.sock</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_sock_path_desc"><a name="p_deploy_install_sock_path_desc"></a><a name="p_deploy_install_sock_path_desc"></a>容器运行时的sock文件路径，对应二进制启动参数<em>-sockPath</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_ctr_strategy"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_ctr_strategy_param"><a name="p_deploy_install_ctr_strategy_param"></a><a name="p_deploy_install_ctr_strategy_param"></a>--ctrStrategy</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_ctr_strategy_type"><a name="p_deploy_install_ctr_strategy_type"></a><a name="p_deploy_install_ctr_strategy_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_ctr_strategy_default"><a name="p_deploy_install_ctr_strategy_default"></a><a name="p_deploy_install_ctr_strategy_default"></a>never</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_ctr_strategy_desc"><a name="p_deploy_install_ctr_strategy_desc"></a><a name="p_deploy_install_ctr_strategy_desc"></a>故障容器启停策略，对应二进制启动参数<em>-ctrStrategy</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_log_level"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_log_level_param"><a name="p_deploy_install_log_level_param"></a><a name="p_deploy_install_log_level_param"></a>--logLevel</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_log_level_type"><a name="p_deploy_install_log_level_type"></a><a name="p_deploy_install_log_level_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_log_level_default"><a name="p_deploy_install_log_level_default"></a><a name="p_deploy_install_log_level_default"></a>0</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_log_level_desc"><a name="p_deploy_install_log_level_desc"></a><a name="p_deploy_install_log_level_desc"></a>日志级别，对应二进制启动参数<em>-logLevel</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_log_path"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_log_path_param"><a name="p_deploy_install_log_path_param"></a><a name="p_deploy_install_log_path_param"></a>--logPath</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_log_path_type"><a name="p_deploy_install_log_path_type"></a><a name="p_deploy_install_log_path_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_log_path_default"><a name="p_deploy_install_log_path_default"></a><a name="p_deploy_install_log_path_default"></a>/var/log/mindx-dl/container-manager/container-manager.log</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_log_path_desc"><a name="p_deploy_install_log_path_desc"></a><a name="p_deploy_install_log_path_desc"></a>日志文件路径，对应二进制启动参数<em>-logPath</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_max_age"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_max_age_param"><a name="p_deploy_install_max_age_param"></a><a name="p_deploy_install_max_age_param"></a>--maxAge</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_max_age_type"><a name="p_deploy_install_max_age_type"></a><a name="p_deploy_install_max_age_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_max_age_default"><a name="p_deploy_install_max_age_default"></a><a name="p_deploy_install_max_age_default"></a>7</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_max_age_desc"><a name="p_deploy_install_max_age_desc"></a><a name="p_deploy_install_max_age_desc"></a>日志备份时间，对应二进制启动参数<em>-maxAge</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_max_backups"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_max_backups_param"><a name="p_deploy_install_max_backups_param"></a><a name="p_deploy_install_max_backups_param"></a>--maxBackups</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_max_backups_type"><a name="p_deploy_install_max_backups_type"></a><a name="p_deploy_install_max_backups_type"></a>int</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_max_backups_default"><a name="p_deploy_install_max_backups_default"></a><a name="p_deploy_install_max_backups_default"></a>30</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_max_backups_desc"><a name="p_deploy_install_max_backups_desc"></a><a name="p_deploy_install_max_backups_desc"></a>转储后日志文件保留个数上限，对应二进制启动参数<em>-maxBackups</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_fault_config"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_fault_config_param"><a name="p_deploy_install_fault_config_param"></a><a name="p_deploy_install_fault_config_param"></a>--faultConfig</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_fault_config_type"><a name="p_deploy_install_fault_config_type"></a><a name="p_deploy_install_fault_config_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_fault_config_default"><a name="p_deploy_install_fault_config_default"></a><a name="p_deploy_install_fault_config_default"></a>""（空）</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_fault_config_desc"><a name="p_deploy_install_fault_config_desc"></a><a name="p_deploy_install_fault_config_desc"></a>自定义故障配置文件路径，对应二进制启动参数<em>-faultConfigPath</em>，详细说明参见<a href="#table8724104319141cm">表2</a>。</p>
</td>
</tr>
<tr id="row_deploy_install_timer_delay"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_timer_delay_param"><a name="p_deploy_install_timer_delay_param"></a><a name="p_deploy_install_timer_delay_param"></a>--timerDelay</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_timer_delay_type"><a name="p_deploy_install_timer_delay_type"></a><a name="p_deploy_install_timer_delay_type"></a>string</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_timer_delay_default"><a name="p_deploy_install_timer_delay_default"></a><a name="p_deploy_install_timer_delay_default"></a>60s</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_timer_delay_desc"><a name="p_deploy_install_timer_delay_desc"></a><a name="p_deploy_install_timer_delay_desc"></a>系统启动后延时启动Container Manager的时间，确保NPU设备就位后再启动服务。支持格式如60s、2min、1h等。</p>
</td>
</tr>
<tr id="row_deploy_install_yes"><td class="cellrowborder" valign="top" width="16.291629162916294%" headers="mcps1.2.6.1.2 "><p id="p_deploy_install_yes_param"><a name="p_deploy_install_yes_param"></a><a name="p_deploy_install_yes_param"></a>-y, --yes</p>
</td>
<td class="cellrowborder" valign="top" width="11.561156115611562%" headers="mcps1.2.6.1.3 "><p id="p_deploy_install_yes_type"><a name="p_deploy_install_yes_type"></a><a name="p_deploy_install_yes_type"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="23.342334233423344%" headers="mcps1.2.6.1.4 "><p id="p_deploy_install_yes_default"><a name="p_deploy_install_yes_default"></a><a name="p_deploy_install_yes_default"></a>-</p>
</td>
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_install_yes_desc"><a name="p_deploy_install_yes_desc"></a><a name="p_deploy_install_yes_desc"></a>跳过安装确认提示，用于自动化脚本场景。</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_uninstall_desc"><a name="p_deploy_uninstall_desc"></a><a name="p_deploy_uninstall_desc"></a>卸载Container Manager服务，包括：停止并禁用systemd服务和定时器、删除systemd单元文件、删除二进制文件。日志目录默认保留。</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p_deploy_upgrade_desc"><a name="p_deploy_upgrade_desc"></a><a name="p_deploy_upgrade_desc"></a>升级Container Manager服务，包括：停止服务、替换二进制文件、重启服务。保留现有服务配置（启动参数等）不变。</p>
<div class="note" id="note_deploy_upgrade"><a name="note_deploy_upgrade"></a><a name="note_deploy_upgrade"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul_deploy_upgrade_note"></a><a name="ul_deploy_upgrade_note"></a><ul id="ul_deploy_upgrade_note"><li>请使用upgrade命令进行升级，多次使用install命令会覆盖现有的服务配置（如启动参数），可能导致服务配置丢失。</li></ul>
</div></div>
</td>
</tr>
</tbody>
</table>

**表 2** Container Manager启动参数

<a name="table8724104319141cm"></a>
<table><thead align="left"><tr id="row57241434113"><th class="cellrowborder" valign="top" width="10.801080108010803%" id="mcps1.2.6.1.1"><p id="p1272416432118"><a name="p1272416432118"></a><a name="p1272416432118"></a>命令</p>
</th>
<th class="cellrowborder" valign="top" width="16.291629162916294%" id="mcps1.2.6.1.2"><p id="p18138161362918"><a name="p18138161362918"></a><a name="p18138161362918"></a>参数</p>
</th>
<th class="cellrowborder" valign="top" width="11.561156115611562%" id="mcps1.2.6.1.3"><p id="p1072419431419"><a name="p1072419431419"></a><a name="p1072419431419"></a>类型</p>
</th>
<th class="cellrowborder" valign="top" width="23.342334233423344%" id="mcps1.2.6.1.4"><p id="p1372464316111"><a name="p1372464316111"></a><a name="p1372464316111"></a>默认值</p>
</th>
<th class="cellrowborder" valign="top" width="38.00380038003801%" id="mcps1.2.6.1.5"><p id="p772517434117"><a name="p772517434117"></a><a name="p772517434117"></a>说明</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p15507184331111"><a name="p15507184331111"></a><a name="p15507184331111"></a>查看帮助信息。</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p129421643102918"><a name="p129421643102918"></a><a name="p129421643102918"></a>查看<span id="ph1220617322468"><a name="ph1220617322468"></a><a name="ph1220617322468"></a>Container Manager</span>的版本信息。</p>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p46466565223"><a name="p46466565223"></a><a name="p46466565223"></a>日志文件。单个日志文件超过20MB时，会触发自动转储功能，文件大小上限不支持修改。转储后文件的命名格式为container-manager-触发转储的时间.log，例如：container-manager-2025-11-07T03-38-24.402.log。</p>
</td>
</tr>
<tr id="row17214348192911"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p3645125662216"><a name="p3645125662216"></a><a name="p3645125662216"></a>-logLevel</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13645175613228"><a name="p13645175613228"></a><a name="p13645175613228"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p9645105618222"><a name="p9645105618222"></a><a name="p9645105618222"></a>0</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p1926353023718"><a name="p1926353023718"></a><a name="p1926353023718"></a>日志级别：</p>
<a name="ul15263163018377"></a><a name="ul15263163018377"></a><ul id="ul15263163018377"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：critical</li></ul>
</td>
</tr>
<tr id="row14307145012915"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p33071750112914"><a name="p33071750112914"></a><a name="p33071750112914"></a>-maxAge</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p8615148104614"><a name="p8615148104614"></a><a name="p8615148104614"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p1461554818467"><a name="p1461554818467"></a><a name="p1461554818467"></a>7</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p335715188373"><a name="p335715188373"></a><a name="p335715188373"></a>日志备份时间，取值范围为[7, 700]，单位为天。</p>
</td>
</tr>
<tr id="row535865213293"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p7358952182915"><a name="p7358952182915"></a><a name="p7358952182915"></a>-maxBackups</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p206151748184617"><a name="p206151748184617"></a><a name="p206151748184617"></a>int</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p17615154874610"><a name="p17615154874610"></a><a name="p17615154874610"></a>30</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16151648144619"><a name="p16151648144619"></a><a name="p16151648144619"></a>转储后日志文件保留个数上限，取值范围为(0, 30]，单位为个。</p>
</td>
</tr>
<tr id="row8414634133110"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p241417348316"><a name="p241417348316"></a><a name="p241417348316"></a>-ctrStrategy</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p13414234183112"><a name="p13414234183112"></a><a name="p13414234183112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p134147348319"><a name="p134147348319"></a><a name="p134147348319"></a>never</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9414134153113"><a name="p9414134153113"></a><a name="p9414134153113"></a>故障容器启停策略：</p>
<a name="ul17352545173818"></a><a name="ul17352545173818"></a><ul id="ul17352545173818"><li>never：不进行容器启停。</li><li>singleRecover：仅启停单个挂载故障芯片的容器。故障产生时，停止容器；故障恢复后，将容器重新拉起。</li><li>ringRecover：启停挂载故障芯片所关联的所有芯片的容器。故障产生时，停止容器；故障恢复后，将容器重新拉起。</li></ul>
<div class="note" id="note16897891164"><a name="note16897891164"></a><a name="note16897891164"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul370062752110"></a><a name="ul370062752110"></a><ul id="ul370062752110"><li><span id="ph646865823518"><a name="ph646865823518"></a><a name="ph646865823518"></a>Container Manager</span>在感知到芯片处于RestartRequest、RestartBusiness、FreeRestartNPU和RestartNPU类型故障时，才会进行容器启停操作。故障类型说明请参见<a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#故障配置说明">故障配置说明</a>中"故障码级别说明"。</li><li>当故障容器启停策略配置为singleRecover或者ringRecover时，不支持用户启动容器时指定容器重启策略，使容器自动重启，二者选其一即可。</li><li>若用户手动干预导致容器停止，可能会造成<span id="ph93985387580"><a name="ph93985387580"></a><a name="ph93985387580"></a>Container Manager</span>内存数据混乱，导致容器状态异常。</li></ul>
</div></div>
</td>
</tr>
<tr id="row16901536173117"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1069033663113"><a name="p1069033663113"></a><a name="p1069033663113"></a>-sockPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p969043633119"><a name="p969043633119"></a><a name="p969043633119"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p13690153610315"><a name="p13690153610315"></a><a name="p13690153610315"></a>/run/docker.sock</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p9690143653110"><a name="p9690143653110"></a><a name="p9690143653110"></a>容器运行时的sock文件，该路径不允许为软链接。</p>
</td>
</tr>
<tr id="row11407174710314"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p1407174713310"><a name="p1407174713310"></a><a name="p1407174713310"></a>-runtimeType</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p14407247203112"><a name="p14407247203112"></a><a name="p14407247203112"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p140711477312"><a name="p140711477312"></a><a name="p140711477312"></a>docker</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p6407647193117"><a name="p6407647193117"></a><a name="p6407647193117"></a>容器运行时类型：</p>
<a name="ul8283112164115"></a><a name="ul8283112164115"></a><ul id="ul8283112164115"><li>docker：容器运行时为docker。</li><li>containerd：容器运行时为containerd。
</li></ul><div class="note" id="note1244216377415"><a name="note1244216377415"></a><a name="note1244216377415"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul7130194664718"></a><a name="ul7130194664718"></a><ul id="ul7130194664718"><li><span id="ph14779959144911"><a name="ph14779959144911"></a><a name="ph14779959144911"></a>Container Manager</span>仅支持管理一种容器运行时启动的容器。</li><li>当容器运行时为containerd时，仅支持管理命名空间不为moby的容器。当多个命名空间下有相同名称的容器，容器管理功能可能会出现异常。</li></ul>
</div></div>
</td>
</tr>
<tr id="row44581192384"><td class="cellrowborder" valign="top" headers="mcps1.2.6.1.1 "><p id="p945879163814"><a name="p945879163814"></a><a name="p945879163814"></a>-faultConfigPath</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.2 "><p id="p6458139183820"><a name="p6458139183820"></a><a name="p6458139183820"></a>string</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.3 "><p id="p3949155543819"><a name="p3949155543819"></a><a name="p3949155543819"></a>""</p>
</td>
<td class="cellrowborder" valign="top" headers="mcps1.2.6.1.4 "><p id="p16458189133819"><a name="p16458189133819"></a><a name="p16458189133819"></a>自定义故障配置文件路径。若不配置，则使用默认的故障码配置。自定义故障配置文件详情请参见<a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#故障级别配置">故障级别配置</a>。</p>
<div class="note" id="note116910214413"><a name="note116910214413"></a><a name="note116910214413"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul1246612216016"></a><a name="ul1246612216016"></a><ul id="ul1246612216016"><li>该路径不允许为软链接。</li><li>该文件权限需不高于640。</li></ul>
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
<td class="cellrowborder" valign="top" width="38.00380038003801%" headers="mcps1.2.6.1.5 "><p id="p541718306324"><a name="p541718306324"></a><a name="p541718306324"></a>查询容器恢复进度，包括容器ID、状态、状态开始时间及描述信息。容器的状态定义及变化规则详细请参见<a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#容器恢复">容器恢复</a>。</p>
<div class="note" id="note18966355162717"><a name="note18966355162717"></a><a name="note18966355162717"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><p id="p179661455192711"><a name="p179661455192711"></a><a name="p179661455192711"></a>如果status查询到的容器信息有误，需确认run服务是否已经终止，或者环境上启动了一个以上的<span id="ph47887203387"><a name="ph47887203387"></a><a name="ph47887203387"></a>Container Manager</span>。</p>
</div></div>
</td>
</tr>
</tbody>
</table>

>[!NOTE]
>Container Manager服务已经启动后，若需要修改Container Manager的启动参数，请修改服务配置文件中的启动参数后，执行以下命令，重启Container Manager系统服务。
>
>```shell
>systemctl daemon-reload && systemctl restart container-manager
>```
>
>或者使用部署脚本重新安装（脚本会自动停止旧服务并覆盖安装）：
>
>```shell
>bash deploy.sh install --ctrStrategy=singleRecover --logLevel=1
>```
