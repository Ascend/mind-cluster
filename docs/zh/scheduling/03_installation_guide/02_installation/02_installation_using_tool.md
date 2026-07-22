# 使用工具安装<a name="ZH-CN_TOPIC_0000002479386368"></a>

借助Ascend Deployer工具可以批量安装集群调度组件，大幅度简化手动安装过程中繁琐的配置操作，简化安装流程，适用于集群场景下批量安装组件。

**工具安装关键步骤**

1. 确认硬件产品和 OS 是否在 Ascend Deployer 支持列表中。
2. 部署 Ascend Deployer 工具，确保工具版本与集群调度组件版本一致。
3. 按 Ascend Deployer 工具指引配置安装参数，批量执行组件安装，步骤包括：
   - 远程连接服务器（可选）
   - 配置服务器安装部署参数
   - 执行安装命令
   - 检查安装结果
   - 配置环境变量

Ascend Deployer工具现支持的硬件产品、OS清单、安装场景请参见《MindCluster Ascend Deployer 用户指南》中的“[支持的产品和OS清单](https://gitcode.com/Ascend/ascend-deployer/blob/dev/docs/zh/01_introduction/02_supported_product_and_os.md)“章节，请根据"支持部署"列的支持情况，选择是否使用Ascend Deployer工具。

如需使用Ascend Deployer工具安装，请参考《MindCluster Ascend Deployer 用户指南》中的“[安装昇腾软件](https://gitcode.com/Ascend/ascend-deployer/blob/dev/docs/zh/05_installation_and_upgrade/02_install_softwares.md)”章节。

>[!NOTE]
>
>- 建议用户在使用工具安装前先了解[手动安装](./01_manual_installation.md)章节中相应组件的使用约束和启动参数，可以更好地帮助用户理解组件的使用场景和功能。
>- 工具版本需要与集群调度组件的版本一致，不同版本之间不可混用。
