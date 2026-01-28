# 一体机特性指南<a name="ZH-CN_TOPIC_0000002524340741"></a>

## NPU硬件故障检测与恢复<a name="ZH-CN_TOPIC_0000002518340693"></a>

在无K8s的场景下，训练或推理进程异常后，没有有效的恢复手段。为了解决上述问题，可以配置容器恢复特性。若要支持容器恢复特性，需要安装Container Manager组件。Container Manager组件的安装操作详细请参见[安装部署](../installation_guide.md#安装部署)。

### 特性说明<a name="ZH-CN_TOPIC_0000002486738074"></a>

<a name="table1866285218270"></a>
<table><thead align="left"><tr id="row7663135222713"><th class="cellrowborder" valign="top" width="14.931493149314932%" id="mcps1.1.4.1.1"><p id="p266355252712"><a name="p266355252712"></a><a name="p266355252712"></a>功能名称</p>
</th>
<th class="cellrowborder" valign="top" width="60.14601460146015%" id="mcps1.1.4.1.2"><p id="p066313523276"><a name="p066313523276"></a><a name="p066313523276"></a>说明</p>
</th>
<th class="cellrowborder" valign="top" width="24.922492249224923%" id="mcps1.1.4.1.3"><p id="p866385212720"><a name="p866385212720"></a><a name="p866385212720"></a>原理介绍及配置步骤</p>
</th>
</tr>
</thead>
<tbody><tr id="row1466395215279"><td class="cellrowborder" valign="top" width="14.931493149314932%" headers="mcps1.1.4.1.1 "><p id="p16631352172718"><a name="p16631352172718"></a><a name="p16631352172718"></a>故障检测</p>
</td>
<td class="cellrowborder" valign="top" width="60.14601460146015%" headers="mcps1.1.4.1.2 "><p id="p1376115912478"><a name="p1376115912478"></a><a name="p1376115912478"></a>该特性具有故障检测功能，支持实时监测350+硬件类故障的故障检测。</p>
</td>
<td class="cellrowborder" valign="top" width="24.922492249224923%" headers="mcps1.1.4.1.3 "><p id="p3664115213274"><a name="p3664115213274"></a><a name="p3664115213274"></a><a href="#故障检测">故障检测</a></p>
</td>
</tr>
<tr id="row8664195222715"><td class="cellrowborder" valign="top" width="14.931493149314932%" headers="mcps1.1.4.1.1 "><p id="p9664252172717"><a name="p9664252172717"></a><a name="p9664252172717"></a>故障处理</p>
</td>
<td class="cellrowborder" valign="top" width="60.14601460146015%" headers="mcps1.1.4.1.2 "><p id="p1411817177434"><a name="p1411817177434"></a><a name="p1411817177434"></a>该特性具有故障处理功能，针对故障级别配置为RestartRequestCodes、RestartBusinessCodes、FreeRestartNPUCodes和RestartNPUCodes的故障，故障发生后不需要人工介入就可自动恢复故障设备。</p>
</td>
<td class="cellrowborder" valign="top" width="24.922492249224923%" headers="mcps1.1.4.1.3 "><p id="p166425282717"><a name="p166425282717"></a><a name="p166425282717"></a><a href="#故障处理">故障处理</a></p>
</td>
</tr>
<tr id="row1781918488293"><td class="cellrowborder" valign="top" width="14.931493149314932%" headers="mcps1.1.4.1.1 "><p id="p381964862916"><a name="p381964862916"></a><a name="p381964862916"></a>容器恢复</p>
</td>
<td class="cellrowborder" valign="top" width="60.14601460146015%" headers="mcps1.1.4.1.2 "><p id="p16958101123713"><a name="p16958101123713"></a><a name="p16958101123713"></a>该特性具有容器恢复功能，用户可配置容器启停的策略，针对故障级别配置为RestartRequestCodes、RestartBusinessCodes、FreeRestartNPUCodes和RestartNPUCodes的故障，故障发生时将容器停止，故障恢复后重新将容器拉起。</p>
</td>
<td class="cellrowborder" valign="top" width="24.922492249224923%" headers="mcps1.1.4.1.3 "><p id="p17819948202912"><a name="p17819948202912"></a><a name="p17819948202912"></a><a href="#容器恢复">容器恢复</a></p>
</td>
</tr>
</tbody>
</table>

>[!NOTE] 说明 
>本特性不适用于算力虚拟化场景，不支持共享设备特性及混插模式。


### 故障检测<a name="ZH-CN_TOPIC_0000002518738073"></a>

#### 方案说明<a name="ZH-CN_TOPIC_0000002518738601"></a>

Container Manager启动时会先注册DCMI故障订阅接口，故障发生时驱动通过该接口将故障事件上报给Container Manager，故障恢复时通过该接口将恢复事件上报给Container Manager。

NPU发生故障时，故障管理框架获取到故障信息后，将该信息上报给NPU驱动的故障管理框架。故障管理框架收到故障信息后，通过DCMI接口上报给Container Manager，如[图1](#fig610813710515)所示。

**图 1**  故障检测原理图<a name="fig610813710515"></a>  
![](../../figures/scheduling/故障检测原理图.png "故障检测原理图")


#### 故障级别配置<a name="ZH-CN_TOPIC_0000002518737701"></a>

##### 故障配置说明<a name="ZH-CN_TOPIC_0000002486577908"></a>

针对芯片故障的不同级别进行分级处理时，Container Manager组件会获取到当前故障的故障码，根据故障码处理级别，对故障进行相应处理。

**默认故障码配置<a name="section44743445012"></a>**

Container Manager启动后，会默认按照如下配置作为当前故障处理依据：

```
{
  "NotHandleFaultCodes":[
    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    "80CD8006","80CD8003","80A38006","80A38003","80A58006","80A58003","80DE1805","80F18006","80F18003","80DF8006",
    "80E01805","80E18400","80E01809","80E18401","80E00209","80F38006","80F38003","80E18006","80D38009","819B800D",
    "80DD8008","80DD8007","80B98006","80BD8006","819B8006","80DE1803","819D8000","81998006","81978006","81978004",
    "815F8006","815F8004","81338006","81338004","817F8006","817F8004","816F8006","816F8004","814F8006","814F8004",
    "81938006","81938004","81478006","81478004","813B8006","813B8004","81578006","81578004","81958006","81958004",
    "81078603","8C2FA009","A4025021","A60250C1","A4025081","A214000D","A414000D","A4028801","A4025101","A2140007",
    "A4140007","A2140008","A4140008","A40250E1","A214000A","A414000A","A4025061","A4025041","A214000B","A414000B",
    "A414000C","A2140009","A4140009","A4303002","80B78006","80B78005","80E1800F","80DE0200","814D8006","8C1F860B",
    "8C1F8608","4C1F8608","819B8003","80DF8401","80DF8400","80818200","80818201","80818202","80818203","80818204",
    "80818205","80F38009","81A3880C","81AD8605","80E20207","81078605","80DE0207","8C2FA001","819B8605","80818C06",
    "8C1F860A","80E18405"
  ],
  "RestartRequestCodes":[
    "80C98008","80C98002","80C98003","80C98009","80CB8002","80CB8008","80CB8009","80CF8003","81318008","80D58000",
    "80D58009","80D98008","80DB800A","80DB8000","80DD8000","80DD8003","80C98000","81AB800D","81AB8003","80BD8000",
    "80BB8009","80BD8003","80BD8009","80BB8000","80BB8003","80BB8008","80BB800A","81AB8008","80C9800A","80CB800A"
  ],
  "RestartBusinessCodes":[
    "8C204E00","A8028802","A4302003","A4302004","A4302005","A4302006","A4302009","A430200A", "A6301002","B4060011",
    "B406009C","B4060008","B4060009","B406000E","A60250A1","A2301001","A2301002","A2303001", "B4060006","B4060007",
    "B406000D","B4060014","B4060010","B4060011","80E01801", "81B38009","81B38004"
  ],
  "FreeRestartNPUCodes":[
    "8C0E4E00","8C104E00","8C0C4E00","8C044E00","8C064E00","8C17A005","8C1DA005","8C19A005","8C0A4E00","8C084E00",
    "A4193217","A4193218","A42A0000","A42F3917","A42F3918","8C464E00","8C124E00"
  ],
  "RestartNPUCodes":[
    "8C03A000","8C1FA006","40F84E00","80E24E00","80E21E01","80E38008","80E3A202","80E3A203","80E39200","819B800A",
    "80E2120D","80E78000","80E78008","80FA4E00","812E4E00","80C78008","80F78009","80F78008","80F78003","80E18404",
    "80FB8005","80A18008","80CD8008","80A38008","80A58008","80DE1801","80F18008","80F18000","80F1800A","80CF8000",
    "80DF8000","80DF8009","80DF8008","80DF800A","80F38008","80F2180D","80E18005","80E18008","80E1800A","812F8000",
    "80B98000","80B98008","80BD8008","80CB8001","81998009","81998008","81978008","815F8008","81338008","817F8008",
    "81478008","813B8008","81578008","81958008","A2141004","A2141006","A2142004","A2142006","A2145004","A4183200",
    "A6023001","A6023002","A6023003","A6023004","A6060000","A6060001","A6060002","A6060003","A6060004","A6060005",
    "A606000A","A606000B","A606000C","A606000F","A606009D","A6060FFF","A607FFFF","A6140001","A6140002","A6140003",
    "A6140004","A6140005","A6140006","A6141003","A6142003","A6143003","A6144003","A6145003","A6192D15","A6193206",
    "A6193215","A6193248","A62F3905","A62FFFFF","A6303003","A6303004","A6360000","A6361000","A6362000","A8021004",
    "A8060FFF","A807FFFF","A82A0000","80B78000","80B58000","81498004","80F78C02","80F78C03","80F78C04","81B38008",
    "80E18000","80E21008","80C98001","80E58005","80E58009","80E58E02","80E58E03","816F8008","814F8008","81938008",
    "80E44E00","80CF8009","80CF8008","813B8002","81338002","81578002","81958002","81938002","81478002","81978002",
    "815F8002","81C9800A","81C7800A","81C5800A","813F800A","8139800A","8145800A","8C4BA00C","80E3A207"
  ],
  "SeparateNPUCodes":[
    "80E3A201","80E18402","80E0020B","817F8002","816F8002","814F8002","9419321B","A2301000","A2301001","A2302001",
    "A4192C1A","A4193216","A419321B","A419321C","A42F390F","A42F3916","A42F391A","A6183207","A62F3934","A8028801",
    "A819320F","A8193234","A8193235","80818c00","80818C05","80DF8402","80818C00","4C4BA00C"
  ]
}
```

>[!NOTE] 说明 
>-   每个故障对应的故障信息请参见[芯片故障码参考文档](../appendix.md#芯片故障码参考文档)。
>-   芯片故障支持配置的故障级别请参见[故障码级别说明](#zh-cn_topic_0000002171521445_section5245155017242)。

**故障码级别说明<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

Container Manager从驱动获取到芯片故障码后，根据故障码对设备及业务的影响将故障划分为以下六种级别，详细说明请参见[表1](#zh-cn_topic_0000002171521445_table7618951152212)。若用户需要修改故障码的故障级别，请参见[（可选）配置芯片故障级别](#可选配置芯片故障级别)。

**表 1**  故障级别及处理说明

<a name="zh-cn_topic_0000002171521445_table7618951152212"></a>
<table><thead align="left"><tr id="zh-cn_topic_0000002171521445_row461812518228"><th class="cellrowborder" valign="top" width="23.830000000000002%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002171521445_p12618851162220"><a name="zh-cn_topic_0000002171521445_p12618851162220"></a><a name="zh-cn_topic_0000002171521445_p12618851162220"></a>故障级别</p>
</th>
<th class="cellrowborder" valign="top" width="44.73%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002171521445_p16618125162219"><a name="zh-cn_topic_0000002171521445_p16618125162219"></a><a name="zh-cn_topic_0000002171521445_p16618125162219"></a>NPU复位策略</p>
</th>
<th class="cellrowborder" valign="top" width="31.44%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002171521445_p171971327125410"><a name="zh-cn_topic_0000002171521445_p171971327125410"></a><a name="zh-cn_topic_0000002171521445_p171971327125410"></a>容器处理策略</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0000002171521445_row961811511228"><td class="cellrowborder" valign="top" width="23.830000000000002%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p7618125114229"><a name="zh-cn_topic_0000002171521445_p7618125114229"></a><a name="zh-cn_topic_0000002171521445_p7618125114229"></a>NotHandleFault</p>
</td>
<td class="cellrowborder" valign="top" width="44.73%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p1261835110227"><a name="zh-cn_topic_0000002171521445_p1261835110227"></a><a name="zh-cn_topic_0000002171521445_p1261835110227"></a>对业务无影响的故障，无需处理。</p>
</td>
<td class="cellrowborder" valign="top" width="31.44%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p719714273546"><a name="zh-cn_topic_0000002171521445_p719714273546"></a><a name="zh-cn_topic_0000002171521445_p719714273546"></a>暂不处理。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row116184515226"><td class="cellrowborder" valign="top" width="23.830000000000002%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p5618751102216"><a name="zh-cn_topic_0000002171521445_p5618751102216"></a><a name="zh-cn_topic_0000002171521445_p5618751102216"></a>RestartRequest</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" width="44.73%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p05771854113911"><a name="zh-cn_topic_0000002171521445_p05771854113911"></a><a name="zh-cn_topic_0000002171521445_p05771854113911"></a><span id="ph6926121810160"><a name="ph6926121810160"></a><a name="ph6926121810160"></a>Container Manager</span>在故障持续60秒后，将故障芯片和关联芯片加入到待复位芯片缓存中。待复位芯片缓存复位逻辑详细请参见<a href="#故障处理">故障处理</a>。</p>
</td>
<td class="cellrowborder" rowspan="4" valign="top" width="31.44%" headers="mcps1.2.4.1.3 "><p id="p11041540152412"><a name="p11041540152412"></a><a name="p11041540152412"></a>当命令run的启动参数<span class="parmname" id="parmname127339182715"><a name="parmname127339182715"></a><a name="parmname127339182715"></a>“-ctrStrategy”</span>配置为<span class="parmvalue" id="parmvalue058714462711"><a name="parmvalue058714462711"></a><a name="parmvalue058714462711"></a>“singleRecover”</span>或者<span class="parmvalue" id="parmvalue1923725032719"><a name="parmvalue1923725032719"></a><a name="parmvalue1923725032719"></a>“ringRecover”</span>时，开启容器启停功能。两个配置参数的差异请参见<a href="../installation_guide.md#container-manager">安装Container Manager</a>中"Container Manager启动参数"表。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row14618105116225"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p15618851132212"><a name="zh-cn_topic_0000002171521445_p15618851132212"></a><a name="zh-cn_topic_0000002171521445_p15618851132212"></a>RestartBusiness</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row561825132214"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p66188511222"><a name="zh-cn_topic_0000002171521445_p66188511222"></a><a name="zh-cn_topic_0000002171521445_p66188511222"></a>FreeRestartNPU</p>
</td>
<td class="cellrowborder" rowspan="2" valign="top" headers="mcps1.2.4.1.2 "><p id="p17596495131"><a name="p17596495131"></a><a name="p17596495131"></a><span id="ph1559049101318"><a name="ph1559049101318"></a><a name="ph1559049101318"></a>Container Manager</span>收到故障后，立即将故障芯片和关联芯片加入到待复位芯片缓存中。待复位芯片缓存复位逻辑详细请参见<a href="#故障处理">故障处理</a>。</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row14618125152210"><td class="cellrowborder" valign="top" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p17618155116227"><a name="zh-cn_topic_0000002171521445_p17618155116227"></a><a name="zh-cn_topic_0000002171521445_p17618155116227"></a>RestartNPU</p>
</td>
</tr>
<tr id="zh-cn_topic_0000002171521445_row1061895115227"><td class="cellrowborder" valign="top" width="23.830000000000002%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002171521445_p961885142215"><a name="zh-cn_topic_0000002171521445_p961885142215"></a><a name="zh-cn_topic_0000002171521445_p961885142215"></a>SeparateNPU</p>
</td>
<td class="cellrowborder" valign="top" width="44.73%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002171521445_p18618151202216"><a name="zh-cn_topic_0000002171521445_p18618151202216"></a><a name="zh-cn_topic_0000002171521445_p18618151202216"></a>故障无法通过复位恢复，需要隔离芯片。</p>
</td>
<td class="cellrowborder" valign="top" width="31.44%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002171521445_p019742745411"><a name="zh-cn_topic_0000002171521445_p019742745411"></a><a name="zh-cn_topic_0000002171521445_p019742745411"></a>暂不处理。</p>
</td>
</tr>
</tbody>
</table>


##### （可选）配置芯片故障级别<a name="ZH-CN_TOPIC_0000002486737872"></a>

如果用户想要自定义故障级别，可以创建自定义故障码配置文件，启动Container Manager组件时，作为“-faultConfigPath“参数的值传入即可。以故障名称dmp\_daemon节点状态检测异常，对应故障码80E21007为例。将当前故障的处理策略NotHandleFault修改为RestartNPU的操作示例如下。

1.  登录环境，进入任意目录（以下以“/home/container-manager“目录为例）。
2.  创建自定义故障码配置文件，以文件名为faultCode.json为例。

    ```
    vi faultCode.json
    ```

3.  按“i”进入编辑模式，将[默认故障码配置](#故障配置说明)中的默认故障码配置拷贝到该文件中。
4.  找到故障码80E21007。

    ```
    "NotHandleFaultCodes":[
       "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE] 说明 
    >同一故障码配置在多个故障级别中，会显示设置成功，但默认按照高等级故障处理。

5.  将故障码80E21007在**NotHandleFaultCodes**中删除，并添加到**RestartNPUCodes**中。

    ```
    "NotHandleFaultCodes":[ 
       "80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    "RestartNPUCodes":[
       "8C204E00","A8028802","A4302003","A4302004","A4302005","A4302006","A4302009","A430200A","80CF8009","80CF8008","80E21007",... 
    ...
       ],
    ```

6.  修改完成后，按“Esc”键，输入:wq!保存并退出。
7.  启动Container Manager。如果Container Manager服务已经安装完成，需要重启Container Manager服务使得配置生效。

    ```
    systemctl daemon-reload && systemctl restart container-manager.service # 重新加载服务配置，且重启已经安装完成的Container Manager服务
    ```

    若日志出现“load custom fault config file from /home/container-manager/faultCode.json success”，表示自定义配置故障码操作成功。

>[!NOTICE] 须知 
>-   故障码配置为系统配置，若用户无特殊需求，请勿随意修改，否则可能会导致系统故障处理功能出错。
>-   自定义故障码配置文件被修改后，需要重启Container Manager使其生效。如果自定义的配置文件内容存在格式错误等问题，Container Manager会直接报错退出。




### 故障处理<a name="ZH-CN_TOPIC_0000002486738174"></a>

Container Manager在RestartRequest和RestartBusiness故障持续60秒，或者获取到FreeRestartNPU和RestartNPU类型故障时，将故障芯片和关联芯片放入待复位缓存中。Container Manager会周期性尝试复位待复位缓存中的芯片，当芯片满足以下条件时，Container Manager调用DCMI接口执行芯片复位操作。

-   当前故障芯片和关联芯片上不存在任务进程。
-   当前故障芯片和关联芯片没有被正在运行的容器占用。
-   当前故障芯片或关联芯片依然存在任意最高为RestartRequest、RestartBusiness、FreeRestartNPU、RestartNPU四种级别的故障。

>[!NOTE] 说明 
>-   Container Manager在周期内成功执行了芯片复位并获取到芯片成功启动的结果后，故障复位功能会暂停30秒等待芯片初始化完成。
>-   芯片连续复位失败3次以后，Container Manager不再尝试复位此芯片。


### 容器恢复<a name="ZH-CN_TOPIC_0000002486578214"></a>

Container Manager在感知到芯片处于RestartRequest、RestartBusiness、FreeRestartNPU和RestartNPU类型故障时，会按照命令run的启动参数“-ctrStrategy“配置的重启策略，进行容器停止与恢复。具体的容器停止与恢复的范围请参见<a href="../installation_guide.md#container-manager">安装Container Manager</a>中"Container Manager启动参数"表。

容器启停过程中，会发生状态变化：

-   当容器正在停止时，容器状态为pausing。当容器状态为pausing，且状态持续时间超过20s时，通过status命令查询到的容器描述信息为"Container pause may fail. Please manually delete the container"。
-   容器停止后，容器状态会变更为paused。当容器状态为paused，且状态持续时间超过400s时，通过status命令查询到的容器描述信息为"Device hot reset may fail. Please check of device status and recovery are required"。
-   当容器正在恢复时，容器状态为resuming。当容器状态为resuming，且状态持续时间超过20s时，通过status命令查询到的容器描述信息为"The device has been recovered, but the container failed to be resumed. Please manually pull up the container"。
-   其余时间，容器状态均为running，描述信息提示为“normal”，通过status命令查询到的容器状态开始时间为Container Manager感知到容器启动的时间或者容器恢复后的时间。

>[!NOTE] 说明 
>-   Container Manager仅恢复由它本身停止的容器。
>-   上述涉及到的容器启停过程中的容器状态，仅为Container Manager自定义，非容器运行时给出的官方定义。
>-   当容器运行时为containerd时，针对没有task的容器，启停会报错。



## 配置推理业务故障恢复<a name="ZH-CN_TOPIC_0000002511630975"></a>

在一体机混部或无K8s的场景下，推理进程异常后，没有有效的恢复手段。本章节提供了推理业务故障后自动恢复的示例。示例中启动脚本作为容器entrypoint，自动拉起推理进程，监控推理进程状态，并在异常后重新拉起推理进程。

-   支持MindIE Server单机推理。
-   不支持MindIE Server多机推理。仅重启其中一个容器中的推理进程，业务无法恢复。

**操作步骤<a name="section169801610181818"></a>**

以下配置过程以Qwen3-1.7B模型为例。

1.  获取MindIE容器镜像。
    -   方式一：进入昇腾镜像仓库的[MindIE镜像下载](https://www.hiascend.com/developer/ascendhub/detail/af85b724a7e5469ebd7ea13c3439d48f)页面，下载MindIE镜像。
    -   方式二：参考《MindIE安装指南》中的“安装MindIE \> 方式三：容器安装方式”章节，自行准备镜像。

2.  在节点上查看MindIE镜像。

    ```
    docker images |grep mindie
    ```

    回显示例如下：

    ```
    …
    swr.cn-south-1.myhuaweicloud.com/ascendhub/mindie   2.1.RC2-800I-A2-py311-openeuler24.03-lts   a4708118cd12        6 weeks ago         16GB
    …
    ```

3.  获取Qwen3-1.7B模型权重。

    ```
    # 创建模型权重保存目录
    mkdir -p /data/atlas_dls/public/infer/model_weight
    cd /data/atlas_dls/public/infer/model_weight/
    # 若未安装git-lfs，需要先安装git-lfs。git-lfs是一个Git扩展，专门用于管理大文件和二进制文件
    yum install -y git-lfs 
    # git启用lfs
    git lfs install
    # 权重下载
    git clone https://www.modelscope.cn/Qwen/Qwen3-1.7B.git
    # 修改权重文件权限
    chmod -R 750 Qwen3-1.7B/
    # （可选）如果使用普通用户镜像，权重路径所属应为镜像内默认的1000用户
    chown -R 1000:1000 Qwen3-1.7B/
    ```

    >[!NOTE] 说明 
    >某些模型下载后，还需进行权重量化，详细请参见[ModelZoo-PyTorch](https://gitcode.com/Ascend/ModelZoo-PyTorch/tree/master/MindIE/LLM)中各类模型的README.md。

4.  从MindIE容器内复制配置文件config.json到节点目录。
    1.  在节点上创建目录。

        ```
        mkdir -p /data/atlas_dls/public/infer/script/Qwen3-1.7B
        ```

    2.  启动容器，将目录“/data/atlas\_dls/public/infer/script/Qwen3-1.7B“挂载到容器中。

        ```
        docker run --rm -it \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B:/data/atlas_dls/public/infer/script/Qwen3-1.7B \
        <mindie image:tag>  /bin/bash
        ```

        请用户将<mindie image:tag\>替换为实际镜像名和tag。

    3.  在容器内，将config.json复制到“/data/atlas\_dls/public/infer/script/Qwen3-1.7B“中。

        ```
        cp  $MIES_INSTALL_PATH/conf/config.json /data/atlas_dls/public/infer/script/Qwen3-1.7B/
        ```

        容器内环境变量MIES\_INSTALL\_PATH为MindIE Server的安装路径，默认为“/usr/local/Ascend/mindie/latest/mindie-service“，请用户替换为实际安装路径。

    4.  退出容器。

        ```
        exit
        ```

    5.  在节点的“/data/atlas\_dls/public/infer/script/Qwen3-1.7B“目录中查看config.json文件。

        ```
        ll
        ```

        回显示例如下：

        ```
        …
        -rw-r----- 1 root root 3920 11月  8 11:53 config.json
        …
        ```

5.  修改config.json文件。
    1.  打开config.json文件。

        ```
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/config.json
        ```

    2.  按“i”进入编辑模式，按实际使用情况修改如下参数。参数说明详细请参见《MindIE LLM开发指南》中的“核心概念与配置 \> 配置参数说明（服务化）”章节。

        ```
        {
            …
            "ServerConfig" :
        {
                "ipAddress" : "127.0.0.1",
                "managementIpAddress" : "127.0.0.2",
                "port" : 1025,
                "managementPort" : 1026,
                "metricsPort" : 1027,
                …
                "httpsEnabled" : false,
                …
            },
         
        "BackendConfig" : {
            …
                "npuDeviceIds" : [[0,1]],
                …
                "ModelDeployConfig" :
                {
                    …
                    "truncation" : false,
                    "ModelConfig" : [
                        {
                            …
                            "modelName" : "qwen3",
                            "modelWeightPath" : "/job/model_weight/",
                            "worldSize" : 2,
                            …
                        }
                    ]
                },
                …
            }
        }
        ```

        其中，modelWeightPath为挂载到容器中的模型权重路径。

        >[!NOTICE] 须知 
        >若网络环境不安全，不开启HTTPS通信（即“httpsEnabled”=false时），会存在较高的网络安全风险；若开启HTTPS双向认证，需要提前准备好服务证书、服务器私钥、验签证书等，详情请参见《MindIE Motor开发指南》中的“集群服务部署 \> 单机（非分布式）服务部署  \> 安装部署 \> 使用Deployer部署服务示例 \> 部署Deployer服务端 \> 准备TLS证书”章节，自行准备TLS证书，并修改config.json相关TLS证书配置。

    3.  按“Esc”键，输入**:wq!**，按“Enter”保存并退出编辑。

6.  进入“[mindcluster-deploy](https://gitcode.com/Ascend/mindxdl-deploy)”仓库，根据[mindcluster-deploy开源仓版本说明](../appendix.md#mindcluster-deploy开源仓版本说明)进入版本对应分支，获取“samples/inference/without-k8s/“目录下的启动脚本infer\_start.sh，放在节点目录“/data/atlas\_dls/public/infer/script/Qwen3-1.7B/“下，并对infer\_start.sh脚本进行编辑。

    1.  打开infer\_start.sh脚本。

        ```
        vi /data/atlas_dls/public/infer/script/Qwen3-1.7B/infer_start.sh
        ```

    2.  按“i”进入编辑模式，按实际情况修改脚本中的相关配置。

        ```
        …
        if [[ -z "${MIES_INSTALL_PATH}" ]]; then
            export MIES_INSTALL_PATH=/usr/local/Ascend/mindie/latest/mindie-service # 镜像中MindIE Server安装目录，若安装路径不一致，请用户自行修改
        fi
        …
        mkdir -p /job/script/alllog/
        INFER_LOG_PATH=/job/script/alllog/output_$(date +%Y%m%d_%H%M%S).log # 日志落盘路径
         
        # config.json
        export MIES_CONFIG_JSON_PATH=/job/script/config.json # 推理任务启动配置文件路径，容器启动时挂载进容器
        # （可选）其他用户自定义步骤
        …
        ```

    3.  按“Esc”键，输入**:wq!**，按“Enter”保存并退出编辑。
    4.  增加脚本可执行权限。

        ```
        chmod +x infer_start.sh
        ```

    “/data/atlas\_dls/public/infer/“的目录结构如下：

    ```
    ├── model_weight
    │   └── Qwen3-1.7B
    └── script
        └── Qwen3-1.7B
            ├── config.json
            └── infer_start.sh
    ```

7.  启动容器，拉起MindIE任务。

    -   使用Ascend Docker Runtime挂载芯片和设备

        ```
        docker run -it -d --net=host --shm-size=1g \ 
        --name <container-name> \
        -e ASCEND_VISIBLE_DEVICES=0,1 \
        -v /usr/local/Ascend/driver:/usr/local/Ascend/driver:ro \
        -v /usr/local/sbin:/usr/local/sbin:ro \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B/:/job/script/ \
        -v /data/atlas_dls/public/infer/model_weight/Qwen3-1.7B/:/job/model_weight/ \
        --entrypoint /job/script/infer_start.sh  <mindie image:tag>  <restart_times>
        ```

        >[!NOTE] 说明 
        >-   <container-name\>表示容器名称。
        >-   请用户将<mindie image:tag\>替换为实际镜像名和tag。
        >-   <restart\_times\>作为参数传入infer\_start.sh中，表示服务重启次数，需替换为数字，不填默认为0。超过重启次数会退出容器。
        >-   请用户按需自行修改环境变量ASCEND\_VISIBLE\_DEVICES的值，以挂载不同数量芯片。芯片ID需要与config.json中npuDeviceIds字段包含的芯片ID保持一致。

    -   不使用Ascend Docker Runtime挂载芯片和设备

        ```
        docker run -it -d --net=host --shm-size=1g \
        --name <container-name> \
        --device=/dev/davinci0:rwm \
        --device=/dev/davinci1:rwm \
        --device=/dev/davinci_manager:rwm \
        --device=/dev/devmm_svm:rwm \
        --device=/dev/hisi_hdc:rwm \
         -v /usr/local/sbin/npu-smi:/usr/local/sbin/npu-smi \
        -v /usr/local/Ascend/driver:/usr/local/Ascend/driver:ro \
        -v /usr/local/sbin:/usr/local/sbin:ro \
        -v /data/atlas_dls/public/infer/script/Qwen3-1.7B/:/job/script/ \
        -v /data/atlas_dls/public/infer/model_weight/Qwen3-1.7B/:/job/model_weight/ \
        --entrypoint /job/script/infer_start.sh  <mindie image:tag>  <restart_times>
        ```

    >[!NOTE] 说明 
    >请用户自行增删“--device”参数，以挂载不同数量芯片和设备。芯片ID需要与config.json中npuDeviceIds字段包含的芯片ID保持一致。

8.  查看容器日志。

    ```
    docker logs -f <container-name>
    ```

    若显示如下信息，说明容器启动成功。

    ```
    …
    Daemon start success!
    …
    ```

9.  新建终端窗口，输入以下命令，访问服务。若请求成功返回，表示推理服务部署成功。

    ```
    curl -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    -X POST -d '{
        "model": "<model_name>", 
    "messages": [ 
            {"role": "system", "content": "you are a helpful assistant."},
            { "role": "user", "content": "How many r are in the word \"strawberry\"" } 
        ], 
        "max_tokens": 256, 
        "stream": false,
        "do_sample": true,
        "ignore_eos": true, 
        "temperature": 0.6,
        "top_p": 0.95,
        "top_k": 20,
        "stream": false }' \
    http://<ipAddress>:<port>/v1/chat/completions
    ```

    >[!NOTE] 说明 
    >-   <model\_name\>需替换为config.json中modelName字段的值。
    >-   <ipAddress\>需替换为config.json中ipAddress字段的值。
    >-   <port\>需替换为config.json中port字段的值。

10. 测试服务故障后是否自动重启。
    1.  在节点上构造服务故障。

        ```
        # 查询NPU卡上的进程信息，包含进程号
        npu-smi info
        # 杀进程，构造故障，请将<process_id>替换为进程号
        kill -9 <process_id>
        ```

    2.  查看容器日志。

        ```
        docker logs -f <container-name>
        ```

        若显示如下信息，说明重启成功。

        ```
        Daemon is killing...
        …
        [EntryPoint Script Log]running job failed. exit code: 137
        [EntryPoint Script Log]restart mindie service daemon, cur: 0, max: 1
        …
        Daemon start success!
        ```

11. 停止容器。

    ```
    docker stop <container-name>
    ```


