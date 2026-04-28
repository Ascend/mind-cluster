# Container Manager<a name="ZH-CN_TOPIC_0000002524428759"></a>

## 操作步骤

Container Manager组件直接在物理机上通过二进制方式运行。

1. 使用root用户登录服务器。
2. 将获取到的Container Manager软件包上传至服务器的任意目录（以下以“/home/container-manager”目录为例）。
3. 进入“/home/container-manager”目录并进行解压操作。

    ```shell
    unzip Ascend-mindxdl-container-manager_{version}_linux-{arch}.zip
    ```

    >[!NOTE] 
    ><i><version\></i>为软件包的版本号；<i><arch\></i>为CPU架构。

4. （可选）创建自定义故障码配置文件，自定义故障码处理级别。配置及使用详情请参见[（可选）配置芯片故障级别](../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#可选配置芯片故障级别)，以下步骤不体现该文件。
5. 创建并编辑container-manager.service文件。
    1. 执行以下命令，创建container-manager.service文件。

        ```shell
        vi container-manager.service
        ```

    2. 参考如下内容，写入container-manager.service文件中。“ExecStart”字段中加粗的内容为启动命令，启动参数说明请参见[表1](#table8724104319141cm)，用户可以根据实际需要进行修改。

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

    3. 按“Esc”键，输入:wq!保存并退出。

6. 创建并编辑container-manager.timer文件。通过配置timer延时启动，可保证Container Manager启动时NPU卡已就位。
    1. 执行以下命令，创建container-manager.timer文件。

        ```shell
        vi container-manager.timer
        ```

    2. 参考以下示例，并将其写入container-manager.timer文件中。

        <pre>
        [Unit]
        Description=Timer for container manager Service
        
        [Timer]
        # 设置Container Manager延时启动时间，请根据实际情况调整
        <strong>OnBootSec=60s</strong> 
        Unit=container-manager.service
        
        [Install]
        WantedBy=timers.target</pre>

    3. 按“Esc”键，输入:wq!保存并退出。

7. 依次执行以下命令，启用Container Manager服务。

    ```shell
    # 准备Container Manager二进制文件到PATH
    cp container-manager /usr/local/bin
    chmod 500 /usr/local/bin/container-manager
    
    # 准备Container Manager系统服务文件
    cp container-manager.service /etc/systemd/system
    cp container-manager.timer /etc/systemd/system      
    
    # 启动Container Manager系统服务
    systemctl enable container-manager.service 
    systemctl enable container-manager.timer 
    systemctl start container-manager.service
    systemctl start container-manager.timer
    ```

## 参数说明<a name="section2042611570392"></a>

**表 1** Container Manager启动参数

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
<div class="note" id="note16897891164"><a name="note16897891164"></a><a name="note16897891164"></a><span class="notetitle">[!NOTE] 说明</span><div class="notebody"><a name="ul370062752110"></a><a name="ul370062752110"></a><ul id="ul370062752110"><li><span id="ph646865823518"><a name="ph646865823518"></a><a name="ph646865823518"></a>Container Manager</span>在感知到芯片处于RestartRequest、RestartBusiness、FreeRestartNPU和RestartNPU类型故障时，才会进行容器启停操作。故障类型说明请参见<a href="../../../usage/appliance/01_npu_hardware_fault_detection_and_rectification.md#故障配置说明">故障配置说明</a>中“故障码级别说明”。</li><li>当故障容器启停策略配置为singleRecover或者ringRecover时，不支持用户启动容器时指定容器重启策略，使容器自动重启，二者选其一即可。</li><li>若用户手动干预导致容器停止，可能会造成<span id="ph93985387580"><a name="ph93985387580"></a><a name="ph93985387580"></a>Container Manager</span>内存数据混乱，导致容器状态异常。</li></ul>
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
