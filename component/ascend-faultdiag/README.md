# MindCluster Ascend FaultDiag
-   [变更通知](#-变更通知)
-   [简介](#简介)
-   [目录结构](#目录结构)
-   [版本说明](#版本说明)
-   [环境部署](#环境部署)
-   [快速入门](#快速入门)
-   [特性介绍](#特性介绍)
-   [API参考](#API参考)
-   [FAQ](#FAQ)
-   [安全声明](#安全声明)
-   [分支维护策略](#分支维护策略)
-   [版本维护策略](#版本维护策略)
-   [免责声明](#免责声明)
-   [License](#License)
-   [建议与交流](#建议与交流)

# 📢 变更通知

- **2025-11-07**: ✨ 补充A3 AI服务器故障模式
- **2025-09-04**: ✨ 适配训练打屏日志变化
- **2025-08-22**: ✨ SDK支持故障类型扩充
- **2025-08-22**: ⚙️ 支持自定义配置
- **2025-08-22**: ✨ MindSpore故障模式补充
- **2025-08-08**: 🌐 国际化支持
- **2025-06-05**: ✨ 根因节点定位能力适配Socket并行建链
- **2025-05-23**: 🚀 提供 **模型级/POD级** 故障诊断分析

# 简介

MindCluster Ascend FaultDiag（故障诊断工具）主要功能如下：提供日志清洗和故障诊断功能，提取训练及推理过程相关日志的关键信息，并根据集群所有节点清洗后的关键信息，分析故障根因节点以及故障事件。

# 目录结构

```
ascend-faultdiag
├─build
├─platform
├─src
│  ├─ascend_fd
│  │  ├─configuration
│  │  ├─controller
│  │  ├─lib
│  │  ├─model
│  │  ├─module
│  │  │  └─mindie_trace_parser
│  │  ├─pkg
│  │  │  ├─customize
│  │  │  │  ├─custom_config
│  │  │  │  └─custom_entity
│  │  │  ├─diag
│  │  │  │  ├─knowledge_graph
│  │  │  │  │  ├─kg_engine
│  │  │  │  │  │  ├─graph
│  │  │  │  │  │  └─model
│  │  │  │  ├─network_congestion
│  │  │  │  ├─node_anomaly
│  │  │  │  │  ├─npu_anomaly
│  │  │  │  │  └─resource_preemption
│  │  │  │  │      └─utils
│  │  │  │  └─root_cluster
│  │  │  ├─parse
│  │  │  │  ├─blacklist
│  │  │  │  ├─knowledge_graph
│  │  │  │  │  ├─parser
│  │  │  │  │  └─utils
│  │  │  │  ├─network_congestion
│  │  │  │  ├─node_anomaly
│  │  │  │  └─root_cluster
│  │  ├─sdk
│  │  ├─utils
│  │  │  ├─constant
│  │  │  ├─fast_parser
│  │  │  └─timehub
│  │  └─wrapper
├─test
│  ├─custom_operation
│  ├─dt
│  └─st
└─toolkits
    ├─exp_covert
    │  └─exp_lib_dir
    └─local_diag
```

# 版本说明

MindCluster Ascend FaultDiag版本配套详情请参考：[版本配套详情](https://www.hiascend.com/developer/download/community)

# 环境部署

MindCluster Ascend FaultDiag支持的Python版本需≥3.7。在安装MindCluster Ascend FaultDiag前，请检查依赖的Python版本是否满足要求。

## 编译与构建

### 环境要求
- Python版本≥3.7.5
- scikit-learn>=1.3.0
- pandas>=1.3.5
- numpy>=1.21.6,<2.0.0
- joblib>=1.2.0,<1.5.0
- ply>=3.11

### 构建
请先克隆仓库，然后在项目根目录执行构建脚本：
```shell
git clone https://gitcode.com/Ascend/mind-cluster.git
cd mind-cluster/component/ascend-faultdiag
./build/build.sh
```

## [获取软件包](https://www.hiascend.com/zh/developer/download/community/result?module=dl%2Bcann)
获取MindCluster Ascend FaultDiag软件包。

## [命令行方式安装](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG010.html)
介绍如何以命令行方式安装MindCluster Ascend FaultDiag。

## [使用MindCluster Ascend Deployer安装](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG011.html)
介绍如何使用MindCluster Ascend Deployer安装MindCluster Ascend FaultDiag。

# 快速入门

**（可选）为普通用户配置环境变量。**

以root用户安装组件，普通用户使用时，请配置环境变量。若无法找到依赖时，请查看是否已安装该依赖或使用权限不符。
- 步骤1：以**root用户**登录并查询组件位置
    ```shell
    which ascend-fd
    ```
    回显示例如下，实际位置请以查询结果为准：
    ```
    /usr/local/python3.7.5/bin/ascend-fd
    ```
- 以**普通用户**登录配置环境变量。
    ```shell
    export PATH=$PATH:/usr/local/python3.7.5/bin
    ```
- 执行命令查看是否配置完成。
    ```shell
    ascend-fd version
    ```
    回显示例如下：
    ```shell
    ascend-fd ${版本号}
    ```

**日志清洗**
- 步骤1：上传日志至服务器。  
    上传至服务器任意目录（例如/home），以使用-i参数为例，将所有日志汇总至同一采集目录下进行清洗，目录结构示例如下。  
    Host主机侧：  
    ```
    采集目录
    |-- messages         # 主机侧操作系统日志
    |-- dmesg                # 主机侧内核消息日志
    |-- crash
        |-- 主机+故障时间目录(eg:127.xx.xx.1-2024-09-23-11:25:29)
            |-- vmcore_dmesg.txt     # 系统崩溃时保存的Host侧内核消息日志文件
    |-- sysmonitor.log       # 主机侧系统监测日志
    |-- rank-0.txt      # 训练控制台日志
    ... 
    |-- rank-7.txt      # 训练控制台日志
    |-- process_log          # CANN应用侧原始日志，目录名需为process_log
    |-- device_log           # Device侧日志，目录名需为device_log
    |-- dl_log                # MindCluster组件日志，目录名需为dl_log
        |-- devicePlugin        # Ascend Device Plugin组件日志
        |-- noded               # NodeD组件日志
        |-- ascend-docker-runtime              # Ascend Docker Runtime组件日志
        |-- volcano-scheduler              # Volcano中的volcano-scheduler组件日志
        |-- volcano-controller              # Volcano中的volcano-controller组件日志
    
        |-- npu-exporter              # NPU Exporter组件日志
    |-- mindie               # MindIE组件日志
        |-- log
            |-- debug        # MindIE组件运行日志
            |-- security     # MindIE组件审计日志
            |-- mindie_cluster_log     # MindIE Pod控制台日志
    |-- amct_log             # AMCT组件日志
    |-- environment_check # NPU网口、状态信息、资源信息
        |-- npu_info_before/after.txt  # 训练前或后NPU网口
    ```
- 步骤2：创建清洗输出目录
    ```shell
    mkdir 清洗输出目录
    ```
- 步骤3：执行命令清洗日志
    ```shell
    ascend-fd parse -i 采集目录  -o 清洗输出目录
    ```
    回显如下：
    ```
    The parse job starts. Please wait. Job id: [****], run log file is [****].
    These job ['模块1', '模块2'...] succeeded.
    The parse job is complete.
    ```
- 步骤4：日志转储  
将每台服务器的清洗输出目录下所有文件进行集中转储，转储目录结构如下。
    ```
    诊断输入目录        
        |--清洗输出目录1 
           |--plog-parser-{pid}-{0/1}.log        # 根因节点分析清洗后日志，包括error、trace等关键信息，按Pid分别保存，{0/1}代表该{pid}的plog日志有/无错误日志
           |--device_ip_info.json                # 设备IP信息
           |--ascend-kg-parser.json              # 故障事件分析清洗结果，推理引擎输入文件
           |--ascend-kg-analyzer.json            # 故障事件分析清洗结果
           |--ascend-rc-parser.json              # 根因节点分析清洗结果   
           |--mindie-cluster-info.json           # MindIE Pod控制台日志清洗结果 
           |--server-info.json.json              # MindIE组件日志清洗结果 
                   
        |--清洗输出目录2
           |--plog-parser-{pid}-{0/1}.log
           |--device_ip_info.json
           |--ascend-kg-parser.json
           |--ascend-kg-analyzer.json               
           |--ascend-rc-parser.json
           |--server-info.json.json              
        ...
        |--清洗输出目录n
    ```
**故障诊断**

- 步骤1：创建诊断结果输出目录。
    ```shell
    mkdir 诊断结果输出目录
    ```
- 步骤二：执行命令进行故障诊断
    ```shell
    ascend-fd diag -i 诊断输入目录 -o 诊断结果输出目录 
    ```
    诊断回显样例以及关键参数说明请见：[故障诊断](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG038.html)

# 特性介绍

MindCluster组件提供资源调度功能，支持NPU集群作业调度、运维监测、故障恢复等功能。具体特性介绍如下：

| 特性名称      | 介绍                                                                                                              | Released |
|-----------|-----------------------------------------------------------------------------------------------------------------|----------|
| 日志清洗与转储   | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG037.html) | ✅        |
| 故障诊断      | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG038.html) | ✅        |
| 单机故障诊断    | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG039.html) | ✅        |
| 超节点故障诊断   | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG126.html) | ✅        |
| 清洗业务流日志   | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG127.html) | ✅        |
| 根因节点清洗及诊断 | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG128.html) | ✅        |
| 故障事件清洗及诊断 | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG129.html) | ✅        |
| 自定义配置文件   | [link](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG142.html) | ✅        |

# API参考

API参考详见：[API参考](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG041.html)。

# FAQ

相关FAQ请参考：[FAQ](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG059.html)。

# 安全声明

- 安全声明详见：[安全加固](https://www.hiascend.com/document/detail/zh/mindcluster/72rc1/faultdiag/faultdiagug/mindxdlFDUG052.html)
- 公网地址详见：[公网地址](https://www.hiascend.com/doc_center/source/zh/mindcluster/72rc1/faultdiag/faultdiagug/resource/MindCluster%207.2.RC1%20Ascend%20FaultDiag%E5%85%AC%E7%BD%91%E5%9C%B0%E5%9D%80.xlsx)

# 分支维护策略

版本分支的维护阶段如下：

| 状态          | 时间     | 说明                                                      |
|-------------|--------|---------------------------------------------------------|
| 计划          | 1-3个月  | 计划特性                                                    |
| 开发          | 3个月    | 开发新特性并修复问题，定期发布新版本                                      | 
| 维护          | 3-12个月 | 常规分支维护3个月，长期支持分支维护12个月。对重大BUG进行修复，不合入新特性，并视BUG的影响发布补丁版本 | 
| 生命周期终止（EOL） | N/A    | 分支不再接受任何修改                                              |

# 版本维护策略

| 版本       | 维护策略 | 当前状态 | 发布日期       | 后续状态                 | EOL日期      |
|----------|------|------|------------|----------------------|------------|
| master   | 长期支持 | 开发   | 在研分支，不发布   | 2025-10-27           | -          |
| v7.3.0   | 长期支持 | 开发   | 在研分支，未发布   | 2025-10-27           | -          |


# 免责声明

- 本仓库代码中包含多个开发分支，这些分支可能包含未完成、实验性或未测试的功能。在正式发布前，这些分支不应被应用于任何生产环境或者依赖关键业务的项目中。请务必使用我们的正式发行版本，以确保代码的稳定性和安全性。
  使用开发分支所导致的任何问题、损失或数据损坏，本项目及其贡献者概不负责。
- 正式版本请参考release版本 <https://gitcode.com/ascend/mind-cluster/releases>

# License

MindCluster以Apache 2.0许可证许可，对应许可证文本可查阅[MindCluster根目录](https://gitcode.com/Ascend/mind-cluster/blob/master/LICENSE)。

# 建议与交流

欢迎大家为社区做贡献。如果有任何疑问或建议，请提交[issue](https://gitcode.com/Ascend/mind-cluster/issues)，我们会尽快回复。感谢您的支持。

# 致谢

MindCluster Ascend FaultDiag由华为公司的下列部门联合贡献：
- 昇腾计算应用使能开发部

感谢来自社区的每一个PR，欢迎贡献MindCluster Ascend FaultDiag！