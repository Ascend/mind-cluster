# 配置芯片故障<a name="ZH-CN_TOPIC_0000002479226466"></a>

## 概述<a name="ZH-CN_TOPIC_0000002511346521_0101"></a>

Ascend Device Plugin和ClusterD均提供了按照故障频率进行人工隔离芯片的能力，两者功能差异如下：

- Ascend Device Plugin基于节点维度进行故障判定，根据实际发生的故障进行频率计数。
- ClusterD基于任务维度和芯片维度综合进行故障判定：
    - 故障芯片未关联任务，或处理故障时任务已不存在时，该故障不计入ClusterD人工隔离频次。
    - 对关联有效任务且经判定为硬件原因的故障，每张芯片每发生一次，该芯片的故障频率计数加一。
    - 若一个任务下30s内多张卡同时出现同一个故障，则认为非硬件故障导致，不计入故障频率。该规则适用于大多数场景，但在Pod被删除后仍有残留进程等情况下，计数可能存在偏差。
    - 仅新发生的故障才会触发频率上限判断。若将配置的阈值调整至等于或低于当前计数，需等待下一次故障时才触发隔离判断。
    - ClusterD重启后，频率计数信息将丢失，故障频率从零开始重新计数。
    - 解除隔离后，若任务调度不符合预期，请检查节点上是否存在 `huawei.com/scheduler.chip1softsharedev.enable=false` 标签，若存在需将其删除。

Ascend Device Plugin和ClusterD的人工隔离芯片功能，理论上涉及的故障码不需要重复。若不想使用Ascend Device Plugin的隔离功能，请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)章节，将faultCustomization.json文件中的人工隔离芯片相关的配置删除；若不想使用ClusterD的隔离功能，请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)章节，将人工隔离芯片功能开关关闭。

若Ascend Device Plugin和ClusterD对同一张芯片都进行了人工隔离，需要各自解除隔离。Ascend Device Plugin解除隔离的方法请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)中"手动恢复强制隔离的芯片"步骤；ClusterD解除隔离的方法请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)中"手动恢复人工隔离的芯片"步骤。

## Ascend Device Plugin<a name="ZH-CN_TOPIC_0000002511346521_02"></a>

### 配置文件说明<a name="ZH-CN_TOPIC_0000002511346521"></a>

故障检测特性针对**芯片故障**，支持按故障级别、故障频率和故障时长的配置进行处理。

- 针对芯片故障的**不同级别**进行分级处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCode.json**中故障码配置的故障级别，对故障进行相应处理。
- 针对芯片故障的**故障频率及时长**进行处理时，Ascend Device Plugin组件会获取到当前故障的故障码，根据**faultCustomization.json**中故障配置的故障频率和时长，对故障进行相应处理。

faultCode.json、faultCustomization.json为系统配置文件，若用户无特殊需求，请勿随意修改。若Ascend Device Plugin默认的频率故障配置中有由软件原因可以触发的故障，用户可自行将该故障码删除。（软件原因会导致一个任务下短时间内反复大量出现某个故障，导致Ascend Device Plugin侧感知到该故障达到了故障频率，将大量设备置为人工隔离状态。）

若用户需要修改故障码对应的故障级别，可以通过由faultCode.json和faultCustomization.json创建的**mindx-dl-fault-config**文件实现。

>[!NOTE]
>
>- 每个故障对应的故障码请参见[芯片故障码参考文档](../../../07_references/05_appendix.md#芯片故障码参考文档)章节。
>- 芯片故障支持配置的故障级别参见[故障级别](#zh-cn_topic_0000002171521445_section5245155017242)。
>- 芯片故障支持配置的故障频率和时长参见[故障频率及时长](#zh-cn_topic_0000002171521445_section115842029104220)。

**faultCode.json中的故障级别<a name="zh-cn_topic_0000002171521445_section5245155017242"></a>**

故障检测特性针对芯片故障的不同级别进行分级处理。若用户需要修改故障码的故障级别，操作指导请参见[（可选）配置芯片故障级别](#可选配置芯片故障级别)。

Ascend Device Plugin从驱动获取到芯片故障码后，将根据故障码对设备及业务的影响将故障划分几个级别，详细说明请参见[表1](../../../06_api/02_ascend_device_plugin.md#自定义芯片故障)。

>[!NOTE]
>
>- 复位芯片前需要停止训练进程，否则复位将失败。
>- 若Ascend Device Plugin通过订阅的方式收到了无法识别的故障码（未保存在faultCode.json中），默认按照订阅接口给的处理意见进行故障处理。若订阅接口收到的故障等级为“提示”或“次要”，则按照NotHandleFault级别处理；若故障等级为其他等级，则按照SeparateNPU级别处理。

**故障频率及时长<a name="zh-cn_topic_0000002171521445_section115842029104220"></a>**

故障检测特性针对芯片故障的故障频率及时长进行处理。某些硬件类故障可能在一次训练任务中反复出现，导致训练任务中断反复进行重调度。集群调度组件针对这些故障对应的故障码，提供了提升故障级别的初始化配置文件faultCustomization.json。

- faultCustomization.json文件提供的初始化配置和故障类型关系如下[初始化配置和故障类型](#zh-cn_topic_0000002171521445_section13684172919539)。
- faultCustomization.json文件的默认配置（默认值）请参见[表2](../../../06_api/02_ascend_device_plugin.md#自定义芯片故障)。
- 若用户需要修改故障频率及时长配置，操作指导请参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)。

**初始化配置和故障类型<a name="zh-cn_topic_0000002171521445_section13684172919539"></a>**

当前faultCustomization.json文件中仅提供对可识别的硬件类故障进行提升故障级别的初始化配置。

单个节点在24小时内发生3次以下故障，则将芯片故障级别提升至需要人工干预的故障级别ManuallySeparateNPU，详细说明请参见[faultCustomization.json参数说明](#zh-cn_topic_0000002171521445_section33036167576)。

下面将以故障名称HBMC Ca Parity错误，对应故障码80E18005为例，将当前的故障级别提升至ManuallySeparateNPU（需要人工干预的故障级别），示例如下。

```json
  "FaultFrequency": [
    {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    },
    {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
    }
  ],
```

>[!NOTE]
>
>- 故障的处理策略为ManuallySeparateNPU时，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)中“手动恢复强制隔离的芯片”步骤进行处理。
>- 除可以识别的硬件故障外，faultCustomization.json文件中还包含以下几类故障。
>     - 无需处理的故障：该类故障出现不影响训练任务及设备，不提供提升故障级别的初始化配置。
>     - 无法识别出是硬件还是软件类故障：该类故障无法准确识别是硬件还是软件故障，且会影响训练任务。该类故障不提供提升故障级别的初始化配置，建议用户根据实际情况手动配置任务支持的断点续训最大次数和达到最大次数后故障的处理策略，可以参见[（可选）配置芯片故障频率及时长](#可选配置芯片故障频率及时长)进行配置。
>     - 软件配置类故障：该类故障为软件配置类问题，正常情况下不会出现。该类故障不提供提升故障级别的初始化配置，建议用户检查软件版本是否配套。

**faultCustomization.json参数说明<a name="zh-cn_topic_0000002171521445_section33036167576"></a>**

用户不手动修改faultCustomization.json文件时，Ascend Device Plugin按照faultCustomization.json的默认配置（默认值）进行故障处理。faultCustomization.json文件参数说明请参见[表2](../../../06_api/02_ascend_device_plugin.md#自定义芯片故障)。

### （可选）配置芯片故障级别<a name="ZH-CN_TOPIC_0000002479226532"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#ZH-CN_TOPIC_0000002511346521)。

如果用户想要自定义故障级别或者优雅容错相关配置，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

- 如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCode.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。
- 若想要使用faultCode.json或faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCode.json、SwitchFaultCode.json或faultCustomization.json文件。
- 如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

**使用faultCode.json配置故障级别<a name="zh-cn_topic_0000001951258609_section112139052513"></a>**

以故障名称dmp\_daemon节点状态检测异常，对应故障码80E21007为例。将当前故障的处理策略NotHandleFaultCodes（无需处理）修改为RestartNPUCodes（隔离芯片，进行任务重调度）的操作示例如下。

1. 登录环境，进入Ascend Device Plugin解压目录。
2. 执行以下命令，创建动态配置故障码所需ConfigMap文件（mindx-dl-fault-config）。

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json
    ```

    回显示例如下。

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **表 1**  参数说明

    <a name="zh-cn_topic_0000001951258609_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000001951258609_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000001951258609_p16631548171910"><a name="zh-cn_topic_0000001951258609_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000001951258609_p1663144816197"><a name="zh-cn_topic_0000001951258609_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000001951258609_p775918210209"><a name="zh-cn_topic_0000001951258609_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001951258609_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1863164816197"><a name="zh-cn_topic_0000001951258609_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1063194861910"><a name="zh-cn_topic_0000001951258609_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p157595292015"><a name="zh-cn_topic_0000001951258609_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000001951258609_ph126311642183015"><a name="zh-cn_topic_0000001951258609_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1963194819195"><a name="zh-cn_topic_0000001951258609_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p76316488192"><a name="zh-cn_topic_0000001951258609_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p276092142019"><a name="zh-cn_topic_0000001951258609_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p964144891914"><a name="zh-cn_topic_0000001951258609_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p1164748191916"><a name="zh-cn_topic_0000001951258609_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p876012211206"><a name="zh-cn_topic_0000001951258609_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001951258609_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000001951258609_p1964748141915"><a name="zh-cn_topic_0000001951258609_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000001951258609_p10641648191915"><a name="zh-cn_topic_0000001951258609_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000001951258609_p147602211206"><a name="zh-cn_topic_0000001951258609_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

3. 执行以下命令，编辑mindx-dl-fault-config文件。

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

4. 在mindx-dl-fault-config文件中，找到故障码80E21007。

    ```json
    "NotHandleFaultCodes":[

    "80E21007","80E38003","80F78006","80C98006","80CB8006","81318006","80A18006","80A18005","80FB8000","8C1F8609",
    ...
      ],
    ...
    ```

    >[!NOTE]
    >同一故障码配置在多个故障级别中，会显示设置成功，但默认按照高等级故障处理。

5. 将故障码80E21007从NotHandleFaultCodes中删除，并添加到RestartNPUCodes中。

    ```json
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

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1. 执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. 通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        若日志出现“load fault code from configmap success”，表示手动配置故障码操作成功。

### （可选）配置芯片故障频率及时长<a name="ZH-CN_TOPIC_0000002511426473"></a>

在制作Ascend Device Plugin镜像时，会将faultCode.json和faultCustomization.json配置文件内置在镜像中，启动Ascend Device Plugin时会读取这两个文件的默认配置，作为当前故障处理依据。faultCode.json和faultCustomization.json的说明请参见[配置文件说明](#ZH-CN_TOPIC_0000002511346521)。

如果用户想要自定义芯片故障频率及时长，可以在集群中创建ConfigMap文件（mindx-dl-fault-config）。

- 如果Ascend Device Plugin启动时，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin会优先按照已存在的mindx-dl-fault-config中配置的内容，作为当前故障处理依据。
- 如果重新安装Ascend Device Plugin后，集群中已经存在mindx-dl-fault-config，Ascend Device Plugin的默认faultCustomization.json将不会生效，使用集群中已经存在的mindx-dl-fault-config。若想要使用faultCustomization.json的默认配置，可以删除mindx-dl-fault-config，使Ascend Device Plugin读取默认faultCustomization.json文件。
- 如果ConfigMap文件内容存在格式错误等问题，Ascend Device Plugin会默认读取镜像中内置的ConfigMap文件的内容，作为当前故障处理依据。

>[!CAUTION]
>修改故障频率为高危操作，如果修改不当，会导致芯片被误隔离。例如，由于任务发生错误导致的软件故障，会短时间内反复大量出现，使Ascend Device Plugin侧感知到该故障达到了故障频率，将大量芯片置为人工隔离状态，导致大量节点无法调度。

**操作步骤<a name="section141902103110"></a>**

以故障码80CB8002为例，如果某张芯片反复发生80CB8002故障，导致训练业务反复重调度，可以手动配置24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

1. 登录环境，进入Ascend Device Plugin解压目录。
2. 执行以下命令，查询是否已经基于faultCode.json文件创建了mindx-dl-fault-config。

    ```shell
    kubectl describe cm -n kube-system mindx-dl-fault-config
    ```

    - 如果mindx-dl-fault-config已经存在，且存在faultCustomization.json的相关字段，执行[步骤4](#zh-cn_topic_0000002136360238_li38432520129)编辑该文件。
    - 如果mindx-dl-fault-config已经存在，但是不存在faultCustomization.json的相关字段，需要先保存mindx-dl-fault-config内容，再删除mindx-dl-fault-config文件后，执行[步骤3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。
    - 如果不存在mindx-dl-fault-config，执行[步骤3](#zh-cn_topic_0000002136360238_li1946014413123)创建该文件。

3. <a name="zh-cn_topic_0000002136360238_li1946014413123"></a>执行以下命令，创建配置芯片故障频率及时长所需ConfigMap文件（mindx-dl-fault-config）。

    ```shell
    kubectl create cm mindx-dl-fault-config -n kube-system --from-literal="PollInterval=300" --from-file=./faultCode.json --from-file=./faultCustomization.json
    ```

    回显示例如下。

    ```ColdFusion
    configmap/mindx-dl-fault-config created
    ```

    **表 2**  参数说明

    <a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_table16314861918"></a>
    <table><thead align="left"><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763648161914"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.2.4.1.1"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p16631548171910"></a>参数名</p>
    </th>
    <th class="cellrowborder" valign="top" width="11.701170117011701%" id="mcps1.2.4.1.2"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1663144816197"></a>是否必选</p>
    </th>
    <th class="cellrowborder" valign="top" width="54.96549654965496%" id="mcps1.2.4.1.3"><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p775918210209"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row36354871915"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1863164816197"></a>mindx-dl-fault-config</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1063194861910"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p157595292015"></a>动态配置故障码所需的<span id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_ph126311642183015"></a>ConfigMap</span>文件名称，不能修改该文件名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row763184812192"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1963194819195"></a>kube-system</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p76316488192"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p276092142019"></a>mindx-dl-fault-config所在命名空间，不能修改该命名空间名称。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row86314881910"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p964144891914"></a>PollInterval</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1164748191916"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p876012211206"></a>不指定该参数则默认取值为300s。用于指定查询mindx-dl-fault-config文件是否更新的周期时间，单位为秒，取值范围为30~3600。PollInterval的修改将在下一个周期生效。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_row176474851911"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p1964748141915"></a>faultCode.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p10641648191915"></a>是</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001721780141_p147602211206"></a>用于保存故障码，必须与faultCode.json文件名称保持一致。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000002136360238_row9289716194614"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.2.4.1.1 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p172981520305"></a>faultCustomization.json</p>
    </td>
    <td class="cellrowborder" valign="top" width="11.701170117011701%" headers="mcps1.2.4.1.2 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p122981952113016"></a>否</p>
    </td>
    <td class="cellrowborder" valign="top" width="54.96549654965496%" headers="mcps1.2.4.1.3 "><p id="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"><a name="zh-cn_topic_0000002136360238_zh-cn_topic_0000001762151497_p7298145218303"></a>用于自定义优雅容错时间、故障频率、故障持续时间（仅支持参数面网络故障）等配置，不指定该参数则没有故障频率配置，其余配置使用默认值进行处理。必须与faultCustomization.json文件名称保持一致。</p>
    </td>
    </tr>
    </tbody>
    </table>

4. <a name="zh-cn_topic_0000002136360238_li38432520129"></a>执行以下命令，编辑mindx-dl-fault-config文件。

    ```shell
    kubectl edit cm -n kube-system mindx-dl-fault-config
    ```

    根据实际情况，修改芯片的故障频率和时长。

    ```yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
    PollInterval: "300"
    # 修改芯片故障的故障级别
    faultCode.json: |
    {
    "NotHandleFaultCodes":[
    ...
    }
    # 修改芯片故障的故障频率和时长
    faultCustomization.json: |
    {
     "GraceTolerance": {
     "WaitProcessReadCMTime": 30,
     "WaitDeviceResetTime": 150,
     "WaitFaultSelfHealingTime": 15
    },
    "FaultFrequency": [
     {
      "EventId": [
        "80C98000","80B78000","80B58000","80A18008","80A38008","80A58008","80B98000","80B98008","80BB8000",
        "80BB8008","80BD8000","80BD8008","80C78008","80C98008","80CB8008","80CD8008","80CF8008","80D98008",
        "80DF8008","80DE1801","80E01801","80E18008","80E38008","80E39200","80E3A202","80E3A203","80E78000",
        "80E78008","80F18000","80F18008","80F38008","80F78008","81318008","81338008","813B8008","81478008",
        "81578008","815F8008","81938008","81958008","81978008"
      ],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["80E18005"],
      "TimeWindow": 86400,
      "Times": 3,
      "FaultHandling": "ManuallySeparateNPU"
     },
     {
      "EventId": ["81078603"],
      "TimeWindow": 86400,
      "Times": 5,
      "FaultHandling": "ManuallySeparateNPU",
      "ReleaseTimeWindow": 172800
     }
    ],
    "FaultDuration": [
     {
      "EventId": ["81078603"],
      "FaultTimeout": 20,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     },
     {
      "EventId": ["81B18603"],
      "FaultTimeout": 5,
      "RecoverTimeout": 60,
      "FaultHandling": "PreSeparateNPU"
     }
    ]
   }
    kind: ConfigMap
    metadata:
    creationTimestamp: "2024-06-20T10:12:07Z"
    name: mindx-dl-fault-config
    namespace: kube-system
    resourceVersion: "52893696"
    selfLink: /api/v1/namespaces/kube-system/configmaps/mindx-dl-fault-config
    uid: bba9e17f-41dd-43b3-848e-3d29cb8c595a
    ```

5. 在mindx-dl-fault-config文件中，在FaultFrequency字段下新增以下代码，设置80CB8002故障在24小时内任务支持的断点续训最大次数为2，达到最大次数后故障的处理策略为ManuallySeparateNPU。

    ```json
    {
      "EventId": ["80CB8002"],
      "TimeWindow": 86400,
      "Times": 2,
      "FaultHandling": "ManuallySeparateNPU"
    }
    ```

6. 修改完成后，按“Esc”键，输入:wq!保存并退出。
7. 等mindx-dl-fault-config文件更新生效（PollInterval取值，不指定则为300s）后，查看操作是否成功。
    1. 执行以下命令，查询Ascend Device Plugin组件日志名称。

        ```shell
        kubectl get pods -A | grep ascend-device-plugin
        ```

        回显示例如下：

        ```ColdFusion
        kube-system      ascend-device-plugin-daemonset-910-jmlf5   1/1     Running   0              6h34m
        ```

    2. 通过查询到的组件日志名称，查询Ascend Device Plugin的组件日志信息。

        ```shell
        kubectl logs -n kube-system ascend-device-plugin-daemonset-910-jmlf5
        ```

        >[!NOTE]
        >- 若日志出现“load fault customization from configmap complete”，表示手动配置故障频率操作成功。
        >- 若日志出现“modify  _xxx_  success”，表示ConfigMap中faultCustomization.json里的<i>xxx</i>参数设置成功。
        >- 若日志出现“insert fault frequency success”，表示记录了一次频率故障发生时间，在频率窗口内，该卡的该故障记录次数达到频率故障触发次数以后，就会上报频率故障对应的故障级别。

8. （可选）手动恢复强制隔离的芯片。故障的处理策略为ManuallySeparateNPU时，故障恢复后该芯片也处于隔离状态，在未达到释放条件时若需要手动恢复强制隔离的芯片。
    1. 执行以下命令，查找该节点的Ascend Device Plugin上报的device-info-cm。

        ```shell
        kubectl get cm -n kube-system | grep deviceinfo | grep {nodeName}
        ```

    2. 执行以下命令，编辑该device-info-cm。

        ```shell
        kubectl edit cm -n kube-system {configMapName}
        ```

    3. 将data下面的ManuallySeparateNPU后面已恢复健康的芯片名称删除。

        ```Yaml
        apiVersion: v1
        kind: ConfigMap
        data:
          DeviceInfoCfg: '{"DeviceInfo":{"DeviceList":{"huawei.com/Ascend910":"Ascend910-1,Ascend910-2,Ascend910-3,Ascend910-4,Ascend910-5,Ascend910-6,Ascend910-7","huawei.com/Ascend910-Fault":"[]","huawei.com/Ascend910-NetworkUnhealthy":"","huawei.com/Ascend910-Unhealthy":""},"UpdateTime":1718702470},"CheckCode":"4f00cf1d220da26a8fdbeb5ba163a751d4b264c48b81d22149257e272ae3b413"}'
          ManuallySeparateNPU: Ascend910-0
        ```

        >[!NOTE]
        >删除ManuallySeparateNPU字段后所有芯片名称，并将取值设置为空“”。

    4. 修改完成后，按“Esc”键，输入:wq!保存并退出。
    5. 等待1个上报周期（若设备信息有变化，那么在健康状态检查周期内就会上报，如果设备信息没有变化，那么上报周期固定为5分钟）后，执行以下命令，查看device-info-cm中ManuallySeparateNPU是否存在刚才删除的芯片名称。若不存在，则芯片恢复健康成功，可继续正常使用该芯片。

        ```shell
        kubectl describe cm -n kube-system {configMapName}
        ```

## ClusterD<a name="ZH-CN_TOPIC_0000002511346521_03"></a>

### 配置说明<a name="ZH-CN_TOPIC_0000002511346521_04"></a>

故障检测特性针对芯片故障，支持按故障频率的配置进行处理。

针对芯片故障的不同级别进行分级处理时，ClusterD组件会获取到当前故障的故障码和故障级别，对于除了NotHandleFault和SubHealthFault级别之外的故障，根据ConfigMap（clusterd-config-cm）中配置的故障频率，将芯片状态置为人工隔离。该ConfigMap的参数说明请参见[表1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md)。

>[!NOTE]
>
>- ConfigMap（clusterd-config-cm）为系统配置，若用户无特殊需求，请勿随意修改。若用户需要修改人工隔离芯片检测开关及故障频率、解除隔离时间等，可以通过修改该ConfigMap实现，修改方法请参见[（可选）配置芯片故障频率](#可选配置芯片故障频率)。
>- 不支持配置故障码检测范围，ClusterD会基于Ascend Device Plugin上报的故障级别进行判断。对于除了NotHandleFault和SubHealthFault级别之外的故障，都会进入人工隔离芯片检测流程。

### （可选）配置芯片故障频率<a name="ZH-CN_TOPIC_0000002511426473_01"></a>

在安装ClusterD时，会自动创建ConfigMap（clusterd-config-cm），作为当前人工隔离芯片的检测依据。该ConfigMap的参数说明请参见[表1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md)。

如果用户想要自定义芯片故障频率，可以通过修改该ConfigMap实现。如果修改后的ConfigMap内容存在格式错误等问题，ClusterD会保留上一次读取成功的配置作为当前人工隔离芯片的检测依据。若ClusterD启动时，读取到的ConfigMap内容错误，则人工隔离芯片检测机制会默认关闭，直到格式和内容正确。

**操作步骤<a name="section14190101"></a>**

以人工隔离芯片的阈值由默认的24小时内出现3次调整为24小时内出现5次为例。

1. 登录环境，执行以下命令，查询当前配置。

    ```shell
    kubectl describe cm -n cluster-system clusterd-config-cm
    ```

    - 如果存在clusterd-config-cm，则执行[步骤3](#li01010203)进行编辑。
    - 如果不存在clusterd-config-cm，则执行[步骤2](#li010102)进行创建。

    >[!NOTE]
    >正常情况下存在clusterd-config-cm。若不存在，需确认ClusterD的安装过程是否存在错误。

2. <a name="li010102"></a>创建人工隔离芯片检测所需的ConfigMap（clusterd-config-cm）。

    将以下内容保存为文件cm.yaml：

    ```Yaml
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: clusterd-config-cm
          namespace: cluster-system
        data:
          manually_separate_policy.conf: |
            enabled: true
            separate:
              fault_window_hours: 24
              fault_threshold: 3
            release:
              fault_free_hours: 48

    ```

    执行以下命令：

    ```shell
    kubectl apply -f cm.yaml
    ```

    回显示例如下，说明创建成功。

    ```ColdFusion
    configmap/clusterd-config-cm created
    ```

3. <a name="li01010203"></a>执行以下命令，编辑clusterd-config-cm。

    ```shell
    kubectl edit cm -n cluster-system clusterd-config-cm
    ```

    根据实际情况，修改人工隔离芯片的故障频率。参数说明请参见[表1](../../../05_developer_guide/00_installation_deployment/00_manual_installation/06_clusterd.md)。

    ```Yaml
    # Please edit the object below. Lines beginning with a '#' will be ignored,
    # and an empty file will abort the edit. If an error occurs while saving this file will be
    # reopened with the relevant failures.
    #
    apiVersion: v1
    data:
      manually_separate_policy.conf: |
        # 修改人工隔离芯片检测开关
        enabled: true
        separate:
          # 修改人工隔离芯片的故障频率
          fault_window_hours: 24
          fault_threshold: 5   # 由3修改为5
        release:
          # 修改解除隔离时间
          fault_free_hours: 48
    kind: ConfigMap
    metadata:
      annotations:
        kubectl.kubernetes.io/last-applied-configuration: |
          {"apiVersion":"v1","data":{"manually_separate_policy.conf":"enabled: true\nseparate:\n  fault_window_hours: 24\n  fault_threshold: 3\nrelease:\n  fault_free_hours: 48\n"},"kind":"ConfigMap","metadata":{"annotations":{},"name":"clusterd-config-cm","namespace":"cluster-system"}}
      creationTimestamp: "2026-02-24T11:25:19Z"
      name: clusterd-config-cm
      namespace: cluster-system
      resourceVersion: "3344125"
      selfLink: /api/v1/namespaces/cluster-system/configmaps/clusterd-config-cm
      uid: 68210bfc-f742-4765-a497-b61e9cc6b1a6
    ```

4. 修改完成后，按“Esc”键，输入:wq!保存并退出。
5. 等clusterd-config-cm更新生效（ClusterD的检测周期为300s）后，查看操作是否成功。
    1. 执行以下命令，查询ClusterD组件日志名称。

        ```shell
        kubectl get pods -A | grep clusterd
        ```

        回显示例如下：

        ```ColdFusion
        mindx-dl      clusterd-559bf4bd6-z9hv4   1/1     Running   0             4m23s
        ```

    2. 通过查询到的组件日志名称，查询ClusterD的组件日志信息。

        ```shell
        kubectl logs -f -n mindx-dl clusterd-559bf4bd6-z9hv4
        ```

        >[!NOTE]
        >- 若日志出现“load manually separate policy config success”，表示手动修改人工隔离芯片的故障频率操作成功。
        >- 若日志出现“node: xx, dev: xx, code: xx is not found in manual fault cache, add”，表示该故障触发人工隔离。
        >- 若日志出现“node: xx, dev: xx, code: xx is found in manual fault cache, update last separate time”，表示已经触发人工隔离芯片的故障，再一次达到了人工隔离的故障频率，会刷新clusterd-manual-info-cm中的LastSeparateTime。clusterd-manual-info-cm的说明请参见[clusterd-manual-info-cm](../../../06_api/04_clusterd/00_cluster_resources.md#clusterd-manual-info-cm)。

6. （可选）手动恢复人工隔离的芯片。故障的处理策略为ManuallySeparateNPU时，故障恢复后该芯片也处于隔离状态，可以手动恢复人工隔离的芯片。

    1. 执行以下命令，编辑ConfigMap clusterd-manual-info-cm。

        ```shell
        kubectl edit cm -n cluster-system clusterd-manual-info-cm
        ```

    2. 将Data下面的Total字段后面需要解除人工隔离的芯片名称删除。例如：Ascend910-2。

        ```json
        Name:         clusterd-manual-info-cm
        Namespace:    cluster-system
        Labels:       <none>
        Annotations:  <none>

        Data
        ====
        localhost.localdomain:
        ----
        {"Total":["Ascend910-0","Ascend910-2","Ascend910-3"],"Detail":{"Ascend910-0":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-2":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}],"Ascend910-3":[{"FaultCode":"8C084E00","FaultLevel":"ManuallySeparateNPU","LastSeparateTime":1770811685650}]}}

        Events:  <none>
        ```

    3. 修改完成后，按“Esc”键，输入:wq!保存并退出。
    4. 等待15s后，执行以下命令，查看clusterd-manual-info-cm中Ascend910-2是否还存在于Total和Detail字段中。同时，需要查看该芯片的ManuallySeparateNPU故障是否存在于cluster-info-device-\${m}中。若不存在，则芯片解除人工隔离成功，可继续正常使用该芯片。

        ```shell
        kubectl describe cm -n cluster-system clusterd-manual-info-cm
        ```

        >[!NOTE]
        >- 仅支持删除Total字段中的芯片，不支持手动添加。其他内容不支持修改。
        >- 手动恢复人工隔离的芯片后，该芯片的故障计数会清零，再次达到频率时才会再次触发人工隔离。
        >- 若需要删除节点上所有的人工隔离芯片，则需删除Total字段后面的所有芯片名称，并将取值设置为空[]。如果想一次性解除所有的人工隔离芯片，可以直接将clusterd-manual-info-cm删除。
        >- ClusterD启动后15s内，暂时先不要修改clusterd-manual-info-cm，以免发生数据错误。
