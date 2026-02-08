# 蚂蚁巡检功能使用说明

## 0. 历史数据清理(可选)

执行新集群诊断任务前, 请务必执行清理脚本

- Linux系统：[clear_cache.sh](../scripts/clear_cache.sh)

- Windows系统：[clear_cache.bat](../scripts/clear_cache.bat)

## 1. 连接信息配置

在根目录[conn.ini](../../conn.ini)文件中, 填写需要诊断的设备的远程连接方式

## 2. 收集诊断信息

在不同的网络平面下，分别执行脚本收集信息：

- Linux系统：[auto_collect.sh](linux/auto_collect.sh)

- Windows系统：[auto_collect.bat](windows/auto_collect.bat)

例如:

- 修改交换机配置, 实现分别连接不同网络平面进行采集
- 更换连接网口, 实现不同的网络平面连接

## 3. 手动合并采集数据(可选)

当分别部署在不同网络平面的诊断工具采集的数据需要汇总诊断时, 仅需将[cache](../../cache)目录打包, 然后传输到执行诊断的环境,
解压覆盖粘贴到cache目录.

## 4. 执行统一诊断

所有网络平面的信息收集完成后，执行脚本进行统一诊断(以mayi为例)：

- Linux系统：[mayi_inspection.sh](linux/mayi_inspection.sh)

- Windows系统：[mayi_inspection.bat](windows/mayi_inspection.bat)

## 6. 查看诊断报告

诊断完成后，会在[report](../../report)目录下生成CSV格式的诊断报告（inspection_errors.csv）。

