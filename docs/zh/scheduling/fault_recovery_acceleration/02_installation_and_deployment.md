# 安装部署

## 安装前必读

### 免责声明

本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。

用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

### 约束限制

- MindIO提供TTP、UCE和ARF三种特性，其中MindIO TTP支持在Atlas 800 训练服务器（型号：9000）上使用，MindIO UCE和MindIO ARF不支持该型号设备。
- 众多大模型框架都支持ZeRO（Zero Redundancy Optimizer，零冗余优化器）来减少对显存的使用，当前MindIO TFT仅支持开启ZeRO-1，支持DP（Data Parallelism，数据并行） Size为偶数，同时使用不同的功能对DP Size有不同的限制：
    - MindIO TTP功能
        - 为了保证故障发生后，有完整的优化器状态数据，要求DP Size能被副本数整除。
        - 开启MoE（Mixture of Experts，混合专家结构）前要求稠密层DP Size大于1；开启MoE后要求稠密层和稀疏层DP Size都大于1。
        - 针对分布式优化器，MindIO TFT在ZeRO-1功能的基础上，通过以算代传，在DP Group上重新切分优化器ZeRO-1范围，实现了优化器数据副本。

    - MindIO UCE和MindIO ARF功能
        - 若要实现从当前Step恢复训练，对DP Size限制与MindIO TTP功能一致。
        - 对于显存有限，不做副本的情况，即DP Size = 1，此时若发生UCE或者节点故障，支持在线从周期性Checkpoint中加载模型权重和优化器参数恢复训练，损失当前Step到上次周期性Checkpoint的Step之间的训练成本。

    - 分布式优化器在开启ZeRO特性后，优化器状态数据全局只有一份，无数据冗余。MindIO TFT通过增加优化器状态冗余数据副本，保证故障场景下优化器状态数据的完整性，但同时该方案会导致片上内存使用增加。在原有的模型配置基础上，直接使用MindIO TFT可能会导致模型训练启动过程中出现片上内存OOM（Out Of Memory，内存不足）异常。在此情况下，需要通过扩容增加训练作业的片上内存总量。

        增加副本对应增加的片上内存大小计算公式：增加片上内存总量（GB） = 模型参数量N（B） \* 12 \* 副本数。其中，模型参数量的单位为B（十亿），通过以上公式，计算出需要增加的片上内存，扩容后，再使用MindIO TFT。

- 训练容错框架中有一个Active Controller与两个Backup Controller，为了包括Active Controller在内多张卡发生故障时，能够顺利切换到Backup Controller完成临终保存，需要状态正常的卡的数量大于world\_size的一半。
- MindIO TFT会对优化器状态数据做副本。MindIO UCE或MindIO ARF修复时，寻找有效副本修复故障卡。当训练集群故障较多且通过副本仍无法拼凑出一个完整副本时，系统将从Step在线修复退化为在线加载周期Checkpoint修复。
- MindIO TFT在生成临终Checkpoint数据时，除了考虑一个完整的数据副本，还要校验数据是否一致。如果发生故障后，存在一个OS（Optimizer State，优化器状态）数据Shard长期处于修改状态，或者OS数据不同Shard间训练迭代不一致，都认为是全局数据不一致，无法生成临终Checkpoint数据。
- MindIO TTP不使用MindIO ACP（Async Checkpoint Persistence，异步Checkpoint保存）功能。MindIO TTP完成临终Checkpoint保存后会结束训练进程。为确保在进程退出前，临终Checkpoint已经保存到持久化存储，约束MindIO TTP写数据不使用异步Checkpoint保存方式，而是直接写入到持久化存储。
- MindIO TFT目前不支持级联故障场景。例如：当MindIO TTP正在保存时，如果出现其他故障，就会保存失败。
- MindIO TFT会增加显存占用，详情请参见[表1 原生优化器与开启故障快速恢复特性后优化器参数的理论数值变化](./03_usage_guidance.md#table_tft_03)。
- 默认开启TLS（Transport Layer Security，传输层安全性协议）安全特性，关闭可能导致伪造Controller连接影响训练进程。
- MindIO ARF需要多个节点（≥2），不支持Controller节点发生故障，不支持级联故障；MindIO ARF修复失败后，由MindCluster控制后续流程。
- 日志保存路径默认在运行脚本同级目录下“logs/ttp\_log.log”文件，可在运行脚本里自行配置，默认日志级别为“INFO”，单日志文件大小限制为10MB，写方式为单个追加写，单日志文件达到大小上限后会新建滚动日志文件，滚动日志文件数量限制为5个，多个文件循环写覆盖旧文件。

## 安装前准备

### 组网规划

**图 1**  部署逻辑示意图
![](../../figures/scheduling/部署逻辑示意图ttp.png "部署逻辑示意图")

深度学习平台与训练任务相关的节点有计算节点和存储节点。各类节点主要功能如下：

- 计算节点：实际执行训练、推理任务的节点，MindIO TFT仅部署在计算节点。
- 存储节点：存储平台数据和用户数据，如平台日志、用户上传的数据集、训练脚本、训练输出的模型等。

网络平面划分为：

- 业务面：用于管理集群业务。管理节点和计算节点之间连接。
- 存储面：用于访问存储节点。管理节点和计算节点连接到存储节点。
- 参数面：用于分布式训练时，训练节点之间的参数交换和连接。

    > [!NOTE]说明
    > - 逻辑部署示意图展示深度学习平台的完整示意图，MindIO TFT特性只需要在计算节点上部署一个SDK（Software Development Kit），不涉及存储节点的安装部署。
    > - MindIO TFT功能SDK需要在计算节点相互通信，发送心跳报文，需要使用业务面网络，SDK在所有运行大模型训练的计算节点对等部署，部署时不区分管理节点和计算节点。

### 环境要求

**硬件环境**

安装前，需要检查以下硬件配置，如[表1](#table_tft_01)所示。

**表 1<a id="table_tft_01"></a>**  硬件环境

|类型|配置参考|
|--|--|
|服务器（单机场景）|<ul><li>Atlas 800 训练服务器（型号：9000）：仅支持MindIO TTP功能</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A3 SuperPoD 超节点</li></ul>|
|服务器（集群场景）|计算节点：<ul><li>Atlas 800 训练服务器（型号：9000）：仅支持MindIO TTP功能</li><li>Atlas 800T A2 训练服务器</li><li>Atlas 900 A3 SuperPoD 超节点</li></ul> 存储节点：存储服务器|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s</li><li>带内管理（SSH）：≥1Gbit/s</li><li>业务面：≥10Gbit/s</li><li>存储面：≥25Gbit/s</li><li>参数面：100Gbit/s</li></ul>|

**软件环境**

安装前，需要完成以下环境的安装，如[表2](#table_tft_02)所示。

**表 2<a id="table_tft_02"></a>**  软件环境

|软件|版本|安装位置|获取方式|
|--|--|--|--|
|操作系统|<ul><li>CentOS 7.6</li><li>Ubuntu 18.04</li><li>Ubuntu 20.04</li><li>Ubuntu 22.04</li></ul>|所有节点|-|
|Python|3.7 ~ 3.11|计算节点|用户安装|
|Torch|2.7.1|计算节点|用户安装|
|torch_npu|7.3.0|计算节点|用户安装|
|CANN|8.5.0|计算节点|用户安装|
|驱动与固件|25.5.0|计算节点|用户安装|

### 准备软件包

**下载软件包**

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

**表 1**  软件下载

|组件名称|软件包|获取地址|
|--|--|--|
|MindIO TFT|内存缓存系统软件包|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases)|

**软件数字签名验证**

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请勿使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[http://support.huawei.com/carrier/digitalSignatureAction](http://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

### （可选）启动haveged服务

1. 确认系统是否开启了haveged服务（建议一直开启）。

    ```bash
    systemctl status haveged.service
    ```

    或

    ```bash
    ps -ef | grep "haveged" | grep -v "grep"
    ```

2. 启动haveged服务，并将其设置为随系统启动，确保haveged服务一直开启。

    ```bash
    systemctl start haveged.service
    systemctl enable haveged.service
    ```

3. 查看屏幕输出随机数的速度。

    ```bash
    cat /dev/random | od -x
    ```

    查看当前熵值。

    ```bash
    cat /proc/sys/kernel/random/entropy_avail
    ```

    正常情况下，熵值在未启动haveged时是100多，启动haveged之后会增大到1000多甚至2000。

## 在计算节点安装MindIO TFT SDK

在大模型训练框架使用的Python环境中，安装MindIO TFT SDK，可使能训练任务的故障恢复，从而加速训练恢复。

**操作步骤**

1. 以安装用户 *{MindIO-install-user}* 登录安装节点。

    > [!NOTE]说明
    > 安装用户设置的口令需符合口令复杂度要求（请参见[口令复杂度要求](./06_appendixes.md#口令复杂度要求)）。密码有效期为90天，您可以在“/etc/login.defs”文件中修改有效期的天数，或者通过 **chage** 命令来设置用户的有效期，详情请参见[设置用户有效期](./06_appendixes.md#设置用户有效期)。

2. 将内存缓存系统软件包上传至设备上安装用户有读写权限的路径下。

    > [!NOTE]说明 
    > - 内存缓存系统软件包以获取的实际包名为准。
    > - 如果Python环境是共享目录，则在任一计算节点上传即可，否则所有计算节点都需要上传安装包。

3. 进入软件包上传路径，解压内存缓存系统软件包。

    ```bash
    unzip Ascend-mindxdl-mindio_{version}_linux-{arch}.zip
    ```

    **表 1**  解压后文件

    |文件|说明|
    |--|--|
    |mindio_acp-*{mindio_acp_version}*-py3-none-linux_*{arch}*.whl|MindIO ACP安装包。|
    |mindio_ttp-*{mindio_ttp_version}*-py3-none-linux_*{arch}*.whl|MindIO TFT安装包。|

4. 进入上传路径，执行以下命令，安装MindIO TFT SDK。

    此处以mindio_ttp-_*{mindio_ttp_version}*_-py3-none-linux_*{arch}*.whl为例，请根据实际情况进行选择。

    ```bash
    pip3 install mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl --force-reinstall --no-index
    ```

    - 首次安装MindIO TFT SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_ttp
        Successfully installed mindio_ttp-{mindio_ttp_version}
        ```

    - 非首次安装MindIO TFT SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_ttp-{mindio_ttp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_ttp
          Attempting uninstall: mindio-ttp
            Found existing installation: mindio_ttp {mindio_ttp_version}
            Uninstalling mindio_ttp-{mindio_ttp_version}:
              Successfully uninstalled mindio_ttp-{mindio_ttp_version}
        Successfully installed mindio_ttp-{mindio_ttp_version}
        ```

5. 将软件安装目录内的可执行文件和代码脚本权限更改为550，避免出现非法篡改。

    ```bash
    chmod -R 550 {MindIO TFT SDK安装目录}
    ```

## 卸载MindIO TFT SDK

**操作步骤**

1. 将软件安装目录内的可执行文件和代码脚本权限更改为750。

    ```bash
    chmod -R 750 {MindIO TFT SDK安装目录}
    ```

2. 卸载MindIO TFT SDK。

    ```bash
    pip3 uninstall mindio_ttp
    ```
