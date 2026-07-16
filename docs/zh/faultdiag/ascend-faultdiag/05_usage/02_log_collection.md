# 日志采集

日志采集是使用 ascend-fd 的第一步。用户需要从各台训练/推理设备上收集日志文件，并按指定结构存放。

## 采集日志归档目录结构

当前日志分为三种类型，Host 侧日志、BMC 日志和 LCNE 日志。

Host 侧日志为 BMC 和 LCNE 以外的所有日志。

BMC 日志为 BMC 软件上收集的日志。

LCNE 日志为 Ascend 950 系列产品 LCNE 组件运行日志。

将所有日志汇总到同一个采集目录下，目录结构如下所示。

```text
采集目录
|-- messages             # 主机侧操作系统日志
|-- dmesg                # 主机侧内核消息日志
|-- crash
    └── 主机+故障时间目录
        └── vmcore_dmesg.txt     # 系统崩溃时的内核日志
|-- sysmonitor.log       # 主机侧系统监测日志
|-- rank-0.txt           # 训练控制台日志（第 0 张卡）
...
|-- rank-7.txt           # 训练控制台日志（第 7 张卡）
|-- dmidecode.txt        # 主机侧硬件信息日志
|-- process_log          # CANN 应用类日志（目录名必须为 process_log）
|-- device_log           # Device 侧日志（目录名必须为 device_log）
|-- dl_log               # MindCluster 组件日志（目录名必须为 dl_log）
    |-- devicePlugin       # Ascend Device Plugin 日志
    |-- noded              # NodeD 日志
    |-- ascend-docker-runtime  # Docker Runtime 日志
    |-- volcano-scheduler      # Volcano 调度器日志
    |-- volcano-controller     # Volcano 控制器日志
    |-- npu-exporter           # NPU Exporter 日志
    └── ttp_log                # MindIO 组件日志
|-- mindie               # MindIE 组件日志
    └── log
        |-- debug        # 运行日志
        |-- security     # 审计日志
        └── mindie_cluster_log  # MindIE Pod 日志
|-- amct_log             # AMCT 组件日志
|-- lcne_log             # LCNE 组件日志
    |-- log.log
    |-- log_1_*.log
    |-- diag_display_info.txt
    └── diagnostic_information
        └── slot_1
            └── tempdir
                └── devm_bddrvadp.log
|-- environment_check    # 环境检查日志，NPU 网口、状态、资源信息
    |-- npu_smi_0_details.csv   # NPU 状态监测
    |-- npu_0_details.csv       # NPU 网口统计
    |-- npu_info_before.txt     # 训练前 NPU 环境检查
    |-- npu_info_after.txt      # 训练后 NPU 环境检查
    └── host_metrics_{core_num}.json  # 主机资源监测
|-- pymotor_vllm_log     # PyMotor/vLLM 日志
└── bmc_log              # BMC 侧日志
    └── dump_info
        |-- AppDump
        |-- DeviceDump
        └── LogDump
```

## 注意事项

- 输入目录的日志文件总大小应限制在 5GB 以下，文件总数量不超过 1000000
- CANN 应用类日志的单个文件应限制在 20MB 以下
- NPU 状态监测指标文件、NPU 网口统计文件应限制在 512MB 以下
- 用户训练/推理日志默认只读取最后 1MB 内容
- 如果在容器中进行训练或推理，请及时将日志保存至宿主机，如训练及推理日志、CANN 应用类日志等

## 组件日志采集

### CANN 应用类日志

训练或推理结束后，执行以下命令将日志复制至 `采集目录/process_log` 下：

```shell
cp -r $HOME/ascend/log/* {采集目录}/process_log
```

目录结构：

```text
|--process_log
    |--debug
        |--plog               # Host 侧应用类日志目录
            └──plog-{pid}_{unix时间}.log
        |--device-0           # Device 侧应用类日志目录
            └──device-{pid}_{unix时间}.log
        |--device-1
        |--device-2
        |--…
    |--run
    |--operation
    └──security
```

> [!NOTE]
>
> - CANN 应用类日志默认存储在 `$HOME/ascend/log`目录下，并支持通过环境变量`ASCEND_PROCESS_LOG_PATH`自定义日志存储路径。
> - 文件说明：由 CANN 打印的应用类日志，包括 Host 侧应用类日志 `plog-{pid}_{unix时间}.log` 和 Device 侧应用类日志 `device-{pid}_{unix时间}.log` 两类，更多日志相关信息请参见[《CANN 日志参考》](https://www.hiascend.com/document/detail/zh/canncommercial/900/maintenref/logreference/logreference_0001.html)中的“[查看日志（Ascend EP）](https://www.hiascend.com/document/detail/zh/canncommercial/900/maintenref/logreference/logreference_0002.html)”章节。

### 用户训练/推理日志

训练或推理结束后，将训练或推理日志复制一份至 `采集目录` 下，并为每张卡的训练及推理日志按照以下格式命名：

- `rank-(rank_id).log` 或 `rank-(rank_id).txt`
- `worker-(worker_id).log` 或 `worker-(worker_id).txt`

**图 1**  训练及推理日志示例

![训练及推理日志示例](../../figures/ascend-faultdiag/训练及推理日志示例.png "训练及推理日志示例")

> - 当使用 AI 框架时，训练及推理日志为 Python 打印在屏幕上的日志，通常用户会通过重定向方式存储在本地，在 PyTorch 框架下，控制台日志仅有一份。

### 环境检查日志

环境检查日志包括训练及推理前、中、后的环境检查日志。

#### 训练或推理前环境检查日志

使用[环境检查日志采集脚本](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector/) 中的 `npu_info_collect.sh` 采集训练或推理前 NPU 环境信息。

执行以下命令进行采集。

```shell
bash npu_info_collect.sh {采集目录}/environment_check/npu_info_before.txt
```

#### 训练或推理中环境检查日志

使用[环境检查日志采集脚本](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector/) 中的 `net_data_collect.py` 采集 NPU 网口统计监测指标。

启动该脚本采集会对训练或推理任务有性能影响。

执行以下命令进行采集。

```shell
python net_data_collect.py -n 8 -it 15 -o {采集目录}/environment_check/
```

脚本参数说明：

| 参数            | 类型   | 必选 | 说明                                   |
|-----------------|--------|------|----------------------------------------|
| -n, --num       | int    | 必选 | NPU 卡数量                             |
| -it, --interval | int    | 必选 | 采集间隔（单位：秒）, 15秒采集一次即可 |
| -o, --output    | string | 必选 | 输出采集目录                           |

> [!NOTE]
>
> - NPU 卡数量必须与训练及推理时使用的 NPU 卡数量一致。
> - 采集间隔为 15 秒，根据实际情况调整。
> - 采集完成，会在 `{采集目录}/environment_check/` 下生成 `npu_0_details.csv`、`npu_1_details.csv` 等文件。

#### 训练或推理后环境检查日志

使用[环境检查日志采集脚本](https://gitcode.com/Ascend/mindxdl-deploy/tree/master/npu_collector/) 中的 `npu_info_collect.sh` 采集训练或推理后 NPU 环境信息。

执行以下命令进行采集。

```shell
bash npu_info_collect.sh {采集目录}/environment_check/npu_info_after.txt
```

### 主机侧日志

训练或推理结束后，主机侧需要收集以下日志：

| 日志说明                     | 命名格式         | 存放路径                   |
|------------------------------|------------------|----------------------------|
| 主机侧操作系统日志           | messages-*       | 采集目录/                  |
| 主机侧内核消息日志           | dmesg            | 采集目录/                  |
| 主机侧系统监测日志           | sysmonitor.log   | 采集目录/                  |
| 系统崩溃时主机侧内核消息日志 | vmcore_dmesg.txt | 采集目录/crash/{系统时间}/ |
| 主机侧硬件信息日志           | dmidecode.txt    | 采集目录/                  |

#### 主机侧操作系统日志

主机侧操作系统日志 `messages` 存储在 `/var/log` 目录中。

用户需手动将训练及推理开始与结束时间对应的日志信息复制一份至 `采集目录` 下。

#### 主机侧内核消息日志

使用以下命令采集最新的 dmesg 日志到采集目录，最大采集 100000 行。

```shell
dmesg -T | tail -n 100000 > {采集目录}/dmesg
```

#### 主机侧系统监测日志

使用以下命令复制 `sysmonitor.log` 日志到采集目录。

```shell
cp -r /var/log/sysmonitor.log {采集目录}/
```

#### 系统崩溃时主机侧内核消息日志

主机侧内核消息日志为系统崩溃时保存的 Host 侧内核消息文件。

执行如下命令复制日志到采集目录下。

```shell
cp -r /var/crash/ {采集目录}/
```

#### 主机侧硬件信息日志

主机侧包含 dmi 硬件信息的 dmidecode 日志。

执行如下命令采集 dmidecode 日志到采集目录下。

```shell
dmidecode > {采集目录}/dmidecode.txt
```

### Device 侧日志

训练或推理结束后，执行命令采集 Device 侧日志。

```shell
msnpureport
```

执行完成后，会在当前目录生成以时间戳命名的日志数据，执行以下命令复制日志到采集目录。

```shell
cp -r {时间戳目录}/slog {采集目录}/device_log
cp -r {时间戳目录}/hisi_logs {采集目录}/device_log
```

目录结构：

- Ascend HDK 23.0.RC3 版本

    ```text
    |--device_log
        |-- slog
            |-- dev-os-3
                |-- debug
                    |--device-os
                        └── device-os_{time}.log # Device 侧 Control CPU 上的系统类日志
                |-- run
                    |--device-os
                        └── device-os_{time}.log # Device 侧 Control CPU 上的系统类日志
                |--device-0
                    └──device-0_{time}.log   # Device 侧非 Control CPU 上的系统类日志
                |--device-2
                |--…
                |--slogd
                └──device_sys_init_ext.log
            |-- dev-os-7
            |-- …
        └──hisi_logs
            |-- device-0
                |-- …
                └── history.log     # 黑匣子日志
            |-- device-2
            |-- …
            └── device_info.txt
    ```

- Ascend HDK 23.0.3 及以上版本

    ```text
    |--device_log
        |-- slog
            |-- dev-os-3
                |-- debug
                    |--device-os
                        |-- device-os_{time}.log # Device 侧 Control CPU 上的系统类日志
                    |--device-0
                        |--device-0_{time}.log   # Device 侧非 Control CPU 上的系统类日志
                    |--device-2
                    |--…
                |-- run
                    |--device-os
                        └── device-os_{time}.log # Device 侧 Control CPU 上的系统类日志
                    └──event
                        └── event_{time}.log # Device Control CPU 的 EVENT 级别系统日志
                |--…
                |--slogd
                └──device_sys_init_ext.log
            |-- dev-os-7
            └── …
        └──hisi_logs
            └── device-0
                |-- …
                |-- history.log                  # 黑匣子日志
                |-- {time}/log/kernel.log        # NPU 芯片内核日志
                |-- {time}/bbox/os/os_info.txt   # Device 侧 OS 基本信息
                └── {time}/mntn/hbm.txt          # Device 侧片上内存日志
            |-- device-2
            |-- …
            └── device_info.txt
    ```

### MindCluster 组件日志

训练或推理结束后，将 MindCluster 组件日志（默认路径为 `/var/log/mindx-dl/`）复制到 `采集目录/dl_log` 下。

执行以下命令采集 MindCluster 组件日志。

```shell
cp -r /var/log/mindx-dl/devicePlugin {采集目录}/dl_log
cp -r /var/log/mindx-dl/noded {采集目录}/dl_log
cp -r /var/log/ascend-docker-runtime {采集目录}/dl_log
cp -r /var/log/mindx-dl/volcano-scheduler {采集目录}/dl_log
cp -r /var/log/mindx-dl/volcano-controller {采集目录}/dl_log
cp -r /var/log/mindx-dl/npu-exporter {采集目录}/dl_log
```

> [!NOTE]
>
> - 默认日志存储路径为`/var/log/mindx-dl/`，如果用户自定义了 MindCluster 组件日志落盘路径，请使用自定义路径。
> - MindCluster 组件日志文件名格式为 `devicePlugin*.log`、`noded*.log`、`runtime-run*.log`、`hook-run*.log`、`volcano-scheduler*.log`、`volcano-controller*.log`、`npu-exporter*.log`。

### MindIE 组件日志

训练或推理结束后，将 MindIE 组件日志（日志名 `mindie-{module}_{pid}_{datetime}.log`）复制至 `采集目录/mindie/log/debug` 下。

采集前先检查环境是否有设置 MindIE 组件日志落盘路径：

```shell
env | grep "MINDIE_LOG_PATH"
```

- 若无结果显示或结果显示中不包含绝对路径，例如回显为以下：

    ```shell
    MINDIE_LOG_PATH="llm: llm"
    ```

    代表日志存储在默认路径下，使用以下命令进入日志默认存储目录，拷贝相关组件日志

    ```shell
    cp -r ~/mindie {采集目录}/
    ```

- 若有结果显示且结果显示中包含绝对路径，例如回显为以下：

    ```shell
    MINDIE_LOG_PATH="llm: /home/working/"
    ```

    则需要进入回显中对应的日志存储目录，拷贝相关组件日志。

    ```shell
    cp -r /home/working {采集目录}
    ```

### MindIE Pod 日志

训练或推理结束后，将 MindIE Pod 组件日志转储至 `采集目录/mindie/log/mindie_cluster_log/` 下。

参考[Pod 日志采集脚本](https://gitcode.com/Ascend/mindxdl-deploy/blob/master/mindie/pod_log_collect.sh)中的 `pod_log_collect.sh` 编写采集脚本。

确认脚本采集输出路径为`采集目录/mindie/log/mindie_cluster_log/`，可在任意目录执行命令采集，执行步骤如下：

- 添加日志输出路径

```shell
log_dir="{采集目录}/mindie/log/mindie_cluster_log/"
```

- 执行采集脚本

```shell
bash pod_log_collect.sh
```

脚本执行完成，会在 `{采集目录}/mindie/log/mindie_cluster_log/` 下生成 `${pod_name}.json` 文件。

日志内容样例:

<!-- markdownlint-disable-next-line MD033 -->
<pre>
……
INFO:root:status of ranktable is not completed, waiting for file update.
INFO:root:status of ranktable is not completed, waiting for file update.
INFO:root:status of ranktable is not completed, waiting for file update.
{"IsMindIEEPJob":true,"status":"completed","server_list":[{"device":[{"device_id":"0","device_ip":"10.0.2.41","super_device_id":"113246208","rank_id":"0"},{"device_id":"1","device_ip":"10.0.3.41","super_device_id":"113311745","rank_id":"1"},{"device_id":"2","device_ip":"10.0.2.42","super_device_id":"113508354","rank_id":"2"},{"device_id":"3","device_ip":"10.0.3.42","super_device_id":"113573891","rank_id":"3"},{"device_id":"4","device_ip":"10.0.2.43","super_device_id":"113770500","rank_id":"4"},{"device_id":"5","device_ip":"10.0.3.43","super_device_id":"113836037","rank_id":"5"},{"device_id":"6","device_ip":"10.0.2.44","super_device_id":"114032646","rank_id":"6"},{"device_id":"7","device_ip":"10.0.3.44","super_device_id":"114098183","rank_id":"7"},{"device_id":"8","device_ip":"10.0.2.45","super_device_id":"114294792","rank_id":"8"},{"device_id":"9","device_ip":"10.0.3.45","super_device_id":"114360329","rank_id":"9"},{"device_id":"10","device_ip":"10.0.2.46","super_device_id":"114556938","rank_id":"10"},{"device_id":"11","device_ip":"10.0.3.46","super_device_id":"114622475","rank_id":"11"},{"device_id":"12","device_ip":"10.0.2.47","super_device_id":"114819084","rank_id":"12"},{"device_id":"13","device_ip":"10.0.3.47","super_device_id":"114884621","rank_id":"13"},{"device_id":"14","device_ip":"10.0.2.48","super_device_id":"115081230","rank_id":"14"},{"device_id":"15","device_ip":"10.0.3.48","super_device_id":"115146767","rank_id":"15"}],"server_id":"10.0.0.1","container_ip":"192.168.247.11"}],"server_count":"1","version":"1.2","super_pod_list":[{"super_pod_id":"1","server_list":[{"server_id":"10.0.0.1"}]}]}
……
</pre>

### AMCT 组件日志

模型压缩时，会根据量化进程数量产生对应数量的日志，通常只会启动一个量化进程，即产生一个对应日志 `amct_{framework}.log`。

训练或推理结束后，将 AMCT 组件日志复制至 `采集目录/amct_log/` 下。

执行以下命令进行采集：

```shell
cp -r ~/amct_log {采集目录}/amct_log
```

### MindIO 组件日志

MindIO 组件运行时，每个进程会产生一个 `ttp_log.log.*` 日志文件。

训练或推理结束后，将 MindIO 组件日志复制至 `采集目录/dl_log/ttp_log/` 下。

执行以下命令进行采集：

```shell
cp -r ~/ttp_log {采集目录}/dl_log/ttp_log
```

### LCNE 日志（原 Bus 日志）

训练或推理结束后，需要采集 LCNE 组件日志

- **Ascend 950 系列**

Ascend 950 系列 LCNE 组件运行时，会产生相关日志文件 `log.log` 。

将 `log.log` 日志复制至 `采集目录/lcne_log/` 下。可按照以下方式进行采集：

1. 进入 Ascend 950 系列产品 1213 后台，在 `/opt/vrpv8/home/logfile` 目录下获取 `log.log` 日志。
2. 进入 Ascend 950 系列产品 1213 前台，执行 **collect diagnostic information** 命令采集日志后，从 Ascend 950 系列产品 1213 后台的 `/opt/vrpv8/home/logfile` 目录下获取 `diagnostic_information_*.zip` 压缩日志文件。需要手动解压所有压缩日志。

- **Atlas A3 系列**

直接将 smartkit 或 CCAE 导出的 LCNE 日志递归解压后放置到 `采集目录/lcne_log/` 下。

> [!NOTE]
>
> - 如果不能确定 LCNE 归属于哪个节点，请单独放置到一个目录，如 `LCNE采集日志/` 下。

### BMC 日志

训练或推理结束后，支持两种方式采集 BMC 日志

- 通过 BMC 网页 `一键下载` 按钮下载 BMC 日志。
- 通过 shell 登录 BMC，使用 `ipmcget -d diaginfo` 命令采集 BMC 日志。

> [!NOTE]
>
> - 两种方式采集的 BMC 文件，均为压缩包，需要手动解压后放到 `采集目录/bmc_log/` 下。
> - 如果不能确定 BMC 归属于哪个节点，请单独放置到一个目录，如 `BMC采集日志/` 下。

### PyMotor/vLLM 日志

MindIE-PyMotor、vLLM、vLLM-Ascend 运行产生的日志。

MindIE-PyMotor 部署完成后，会自动启动 MindIE-PyMotor、vLLM 和 vLLM-Ascend 日志的收集。

要求用户手动将日志复制至 `采集目录/pymotor_vllm_log/` 下。

> [!NOTE]
>
> MindIE-PyMotor 部署详情请参见[MindIE-PyMotor 部署](https://gitcode.com/Ascend/MindIE-PyMotor/blob/master/docs/zh/user_guide/deployment/k8s/pd_aggregation_deployment.md#%E6%9F%A5%E7%9C%8B%E9%9B%86%E7%BE%A4%E7%8A%B6%E6%80%81%E4%B8%8E%E6%97%A5%E5%BF%97)。
> 日志名格式为 `mindie-motor-controller-*.log`、`mindie-motor-coordinator-*.log`、`vllm-d0-*.log`、`vllm-p0-*.log`。
