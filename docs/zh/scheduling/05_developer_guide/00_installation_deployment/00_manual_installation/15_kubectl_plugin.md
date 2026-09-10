# Kubectl Plugin<a name="ZH-CN_TOPIC_0000002524428760"></a>

Kubectl Plugin是集群运维Agent特性的客户端命令行工具，包含 `kubectl ascend_diag` 和 `kubectl clusterops` 两个命令，安装在用户机上。Kubectl Plugin将故障诊断能力封装为kubectl命令，用户通过 `kubectl ascend_diag` 命令一键触发诊断，通过 `kubectl clusterops` 命令管理LLM服务配置。

- 使用[集群运维Agent](../../../01_introduction/02_feature_description.md#ZH-CN_TOPIC_0000002524312690)特性的用户，必须安装Kubectl Plugin。
- Kubectl Plugin为纯Python标准库实现的客户端工具，无需pip安装。
- 安装Kubectl Plugin前，需先完成Agent Core和Node Collector的部署，详细说明请参见 [Agent Core](./16_agent_core.md) 和 [Node Collector](./17_node_collector.md)。

## 操作步骤<a name="section15023132772914"></a>

1. 参考[获取软件包](./00_obtaining_software_packages.md)章节，下载Ascend ClusterOps Agent软件包。

2. 将软件包上传至用户机服务器并解压。解压后Kubectl Plugin位于kubectl-plugin子目录下，目录结构如下。

    ```shell
    Ascend-mindxdl-ascend-clusterops-agent_{version}_linux.zip
    └── kubectl-plugin/           # Kubectl Plugin
        ├── install.sh            # 一键安装脚本
        ├── kubectl-ascend_diag   # kubectl ascend_diag命令
        └── kubectl-clusterops    # kubectl clusterops命令
    ```

3. 以root用户登录用户机，进入kubectl-plugin目录，执行以下命令安装插件。

    ```shell
    cd kubectl-plugin
    bash install.sh
    ```

    回显示例如下，表示安装成功。

    ```ColdFusion
    installed. 验证:
    OK: kubectl ascend-diag / kubectl clusterops 可用
    ```

4. 验证插件是否可用。

    ```shell
    kubectl ascend-diag --help
    kubectl clusterops --help
    ```

    回显示例如下，表示插件可用。

    ```text
    $ kubectl ascend-diag --help
    usage: kubectl-ascend_diag [-h] [--job JOB] [-n NAMESPACE] [--refresh] [--json] [--collect-manifest PATH]

    Ascend fault diagnosis kubectl plugin

    $ kubectl clusterops --help
    usage: kubectl-clusterops [-h] [--create-llm-config] [--clear-llm-config] [--base-url BASE_URL] [--model MODEL]

    cluster operations kubectl plugin
    ```

5. （可选）配置LLM服务。如需对诊断报告进行智能总结，可通过`kubectl clusterops`命令创建LLM配置；不配置时，诊断报告回退为原始诊断结果。配置及清除LLM的具体方法请参见[配置和清除LLM](../../../04_usage/14_clusterops_agent/04_configuring_llm.md)。

6. 验证诊断功能。执行以下命令，按任务维度触发集群运维Agent，确认诊断链路正常。

    ```shell
    kubectl ascend_diag --job {job_name}
    ```

    命令执行后打印诊断报告（包含根因、故障事件、处置建议等字段）即表示链路正常，示例回显如下。

    ```text
    [Root Cause] HCCS链路异常，导致集合通信超时
    Root cause device: node-3:3
    [1] Fault: 940000010
        Type: hw npu hbm
        Name: HBM单粒子翻转（SEU）
        Desc: 检测到HBM错误并触发复位
        Suggestion: 1. 检查故障设备所在节点硬件状态
    ```

    完整的诊断输出说明与真实样例请参见[使用故障诊断](../../../04_usage/14_clusterops_agent/03_using_fault_diagnosis.md)。
