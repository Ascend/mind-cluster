# 日志清洗与转储<a name="ZH-CN_TOPIC_0000001541788946"></a>

>[!NOTE] 
>
>- 清洗命令指定的输出目录磁盘空间需大于5G，空间不足可能导致部分清洗结果丢失，进而导致诊断结果异常或不准确。
>- MindCluster Ascend FaultDiag在清洗时会读取用户采集的日志文件及监测指标文件，请用户自行确认目录中无敏感信息，避免信息泄露。
>- 在进行清洗时，请保证待清洗目录仅包含单台训练设备的原始日志及监测指标文件，若包含其他设备相关文件，可能导致清洗失败。
>- 执行清洗命令时，若需要清洗设备资源、网络拥塞两个性能劣化检测模块的数据，需指定--performance\(-p\)参数。当不指定时，程序默认仅清洗根因节点、故障事件两个模块的对应数据。

1. （可选）以root用户安装组件，普通用户使用时，请配置环境变量。若无法找到依赖时，请查看是否已安装该依赖或使用权限不符。
    1. 以**root用户**登录并查询组件位置。

        ```shell
        which ascend-fd
        ```

        回显示例如下，实际位置请以查询结果为主：

        ```ColdFusion
        /usr/local/python3.7.5/bin/ascend-fd
        ```

    2. 以**普通用户**登录配置环境变量。

        ```shell
        export PATH=$PATH:/usr/local/python3.7.5/bin
        ```

    3. 执行命令查看是否配置完成。

        ```shell
        ascend-fd -h
        ```

        显示以下内容即表示配置完成。

        ```ColdFusion
        usage: ascend-fd [-h] {version,parse,diag,blacklist,config,entity,single-diag} ...
        Ascend Fault Diag
        positional arguments:
          {version,parse,diag,blacklist,config,entity,single-diag}
            version             show ascend-fd version
            parse               parse origin log files
            diag                diag parsed log files
            blacklist           filter invalid CANN logs by blacklist for parsing
            config              custom configuration parsing files
            entity              perform operations on the user-defined faulty entity.
            single-diag         single parse and diag log files
        optional arguments:
          -h, --help            show this help message and exit
        ```

2. 参见[日志采集](./03_collecting_logs.md)完成训练设备日志收集。

    上传至服务器任意目录（例如/home），以使用-i参数为例，将所有日志汇总至同一采集目录下进行清洗，目录结构示例如下。

    - Host主机侧：

        ```text
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
            |-- ttp_log                   # MindIO组件日志
        |-- mindie               # MindIE组件日志
            |-- log
                |-- debug        # MindIE组件运行日志
                |-- security     # MindIE组件审计日志
                |-- mindie_cluster_log     # MindIE Pod控制台日志
        |-- amct_log             # AMCT组件日志
        |-- bus_log              # Ascend 950代际LCNE组件日志
        |-- environment_check # NPU网口、状态信息、资源信息
            |-- npu_smi_0_details.csv   # NPU状态监测指标文件
             ...
            |-- npu_smi_7_details.csv   # NPU状态监测指标文件
            |-- npu_0_details.csv       # NPU网口统计监测指标文件
             ...    
            |-- npu_7_details.csv       # NPU网口统计监测指标文件
            |-- npu_info_before/after.txt  # 训练前或后NPU网口
            |-- host_metrics_{core_num}.json # 主机资源监测指标文件
        ```

    - BMC及LCNE侧：

        将Computing Toolkit或CCAE导出的BMC及LCNE侧日志，递归解压后进行单机放置、单机清洗。

        ```shell
        ascend-fd parse --lcne_log 解压后的单节点LCNE日志目录 -o 清洗结果输出目录
        ascend-fd parse --bmc_log 解压后的单节点BMC日志目录 -o 清洗结果输出目录
        ```

        >[!NOTE] 
        >- 使用CCAE进行日志收集，可参考[灵衢日志采集](https://support.huawei.com/hedex/hdx.do?docid=EDOC1100485430&id=ZH-CN_TOPIC_0000002240474597)。
        >- 使用Computing Toolkit进行日志收集，可参考[《Computing Toolkit 用户指南》](https://support.huawei.com/carrier/productNewOffering?col=product&path=PBI1-262732867/PBI1-262735884/PBI1-261914673/PBI1-264314551)\>使用Computing Toolkit\>日志收集\>使用指导\>服务器（BMC）、IES、SWITCH日志收集。

3. 创建日志清洗输出目录。

    ```shell
    mkdir 清洗输出目录
    ```

4. 执行命令开始清洗日志。

    ```shell
    ascend-fd parse -i 采集目录  -o 清洗输出目录 --performance
    ```

    回显如下：

    ```ColdFusion
    The parse job starts. Please wait. Job id: [****], run log file is [****].
    These job ['模块1', '模块2'...] succeeded.
    The parse job is complete.
    ```

    清洗输出目录结构：

    ```text
    └── 清洗输出目录 
       ├── ascend-kg-parser.json        # 故障事件分析清洗结果，推理引擎输入文件
       ├── ascend-kg-analyzer.json      # 故障事件分析清洗结果
       ├── ascend-rc-parser.json        # 根因节点分析清洗结果
       ├── device_ip_info.json          # 设备IP信息
       ├── mindie-cluster-info.json     # MindIE Pod控制台日志清洗结果
       ├── server-info.json             # MindIE组件日志清洗结果
       ├── nad_clean.csv                # 计算降频清洗输出结果
       ├── nic_clean.csv                # 网络拥塞清洗输出结果
       ├── process_{core_num}.csv       # CPU资源抢占清洗输出结果
       ├── plog-parser-{pid}-{0/1}.log # 根因节点分析清洗后日志，包括error、trace等关键信息，按Pid分别保存
        ...
       └── plog-parser-{pid}-{0/1}.log
    ```

5. 日志转储。

    将每台服务器的清洗输出目录下所有文件进行集中转储，转储目录结构如下。

    ```text
    诊断输入目录        
        |--清洗输出目录1 
           |--plog-parser-{pid}-{0/1}.log        # 根因节点分析清洗后日志，包括error、trace等关键信息，按Pid分别保存
           |--nic_clean.csv                      # 网络拥塞清洗输出结果
           |--nad_clean.csv                      # 计算降频清洗输出结果
           |--mem_used.csv                       # 内存资源抢占清洗输出结果，预留文件，当前暂未使用，
           |--process_{core_num}.csv             # CPU资源抢占清洗输出结果
           |--device_ip_info.json                # 设备IP信息
           |--ascend-kg-parser.json              # 故障事件分析清洗结果，推理引擎输入文件
           |--ascend-kg-analyzer.json            # 故障事件分析清洗结果
           |--ascend-rc-parser.json              # 根因节点分析清洗结果   
           |--mindie-cluster-info.json           # MindIE Pod控制台日志清洗结果 
           |--server-info.json                   # MindIE组件日志清洗结果 
                   
        |--清洗输出目录2
           |--plog-parser-{pid}-{0/1}.log        
           |--nic_clean.csv  
           |--nad_clean.csv  
           |--mem_used.csv  
           |--process_{core_num}.csv
           |--device_ip_info.json
           |--ascend-kg-parser.json
           |--ascend-kg-analyzer.json               
           |--ascend-rc-parser.json
           |--server-info.json                   ...
        |--清洗输出目录n
    ```

>[!NOTE] 
>
>- 清洗输出目录的名称建议修改为能标识出设备节点信息的目录名，例如：host1-192.168.x.x。
>- MindIE Pod控制台日志清洗结果仅需在一个节点内存储即可。
