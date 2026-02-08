# 集群自动诊断样例使用说明

## 0. 历史数据清理(可选)

执行新集群诊断任务前, 请务必执行清理脚本

- Linux系统：[clear_cache.sh](../scripts/clear_cache.sh)

- Windows系统：[clear_cache.bat](../scripts/clear_cache.bat)

## 1. 连接信息配置

在根目录[conn.ini](../../conn.ini)文件中, 填写需要诊断的设备的远程连接方式

## 2. 自动收集诊断信息

在不同的网络平面下，分别执行脚本收集信息：

- Linux系统：[auto_collect.sh](linux/auto_collect.sh)

- Windows系统：[auto_collect.bat](windows/auto_collect.bat)

例如:

- 修改交换机配置, 实现分别连接不同网络平面进行采集
- 更换连接网口, 实现不同的网络平面连接

## 3. bmc日志采集(可选)

支持批量bmc日志采集, 收集bmc日志较慢, 但可以支持更多诊断能力

- Linux系统：[collect_bmc_log.sh](linux/collect_bmc_log.sh)

- Windows系统：[collect_bmc_log.bat](windows/collect_bmc_log.bat)

## 4. 手动采集诊断信息

### 4.1 服务器带内日志&命令回显

- 通过 `tool_log_collection_out_version_all_<version>.sh`  
- 通过 `A3device日志一键采集脚本<version>.sh`

采集的带内日志(压缩包或文件夹均可)放置到工具[host_dump_cache](../../cache/host_dump_cache)目录下即可(若需要最新脚本,
请站内私信或华为内网联系wangruiju)

### 4.2 交换机命令行回显

- 方式1: 使用交换机 `display diagnostic-information <filename>` 命令导出命令回显结果集(推荐, 信息较全)
- 方式2: 查询关键命令后直接复制shell回显页面, 导出文本文件, 收集命令参考[cmd](../cmd)目录下的文本

将以上方式收集到的文本文件放置到工具[switch_cli_output_cache](../../cache/switch_cli_output_cache)目录下即可

### 4.3 BMC日志

手动通过bmc网页下载或通过命令 `ipmcget -d diaginfo` 采集的日志tar.gz包,
直接放到工具[bmc_dump_cache](../../cache/bmc_dump_cache)目录下即可

## 5. 手动合并采集数据(可选)

当分别部署在不同网络平面的诊断工具采集的数据需要汇总诊断时, 仅需将[cache](../../cache)目录打包, 然后传输到执行诊断的环境,
解压覆盖粘贴到cache目录.

## 6. 执行统一诊断

所有网络平面的信息收集完成后，执行脚本进行统一诊断：

- Linux系统：[auto_diag.sh](linux/auto_diag.sh)

- Windows系统：[auto_diag.bat](windows/auto_diag.bat)

```
注: 诊断可以不必收集完所有信息, 但越多的信息可以支持更多范围的诊断
```

## 7. 查看诊断报告

诊断完成后，会在[report](../../report)目录下生成CSV格式的诊断报告。

