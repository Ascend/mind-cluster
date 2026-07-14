# Container Manager<a name="ZH-CN_TOPIC_0000002524428759"></a>

## Procedure

Container Manager runs on a physical machine in binary mode.

1. Log in to the server as the `root` user.
2. Upload the obtained Container Manager package to any directory on the server (for example, `/home/container-manager`).
3. Go to the `/home/container-manager` directory and decompress the package.

    ```shell
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

    >[!NOTE] 
    ><i><version\></i> indicates the package version, and <i><arch\></i> indicates the CPU architecture.

4. (Optional) Create a custom fault code configuration file and customize the fault processing level by referring to [(Optional) Configuring Chip Fault Levels](../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#optional-configuring-chip-fault-levels). The following steps do not include this file.
5. Create and edit the `container-manager.service` file.
    1. Run the following commands to create `container-manager.service`:

        ```shell
        vi container-manager.service
        ```

    2. Write the following information to `container-manager.service`. The content in bold in the `ExecStart` field is the startup command. For details about the startup parameters, see [Table 1](#table8724104319141cm). You can modify the parameters as required.

        <pre>
        [Unit]
        Description=Ascend container manager
        Documentation=hiascend.com
        
        [Service]
        ExecStart=/bin/bash -c "<strong>container-manager run -ctrStrategy ringRecover -logPath=/var/log/mindx-dl/container-manager/container-manager.log >/dev/null  2>&1 &</strong>"
        Restart=always
        RestartSec=2
        KillMode=process
        Environment="GOGC=50"
        Environment="GOMAXPROCS=2"
        Environment="GODEBUG=madvdontneed=1"
        Type=forking
        User=root
        Group=root
        
        [Install]
        WantedBy=multi-user.target</pre>

    3. Press `Esc` and enter `:wq!` to save the settings and exit.

6. Create and edit the `container-manager.timer` file. Configuring a timer to start Container Manager after a delay can ensure that the NPU is ready when Container Manager is started.
    1. Run the following commands to create `container-manager.timer`:

        ```shell
        vi container-manager.timer
        ```

    2. Write the following information into `container-manager.timer`.

        <pre>
        [Unit]
        Description=Timer for container manager Service
        
        [Timer]
        # Set a delay for starting Container Manager. Adjust the time as required.
        <strong>OnBootSec=60s</strong> 
        Unit=container-manager.service
        
        [Install]
        WantedBy=timers.target</pre>

    3. Press `Esc` and enter `:wq!` to save the settings and exit.

7. Run the following commands to restart the Container Manager service:

    ```shell
    # Set the Container Manager binary file path.
    cp container-manager /usr/local/bin
    chmod 500 /usr/local/bin/container-manager
    
    # Prepare the Container Manager service file.
    cp container-manager.service /etc/systemd/system
    cp container-manager.timer /etc/systemd/system      
    
    # Start the Container Manager service.
    systemctl enable container-manager.service 
    systemctl enable container-manager.timer 
    systemctl start container-manager.service
    systemctl start container-manager.timer
    ```

## Parameter Description<a name="section2042611570392"></a>

**Table 1** Container Manager startup parameters

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
<div class="note" id="note16897891164"><a name="note16897891164"></a><a name="note16897891164"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><a name="ul370062752110"></a><a name="ul370062752110"></a><ul id="ul370062752110"><li><span id="ph646865823518"><a name="ph646865823518"></a><a name="ph646865823518"></a>Container Manager</span> only performs container start/stop operations when it detects that a chip is in the RestartRequest, RestartBusiness, FreeRestartNPU, or RestartNPU fault state. For details about fault types, see "Fault Code Level Description" in <a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#fault-configuration-description">Fault Configuration Description</a>.</li><li>When the faulty container start/stop strategy is set to singleRecover or ringRecover, users are not supported to specify a container restart policy to enable automatic container restart when starting containers. Choose one of the two options.</li><li>If a container is stopped due to manual intervention, it may cause data inconsistency in the memory of <span id="ph93985387580"><a name="ph93985387580"></a><a name="ph93985387580"></a>Container Manager</span>, leading to abnormal container status.</li></ul>
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
</li></ul><div class="note" id="note1244216377415"><a name="note1244216377415"></a><a name="note1244216377415"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><a name="ul7130194664718"></a><a name="ul7130194664718"></a><ul id="ul7130194664718"><li><span id="ph14779959144911"><a name="ph14779959144911"></a><a name="ph14779959144911"></a>Container Manager</span> can only manage containers started by one container runtime.</li><li>When the container runtime is containerd, only containers whose namespace is not moby can be managed. If containers with the same name exist in multiple namespaces, the container management function may be abnormal.</li></ul>
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
<div class="note" id="note116910214413"><a name="note116910214413"></a><a name="note116910214413"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><a name="ul1246612216016"></a><a name="ul1246612216016"></a><ul id="ul1246612216016"><li>This path is not allowed to be a symbolic link.</li><li>The file permission must be no higher than 640.</li></ul>
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
<div class="note" id="note18966355162717"><a name="note18966355162717"></a><a name="note18966355162717"></a><span class="notetitle">[!NOTE]</span><div class="notebody"><p id="p179661455192711"><a name="p179661455192711"></a><a name="p179661455192711"></a>If the container information queried by status is incorrect, check whether the run service has been terminated or more than one <span id="ph47887203387"><a name="ph47887203387"></a><a name="ph47887203387"></a>Container Manager</span> is started in the environment.</p>
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
