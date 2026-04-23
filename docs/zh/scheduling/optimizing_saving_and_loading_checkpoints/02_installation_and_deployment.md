# 安装部署

## 安装前必读

### 免责声明

本文档可能包含第三方信息、产品、服务、软件、组件、数据或内容（统称“第三方内容”）。华为不控制且不对第三方内容承担任何责任，包括但不限于准确性、兼容性、可靠性、可用性、合法性、适当性、性能、不侵权、更新状态等，除非本文档另有明确说明。在本文档中提及或引用任何第三方内容不代表华为对第三方内容的认可或保证。

用户若需要第三方许可，须通过合法途径获取第三方许可，除非本文档另有明确说明。

### 约束限制

- 训练[故障快速恢复](../fault_recovery_acceleration/01_product_description.md)框架正在向MindIO ACP保存Checkpoint时，如果遇到Checkpoint保存失败，当前正在保存的Checkpoint不能作为训练恢复点，训练框架需要从上一次完整的Checkpoint点进行恢复。
- 在训练过程中发生MindIO ACP故障，已经下发的业务，MindIO ACP SDK会重试3次连接，3次都失败则对接原生存储方式，重试最长等待60s；在训练开始前发生MindIO ACP故障，MindIO ACP SDK则会跳过对接MindIO ACP，Checkpoint的数据直接对接原生数据存储方式。
- 本特性不配套MindSpore 2.7.0之前的版本，功能无法使用。

## 安装前准备

### 组网规划

**图 1**  部署逻辑示意图
![](../../figures/scheduling/部署逻辑示意图acp.png "部署逻辑示意图")

深度学习平台与训练任务相关的节点有计算节点和存储节点。各类节点主要功能如下：

- 计算节点：实际执行训练、推理任务的节点，MindIO ACP仅部署在计算节点。
- 存储节点：存储平台数据和用户数据，如平台日志、用户上传的数据集、训练脚本、训练输出的模型等。

网络平面划分为：

- 业务面：用于管理集群业务。管理节点和计算节点之间连接。
- 存储面：用于访问存储节点。管理节点和计算节点连接到存储节点。
- 参数面：用于分布式训练时，训练节点之间的参数交换和连接。

> [!NOTE]说明
>
> - 逻辑部署示意图展示深度学习平台的完整示意图，MindIO ACP作为计算节点上部署的一个组件，不涉及管理节点和存储节点的安装部署。
> - MindIO ACP是单节点内存缓存系统，训练Checkpoint数据通过共享内存方式访问MindIO ACP，不涉及网络平面划分。

### 环境要求

**硬件环境**

安装前，需要检查以下硬件配置，如[表1](#table_acp_01)所示。

**表 1 <a id="table_acp_01"></a>**  硬件环境

|类型|配置参考|
|--|--|
|服务器（单机场景）|Atlas 800 训练服务器（型号：9000）|
|服务器（集群场景）|<ul><li>计算节点：Atlas 800 训练服务器（型号：9000）</li><li>存储节点：存储服务器</li></ul>|
|内存|<ul><li>推荐配置：≥64GB</li><li>最低配置：≥32GB</li></ul>|
|磁盘空间|≥1TB <br> 磁盘空间规划请参见[表3](#table_acp_03)|
|网络|<ul><li>带外管理（BMC）：≥1Gbit/s</li><li>带内管理（SSH）：≥1Gbit/s</li><li>业务面：≥10Gbit/s</li><li>存储面：≥25Gbit/s</li><li>参数面：100Gbit/s</li></ul>|

**软件环境**

安装前，需要完成以下环境的安装，如[表2](#table_acp_02)所示。

**表 2 <a id="table_acp_02"></a>**  软件环境

|软件|版本|安装位置|获取方式|
|--|--|--|--|
|操作系统|<ul><li>CentOS 7.6 Arm</li><li>CentOS 7.6 x86</li><li>openEuler 20.03 Arm</li><li>openEuler 20.03 x86</li><li>openEuler 22.03 Arm</li><li>openEuler 22.03 x86</li><li>Ubuntu 20.04 Arm</li><li>Ubuntu 20.04 x86</li><li>Ubuntu 18.04.5 Arm</li><li>Ubuntu 18.04.5 x86</li><li>Ubuntu 18.04.1 Arm</li><li>Ubuntu 18.04.1 x86</li><li>Kylin V10 SP2 Arm</li><li>Kylin V10 SP2 x86</li><li>UOS20 1020e Arm</li></ul>|所有节点|-|
|Python|3.7或更高版本|计算节点|用户安装|
|Torch|2.7.1|计算节点|用户安装|
|MindSpore|2.7.0或更高版本|计算节点|用户安装|

**操作系统磁盘分区**

操作系统磁盘分区推荐如[表3](#table_acp_03)所示。

**表 3 <a id="table_acp_03"></a>**  磁盘分区

|分区|说明|大小|bootable flag|
|--|--|--|--|
|/boot|启动分区|500MB|on|
|/var|软件运行所产生的数据存放分区，如日志、缓存等|>300GB|off|
|/|主分区|>300GB|off|

### 准备软件包

**下载软件包**

下载本软件即表示您同意[华为企业业务最终用户许可协议（EULA）](https://e.huawei.com/cn/about/eula)的条款和条件。

**表 1**  软件下载

|组件名称|软件包|获取地址|
|--|--|--|
|MindIO ACP|内存缓存系统软件包|[获取链接](https://gitcode.com/Ascend/mind-cluster/releases)|

**软件数字签名验证**

为了防止软件包在传递过程或存储期间被恶意篡改，下载软件包时需下载对应的数字签名文件用于完整性验证。

在软件包下载之后，请参见《[OpenPGP签名验证指南](https://support.huawei.com/enterprise/zh/doc/EDOC1100209376)》，对从Support网站下载的软件包进行PGP数字签名校验。如果校验失败，请勿使用该软件包，先联系华为技术支持工程师解决。

使用软件包安装/升级之前，也需要按上述过程先验证软件包的数字签名，确保软件包未被篡改。

运营商客户请访问：[http://support.huawei.com/carrier/digitalSignatureAction](http://support.huawei.com/carrier/digitalSignatureAction)

企业客户请访问：[https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054](https://support.huawei.com/enterprise/zh/tool/pgp-verify-TL1000000054)

## 在计算节点安装MindIO ACP SDK

通过使用MindIO ACP SDK对接Torch和MindSpore，加速Torch和MindSpore训练Checkpoint save和load操作。

**操作步骤**

1. 以安装用户 *{MindIO-install-user}* 登录安装节点。

    >[!NOTE]说明
    >安装用户设置的口令需符合口令复杂度要求（请参见[口令复杂度要求](./07_appendixes.md#口令复杂度要求)）。密码有效期为90天，您可以在“/etc/login.defs“文件中修改有效期的天数，或者通过 **chage** 命令来设置用户的有效期，详情请参见[设置用户有效期](./07_appendixes.md#设置用户有效期)。

2. 将内存缓存系统软件包上传至设备中安装用户有读写权限的路径下。

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

4. 进入上传路径，安装MindIO ACP SDK。

    ```bash
    pip3 install mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl --force-reinstall
    ```

    - 首次安装MindIO ACP SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl
        Installing collected packages: mindio_acp
        Successfully installed mindio_acp-{version}
        ```

    - 非首次安装MindIO ACP SDK回显如下，表示安装成功。

        ```bash
        Processing ./mindio_acp-{mindio_acp_version}-py3-none-linux_{arch}.whl
         Installing collected packages: mindio_acp
           Attempting uninstall: mindio_acp
             Found existing installation: mindio_acp {mindio_acp_version}
             Uninstalling mindio_acp-{mindio_acp_version}:
               Successfully uninstalled mindio_acp-{mindio_acp_version}
         Successfully installed mindio_acp-{mindio_acp_version}
        ```

5. 将软件安装目录内的可执行文件和代码脚本权限更改为550，避免出现非法篡改。

    ```bash
    chmod -R 550 {MindIO ACP SDK安装目录}
    ```

## 卸载MindIO ACP SDK

**操作步骤**

1. 将软件安装目录内的可执行文件和代码脚本权限更改为750。

    ```bash
    chmod -R 750 {MindIO ACP SDK安装目录}
    ```

2. 卸载MindIO ACP SDK。

    ```bash
    pip3 uninstall mindio_acp
    ```
